package utils

import (
	"encoding/json"
	"net/http"
	"strings"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error   string                 `json:"error"`
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// SuccessResponse represents a standardized success response
type SuccessResponse struct {
	Data interface{} `json:"data"`
	Meta *Meta       `json:"meta,omitempty"`
}

// Meta contains pagination and additional metadata
type Meta struct {
	Cursor    *string `json:"cursor,omitempty"`
	HasMore   bool    `json:"has_more"`
	Limit     int     `json:"limit"`
	Total     *int    `json:"total,omitempty"`
}

// RespondJSON writes a JSON response
func RespondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// RespondSuccess writes a success response with optional metadata
func RespondSuccess(w http.ResponseWriter, data interface{}, meta *Meta) {
	RespondJSON(w, http.StatusOK, SuccessResponse{
		Data: data,
		Meta: meta,
	})
}

// RespondError writes an error response
func RespondError(w http.ResponseWriter, err error) {
	statusCode, errorResp := MapError(err)
	RespondJSON(w, statusCode, errorResp)
}

// RespondCreated writes a 201 Created response
func RespondCreated(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(SuccessResponse{Data: data}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// RespondNoContent writes a 204 No Content response
func RespondNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// MapError maps domain errors to HTTP status codes and error responses
func MapError(err error) (int, ErrorResponse) {
	if err == nil {
		return http.StatusOK, ErrorResponse{}
	}

	errMsg := err.Error()
	errLower := strings.ToLower(errMsg)

	// Check for not found errors
	if strings.Contains(errLower, "not found") || strings.Contains(errLower, "does not exist") {
		return http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Code:    "RESOURCE_NOT_FOUND",
			Message: errMsg,
		}
	}

	// Check for validation errors
	if strings.Contains(errLower, "validation") || strings.Contains(errLower, "invalid") {
		return http.StatusBadRequest, ErrorResponse{
			Error:   "validation_error",
			Code:    "VALIDATION_ERROR",
			Message: errMsg,
		}
	}

	// Check for database errors
	if strings.Contains(errLower, "database") || strings.Contains(errLower, "sql") {
		return http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Code:    "DATABASE_ERROR",
			Message: "An internal error occurred",
			Details: map[string]interface{}{
				"type": "database",
			},
		}
	}

	// Default to internal server error
	return http.StatusInternalServerError, ErrorResponse{
		Error:   "internal_error",
		Code:    "INTERNAL_SERVER_ERROR",
		Message: "An unexpected error occurred",
	}
}

// RespondValidationError writes a validation error response
func RespondValidationError(w http.ResponseWriter, field, message string) {
	RespondJSON(w, http.StatusBadRequest, ErrorResponse{
		Error:   "validation_error",
		Code:    "VALIDATION_ERROR",
		Message: message,
		Details: map[string]interface{}{
			"field": field,
		},
	})
}

// RespondUnauthorized writes a 401 Unauthorized response
func RespondUnauthorized(w http.ResponseWriter, message string) {
	RespondJSON(w, http.StatusUnauthorized, ErrorResponse{
		Error:   "unauthorized",
		Code:    "UNAUTHORIZED",
		Message: message,
	})
}

// RespondForbidden writes a 403 Forbidden response
func RespondForbidden(w http.ResponseWriter, message string) {
	RespondJSON(w, http.StatusForbidden, ErrorResponse{
		Error:   "forbidden",
		Code:    "FORBIDDEN",
		Message: message,
	})
}
