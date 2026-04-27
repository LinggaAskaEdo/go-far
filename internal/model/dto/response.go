package dto

import (
	"go-far/internal/infra/token"
	"go-far/internal/model/entity"
	appErr "go-far/internal/model/errors"
)

type TokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresAt    int64  `json:"expiresAt"`
}

// UserTokenData is a helper struct for JWT token generation
type UserTokenData struct {
	PublicID string
	Username string
	Role     entity.Role
}

type Meta struct {
	Error      *appErr.AppError `json:"error,omitempty" swaggertype:"primitive,object" extensions:"x-order=4"`
	Path       string           `json:"path" extensions:"x-order=0"`
	Status     string           `json:"status" extensions:"x-order=2"`
	Message    string           `json:"message" extensions:"x-order=3"`
	Timestamp  string           `json:"timestamp" extensions:"x-order=5"`
	StatusCode int              `json:"status_code" extensions:"x-order=1"`
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
	Dependencies map[string]string `json:"dependencies"`
	Status       string            `json:"status"`
	Timestamp    string            `json:"timestamp"`
}

// ToTokenResponse converts token.TokenDetails to a DTO for API responses
func ToTokenResponse(td *token.TokenDetails) TokenResponse {
	return TokenResponse{
		AccessToken:  td.AccessToken,
		RefreshToken: td.RefreshToken,
		ExpiresAt:    td.ExpiresAt,
	}
}
