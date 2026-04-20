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
	Age      int         `json:"age" validate:"required,min=1,max=150"`
	Role     entity.Role `json:"role" validate:"required,role_valid"`
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
	Age      int         `json:"age" validate:"required,min=1,max=150"`
	Role     entity.Role `json:"role" validate:"required,role_valid"`
}

type UpdateUserRequest struct {
	Name     string      `json:"name" validate:"omitempty,min=2,max=100"`
	Email    string      `json:"email" validate:"omitempty,email"`
	Age      int         `json:"age" validate:"omitempty,min=1,max=150"`
	Role     entity.Role `json:"role" validate:"omitempty,role_valid"`
	IsActive *bool       `json:"is_active" validate:"omitempty"`
}

type UserFilter struct {
	ID       string `form:"id"`
	Name     string `form:"name"`
	Email    string `form:"email"`
	MinAge   int    `form:"min_age"`
	MaxAge   int    `form:"max_age"`
	Page     int64  `form:"page" validate:"min=1"`
	PageSize int64  `form:"page_size" validate:"min=1,max=100"`
	SortBy   string `form:"sort_by"`
	SortDir  string `form:"sort_dir" validate:"omitempty,oneof=asc desc"`
}

type UserFilterV2 struct {
	ID       string `param:"id" db:"id"`
	Name     string `param:"name" db:"name"`
	Email    string `param:"email" db:"email"`
	MinAge   int    `param:"min_age__gte" db:"age"`
	MaxAge   int    `param:"max_age__lte" db:"age"`
	Page     int64  `param:"-"`
	PageSize int64  `param:"-"`
	SortBy   string `param:"-"`
	SortDir  string `param:"-"`
}

type CreateCarRequest struct {
	Brand        string `json:"brand" validate:"required,min=2,max=100"`
	Model        string `json:"model" validate:"required,min=2,max=100"`
	Year         int    `json:"year" validate:"required,gte=1900,lte=2100"`
	Color        string `json:"color" validate:"omitempty,max=50"`
	LicensePlate string `json:"license_plate" validate:"required,min=3,max=20"`
}

type BulkCreateCarsRequest struct {
	Cars []CreateCarRequest `json:"cars" validate:"required,min=1,max=50,dive"`
}

type UpdateCarRequest struct {
	Brand        string `json:"brand" validate:"omitempty,min=2,max=100"`
	Model        string `json:"model" validate:"omitempty,min=2,max=100"`
	Year         int    `json:"year" validate:"omitempty,gte=1900,lte=2100"`
	Color        string `json:"color" validate:"omitempty,max=50"`
	LicensePlate string `json:"license_plate" validate:"omitempty,min=3,max=20"`
	IsAvailable  *bool  `json:"is_available" validate:"omitempty"`
}

type TransferCarRequest struct {
	NewUserID uuid.UUID `json:"new_user_id" validate:"required"`
}

type BulkUpdateAvailabilityRequest struct {
	CarIDs      []uuid.UUID `json:"car_ids" validate:"required,min=1"`
	IsAvailable bool        `json:"is_available" validate:"required"`
}
