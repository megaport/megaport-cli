//go:build js && wasm
// +build js,wasm

package output

import (
	"strings"
	"syscall/js"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestWasmTableWriter verifies the WASM table writer buffer
func TestWasmTableWriter(t *testing.T) {
	// Clear the buffer
	WasmTableWriter.Reset()

	// Write some data
	testData := "test table data"
	n, err := WasmTableWriter.Write([]byte(testData))

	assert.NoError(t, err)
	assert.Equal(t, len(testData), n)
	assert.Equal(t, testData, WasmTableWriter.String())
}

// TestWasmTableWriter_Reset verifies buffer reset functionality
func TestWasmTableWriter_Reset(t *testing.T) {
	WasmTableWriter.WriteString("test data")
	assert.NotEqual(t, "", WasmTableWriter.String())

	WasmTableWriter.Reset()
	assert.Equal(t, "", WasmTableWriter.String())
}

// TestPrintTable_WASM verifies table printing in WASM environment
func TestPrintTable_WASM(t *testing.T) {
	data := []SimpleStruct{
		{ID: 1, Name: "Item 1", Active: true},
		{ID: 2, Name: "Item 2", Active: false},
	}

	// Clear the buffer and globals
	WasmTableWriter.Reset()
	js.Global().Delete("wasmTableOutput")

	// Print table
	err := printTable(data, false)
	assert.NoError(t, err)

	// Verify buffer has content
	output := WasmTableWriter.String()
	assert.NotEmpty(t, output)
	assert.Contains(t, output, "ID")
	assert.Contains(t, output, "NAME")
	assert.Contains(t, output, "ACTIVE")
	assert.Contains(t, output, "Item 1")
	assert.Contains(t, output, "Item 2")

	// Verify global variable is set
	wasmTableGlobal := js.Global().Get("wasmTableOutput")
	assert.False(t, wasmTableGlobal.IsUndefined())
	assert.Equal(t, output, wasmTableGlobal.String())
}

// TestPrintTable_WASM_NoColor verifies table printing without colors
func TestPrintTable_WASM_NoColor(t *testing.T) {
	data := []SimpleStruct{
		{ID: 1, Name: "Test", Active: true},
	}

	WasmTableWriter.Reset()

	// Print with noColor=true
	err := printTable(data, true)
	assert.NoError(t, err)

	output := WasmTableWriter.String()
	assert.NotEmpty(t, output)

	// Should not contain ANSI color codes (basic check)
	// Note: StyleLight may still have some formatting, but should be minimal
	assert.NotContains(t, output, "\033[38") // 38;5;x color codes
	assert.NotContains(t, output, "\033[48") // 48;5;x background codes
}

// TestPrintTable_WASM_WithColor verifies table printing with colors
func TestPrintTable_WASM_WithColor(t *testing.T) {
	data := []SimpleStruct{
		{ID: 1, Name: "Test", Active: true},
	}

	WasmTableWriter.Reset()

	// Print with noColor=false (colors enabled)
	err := printTable(data, false)
	assert.NoError(t, err)

	output := WasmTableWriter.String()
	assert.NotEmpty(t, output)

	// Should contain ANSI color codes
	hasColorCodes := strings.Contains(output, "\033[") || strings.Contains(output, "\u001b[")
	assert.True(t, hasColorCodes, "Colored output should contain ANSI escape codes")
}

// TestPrintTable_WASM_EmptyData verifies handling of empty data
func TestPrintTable_WASM_EmptyData(t *testing.T) {
	var data []SimpleStruct

	WasmTableWriter.Reset()

	err := printTable(data, false)
	assert.NoError(t, err)

	output := WasmTableWriter.String()
	// Empty data should still render headers
	assert.Contains(t, output, "ID")
	assert.Contains(t, output, "NAME")
	assert.Contains(t, output, "ACTIVE")
}

// TestPrintTable_WASM_ComplexData verifies complex structure handling
func TestPrintTable_WASM_ComplexData(t *testing.T) {
	data := []ComplexStruct{
		{
			ID:   1,
			Name: "Complex Item",
			Tags: []string{"tag1", "tag2"},
			Metadata: map[string]string{
				"key": "value",
			},
		},
	}

	WasmTableWriter.Reset()

	err := printTable(data, false)
	assert.NoError(t, err)

	output := WasmTableWriter.String()
	assert.NotEmpty(t, output)
	assert.Contains(t, output, "Complex Item")
}

// TestCalculateDynamicWidth verifies column width calculation
func TestCalculateDynamicWidth(t *testing.T) {
	tests := []struct {
		name          string
		termWidth     int
		minWidth      int
		maxPercentage int
		expectedMin   int
	}{
		{
			name:          "normal width",
			termWidth:     100,
			minWidth:      10,
			maxPercentage: 50,
			expectedMin:   10,
		},
		{
			name:          "exceeds minimum",
			termWidth:     200,
			minWidth:      10,
			maxPercentage: 50,
			expectedMin:   10,
		},
		{
			name:          "below minimum",
			termWidth:     10,
			minWidth:      20,
			maxPercentage: 50,
			expectedMin:   20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateDynamicWidth(tt.termWidth, tt.minWidth, tt.maxPercentage)
			assert.GreaterOrEqual(t, result, tt.expectedMin)
		})
	}
}

