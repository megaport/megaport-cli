package validation

import (
	"fmt"
)

// Constants for valid values
var (
	ValidContractTerms   = []int{1, 12, 24, 36}
	ValidMCRPortSpeeds   = []int{1000, 2500, 5000, 10000, 25000, 50000, 100000}
	ValidPortSpeeds      = []int{1000, 10000, 100000}
	ValidMVEProductSizes = []string{"SMALL", "MEDIUM", "LARGE"}
)

// ValidateContractTerm validates a contract term
func ValidateContractTerm(term int) error {
	for _, validTerm := range ValidContractTerms {
		if term == validTerm {
			return nil
		}
	}
	return NewValidationError("contract term", term,
		fmt.Sprintf("must be one of: %v", ValidContractTerms))
}

// ValidateMCRPortSpeed validates an MCR port speed
func ValidateMCRPortSpeed(speed int) error {
	for _, validSpeed := range ValidMCRPortSpeeds {
		if speed == validSpeed {
			return nil
		}
	}
	return NewValidationError("MCR port speed", speed,
		fmt.Sprintf("must be one of: %v", ValidMCRPortSpeeds))
}

// ValidatePortSpeed validates a port speed
func ValidatePortSpeed(speed int) error {
	for _, validSpeed := range ValidPortSpeeds {
		if speed == validSpeed {
			return nil
		}
	}
	return NewValidationError("port speed", speed,
		fmt.Sprintf("must be one of: %v", ValidPortSpeeds))
}

// ValidateVLAN validates if the VLAN is within the allowed range.
// A valid VLAN can be:
// -1 (untagged)
// 0 (auto-assigned)
// 2-4093 (actual VLAN values, with 1 being reserved)
func ValidateVLAN(vlan int) error {
	if vlan == -1 || vlan == 0 || (vlan >= 2 && vlan <= 4093) {
		return nil
	}
	return NewValidationError("VLAN", vlan,
		"must be -1 (untagged), 0 (auto-assigned), or between 2-4093 (1 is reserved)")
}

// ValidateRateLimit validates a VXC rate limit
func ValidateRateLimit(rateLimit int) error {
	if rateLimit <= 0 {
		return NewValidationError("rate limit", rateLimit, "must be a positive integer")
	}
	return nil
}

// ValidateMVEProductSize validates an MVE product size
func ValidateMVEProductSize(size string) error {
	for _, validSize := range ValidMVEProductSizes {
		if size == validSize {
			return nil
		}
	}
	return NewValidationError("MVE product size", size,
		fmt.Sprintf("must be one of: %v", ValidMVEProductSizes))
}
