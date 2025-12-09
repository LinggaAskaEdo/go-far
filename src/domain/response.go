package domain

type Response struct {
	Success bool       `json:"success"`
	Data    any        `json:"data,omitempty"`
	Error   *ErrorData `json:"error,omitempty"`
}

type ErrorData struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type PaginatedResponse struct {
	Success bool     `json:"success"`
	Data    any      `json:"data"`
	Meta    MetaData `json:"meta"`
}

type MetaData struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
}

func SuccessResponse(data any) Response {
	return Response{
		Success: true,
		Data:    data,
	}
}

func ErrorResponse(code, message string) Response {
	return Response{
		Success: false,
		Error: &ErrorData{
			Code:    code,
			Message: message,
		},
	}
}

func PaginatedSuccessResponse(data any, meta MetaData) PaginatedResponse {
	return PaginatedResponse{
		Success: true,
		Data:    data,
		Meta:    meta,
	}
}
