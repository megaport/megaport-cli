package cmd

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Test structures
type SimpleStruct struct {
	ID     int    `json:"id" csv:"id" header:"ID"`
	Name   string `json:"name" csv:"name" header:"Name"`
	Active bool   `json:"active" csv:"active" header:"Active"`
}

type ComplexStruct struct {
	ID         int               `json:"id" csv:"id" header:"ID"`
	Name       string            `json:"name" csv:"name" header:"Name"`
	Created    time.Time         `json:"created" csv:"created" header:"Created"`
	Tags       []string          `json:"tags" csv:"tags" header:"Tags"`
	Metadata   map[string]string `json:"metadata" csv:"metadata" header:"Metadata"`
	Reference  *SimpleStruct     `json:"reference" csv:"reference" header:"Reference"`
	unexported string            // This should be skipped
	Ignored    int               `json:"-" csv:"-" header:"-"` // This should be skipped
}

type CustomTagStruct struct {
	ID   int    `json:"id" csv:"csv_id" header:"Custom ID"`
	Name string `json:"name" csv:"csv_name" header:"Custom Name"`
}

type NoTagStruct struct {
	ID   int
	Name string
}

// Tests for printTable
func TestPrintTable_SimpleStruct(t *testing.T) {
	data := []SimpleStruct{
		{ID: 1, Name: "Item 1", Active: true},
		{ID: 2, Name: "Item 2", Active: false},
	}

	output := captureOutput(func() {
		err := printTable(data)
		assert.NoError(t, err)
	})

	// Use spaces instead of tabs to match tabwriter output
	assert.Contains(t, output, "ID")
	assert.Contains(t, output, "Name")
	assert.Contains(t, output, "Active")
	assert.Contains(t, output, "1")
	assert.Contains(t, output, "Item 1")
	assert.Contains(t, output, "true")
	assert.Contains(t, output, "2")
	assert.Contains(t, output, "Item 2")
	assert.Contains(t, output, "false")
}

func TestPrintTable_ComplexStruct(t *testing.T) {
	now := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

	data := []ComplexStruct{
		{
			ID:      1,
			Name:    "Complex Item",
			Created: now,
			Tags:    []string{"tag1", "tag2"},
			Metadata: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			Reference:  &SimpleStruct{ID: 100, Name: "Referenced", Active: true},
			unexported: "hidden",
			Ignored:    999,
		},
	}

	output := captureOutput(func() {
		err := printTable(data)
		assert.NoError(t, err)
	})

	// The output should contain the fields (except unexported and ignored)
	assert.Contains(t, output, "ID")
	assert.Contains(t, output, "Name")
	assert.Contains(t, output, "Created")
	assert.Contains(t, output, "Tags")
	assert.Contains(t, output, "Metadata")
	assert.Contains(t, output, "Reference")

	// Check specific values
	assert.Contains(t, output, "1")
	assert.Contains(t, output, "Complex Item")
	assert.Contains(t, output, "2023-01-01")

	// The unexported and ignored fields should not be in the output
	assert.NotContains(t, output, "unexported")
	assert.NotContains(t, output, "hidden")
	assert.NotContains(t, output, "Ignored")
	assert.NotContains(t, output, "999")
}

func TestPrintTable_EmptySlice(t *testing.T) {
	data := []SimpleStruct{}

	output := captureOutput(func() {
		err := printTable(data)
		assert.NoError(t, err)
	})

	// Check for headers (ignoring exact spacing)
	assert.Contains(t, output, "ID")
	assert.Contains(t, output, "Name")
	assert.Contains(t, output, "Active")
}

func TestPrintTable_NilSlice(t *testing.T) {
	var data []SimpleStruct = nil

	_ = captureOutput(func() {
		err := printTable(data)
		assert.NoError(t, err)
	})

	// Should not panic but might not output anything useful
	assert.NotPanics(t, func() {
		_ = printTable(data)
	})
}

func TestPrintTable_MixedSlice(t *testing.T) {
	data := []*SimpleStruct{
		{ID: 1, Name: "Item 1", Active: true},
		nil,
		{ID: 3, Name: "Item 3", Active: false},
	}

	output := captureOutput(func() {
		err := printTable(data)
		assert.NoError(t, err)
	})

	// Should skip the nil entry
	assert.Contains(t, output, "Item 1")
	assert.Contains(t, output, "Item 3")
	assert.NotContains(t, output, "Item 2") // There is no Item 2
}

func TestPrintTable_CustomTags(t *testing.T) {
	data := []CustomTagStruct{
		{ID: 1, Name: "Custom 1"},
		{ID: 2, Name: "Custom 2"},
	}

	output := captureOutput(func() {
		err := printTable(data)
		assert.NoError(t, err)
	})

	// Check for expected content rather than exact spacing
	assert.Contains(t, output, "Custom ID")
	assert.Contains(t, output, "Custom Name")
	assert.Contains(t, output, "1")
	assert.Contains(t, output, "Custom 1")
	assert.Contains(t, output, "2")
	assert.Contains(t, output, "Custom 2")
}

func TestPrintTable_NoTags(t *testing.T) {
	data := []NoTagStruct{
		{ID: 1, Name: "No Tags 1"},
		{ID: 2, Name: "No Tags 2"},
	}

	output := captureOutput(func() {
		err := printTable(data)
		assert.NoError(t, err)
	})

	// Should use field names as headers
	assert.Contains(t, output, "ID")
	assert.Contains(t, output, "Name")
	assert.Contains(t, output, "No Tags 1")
	assert.Contains(t, output, "No Tags 2")
}

