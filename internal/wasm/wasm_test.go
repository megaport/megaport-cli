//go:build js && wasm
// +build js,wasm

package wasm

import (
	"bytes"
	"encoding/json"
	"strings"
	"syscall/js"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestResetOutputBuffers verifies that buffer reset clears all content
func TestResetOutputBuffers(t *testing.T) {
	// Write some data to buffers
	stdoutBuffer.WriteString("test stdout")
	stderrBuffer.WriteString("test stderr")
	WasmOutputBuffer.Write([]byte("test direct"))

	// Set globals
	js.Global().Set("wasmJSONOutput", "test json")
	js.Global().Set("wasmCSVOutput", "test csv")
	js.Global().Set("wasmTableOutput", "test table")

	// Reset
	ResetOutputBuffers()

	// Verify all buffers are empty
	assert.Equal(t, "", stdoutBuffer.String(), "stdout buffer should be empty")
	assert.Equal(t, "", stderrBuffer.String(), "stderr buffer should be empty")
	assert.Equal(t, "", WasmOutputBuffer.String(), "direct buffer should be empty")

	// Verify globals are cleared
	assert.True(t, js.Global().Get("wasmJSONOutput").IsUndefined(), "wasmJSONOutput should be undefined")
	assert.True(t, js.Global().Get("wasmCSVOutput").IsUndefined(), "wasmCSVOutput should be undefined")
	assert.True(t, js.Global().Get("wasmTableOutput").IsUndefined(), "wasmTableOutput should be undefined")
}

// TestGetCapturedOutput_Priority verifies output priority order
func TestGetCapturedOutput_Priority(t *testing.T) {
	tests := []struct {
		name           string
		setupFn        func()
		expectedSource string
		expectedOutput string
	}{
		{
			name: "JSON has highest priority",
			setupFn: func() {
				ResetOutputBuffers()
				js.Global().Set("wasmJSONOutput", "json output")
				js.Global().Set("wasmCSVOutput", "csv output")
				js.Global().Set("wasmTableOutput", "table output")
				stdoutBuffer.WriteString("stdout output")
				WasmOutputBuffer.Write([]byte("direct output"))
			},
			expectedSource: "JSON buffer",
			expectedOutput: "json output",
		},
		{
			name: "CSV has second priority",
			setupFn: func() {
				ResetOutputBuffers()
				js.Global().Set("wasmCSVOutput", "csv output")
				js.Global().Set("wasmTableOutput", "table output")
				stdoutBuffer.WriteString("stdout output")
				WasmOutputBuffer.Write([]byte("direct output"))
			},
			expectedSource: "CSV buffer",
			expectedOutput: "csv output",
		},
		{
			name: "Table has third priority",
			setupFn: func() {
				ResetOutputBuffers()
				js.Global().Set("wasmTableOutput", "table output")
				stdoutBuffer.WriteString("stdout output")
				WasmOutputBuffer.Write([]byte("direct output"))
			},
			expectedSource: "table buffer",
			expectedOutput: "table output",
		},
		{
			name: "Direct buffer has fourth priority",
			setupFn: func() {
				ResetOutputBuffers()
				stdoutBuffer.WriteString("stdout output")
				WasmOutputBuffer.Write([]byte("direct output"))
			},
			expectedSource: "direct buffer",
			expectedOutput: "direct output",
		},
		{
			name: "Stdout/stderr is lowest priority",
			setupFn: func() {
				ResetOutputBuffers()
				stdoutBuffer.WriteString("stdout output")
				stderrBuffer.WriteString("stderr output")
			},
			expectedSource: "combined stdout/stderr",
			expectedOutput: "stdout outputstderr output",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFn()
			output := GetCapturedOutput()
			assert.Equal(t, tt.expectedOutput, output)
		})
	}
}

