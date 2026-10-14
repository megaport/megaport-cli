// These tests call undefined JS globals (fetch, config-file/localStorage
// bridges) that only exist in a real browser host, so they panic under the
// headless node runtime. The `browser` tag keeps them out of the default CI
// run; a browser harness opts in with -tags js,wasm,browser.
//go:build js && wasm && browser

package wasm

import (
	"encoding/json"
	"syscall/js"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestResetOutputBuffers verifies that buffer reset clears all content
func TestResetOutputBuffers(t *testing.T) {
	// Write some data to buffers
	stdoutBuffer.WriteString("test stdout")
	stderrBuffer.WriteString("test stderr")
	_, _ = WasmOutputBuffer.Write([]byte("test direct"))

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

// TestEnableDebugMode verifies debug mode activation
func TestEnableDebugMode(t *testing.T) {
	// Enable debug mode
	EnableDebugMode()

	// Verify debug mode is enabled
	assert.True(t, debugMode.Load())

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
	contentStr, ok := resultMap["content"].(string)
	assert.True(t, ok, "content should be a string")
	err := json.Unmarshal([]byte(contentStr), &configData)
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

	resultBool, ok := result.(bool)
	assert.True(t, ok, "result should be a bool")
	assert.Equal(t, true, resultBool)

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

	resultVal, ok := result.(js.Value)
	assert.True(t, ok, "result should be a js.Value")
	assert.Equal(t, js.TypeString, resultVal.Type())
	assert.Equal(t, "test_value", resultVal.String())
}

// TestResetWasmOutput_JSFunction verifies JS function for resetting output
func TestResetWasmOutput_JSFunction(t *testing.T) {
	RegisterJSFunctions()

	// Add some content to buffers
	stdoutBuffer.WriteString("test")
	_, _ = WasmOutputBuffer.Write([]byte("test"))

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
	_, _ = WasmOutputBuffer.Write([]byte("direct content"))

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
	initialState := debugMode.Load()

	// Toggle
	result := toggleFunc.Invoke()

	// Should return opposite state
	assert.Equal(t, !initialState, result.Bool())
	assert.Equal(t, !initialState, debugMode.Load())

	// Toggle again
	result = toggleFunc.Invoke()
	assert.Equal(t, initialState, result.Bool())
	assert.Equal(t, initialState, debugMode.Load())
}
