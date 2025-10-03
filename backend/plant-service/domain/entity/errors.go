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

// NotFoundError represents an error when a resource cannot be found
type NotFoundError struct {
	Resource string
	ID       string
}

// Error implements the error interface
func (e *NotFoundError) Error() string {
	if e.ID != "" {
		return fmt.Sprintf("%s not found: %s", e.Resource, e.ID)
	}
	return fmt.Sprintf("%s not found", e.Resource)
}

// NewNotFoundError creates a new NotFoundError
func NewNotFoundError(resource, id string) *NotFoundError {
	return &NotFoundError{
		Resource: resource,
		ID:       id,
	}
}

// Is implements error comparison for errors.Is
func (e *NotFoundError) Is(target error) bool {
	_, ok := target.(*NotFoundError)
	return ok
}

// DatabaseError represents a database-related error
type DatabaseError struct {
	Operation string
	Cause     error
}

// Error implements the error interface
func (e *DatabaseError) Error() string {
	return fmt.Sprintf("database error during %s: %v", e.Operation, e.Cause)
}

// Unwrap implements error unwrapping
func (e *DatabaseError) Unwrap() error {
	return e.Cause
}

// NewDatabaseError creates a new DatabaseError
func NewDatabaseError(operation string, cause error) *DatabaseError {
	return &DatabaseError{
		Operation: operation,
		Cause:     cause,
	}
}

// InvalidInputError represents an error when input validation fails
type InvalidInputError struct {
	Field   string
	Value   interface{}
	Message string
}

// Error implements the error interface
func (e *InvalidInputError) Error() string {
	if e.Value != nil {
		return fmt.Sprintf("invalid input for field '%s' (value: %v): %s", e.Field, e.Value, e.Message)
	}
	return fmt.Sprintf("invalid input for field '%s': %s", e.Field, e.Message)
}

// NewInvalidInputError creates a new InvalidInputError
func NewInvalidInputError(field string, value interface{}, message string) *InvalidInputError {
	return &InvalidInputError{
		Field:   field,
		Value:   value,
		Message: message,
	}
}

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
