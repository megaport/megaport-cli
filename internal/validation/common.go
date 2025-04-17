// Package validation provides validation functions for Megaport resources and configurations.
// It contains utilities to validate inputs for Megaport API calls, ensuring that input parameters
// meet required criteria before they are submitted to the API.
package validation

import (
	"fmt"
)

var (
	// ValidContractTerms lists the allowed contract term durations in months.
	ValidContractTerms = []int{1, 12, 24, 36}
	// ValidMCRPortSpeeds lists the supported MCR port speeds in Mbps.
	ValidMCRPortSpeeds = []int{1000, 2500, 5000, 10000, 25000, 50000, 100000}
	// ValidPortSpeeds lists the supported port speeds in Mbps.
	ValidPortSpeeds = []int{1000, 10000, 100000}
	// ValidMVEProductSizes lists the supported MVE product sizes.
	ValidMVEProductSizes = []string{"SMALL", "MEDIUM", "LARGE", "X_LARGE_12"}
)

const (
	// MaxVLAN is the maximum VLAN ID.
	MaxVLAN = 4094
	// UntaggedVLAN indicates a packet should be untagged.
	UntaggedVLAN = -1
	// AutoAssignVLAN indicates the system should auto-assign a VLAN.
	// This was previously known as MinVLAN.
	AutoAssignVLAN = 0
	// ReservedVLAN is a VLAN reserved by the system.
	ReservedVLAN = 1
	// MinAssignableVLAN is the lowest VLAN ID assignable to a user.
	MinAssignableVLAN = 2
	// MaxAssignableVLAN is the highest VLAN ID assignable to a user.
	MaxAssignableVLAN = 4093
	// MaxPortNameLength is the maximum allowed length of a port name.
	MaxPortNameLength = 64
	// MaxAWSConnectionNameLength is the maximum allowed length of an AWS connection name.
	MaxAWSConnectionNameLength = 255
	// MaxMVENameLength is the maximum allowed length of an MVE name.
	MaxMVENameLength = 64
)

// ValidateContractTerm validates if a contract term is one of the allowed values.
// Returns an error if the term is not a valid contract term duration.
func ValidateContractTerm(term int) error {
	for _, validTerm := range ValidContractTerms {
		if term == validTerm {
			return nil
		}
	}
	return NewValidationError("contract term", term,
		fmt.Sprintf("must be one of: %v", ValidContractTerms))
}

// ValidateMCRPortSpeed validates if a port speed is one of the allowed values for MCR.
// Returns an error if the speed is not a valid MCR port speed.
func ValidateMCRPortSpeed(speed int) error {
	for _, validSpeed := range ValidMCRPortSpeeds {
		if speed == validSpeed {
			return nil
		}
	}
	return NewValidationError("MCR port speed", speed,
		fmt.Sprintf("must be one of: %v", ValidMCRPortSpeeds))
}

// ValidatePortSpeed validates if a port speed is one of the allowed values for ports.
// Returns an error if the speed is not a valid port speed.
func ValidatePortSpeed(speed int) error {
	for _, validSpeed := range ValidPortSpeeds {
		if speed == validSpeed {
			return nil
		}
	}
	return NewValidationError("port speed", speed,
		fmt.Sprintf("must be one of: %v", ValidPortSpeeds))
}

// ValidateVLAN validates if a VLAN ID is valid.
// Valid values include:
// - AutoAssignVLAN (0): System will auto-assign a VLAN
// - UntaggedVLAN (-1): Packet will be untagged
// - Values between MinAssignableVLAN and MaxVLAN (inclusive)
// Returns an error if the VLAN ID is not valid.
func ValidateVLAN(vlan int) error {
	if vlan == AutoAssignVLAN || vlan == UntaggedVLAN || (vlan >= MinAssignableVLAN && vlan <= MaxVLAN) {
		return nil
	}
	return NewValidationError("VLAN ID", vlan, fmt.Sprintf("must be %d, %d, or between %d-%d", AutoAssignVLAN, UntaggedVLAN, MinAssignableVLAN, MaxVLAN))
}

// ValidateRateLimit validates if a rate limit is a positive integer.
// Returns an error if the rate limit is not a positive integer.
func ValidateRateLimit(rateLimit int) error {
	if rateLimit <= 0 {
		return NewValidationError("rate limit", rateLimit, "must be a positive integer")
	}
	return nil
}

// ExtractFieldsWithTypes extracts fields from a configuration map according to their expected types.
// This helper function is used to convert untyped map data to correctly typed fields.
func ExtractFieldsWithTypes(config map[string]interface{}, fields map[string]string) map[string]interface{} {
	result := make(map[string]interface{})
	for field, fieldType := range fields {
		switch fieldType {
		case "string":
			val, _ := GetStringFromInterface(config[field])
			result[field] = val
		case "int":
			val, _ := GetIntFromInterface(config[field])
			result[field] = val
		case "bool":
			val, _ := GetBoolFromInterface(config[field])
			result[field] = val
		case "string_slice":
			val, _ := GetSliceInterfaceFromInterface(config[field])
			result[field] = val
		case "map_slice":
			val, _ := GetSliceMapStringInterfaceFromInterface(config[field])
			result[field] = val
		}
	}
	return result
}
