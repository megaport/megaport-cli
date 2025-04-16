package validation

import (
	"fmt"
)

var (
	ValidContractTerms   = []int{1, 12, 24, 36}
	ValidMCRPortSpeeds   = []int{1000, 2500, 5000, 10000, 25000, 50000, 100000}
	ValidPortSpeeds      = []int{1000, 10000, 100000}
	ValidMVEProductSizes = []string{"SMALL", "MEDIUM", "LARGE"}
)

const (
	MinVLAN                    = 0
	MaxVLAN                    = 4094
	UntaggedVLAN               = -1
	AutoAssignVLAN             = 0
	ReservedVLAN               = 1
	MinAssignableVLAN          = 2
	MaxAssignableVLAN          = 4093
	MaxPortNameLength          = 64
	MaxAWSConnectionNameLength = 255
	MaxMVENameLength           = 64
)

func ValidateContractTerm(term int) error {
	for _, validTerm := range ValidContractTerms {
		if term == validTerm {
			return nil
		}
	}
	return NewValidationError("contract term", term,
		fmt.Sprintf("must be one of: %v", ValidContractTerms))
}

func ValidateMCRPortSpeed(speed int) error {
	for _, validSpeed := range ValidMCRPortSpeeds {
		if speed == validSpeed {
			return nil
		}
	}
	return NewValidationError("MCR port speed", speed,
		fmt.Sprintf("must be one of: %v", ValidMCRPortSpeeds))
}

func ValidatePortSpeed(speed int) error {
	for _, validSpeed := range ValidPortSpeeds {
		if speed == validSpeed {
			return nil
		}
	}
	return NewValidationError("port speed", speed,
		fmt.Sprintf("must be one of: %v", ValidPortSpeeds))
}

func ValidateVLAN(vlan int) error {
	if vlan == AutoAssignVLAN || vlan == UntaggedVLAN || (vlan >= MinAssignableVLAN && vlan <= MaxVLAN) {
		return nil
	}
	return NewValidationError("VLAN ID", vlan, fmt.Sprintf("must be %d, %d, or between %d-%d", AutoAssignVLAN, UntaggedVLAN, MinAssignableVLAN, MaxVLAN))
}

func ValidateRateLimit(rateLimit int) error {
	if rateLimit <= 0 {
		return NewValidationError("rate limit", rateLimit, "must be a positive integer")
	}
	return nil
}

func ValidateMVEProductSize(size string) error {
	for _, validSize := range ValidMVEProductSizes {
		if size == validSize {
			return nil
		}
	}
	return NewValidationError("MVE product size", size,
		fmt.Sprintf("must be one of: %v", ValidMVEProductSizes))
}

func ValidateFieldPresence(config map[string]interface{}, requiredFields []string) string {
	for _, field := range requiredFields {
		val, exists := config[field]
		if !exists || val == nil {
			return field
		}
		if strVal, isStr := val.(string); isStr && strVal == "" {
			return field
		}
	}
	return ""
}

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
