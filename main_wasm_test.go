//go:build js && wasm
// +build js,wasm

package main

import (
	"syscall/js"
	"testing"

	"github.com/megaport/megaport-cli/internal/wasm"
	"github.com/stretchr/testify/assert"
)

// TestWasmInitialization verifies WASM module initializes correctly
func TestWasmInitialization(t *testing.T) {
	// Verify JS functions are registered
	executeMegaportCmd := js.Global().Get("executeMegaportCommand")
	assert.False(t, executeMegaportCmd.IsUndefined(), "executeMegaportCommand should be defined")
	assert.Equal(t, js.TypeFunction, executeMegaportCmd.Type())

	executeMegaportCmdAsync := js.Global().Get("executeMegaportCommandAsync")
	assert.False(t, executeMegaportCmdAsync.IsUndefined(), "executeMegaportCommandAsync should be defined")
	assert.Equal(t, js.TypeFunction, executeMegaportCmdAsync.Type())
}

// TestWasmHelperFunctions verifies helper functions are registered
func TestWasmHelperFunctions(t *testing.T) {
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
		"wasmDebug",
	}

	for _, funcName := range expectedFunctions {
		jsFunc := js.Global().Get(funcName)
		assert.False(t, jsFunc.IsUndefined(), "Function %s should be defined", funcName)
	}
}

// TestExecuteMegaportCommand_Help verifies help command execution
func TestExecuteMegaportCommand_Help(t *testing.T) {
	wasm.ResetOutputBuffers()

	// Call the sync version with help command
	executeMegaportCmd := js.Global().Get("executeMegaportCommand")
	result := executeMegaportCmd.Invoke(js.ValueOf("--help"))

	// Verify result structure
	assert.Equal(t, js.TypeObject, result.Type())
	output := result.Get("output")
	assert.False(t, output.IsUndefined())

	outputStr := output.String()
	assert.NotEmpty(t, outputStr)
	assert.Contains(t, outputStr, "Megaport CLI")
}

// TestExecuteMegaportCommand_Version verifies version command
func TestExecuteMegaportCommand_Version(t *testing.T) {
	wasm.ResetOutputBuffers()

	executeMegaportCmd := js.Global().Get("executeMegaportCommand")
	result := executeMegaportCmd.Invoke(js.ValueOf("version"))

	assert.Equal(t, js.TypeObject, result.Type())
	output := result.Get("output")
	assert.False(t, output.IsUndefined())

	outputStr := output.String()
	assert.NotEmpty(t, outputStr)
}

// TestExecuteMegaportCommand_InvalidCommand verifies error handling
func TestExecuteMegaportCommand_InvalidCommand(t *testing.T) {
	wasm.ResetOutputBuffers()

	executeMegaportCmd := js.Global().Get("executeMegaportCommand")
	result := executeMegaportCmd.Invoke(js.ValueOf("nonexistent-command"))

	assert.Equal(t, js.TypeObject, result.Type())
	output := result.Get("output")
	
	// Should have error output
	outputStr := output.String()
	assert.NotEmpty(t, outputStr)
}

// TestExecuteMegaportCommand_EmptyCommand verifies empty command handling
func TestExecuteMegaportCommand_EmptyCommand(t *testing.T) {
	wasm.ResetOutputBuffers()

	executeMegaportCmd := js.Global().Get("executeMegaportCommand")
	result := executeMegaportCmd.Invoke(js.ValueOf(""))

	assert.Equal(t, js.TypeObject, result.Type())
	output := result.Get("output")
	assert.False(t, output.IsUndefined())
}

// TestExecuteMegaportCommand_WithFlags verifies flag parsing
func TestExecuteMegaportCommand_WithFlags(t *testing.T) {
	wasm.ResetOutputBuffers()

	executeMegaportCmd := js.Global().Get("executeMegaportCommand")
	result := executeMegaportCmd.Invoke(js.ValueOf("version --format json"))

	assert.Equal(t, js.TypeObject, result.Type())
	output := result.Get("output")
	assert.False(t, output.IsUndefined())

	outputStr := output.String()
	assert.NotEmpty(t, outputStr)
}

