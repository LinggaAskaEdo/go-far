package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"go-far/src/preference"

	"github.com/rs/xid"
	"go.opentelemetry.io/otel/trace"
)

// contextKey is a custom type for context keys
type contextKey string

const startTimeKey contextKey = "request_start_time"

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

// Handler returns the main middleware handler for request logging and tracing
func (mw *middleware) Handler() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/swagger/") {
				next.ServeHTTP(w, r)
				return
			}

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

func (mw *middleware) prepareContext(ctx context.Context, r *http.Request) (context.Context, string, string) {
	span := trace.SpanFromContext(ctx)
	spanContext := span.SpanContext()

	traceID := spanContext.TraceID().String()
	spanID := spanContext.SpanID().String()

	if r.Header.Get("X-Request-ID") == "" {
		spanID = xid.New().String()
	}

	ctx = mw.attachTraceSpanIDs(ctx, traceID, spanID)
	r.Header.Set("X-Request-ID", spanID)

	return ctx, traceID, spanID
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
	return mw.log.With().
		Str(string(preference.CONTEXT_KEY_LOG_TRACE_ID), ctx.Value(preference.CONTEXT_KEY_LOG_TRACE_ID).(string)).
		Str(string(preference.CONTEXT_KEY_LOG_SPAN_ID), ctx.Value(preference.CONTEXT_KEY_LOG_SPAN_ID).(string)).
		Logger().
		WithContext(ctx)
}
