package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"go-far/src/preference"

	"github.com/rs/xid"
	"go.opentelemetry.io/otel/trace"
)

type errorResp struct {
	Error string `json:"error"`
}

// contextKey is a custom type for context keys
type contextKey string

const startTimeKey contextKey = "request_start_time"

// AuthUser holds authenticated user info from the token
type AuthUser struct {
	UserID   string
	Username string
	Role     string
}

// responseWriter wraps http.ResponseWriter to capture the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// withStartTime adds start time to context
func withStartTime(ctx context.Context, t time.Time) context.Context {
	return context.WithValue(ctx, startTimeKey, t)
}

// WithAuthUser stores authenticated user info in context
func WithAuthUser(ctx context.Context, user *AuthUser) context.Context {
	return context.WithValue(ctx, preference.ContextKeyAuthUser, user)
}

// GetAuthUser retrieves authenticated user info from context
func GetAuthUser(ctx context.Context) (*AuthUser, bool) {
	user, ok := ctx.Value(preference.ContextKeyAuthUser).(*AuthUser)
	return user, ok
}

// isPublicPath checks if a path is exempt from authentication
func (mw *middleware) isPublicPath(path string) bool {
	if mw.publicPaths[path] {
		return true
	}

	for pub := range mw.publicPaths {
		if strings.HasSuffix(pub, "/") && strings.HasPrefix(path, pub) {
			return true
		}
	}

	return false
}

// Handler returns the main middleware handler for logging, tracing, and authentication
func (mw *middleware) Handler() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			isPublic := mw.isPublicPath(path)

			// Authentication (skip public paths)
			if !isPublic {
				authHeader := r.Header.Get(preference.HeaderAuthorization)
				if authHeader == "" {
					mw.writeJSONError(w, http.StatusUnauthorized, "missing authorization header")
					return
				}

				accessDetails, err := mw.tkn.ValidateToken(r)
				if err != nil {
					mw.writeJSONError(w, http.StatusUnauthorized, "invalid token")
					return
				}

				authUser := &AuthUser{
					UserID:   accessDetails.UserID,
					Username: accessDetails.Username,
					Role:     accessDetails.Role,
				}

				r = r.WithContext(WithAuthUser(r.Context(), authUser))
			}

			// Skip logging for swagger
			if strings.HasPrefix(path, "/swagger/") {
				next.ServeHTTP(w, r)
				return
			}

			// Logging and tracing
			start := time.Now()
			ctx := withStartTime(r.Context(), start)
			ctx, traceID, spanID := mw.prepareContext(ctx, r)

			mw.logRequestStart(r, traceID, spanID)

			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(rw, r.WithContext(ctx))

			mw.logRequestEnd(traceID, spanID, start, rw.statusCode)
		})
	}
}

func (mw *middleware) prepareContext(ctx context.Context, r *http.Request) (ctxOut context.Context, traceID, spanID string) {
	span := trace.SpanFromContext(ctx)
	spanContext := span.SpanContext()

	traceID = spanContext.TraceID().String()

	if requestID := r.Header.Get(preference.HeaderXRequestID); requestID != "" {
		spanID = requestID
	} else {
		spanID = xid.New().String()
		r.Header.Set(preference.HeaderXRequestID, spanID)
	}

	ctxOut = mw.attachTraceSpanIDs(ctx, traceID, spanID)

	return
}

func (mw *middleware) logRequestStart(r *http.Request, traceID, spanID string) {
	path := r.URL.Path
	if r.URL.RawQuery != "" {
		path = path + "?" + r.URL.RawQuery
	}

	mw.log.Info().
		Str(preference.EVENT, "START").
		Str("trace_id", traceID).
		Str("span_id", spanID).
		Str(preference.METHOD, r.Method).
		Str(preference.URL, path).
		Str(preference.USER_AGENT, r.UserAgent()).
		Str(preference.ADDR, r.Host).
		Send()
}

func (mw *middleware) logRequestEnd(traceID, spanID string, start time.Time, statusCode int) {
	latency := time.Since(start)
	if latency > time.Minute {
		latency = latency.Truncate(time.Second)
	}

	mw.log.Info().
		Str(preference.EVENT, "END").
		Str("trace_id", traceID).
		Str("span_id", spanID).
		Str(preference.LATENCY, latency.String()).
		Int(preference.STATUS, statusCode).
		Send()
}

func (mw *middleware) attachTraceSpanIDs(ctx context.Context, traceID, spanID string) context.Context {
	ctx = context.WithValue(ctx, preference.CONTEXT_KEY_LOG_TRACE_ID, traceID)
	ctx = context.WithValue(ctx, preference.CONTEXT_KEY_LOG_SPAN_ID, spanID)
	return mw.attachLogger(ctx)
}

func (mw *middleware) attachLogger(ctx context.Context) context.Context {
	traceID, _ := ctx.Value(preference.CONTEXT_KEY_LOG_TRACE_ID).(string)
	spanID, _ := ctx.Value(preference.CONTEXT_KEY_LOG_SPAN_ID).(string)

	return mw.log.With().
		Str(string(preference.CONTEXT_KEY_LOG_TRACE_ID), traceID).
		Str(string(preference.CONTEXT_KEY_LOG_SPAN_ID), spanID).
		Logger().
		WithContext(ctx)
}

func (mw *middleware) writeJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set(preference.HeaderContentType, preference.ContentTypeJSON)
	w.WriteHeader(status)
	if data, err := json.Marshal(errorResp{Error: msg}); err == nil {
		_, _ = w.Write(data)
	}
}
