package output

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var noColor = false

// Test structures
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
	Ignored   int               `json:"-" csv:"-" header:"-"` // This should be skipped
}

type CustomTagStruct struct {
	ID   int    `json:"id" csv:"csv_id" header:"Custom ID"`
	Name string `json:"name" csv:"csv_name" header:"Custom Name"`
}

type NoTagStruct struct {
	ID   int
	Name string
}

// Tests for printCSV
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

	// The CSV output should have headers
	assert.Contains(t, output, "id,name,created,tags,metadata,reference")

	// Check specific values
	assert.Contains(t, output, "1,Complex Item,2023-01-01")
}

func TestPrintCSV_EmptySlice(t *testing.T) {
	data := []SimpleStruct{}

	output := CaptureOutput(func() {
		err := printCSV(data)
		assert.NoError(t, err)
	})

	// Should output headers even with empty data
	expected := "id,name,active\n"
	assert.Equal(t, expected, output)
}

func TestPrintCSV_NilSlice(t *testing.T) {
	var data []SimpleStruct = nil

	_ = CaptureOutput(func() {
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

	output := CaptureOutput(func() {
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
			output := CaptureOutput(func() {
				err := PrintOutput(data, format, noColor)
				assert.NoError(t, err)
			})

			// All formats should include the exported fields
			assert.Contains(t, output, "Test Item")
			assert.Contains(t, output, "2023-01-01")

			// None should contain unexported fields
			assert.NotContains(t, output, "hidden")

			// All should handle nil pointer gracefully
			assert.NotPanics(t, func() {
				_ = PrintOutput(data, format, noColor)
			})
		})
	}
}

// Test error handling
func TestPrintOutput_InvalidFormat(t *testing.T) {
	data := []SimpleStruct{{ID: 1, Name: "Test"}}

	err := PrintOutput(data, "invalid", noColor)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid output format")
}

func TestCalculateColumnWidths(t *testing.T) {
	// Empty input case
	emptyWidths := calculateColumnWidths([][]string{})
	assert.Nil(t, emptyWidths)

	// Regular case with varied column widths
	rows := [][]string{
		{"Header1", "Header2", "LongHeader3"},
		{"Value1", "LongerValue2", "Val3"},
		{"X", "Y", "VeryLongValue3"},
	}

	widths := calculateColumnWidths(rows)
	assert.Equal(t, 3, len(widths))
	assert.Equal(t, 7, widths[0])  // "Value1" length
	assert.Equal(t, 12, widths[1]) // "LongerValue2" length
	assert.Equal(t, 14, widths[2]) // "VeryLongValue3" length
}

func TestColorizeValue(t *testing.T) {
	// Save original noColor setting
	originalNoColor := noColor
	defer func() { noColor = originalNoColor }()

	// Test with noColor = true
	noColor = true

	// Test various field types
	assert.Equal(t, "ACTIVE", colorizeValue("ACTIVE", "status", noColor))
	assert.Equal(t, "123", colorizeValue("123", "id", noColor))
	assert.Equal(t, "Important", colorizeValue("Important", "name", noColor))
	assert.Equal(t, "12.50", colorizeValue("12.50", "price", noColor))

	// Test with noColor = false
	noColor = false
	// Just verify it doesn't panic (exact output depends on color codes)
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

	// Check headers are extracted correctly with priority
	assert.Equal(t, 2, len(headers))
	assert.Contains(t, headers, "Custom ID")   // From header tag
	assert.Contains(t, headers, "custom_name") // From csv tag

	// Check indices
	assert.Equal(t, []int{0, 1}, indices) // Should have indices for first two fields only
}

// Tests for printPrettyTable functionality

func TestPrintPrettyTable_SimpleStruct(t *testing.T) {
	// Create test data
	data := []SimpleStruct{
		{ID: 1, Name: "Item 1", Active: true},
		{ID: 2, Name: "Item 2", Active: false},
	}

	// Capture the output
	output := CaptureOutput(func() {
		err := PrintOutput(data, "table", noColor)
		assert.NoError(t, err)
	})

	// Check for box drawing characters and correct uppercase headers
	assert.Contains(t, output, "│")
	assert.Contains(t, output, "─")
	assert.Contains(t, output, "ID")
	assert.Contains(t, output, "NAME")
	assert.Contains(t, output, "ACTIVE")

	// Check for proper content alignment with spaces
	assert.Contains(t, output, " 1 ")
	assert.Contains(t, output, " Item 1 ")
	assert.Contains(t, output, " true ")
}

func TestPrintPrettyTable_ComplexStruct(t *testing.T) {
	// Create test data with complex structures
	reference := SimpleStruct{ID: 100, Name: "Reference", Active: true}
	now := time.Now()
	data := []ComplexStruct{
		{
			ID:        1,
			Name:      "Complex Item",
			Created:   now,
			Tags:      []string{"tag1", "tag2"},
			Metadata:  map[string]string{"key": "value"},
			Reference: &reference,
		},
	}

	// Capture the output
	output := CaptureOutput(func() {
		err := PrintOutput(data, "table", noColor)
		assert.NoError(t, err)
	})

	// Check headers and formatting
	assert.Contains(t, output, "ID")
	assert.Contains(t, output, "NAME")
	assert.Contains(t, output, "CREATED")
	assert.Contains(t, output, "TAGS")
	assert.Contains(t, output, "METADATA")
	assert.Contains(t, output, "REFERENCE")
	assert.Contains(t, output, "Complex Item")
}

func TestPrintPrettyTable_EmptySlice(t *testing.T) {
	// Create empty slice
	var data []SimpleStruct

	// Capture the output
	output := CaptureOutput(func() {
		err := PrintOutput(data, "table", noColor)
		assert.NoError(t, err)
	})

	// Should still have headers with formatting
	assert.Contains(t, output, "ID")
	assert.Contains(t, output, "NAME")
	assert.Contains(t, output, "ACTIVE")
	assert.Contains(t, output, "│")
	assert.Contains(t, output, "─")
	// Verify there's a header row and separator but no data rows
	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.Equal(t, 2, len(lines), "Empty table should have only header and separator")
}

func TestPrintPrettyTable_NilSlice(t *testing.T) {
	// Create nil slice
	var data []SimpleStruct = nil

	// Capture the output
	output := CaptureOutput(func() {
		err := PrintOutput(data, "table", noColor)
		assert.NoError(t, err)
	})

	// Should still have headers with formatting
	assert.Contains(t, output, "ID")
	assert.Contains(t, output, "NAME")
	assert.Contains(t, output, "ACTIVE")
	assert.Contains(t, output, "│")
	assert.Contains(t, output, "─")
}

func TestPrintPrettyTable_CustomHeaders(t *testing.T) {
	// Test custom header tags
	data := []CustomTagStruct{
		{ID: 1, Name: "Custom Item"},
	}

	// Capture the output
	output := CaptureOutput(func() {
		err := PrintOutput(data, "table", noColor)
		assert.NoError(t, err)
	})

	// Check for custom headers in uppercase
	assert.Contains(t, output, "CUSTOM ID")
	assert.Contains(t, output, "CUSTOM NAME")
}

func TestTableColorization(t *testing.T) {
	// Define a struct with fields that should trigger different colorization rules
	type ColorTestStruct struct {
		ID       string `json:"id" header:"ID"`
		Status   string `json:"status" header:"STATUS"`
		Name     string `json:"name" header:"NAME"`
		Price    string `json:"price" header:"PRICE"`
		Speed    int    `json:"speed" header:"SPEED"`
		Location string `json:"location" header:"LOCATION"`
		Empty    string `json:"empty" header:"EMPTY"`
	}

	data := []ColorTestStruct{
		{
			ID:       "test-123",
			Status:   "ACTIVE",
			Name:     "Test Item",
			Price:    "99.99",
			Speed:    1000,
			Location: "NYC",
			Empty:    "",
		},
	}

	// Test with color enabled
	outputColor := CaptureOutput(func() {
		err := PrintOutput(data, "table", false) // false = color enabled
		assert.NoError(t, err)
	})

	// Test with color disabled
	outputNoColor := CaptureOutput(func() {
		err := PrintOutput(data, "table", true) // true = color disabled
		assert.NoError(t, err)
	})

	// Color output should have ANSI sequences
	hasColorCodes := strings.Contains(outputColor, "\033[") ||
		strings.Contains(outputColor, "\u001b[")

	// No color output should not have ANSI sequences
	noColorCodes := !strings.Contains(outputNoColor, "\033[") &&
		!strings.Contains(outputNoColor, "\u001b[")

	assert.True(t, hasColorCodes, "Color output should contain ANSI color codes")
	assert.True(t, noColorCodes, "No-color output should not contain ANSI color codes")
}

func TestPrintPrettyTable_TableStyle(t *testing.T) {
	// Create test data
	data := []SimpleStruct{
		{ID: 1, Name: "Test Item", Active: true},
	}

	// Capture the output
	output := CaptureOutput(func() {
		err := PrintOutput(data, "table", noColor)
		assert.NoError(t, err)
	})

	// Check for Megaport style elements - box drawing characters
	assert.Contains(t, output, "│") // Vertical separator
	assert.Contains(t, output, "─") // Horizontal separator

	// Check header capitalization
	assert.Contains(t, output, "ID")
	assert.Contains(t, output, "NAME")
	assert.Contains(t, output, "ACTIVE")

	// Check proper spacing (with padding)
	assert.Contains(t, output, " ID ")
	assert.Contains(t, output, " NAME ")
	assert.Contains(t, output, " ACTIVE ")
}
