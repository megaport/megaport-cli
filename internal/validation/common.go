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
	// MaxPortNameLength defines the maximum length of a port name in characters.
	MaxPortNameLength = 64
	// MaxAWSConnectionNameLength is the maximum allowed length of an AWS connection name.
	MaxAWSConnectionNameLength = 255
	// MaxMVENameLength defines the maximum length of an MVE name in characters.
	MaxMVENameLength = 64
	// AutoAssignVLAN indicates the VLAN should be automatically assigned by the system.
	AutoAssignVLAN = 0
	// UntaggedVLAN indicates packets should be untagged (no VLAN tag).
	UntaggedVLAN = -1
	// MinAssignableVLAN is the lowest VLAN ID that can be assigned to traffic.
	MinAssignableVLAN = 2
	// MaxAssignableVLAN is the highest VLAN ID that can typically be assigned by users.
	MaxAssignableVLAN = 4093
	// MaxVLAN is the maximum possible VLAN ID according to IEEE 802.1Q.
	MaxVLAN = 4094
	// ReservedVLAN identifies a VLAN ID that is reserved and cannot be used.
	ReservedVLAN = 1
)

// ValidateContractTerm validates if a contract term is one of the allowed values.
// Contract terms define the duration of the service commitment in months.
//
// Parameters:
//   - term: The contract term in months to validate
//
// Validation checks:
//   - Term must be one of the predefined valid values (ValidContractTerms)
//   - Typically valid values are 1, 12, 24, or 36 months
//
// Returns:
//   - A ValidationError if the term is not valid
//   - nil if the validation passes
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
// This function ensures that the specified port speed is supported for Megaport Cloud Routers.
//
// Parameters:
//   - speed: The port speed in Mbps to validate
//
// Validation checks:
//   - Speed must be one of the predefined valid values (ValidMCRPortSpeeds)
//   - Typically valid values are 1000, 2500, 5000, 10000, 25000, 50000, or 100000 Mbps
//
// Returns:
//   - A ValidationError if the speed is not valid
//   - nil if the validation passes
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
// This function ensures that the specified port speed is supported for Megaport physical ports.
//
// Parameters:
//   - speed: The port speed in Mbps to validate
//
// Validation checks:
//   - Speed must be one of the predefined valid values (ValidPortSpeeds)
//   - Typically valid values are 1000, 10000, or 100000 Mbps
//
// Returns:
//   - A ValidationError if the speed is not valid
//   - nil if the validation passes
func ValidatePortSpeed(speed int) error {
	for _, validSpeed := range ValidPortSpeeds {
		if speed == validSpeed {
			return nil
		}
	}
	return NewValidationError("port speed", speed,
		fmt.Sprintf("must be one of: %v", ValidPortSpeeds))
}

// ValidateVLAN validates if a VLAN ID is valid for use in Megaport configurations.
// This function ensures the VLAN ID meets the requirements of IEEE 802.1Q standards and Megaport-specific constraints.
//
// Parameters:
//   - vlan: The VLAN ID to validate
//
// Validation checks:
//   - VLAN must be one of the following:
//   - AutoAssignVLAN (0): System will auto-assign a VLAN
//   - UntaggedVLAN (-1): Packet will be untagged
//   - A value between MinAssignableVLAN (2) and MaxVLAN (4094) inclusive
//
// Returns:
//   - A ValidationError if the VLAN ID is not valid
//   - nil if the validation passes
func ValidateVLAN(vlan int) error {
	if vlan == AutoAssignVLAN || vlan == UntaggedVLAN || (vlan >= MinAssignableVLAN && vlan <= MaxVLAN) {
		return nil
	}
	return NewValidationError("VLAN ID", vlan, fmt.Sprintf("must be %d, %d, or between %d-%d", AutoAssignVLAN, UntaggedVLAN, MinAssignableVLAN, MaxVLAN))
}

// ValidateRateLimit validates if a rate limit is a positive integer.
// This function ensures the rate limit value is valid for bandwidth constraints.
//
// Parameters:
//   - rateLimit: The rate limit in Mbps to validate
//
// Validation checks:
//   - Rate limit must be a positive integer (greater than zero)
//   - Rate limit represents bandwidth in Mbps
//
// Returns:
//   - A ValidationError if the rate limit is not valid
//   - nil if the validation passes
func ValidateRateLimit(rateLimit int) error {
	if rateLimit <= 0 {
		return NewValidationError("rate limit", rateLimit, "must be a positive integer")
	}
	return nil
}

// ExtractFieldsWithTypes extracts fields from a configuration map according to their expected types.
// This helper function is used to convert untyped map data (typically from JSON deserialization)
// to correctly typed fields for further processing or validation. It handles type conversion
// intelligently based on the specified expected types.
//
// Parameters:
//   - config: A map containing mixed type values, typically from JSON deserialization
//   - fields: A map where key is the field name and value is the expected type name
//
// Supported type names in the fields map:
//   - "string": Extracts the value as a string
//   - "int": Extracts the value as an integer
//   - "bool": Extracts the value as a boolean
//   - "string_slice": Extracts the value as a slice of strings/interfaces
//   - "map_slice": Extracts the value as a slice of map[string]interface{}
//
// The function calls the appropriate type conversion helper function for each field
// based on the specified expected type.
//
// Returns:
//   - A new map with the extracted values, correctly typed according to the fields map
//
// Example:
//
//	config := map[string]interface{}{
//	    "name": "test",
//	    "port": 8080,
//	    "enabled": true,
//	    "tags": []interface{}{"tag1", "tag2"},
//	}
//	fields := map[string]string{
//	    "name": "string",
//	    "port": "int",
//	    "enabled": "bool",
//	    "tags": "string_slice",
//	}
//	result := ExtractFieldsWithTypes(config, fields)
//	// result will contain the extracted values with proper types
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
