package dto

type UserFilter struct {
	ID       string `form:"id"`
	Name     string `form:"name"`
	Email    string `form:"email"`
	SortBy   string `form:"sort_by"`
	SortDir  string `form:"sort_dir" validate:"omitempty,oneof=asc desc"`
	MinAge   int    `form:"min_age"`
	MaxAge   int    `form:"max_age"`
	Page     int64  `form:"page" validate:"min=1"`
	PageSize int64  `form:"page_size" validate:"min=1,max=100"`
}

func (f *UserFilter) NamePattern() string {
	return "%" + f.Name + "%"
}

func (f *UserFilter) EmailPattern() string {
	return "%" + f.Email + "%"
}

func (f *UserFilter) Limit() int64 {
	return f.PageSize
}

func (f *UserFilter) Offset() int64 {
	return (f.Page - 1) * f.PageSize
}

type UserFilterV2 struct {
	ID       string `param:"id" db:"id"`
	Name     string `param:"name" db:"name"`
	Email    string `param:"email" db:"email"`
	SortBy   string `param:"-"`
	SortDir  string `param:"-"`
	MinAge   int    `param:"min_age__gte" db:"age"`
	MaxAge   int    `param:"max_age__lte" db:"age"`
	Page     int64  `param:"-"`
	PageSize int64  `param:"-"`
}
