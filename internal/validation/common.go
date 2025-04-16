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

// VLAN Constants
const (
	MinVLAN           = 0    // Minimum possible VLAN ID (Untagged)
	MaxVLAN           = 4094 // Maximum possible VLAN ID (some systems use 4095, Megaport API seems to cap at 4094 for assignable)
	UntaggedVLAN      = -1   // Represents an untagged VLAN
	AutoAssignVLAN    = 0    // Represents automatic VLAN assignment
	ReservedVLAN      = 1    // Reserved VLAN ID (often not assignable)
	MinAssignableVLAN = 2    // Minimum VLAN ID that can typically be assigned by a user
	MaxAssignableVLAN = 4093 // Maximum VLAN ID that can typically be assigned by a user (4094 often reserved)

	// Length/size constants
	MaxPortNameLength          = 64
	MaxAWSConnectionNameLength = 255
	MaxMVENameLength           = 64
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

// ValidateVLAN checks if a VLAN ID is valid for general use cases.
// Allows AutoAssignVLAN (-1), UntaggedVLAN (0), or assignable range (2-4094).
func ValidateVLAN(vlan int) error {
	if vlan == AutoAssignVLAN || vlan == UntaggedVLAN || (vlan >= MinAssignableVLAN && vlan <= MaxVLAN) {
		return nil
	}
	return NewValidationError("VLAN ID", vlan, fmt.Sprintf("must be %d, %d, or between %d-%d", AutoAssignVLAN, UntaggedVLAN, MinAssignableVLAN, MaxVLAN))
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