// TestDirectOutputBuffer_Write verifies thread-safe writes
func TestDirectOutputBuffer_Write(t *testing.T) {
	buffer := &DirectOutputBuffer{
		buffer: &bytes.Buffer{},
	}

	testData := []byte("test data")
	n, err := buffer.Write(testData)

	assert.NoError(t, err)
	assert.Equal(t, len(testData), n)
	assert.Equal(t, "test data", buffer.String())
}

// TestDirectOutputBuffer_Concurrent verifies thread safety
func TestDirectOutputBuffer_Concurrent(t *testing.T) {
	buffer := &DirectOutputBuffer{
		buffer: &bytes.Buffer{},
	}

	done := make(chan bool)
	iterations := 100

	// Multiple goroutines writing concurrently
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < iterations; j++ {
				buffer.Write([]byte("x"))
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should have exactly 1000 'x' characters (10 goroutines * 100 iterations)
	result := buffer.String()
	assert.Equal(t, 1000, len(result))
	assert.Equal(t, strings.Repeat("x", 1000), result)
}

// TestSplitArgs verifies command argument parsing
func TestSplitArgs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "simple command",
			input:    "locations list",
			expected: []string{"locations", "list"},
		},
		{
			name:     "command with flags",
			input:    "ports list --format json",
			expected: []string{"ports", "list", "--format", "json"},
		},
		{
			name:     "command with double quotes",
			input:    `ports create --name "My Port"`,
			expected: []string{"ports", "create", "--name", "My Port"},
		},
		{
			name:     "command with single quotes",
			input:    `ports create --name 'My Port'`,
			expected: []string{"ports", "create", "--name", "My Port"},
		},
		{
			name:     "removes program name",
			input:    "megaport-cli locations list",
			expected: []string{"locations", "list"},
		},
		{
			name:     "removes program name with path",
			input:    "./megaport-cli locations list",
			expected: []string{"locations", "list"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "whitespace only",
			input:    "   ",
			expected: []string{},
		},
		{
			name:     "complex with mixed quotes",
			input:    `ports create --name "Port 1" --description 'Test port' --speed 1000`,
			expected: []string{"ports", "create", "--name", "Port 1", "--description", "Test port", "--speed", "1000"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SplitArgs(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestEnableDebugMode verifies debug mode activation
func TestEnableDebugMode(t *testing.T) {
	// Enable debug mode
	EnableDebugMode()

	// Verify debug mode is enabled
	assert.True(t, debugMode)

	// Verify JS function is registered
	wasmDebugFunc := js.Global().Get("wasmDebug")
	assert.False(t, wasmDebugFunc.IsUndefined())

	// Call the JS function and verify it returns true
	result := wasmDebugFunc.Invoke()
	assert.Equal(t, true, result.Bool())
}

// TestRegisterJSFunctions verifies all JS functions are registered
func TestRegisterJSFunctions(t *testing.T) {
	RegisterJSFunctions()

	// List of expected functions
	expectedFunctions := []string{
		"readConfigFile",
		"writeConfigFile",
		"debugAuthInfo",
		"saveToLocalStorage",
		"loadFromLocalStorage",
		"resetWasmOutput",
		"getWasmOutput",
		"toggleWasmDebug",
		"dumpBuffers",
		"logLocationCommand",
	}

	// Verify each function is registered
	for _, funcName := range expectedFunctions {
		jsFunc := js.Global().Get(funcName)
		assert.False(t, jsFunc.IsUndefined(), "Function %s should be registered", funcName)
		assert.Equal(t, js.TypeFunction, jsFunc.Type(), "Function %s should be a function", funcName)
	}
}

// TestReadConfigFile_NotFound verifies handling of missing config
func TestReadConfigFile_NotFound(t *testing.T) {
	// Clear localStorage
	js.Global().Get("localStorage").Call("removeItem", "megaport_fs_test.json")

	// Call readConfigFile
	result := readConfigFile(js.Null(), []js.Value{js.ValueOf("test.json")})

	// Should return error map
	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok, "Result should be a map")
	assert.Contains(t, resultMap, "error")
}

// TestReadConfigFile_ConfigJSON verifies default config creation
func TestReadConfigFile_ConfigJSON(t *testing.T) {
	// Clear localStorage
	js.Global().Get("localStorage").Call("removeItem", "megaport_fs_config.json")

	// Call readConfigFile for config.json
	result := readConfigFile(js.Null(), []js.Value{js.ValueOf("config.json")})

	// Should return default config
	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok, "Result should be a map")
	assert.Contains(t, resultMap, "content")

	// Parse the content
	var configData map[string]interface{}
	err := json.Unmarshal([]byte(resultMap["content"].(string)), &configData)
	assert.NoError(t, err)

	// Verify default structure
	assert.Contains(t, configData, "profiles")
	assert.Contains(t, configData, "activeProfile")
	assert.Contains(t, configData, "defaults")
}

// TestWriteConfigFile verifies file writing to localStorage
func TestWriteConfigFile(t *testing.T) {
	testContent := `{"test": "data"}`

	// Write file
	result := writeConfigFile(js.Null(), []js.Value{
		js.ValueOf("test.json"),
		js.ValueOf(testContent),
	})

	// Should return success
	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok, "Result should be a map")
	assert.Contains(t, resultMap, "success")
	assert.Equal(t, true, resultMap["success"])

	// Verify it was written to localStorage
	stored := js.Global().Get("localStorage").Call("getItem", "megaport_fs_test.json")
	assert.False(t, stored.IsNull())
	assert.Equal(t, testContent, stored.String())
}

// TestSaveToLocalStorage verifies localStorage save
func TestSaveToLocalStorage(t *testing.T) {
	result := saveToLocalStorage(js.Null(), []js.Value{
		js.ValueOf("test_key"),
		js.ValueOf("test_value"),
	})

	assert.Equal(t, true, result.(bool))

	// Verify in localStorage
	stored := js.Global().Get("localStorage").Call("getItem", "test_key")
	assert.Equal(t, "test_value", stored.String())
}

// TestLoadFromLocalStorage verifies localStorage load
func TestLoadFromLocalStorage(t *testing.T) {
	// Set a value
	js.Global().Get("localStorage").Call("setItem", "test_key", "test_value")

	// Load it
	result := loadFromLocalStorage(js.Null(), []js.Value{
		js.ValueOf("test_key"),
	})

	assert.Equal(t, js.TypeString, result.(js.Value).Type())
	assert.Equal(t, "test_value", result.(js.Value).String())
}

// TestResetWasmOutput_JSFunction verifies JS function for resetting output
func TestResetWasmOutput_JSFunction(t *testing.T) {
	RegisterJSFunctions()

	// Add some content to buffers
	stdoutBuffer.WriteString("test")
	WasmOutputBuffer.Write([]byte("test"))

	// Call the JS function
	resetFunc := js.Global().Get("resetWasmOutput")
	result := resetFunc.Invoke()

	// Should return true
	assert.Equal(t, true, result.Bool())

	// Buffers should be empty
	assert.Equal(t, "", stdoutBuffer.String())
	assert.Equal(t, "", WasmOutputBuffer.String())
}

// TestGetWasmOutput_JSFunction verifies JS function for getting output
func TestGetWasmOutput_JSFunction(t *testing.T) {
	RegisterJSFunctions()
	ResetOutputBuffers()

	// Add some content
	testOutput := "test output content"
	stdoutBuffer.WriteString(testOutput)

	// Call the JS function
	getFunc := js.Global().Get("getWasmOutput")
	result := getFunc.Invoke()

	// Should return the output
	assert.Equal(t, testOutput, result.String())
}

// TestDumpBuffers_JSFunction verifies JS function for dumping all buffers
func TestDumpBuffers_JSFunction(t *testing.T) {
	RegisterJSFunctions()
	ResetOutputBuffers()

	// Add content to different buffers
	stdoutBuffer.WriteString("stdout content")
	stderrBuffer.WriteString("stderr content")
	WasmOutputBuffer.Write([]byte("direct content"))

	// Call the JS function
	dumpFunc := js.Global().Get("dumpBuffers")
	result := dumpFunc.Invoke()

	// Should return an object with all buffers
	assert.Equal(t, "stdout content", result.Get("stdout").String())
	assert.Equal(t, "stderr content", result.Get("stderr").String())
	assert.Equal(t, "direct content", result.Get("direct").String())
}

// TestToggleWasmDebug_JSFunction verifies JS function for toggling debug mode
func TestToggleWasmDebug_JSFunction(t *testing.T) {
	RegisterJSFunctions()

	// Get initial state
	toggleFunc := js.Global().Get("toggleWasmDebug")
	initialState := debugMode

	// Toggle
	result := toggleFunc.Invoke()

	// Should return opposite state
	assert.Equal(t, !initialState, result.Bool())
	assert.Equal(t, !initialState, debugMode)

	// Toggle again
	result = toggleFunc.Invoke()
	assert.Equal(t, initialState, result.Bool())
	assert.Equal(t, initialState, debugMode)
}

// TestCaptureOutput verifies output capture functionality
func TestCaptureOutput(t *testing.T) {
	testOutput := "test output from function"

	captured := CaptureOutput(func() {
		WasmOutputBuffer.Write([]byte(testOutput))
	})

	// Should contain the output (may have additional content from pipes)
	assert.Contains(t, captured, testOutput)
}

// TestDirectOutputBuffer_Reset verifies buffer reset
func TestDirectOutputBuffer_Reset(t *testing.T) {
	buffer := &DirectOutputBuffer{
		buffer: &bytes.Buffer{},
	}

	buffer.Write([]byte("test data"))
	assert.NotEqual(t, "", buffer.String())

	buffer.Reset()
	assert.Equal(t, "", buffer.String())
}

// TestCustomWriter verifies custom writer implementation
func TestCustomWriter(t *testing.T) {
	var buf bytes.Buffer
	writer := &customWriter{writer: &buf}

	testData := []byte("test data")
	n, err := writer.Write(testData)

	assert.NoError(t, err)
	assert.Equal(t, len(testData), n)
	assert.Equal(t, "test data", buf.String())
}

// TestSplitArgs_EdgeCases verifies edge cases in argument parsing
func TestSplitArgs_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "unclosed double quote",
			input:    `command --flag "unclosed`,
			expected: []string{"command", "--flag", "unclosed"},
		},
		{
			name:     "unclosed single quote",
			input:    `command --flag 'unclosed`,
			expected: []string{"command", "--flag", "unclosed"},
		},
		{
			name:     "empty quotes",
			input:    `command --flag "" --other ''`,
			expected: []string{"command", "--flag", "", "--other", ""},
		},
		{
			name:     "multiple spaces",
			input:    "command    with    spaces",
			expected: []string{"command", "with", "spaces"},
		},
		{
			name:     "quotes with spaces inside",
			input:    `command "arg with   spaces"`,
			expected: []string{"command", "arg with   spaces"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SplitArgs(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestBufferThreadSafety verifies all buffers are thread-safe
func TestBufferThreadSafety(t *testing.T) {
	ResetOutputBuffers()

	done := make(chan bool)
	iterations := 100

	// Test stdout buffer
	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < iterations; j++ {
				bufferMutex.Lock()
				stdoutBuffer.WriteString("x")
				bufferMutex.Unlock()
			}
			done <- true
		}()
	}

	// Wait for completion
	for i := 0; i < 5; i++ {
		<-done
	}

	// Should have exactly 500 'x' characters
	result := stdoutBuffer.String()
	assert.Equal(t, 500, len(result))
}
