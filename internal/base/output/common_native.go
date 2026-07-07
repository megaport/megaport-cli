//go:build !wasm

package output

import (
	"reflect"
	"strings"
)

// extractCSVFieldInfo extracts CSV-specific field metadata from the first element of data.
// CSV uses the csv tag as the header name (falling back to json tag), and skips fields
// that have neither a csv nor json tag. This differs from extractFieldInfo, which does
// not require csv/json tags and instead falls back to header/csv/output tags or the field name.
func extractCSVFieldInfo[T OutputFields](data []T) (headers, jsonNames []string, fieldIndices []int, err error) {
	var sample T
	if len(data) > 0 {
		sample = data[0]
	}
	sampleVal := reflect.ValueOf(sample)
	if !sampleVal.IsValid() {
		return nil, nil, nil, nil
	}
	t := sampleVal.Type()
	if t.Kind() == reflect.Pointer {
		if sampleVal.IsNil() {
			if t.Elem().Kind() != reflect.Struct {
				return nil, nil, nil, nil
			}
			t = t.Elem()
		} else {
			t = sampleVal.Elem().Type()
		}
	}
	if t.Kind() != reflect.Struct {
		return nil, nil, nil, nil
	}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue
		}
		if !isOutputCompatibleType(field.Type) {
			continue
		}
		csvTag := field.Tag.Get("csv")
		if csvTag == "-" {
			continue
		}
		jsonTag := field.Tag.Get("json")
		// Strip json tag options (e.g. "name,omitempty" -> "name") before
		// using as a fallback header or for --fields matching.
		jsonName := jsonTag
		if idx := strings.Index(jsonName, ","); idx != -1 {
			jsonName = jsonName[:idx]
		}
		if csvTag == "" {
			if jsonName == "" || jsonName == "-" {
				continue
			}
			csvTag = jsonName
		}
		jn := jsonName
		if jn == "" || jn == "-" {
			jn = strings.ToLower(field.Name)
		}
		headers = append(headers, csvTag)
		jsonNames = append(jsonNames, jn)
		fieldIndices = append(fieldIndices, i)
	}
	return headers, jsonNames, fieldIndices, nil
}

// isNilOrInvalid returns true if item is a nil pointer or an invalid reflect value.
func isNilOrInvalid(item interface{}) bool {
	v := reflect.ValueOf(item)
	return !v.IsValid() || (v.Kind() == reflect.Pointer && v.IsNil())
}
