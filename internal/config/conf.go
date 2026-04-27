package config

import (
	"go-far/internal/infra/database"
	app "go-far/internal/infra/grace"
	httpclient "go-far/internal/infra/http/client"
	httpserver "go-far/internal/infra/http/server"
	"go-far/internal/infra/logger"
	"go-far/internal/infra/metrics"
	"go-far/internal/infra/middleware"
	"go-far/internal/infra/query"
	"go-far/internal/infra/redis"
	cfgscheduler "go-far/internal/infra/scheduler"
	"go-far/internal/infra/token"
	"go-far/internal/infra/tracer"

	"github.com/yuseferi/envyaml"
)

type Config struct {
	App        *app.AppOptions                `yaml:"app"`
	HTTP       HTTPConfig                     `yaml:"http"`
	Database   DatabaseConfig                 `yaml:"database"`
	Redis      *redis.RedisOptions            `yaml:"redis"`
	Logger     *logger.LoggerOptions          `yaml:"logger"`
	Middleware *middleware.MiddlewareOptions  `yaml:"middleware"`
	Scheduler  *cfgscheduler.SchedulerOptions `yaml:"scheduler"`
	Token      *token.TokenOptions            `yaml:"token"`
	Queries    *query.QueriesOptions          `yaml:"queries"`
	Tracer     *tracer.TracerOptions          `yaml:"tracer"`
	Metric     *metrics.MetricsOptions        `yaml:"metric"`
}

type HTTPConfig struct {
	Server *httpserver.HttpServerOptions `yaml:"server"`
	Client *httpclient.HttpClientOptions `yaml:"client"`
}

type DatabaseConfig struct {
	Postgres *database.DatabaseOptions `yaml:"postgres"`
}

func InitConfig() (*Config, error) {
	var cfg Config
	if err := envyaml.LoadConfig("./configs/config.yaml", &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
