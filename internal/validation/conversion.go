package validation

import (
	"strconv"
	"strings"
)

// GetIntFromInterface safely converts an interface value to int.
// It handles int, float64 (truncating), and string representations of integers.
func GetIntFromInterface(value interface{}) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case float64:
		// Allow conversion from float64, truncating the decimal part
		return int(v), true
	case string:
		// Attempt to parse string as int
		if i, err := strconv.Atoi(v); err == nil {
			return i, true
		}
	}
	return 0, false
}

// GetStringFromInterface safely converts an interface value to string.
func GetStringFromInterface(value interface{}) (string, bool) {
	if v, ok := value.(string); ok {
		return v, true
	}
	return "", false
}

// GetFloatFromInterface safely converts an interface value to float64.
// It handles float64, int, and string representations of floats.
func GetFloatFromInterface(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case int:
		// Allow conversion from int
		return float64(v), true
	case string:
		// Attempt to parse string as float
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

// GetBoolFromInterface safely converts an interface value to bool.
// It handles bool and string representations ("true", "false", case-insensitive).
func GetBoolFromInterface(value interface{}) (bool, bool) {
	switch v := value.(type) {
	case bool:
		return v, true
	case string:
		// Attempt to parse string as bool
		lowerV := strings.ToLower(v)
		if lowerV == "true" {
			return true, true
		}
		if lowerV == "false" {
			return false, true
		}
	}
	return false, false
}

// GetMapStringInterfaceFromInterface safely converts an interface value to map[string]interface{}.
func GetMapStringInterfaceFromInterface(value interface{}) (map[string]interface{}, bool) {
	if v, ok := value.(map[string]interface{}); ok {
		return v, true
	}
	return nil, false
}

// GetSliceMapStringInterfaceFromInterface safely converts an interface value to []map[string]interface{}.
// It handles []map[string]interface{} and []interface{} where elements are map[string]interface{}.
func GetSliceMapStringInterfaceFromInterface(value interface{}) ([]map[string]interface{}, bool) {
	if v, ok := value.([]map[string]interface{}); ok {
		return v, true
	}
	// Handle conversion from []interface{} if elements are map[string]interface{}
	if v, ok := value.([]interface{}); ok {
		result := make([]map[string]interface{}, 0, len(v))
		for _, item := range v {
			if mapItem, isMap := item.(map[string]interface{}); isMap {
				result = append(result, mapItem)
			} else {
				// If any element is not the correct type, the conversion fails
				return nil, false
			}
		}
		return result, true
	}
	return nil, false
}

// GetSliceInterfaceFromInterface safely converts an interface value to []interface{}.
func GetSliceInterfaceFromInterface(value interface{}) ([]interface{}, bool) {
	if v, ok := value.([]interface{}); ok {
		return v, true
	}
	return nil, false
}
