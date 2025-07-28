//go:build !js && !wasm
// +build !js,!wasm

package wasm

import (
	"bytes"
)

// WasmOutputBuffer is a stub for non-WASM builds
var WasmOutputBuffer = &StubBuffer{
	buffer: &bytes.Buffer{},
}

// StubBuffer provides a no-op implementation for non-WASM builds
type StubBuffer struct {
	buffer *bytes.Buffer
}

func (d *StubBuffer) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (d *StubBuffer) String() string {
	return ""
}

func (d *StubBuffer) Reset() {
	// No-op
}

// No-op function stubs for non-WASM environments
func ResetOutputBuffers()       {}
func GetCapturedOutput() string { return "" }
func SetupIO()                  {}
func CaptureOutput(fn func()) string {
	fn()
	return ""
}
func SplitArgs(cmdString string) []string { return nil }
