package main

import (
	"fmt"
	"os"

	"go-far/src/config/auth"
	"go-far/src/config/database"
	"go-far/src/config/logger"
	"go-far/src/config/middleware"
	"go-far/src/config/query"
	"go-far/src/config/redis"
	cfgscheduler "go-far/src/config/scheduler"
	"go-far/src/config/server"
	"go-far/src/config/tracer"

	"github.com/goccy/go-yaml"
)

type Config struct {
	Server     server.ServerOptions          `yaml:"server"`
	Logger     logger.LoggerOptions          `yaml:"logger"`
	Postgres   database.DatabaseOptions      `yaml:"postgres"`
	MySQL      database.DatabaseOptions      `yaml:"mysql"`
	Redis      redis.RedisOptions            `yaml:"redis"`
	Queries    query.QueriesOptions          `yaml:"queries"`
	Auth       auth.AuthOptions              `yaml:"auth"`
	Middleware middleware.MiddlewareOptions  `yaml:"middleware"`
	Gin        server.GinOptions             `yaml:"gin"`
	Scheduler  cfgscheduler.SchedulerOptions `yaml:"scheduler"`
	Tracer     tracer.TracerOptions          `yaml:"tracer"`
}

func InitConfig() (*Config, error) {
	configPath := "config.yaml"

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Override with environment variables if present
	overrideWithEnv(&cfg)

	return &cfg, nil
}

func overrideWithEnv(cfg *Config) {
	if val := os.Getenv("SERVER_PORT"); val != "" {
		cfg.Server.Port = parseInt(val, cfg.Server.Port)
	}

	if val := os.Getenv("LOG_LEVEL"); val != "" {
		cfg.Logger.Level = val
	}

	if val := os.Getenv("DB_HOST"); val != "" {
		cfg.Postgres.Host = val
	}

	if val := os.Getenv("DB_PORT"); val != "" {
		cfg.Postgres.Port = parseInt(val, cfg.Postgres.Port)
	}

	if val := os.Getenv("DB_USER"); val != "" {
		cfg.Postgres.User = val
	}

	if val := os.Getenv("DB_PASSWORD"); val != "" {
		cfg.Postgres.Password = val
	}

	if val := os.Getenv("DB_NAME"); val != "" {
		cfg.Postgres.DBName = val
	}

	if val := os.Getenv("REDIS_ADDRESS"); val != "" {
		cfg.Redis.Address = val
	}

	if val := os.Getenv("REDIS_PASSWORD"); val != "" {
		cfg.Redis.Password = val
	}

	if val := os.Getenv("TRACER_ENDPOINT"); val != "" {
		cfg.Tracer.Endpoint = val
	}
}

func parseInt(s string, defaultVal int) int {
	var val int
	if _, err := fmt.Sscanf(s, "%d", &val); err == nil {
		return val
	}

	return defaultVal
}
