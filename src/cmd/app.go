package main

import (
	"flag"

	"go-far/src/config"
	resthandler "go-far/src/handler/rest"
	"go-far/src/preference"
	"go-far/src/repository"
	"go-far/src/service"
	"go-far/src/util"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

var (
	minJitter    int
	maxJitter    int
	sqlClient0   *sqlx.DB
	redisClient0 *redis.Client
	redisClient1 *redis.Client
	redisClient2 *redis.Client
	app          config.App
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
	sqlClient0 = config.InitDB(log, conf.Postgres)

	// Redis Initialization
	redisClient0 = config.InitRedis(log, conf.Redis, preference.REDIS_APPS)
	redisClient1 = config.InitRedis(log, conf.Redis, preference.REDIS_AUTH)
	redisClient2 = config.InitRedis(log, conf.Redis, preference.REDIS_LIMITER)

	// Query Loader Initialization
	queryLoader, err := config.InitQueryLoader(log, conf.Queries)
	if err != nil {
		log.Panic().Err(err).Msg("Failed to load queries")
	}

	// Initialize dependencies
	repository := repository.InitRepository(sqlClient0, redisClient0, queryLoader, conf.Redis.CacheTTL)
	service := service.InitService(repository)

	// Initialize validator
	util.Validator()

	// Auth Initialization
	auth := config.InitAuth(log, conf.Auth, redisClient1)

	// Middleware Initialization
	middleware := config.InitMiddleware(log, auth, redisClient2)

	// HTTP Gin Initialization
	httpGin := config.InitHttpGin(log, middleware)

	// REST Handler Initialization
	resthandler.InitRestHandler(httpGin, auth, middleware, service)

	// HTTP Server Initialization
	httpServer := config.InitHttpServer(log, conf.Server, httpGin)

	// App Initialization
	app = config.InitGrace(log, httpServer)
}

// @title			Go Far
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
// @BasePath		/api/v1
// @schemes		http https
func main() {
	defer func() {
		if redisClient0 != nil {
			redisClient0.Close()
		}

		if redisClient1 != nil {
			redisClient1.Close()
		}

		if redisClient2 != nil {
			redisClient2.Close()
		}

		if sqlClient0 != nil {
			sqlClient0.Close()
		}
	}()

	app.Serve()
}
