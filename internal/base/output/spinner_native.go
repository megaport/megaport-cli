//go:build !wasm
// +build !wasm

package output

import (
	"time"
)

// NewSpinnerWasm creates a spinner optimized for WASM display
// This is the non-WASM build version that uses the regular spinner
func NewSpinnerWasm(noColor bool, outputFormat string) *Spinner {
	return &Spinner{
		stop:         make(chan bool, 1),
		frameRate:    150 * time.Millisecond, // Slightly slower for better visibility
		noColor:      noColor,
		outputFormat: outputFormat,
		style:        SpinnerStyleFancy,
		wasmSpinner:  nil, // No WASM spinner in non-WASM builds
	}
}
