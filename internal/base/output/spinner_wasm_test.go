//go:build js && wasm

package output

import (
	"testing"

	"github.com/megaport/megaport-cli/internal/wasm"
	"github.com/stretchr/testify/assert"
)

// TestWasmSpinner_StopWithSuccess verifies the WasmSpinner implementation
// writes the success message to the captured output buffer.
func TestWasmSpinner_StopWithSuccess(t *testing.T) {
	t.Cleanup(func() {
		ResetState()
		wasm.WasmOutputBuffer.Reset()
	})
	wasm.WasmOutputBuffer.Reset()
	SetVerbosity("normal")

	spinner := NewWasmSpinner("Working...", true, "table")
	spinner.StopWithSuccess("Successfully logged in to Megaport")

	assert.Contains(t, wasm.WasmOutputBuffer.String(), "Successfully logged in to Megaport")
}

// TestSpinner_StopWithSuccess_DelegatesToWasm verifies that Spinner.StopWithSuccess
// delegates to its wasmSpinner field (typed as SpinnerInterface) and routes into
// the captured buffer rather than os.Stderr. This guards against the interface
// regressing to only Start/Stop, which would make delegation impossible.
func TestSpinner_StopWithSuccess_DelegatesToWasm(t *testing.T) {
	t.Cleanup(func() {
		ResetState()
		wasm.WasmOutputBuffer.Reset()
	})
	wasm.WasmOutputBuffer.Reset()
	SetVerbosity("normal")

	spinner := NewSpinnerWithOutput(true, "table")
	spinner.StopWithSuccess("Found location with ID: 123")

	assert.Contains(t, wasm.WasmOutputBuffer.String(), "Found location with ID: 123")
}

// TestSpinner_StopWithSuccess_QuietSuppressesWasm verifies quiet mode still
// suppresses the success message when delegating to the WASM spinner.
func TestSpinner_StopWithSuccess_QuietSuppressesWasm(t *testing.T) {
	t.Cleanup(func() {
		ResetState()
		wasm.WasmOutputBuffer.Reset()
	})
	wasm.WasmOutputBuffer.Reset()
	SetVerbosity("quiet")

	spinner := NewSpinnerWithOutput(true, "table")
	spinner.StopWithSuccess("done")

	assert.Empty(t, wasm.WasmOutputBuffer.String(), "StopWithSuccess should produce no output in quiet mode")
}
