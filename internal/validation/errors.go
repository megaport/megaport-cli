package validation

import "fmt"

type ValidationError struct {
	Field  string
	Value  any
	Reason string
}

func NewValidationError(field string, value interface{}, reason string) *ValidationError {
	return &ValidationError{
		Field:  field,
		Value:  value,
		Reason: reason,
	}
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("Invalid %s: %v - %s", e.Field, e.Value, e.Reason)
}
