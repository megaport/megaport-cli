package output

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// Output is a marker interface for output types
type Output interface {
	isOuput()
}

// OutputFields is a constraint for types that can be output
type OutputFields interface {
	any
}

// ResourceTag represents a key-value tag pair
type ResourceTag struct {
	Key   string `json:"key" header:"KEY"`
	Value string `json:"value" header:"VALUE"`
}

// PrintOutput prints data in the specified format
func PrintOutput[T OutputFields](data []T, format string, noColor bool) error {
	validFormats := map[string]bool{
		"table": true,
		"json":  true,
		"csv":   true,
	}
	if !validFormats[format] {
		return fmt.Errorf("invalid output format: %s", format)
	}
	switch format {
	case "json":
		return printJSON(data)
	case "csv":
		return printCSV(data)
	default:
		return printTable(data, noColor)
	}
}

// getStructTypeInfo extracts header names and field indices from a struct type
func getStructTypeInfo[T OutputFields](data []T) ([]string, []int, error) {
	var sample T
	if len(data) > 0 {
		sample = data[0]
	}
	sampleVal := reflect.ValueOf(sample)
	if !sampleVal.IsValid() {
		fmt.Println("")
		return nil, nil, nil
	}
	itemType := sampleVal.Type()
	if itemType.Kind() == reflect.Ptr {
		if sampleVal.IsNil() {
			if itemType.Elem().Kind() != reflect.Struct {
				return nil, nil, nil
			}
			itemType = itemType.Elem()
		} else {
			sampleVal = sampleVal.Elem()
			itemType = sampleVal.Type()
		}
	}
	if itemType.Kind() != reflect.Struct {
		return nil, nil, nil
	}
	headers, fieldIndices := extractFieldInfo(itemType)
	return headers, fieldIndices, nil
}

// extractFieldInfo extracts field information from a struct type
func extractFieldInfo(itemType reflect.Type) ([]string, []int) {
	var headers []string
	var fieldIndices []int
	for i := 0; i < itemType.NumField(); i++ {
		field := itemType.Field(i)
		if field.PkgPath != "" {
			continue
		}
		if !isOutputCompatibleType(field.Type) {
			continue
		}
		headerTag := field.Tag.Get("header")
		if headerTag == "-" {
			continue
		}
		if headerTag == "" {
			headerTag = field.Tag.Get("csv")
			if headerTag == "-" {
				continue
			}
		}
		if headerTag == "" {
			headerTag = field.Tag.Get("output")
			if headerTag == "-" {
				continue
			}
		}
		if headerTag == "" {
			headerTag = strings.ToUpper(field.Name)
		}
		headers = append(headers, headerTag)
		fieldIndices = append(fieldIndices, i)
	}
	return headers, fieldIndices
}

// isOutputCompatibleType checks if a type can be output
func isOutputCompatibleType(t reflect.Type) bool {
	// Handle pointer types by checking the element type
	if t.Kind() == reflect.Ptr {
		return isOutputCompatibleType(t.Elem())
	}
	
	// Special case: time.Time is always compatible
	if t == reflect.TypeOf(time.Time{}) {
		return true
	}
	
	switch t.Kind() {
	case reflect.Slice, reflect.Array:
		// Slices and arrays are compatible (we can serialize them)
		return true
	case reflect.Map:
		// Maps are compatible (we can serialize them)
		return true
	case reflect.Struct:
		// Structs are compatible (we can serialize them)
		return true
	case reflect.Chan, reflect.Func:
		// Channels and functions cannot be serialized
		return false
	default:
		// All primitive types are compatible
		return true
	}
}

// extractRowData extracts field values from a struct for table/CSV output
func extractRowData(item interface{}, fieldIndices []int) []string {
	itemVal := reflect.ValueOf(item)
	if itemVal.Kind() == reflect.Ptr {
		if itemVal.IsNil() {
			return make([]string, len(fieldIndices))
		}
		itemVal = itemVal.Elem()
	}
	if itemVal.Kind() != reflect.Struct {
		return nil
	}
	values := make([]string, len(fieldIndices))
	for i, idx := range fieldIndices {
		fieldVal := itemVal.Field(idx)
		values[i] = formatFieldValue(fieldVal)
	}
	return values
}

// formatFieldValue formats a field value for display
func formatFieldValue(fieldVal reflect.Value) string {
	if !fieldVal.IsValid() {
		return ""
	}
	if fieldVal.Kind() == reflect.Ptr {
		if fieldVal.IsNil() {
			return ""
		}
		fieldVal = fieldVal.Elem()
	}
	
	// Special handling for time.Time - format as ISO date
	if fieldVal.Type() == reflect.TypeOf(time.Time{}) {
		t := fieldVal.Interface().(time.Time)
		if t.IsZero() {
			return ""
		}
		return t.Format("2006-01-02")
	}
	
	switch fieldVal.Kind() {
	case reflect.Slice, reflect.Array:
		if fieldVal.Len() == 0 {
			return ""
		}
		// For slices of primitives, try JSON serialization first
		if fieldVal.CanInterface() {
			val := fieldVal.Interface()
			if jsonBytes, err := json.Marshal(val); err == nil {
				return string(jsonBytes)
			}
		}
		// Fallback to comma-separated values
		var parts []string
		for i := 0; i < fieldVal.Len(); i++ {
			elem := fieldVal.Index(i)
			parts = append(parts, formatFieldValue(elem))
		}
		return strings.Join(parts, ", ")
	case reflect.Map:
		if fieldVal.Len() == 0 {
			return ""
		}
		// Serialize maps as JSON
		if fieldVal.CanInterface() {
			val := fieldVal.Interface()
			if jsonBytes, err := json.Marshal(val); err == nil {
				return string(jsonBytes)
			}
		}
		return fmt.Sprintf("%v", fieldVal.Interface())
	case reflect.Struct:
		// Serialize structs as JSON
		if fieldVal.CanInterface() {
			val := fieldVal.Interface()
			if jsonBytes, err := json.Marshal(val); err == nil {
				return string(jsonBytes)
			}
		}
		return fmt.Sprintf("%v", fieldVal.Interface())
	case reflect.Bool:
		if fieldVal.Bool() {
			return "true"
		}
		return "false"
	case reflect.String:
		return fieldVal.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", fieldVal.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", fieldVal.Uint())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%g", fieldVal.Float())
	default:
		val := fieldVal.Interface()
		if val == nil {
			return ""
		}
		return fmt.Sprintf("%v", val)
	}
}
