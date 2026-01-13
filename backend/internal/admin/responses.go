package admin

import (
	"encoding/json"
	"net/http"
)

// ApiResponse is a standard API response wrapper
type ApiResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// PaginatedResponse wraps paginated results
type PaginatedResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Page    int         `json:"page"`
	Pages   int         `json:"pages"`
	Total   int64       `json:"total"`
	Limit   int         `json:"limit"`
	Error   string      `json:"error,omitempty"`
}

// WriteJSON writes a JSON response to the response writer
func WriteJSON(w http.ResponseWriter, statusCode int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(data)
}

// WriteError writes an error response
func WriteError(w http.ResponseWriter, statusCode int, message string) error {
	return WriteJSON(w, statusCode, ApiResponse{
		Success: false,
		Error:   message,
	})
}

// WriteSuccess writes a success response
func WriteSuccess(w http.ResponseWriter, statusCode int, data interface{}) error {
	return WriteJSON(w, statusCode, ApiResponse{
		Success: true,
		Data:    data,
	})
}

// WritePaginated writes a paginated response
func WritePaginated(w http.ResponseWriter, statusCode int, data interface{}, page, pages int, total int64, limit int) error {
	return WriteJSON(w, statusCode, PaginatedResponse{
		Success: true,
		Data:    data,
		Page:    page,
		Pages:   pages,
		Total:   total,
		Limit:   limit,
	})
}
