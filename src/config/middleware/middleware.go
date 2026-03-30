package middleware

import (
	"context"
	"errors"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"go-far/src/config/auth"
	"go-far/src/preference"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/rs/xid"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
)

const TimeFormat = time.RFC3339

// Lua script for atomic rate limiting (reduces Redis round-trips from 6 to 2)
const rateLimitLuaScript = `
	local routeKey = KEYS[1]
	local globalKey = KEYS[2]
	local routeLimit = tonumber(ARGV[1])
	local globalLimit = tonumber(ARGV[2])
	local routeDuration = tonumber(ARGV[3])
	local globalDuration = tonumber(ARGV[4])

	-- Check and increment route limit
	local routeCount = tonumber(redis.call('INCR', routeKey))
	if routeCount == 1 then
		redis.call('EXPIRE', routeKey, routeDuration)
	end

	-- Check route limit
	if routeCount > routeLimit then
		local routeTTL = redis.call('TTL', routeKey)
		return {0, routeCount, 0, routeTTL, 'route'}
	end

	-- Check and increment global limit
	local globalCount = tonumber(redis.call('INCR', globalKey))
	if globalCount == 1 then
		redis.call('EXPIRE', globalKey, globalDuration)
	end

	-- Check global limit
	if globalCount > globalLimit then
		local globalTTL = redis.call('TTL', globalKey)
		return {0, globalCount, 0, globalTTL, 'global'}
	end

	-- Success: return counts and TTLs
	local routeTTL = redis.call('TTL', routeKey)
	local globalTTL = redis.call('TTL', globalKey)
	return {1, routeCount, globalCount, routeTTL, globalTTL}
`

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
	Handler() gin.HandlerFunc
	CORS() gin.HandlerFunc
	Limiter(command string, limit int) gin.HandlerFunc
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

// Handler returns the main middleware handler
func (mw *middleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		path := c.Request.URL.Path

		if !strings.HasPrefix(path, "/swagger/") {
			start := time.Now()

			span := trace.SpanFromContext(ctx)
			spanContext := span.SpanContext()
			traceID := spanContext.TraceID().String()
			spanID := spanContext.SpanID().String()

			reqID := c.GetHeader("X-Request-ID")
			if reqID == "" {
				spanID = xid.New().String()
			}

			ctx = mw.attachTraceSpanIDs(ctx, traceID, spanID)
			ctx = mw.attachLogger(ctx)

			c.Header("X-Request-ID", spanID)

			raw := c.Request.URL.RawQuery
			if raw != "" {
				path = path + "?" + raw
			}

			mw.log.Info().
				Str(preference.EVENT, "START").
				Str("trace_id", traceID).
				Str("span_id", spanID).
				Str(preference.METHOD, c.Request.Method).
				Str(preference.URL, path).
				Str(preference.USER_AGENT, c.Request.UserAgent()).
				Str(preference.ADDR, c.Request.Host).
				Send()

			c.Request = c.Request.WithContext(ctx)
			c.Next()

			param := gin.LogFormatterParams{}
			param.TimeStamp = time.Now()
			param.Latency = param.TimeStamp.Sub(start)
			if param.Latency > time.Minute {
				param.Latency = param.Latency.Truncate(time.Second)
			}
			param.StatusCode = c.Writer.Status()

			mw.log.Info().
				Str(preference.EVENT, "END").
				Str("trace_id", traceID).
				Str("span_id", spanID).
				Str(preference.LATENCY, param.Latency.String()).
				Int(preference.STATUS, param.StatusCode).
				Send()
		}
	}
}

// CORS returns the CORS middleware handler
func (mw *middleware) CORS() gin.HandlerFunc {
	allowedOrigins := getAllowedOrigins()
	strMethods := []string{"GET", "POST", "PUT", "DELETE"}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		if isOriginAllowed(origin, allowedOrigins) {
			c.Header("Access-Control-Allow-Origin", origin)
		} else {
			c.Header("Access-Control-Allow-Origin", "*")
		}

		c.Header("Access-Control-Allow-Headers", "*")
		c.Header("Access-Control-Allow-Methods", strings.Join(strMethods, ", "))
		c.Header("X-Frame-Options", "DENY")
		c.Header("Content-Security-Policy", "default-src 'self'; connect-src *; font-src *; script-src-elem * 'unsafe-inline'; img-src * data:; style-src * 'unsafe-inline';")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		c.Header("Referrer-Policy", "strict-origin")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Permissions-Policy", "geolocation=(),midi=(),sync-xhr=(),microphone=(),camera=(),magnetometer=(),gyroscope=(),fullscreen=(self),payment=()")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		if !slices.Contains(strMethods, c.Request.Method) {
			c.AbortWithStatus(http.StatusMethodNotAllowed)
			return
		}

		c.Next()
	}
}

