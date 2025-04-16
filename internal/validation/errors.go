package validation

import "fmt"

// Error types and helper functions for validation

// ValidationError represents a validation failure with context
type ValidationError struct {
	Field  string
	Value  interface{}
	Reason string
}

// NewValidationError creates a new validation error
func NewValidationError(field string, value interface{}, reason string) *ValidationError {
	return &ValidationError{
		Field:  field,
		Value:  value,
		Reason: reason,
	}
}

// Error implements the error interface with a consistent format
func (e *ValidationError) Error() string {
	return fmt.Sprintf("Invalid %s: %v - %s", e.Field, e.Value, e.Reason)
}