// Tests for printCSV
func TestPrintCSV_SimpleStruct(t *testing.T) {
	data := []SimpleStruct{
		{ID: 1, Name: "Item 1", Active: true},
		{ID: 2, Name: "Item 2", Active: false},
	}

	output := captureOutput(func() {
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

	output := captureOutput(func() {
		err := printCSV(data)
		assert.NoError(t, err)
	})

	// The CSV output should have headers
	assert.Contains(t, output, "id,name,created,tags,metadata,reference")

	// Check specific values
	assert.Contains(t, output, "1,Complex Item,2023-01-01")
}

func TestPrintCSV_EmptySlice(t *testing.T) {
	data := []SimpleStruct{}

	output := captureOutput(func() {
		err := printCSV(data)
		assert.NoError(t, err)
	})

	// Should output headers even with empty data
	expected := "id,name,active\n"
	assert.Equal(t, expected, output)
}

func TestPrintCSV_NilSlice(t *testing.T) {
	var data []SimpleStruct = nil

	_ = captureOutput(func() {
		err := printCSV(data)
		assert.NoError(t, err)
	})

	// Should not panic but might not output anything useful
	assert.NotPanics(t, func() {
		_ = printCSV(data)
	})
}

// Tests for printJSON
func TestPrintJSON(t *testing.T) {
	data := []SimpleStruct{
		{ID: 1, Name: "Item 1", Active: true},
		{ID: 2, Name: "Item 2", Active: false},
	}

	output := captureOutput(func() {
		err := printJSON(data)
		assert.NoError(t, err)
	})

	// Parse the output back to JSON to compare
	var parsed []map[string]interface{}
	err := json.Unmarshal([]byte(output), &parsed)
	assert.NoError(t, err)

	expected := []map[string]interface{}{
		{"id": float64(1), "name": "Item 1", "active": true},
		{"id": float64(2), "name": "Item 2", "active": false},
	}

	assert.Equal(t, expected, parsed)
}

// Tests for isOutputCompatibleType
func TestIsOutputCompatibleType(t *testing.T) {
	tests := []struct {
		name      string
		fieldType reflect.Type
		expected  bool
	}{
		{
			name:      "String",
			fieldType: reflect.TypeOf(""),
			expected:  true,
		},
		{
			name:      "Int",
			fieldType: reflect.TypeOf(0),
			expected:  true,
		},
		{
			name:      "Bool",
			fieldType: reflect.TypeOf(false),
			expected:  true,
		},
		{
			name:      "Float",
			fieldType: reflect.TypeOf(0.0),
			expected:  true,
		},
		{
			name:      "Slice",
			fieldType: reflect.TypeOf([]string{}),
			expected:  true,
		},
		{
			name:      "Map",
			fieldType: reflect.TypeOf(map[string]string{}),
			expected:  true,
		},
		{
			name:      "Struct",
			fieldType: reflect.TypeOf(struct{}{}),
			expected:  true,
		},
		{
			name:      "Pointer",
			fieldType: reflect.TypeOf(&struct{}{}),
			expected:  true,
		},
		{
			name:      "Time",
			fieldType: reflect.TypeOf(time.Time{}),
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isOutputCompatibleType(tt.fieldType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Tests for formatFieldValue
func TestFormatFieldValue(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{
			name:     "String",
			value:    "test",
			expected: "test",
		},
		{
			name:     "Int",
			value:    42,
			expected: "42",
		},
		{
			name:     "Bool",
			value:    true,
			expected: "true",
		},
		{
			name:     "Float",
			value:    3.14,
			expected: "3.14",
		},
		{
			name:     "Time",
			value:    time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: "2023-01-01",
		},
		{
			name:     "Slice",
			value:    []string{"one", "two"},
			expected: `["one","two"]`,
		},
		{
			name:     "Map",
			value:    map[string]string{"key": "value"},
			expected: `{"key":"value"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := reflect.ValueOf(tt.value)
			result := formatFieldValue(v)

			// Handle the special case of JSON formatting
			if tt.name == "Slice" || tt.name == "Map" {
				// Compare after normalizing JSON
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

// Test output with mixed fields
func TestPrintOutput_MixedFields(t *testing.T) {
	type MixedStruct struct {
		ID         int       `json:"id" csv:"id" header:"ID"`
		Name       string    `json:"name" csv:"name" header:"Name"`
		Created    time.Time `json:"created" csv:"created" header:"Created"`
		Active     bool      `json:"active" csv:"active" header:"Active"`
		NilPtr     *string   `json:"nil_ptr" csv:"nil_ptr" header:"Nil Pointer"`
		unexported string    // This should be skipped
	}

	timeVal := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	data := []MixedStruct{
		{
			ID:         1,
			Name:       "Test Item",
			Created:    timeVal,
			Active:     true,
			NilPtr:     nil,
			unexported: "hidden",
		},
	}

	formats := []string{"table", "csv", "json"}
	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			output := captureOutput(func() {
				err := printOutput(data, format)
				assert.NoError(t, err)
			})

			// All formats should include the exported fields
			assert.Contains(t, output, "Test Item")
			assert.Contains(t, output, "2023-01-01")

			// None should contain unexported fields
			assert.NotContains(t, output, "hidden")

			// All should handle nil pointer gracefully
			assert.NotPanics(t, func() {
				_ = printOutput(data, format)
			})
		})
	}
}

// Test error handling
func TestPrintOutput_InvalidFormat(t *testing.T) {
	data := []SimpleStruct{{ID: 1, Name: "Test"}}

	err := printOutput(data, "invalid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid output format")
}
