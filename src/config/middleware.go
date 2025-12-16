package config

import (
	"context"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"go-far/src/preference"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/rs/xid"
	"github.com/rs/zerolog"
)

var onceMiddlewre = &sync.Once{}

type Middleware interface {
	Handler() gin.HandlerFunc
	CORS() gin.HandlerFunc
	// Limiter(command string, limit int) gin.HandlerFunc
	// JWT() gin.HandlerFunc
	// KC() gin.HandlerFunc
}

type middleware struct {
	log  zerolog.Logger
	auth Auth
	// limit     int
	// deadline  int64
	// shaScript map[string]string
	// period    time.Duration
	rdb *redis.Client
}

type MiddlewareOptions struct {
	RateLimiter RateLimiterOptions `yaml:"rate_limiter"`
}

type RateLimiterOptions struct {
	Enabled          bool                    `yaml:"enabled"`
	GlobalLimiter    GlobalLimiterOptions    `yaml:"global"`
	EndpointsLimiter EndpointsLimiterOptions `yaml:"endpoints"`
}

type GlobalLimiterOptions struct {
	Limit  int           `yaml:"limit"`
	Window time.Duration `yaml:"window"`
}

type EndpointsLimiterOptions struct {
	CreateUser CreateUserOptions `yaml:"create_user"`
	ListUser   ListUserOptions   `yaml:"list_user"`
}

type CreateUserOptions struct {
	Limit  int           `yaml:"limit"`
	Window time.Duration `yaml:"window"`
}

type ListUserOptions struct {
	Limit  int           `yaml:"limit"`
	Window time.Duration `yaml:"window"`
}

func InitMiddleware(log zerolog.Logger, auth Auth, rdb *redis.Client) Middleware {
	var m *middleware

	onceMiddlewre.Do(func() {
		m = &middleware{
			log:  log,
			auth: auth,
			rdb:  rdb,
		}
	})

	return m
}

func (mw *middleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		if !strings.HasPrefix(path, "/swagger/") { // skip logging swagger request
			start := time.Now()

			ctx := c.Request.Context()
			ctx = mw.attachReqID(ctx)
			ctx = mw.attachLogger(ctx)

			raw := c.Request.URL.RawQuery
			if raw != "" {
				path = path + "?" + raw
			}

			mw.log.Info().
				Str(preference.EVENT, "START").
				Str(string(preference.CONTEXT_KEY_LOG_REQUEST_ID), mw.getRequestID(ctx)).
				Str(preference.METHOD, c.Request.Method).
				Str(preference.URL, path).
				Str(preference.USER_AGENT, c.Request.UserAgent()).
				Str(preference.ADDR, c.Request.Host).
				Send()

			// Process request
			c.Request = c.Request.WithContext(ctx)
			c.Next()

			// Fill the params
			param := gin.LogFormatterParams{}

			param.TimeStamp = time.Now() // Stop timer
			param.Latency = param.TimeStamp.Sub(start)
			if param.Latency > time.Minute {
				param.Latency = param.Latency.Truncate(time.Second)
			}

			param.StatusCode = c.Writer.Status()

			mw.log.Info().
				Str(preference.EVENT, "END").
				Str(string(preference.CONTEXT_KEY_LOG_REQUEST_ID), mw.getRequestID(ctx)).
				Str(preference.LATENCY, param.Latency.String()).
				Int(preference.STATUS, param.StatusCode).
				Send()
		}
	}
}

func (mw *middleware) attachReqID(ctx context.Context) context.Context {
	return context.WithValue(ctx, preference.CONTEXT_KEY_REQUEST_ID, xid.New().String())
}

func (mw *middleware) attachLogger(ctx context.Context) context.Context {
	return mw.log.With().Str(string(preference.CONTEXT_KEY_LOG_REQUEST_ID), mw.getRequestID(ctx)).Logger().WithContext(ctx)
}

func (mw *middleware) getRequestID(ctx context.Context) string {
	reqID := ctx.Value(preference.CONTEXT_KEY_REQUEST_ID)

	if ret, ok := reqID.(string); ok {
		return ret
	}

	return ""
}

func (mw *middleware) CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		strMethods := []string{"GET", "POST", "PUT", "DELETE"}

		c.Header("Access-Control-Allow-Headers:", "*")
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", strings.Join(strMethods, ", "))
		c.Header("X-Frame-Options", "DENY")
		c.Header("Content-Security-Policy", "default-src 'self'; connect-src *; font-src *; script-src-elem * 'unsafe-inline'; img-src * data:; style-src * 'unsafe-inline';")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		c.Header("Referrer-Policy", "strict-origin")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Permissions-Policy", "geolocation=(),midi=(),sync-xhr=(),microphone=(),camera=(),magnetometer=(),gyroscope=(),fullscreen=(self),payment=()")

		if !slices.Contains(strMethods, c.Request.Method) {
			c.AbortWithStatus(http.StatusMethodNotAllowed)
			return
		}

		c.Next()
	}
}
