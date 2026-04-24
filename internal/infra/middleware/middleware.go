package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"go-far/internal/infra/token"
	appErr "go-far/internal/model/errors"
	"go-far/internal/preference"

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
	AuthLimiter() func(http.Handler) http.Handler
}

type middleware struct {
	tkn            token.Token
	log            *zerolog.Logger
	opt            *MiddlewareOptions
	rdb            *redis.Client
	publicPaths    map[string]bool
	roleRateLimit  RoleRateLimitOptions
	authRateLimit  AuthRateLimitOptions
	limit          int
	period         time.Duration
	authLimit      int
	authPeriod     time.Duration
	tracingEnabled bool
}

// MiddlewareOptions holds middleware configuration
type MiddlewareOptions struct {
	PublicPaths   []string             `yaml:"public_paths"`
	RateLimiter   RateLimiterOptions   `yaml:"rate_limiter"`
	RoleRateLimit RoleRateLimitOptions `yaml:"role_rate_limit"`
	AuthRateLimit AuthRateLimitOptions `yaml:"auth_rate_limit"`
}

// RateLimiterOptions holds rate limiter configuration
type RateLimiterOptions struct {
	Command string `yaml:"command"`
	Limit   int    `yaml:"limit"`
}

// AuthRateLimitOptions holds auth endpoint rate limit configuration
type AuthRateLimitOptions struct {
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
func InitMiddleware(log *zerolog.Logger, opt *MiddlewareOptions, tkn token.Token, rdb *redis.Client, tracingEnabled bool) Middleware {
	onceMiddleware.Do(func() {
		// --- Main rate limiter (mandatory) ---
		limit := opt.RateLimiter.Limit
		period, err := parsePeriod(opt.RateLimiter.Command)
		if err != nil {
			log.Panic().Err(err).Send()
		}

		// --- Auth rate limiter (optional, with defaults) ---
		authLimit := opt.AuthRateLimit.Limit
		authPeriod := time.Minute // default
		if opt.AuthRateLimit.Command != "" {
			if p, err := parsePeriod(opt.AuthRateLimit.Command); err == nil {
				authPeriod = p
			}
		}

		if authLimit == 0 {
			authLimit = 3
		}

		publicPathsMap := make(map[string]bool, len(opt.PublicPaths))
		for _, p := range opt.PublicPaths {
			publicPathsMap[p] = true
		}

		middlewareInst = &middleware{
			log:            log,
			opt:            opt,
			tkn:            tkn,
			rdb:            rdb,
			limit:          limit,
			period:         period,
			roleRateLimit:  opt.RoleRateLimit,
			authRateLimit:  opt.AuthRateLimit,
			authLimit:      authLimit,
			authPeriod:     authPeriod,
			publicPaths:    publicPathsMap,
			tracingEnabled: tracingEnabled,
		}
	})

	return middlewareInst
}

// parsePeriod converts a command string like "10-minute" into a time.Duration.
// Returns an error if the format is invalid or the unit is unrecognized.
func parsePeriod(command string) (time.Duration, error) {
	parts := strings.Split(command, "-")
	if len(parts) != 2 {
		return 0, appErr.New(preference.FormatError, appErr.CodeHTTPBadRequest)
	}

	unit, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, appErr.New("invalid rate limit format", appErr.CodeHTTPBadRequest)
	}

	unitKey := strings.ToUpper(parts[1])
	if t, ok := timeDict[unitKey]; ok {
		return time.Duration(unit) * t, nil
	}

	return 0, appErr.New(preference.FormatError, appErr.CodeHTTPBadRequest)
}
