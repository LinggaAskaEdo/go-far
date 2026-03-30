package rest

import (
	"fmt"
	"net/http"
	"time"

	"go-far/src/dto"
	x "go-far/src/errors"
	"go-far/src/preference"

	"github.com/gin-gonic/gin"
)

// Health godoc
//
//	@Summary		Health check endpoint
//	@Description	Returns the health status of the service
//	@Tags			health
//	@Produce		json
//	@Success		200	{object}	dto.HttpSuccessResp{data=dto.HealthStatus}
//	@Router			/health [get]
func (e *rest) Health(c *gin.Context) {
	ctx := c.Request.Context()

	status := dto.HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now().Format(time.RFC3339),
		Service:   "go-far-app",
		Version:   "1.0.0",
	}

	e.httpRespSuccess(c, http.StatusOK, status, nil)
	_ = ctx
}

// Ready godoc
//
//	@Summary		Readiness check endpoint
//	@Description	Returns the readiness status of the service (checks dependencies)
//	@Tags			health
//	@Produce		json
//	@Success		200	{object}	dto.HttpSuccessResp{data=dto.ReadinessStatus}
//	@Failure		503	{object}	dto.HTTPErrorResp
//	@Router			/ready [get]
func (e *rest) Ready(c *gin.Context) {
	ctx := c.Request.Context()

	// Add dependency checks here (database, redis, etc.)
	// For now, return a simple ready response
	status := dto.ReadinessStatus{
		Status:       "ready",
		Timestamp:    time.Now().Format(time.RFC3339),
		Dependencies: map[string]string{"database": "unknown", "redis": "unknown"},
	}

	e.httpRespSuccess(c, http.StatusOK, status, nil)
	_ = ctx
}

func (e *rest) httpRespSuccess(c *gin.Context, statusCode int, resp any, p *dto.Pagination) {
	meta := dto.Meta{
		Path:       c.Request.URL.Path,
		StatusCode: statusCode,
		Status:     http.StatusText(statusCode),
		Message:    fmt.Sprintf("%s %s [%d] %s", c.Request.Method, c.Request.RequestURI, statusCode, http.StatusText(statusCode)),
		Error:      nil,
		Timestamp:  time.Now().Format(time.RFC3339),
	}

	httpResp := &dto.HttpSuccessResp{
		Meta:       meta,
		Data:       any(resp),
		Pagination: p,
	}

	c.JSON(statusCode, httpResp)
}

func (e *rest) httpRespError(c *gin.Context, err error) {
	lang := preference.LANG_ID

	appLangHeader := http.CanonicalHeaderKey(preference.APP_LANG)
	if c.Request.Header[appLangHeader] != nil && c.Request.Header[appLangHeader][0] == preference.LANG_EN {
		lang = preference.LANG_EN
	}

	statusCode, displayError := x.Compile(x.COMMON, err, lang, true)
	statusStr := http.StatusText(statusCode)

	jsonErrResp := &dto.HTTPErrorResp{
		Meta: dto.Meta{
			Path:       c.Request.URL.Path,
			StatusCode: statusCode,
			Status:     statusStr,
			Message:    fmt.Sprintf("%s %s [%d] %s", c.Request.Method, c.Request.RequestURI, statusCode, http.StatusText(statusCode)),
			Error:      &displayError,
			Timestamp:  time.Now().Format(time.RFC3339),
		},
	}

	c.JSON(statusCode, jsonErrResp)
}
