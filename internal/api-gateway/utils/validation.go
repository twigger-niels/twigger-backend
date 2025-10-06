package utils

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// GetPathParam extracts a path parameter from the request
func GetPathParam(r *http.Request, key string) string {
	vars := mux.Vars(r)
	return vars[key]
}

// GetQueryParam extracts a query parameter from the request
func GetQueryParam(r *http.Request, key string) string {
	return r.URL.Query().Get(key)
}

// GetQueryParamInt extracts an integer query parameter with a default value
func GetQueryParamInt(r *http.Request, key string, defaultValue int) int {
	value := r.URL.Query().Get(key)
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return intValue
}

// GetQueryParamBool extracts a boolean query parameter with a default value
func GetQueryParamBool(r *http.Request, key string, defaultValue bool) bool {
	value := r.URL.Query().Get(key)
	if value == "" {
		return defaultValue
	}

	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}

	return boolValue
}

// ValidateLimit validates and clamps a limit parameter
func ValidateLimit(limit int, max int) int {
	if limit <= 0 {
		return 20 // default
	}
	if limit > max {
		return max
	}
	return limit
}

// ValidateUUID checks if a string is a valid UUID format (basic check)
func ValidateUUID(id string) error {
	if len(id) != 36 {
		return fmt.Errorf("invalid UUID format: must be 36 characters")
	}
	// Basic UUID format check (8-4-4-4-12)
	if id[8] != '-' || id[13] != '-' || id[18] != '-' || id[23] != '-' {
		return fmt.Errorf("invalid UUID format")
	}
	return nil
}

// ValidateRequired checks if a value is not empty
func ValidateRequired(value, field string) error {
	if value == "" {
		return fmt.Errorf("%s is required", field)
	}
	return nil
}

// ValidateCoordinates validates latitude and longitude
func ValidateCoordinates(lat, lng float64) error {
	if lat < -90 || lat > 90 {
		return fmt.Errorf("latitude must be between -90 and 90 degrees, got: %f", lat)
	}
	if lng < -180 || lng > 180 {
		return fmt.Errorf("longitude must be between -180 and 180 degrees, got: %f", lng)
	}
	return nil
}

// ParseFloat64 parses a string to float64 with error handling
func ParseFloat64(value string) (float64, error) {
	if value == "" {
		return 0, fmt.Errorf("value is empty")
	}
	return strconv.ParseFloat(value, 64)
}

// ParseFloat64OrDefault parses a string to float64 or returns default value
func ParseFloat64OrDefault(value string, defaultValue float64) float64 {
	if value == "" {
		return defaultValue
	}
	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return defaultValue
	}
	return f
}

// StringToBoolPtr converts a string to *bool for filter purposes
func StringToBoolPtr(value string) *bool {
	b, err := strconv.ParseBool(value)
	if err != nil {
		return nil
	}
	return &b
}
