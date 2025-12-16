package dto

type Pagination struct {
	CurrentPage     int64   `json:"current_page" extensions:"x-order=0"`
	CurrentElements int64   `json:"current_elements" extensions:"x-order=1"`
	TotalPages      int64   `json:"total_pages" extensions:"x-order=2"`
	TotalElements   int64   `json:"total_elements" extensions:"x-order=3"`
	SortBy          string  `json:"sort_by" extensions:"x-order=4"`
	SortDir         string  `json:"sort_dir" extensions:"x-order=5"`
	CursorStart     *string `json:"cursor_start,omitempty" extensions:"x-order=6"`
	CursorEnd       *string `json:"cursor_end,omitempty" extensions:"x-order=7"`
}
