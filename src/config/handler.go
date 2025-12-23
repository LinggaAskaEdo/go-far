package config

import (
	"context"
	"strings"
	"time"

	"go-far/src/preference"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
)

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
