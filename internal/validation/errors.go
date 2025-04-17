package validation

import "fmt"

// ValidationError represents an error that occurs during validation of user input.
// It contains information about the field being validated, its value, and the reason for validation failure.
type ValidationError struct {
	Field  string
	Value  any
	Reason string
}

// NewValidationError creates a new ValidationError with the specified field name, value, and reason.
func NewValidationError(field string, value interface{}, reason string) *ValidationError {
	return &ValidationError{
		Field:  field,
		Value:  value,
		Reason: reason,
	}
}

// Error implements the error interface for ValidationError.
// Returns a formatted string with field name, value, and reason for the validation error.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("Invalid %s: %v - %s", e.Field, e.Value, e.Reason)
}
