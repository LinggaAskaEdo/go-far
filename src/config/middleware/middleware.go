package middleware

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"go-far/src/config/token"
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
	RoleLimiter() func(http.Handler) http.Handler
}

type middleware struct {
	log           zerolog.Logger
	tkn           token.Token
	opt           MiddlewareOptions
	rdb           *redis.Client
	limit         int
	period        time.Duration
	roleRateLimit RoleRateLimitOptions
	publicPaths   []string
}

// MiddlewareOptions holds middleware configuration
type MiddlewareOptions struct {
	PublicPaths   []string             `yaml:"public_paths"`
	RateLimiter   RateLimiterOptions   `yaml:"rate_limiter"`
	RoleRateLimit RoleRateLimitOptions `yaml:"role_rate_limit"`
}

// RateLimiterOptions holds rate limiter configuration
type RateLimiterOptions struct {
	Command string `yaml:"command"`
	Limit   int    `yaml:"limit"`
}

// RoleRateLimitOptions holds role-based rate limiter configuration
type RoleRateLimitOptions struct {
	Admin RoleRateLimit `yaml:"admin"`
	User  RoleRateLimit `yaml:"user"`
	Guest RoleRateLimit `yaml:"guest"`
}

// RoleRateLimit holds rate limit config for a single role
type RoleRateLimit struct {
	Command string `yaml:"command"`
	Limit   int    `yaml:"limit"`
}

// InitMiddleware initializes the middleware
func InitMiddleware(log zerolog.Logger, opt MiddlewareOptions, tkn token.Token, rdb *redis.Client) Middleware {
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
			log:           log,
			opt:           opt,
			tkn:           tkn,
			rdb:           rdb,
			limit:         limit,
			period:        period,
			roleRateLimit: opt.RoleRateLimit,
			publicPaths:   opt.PublicPaths,
		}
	})

	return middlewareInst
}
