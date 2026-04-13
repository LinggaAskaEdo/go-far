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

func (mw *middleware) writeJSONError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set(preference.HeaderContentType, preference.ContentTypeJSON)
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// RoleLimiter returns a rate limiting middleware based on user role
func (mw *middleware) RoleLimiter() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authUser, ok := GetAuthUser(r.Context())
			if !ok {
				http.Error(w, `{"error":"unauthenticated"}`, http.StatusUnauthorized)
				return
			}

			command, limit := mw.getRoleRateLimit(authUser.Role)

			duration, err := mw.parseCommand(command)
			if err != nil {
				mw.writeJSONError(w, err.Error(), http.StatusInternalServerError)
				return
			}

			now := time.Now()
			limitResult, err := mw.evalRoleRateLimit(r.URL.Path, r.Method, authUser, limit, duration)
			if err != nil {
				mw.writeJSONError(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if !limitResult.Allowed {
				mw.writeRoleRateLimitExceeded(w, limitResult, now, limit, authUser.Role)
				return
			}

			mw.setRoleRateLimitHeaders(w, limitResult, now, limit)
			next.ServeHTTP(w, r)
		})
	}
}

func (mw *middleware) evalRoleRateLimit(path, method string, authUser *AuthUser, limit int, duration time.Duration) (rateLimitResult, error) {
	routeKey := "ratelimit:route:" + path + ":" + method + ":" + authUser.UserID
	globalKey := "ratelimit:user:" + authUser.UserID

	result, err := mw.rdb.Eval(context.Background(), rateLimitLuaScript, []string{routeKey, globalKey},
		limit,                   // route limit
		limit,                   // global limit (same as role limit)
		int(duration.Seconds()), // route duration
		int(duration.Seconds()), // global duration
	).Result()
	if err != nil {
		return rateLimitResult{}, err
	}

	resultArr, ok := result.([]interface{})
	if !ok || len(resultArr) < 5 {
		return rateLimitResult{}, errors.New("invalid rate limit response")
	}

	return parseRateLimitResult(resultArr), nil
}

func (mw *middleware) writeRoleRateLimitExceeded(w http.ResponseWriter, result rateLimitResult, now time.Time, limit int, role string) {
	w.Header().Set(preference.HeaderContentType, preference.ContentTypeJSON)
	w.Header().Set(preference.HeaderXRateLimitLimitGlobal, strconv.Itoa(limit))
	w.Header().Set(preference.HeaderXRateLimitRemainingGlobal, "0")
	w.Header().Set(preference.HeaderXRateLimitResetGlobal, formatTime(now, result.GlobalTTL))
	w.WriteHeader(http.StatusTooManyRequests)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": "role rate limit exceeded", "role": role})
}

func (mw *middleware) setRoleRateLimitHeaders(w http.ResponseWriter, result rateLimitResult, now time.Time, limit int) {
	w.Header().Set(preference.HeaderXRateLimitLimitGlobal, strconv.Itoa(limit))
	w.Header().Set(preference.HeaderXRateLimitRemainingGlobal, formatRemaining(int64(limit), result.GlobalCount))
	w.Header().Set(preference.HeaderXRateLimitResetGlobal, formatTime(now, result.GlobalTTL))
}

// getRoleRateLimit returns the rate limit config for a given role
func (mw *middleware) getRoleRateLimit(role string) (command string, limit int) {
	switch role {
	case "admin":
		return mw.roleRateLimit.Admin.Command, mw.roleRateLimit.Admin.Limit
	case "user":
		return mw.roleRateLimit.User.Command, mw.roleRateLimit.User.Limit
	case "guest":
		return mw.roleRateLimit.Guest.Command, mw.roleRateLimit.Guest.Limit
	default:
		return mw.roleRateLimit.Guest.Command, mw.roleRateLimit.Guest.Limit
	}
}
