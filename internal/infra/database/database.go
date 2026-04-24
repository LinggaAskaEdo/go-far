package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type DatabaseOptions struct {
	Driver          string        `yaml:"driver"`
	Host            string        `yaml:"host"`
	User            string        `yaml:"user"`
	Password        string        `yaml:"password" env:"POSTGRES_DOCKER_PASSWORD"`
	DBName          string        `yaml:"dbname"`
	Port            int           `yaml:"port"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time"`
	MaxOpenConns    int32         `yaml:"max_open_conns"`
	MaxIdleConns    int32         `yaml:"max_idle_conns"`
	Enabled         bool          `yaml:"enabled"`
	SSLMode         bool          `yaml:"sslmode"`
}

func InitDB(log *zerolog.Logger, opt *DatabaseOptions) *pgxpool.Pool {
	if !opt.Enabled {
		return nil
	}

	config, err := pgxpool.ParseConfig(getURI(opt))
	if err != nil {
		log.Panic().Err(err).Msg("failed to parse database config")
	}

	config.MaxConns = opt.MaxOpenConns
	config.MinConns = opt.MaxIdleConns
	config.MaxConnLifetime = opt.ConnMaxLifetime
	config.MaxConnIdleTime = opt.ConnMaxIdleTime

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Panic().Err(err).Msg(strings.ToUpper(opt.Driver) + " status: FAILED")
	}

	log.Info().Msg(strings.ToUpper(opt.Driver) + " status: OK")

	return pool
}

func getURI(opt *DatabaseOptions) string {
	ssl := "disable"
	if opt.SSLMode {
		ssl = "require"
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		opt.User, opt.Password, opt.Host, opt.Port, opt.DBName, ssl)
}
