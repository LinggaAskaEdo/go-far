package middleware

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"go-far/src/config/auth"
	"go-far/src/preference"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

const TimeFormat = time.RFC3339

var (
	onceMiddleware = &sync.Once{}
	middlewareInst Middleware

	timeDict = map[string]time.Duration{
		"S": time.Second,
		"M": time.Minute,
		"H": time.Hour,
		"D": time.Hour * 24,
	}
)

// Middleware defines the middleware interface
type Middleware interface {
	Handler() func(http.Handler) http.Handler
	CORS() func(http.Handler) http.Handler
	Limiter(command string, limit int) func(http.Handler) http.Handler
}

type middleware struct {
	log    zerolog.Logger
	auth   auth.Auth
	opt    MiddlewareOptions
	rdb    *redis.Client
	limit  int
	period time.Duration
}

// MiddlewareOptions holds middleware configuration
type MiddlewareOptions struct {
	RateLimiter RateLimiterOptions `yaml:"rate_limiter"`
}

// RateLimiterOptions holds rate limiter configuration
type RateLimiterOptions struct {
	Command string `yaml:"command"`
	Limit   int    `yaml:"limit"`
}

// InitMiddleware initializes the middleware
func InitMiddleware(log zerolog.Logger, opt MiddlewareOptions, auth auth.Auth, rdb *redis.Client) Middleware {
	onceMiddleware.Do(func() {
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

		middlewareInst = &middleware{
			log:    log,
			opt:    opt,
			auth:   auth,
			rdb:    rdb,
			limit:  limit,
			period: period,
		}
	})

	return middlewareInst
}
