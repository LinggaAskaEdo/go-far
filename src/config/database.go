package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"go-far/src/preference"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
)

type DatabaseOptions struct {
	Enabled         bool          `yaml:"enabled"`
	Driver          string        `yaml:"driver"`
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	User            string        `yaml:"user"`
	Password        string        `yaml:"password"`
	DBName          string        `yaml:"dbname"`
	SSLMode         bool          `yaml:"sslmode"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time"`
}

func InitDB(log zerolog.Logger, opt DatabaseOptions) *sqlx.DB {
	if !opt.Enabled {
		return nil
	}

	driver, host, err := getURI(opt)
	if err != nil {
		log.Panic().Err(err).Msg(fmt.Sprintf("%s status: FAILED", strings.ToUpper(opt.Driver)))
	}

	db, err := sqlx.Connect(driver, host)
	if err != nil {
		log.Panic().Err(err).Msg(fmt.Sprintf("%s status: FAILED", strings.ToUpper(opt.Driver)))
	}

	log.Debug().Msg(fmt.Sprintf("%s status: OK", strings.ToUpper(opt.Driver)))

	db.SetMaxOpenConns(opt.MaxOpenConns)
	db.SetMaxIdleConns(opt.MaxIdleConns)
	db.SetConnMaxLifetime(opt.ConnMaxLifetime)
	db.SetConnMaxIdleTime(opt.ConnMaxIdleTime)

	return db
}

func getURI(opt DatabaseOptions) (string, string, error) {
	switch opt.Driver {
	case preference.POSTGRES:
		ssl := `disable`
		if opt.SSLMode {
			ssl = `require`
		}

		return opt.Driver, fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", opt.Host, opt.Port, opt.User, opt.Password, opt.DBName, ssl), nil

	case preference.MYSQL:
		ssl := `false`
		if opt.SSLMode {
			ssl = `true`
		}

		return opt.Driver, fmt.Sprintf("%s:%s@tcp(%s:%v)/%s?tls=%s&parseTime=%t", opt.User, opt.Password, opt.Host, opt.Port, opt.DBName, ssl, true), nil

	default:
		return "", "", errors.New("DB Driver is not supported ")
	}
}
