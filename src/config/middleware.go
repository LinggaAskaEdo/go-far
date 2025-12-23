package config

import (
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"

	"go-far/src/preference"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

const TimeFormat = time.RFC3339

var (
	onceMiddlewre = &sync.Once{}

	timeDict = map[string]time.Duration{
		"S": time.Second,
		"M": time.Minute,
		"H": time.Hour,
		"D": time.Hour * 24,
	}
)

type Middleware interface {
	Handler() gin.HandlerFunc
	CORS() gin.HandlerFunc
	Limiter(command string, limit int) gin.HandlerFunc
}

type middleware struct {
	log    zerolog.Logger
	auth   Auth
	opt    MiddlewareOptions
	rdb    *redis.Client
	limit  int
	period time.Duration
}

type MiddlewareOptions struct {
	RateLimiter RateLimiterOptions `yaml:"rate_limiter"`
}

type RateLimiterOptions struct {
	Command string `yaml:"command"`
	Limit   int    `yaml:"limit"`
}

func InitMiddleware(log zerolog.Logger, opt MiddlewareOptions, auth Auth, rdb *redis.Client) Middleware {
	var m *middleware

	onceMiddlewre.Do(func() {
		var limit int
		var period time.Duration

		limit = opt.RateLimiter.Limit

		values := strings.Split(opt.RateLimiter.Command, "-")
		if len(values) != 2 {
			log.Panic().Err(errors.New(preference.FormatError)).Send()
		}

		unit, err := strconv.Atoi(values[0])
		if err != nil {
			log.Panic().Err(errors.New(preference.FormatError)).Send()
		}

		if t, ok := timeDict[strings.ToUpper(values[1])]; ok {
			period = time.Duration(unit) * t
		} else {
			log.Panic().Err(errors.New(preference.FormatError)).Send()
		}

		m = &middleware{
			log:    log,
			opt:    opt,
			auth:   auth,
			rdb:    rdb,
			limit:  limit,
			period: period,
		}
	})

	return m
}
