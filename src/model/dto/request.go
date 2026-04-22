package dto

import (
	"go-far/src/model/entity"

	"github.com/google/uuid"
)

// Auth related DTOs
type RegisterRequest struct {
	Name     string      `json:"name" validate:"required,min=2,max=100"`
	Email    string      `json:"email" validate:"required,email"`
	Password string      `json:"password" validate:"required,min=8,max=100"`
	Role     entity.Role `json:"role" validate:"required,role_valid"`
	Age      int         `json:"age" validate:"required,min=1,max=150"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

// User related DTOs
type CreateUserRequest struct {
	Name     string      `json:"name" validate:"required,min=2,max=100"`
	Email    string      `json:"email" validate:"required,email"`
	Password string      `json:"password" validate:"required,min=8,max=100"`
	Role     entity.Role `json:"role" validate:"required,role_valid"`
	Age      int         `json:"age" validate:"required,min=1,max=150"`
}

type UpdateUserRequest struct {
	IsActive *bool       `json:"is_active" validate:"omitempty,boolean"`
	Name     string      `json:"name" validate:"omitempty,min=2,max=100"`
	Email    string      `json:"email" validate:"omitempty,email"`
	Role     entity.Role `json:"role" validate:"omitempty,role_valid"`
	Age      int         `json:"age" validate:"omitempty,min=1,max=150"`
}

type CreateCarRequest struct {
	Brand        string `json:"brand" validate:"required,min=2,max=100"`
	Model        string `json:"model" validate:"required,min=2,max=100"`
	Color        string `json:"color" validate:"omitempty,max=50"`
	LicensePlate string `json:"license_plate" validate:"required,min=3,max=20"`
	Year         int    `json:"year" validate:"required,gte=1900,lte=2100"`
}

type BulkCreateCarsRequest struct {
	Cars []CreateCarRequest `json:"cars" validate:"required,min=1,max=50,dive"`
}

type UpdateCarRequest struct {
	IsAvailable  *bool  `json:"is_available" validate:"omitempty,boolean"`
	Brand        string `json:"brand" validate:"omitempty,min=2,max=100"`
	Model        string `json:"model" validate:"omitempty,min=2,max=100"`
	Color        string `json:"color" validate:"omitempty,max=50"`
	LicensePlate string `json:"license_plate" validate:"omitempty,min=3,max=20"`
	Year         int    `json:"year" validate:"omitempty,gte=1900,lte=2100"`
}

type TransferCarRequest struct {
	NewUserID uuid.UUID `json:"new_user_id" validate:"required,uuid"`
}

type BulkUpdateAvailabilityRequest struct {
	CarIDs      []uuid.UUID `json:"car_ids" validate:"required,min=1"`
	IsAvailable bool        `json:"is_available" validate:"required,boolean"`
}
