package validation

import (
	"strconv"
	"strings"
)

// GetIntFromInterface attempts to convert an interface{} value to an int.
// Returns the converted int value and a boolean indicating success.
func GetIntFromInterface(value interface{}) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case float64:
		return int(v), true
	case string:
		if i, err := strconv.Atoi(v); err == nil {
			return i, true
		}
	}
	return 0, false
}

// GetStringFromInterface attempts to convert an interface{} value to a string.
// Returns the converted string value and a boolean indicating success.
func GetStringFromInterface(value interface{}) (string, bool) {
	if v, ok := value.(string); ok {
		return v, true
	}
	return "", false
}

// GetFloatFromInterface attempts to convert an interface{} value to a float64.
// Returns the converted float64 value and a boolean indicating success.
func GetFloatFromInterface(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

// GetBoolFromInterface attempts to convert an interface{} value to a bool.
// Returns the converted bool value and a boolean indicating success.
func GetBoolFromInterface(value interface{}) (bool, bool) {
	switch v := value.(type) {
	case bool:
		return v, true
	case string:
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

// GetMapStringInterfaceFromInterface attempts to convert an interface{} value to a map[string]interface{}.
// Returns the converted map and a boolean indicating success.
func GetMapStringInterfaceFromInterface(value interface{}) (map[string]interface{}, bool) {
	if v, ok := value.(map[string]interface{}); ok {
		return v, true
	}
	return nil, false
}

// GetSliceMapStringInterfaceFromInterface attempts to convert an interface{} value to a []map[string]interface{}.
// Handles both direct slice conversions and conversion of []interface{} where each element is a map.
// Returns the converted slice of maps and a boolean indicating success.
func GetSliceMapStringInterfaceFromInterface(value interface{}) ([]map[string]interface{}, bool) {
	if v, ok := value.([]map[string]interface{}); ok {
		return v, true
	}
	if v, ok := value.([]interface{}); ok {
		result := make([]map[string]interface{}, 0, len(v))
		for _, item := range v {
			if mapItem, isMap := item.(map[string]interface{}); isMap {
				result = append(result, mapItem)
			} else {
				return nil, false
			}
		}
		return result, true
	}
	return nil, false
}

// GetSliceInterfaceFromInterface attempts to convert an interface{} value to a []interface{}.
// Returns the converted slice and a boolean indicating success.
func GetSliceInterfaceFromInterface(value interface{}) ([]interface{}, bool) {
	if v, ok := value.([]interface{}); ok {
		return v, true
	}
	return nil, false
}