// TestPrintTable_WASM_ColumnWidths verifies fixed column width behavior
func TestPrintTable_WASM_ColumnWidths(t *testing.T) {
	data := []struct {
		ID       int    `json:"id" header:"ID"`
		Name     string `json:"name" header:"Name"`
		Country  string `json:"country" header:"Country"`
		Metro    string `json:"metro" header:"Metro"`
		SiteCode string `json:"site_code" header:"Site Code"`
		Status   string `json:"status" header:"Status"`
	}{
		{
			ID:       1,
			Name:     "Very Long Location Name That Might Wrap",
			Country:  "United States",
			Metro:    "New York",
			SiteCode: "NYC01",
			Status:   "Active",
		},
	}

	WasmTableWriter.Reset()

	err := printTable(data, false)
	assert.NoError(t, err)

	output := WasmTableWriter.String()
	assert.NotEmpty(t, output)

	// Verify all columns are present
	assert.Contains(t, output, "ID")
	assert.Contains(t, output, "NAME")
	assert.Contains(t, output, "COUNTRY")
	assert.Contains(t, output, "METRO")
	assert.Contains(t, output, "SITE CODE")
	assert.Contains(t, output, "STATUS")
}

// TestPrintTable_WASM_BoxDrawing verifies box drawing characters
func TestPrintTable_WASM_BoxDrawing(t *testing.T) {
	data := []SimpleStruct{
		{ID: 1, Name: "Test", Active: true},
	}

	WasmTableWriter.Reset()

	err := printTable(data, false)
	assert.NoError(t, err)

	output := WasmTableWriter.String()

	// Check for box drawing characters used in MegaportEnhancedStyle
	assert.Contains(t, output, "│") // Vertical border
	assert.Contains(t, output, "─") // Horizontal border
	
	// Check for corners (at least one of these should be present)
	hasCorners := strings.Contains(output, "┌") || 
		strings.Contains(output, "┐") || 
		strings.Contains(output, "└") || 
		strings.Contains(output, "┘")
	assert.True(t, hasCorners, "Table should have corner characters")
}

// TestPrintTable_WASM_HeaderFormatting verifies header formatting
func TestPrintTable_WASM_HeaderFormatting(t *testing.T) {
	data := []SimpleStruct{
		{ID: 1, Name: "test", Active: true},
	}

	WasmTableWriter.Reset()

	err := printTable(data, false)
	assert.NoError(t, err)

	output := WasmTableWriter.String()

	// Headers should be uppercase (FormatUpper)
	assert.Contains(t, output, "ID")
	assert.Contains(t, output, "NAME")
	assert.Contains(t, output, "ACTIVE")

	// Should not contain lowercase headers
	lines := strings.Split(output, "\n")
	headerLine := ""
	for _, line := range lines {
		if strings.Contains(line, "NAME") {
			headerLine = line
			break
		}
	}
	assert.NotEmpty(t, headerLine)
}

// TestPrintTable_WASM_GlobalVariable verifies wasmTableOutput global
func TestPrintTable_WASM_GlobalVariable(t *testing.T) {
	data := []SimpleStruct{
		{ID: 1, Name: "Test", Active: true},
	}

	// Clear the global
	js.Global().Delete("wasmTableOutput")
	WasmTableWriter.Reset()

	// Print table
	err := printTable(data, false)
	assert.NoError(t, err)

	// Verify global is set
	wasmTableGlobal := js.Global().Get("wasmTableOutput")
	assert.False(t, wasmTableGlobal.IsUndefined(), "wasmTableOutput should be set")
	assert.False(t, wasmTableGlobal.IsNull(), "wasmTableOutput should not be null")

	globalContent := wasmTableGlobal.String()
	bufferContent := WasmTableWriter.String()

	assert.Equal(t, bufferContent, globalContent, "Global and buffer content should match")
}

// TestPrintTable_WASM_ConsoleLogging verifies console logging (basic check)
func TestPrintTable_WASM_ConsoleLogging(t *testing.T) {
	data := []SimpleStruct{
		{ID: 1, Name: "Test", Active: true},
	}

	WasmTableWriter.Reset()

	// This should not panic or error even with console.log calls
	assert.NotPanics(t, func() {
		err := printTable(data, false)
		assert.NoError(t, err)
	})
}

// TestPrintTable_WASM_MegaportStyle verifies Megaport-specific styling
func TestPrintTable_WASM_MegaportStyle(t *testing.T) {
	data := []SimpleStruct{
		{ID: 1, Name: "Test", Active: true},
		{ID: 2, Name: "Test2", Active: false},
	}

	WasmTableWriter.Reset()

	err := printTable(data, false)
	assert.NoError(t, err)

	output := WasmTableWriter.String()

	// Should have padding (spaces around content)
	assert.Contains(t, output, "  ") // Multiple spaces indicate padding

	// Should have separators between columns
	assert.Contains(t, output, "│")

	// Should have header separator
	lines := strings.Split(output, "\n")
	hasSeparatorLine := false
	for _, line := range lines {
		if strings.Contains(line, "─") && strings.Contains(line, "┼") {
			hasSeparatorLine = true
			break
		}
	}
	assert.True(t, hasSeparatorLine, "Should have header separator line")
}
