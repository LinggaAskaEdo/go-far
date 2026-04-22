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
	Scheduler  cfgscheduler.SchedulerOptions `yaml:"scheduler"`
	Tracer     tracer.TracerOptions          `yaml:"tracer"`
	Queries    query.QueriesOptions          `yaml:"queries"`
	HTTP       server.HttpOptions            `yaml:"http"`
	Middleware middleware.MiddlewareOptions  `yaml:"middleware"`
	Server     server.ServerOptions          `yaml:"server"`
	Postgres   database.DatabaseOptions      `yaml:"postgres"`
	MySQL      database.DatabaseOptions      `yaml:"mysql"`
	Logger     logger.LoggerOptions          `yaml:"logger"`
	Redis      redis.RedisOptions            `yaml:"redis"`
	HttpClient httpclient.HttpClientOptions  `yaml:"http_client"`
	Token      token.TokenOptions            `yaml:"token"`
}

func InitConfig() (*Config, error) {
	var cfg Config
	if err := envyaml.LoadConfig("config.yaml", &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