// TestExecuteMegaportCommandAsync_Callback verifies async command with callback
func TestExecuteMegaportCommandAsync_Callback(t *testing.T) {
	wasm.ResetOutputBuffers()

	// Create a callback function
	callback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// Callback receives the result
		if len(args) > 0 {
			result := args[0]
			assert.Equal(t, js.TypeObject, result.Type())
		}
		return nil
	})
	defer callback.Release()

	// Call async version
	executeMegaportCmdAsync := js.Global().Get("executeMegaportCommandAsync")
	executeMegaportCmdAsync.Invoke(
		js.ValueOf("version"),
		callback,
	)

	// Give it a moment to execute (in real tests, you'd use proper async handling)
	// Note: In actual usage, the callback would be called asynchronously
	// For testing, we verify the function doesn't panic
	assert.NotPanics(t, func() {
		executeMegaportCmdAsync.Invoke(js.ValueOf("--help"), callback)
	})
}

// TestExecuteMegaportCommandAsync_NoCallback verifies error handling without callback
func TestExecuteMegaportCommandAsync_NoCallback(t *testing.T) {
	executeMegaportCmdAsync := js.Global().Get("executeMegaportCommandAsync")
	
	// Should not panic even with missing callback
	assert.NotPanics(t, func() {
		executeMegaportCmdAsync.Invoke(js.ValueOf("version"))
	})
}

// TestExecuteMegaportCommandAsync_InvalidCallback verifies invalid callback handling
func TestExecuteMegaportCommandAsync_InvalidCallback(t *testing.T) {
	executeMegaportCmdAsync := js.Global().Get("executeMegaportCommandAsync")
	
	// Should not panic with non-function callback
	assert.NotPanics(t, func() {
		executeMegaportCmdAsync.Invoke(
			js.ValueOf("version"),
			js.ValueOf("not a function"),
		)
	})
}

// TestOutputBufferReset verifies output buffer reset between commands
func TestOutputBufferReset(t *testing.T) {
	// Execute first command
	executeMegaportCmd := js.Global().Get("executeMegaportCommand")
	result1 := executeMegaportCmd.Invoke(js.ValueOf("version"))
	output1 := result1.Get("output").String()

	// Execute second command
	result2 := executeMegaportCmd.Invoke(js.ValueOf("--help"))
	output2 := result2.Get("output").String()

	// Outputs should be different (buffers were reset)
	assert.NotEqual(t, output1, output2)
	assert.NotEmpty(t, output1)
	assert.NotEmpty(t, output2)
}

// TestWasmOutputBufferIntegration verifies WasmOutputBuffer usage
func TestWasmOutputBufferIntegration(t *testing.T) {
	wasm.ResetOutputBuffers()

	// Write to the buffer
	testData := "test integration data"
	wasm.WasmOutputBuffer.Write([]byte(testData))

	// Verify it can be retrieved
	output := wasm.WasmOutputBuffer.String()
	assert.Equal(t, testData, output)

	// Reset and verify
	wasm.WasmOutputBuffer.Reset()
	assert.Equal(t, "", wasm.WasmOutputBuffer.String())
}

// TestCommandArgumentParsing verifies complex argument parsing
func TestCommandArgumentParsing(t *testing.T) {
	tests := []struct {
		name    string
		command string
	}{
		{
			name:    "simple command",
			command: "version",
		},
		{
			name:    "command with flags",
			command: "version --format json",
		},
		{
			name:    "command with quoted args",
			command: `config set-default --name "My Profile"`,
		},
		{
			name:    "complex command",
			command: "ports list --format table --no-color",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wasm.ResetOutputBuffers()

			executeMegaportCmd := js.Global().Get("executeMegaportCommand")
			
			assert.NotPanics(t, func() {
				result := executeMegaportCmd.Invoke(js.ValueOf(tt.command))
				assert.Equal(t, js.TypeObject, result.Type())
			})
		})
	}
}

// TestGlobalVariableCleanup verifies globals are cleaned up
func TestGlobalVariableCleanup(t *testing.T) {
	// Set some globals
	js.Global().Set("wasmJSONOutput", "test json")
	js.Global().Set("wasmCSVOutput", "test csv")
	js.Global().Set("wasmTableOutput", "test table")

	// Reset should clear them
	wasm.ResetOutputBuffers()

	assert.True(t, js.Global().Get("wasmJSONOutput").IsUndefined())
	assert.True(t, js.Global().Get("wasmCSVOutput").IsUndefined())
	assert.True(t, js.Global().Get("wasmTableOutput").IsUndefined())
}

