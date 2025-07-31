package output

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var noColor = false

type SimpleStruct struct {
	ID     int    `json:"id" csv:"id" header:"ID"`
	Name   string `json:"name" csv:"name" header:"Name"`
	Active bool   `json:"active" csv:"active" header:"Active"`
}

type ComplexStruct struct {
	ID        int               `json:"id" csv:"id" header:"ID"`
	Name      string            `json:"name" csv:"name" header:"Name"`
	Created   time.Time         `json:"created" csv:"created" header:"Created"`
	Tags      []string          `json:"tags" csv:"tags" header:"Tags"`
	Metadata  map[string]string `json:"metadata" csv:"metadata" header:"Metadata"`
	Reference *SimpleStruct     `json:"reference" csv:"reference" header:"Reference"`
	Ignored   int               `json:"-" csv:"-" header:"-"`
}

type CustomTagStruct struct {
	ID   int    `json:"id" csv:"csv_id" header:"Custom ID"`
	Name string `json:"name" csv:"csv_name" header:"Custom Name"`
}

type NoTagStruct struct {
	ID   int
	Name string
}

func TestPrintCSV_SimpleStruct(t *testing.T) {
	data := []SimpleStruct{
		{ID: 1, Name: "Item 1", Active: true},
		{ID: 2, Name: "Item 2", Active: false},
	}

	output := CaptureOutput(func() {
		err := printCSV(data)
		assert.NoError(t, err)
	})

	expected := "id,name,active\n" +
		"1,Item 1,true\n" +
		"2,Item 2,false\n"

	assert.Equal(t, expected, output)
}

func TestPrintCSV_ComplexStruct(t *testing.T) {
	now := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

	data := []ComplexStruct{
		{
			ID:      1,
			Name:    "Complex Item",
			Created: now,
			Tags:    []string{"tag1", "tag2"},
			Metadata: map[string]string{
				"key1": "value1",
			},
			Reference: &SimpleStruct{ID: 100, Name: "Referenced", Active: true},
		},
	}

	output := CaptureOutput(func() {
		err := printCSV(data)
		assert.NoError(t, err)
	})

	assert.Contains(t, output, "id,name,created,tags,metadata,reference")
	assert.Contains(t, output, "1,Complex Item,2023-01-01")
}

func TestPrintCSV_EmptySlice(t *testing.T) {
	data := []SimpleStruct{}

	output := CaptureOutput(func() {
		err := printCSV(data)
		assert.NoError(t, err)
	})

	expected := "id,name,active\n"
	assert.Equal(t, expected, output)
}

func TestPrintCSV_NilSlice(t *testing.T) {
	var data []SimpleStruct = nil

	_ = CaptureOutput(func() {
		err := printCSV(data)
		assert.NoError(t, err)
	})

	assert.NotPanics(t, func() {
		_ = printCSV(data)
	})
}

func TestPrintJSON(t *testing.T) {
	data := []SimpleStruct{
		{ID: 1, Name: "Item 1", Active: true},
		{ID: 2, Name: "Item 2", Active: false},
	}

	output := CaptureOutput(func() {
		err := printJSON(data)
		assert.NoError(t, err)
	})

	var parsed []map[string]interface{}
	err := json.Unmarshal([]byte(output), &parsed)
	assert.NoError(t, err)

	expected := []map[string]interface{}{
		{"id": float64(1), "name": "Item 1", "active": true},
		{"id": float64(2), "name": "Item 2", "active": false},
	}

	assert.Equal(t, expected, parsed)
}

