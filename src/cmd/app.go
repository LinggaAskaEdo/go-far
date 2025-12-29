package main

import (
	"flag"

	_ "go-far/docs"
	"go-far/src/config"
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
	scheduler *config.Scheduler

	tracer config.Tracer
	app    config.App
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
	log := config.InitLogger(conf.Logger)

	// SQL Initialization
	sql0 = config.InitDB(log, conf.Postgres)

	// Redis Initialization
	redis0 = config.InitRedis(log, conf.Redis, preference.REDIS_APPS)
	redis1 = config.InitRedis(log, conf.Redis, preference.REDIS_AUTH)
	redis2 = config.InitRedis(log, conf.Redis, preference.REDIS_LIMITER)

	// Query Loader Initialization
	queryLoader := config.InitQueryLoader(log, conf.Queries)

	// Initialize dependencies
	repository := repository.InitRepository(sql0, redis0, queryLoader, conf.Redis.CacheTTL)
	service := service.InitService(repository)

	// Initialize validator
	config.InitValidator(log)

	// Auth Initialization
	auth := config.InitAuth(log, conf.Auth, redis1)

	// Middleware Initialization
	middleware := config.InitMiddleware(log, conf.Middleware, auth, redis2)

	// HTTP Gin Initialization
	httpGin := config.InitHttpGin(log, middleware, conf.Gin)

	// REST Handler Initialization
	restHandler.InitRestHandler(httpGin, auth, middleware, service)

	//Scheduler Initialization
	scheduler = config.InitScheduler(log, conf.Scheduler)
	schedHandler.InitSchedulerHandler(log, scheduler, service, conf.Scheduler.SchedulerJobs)

	// HTTP Server Initialization
	httpServer := config.InitHttpServer(log, conf.Server, httpGin)

	// Tracer Initialization
	tracer = config.InitTracer(log)

	// App Initialization
	app = config.InitGrace(log, httpServer, tracer)
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
