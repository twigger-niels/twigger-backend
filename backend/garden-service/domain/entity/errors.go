package entity

import "fmt"

// NotFoundError represents an error when a resource is not found
type NotFoundError struct {
	ResourceType string
	ResourceID   string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s not found: %s", e.ResourceType, e.ResourceID)
}

// NewNotFoundError creates a new NotFoundError
func NewNotFoundError(resourceType, resourceID string) error {
	return &NotFoundError{
		ResourceType: resourceType,
		ResourceID:   resourceID,
	}
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

// NewValidationError creates a new ValidationError
func NewValidationError(field, message string) error {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

// SpatialError represents an error with spatial data or operations
type SpatialError struct {
	Operation string
	Message   string
}

func (e *SpatialError) Error() string {
	return fmt.Sprintf("spatial error in %s: %s", e.Operation, e.Message)
}

// NewSpatialError creates a new SpatialError
func NewSpatialError(operation, message string) error {
	return &SpatialError{
		Operation: operation,
		Message:   message,
	}
}

// DatabaseError represents a database operation error
type DatabaseError struct {
	Operation string
	Err       error
}

func (e *DatabaseError) Error() string {
	return fmt.Sprintf("database error in %s: %v", e.Operation, e.Err)
}

func (e *DatabaseError) Unwrap() error {
	return e.Err
}

// NewDatabaseError creates a new DatabaseError
func NewDatabaseError(operation string, err error) error {
	return &DatabaseError{
		Operation: operation,
		Err:       err,
	}
}

// InvalidInputError represents an invalid input error
type InvalidInputError struct {
	Field   string
	Message string
}

func (e *InvalidInputError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("invalid input for field '%s': %s", e.Field, e.Message)
	}
	return fmt.Sprintf("invalid input: %s", e.Message)
}

// NewInvalidInputError creates a new InvalidInputError
func NewInvalidInputError(field, message string) error {
	return &InvalidInputError{
		Field:   field,
		Message: message,
	}
}
