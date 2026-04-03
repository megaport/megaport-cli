package output

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"
)

// outputFields holds the user-selected fields from --fields flag.
// nil means all fields are shown. Protected by outputFieldsMu.
var (
	outputFields   []string
	outputFieldsMu sync.RWMutex
)

// SetOutputFields sets the field filter applied by all PrintOutput calls.
// Only fields whose json tag name or header display name (case-insensitive) appears
// in fields will be included in output. Pass nil to restore full output (all fields).
// This function is goroutine-safe. Tests should call defer SetOutputFields(nil) to
// reset state between test cases.
func SetOutputFields(fields []string) {
	outputFieldsMu.Lock()
	defer outputFieldsMu.Unlock()
	outputFields = fields
}

// getOutputFields returns a copy of the current field filter under a read lock.
func getOutputFields() []string {
	outputFieldsMu.RLock()
	defer outputFieldsMu.RUnlock()
	if outputFields == nil {
		return nil
	}
	cp := make([]string, len(outputFields))
	copy(cp, outputFields)
	return cp
}

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
	Key   string `json:"key" header:"Key"`
	Value string `json:"value" header:"Value"`
}

// PrintOutput prints data in the specified format
func PrintOutput[T OutputFields](data []T, format string, noColor bool) error {
	validFormats := map[string]bool{
		"table": true,
		"json":  true,
		"csv":   true,
		"xml":   true,
	}
	if !validFormats[format] {
		return fmt.Errorf("invalid output format: %s", format)
	}
	switch format {
	case "json":
		return printJSON(data)
	case "csv":
		return printCSV(data)
	case "xml":
		return printXML(data)
	default:
		return printTable(data, noColor)
	}
}

// getStructTypeInfo extracts header names, json names, and field indices from a struct type.
// jsonNames are the json tag values (used for --fields matching); headers are the display names.
func getStructTypeInfo[T OutputFields](data []T) (headers, jsonNames []string, fieldIndices []int, err error) {
	var sample T
	if len(data) > 0 {
		sample = data[0]
	}
	sampleVal := reflect.ValueOf(sample)
	if !sampleVal.IsValid() {
		return nil, nil, nil, nil
	}
	itemType := sampleVal.Type()
	if itemType.Kind() == reflect.Ptr {
		if sampleVal.IsNil() {
			if itemType.Elem().Kind() != reflect.Struct {
				return nil, nil, nil, nil
			}
			itemType = itemType.Elem()
		} else {
			sampleVal = sampleVal.Elem()
			itemType = sampleVal.Type()
		}
	}
	if itemType.Kind() != reflect.Struct {
		return nil, nil, nil, nil
	}
	headers, jsonNames, fieldIndices = extractFieldInfo(itemType)
	return headers, jsonNames, fieldIndices, nil
}

// extractFieldInfo extracts field information from a struct type.
// Returns headers (display names), jsonNames (json tag names for --fields matching), and field indices.
func extractFieldInfo(itemType reflect.Type) (headers, jsonNames []string, fieldIndices []int) {
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

		// Derive the json name for --fields matching.
		jsonName := field.Tag.Get("json")
		if idx := strings.Index(jsonName, ","); idx != -1 {
			jsonName = jsonName[:idx]
		}
		if jsonName == "" || jsonName == "-" {
			jsonName = strings.ToLower(field.Name)
		}

		headers = append(headers, headerTag)
		jsonNames = append(jsonNames, jsonName)
		fieldIndices = append(fieldIndices, i)
	}
	return headers, jsonNames, fieldIndices
}

// filterByFields filters (headers, jsonNames, indices) to only the user-selected fields.
// Matching is case-insensitive: json names are tried first, then header names.
// Duplicate selections are silently deduplicated. Returns an error listing available
// json names if any selected field is unknown.
func filterByFields(headers, jsonNames []string, indices []int, selected []string) ([]string, []string, []int, error) {
	if len(selected) == 0 {
		return headers, jsonNames, indices, nil
	}
	// Two separate maps avoids the collision that occurs when a header name
	// happens to match a json name of a different field.
	byJSON := make(map[string]int, len(jsonNames))
	for i, jn := range jsonNames {
		byJSON[strings.ToLower(jn)] = i
	}
	byHeader := make(map[string]int, len(headers))
	for i, h := range headers {
		byHeader[strings.ToLower(h)] = i
	}

	// Pre-build the available-fields string once so it is not rebuilt on every error.
	available := make([]string, 0, len(jsonNames))
	for j, jn := range jsonNames {
		if j < len(headers) && !strings.EqualFold(headers[j], jn) {
			available = append(available, fmt.Sprintf("%s (or %q)", jn, headers[j]))
		} else {
			available = append(available, jn)
		}
	}
	availableStr := strings.Join(available, ", ")

	seen := make(map[int]bool, len(selected))
	var outHeaders, outJSONNames []string
	var outIndices []int
	for _, sel := range selected {
		key := strings.ToLower(strings.TrimSpace(sel))
		// Prefer json name match; fall back to header display name.
		i, ok := byJSON[key]
		if !ok {
			i, ok = byHeader[key]
		}
		if !ok {
			return nil, nil, nil, fmt.Errorf("unknown field %q, available fields: %s", sel, availableStr)
		}
		if seen[i] {
			continue // deduplicate repeated field selections
		}
		seen[i] = true
		if len(headers) > 0 {
			outHeaders = append(outHeaders, headers[i])
		}
		outJSONNames = append(outJSONNames, jsonNames[i])
		outIndices = append(outIndices, indices[i])
	}
	return outHeaders, outJSONNames, outIndices, nil
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
		t, ok := fieldVal.Interface().(time.Time)
		if !ok || t.IsZero() {
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
