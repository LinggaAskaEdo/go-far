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

	"github.com/rs/zerolog"
)

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
