package main

import (
	"flag"

	_ "go-far/docs"
	"go-far/src/config/auth"
	"go-far/src/config/database"
	"go-far/src/config/grace"
	"go-far/src/config/logger"
	"go-far/src/config/middleware"
	"go-far/src/config/query"
	cfgredis "go-far/src/config/redis"
	cfgscheduler "go-far/src/config/scheduler"
	"go-far/src/config/server"
	"go-far/src/config/tracer"
	restHandler "go-far/src/handler/rest"
	schedHandler "go-far/src/handler/scheduler"
	"go-far/src/preference"
	"go-far/src/repository"
	"go-far/src/service"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

var (
	minJitter int
	maxJitter int

	sql0      *sqlx.DB
	redis0    *redis.Client
	redis1    *redis.Client
	redis2    *redis.Client
	scheduler *cfgscheduler.Scheduler

	tracerInst tracer.Tracer
	app        grace.App
)

func init() {
	flag.IntVar(&minJitter, "minSleep", DefaultMinJitter, "min. sleep duration during app initialization")
	flag.IntVar(&maxJitter, "maxSleep", DefaultMaxJitter, "max. sleep duration during app initialization")
	flag.Parse()

	// Add sleep with Jitter to drag the the initialization time among instances
	sleepWithJitter(minJitter, maxJitter)

	// Config Initialization
	conf, err := InitConfig()
	if err != nil {
		panic(err)
	}

	// Logger Initialization
	log := logger.InitLogger(conf.Logger)

	// SQL Initialization
	sql0 = database.InitDB(log, conf.Postgres)

	// Redis Initialization
	redis0 = cfgredis.InitRedis(log, conf.Redis, preference.REDIS_APPS)
	redis1 = cfgredis.InitRedis(log, conf.Redis, preference.REDIS_AUTH)
	redis2 = cfgredis.InitRedis(log, conf.Redis, preference.REDIS_LIMITER)

	// Query Loader Initialization
	queryLoader := query.InitQueryLoader(log, conf.Queries)

	// Initialize dependencies
	repository := repository.InitRepository(sql0, redis0, queryLoader, conf.Redis.CacheTTL)
	service := service.InitService(repository)

	// Initialize validator
	middleware.InitValidator(log)

	// Auth Initialization
	authInst := auth.InitAuth(log, conf.Auth, redis1)

	// Middleware Initialization
	mw := middleware.InitMiddleware(log, conf.Middleware, authInst, redis2)

	// HTTP Gin Initialization
	httpGin := server.InitHttpGin(log, mw, conf.Gin)

	// REST Handler Initialization
	restHandler.InitRestHandler(httpGin, authInst, mw, service)

	// Scheduler Initialization
	scheduler = cfgscheduler.InitScheduler(log, conf.Scheduler)
	schedHandler.InitSchedulerHandler(log, scheduler, service, conf.Scheduler.SchedulerJobs)

	// HTTP Server Initialization
	httpServer := server.InitHttpServer(log, conf.Server, httpGin)

	// Tracer Initialization
	tracerInst = tracer.InitTracer(log, conf.Tracer)

	// App Initialization
	app = grace.InitGrace(log, httpServer, tracerInst)
}

// @title			Go-Far
// @version		1.0
// @description	Clean Architecture CRUD API with Go
// @termsOfService	http://swagger.io/terms/
// @contact.name	API Support
// @contact.url	http://www.swagger.io/support
// @contact.email	support@swagger.io
// @license.name	Apache 2.0
// @license.url	http://www.apache.org/licenses/LICENSE-2.0.html
//
// @host			localhost:8181
// @schemes		http https
func main() {
	defer func() {
		if redis0 != nil {
			redis0.Close()
		}

		if redis1 != nil {
			redis1.Close()
		}

		if redis2 != nil {
			redis2.Close()
		}

		if sql0 != nil {
			sql0.Close()
		}

		if scheduler != nil {
			scheduler.Stop()
		}
	}()

	app.Serve()
}
