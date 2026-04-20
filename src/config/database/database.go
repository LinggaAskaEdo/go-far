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
	Enabled         bool          `yaml:"enabled"`
	Driver          string        `yaml:"driver"`
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	User            string        `yaml:"user"`
	Password        string        `yaml:"password" env:"POSTGRES_DOCKER_PASSWORD"`
	DBName          string        `yaml:"dbname"`
	SSLMode         bool          `yaml:"sslmode"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time"`
}

func InitDB(log zerolog.Logger, opt DatabaseOptions) *pgxpool.Pool {
	if !opt.Enabled {
		return nil
	}

	config, err := pgxpool.ParseConfig(getURI(opt))
	if err != nil {
		log.Panic().Err(err).Msg("failed to parse database config")
	}

	config.MaxConns = int32(opt.MaxOpenConns)
	config.MinConns = int32(opt.MaxIdleConns)
	config.MaxConnLifetime = opt.ConnMaxLifetime
	config.MaxConnIdleTime = opt.ConnMaxIdleTime

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Panic().Err(err).Msg(fmt.Sprintf("%s status: FAILED", strings.ToUpper(opt.Driver)))
	}

	log.Debug().Msg(fmt.Sprintf("%s status: OK", strings.ToUpper(opt.Driver)))

	return pool
}

func getURI(opt DatabaseOptions) string {
	ssl := "disable"
	if opt.SSLMode {
		ssl = "require"
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		opt.User, opt.Password, opt.Host, opt.Port, opt.DBName, ssl)
}
