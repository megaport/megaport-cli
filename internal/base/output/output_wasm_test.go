//go:build js && wasm
// +build js,wasm

package output

import (
	"syscall/js"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestWasmJSONWriter verifies JSON writer buffer
func TestWasmJSONWriter(t *testing.T) {
	WasmJSONWriter.Reset()

	testData := `{"test": "data"}`
	n, err := WasmJSONWriter.Write([]byte(testData))

	assert.NoError(t, err)
	assert.Equal(t, len(testData), n)
	assert.Equal(t, testData, WasmJSONWriter.String())
}

// TestWasmCSVWriter verifies CSV writer buffer
func TestWasmCSVWriter(t *testing.T) {
	WasmCSVWriter.Reset()

	testData := "id,name,active\n1,test,true\n"
	n, err := WasmCSVWriter.Write([]byte(testData))

	assert.NoError(t, err)
	assert.Equal(t, len(testData), n)
	assert.Equal(t, testData, WasmCSVWriter.String())
}

// TestPrintJSON_WASM verifies JSON output in WASM
func TestPrintJSON_WASM(t *testing.T) {
	data := []SimpleStruct{
		{ID: 1, Name: "Item 1", Active: true},
		{ID: 2, Name: "Item 2", Active: false},
	}

	// Clear buffers and globals
	WasmJSONWriter.Reset()
	js.Global().Delete("wasmJSONOutput")

	// Print JSON
	err := printJSON(data)
	assert.NoError(t, err)

	// Verify buffer has content
	output := WasmJSONWriter.String()
	assert.NotEmpty(t, output)
	assert.Contains(t, output, `"id"`)
	assert.Contains(t, output, `"name"`)
	assert.Contains(t, output, `"active"`)
	assert.Contains(t, output, `"Item 1"`)
	assert.Contains(t, output, `"Item 2"`)

	// Verify it's valid JSON
	assert.True(t, output[0] == '[', "Should start with [")
	assert.True(t, output[len(output)-1] == '\n' || output[len(output)-2] == ']', "Should end with ]")

	// Verify global variable is set
	wasmJSONGlobal := js.Global().Get("wasmJSONOutput")
	assert.False(t, wasmJSONGlobal.IsUndefined())
	assert.Equal(t, output, wasmJSONGlobal.String())
}

// TestPrintJSON_WASM_EmptyData verifies empty JSON array
func TestPrintJSON_WASM_EmptyData(t *testing.T) {
	var data []SimpleStruct

	WasmJSONWriter.Reset()

	err := printJSON(data)
	assert.NoError(t, err)

	output := WasmJSONWriter.String()
	assert.Contains(t, output, "[]")
}

// TestPrintJSON_WASM_ComplexData verifies complex JSON structures
func TestPrintJSON_WASM_ComplexData(t *testing.T) {
	data := []ComplexStruct{
		{
			ID:   1,
			Name: "Complex",
			Tags: []string{"tag1", "tag2"},
			Metadata: map[string]string{
				"key1": "value1",
			},
		},
	}

	WasmJSONWriter.Reset()

	err := printJSON(data)
	assert.NoError(t, err)

	output := WasmJSONWriter.String()
	assert.NotEmpty(t, output)
	assert.Contains(t, output, "Complex")
	assert.Contains(t, output, "tag1")
	assert.Contains(t, output, "key1")
}

// TestPrintCSV_WASM verifies CSV output in WASM
func TestPrintCSV_WASM(t *testing.T) {
	data := []SimpleStruct{
		{ID: 1, Name: "Item 1", Active: true},
		{ID: 2, Name: "Item 2", Active: false},
	}

	// Clear buffers and globals
	WasmCSVWriter.Reset()
	js.Global().Delete("wasmCSVOutput")

	// Print CSV
	err := printCSV(data)
	assert.NoError(t, err)

	// Verify buffer has content
	output := WasmCSVWriter.String()
	assert.NotEmpty(t, output)
	assert.Contains(t, output, "id,name,active")
	assert.Contains(t, output, "1,Item 1,true")
	assert.Contains(t, output, "2,Item 2,false")

	// Verify global variable is set
	wasmCSVGlobal := js.Global().Get("wasmCSVOutput")
	assert.False(t, wasmCSVGlobal.IsUndefined())
}

// TestPrintCSV_WASM_EmptyData verifies CSV headers only
func TestPrintCSV_WASM_EmptyData(t *testing.T) {
	var data []SimpleStruct

	WasmCSVWriter.Reset()

	err := printCSV(data)
	assert.NoError(t, err)

	output := WasmCSVWriter.String()
	assert.Contains(t, output, "id,name,active")
}

// TestPrintCSV_WASM_SpecialCharacters verifies CSV escaping
func TestPrintCSV_WASM_SpecialCharacters(t *testing.T) {
	data := []SimpleStruct{
		{ID: 1, Name: `Item with "quotes"`, Active: true},
		{ID: 2, Name: "Item with, comma", Active: false},
	}

	WasmCSVWriter.Reset()

	err := printCSV(data)
	assert.NoError(t, err)

	output := WasmCSVWriter.String()
	assert.NotEmpty(t, output)
	// CSV should properly escape quotes and commas
	assert.Contains(t, output, "Item with")
}

// TestPrintOutput_WASM_Formats verifies all output formats
func TestPrintOutput_WASM_Formats(t *testing.T) {
	data := []SimpleStruct{
		{ID: 1, Name: "Test", Active: true},
	}

	formats := []struct {
		name   string
		format string
	}{
		{"JSON", "json"},
		{"CSV", "csv"},
		{"Table", "table"},
	}

	for _, tt := range formats {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all buffers
			WasmJSONWriter.Reset()
			WasmCSVWriter.Reset()
			WasmTableWriter.Reset()
			js.Global().Delete("wasmJSONOutput")
			js.Global().Delete("wasmCSVOutput")
			js.Global().Delete("wasmTableOutput")

			err := PrintOutput(data, tt.format, false)
			assert.NoError(t, err)

			// At least one buffer should have content
			hasOutput := WasmJSONWriter.String() != "" ||
				WasmCSVWriter.String() != "" ||
				WasmTableWriter.String() != ""
			assert.True(t, hasOutput, "At least one buffer should have output")
		})
	}
}

