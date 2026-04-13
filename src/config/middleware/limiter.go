package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go-far/src/preference"
)

// Header constants to avoid string duplication
const (
	headerContentType               = "Content-Type"
	headerXRateLimitLimitGlobal     = "X-RateLimit-Limit-global"
	headerXRateLimitRemainingGlobal = "X-RateLimit-Remaining-global"
	headerXRateLimitResetGlobal     = "X-RateLimit-Reset-global"
	headerXRateLimitLimitRoute      = "X-RateLimit-Limit-route"
	headerXRateLimitRemainingRoute  = "X-RateLimit-Remaining-route"
	headerXRateLimitResetRoute      = "X-RateLimit-Reset-route"
)

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

// rateLimitResult holds parsed rate limiting data
type rateLimitResult struct {
	Allowed      bool
	RouteCount   int64
	GlobalCount  int64
	RouteTTL     int64
	GlobalTTL    int64
	ExceededType string // "route" or "global"
}

// Limiter returns the rate limiting middleware handler
func (mw *middleware) Limiter(command string, limit int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			duration, err := mw.parseCommand(command)
			if err != nil {
				mw.writeJSONError(w, err.Error(), http.StatusInternalServerError)
				return
			}

			now := time.Now()
			clientIP := getClientIP(r)

			routeKey := buildRateLimitKey("route", r.URL.Path, r.Method, clientIP)
			globalKey := buildRateLimitKey("global", "", "", clientIP)

			result, err := mw.rdb.Eval(context.Background(), rateLimitLuaScript, []string{routeKey, globalKey},
				limit,                    // route limit
				mw.limit,                 // global limit
				int(duration.Seconds()),  // route duration
				int(mw.period.Seconds()), // global duration
			).Result()
			if err != nil {
				mw.writeJSONError(w, err.Error(), http.StatusInternalServerError)
				return
			}

			resultArr, ok := result.([]interface{})
			if !ok || len(resultArr) < 5 {
				mw.writeJSONError(w, "invalid rate limit response", http.StatusInternalServerError)
				return
			}

			limitResult := parseRateLimitResult(resultArr)

			if !limitResult.Allowed {
				mw.handleRateLimitExceeded(w, limitResult, now, limit)
				return
			}

			mw.setRateLimitHeaders(w, limitResult, now, mw.limit, limit)
			next.ServeHTTP(w, r)
		})
	}
}

// parseRateLimitResult converts raw Redis result to structured data
func parseRateLimitResult(resultArr []interface{}) rateLimitResult {
	return rateLimitResult{
		Allowed:      resultArr[0].(int64) == 1,
		RouteCount:   resultArr[1].(int64),
		GlobalCount:  resultArr[2].(int64),
		RouteTTL:     resultArr[3].(int64),
		GlobalTTL:    resultArr[4].(int64),
		ExceededType: getStringOrDefault(resultArr, 5, ""),
	}
}

// getStringOrDefault safely extracts a string from result array
func getStringOrDefault(resultArr []interface{}, idx int, defaultVal string) string {
	if idx < len(resultArr) {
		if s, ok := resultArr[idx].(string); ok {
			return s
		}
	}

	return defaultVal
}

func (mw *middleware) handleRateLimitExceeded(w http.ResponseWriter, result rateLimitResult, now time.Time, routeLimit int) {
	w.Header().Set(headerContentType, contentTypeJSON)

	if result.ExceededType == "route" {
		w.Header().Set(headerXRateLimitLimitRoute, strconv.Itoa(routeLimit))
		w.Header().Set(headerXRateLimitRemainingRoute, "0")
		w.Header().Set(headerXRateLimitResetRoute, formatTime(now, result.RouteTTL))
		w.WriteHeader(http.StatusTooManyRequests)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "route rate limit exceeded"})

		return
	}

	w.Header().Set(headerXRateLimitLimitGlobal, strconv.Itoa(mw.limit))
	w.Header().Set(headerXRateLimitRemainingGlobal, "0")
	w.Header().Set(headerXRateLimitResetGlobal, formatTime(now, result.GlobalTTL))
	w.WriteHeader(http.StatusTooManyRequests)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": "global rate limit exceeded"})
}

func (mw *middleware) setRateLimitHeaders(w http.ResponseWriter, result rateLimitResult, now time.Time, globalLimit, routeLimit int) {
	w.Header().Set(headerXRateLimitLimitGlobal, strconv.Itoa(globalLimit))
	w.Header().Set(headerXRateLimitRemainingGlobal, formatRemaining(int64(globalLimit), result.GlobalCount))
	w.Header().Set(headerXRateLimitResetGlobal, formatTime(now, result.GlobalTTL))
	w.Header().Set(headerXRateLimitLimitRoute, strconv.Itoa(routeLimit))
	w.Header().Set(headerXRateLimitRemainingRoute, formatRemaining(int64(routeLimit), result.RouteCount))
	w.Header().Set(headerXRateLimitResetRoute, formatTime(now, result.RouteTTL))
}

// formatRemaining calculates remaining requests
func formatRemaining(limit, count int64) string {
	remaining := limit - count
	if remaining < 0 {
		remaining = 0
	}

	return strconv.FormatInt(remaining, 10)
}

// formatTime formats duration as RFC3339
func formatTime(now time.Time, ttlSeconds int64) string {
	return now.Add(time.Duration(ttlSeconds) * time.Second).Format(TimeFormat)
}

// buildRateLimitKey constructs a Redis key for rate limiting
func buildRateLimitKey(prefix, path, method, clientIP string) string {
	if path != "" && method != "" {
		return "ratelimit:" + prefix + ":" + path + ":" + method + ":" + clientIP
	}

	return "ratelimit:" + prefix + ":" + clientIP
}

// Content-Type value constant
const contentTypeJSON = "application/json"

func (mw *middleware) writeJSONError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set(headerContentType, contentTypeJSON)
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
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

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}

	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	return r.RemoteAddr
}
