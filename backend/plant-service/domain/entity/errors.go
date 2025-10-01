package entity

import "fmt"

// Domain errors
var (
	// ErrPlantNotFound is returned when a plant cannot be found
	ErrPlantNotFound = fmt.Errorf("plant not found")

	// ErrPlantAlreadyExists is returned when attempting to create a duplicate plant
	ErrPlantAlreadyExists = fmt.Errorf("plant already exists")

	// ErrInvalidPlantID is returned when a plant ID is invalid
	ErrInvalidPlantID = fmt.Errorf("invalid plant ID")

	// ErrInvalidSpeciesID is returned when a species ID is invalid
	ErrInvalidSpeciesID = fmt.Errorf("invalid species ID")

	// ErrCompanionNotFound is returned when a companion relationship cannot be found
	ErrCompanionNotFound = fmt.Errorf("companion relationship not found")

	// ErrInvalidSearchQuery is returned when a search query is invalid
	ErrInvalidSearchQuery = fmt.Errorf("invalid search query")

	// ErrInvalidFilter is returned when a filter parameter is invalid
	ErrInvalidFilter = fmt.Errorf("invalid filter")

	// ErrDatabaseConnection is returned when database connection fails
	ErrDatabaseConnection = fmt.Errorf("database connection error")
)

// ValidationError represents a validation error with context
type ValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}
