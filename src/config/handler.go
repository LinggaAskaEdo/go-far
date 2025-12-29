package config

import (
	"context"
	"strings"
	"time"

	"go-far/src/preference"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
	"go.opentelemetry.io/otel/trace"
)

func (mw *middleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		path := c.Request.URL.Path

		if !strings.HasPrefix(path, "/swagger/") { // skip logging swagger request
			start := time.Now()

			// Get trace context from OpenTelemetry
			span := trace.SpanFromContext(ctx)
			spanContext := span.SpanContext()
			traceID := spanContext.TraceID().String()
			spanID := spanContext.SpanID().String()

			reqID := c.GetHeader("X-Request-ID")
			if reqID == "" {
				spanID = xid.New().String()
			}

			// ctx = mw.attachReqID(ctx, reqID)
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
				Str("trace_id", traceID).
				Str("span_id", spanID).
				Str(preference.LATENCY, param.Latency.String()).
				Int(preference.STATUS, param.StatusCode).
				Send()
		}
	}
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
