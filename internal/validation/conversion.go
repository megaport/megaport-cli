package validation

import (
	"strconv"
	"strings"
)

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

func GetStringFromInterface(value interface{}) (string, bool) {
	if v, ok := value.(string); ok {
		return v, true
	}
	return "", false
}

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

func GetMapStringInterfaceFromInterface(value interface{}) (map[string]interface{}, bool) {
	if v, ok := value.(map[string]interface{}); ok {
		return v, true
	}
	return nil, false
}

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

func GetSliceInterfaceFromInterface(value interface{}) ([]interface{}, bool) {
	if v, ok := value.([]interface{}); ok {
		return v, true
	}
	return nil, false
}