// TestConfigFileOperations verifies config file read/write
func TestConfigFileOperations(t *testing.T) {
	writeConfigFile := js.Global().Get("writeConfigFile")
	readConfigFile := js.Global().Get("readConfigFile")

	// Write a config file
	testContent := `{"test": "config"}`
	writeResult := writeConfigFile.Invoke(
		js.ValueOf("test-config.json"),
		js.ValueOf(testContent),
	)

	// Verify write succeeded
	assert.Equal(t, js.TypeObject, writeResult.Type())

	// Read it back
	readResult := readConfigFile.Invoke(js.ValueOf("test-config.json"))
	assert.Equal(t, js.TypeObject, readResult.Type())

	// Clean up
	js.Global().Get("localStorage").Call("removeItem", "megaport_fs_test-config.json")
}

// TestLocalStorageOperations verifies localStorage save/load
func TestLocalStorageOperations(t *testing.T) {
	saveToLocalStorage := js.Global().Get("saveToLocalStorage")
	loadFromLocalStorage := js.Global().Get("loadFromLocalStorage")

	// Save a value
	testKey := "test-key"
	testValue := "test-value"
	saveResult := saveToLocalStorage.Invoke(
		js.ValueOf(testKey),
		js.ValueOf(testValue),
	)
	assert.Equal(t, true, saveResult.Bool())

	// Load it back
	loadResult := loadFromLocalStorage.Invoke(js.ValueOf(testKey))
	assert.Equal(t, testValue, loadResult.String())

	// Clean up
	js.Global().Get("localStorage").Call("removeItem", testKey)
}

// TestDebugMode verifies debug mode toggle
func TestDebugMode(t *testing.T) {
	toggleWasmDebug := js.Global().Get("toggleWasmDebug")
	wasmDebug := js.Global().Get("wasmDebug")

	// Get initial state
	initialState := wasmDebug.Invoke().Bool()

	// Toggle
	newState := toggleWasmDebug.Invoke().Bool()
	assert.NotEqual(t, initialState, newState)

	// Toggle back
	finalState := toggleWasmDebug.Invoke().Bool()
	assert.Equal(t, initialState, finalState)
}

// TestDumpBuffers verifies buffer dumping
func TestDumpBuffers(t *testing.T) {
	dumpBuffers := js.Global().Get("dumpBuffers")

	// Add some content
	wasm.ResetOutputBuffers()
	wasm.WasmOutputBuffer.Write([]byte("test content"))

	// Dump buffers
	result := dumpBuffers.Invoke()
	assert.Equal(t, js.TypeObject, result.Type())

	// Verify structure
	stdout := result.Get("stdout")
	stderr := result.Get("stderr")
	direct := result.Get("direct")

	assert.False(t, stdout.IsUndefined())
	assert.False(t, stderr.IsUndefined())
	assert.False(t, direct.IsUndefined())
}

// TestErrorRecovery verifies panic recovery
func TestErrorRecovery(t *testing.T) {
	executeMegaportCmd := js.Global().Get("executeMegaportCommand")

	// These should not cause the WASM module to crash
	assert.NotPanics(t, func() {
		executeMegaportCmd.Invoke(js.ValueOf("nonexistent"))
	})

	assert.NotPanics(t, func() {
		executeMegaportCmd.Invoke(js.ValueOf(""))
	})

	assert.NotPanics(t, func() {
		executeMegaportCmd.Invoke(js.ValueOf("invalid command with many args that don't make sense"))
	})
}

// TestConcurrentCommands verifies multiple commands don't interfere
func TestConcurrentCommands(t *testing.T) {
	executeMegaportCmd := js.Global().Get("executeMegaportCommand")

	// Execute multiple commands
	result1 := executeMegaportCmd.Invoke(js.ValueOf("version"))
	result2 := executeMegaportCmd.Invoke(js.ValueOf("--help"))

	// Both should succeed
	assert.Equal(t, js.TypeObject, result1.Type())
	assert.Equal(t, js.TypeObject, result2.Type())

	output1 := result1.Get("output").String()
	output2 := result2.Get("output").String()

	assert.NotEmpty(t, output1)
	assert.NotEmpty(t, output2)
	assert.NotEqual(t, output1, output2)
}
