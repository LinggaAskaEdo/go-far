package main

import (
	"go-far/src/config/database"
	"go-far/src/config/httpclient"
	"go-far/src/config/logger"
	"go-far/src/config/middleware"
	"go-far/src/config/query"
	"go-far/src/config/redis"
	cfgscheduler "go-far/src/config/scheduler"
	"go-far/src/config/server"
	"go-far/src/config/token"
	"go-far/src/config/tracer"

	"github.com/yuseferi/envyaml"
)

type Config struct {
	Server     server.ServerOptions          `yaml:"server"`
	Logger     logger.LoggerOptions          `yaml:"logger"`
	Postgres   database.DatabaseOptions      `yaml:"postgres"`
	MySQL      database.DatabaseOptions      `yaml:"mysql"`
	Redis      redis.RedisOptions            `yaml:"redis"`
	Queries    query.QueriesOptions          `yaml:"queries"`
	Token      token.TokenOptions            `yaml:"token"`
	Middleware middleware.MiddlewareOptions  `yaml:"middleware"`
	HTTP       server.HttpOptions            `yaml:"http"`
	Scheduler  cfgscheduler.SchedulerOptions `yaml:"scheduler"`
	Tracer     tracer.TracerOptions          `yaml:"tracer"`
	HttpClient httpclient.HttpClientOptions  `yaml:"http_client"`
}

func InitConfig() (*Config, error) {
	var cfg Config
	if err := envyaml.LoadConfig("config.yaml", &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
