package validation

import (
	"fmt"
	"net"
)

func IsValidationError(err error) bool {
	_, ok := err.(*ValidationError)
	return ok
}

func ValidateIntRange(value int, minValue int, maxValue int, fieldName string) error {
	if value < minValue || value > maxValue {
		return NewValidationError(fieldName, value,
			fmt.Sprintf("must be between %d-%d", minValue, maxValue))
	}
	return nil
}

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
