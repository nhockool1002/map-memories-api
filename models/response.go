package models

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Success    bool        `json:"success"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

// Pagination represents pagination metadata
type Pagination struct {
	CurrentPage int   `json:"current_page"`
	PerPage     int   `json:"per_page"`
	Total       int64 `json:"total"`
	TotalPages  int   `json:"total_pages"`
	HasNext     bool  `json:"has_next"`
	HasPrev     bool  `json:"has_prev"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// ValidationError represents validation errors
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   interface{} `json:"value,omitempty"`
}

// SuccessResponse creates a successful API response
func SuccessResponse(message string, data interface{}) APIResponse {
	return APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
}

// ErrorResponseWithCode creates an error API response with code
func ErrorResponseWithCode(message string, code string, details interface{}) APIResponse {
	return APIResponse{
		Success: false,
		Message: message,
		Error: ErrorResponse{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}

// PaginatedSuccessResponse creates a successful paginated API response
func PaginatedSuccessResponse(message string, data interface{}, pagination Pagination) PaginatedResponse {
	return PaginatedResponse{
		Success:    true,
		Message:    message,
		Data:       data,
		Pagination: pagination,
	}
}

// CalculatePagination calculates pagination metadata
func CalculatePagination(page, limit int, total int64) Pagination {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))
	hasNext := page < totalPages
	hasPrev := page > 1

	return Pagination{
		CurrentPage: page,
		PerPage:     limit,
		Total:       total,
		TotalPages:  totalPages,
		HasNext:     hasNext,
		HasPrev:     hasPrev,
	}
}