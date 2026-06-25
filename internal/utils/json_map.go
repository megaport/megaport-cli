package utils

import "fmt"

// These helpers read typed values out of a JSON object decoded into
// map[string]interface{}. An absent key is reported via present=false and is
// never an error, so callers keep optional fields optional. A key that is
// present but holds the wrong JSON type returns an error instead of being
// silently dropped, matching how the struct-based decoders surface a
// json.UnmarshalTypeError.

// JSONString returns the string at m[key].
func JSONString(m map[string]interface{}, key string) (value string, present bool, err error) {
	raw, ok := m[key]
	if !ok {
		return "", false, nil
	}
	s, ok := raw.(string)
	if !ok {
		return "", true, fmt.Errorf("%s must be a string", key)
	}
	return s, true, nil
}

// JSONNumber returns the number at m[key]. JSON numbers decode to float64.
func JSONNumber(m map[string]interface{}, key string) (value float64, present bool, err error) {
	raw, ok := m[key]
	if !ok {
		return 0, false, nil
	}
	f, ok := raw.(float64)
	if !ok {
		return 0, true, fmt.Errorf("%s must be a number", key)
	}
	return f, true, nil
}

// JSONBool returns the boolean at m[key].
func JSONBool(m map[string]interface{}, key string) (value bool, present bool, err error) {
	raw, ok := m[key]
	if !ok {
		return false, false, nil
	}
	b, ok := raw.(bool)
	if !ok {
		return false, true, fmt.Errorf("%s must be a boolean", key)
	}
	return b, true, nil
}

// JSONObject returns the nested object at m[key].
func JSONObject(m map[string]interface{}, key string) (value map[string]interface{}, present bool, err error) {
	raw, ok := m[key]
	if !ok {
		return nil, false, nil
	}
	obj, ok := raw.(map[string]interface{})
	if !ok {
		return nil, true, fmt.Errorf("%s must be an object", key)
	}
	return obj, true, nil
}

// JSONArray returns the array at m[key].
func JSONArray(m map[string]interface{}, key string) (value []interface{}, present bool, err error) {
	raw, ok := m[key]
	if !ok {
		return nil, false, nil
	}
	arr, ok := raw.([]interface{})
	if !ok {
		return nil, true, fmt.Errorf("%s must be an array", key)
	}
	return arr, true, nil
}
