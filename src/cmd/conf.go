package main

import (
	"os"
	"regexp"
	"strconv"
	"strings"

	"go-far/src/config/database"
	"go-far/src/config/logger"
	"go-far/src/config/middleware"
	"go-far/src/config/query"
	"go-far/src/config/redis"
	cfgscheduler "go-far/src/config/scheduler"
	"go-far/src/config/server"
	"go-far/src/config/token"
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
	Token      token.TokenOptions            `yaml:"token"`
	Middleware middleware.MiddlewareOptions  `yaml:"middleware"`
	HTTP       server.HttpOptions            `yaml:"http"`
	Scheduler  cfgscheduler.SchedulerOptions `yaml:"scheduler"`
	Tracer     tracer.TracerOptions          `yaml:"tracer"`
}

var envVarRegex = regexp.MustCompile(`\$\{(\w+)\}`)

// loadEnvFile reads a key=value file and sets each as an environment variable
func loadEnvFile(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if key, val, ok := strings.Cut(line, "="); ok {
			os.Setenv(strings.TrimSpace(key), strings.TrimSpace(val))
		}
	}
}

func InitConfig() (*Config, error) {
	loadEnvFile("/etc/environment")
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(resolveEnvVars(data), &cfg); err != nil {
		return nil, err
	}

	overrideWithEnv(&cfg)

	return &cfg, nil
}

func resolveEnvVars(content []byte) []byte {
	return envVarRegex.ReplaceAllFunc(content, func(match []byte) []byte {
		varName := string(match[2 : len(match)-1])
		if val := os.Getenv(varName); val != "" {
			return []byte(val)
		}

		return match
	})
}

func overrideWithEnv(cfg *Config) {
	if v := os.Getenv("SERVER_PORT"); v != "" {
		cfg.Server.Port = parseInt(v, cfg.Server.Port)
	}
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		cfg.Logger.Level = v
	}

	// Postgres
	if v := os.Getenv("DB_HOST"); v != "" {
		cfg.Postgres.Host = v
	}
	if v := os.Getenv("DB_PORT"); v != "" {
		cfg.Postgres.Port = parseInt(v, cfg.Postgres.Port)
	}
	if v := os.Getenv("DB_USER"); v != "" {
		cfg.Postgres.User = v
	}
	if v := os.Getenv("DB_NAME"); v != "" {
		cfg.Postgres.DBName = v
	}

	// MySQL
	if v := os.Getenv("MYSQL_HOST"); v != "" {
		cfg.MySQL.Host = v
	}
	if v := os.Getenv("MYSQL_PORT"); v != "" {
		cfg.MySQL.Port = parseInt(v, cfg.MySQL.Port)
	}
	if v := os.Getenv("MYSQL_USER"); v != "" {
		cfg.MySQL.User = v
	}
	if v := os.Getenv("MYSQL_DB_NAME"); v != "" {
		cfg.MySQL.DBName = v
	}

	// Redis
	if v := os.Getenv("REDIS_ADDRESS"); v != "" {
		cfg.Redis.Address = v
	}
	if v := os.Getenv("REDIS_PASSWORD"); v != "" {
		cfg.Redis.Password = v
	}

	// Tracer
	if v := os.Getenv("TRACER_ENDPOINT"); v != "" {
		cfg.Tracer.Endpoint = v
	}
}

func parseInt(s string, defaultVal int) int {
	if v, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
		return v
	}

	return defaultVal
}
