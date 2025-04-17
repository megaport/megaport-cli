package validation

import (
	"fmt"
	"net"
)

// IsValidationError checks if an error is an instance of ValidationError.
// Returns true if the error is a ValidationError, false otherwise.
func IsValidationError(err error) bool {
	_, ok := err.(*ValidationError)
	return ok
}

// ValidateIntRange validates that an integer value falls within the specified range.
// Returns a ValidationError if the value is outside the range, nil otherwise.
func ValidateIntRange(value int, minValue int, maxValue int, fieldName string) error {
	if value < minValue || value > maxValue {
		return NewValidationError(fieldName, value,
			fmt.Sprintf("must be between %d-%d", minValue, maxValue))
	}
	return nil
}

// ValidateStringOneOf validates that a string value is one of the allowed values.
// Returns a ValidationError if the string is empty or not in the list of valid values.
func ValidateStringOneOf(value string, validValues []string, fieldName string) error {
	if value == "" {
		return NewValidationError(fieldName, value, "cannot be empty")
	}
	for _, validValue := range validValues {
		if value == validValue {
			return nil
		}
	}
	return NewValidationError(fieldName, value,
		fmt.Sprintf("must be one of: %v", validValues))
}

// ValidateIPv4 validates that a string is a valid IPv4 address.
// Returns a ValidationError if the string is empty or not a valid IPv4 address.
func ValidateIPv4(ip string, fieldName string) error {
	if ip == "" {
		return NewValidationError(fieldName, ip, "cannot be empty")
	}
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil || parsedIP.To4() == nil {
		return NewValidationError(fieldName, ip, "must be a valid IPv4 address")
	}
	return nil
}

// ValidateCIDR validates that a string is in valid CIDR notation.
// Returns a ValidationError if the string is empty or not in valid CIDR format.
func ValidateCIDR(cidr string, fieldName string) error {
	if cidr == "" {
		return NewValidationError(fieldName, cidr, "cannot be empty")
	}
	_, _, err := net.ParseCIDR(cidr)
	if err != nil {
		return NewValidationError(fieldName, cidr, "must be a valid CIDR notation")
	}
	return nil
}
