package dto

import "github.com/google/uuid"

// user related DTOs
type CreateUserRequest struct {
	Name  string `json:"name" binding:"required,min=2,max=100"`
	Email string `json:"email" binding:"required,email"`
	Age   int    `json:"age" binding:"required,min=1,max=150"`
}

type UpdateUserRequest struct {
	Name  string `json:"name" binding:"omitempty,min=2,max=100"`
	Email string `json:"email" binding:"omitempty,email"`
	Age   int    `json:"age" binding:"omitempty,min=1,max=150"`
}

type UserFilter struct {
	Name     string `form:"name"`
	Email    string `form:"email"`
	MinAge   int    `form:"min_age"`
	MaxAge   int    `form:"max_age"`
	Page     int64  `form:"page" binding:"min=1"`
	PageSize int64  `form:"page_size" binding:"min=1,max=100"`
	SortBy   string `form:"sort_by"`
	SortDir  string `form:"sort_dir" binding:"omitempty,oneof=asc desc"`
}

type CreateCarRequest struct {
	UserID       uuid.UUID `json:"user_id" binding:"required"`
	Brand        string    `json:"brand" binding:"required,min=2,max=100"`
	Model        string    `json:"model" binding:"required,min=2,max=100"`
	Year         int       `json:"year" binding:"required,gte=1900,lte=2100"`
	Color        string    `json:"color" binding:"omitempty,max=50"`
	LicensePlate string    `json:"license_plate" binding:"required,min=3,max=20"`
}

type BulkCreateCarsRequest struct {
	UserID uuid.UUID          `json:"user_id" binding:"required"`
	Cars   []CreateCarRequest `json:"cars" binding:"required,min=1,max=50,dive"`
}

type UpdateCarRequest struct {
	Brand        string `json:"brand" binding:"omitempty,min=2,max=100"`
	Model        string `json:"model" binding:"omitempty,min=2,max=100"`
	Year         int    `json:"year" binding:"omitempty,gte=1900,lte=2100"`
	Color        string `json:"color" binding:"omitempty,max=50"`
	LicensePlate string `json:"license_plate" binding:"omitempty,min=3,max=20"`
	IsAvailable  *bool  `json:"is_available" binding:"omitempty"`
}

type TransferCarRequest struct {
	NewUserID uuid.UUID `json:"new_user_id" binding:"required"`
}

type BulkUpdateAvailabilityRequest struct {
	CarIDs      []uuid.UUID `json:"car_ids" binding:"required,min=1"`
	IsAvailable bool        `json:"is_available"`
}
