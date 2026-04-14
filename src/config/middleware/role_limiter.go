package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go-far/src/model/entity"
	"go-far/src/preference"

	"github.com/rs/zerolog"
)

// Lua script for atomic rate limiting (single key, no race conditions)
const rateLimitLuaScript = `
	local key = KEYS[1]
	local limit = tonumber(ARGV[1])
	local duration = tonumber(ARGV[2])

	local count = tonumber(redis.call('INCR', key))
	if count == 1 then
		redis.call('EXPIRE', key, duration)
	end

	local ttl = redis.call('TTL', key)

	if count > limit then
		return {0, count, ttl}
	end

	return {1, count, ttl}
`

// rateLimitResult holds parsed rate limiting data
type rateLimitResult struct {
	Allowed bool
	Count   int64
	TTL     int64
}

// parseRateLimitResult converts raw Redis result to structured data
func parseRateLimitResult(resultArr []any) rateLimitResult {
	return rateLimitResult{
		Allowed: resultArr[0].(int64) == 1,
		Count:   resultArr[1].(int64),
		TTL:     resultArr[2].(int64),
	}
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

// AuthLimiter returns a rate limiting middleware for auth endpoints (login, register, refresh)
// This prevents brute force attacks and mass registration abuse by limiting requests per IP.
func (mw *middleware) AuthLimiter() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Use client IP as the rate limit key
			ip := getClientIP(r)
			key := "ratelimit:auth:" + ip

			now := time.Now()
			limitResult, err := mw.evalAuthRateLimit(key)
			if err != nil {
				zerolog.Ctx(r.Context()).Error().Err(err).Msg("eval auth rate limit failed")
				next.ServeHTTP(w, r)
				return
			}

			if !limitResult.Allowed {
				mw.writeAuthRateLimitExceeded(w, limitResult, now)
				return
			}

			mw.setAuthRateLimitHeaders(w, limitResult, now)
			next.ServeHTTP(w, r)
		})
	}
}

// getClientIP extracts the client IP from request headers
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxied requests)
	if xff := r.Header.Get(preference.HeaderXForwardedFor); xff != "" {
		// Take the first IP in the comma-separated list
		if idx := strings.Index(xff, ","); idx > 0 {
			return xff[:idx]
		}
		return xff
	}

	// Check X-Real-IP header
	if xri := r.Header.Get(preference.HeaderXRealIP); xri != "" {
		return xri
	}

	// Fallback to RemoteAddr
	return r.RemoteAddr
}

// evalAuthRateLimit evaluates rate limit for auth endpoints
func (mw *middleware) evalAuthRateLimit(key string) (rateLimitResult, error) {
	result, err := mw.rdb.Eval(context.Background(), rateLimitLuaScript, []string{key},
		mw.authLimit,                 // rate limit
		int(mw.authPeriod.Seconds()), // window duration in seconds
	).Result()
	if err != nil {
		return rateLimitResult{}, err
	}

	resultArr, ok := result.([]any)
	if !ok || len(resultArr) < 3 {
		return rateLimitResult{}, errors.New("invalid rate limit response")
	}

	return parseRateLimitResult(resultArr), nil
}

// writeAuthRateLimitExceeded writes the rate limit exceeded response for auth endpoints
func (mw *middleware) writeAuthRateLimitExceeded(w http.ResponseWriter, result rateLimitResult, now time.Time) {
	w.Header().Set(preference.HeaderContentType, preference.ContentTypeJSON)
	w.Header().Set("X-RateLimit-Limit-auth", strconv.Itoa(mw.authLimit))
	w.Header().Set("X-RateLimit-Remaining-auth", "0")
	w.Header().Set("X-RateLimit-Reset-auth", formatTime(now, result.TTL))
	w.WriteHeader(http.StatusTooManyRequests)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": "auth rate limit exceeded, please try again later",
	})
}

