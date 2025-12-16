package dto

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
