package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"go-far/internal/preference"

	"github.com/rs/xid"
	"go.opentelemetry.io/otel/trace"
)

type errorResp struct {
	Error string `json:"error"`
}

// contextKey is a custom type for context keys
type contextKey string

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

const startTimeKey contextKey = "request_start_time"

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

// GenerateTraceID generates a new trace ID
func GenerateTraceID() string {
	return xid.New().String()
}

// GenerateSpanID generates a new span ID
func GenerateSpanID() string {
	return xid.New().String()
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
		// Handle exact slash suffix (existing behavior)
		if strings.HasSuffix(pub, "/") && strings.HasPrefix(path, pub) {
			return true
		}

		// Handle wildcard pattern "/*"
		if strings.HasSuffix(pub, "/*") {
			prefix, _ := strings.CutSuffix(pub, "/*")
			if strings.HasPrefix(path, prefix) {
				return true
			}
		}
	}

	return false
}

// Handler returns the main middleware handler for logging, tracing, and authentication
func (mw *middleware) Handler() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			start := time.Now()
			statusCode := http.StatusOK

			// Skip auth and logging for swagger, metrics, debug
			if mw.shouldSkipAuthAndLog(path) {
				next.ServeHTTP(w, r)
				mw.recordMetrics(r, statusCode, start)
				return
			}

			// Authentication
			if !mw.isPublicPath(path) {
				authHeader := r.Header.Get(preference.HeaderAuthorization)
				if authHeader == "" {
					mw.writeJSONError(w, http.StatusUnauthorized, "missing authorization header")
					mw.recordMetrics(r, http.StatusUnauthorized, start)
					return
				}

				accessDetails, err := mw.tkn.ValidateToken(r)
				if err != nil {
					mw.writeJSONError(w, http.StatusUnauthorized, "invalid token")
					mw.recordMetrics(r, http.StatusUnauthorized, start)
					return
				}

				authUser := &AuthUser{
					UserID:   accessDetails.UserID,
					Username: accessDetails.Username,
					Role:     accessDetails.Role,
				}

				r = r.WithContext(WithAuthUser(r.Context(), authUser))
			}

			// Logging and tracing
			ctx := withStartTime(r.Context(), start)
			ctx, traceID, spanID := mw.prepareContext(ctx, r)

			mw.logRequestStart(r, traceID, spanID)

			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(rw, r.WithContext(ctx))

			mw.logRequestEnd(traceID, spanID, start, rw.statusCode)
			mw.recordMetrics(r, rw.statusCode, start)
		})
	}
}

func (mw *middleware) shouldSkipAuthAndLog(path string) bool {
	return strings.HasPrefix(path, "/swagger/") ||
		strings.HasPrefix(path, "/metrics") ||
		strings.HasPrefix(path, "/debug/")
}

func (mw *middleware) recordMetrics(r *http.Request, status int, start time.Time) {
	if mw.metrics != nil {
		mw.metrics.RecordHttpRequestDuration(r.Method, r.URL.Path, status, time.Since(start))
	}
}

func (mw *middleware) prepareContext(ctx context.Context, r *http.Request) (ctxOut context.Context, traceID, spanID string) {
	span := trace.SpanFromContext(ctx)
	spanContext := span.SpanContext()

	if spanContext.HasTraceID() {
		traceID = spanContext.TraceID().String()
	} else {
		traceID = xid.New().String()
	}

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
	if path == "/metrics" || path == "/debug/vars" {
		return
	}

	if r.URL.RawQuery != "" {
		path = path + "?" + r.URL.RawQuery
	}

	event := mw.log.Info().
		Str(preference.EVENT, "START")

	if mw.tracingEnabled {
		event = event.
			Str(string(preference.CONTEXT_KEY_LOG_TRACE_ID), traceID).
			Str(string(preference.CONTEXT_KEY_LOG_SPAN_ID), spanID)
	} else {
		event = event.
			Str(string(preference.CONTEXT_KEY_LOG_REQUEST_ID), spanID)
	}

	event = event.
		Str(preference.METHOD, r.Method).
		Str(preference.URL, path).
		Str(preference.USER_AGENT, r.UserAgent()).
		Str(preference.ADDR, r.Host)

	event.Send()
}

func (mw *middleware) logRequestEnd(traceID, spanID string, start time.Time, statusCode int) {
	latency := time.Since(start)
	if latency > time.Minute {
		latency = latency.Truncate(time.Second)
	}

	event := mw.log.Info().
		Str(preference.EVENT, "END")

	if mw.tracingEnabled {
		event = event.
			Str(string(preference.CONTEXT_KEY_LOG_TRACE_ID), traceID).
			Str(string(preference.CONTEXT_KEY_LOG_SPAN_ID), spanID)
	} else {
		event = event.
			Str(string(preference.CONTEXT_KEY_LOG_REQUEST_ID), spanID)
	}

	event = event.
		Str(preference.LATENCY, latency.String()).
		Int(preference.STATUS, statusCode)

	event.Send()
}

func (mw *middleware) attachTraceSpanIDs(ctx context.Context, traceID, spanID string) context.Context {
	ctx = context.WithValue(ctx, preference.CONTEXT_KEY_LOG_TRACE_ID, traceID)
	ctx = context.WithValue(ctx, preference.CONTEXT_KEY_LOG_SPAN_ID, spanID)

	return mw.attachLogger(ctx)
}

func (mw *middleware) attachLogger(ctx context.Context) context.Context {
	logBuilder := mw.log.With()

	if mw.tracingEnabled {
		traceID, _ := ctx.Value(preference.CONTEXT_KEY_LOG_TRACE_ID).(string)
		spanID, _ := ctx.Value(preference.CONTEXT_KEY_LOG_SPAN_ID).(string)

		logBuilder = logBuilder.
			Str(string(preference.CONTEXT_KEY_LOG_TRACE_ID), traceID).
			Str(string(preference.CONTEXT_KEY_LOG_SPAN_ID), spanID)
	} else {
		logBuilder = logBuilder.
			Str(string(preference.CONTEXT_KEY_LOG_REQUEST_ID), ctx.Value(preference.CONTEXT_KEY_LOG_SPAN_ID).(string))
	}

	return logBuilder.Logger().WithContext(ctx)
}

func (mw *middleware) writeJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set(preference.HeaderContentType, preference.ContentTypeJSON)
	w.WriteHeader(status)
	if data, err := json.Marshal(errorResp{Error: msg}); err == nil {
		_, _ = w.Write(data)
	}
}