// Limiter returns the rate limiting middleware handler
func (mw *middleware) Limiter(command string, limit int) gin.HandlerFunc {
	return func(c *gin.Context) {
		duration, err := mw.parseCommand(command)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		now := time.Now()
		clientIP := c.ClientIP()

		routeKey := "ratelimit:route:" + c.FullPath() + ":" + c.Request.Method + ":" + clientIP
		globalKey := "ratelimit:global:" + clientIP

		ctx := context.Background()

		result, err := mw.rdb.Eval(ctx, rateLimitLuaScript, []string{routeKey, globalKey},
			limit,                    // route limit
			mw.limit,                 // global limit
			int(duration.Seconds()),  // route duration
			int(mw.period.Seconds()), // global duration
		).Result()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		resultArr, ok := result.([]interface{})
		if !ok || len(resultArr) < 5 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid rate limit response"})
			c.Abort()
			return
		}

		success := resultArr[0].(int64) == 1
		count1 := resultArr[1].(int64)
		count2 := resultArr[2].(int64)
		ttl1 := resultArr[3].(int64)
		ttl2 := resultArr[4].(int64)

		var resetTime string
		if success {
			resetTime = now.Add(time.Duration(ttl1) * time.Second).Format(TimeFormat)
		} else {
			exceededType := resultArr[5].(string)
			if exceededType == "route" {
				c.Header("X-RateLimit-Limit-route", strconv.Itoa(limit))
				c.Header("X-RateLimit-Remaining-route", "0")
				c.Header("X-RateLimit-Reset-route", now.Add(time.Duration(ttl1)*time.Second).Format(TimeFormat))
				c.JSON(http.StatusTooManyRequests, gin.H{"error": "route rate limit exceeded"})
				c.Abort()
				return
			} else {
				resetTime = now.Add(time.Duration(ttl2) * time.Second).Format(TimeFormat)
				c.Header("X-RateLimit-Limit-global", strconv.Itoa(mw.limit))
				c.Header("X-RateLimit-Remaining-global", "0")
				c.Header("X-RateLimit-Reset-global", resetTime)
				c.JSON(http.StatusTooManyRequests, gin.H{"error": "global rate limit exceeded"})
				c.Abort()
				return
			}
		}

		c.Header("X-RateLimit-Limit-global", strconv.Itoa(mw.limit))
		c.Header("X-RateLimit-Remaining-global", strconv.FormatInt(int64(mw.limit)-count2, 10))
		c.Header("X-RateLimit-Reset-global", now.Add(time.Duration(ttl2)*time.Second).Format(TimeFormat))
		c.Header("X-RateLimit-Limit-route", strconv.Itoa(limit))
		c.Header("X-RateLimit-Remaining-route", strconv.FormatInt(int64(limit)-count1, 10))
		c.Header("X-RateLimit-Reset-route", resetTime)

		c.Next()
	}
}

func (mw *middleware) parseCommand(command string) (time.Duration, error) {
	values := strings.Split(command, "-")
	if len(values) != 2 {
		return 0, errors.New(preference.FormatError)
	}

	unit, err := strconv.Atoi(values[0])
	if err != nil {
		return 0, errors.New(preference.FormatError)
	}

	if unit <= 0 {
		return 0, errors.New(preference.CommandError)
	}

	if t, ok := timeDict[strings.ToUpper(values[1])]; ok {
		return time.Duration(unit) * t, nil
	}

	return 0, errors.New(preference.FormatError)
}

func (mw *middleware) attachTraceSpanIDs(ctx context.Context, traceID, spanID string) context.Context {
	ctx = context.WithValue(ctx, preference.CONTEXT_KEY_LOG_TRACE_ID, traceID)
	ctx = context.WithValue(ctx, preference.CONTEXT_KEY_LOG_SPAN_ID, spanID)
	return ctx
}

func (mw *middleware) attachLogger(ctx context.Context) context.Context {
	return mw.log.With().
		Str(string(preference.CONTEXT_KEY_LOG_TRACE_ID), ctx.Value(preference.CONTEXT_KEY_LOG_TRACE_ID).(string)).
		Str(string(preference.CONTEXT_KEY_LOG_SPAN_ID), ctx.Value(preference.CONTEXT_KEY_LOG_SPAN_ID).(string)).
		Logger().
		WithContext(ctx)
}

func getAllowedOrigins() []string {
	allowedOriginsEnv := os.Getenv("ALLOWED_ORIGINS")
	if allowedOriginsEnv == "" {
		return []string{"http://localhost:3000", "http://localhost:8080"}
	}

	origins := strings.Split(allowedOriginsEnv, ",")
	for i, origin := range origins {
		origins[i] = strings.TrimSpace(origin)
	}
	return origins
}

func isOriginAllowed(origin string, allowedOrigins []string) bool {
	if origin == "" {
		return false
	}
	return slices.Contains(allowedOrigins, origin)
}