// setAuthRateLimitHeaders sets rate limit headers for auth endpoints
func (mw *middleware) setAuthRateLimitHeaders(w http.ResponseWriter, result rateLimitResult, now time.Time) {
	w.Header().Set("X-RateLimit-Limit-auth", strconv.Itoa(mw.authLimit))
	w.Header().Set("X-RateLimit-Remaining-auth", formatRemaining(int64(mw.authLimit), result.Count))
	w.Header().Set("X-RateLimit-Reset-auth", formatTime(now, result.TTL))
}

func (mw *middleware) evalRoleRateLimit(path, method string, authUser *AuthUser, limit int, duration time.Duration) (rateLimitResult, error) {
	// Normalize path to remove dynamic segments (UUIDs, numeric IDs)
	normalizedPath := normalizePath(path)

	key := "ratelimit:role:" + normalizedPath + ":" + method + ":" + authUser.UserID

	result, err := mw.rdb.Eval(context.Background(), rateLimitLuaScript, []string{key},
		limit,                   // rate limit
		int(duration.Seconds()), // window duration in seconds
	).Result()
	if err != nil {
		return rateLimitResult{}, err
	}

	resultArr, ok := result.([]any)
	if !ok || len(resultArr) < 3 {
		return rateLimitResult{}, errors.New("invalid rate limit response")
	}

	return parseRateLimitResult(resultArr), nil
}

func (mw *middleware) writeRoleRateLimitExceeded(w http.ResponseWriter, result rateLimitResult, now time.Time, limit int, role string) {
	w.Header().Set(preference.HeaderContentType, preference.ContentTypeJSON)
	w.Header().Set(preference.HeaderXRateLimitLimitGlobal, strconv.Itoa(limit))
	w.Header().Set(preference.HeaderXRateLimitRemainingGlobal, "0")
	w.Header().Set(preference.HeaderXRateLimitResetGlobal, formatTime(now, result.TTL))
	w.WriteHeader(http.StatusTooManyRequests)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": "role rate limit exceeded", "role": role})
}

func (mw *middleware) setRoleRateLimitHeaders(w http.ResponseWriter, result rateLimitResult, now time.Time, limit int) {
	w.Header().Set(preference.HeaderXRateLimitLimitGlobal, strconv.Itoa(limit))
	w.Header().Set(preference.HeaderXRateLimitRemainingGlobal, formatRemaining(int64(limit), result.Count))
	w.Header().Set(preference.HeaderXRateLimitResetGlobal, formatTime(now, result.TTL))
}

// getRoleRateLimit returns the rate limit config for a given role
func (mw *middleware) getRoleRateLimit(role string) (command string, limit int) {
	switch entity.Role(role) {
	case entity.RoleAdmin:
		return mw.roleRateLimit.Admin.Command, mw.roleRateLimit.Admin.Limit
	case entity.RoleUser:
		return mw.roleRateLimit.User.Command, mw.roleRateLimit.User.Limit
	case entity.RoleGuest:
		return mw.roleRateLimit.Guest.Command, mw.roleRateLimit.Guest.Limit
	default:
		// Unknown role defaults to guest limits (safest)
		return mw.roleRateLimit.Guest.Command, mw.roleRateLimit.Guest.Limit
	}
}

// normalizePath normalizes URL paths by replacing dynamic segments (UUIDs, numeric IDs) with {id}
// This ensures rate limiting works across all resources for a user, not per-resource.
// Examples:
//   - /users/5f0f2b8c0f5e4e001c8b4567 -> /users/{id}
//   - /cars/123/owner -> /cars/{id}/owner
func normalizePath(path string) string {
	// Match UUID-like patterns (xid: 20-char alphanumeric with dashes)
	uuidRegex := regexp.MustCompile(`/[a-f0-9]{20,}`)
	path = uuidRegex.ReplaceAllString(path, "/{id}")

	// Match numeric IDs
	numericRegex := regexp.MustCompile(`/\d+`)
	path = numericRegex.ReplaceAllString(path, "/{id}")

	return path
}
