package main

import (
	"fmt"
	"os"

	"go-far/src/config"

	"github.com/goccy/go-yaml"
)

type Config struct {
	Server     config.ServerOptions     `yaml:"server"`
	Logger     config.LoggerOptions     `yaml:"logger"`
	Postgres   config.DatabaseOptions   `yaml:"postgres"`
	MySQL      config.DatabaseOptions   `yaml:"mysql"`
	Redis      config.RedisOptions      `yaml:"redis"`
	Queries    config.QueriesOptions    `yaml:"queries"`
	Auth       config.AuthOptions       `yaml:"auth"`
	Middleware config.MiddlewareOptions `yaml:"middleware"`
	Scheduler  config.SchedulerOptions  `yaml:"scheduler"`
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

	if val := os.Getenv("POSTGRES_HOST"); val != "" {
		cfg.Postgres.Host = val
	}

	if val := os.Getenv("POSTGRES_PORT"); val != "" {
		cfg.Postgres.Port = parseInt(val, cfg.Postgres.Port)
	}

	if val := os.Getenv("POSTGRES_USER"); val != "" {
		cfg.Postgres.User = val
	}

	if val := os.Getenv("POSTGRES_PASSWORD"); val != "" {
		cfg.Postgres.Password = val
	}

	if val := os.Getenv("POSTGRES_DB_NAME"); val != "" {
		cfg.Postgres.DBName = val
	}

	if val := os.Getenv("REDIS_ADDRESS"); val != "" {
		cfg.Redis.Address = val
	}

	if val := os.Getenv("REDIS_PASSWORD"); val != "" {
		cfg.Redis.Password = val
	}
}

func parseInt(s string, defaultVal int) int {
	var val int
	if _, err := fmt.Sscanf(s, "%d", &val); err == nil {
		return val
	}

	return defaultVal
}