func TestIsOutputCompatibleType(t *testing.T) {
	tests := []struct {
		name      string
		fieldType reflect.Type
		expected  bool
	}{
		{"String", reflect.TypeOf(""), true},
		{"Int", reflect.TypeOf(0), true},
		{"Bool", reflect.TypeOf(false), true},
		{"Float", reflect.TypeOf(0.0), true},
		{"Slice", reflect.TypeOf([]string{}), true},
		{"Map", reflect.TypeOf(map[string]string{}), true},
		{"Struct", reflect.TypeOf(struct{}{}), true},
		{"Pointer", reflect.TypeOf(&struct{}{}), true},
		{"Time", reflect.TypeOf(time.Time{}), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isOutputCompatibleType(tt.fieldType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatFieldValue(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{"String", "test", "test"},
		{"Int", 42, "42"},
		{"Bool", true, "true"},
		{"Float", 3.14, "3.14"},
		{"Time", time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), "2023-01-01"},
		{"Slice", []string{"one", "two"}, `["one","two"]`},
		{"Map", map[string]string{"key": "value"}, `{"key":"value"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := reflect.ValueOf(tt.value)
			result := formatFieldValue(v)

			if tt.name == "Slice" || tt.name == "Map" {
				var expected, actual interface{}
				err := json.Unmarshal([]byte(tt.expected), &expected)
				assert.NoError(t, err)
				err = json.Unmarshal([]byte(result), &actual)
				assert.NoError(t, err)
				assert.Equal(t, expected, actual)
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestPrintOutput_MixedFields(t *testing.T) {
	type MixedStruct struct {
		ID         int       `json:"id" csv:"id" header:"ID"`
		Name       string    `json:"name" csv:"name" header:"Name"`
		Created    time.Time `json:"created" csv:"created" header:"Created"`
		Active     bool      `json:"active" csv:"active" header:"Active"`
		NilPtr     *string   `json:"nil_ptr" csv:"nil_ptr" header:"Nil Pointer"`
		unexported string
	}

	timeVal := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	data := []MixedStruct{{ID: 1, Name: "Test Item", Created: timeVal, Active: true, NilPtr: nil, unexported: "hidden"}}

	formats := []string{"table", "csv", "json"}
	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			output := CaptureOutput(func() {
				err := PrintOutput(data, format, noColor)
				assert.NoError(t, err)
			})

			assert.Contains(t, output, "Test Item")
			assert.Contains(t, output, "2023-01-01")
			assert.NotContains(t, output, "hidden")
			assert.NotPanics(t, func() {
				_ = PrintOutput(data, format, noColor)
			})
		})
	}
}

func TestPrintOutput_InvalidFormat(t *testing.T) {
	data := []SimpleStruct{{ID: 1, Name: "Test"}}

	err := PrintOutput(data, "invalid", noColor)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid output format")
}

func TestCalculateColumnWidths(t *testing.T) {
	emptyWidths := calculateColumnWidths([][]string{})
	assert.Nil(t, emptyWidths)

	rows := [][]string{
		{"Header1", "Header2", "LongHeader3"},
		{"Value1", "LongerValue2", "Val3"},
		{"X", "Y", "VeryLongValue3"},
	}

	widths := calculateColumnWidths(rows)
	assert.Equal(t, 3, len(widths))
	assert.Equal(t, 7, widths[0])
	assert.Equal(t, 12, widths[1])
	assert.Equal(t, 14, widths[2])
}

func TestColorizeValue(t *testing.T) {
	originalNoColor := noColor
	defer func() { noColor = originalNoColor }()

	noColor = true
	assert.Equal(t, "ACTIVE", colorizeValue("ACTIVE", "status", noColor))
	assert.Equal(t, "123", colorizeValue("123", "id", noColor))
	assert.Equal(t, "Important", colorizeValue("Important", "name", noColor))
	assert.Equal(t, "12.50", colorizeValue("12.50", "price", noColor))

	noColor = false
	assert.NotPanics(t, func() {
		_ = colorizeValue("ACTIVE", "status", noColor)
		_ = colorizeValue("123", "id", noColor)
		_ = colorizeValue("Important", "name", noColor)
		_ = colorizeValue("12.50", "price", noColor)
	})
}

func TestExtractFieldInfo(t *testing.T) {
	type TestStruct struct {
		ID       int    `json:"id" header:"Custom ID"`
		Name     string `json:"name" csv:"custom_name"`
		Internal string `json:"-" csv:"-" header:"-"`
	}

	headers, indices := extractFieldInfo(reflect.TypeOf(TestStruct{}))

	assert.Equal(t, 2, len(headers))
	assert.Contains(t, headers, "Custom ID")
	assert.Contains(t, headers, "custom_name")
	assert.Equal(t, []int{0, 1}, indices)
}

func TestPrintPrettyTable_SimpleStruct(t *testing.T) {
	data := []SimpleStruct{
		{ID: 1, Name: "Item 1", Active: true},
		{ID: 2, Name: "Item 2", Active: false},
	}

	output := CaptureOutput(func() {
		err := PrintOutput(data, "table", noColor)
		assert.NoError(t, err)
	})

	assert.Contains(t, output, "│")
	assert.Contains(t, output, "─")
	assert.Contains(t, output, "ID")
	assert.Contains(t, output, "NAME")
	assert.Contains(t, output, "ACTIVE")
	assert.Contains(t, output, " 1 ")
	assert.Contains(t, output, " Item 1 ")
	assert.Contains(t, output, " true ")
}

func TestPrintPrettyTable_ComplexStruct(t *testing.T) {
	reference := SimpleStruct{ID: 100, Name: "Reference", Active: true}
	now := time.Now()
	data := []ComplexStruct{{ID: 1, Name: "Complex Item", Created: now, Tags: []string{"tag1", "tag2"}, Metadata: map[string]string{"key": "value"}, Reference: &reference}}

	output := CaptureOutput(func() {
		err := PrintOutput(data, "table", noColor)
		assert.NoError(t, err)
	})

	assert.Contains(t, output, "ID")
	assert.Contains(t, output, "NAME")
	assert.Contains(t, output, "CREATED")
	assert.Contains(t, output, "TAGS")
	assert.Contains(t, output, "METADATA")
	assert.Contains(t, output, "REFERENCE")
	assert.Contains(t, output, "Complex Item")
}

func TestPrintPrettyTable_EmptySlice(t *testing.T) {
	var data []SimpleStruct

	output := CaptureOutput(func() {
		err := PrintOutput(data, "table", noColor)
		assert.NoError(t, err)
	})

	assert.Contains(t, output, "ID")
	assert.Contains(t, output, "NAME")
	assert.Contains(t, output, "ACTIVE")
	assert.Contains(t, output, "│")
	assert.Contains(t, output, "─")
	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.Equal(t, 4, len(lines), "Empty table should have header, separator and closing line")
}

func TestPrintPrettyTable_NilSlice(t *testing.T) {
	var data []SimpleStruct = nil

	output := CaptureOutput(func() {
		err := PrintOutput(data, "table", noColor)
		assert.NoError(t, err)
	})

	assert.Contains(t, output, "ID")
	assert.Contains(t, output, "NAME")
	assert.Contains(t, output, "ACTIVE")
	assert.Contains(t, output, "│")
	assert.Contains(t, output, "─")
}

func TestPrintPrettyTable_CustomHeaders(t *testing.T) {
	data := []CustomTagStruct{{ID: 1, Name: "Custom Item"}}

	output := CaptureOutput(func() {
		err := PrintOutput(data, "table", noColor)
		assert.NoError(t, err)
	})

	assert.Contains(t, output, "CUSTOM ID")
	assert.Contains(t, output, "CUSTOM NAME")
}

func TestTableColorization(t *testing.T) {
	type ColorTestStruct struct {
		ID       string `json:"id" header:"ID"`
		Status   string `json:"status" header:"STATUS"`
		Name     string `json:"name" header:"NAME"`
		Price    string `json:"price" header:"PRICE"`
		Speed    int    `json:"speed" header:"SPEED"`
		Location string `json:"location" header:"LOCATION"`
		Empty    string `json:"empty" header:"EMPTY"`
	}

	data := []ColorTestStruct{{ID: "test-123", Status: "ACTIVE", Name: "Test Item", Price: "99.99", Speed: 1000, Location: "NYC", Empty: ""}}

	outputColor := CaptureOutput(func() {
		err := PrintOutput(data, "table", false)
		assert.NoError(t, err)
	})

	outputNoColor := CaptureOutput(func() {
		err := PrintOutput(data, "table", true)
		assert.NoError(t, err)
	})

	hasColorCodes := strings.Contains(outputColor, "\033[") ||
		strings.Contains(outputColor, "\u001b[")

	noColorCodes := !strings.Contains(outputNoColor, "\033[") &&
		!strings.Contains(outputNoColor, "\u001b[")

	assert.True(t, hasColorCodes, "Color output should contain ANSI color codes")
	assert.True(t, noColorCodes, "No-color output should not contain ANSI color codes")
}

func TestPrintPrettyTable_TableStyle(t *testing.T) {
	data := []SimpleStruct{{ID: 1, Name: "Test Item", Active: true}}

	output := CaptureOutput(func() {
		err := PrintOutput(data, "table", noColor)
		assert.NoError(t, err)
	})

	assert.Contains(t, output, "│")
	assert.Contains(t, output, "─")
	assert.Contains(t, output, "ID")
	assert.Contains(t, output, "NAME")
	assert.Contains(t, output, "ACTIVE")
	assert.Contains(t, output, " ID ")
	assert.Contains(t, output, " NAME ")
	assert.Contains(t, output, " ACTIVE ")
}

func TestJSONOutput_JQCompatibility(t *testing.T) {
	// Test that JSON output is clean and can be parsed by jq-like parsers
	data := []SimpleStruct{
		{ID: 1, Name: "Item 1", Active: true},
		{ID: 2, Name: "Item 2", Active: false},
	}

	output := CaptureOutput(func() {
		err := PrintOutput(data, "json", noColor)
		assert.NoError(t, err)
	})

	// Verify the output is valid JSON
	var parsed []map[string]interface{}
	err := json.Unmarshal([]byte(output), &parsed)
	assert.NoError(t, err, "JSON output should be valid and parseable")

	// Verify the structure matches expectations
	assert.Len(t, parsed, 2, "Should have 2 items")
	assert.Equal(t, float64(1), parsed[0]["id"])
	assert.Equal(t, "Item 1", parsed[0]["name"])
	assert.Equal(t, true, parsed[0]["active"])

	// Verify no extra whitespace or control characters
	trimmed := strings.TrimSpace(output)
	assert.True(t, strings.HasPrefix(trimmed, "["), "JSON should start with [")
	assert.True(t, strings.HasSuffix(trimmed, "]"), "JSON should end with ]")

	// Verify no ANSI color codes in JSON output
	assert.False(t, strings.Contains(output, "\033["), "JSON should not contain ANSI escape codes")
	assert.False(t, strings.Contains(output, "\u001b["), "JSON should not contain ANSI escape codes")
}

func TestJSONOutput_ComplexData_JQCompatibility(t *testing.T) {
	// Test complex data structures for jq compatibility
	now := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	reference := &SimpleStruct{ID: 100, Name: "Referenced Item", Active: true}

	data := []ComplexStruct{
		{
			ID:        1,
			Name:      "Complex Item",
			Created:   now,
			Tags:      []string{"tag1", "tag2", "tag3"},
			Metadata:  map[string]string{"key1": "value1", "key2": "value2"},
			Reference: reference,
		},
	}

	output := CaptureOutput(func() {
		err := PrintOutput(data, "json", noColor)
		assert.NoError(t, err)
	})

	// Verify the output is valid JSON
	var parsed []map[string]interface{}
	err := json.Unmarshal([]byte(output), &parsed)
	assert.NoError(t, err, "Complex JSON output should be valid and parseable")

	// Verify complex structures are properly serialized
	assert.Len(t, parsed, 1)
	item := parsed[0]

	assert.Equal(t, float64(1), item["id"])
	assert.Equal(t, "Complex Item", item["name"])
	assert.Contains(t, item["created"], "2023-01-01T12:00:00Z")

	// Verify arrays are properly serialized
	tags, ok := item["tags"].([]interface{})
	assert.True(t, ok, "Tags should be an array")
	assert.Len(t, tags, 3)
	assert.Equal(t, "tag1", tags[0])

	// Verify maps are properly serialized
	metadata, ok := item["metadata"].(map[string]interface{})
	assert.True(t, ok, "Metadata should be a map")
	assert.Equal(t, "value1", metadata["key1"])

	// Verify nested objects are properly serialized
	ref, ok := item["reference"].(map[string]interface{})
	assert.True(t, ok, "Reference should be an object")
	assert.Equal(t, float64(100), ref["id"])
	assert.Equal(t, "Referenced Item", ref["name"])
}

func TestJSONOutput_EmptyData_JQCompatibility(t *testing.T) {
	// Test empty data produces valid JSON
	var data []SimpleStruct

	output := CaptureOutput(func() {
		err := PrintOutput(data, "json", noColor)
		assert.NoError(t, err)
	})

	// Verify empty array is valid JSON
	var parsed []interface{}
	err := json.Unmarshal([]byte(output), &parsed)
	assert.NoError(t, err, "Empty JSON array should be valid")
	assert.Len(t, parsed, 0, "Empty array should have no elements")

	// Verify it's exactly "[]" (or "[]" with whitespace)
	trimmed := strings.TrimSpace(output)
	assert.Equal(t, "[]", trimmed, "Empty data should produce clean empty JSON array")
}

func TestJSONOutput_NilData_JQCompatibility(t *testing.T) {
	// Test nil data produces valid JSON
	var data []SimpleStruct = nil

	output := CaptureOutput(func() {
		err := PrintOutput(data, "json", noColor)
		assert.NoError(t, err)
	})

	// Verify nil produces valid JSON
	var parsed []interface{}
	err := json.Unmarshal([]byte(output), &parsed)
	assert.NoError(t, err, "Nil data JSON should be valid")
	assert.Len(t, parsed, 0, "Nil data should produce empty array")
}

func TestJSONOutput_SpecialCharacters_JQCompatibility(t *testing.T) {
	// Test data with special characters that need proper JSON escaping
	data := []SimpleStruct{
		{ID: 1, Name: "Item with \"quotes\" and \nnewlines\tand\ttabs", Active: true},
		{ID: 2, Name: "Item with unicode: 🚀 and backslash: \\", Active: false},
	}

	output := CaptureOutput(func() {
		err := PrintOutput(data, "json", noColor)
		assert.NoError(t, err)
	})

	// Verify the output is valid JSON despite special characters
	var parsed []map[string]interface{}
	err := json.Unmarshal([]byte(output), &parsed)
	assert.NoError(t, err, "JSON with special characters should be valid")

	// Verify special characters are properly escaped and preserved
	assert.Contains(t, parsed[0]["name"], "quotes")
	assert.Contains(t, parsed[0]["name"], "newlines")
	assert.Contains(t, parsed[0]["name"], "tabs")
	assert.Contains(t, parsed[1]["name"], "🚀")
	assert.Contains(t, parsed[1]["name"], "\\")
}

func TestJSONOutput_LargeData_JQCompatibility(t *testing.T) {
	// Test with larger dataset to ensure performance and correctness
	var data []SimpleStruct
	for i := 1; i <= 100; i++ {
		data = append(data, SimpleStruct{
			ID:     i,
			Name:   fmt.Sprintf("Item %d", i),
			Active: i%2 == 0,
		})
	}

	output := CaptureOutput(func() {
		err := PrintOutput(data, "json", noColor)
		assert.NoError(t, err)
	})

	// Verify large JSON is still valid
	var parsed []map[string]interface{}
	err := json.Unmarshal([]byte(output), &parsed)
	assert.NoError(t, err, "Large JSON output should be valid")
	assert.Len(t, parsed, 100, "Should have all 100 items")

	// Spot check first and last items
	assert.Equal(t, float64(1), parsed[0]["id"])
	assert.Equal(t, "Item 1", parsed[0]["name"])
	assert.Equal(t, false, parsed[0]["active"])

	assert.Equal(t, float64(100), parsed[99]["id"])
	assert.Equal(t, "Item 100", parsed[99]["name"])
	assert.Equal(t, true, parsed[99]["active"])
}

func TestJSONOutput_NoTrailingNewlines(t *testing.T) {
	// Ensure JSON output doesn't have trailing newlines that could interfere with jq
	data := []SimpleStruct{
		{ID: 1, Name: "Test", Active: true},
	}

	output := CaptureOutput(func() {
		err := PrintOutput(data, "json", noColor)
		assert.NoError(t, err)
	})

	// JSON should end with ] and a single newline, not multiple newlines
	assert.True(t, strings.HasSuffix(output, "]\n"), "JSON should end with ] followed by single newline")
	assert.False(t, strings.HasSuffix(output, "]\n\n"), "JSON should not have multiple trailing newlines")
}