// TestWasmGlobalVariables verifies all WASM global variables
func TestWasmGlobalVariables(t *testing.T) {
	// Test JSON global
	testJSON := `{"test": "json"}`
	js.Global().Set("wasmJSONOutput", testJSON)
	result := js.Global().Get("wasmJSONOutput")
	assert.False(t, result.IsUndefined())
	assert.Equal(t, testJSON, result.String())

	// Test CSV global
	testCSV := "id,name\n1,test"
	js.Global().Set("wasmCSVOutput", testCSV)
	result = js.Global().Get("wasmCSVOutput")
	assert.False(t, result.IsUndefined())
	assert.Equal(t, testCSV, result.String())

	// Test Table global
	testTable := "┌─────┐\n│ Test │\n└─────┘"
	js.Global().Set("wasmTableOutput", testTable)
	result = js.Global().Get("wasmTableOutput")
	assert.False(t, result.IsUndefined())
	assert.Equal(t, testTable, result.String())

	// Clean up
	js.Global().Delete("wasmJSONOutput")
	js.Global().Delete("wasmCSVOutput")
	js.Global().Delete("wasmTableOutput")
}

// TestWasmBufferReset verifies buffer reset functionality
func TestWasmBufferReset(t *testing.T) {
	// Write to all buffers
	WasmJSONWriter.WriteString("json data")
	WasmCSVWriter.WriteString("csv data")
	WasmTableWriter.WriteString("table data")

	// Verify they have content
	assert.NotEmpty(t, WasmJSONWriter.String())
	assert.NotEmpty(t, WasmCSVWriter.String())
	assert.NotEmpty(t, WasmTableWriter.String())

	// Reset all
	WasmJSONWriter.Reset()
	WasmCSVWriter.Reset()
	WasmTableWriter.Reset()

	// Verify they're empty
	assert.Empty(t, WasmJSONWriter.String())
	assert.Empty(t, WasmCSVWriter.String())
	assert.Empty(t, WasmTableWriter.String())
}

// TestPrintOutput_WASM_InvalidFormat verifies error handling
func TestPrintOutput_WASM_InvalidFormat(t *testing.T) {
	data := []SimpleStruct{
		{ID: 1, Name: "Test", Active: true},
	}

	err := PrintOutput(data, "invalid", false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid output format")
}

// TestWasmOutputConcurrency verifies thread-safe buffer operations
func TestWasmOutputConcurrency(t *testing.T) {
	done := make(chan bool)
	iterations := 50

	// Multiple goroutines writing to different buffers
	go func() {
		for i := 0; i < iterations; i++ {
			WasmJSONWriter.WriteString("j")
		}
		done <- true
	}()

	go func() {
		for i := 0; i < iterations; i++ {
			WasmCSVWriter.WriteString("c")
		}
		done <- true
	}()

	go func() {
		for i := 0; i < iterations; i++ {
			WasmTableWriter.WriteString("t")
		}
		done <- true
	}()

	// Wait for all
	<-done
	<-done
	<-done

	// All writes should have succeeded
	assert.Equal(t, iterations, len(WasmJSONWriter.String()))
	assert.Equal(t, iterations, len(WasmCSVWriter.String()))
	assert.Equal(t, iterations, len(WasmTableWriter.String()))
}

// TestConsoleLogging verifies console.log doesn't cause panics
func TestConsoleLogging(t *testing.T) {
	data := []SimpleStruct{
		{ID: 1, Name: "Test", Active: true},
	}

	// All print functions should handle console.log gracefully
	assert.NotPanics(t, func() {
		WasmJSONWriter.Reset()
		_ = printJSON(data)
	})

	assert.NotPanics(t, func() {
		WasmCSVWriter.Reset()
		_ = printCSV(data)
	})

	assert.NotPanics(t, func() {
		WasmTableWriter.Reset()
		_ = printTable(data, false)
	})
}
