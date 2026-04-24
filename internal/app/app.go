package app

import (
	"flag"
	"log"
	"time"

	"go-far/internal/config"
	httpHandler "go-far/internal/handler/http"
	schedHandler "go-far/internal/handler/scheduler"
	"go-far/internal/infra/database"
	"go-far/internal/infra/grace"
	httpclient "go-far/internal/infra/http/client"
	httpmux "go-far/internal/infra/http/mux"
	httpserver "go-far/internal/infra/http/server"
	"go-far/internal/infra/logger"
	"go-far/internal/infra/middleware"
	"go-far/internal/infra/query"
	cfgredis "go-far/internal/infra/redis"
	cfgscheduler "go-far/internal/infra/scheduler"
	"go-far/internal/infra/token"
	"go-far/internal/infra/tracer"
	"go-far/internal/infra/validator"
	"go-far/internal/preference"
	"go-far/internal/repository"
	"go-far/internal/service"
	"go-far/internal/util"
)

const (
	DefaultMinJitter = 100
	DefaultMaxJitter = 2000
)

// Run initializes everything and starts the server.
// It blocks until a shutdown signal is received, then exits gracefully.
func Run() {
	// Add sleep with Jitter to drag the the initialization time among instances
	minJitter, maxJitter := parseFlags()
	sleepWithJitter(minJitter, maxJitter)

	// Config Initialization
	conf, err := config.InitConfig()
	if err != nil {
		panic(err)
	}

	// Logger Initialization
	log := logger.InitLogger(conf.Logger)

	// SQL Initialization
	sql0 := database.InitDB(log, conf.Database.Postgres)
	defer sql0.Close()

	// Redis Initialization
	redis0 := cfgredis.InitRedis(log, conf.Redis, preference.REDIS_APPS)
	defer func() {
		if err := redis0.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close Redis APPS connection")
		}
	}()

	redis1 := cfgredis.InitRedis(log, conf.Redis, preference.REDIS_AUTH)
	defer func() {
		if err := redis1.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close Redis AUTH connection")
		}
	}()

	redis2 := cfgredis.InitRedis(log, conf.Redis, preference.REDIS_LIMITER)
	defer func() {
		if err := redis2.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close Redis LIMITER connection")
		}
	}()

	// HTTP Client Initialization
	httpClient := httpclient.InitHttpClient(log, conf.HTTP.Client)

	// Query Loader Initialization
	queryLoader := query.InitQueryLoader(log, conf.Queries)

	// Business Layers Initialization
	repo := repository.InitRepository(sql0, redis0, queryLoader, conf.Redis.CacheTTL)
	svc := service.InitService(repo)

	// Tracer Initialization
	if conf.Tracer.Enabled {
		tracerInst := tracer.InitTracer(log, conf.Tracer)
		defer tracerInst.Stop()
	}

	// Auth & MiddlewareInitialization
	authToken := token.InitToken(log, conf.Token, redis1)
	mw := middleware.InitMiddleware(log, conf.Middleware, authToken, redis2, conf.Tracer.Enabled)

	// HTTP router & validator
	httpMux := httpmux.InitHttpMux(log)
	validator.InitValidator(log)

	// REST Handler Initialization (registers routes on mux)
	httpHandler.InitHttpHandler(httpMux, authToken, mw, svc, svc.User, sql0, redis0)

	// Scheduler Initialization
	if conf.Scheduler.Enabled {
		scheduler := cfgscheduler.InitScheduler(log, conf.Scheduler)
		schedHandler.InitSchedulerHandler(log, scheduler, svc, &conf.Scheduler.SchedulerJobs, httpClient, conf.Scheduler.Enabled)
		defer scheduler.Stop()
	}

	// HTTP Server Initialization
	httpServer := httpserver.InitHttpServer(log, conf.HTTP.Server, mw, httpMux)

	// App Initialization
	app := grace.InitGrace(log, httpServer, conf.HTTP.Server.ShutdownTimeout)
	app.Serve()
}

func parseFlags() (minJitter, maxJitter int) {
	flag.IntVar(&minJitter, "minSleep", DefaultMinJitter, "min. sleep duration during app initialization")
	flag.IntVar(&maxJitter, "maxSleep", DefaultMaxJitter, "max. sleep duration during app initialization")
	flag.Parse()

	return
}

func sleepWithJitter(low, high int) {
	if low < 1 {
		low = DefaultMinJitter
	}

	if high < 1 || high < low {
		high = DefaultMaxJitter
	}

	rnd := util.RandomInt(high-low) + low
	time.Sleep(time.Duration(rnd) * time.Millisecond)

	log.Printf("%d ms sleep during initialization", rnd)
}
