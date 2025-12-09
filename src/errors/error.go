package exception

import (
	"fmt"
	"net/http"
)

type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func NewAppError(code, message string, status int) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Status:  status,
	}
}

var (
	ErrNotFound = NewAppError(
		"NOT_FOUND",
		"Resource not found",
		http.StatusNotFound,
	)

	ErrBadRequest = NewAppError(
		"BAD_REQUEST",
		"Invalid request",
		http.StatusBadRequest,
	)

	ErrInternalServer = NewAppError(
		"INTERNAL_SERVER_ERROR",
		"Internal server error",
		http.StatusInternalServerError,
	)

	ErrUnauthorized = NewAppError(
		"UNAUTHORIZED",
		"Unauthorized access",
		http.StatusUnauthorized,
	)

	ErrConflict = NewAppError(
		"CONFLICT",
		"Resource already exists",
		http.StatusConflict,
	)
)

func NotFoundError(message string) *AppError {
	return NewAppError("NOT_FOUND", message, http.StatusNotFound)
}

func BadRequestError(message string) *AppError {
	return NewAppError("BAD_REQUEST", message, http.StatusBadRequest)
}

func InternalServerError(message string) *AppError {
	return NewAppError("INTERNAL_SERVER_ERROR", message, http.StatusInternalServerError)
}

func ConflictError(message string) *AppError {
	return NewAppError("CONFLICT", message, http.StatusConflict)
}

func UnauthorizedError(message string) *AppError {
	return NewAppError("UNAUTHORIZED", message, http.StatusUnauthorized)
}
