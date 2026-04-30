package app

import (
	"flag"
	"log"
	"net/http"
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
	"go-far/internal/infra/metrics"
	"go-far/internal/infra/middleware"
	"go-far/internal/infra/pyroscope"
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

	"github.com/rs/zerolog"
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

	// Redis Initialization - Apps
	redis0 := cfgredis.InitRedis(log, conf.Redis, preference.REDIS_APPS)
	defer func() {
		if err := redis0.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close Redis APPS connection")
		}
	}()

	// Redis Initialization - Auth
	redis1 := cfgredis.InitRedis(log, conf.Redis, preference.REDIS_AUTH)
	defer func() {
		if err := redis1.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close Redis AUTH connection")
		}
	}()

	// Redis Initialization - Limiter
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
	var tracerInst tracer.Tracer
	if conf.Tracer.Enabled {
		tracerInst = tracer.InitTracer(log, conf.Tracer)
		defer tracerInst.Stop()
	}

	// Metrics Initialization
	var metricsInst metrics.Metrics
	if conf.Metric.Enabled {
		metricsInst = metrics.InitMetrics(log, sql0, redis0, redis1, redis2)
	}

	// Auth & MiddlewareInitialization
	authToken := token.InitToken(log, conf.Token, redis1)
	mw := middleware.InitMiddleware(log, conf.Middleware, authToken, redis2, conf.Tracer.Enabled, metricsInst)

	// HTTP router & validator
	httpMux := httpmux.InitHttpMux(log, metricsInst)
	validator.InitValidator(log)

	// REST Handler Initialization (registers routes on mux)
	httpHandler.InitHttpHandler(httpMux, authToken, mw, svc, svc.User, sql0, redis0)

	// Scheduler Initialization
	scheduler := initScheduler(log, conf, svc, httpClient)
	if scheduler != nil {
		defer scheduler.Stop()
	}

	// 🔍 Start Pyroscope Profiler (VisualVM-like continuous profiling)
	if conf.Pyroscope.Enabled {
		pyro := pyroscope.InitPyroscope(log, conf.App)
		defer pyro.Stop()
	}

	// HTTP Server Initialization
	httpServer := httpserver.InitHttpServer(log, conf.HTTP.Server, mw, httpMux)

	// Metrics Server Initialization
	var metricsServer *http.Server
	if conf.Metric.Enabled && conf.HTTP.MetricsServer != nil {
		metricsServer = httpserver.InitMetricsServer(log, conf.HTTP.MetricsServer, metricsInst.HTTPHandler())
	}

	// App Initialization
	app := grace.InitGrace(log, httpServer, metricsServer, conf.App.ShutdownTimeout)
	app.Serve()
}

func initScheduler(log *zerolog.Logger, conf *config.Config, svc *service.Service, httpClient *http.Client) *cfgscheduler.Scheduler {
	if !conf.Scheduler.Enabled {
		return nil
	}

	scheduler, schedulerMetrics := cfgscheduler.InitScheduler(log, conf.Scheduler, conf.Tracer.Enabled, metrics.GetRegistry())
	schedHandler.InitSchedulerHandler(&schedHandler.SchedulerHandlerOptions{
		Log:            log,
		Sch:            scheduler,
		Svc:            svc,
		Jobs:           &conf.Scheduler.SchedulerJobs,
		HTTPClient:     httpClient,
		Metrics:        schedulerMetrics,
		Enabled:        conf.Scheduler.Enabled,
		TracingEnabled: conf.Tracer.Enabled,
	})

	return scheduler
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
