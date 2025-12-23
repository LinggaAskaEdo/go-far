package config

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go-far/src/preference"

	"github.com/gin-gonic/gin"
)

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

		routeKey := c.FullPath() + ":" + c.Request.Method + ":" + clientIP
		globalKey := "global:" + clientIP

		ctx := context.Background()

		// Check and increment route-specific limit
		routeCount, err := mw.rdb.Incr(ctx, routeKey).Result()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		// Set expiration on first request
		if routeCount == 1 {
			mw.rdb.Expire(ctx, routeKey, duration)
		}

		// Get TTL for route key
		routeTTL, _ := mw.rdb.TTL(ctx, routeKey).Result()
		routeReset := now.Add(routeTTL).Format(TimeFormat)

		// Check route limit
		if routeCount > int64(limit) {
			c.Header("X-RateLimit-Limit-route", strconv.Itoa(limit))
			c.Header("X-RateLimit-Remaining-route", "0")
			c.Header("X-RateLimit-Reset-route", routeReset)
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "route rate limit exceeded", "reset": routeReset})
			c.Abort()
			return
		}

		// Check and increment global limit
		globalCount, err := mw.rdb.Incr(ctx, globalKey).Result()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		// Set expiration on first request
		if globalCount == 1 {
			mw.rdb.Expire(ctx, globalKey, mw.period)
		}

		// Get TTL for global key
		globalTTL, _ := mw.rdb.TTL(ctx, globalKey).Result()
		globalReset := now.Add(globalTTL).Format(TimeFormat)

		// Check global limit
		if globalCount > int64(mw.limit) {
			c.Header("X-RateLimit-Limit-global", strconv.Itoa(mw.limit))
			c.Header("X-RateLimit-Remaining-global", "0")
			c.Header("X-RateLimit-Reset-global", globalReset)
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "global rate limit exceeded", "reset": globalReset})
			c.Abort()
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit-global", strconv.Itoa(mw.limit))
		c.Header("X-RateLimit-Remaining-global", strconv.FormatInt(int64(mw.limit)-globalCount, 10))
		c.Header("X-RateLimit-Reset-global", globalReset)
		c.Header("X-RateLimit-Limit-route", strconv.Itoa(limit))
		c.Header("X-RateLimit-Remaining-route", strconv.FormatInt(int64(limit)-routeCount, 10))
		c.Header("X-RateLimit-Reset-route", routeReset)

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
