package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go-far/src/model/dto"
	x "go-far/src/model/errors"
	"go-far/src/preference"
)

// Health godoc
//
//	@Summary		Health check endpoint
//	@Description	Returns the health status of the service
//	@Tags			health
//	@Produce		json
//	@Success		200	{object}	dto.HttpSuccessResp{data=dto.HealthStatus}
//	@Router			/health [get]
func (e *rest) Health(w http.ResponseWriter, r *http.Request) {
	status := dto.HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now().Format(time.RFC3339),
		Service:   "go-far",
		Version:   "1.7.0",
	}

	e.httpRespSuccess(w, r, http.StatusOK, status, nil)
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
func (e *rest) Ready(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	depStatus := map[string]string{}
	ready := true

	// Check database
	if err := e.sql0.PingContext(ctx); err != nil {
		depStatus["database"] = preference.StatusNotReady
		ready = false
	} else {
		depStatus["database"] = preference.StatusReady
	}

	// Check Redis
	if err := e.redis.Ping(ctx).Err(); err != nil {
		depStatus["redis"] = preference.StatusNotReady
		ready = false
	} else {
		depStatus["redis"] = preference.StatusReady
	}

	if !ready {
		status := dto.ReadinessStatus{
			Status:       preference.StatusNotReady,
			Timestamp:    time.Now().Format(time.RFC3339),
			Dependencies: depStatus,
		}
		e.httpRespSuccess(w, r, http.StatusServiceUnavailable, status, nil)
		return
	}

	status := dto.ReadinessStatus{
		Status:       preference.StatusReady,
		Timestamp:    time.Now().Format(time.RFC3339),
		Dependencies: depStatus,
	}

	e.httpRespSuccess(w, r, http.StatusOK, status, nil)
}

func (e *rest) httpRespSuccess(w http.ResponseWriter, r *http.Request, statusCode int, resp any, p *dto.Pagination) {
	meta := dto.Meta{
		Path:       r.URL.Path,
		StatusCode: statusCode,
		Status:     http.StatusText(statusCode),
		Message:    fmt.Sprintf("%s %s [%d] %s", r.Method, r.RequestURI, statusCode, http.StatusText(statusCode)),
		Error:      nil,
		Timestamp:  time.Now().Format(time.RFC3339),
	}

	httpResp := &dto.HttpSuccessResp{
		Meta:       meta,
		Data:       any(resp),
		Pagination: p,
	}

	writeJSON(w, statusCode, httpResp)
}

func (e *rest) httpRespError(w http.ResponseWriter, r *http.Request, err error) {
	lang := preference.LANG_ID

	appLangHeader := http.CanonicalHeaderKey(preference.APP_LANG)
	if r.Header[appLangHeader] != nil && r.Header[appLangHeader][0] == preference.LANG_EN {
		lang = preference.LANG_EN
	}

	statusCode, displayError := x.Compile(x.COMMON, err, lang, true)
	statusStr := http.StatusText(statusCode)

	jsonErrResp := &dto.HTTPErrorResp{
		Meta: dto.Meta{
			Path:       r.URL.Path,
			StatusCode: statusCode,
			Status:     statusStr,
			Message:    fmt.Sprintf("%s %s [%d] %s", r.Method, r.RequestURI, statusCode, http.StatusText(statusCode)),
			Error:      &displayError,
			Timestamp:  time.Now().Format(time.RFC3339),
		},
	}

	writeJSON(w, statusCode, jsonErrResp)
}

func writeJSON(w http.ResponseWriter, statusCode int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(v)
}
