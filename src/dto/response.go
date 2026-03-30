package dto

import (
	x "go-far/src/errors"
)

type Meta struct {
	Path       string      `json:"path" extensions:"x-order=0"`
	StatusCode int         `json:"status_code" extensions:"x-order=1"`
	Status     string      `json:"status" extensions:"x-order=2"`
	Message    string      `json:"message" extensions:"x-order=3"`
	Error      *x.AppError `json:"error,omitempty" swaggertype:"primitive,object" extensions:"x-order=4"`
	Timestamp  string      `json:"timestamp" extensions:"x-order=5"`
}

type HttpSuccessResp struct {
	Meta       Meta        `json:"metadata" extensions:"x-order=0"`
	Data       any         `json:"data,omitempty" extensions:"x-order=1"`
	Pagination *Pagination `json:"pagination,omitempty" extensions:"x-order=2"`
}

type HTTPErrorResp struct {
	Meta Meta `json:"metadata"`
}

// HealthStatus represents the health check response
type HealthStatus struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Service   string `json:"service"`
	Version   string `json:"version"`
}

// ReadinessStatus represents the readiness check response
type ReadinessStatus struct {
	Status       string            `json:"status"`
	Timestamp    string            `json:"timestamp"`
	Dependencies map[string]string `json:"dependencies"`
}
