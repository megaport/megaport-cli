//go:build js && wasm
// +build js,wasm

package output

import (
	"fmt"
	"syscall/js"
	"time"

	"github.com/fatih/color"
)

// WasmSpinner is a WASM-specific spinner that uses JavaScript callbacks
// instead of trying to update in-place (which doesn't work in buffered WASM output)
type WasmSpinner struct {
	message      string
	noColor      bool
	stopped      bool
	stopChan     chan bool
	jsSpinnerID  js.Value
	outputFormat string
}

// NewWasmSpinner creates a spinner that works in the WASM environment
func NewWasmSpinner(message string, noColor bool, outputFormat string) *WasmSpinner {
	return &WasmSpinner{
		message:      message,
		noColor:      noColor,
		stopped:      false,
		stopChan:     make(chan bool),
		outputFormat: outputFormat,
	}
}

// Start begins the spinner animation via JavaScript
// This implements SpinnerInterface.Start(message string)
func (s *WasmSpinner) Start(message string) {
	// Store the message for later use
	s.message = message
	
	if js.Global().Get("wasmStartSpinner").IsUndefined() {
		// Fallback: just print the message
		s.printStartMessage()
		return
	}

	// Call JavaScript function to start spinner animation
	s.jsSpinnerID = js.Global().Call("wasmStartSpinner", message)
	
	// Also log to console for debugging
	js.Global().Get("console").Call("log", "🔄 WASM Spinner started (ID: "+s.jsSpinnerID.String()+"): "+message)
}

// Stop stops the spinner
// This implements SpinnerInterface.Stop()
func (s *WasmSpinner) Stop() {
	if s.stopped {
		return
	}
	s.stopped = true
	
	// Stop the JavaScript spinner
	if !s.jsSpinnerID.IsUndefined() && !s.jsSpinnerID.IsNull() {
		if !js.Global().Get("wasmStopSpinner").IsUndefined() {
			js.Global().Call("wasmStopSpinner", s.jsSpinnerID)
		}
	}
	
	js.Global().Get("console").Call("log", "⏹️ WASM Spinner stopped: "+s.message)
}

// StopWithSuccess stops the spinner and shows a success message
func (s *WasmSpinner) StopWithSuccess(msg string) {
	s.Stop()
	
	if s.noColor {
		fmt.Printf("✓ %s\n", msg)
	} else {
		fmt.Print(color.GreenString("✓ "))
		fmt.Println(msg)
	}
}

// printStartMessage prints a static loading message
func (s *WasmSpinner) printStartMessage() {
	if s.noColor {
		fmt.Printf("⏳ %s\n", s.message)
	} else {
		fmt.Print(color.CyanString("⏳ "))
		fmt.Println(s.message)
	}
}

// Override spinner functions to use WASM-specific implementation
func init() {
	// In WASM, spinners don't work well due to buffered output
	// We'll rely on JavaScript-side spinner implementation
	js.Global().Get("console").Call("log", "🔧 WASM spinner module initialized")
}

// NewSpinnerWasm creates a spinner optimized for WASM display
// This function is used when building for WASM to inject the WasmSpinner
func NewSpinnerWasm(noColor bool, outputFormat string) *Spinner {
	wasmSpinner := NewWasmSpinner("", noColor, outputFormat)
	
	return &Spinner{
		stop:         make(chan bool),
		frameRate:    150 * time.Millisecond,
		noColor:      noColor,
		outputFormat: outputFormat,
		style:        "wasm",
		wasmSpinner:  wasmSpinner, // Inject the WASM spinner
	}
}

// PrintResourceListingWasm creates an enhanced spinner for WASM
func PrintResourceListingWasm(resourceType string, noColor bool, outputFormat string) *Spinner {
	msg := "Listing " + resourceType + "s..."
	// Use WASM-enabled spinner
	spinner := NewSpinnerWasm(noColor, outputFormat)
	spinner.Start(msg)
	return spinner
}

// PrintResourceGettingWasm creates an enhanced spinner for WASM
func PrintResourceGettingWasm(resourceType, uid string, noColor bool, outputFormat string) *Spinner {
	uidFormatted := FormatUID(uid, noColor)
	msg := "Getting " + resourceType + " " + uidFormatted + " details..."
	spinner := NewSpinnerWasm(noColor, outputFormat)
	spinner.Start(msg)
	return spinner
}

// PrintLoggingInWasm creates an enhanced spinner for WASM login
func PrintLoggingInWasm(noColor bool, outputFormat string) *Spinner {
	msg := "Logging in to Megaport..."
	spinner := NewSpinnerWasm(noColor, outputFormat)
	spinner.Start(msg)
	return spinner
}

// WasmLoadingMessage shows a static loading indicator that works in buffered output
func WasmLoadingMessage(message string, noColor bool) {
	if noColor {
		fmt.Printf("⏳ %s\n", message)
	} else {
		// Create a prominent loading box
		border := color.New(color.FgHiCyan, color.Bold).Sprint("╔════════════════════════════════════════╗")
		bottom := color.New(color.FgHiCyan, color.Bold).Sprint("╚════════════════════════════════════════╝")
		icon := color.New(color.FgHiCyan, color.Bold).Sprint("⏳")
		text := color.New(color.FgHiWhite, color.Bold).Sprint(message)
		
		fmt.Println(border)
		fmt.Printf("║ %s  %-35s ║\n", icon, text)
		fmt.Println(bottom)
	}
	
	// Notify JavaScript
	if !js.Global().Get("wasmShowLoading").IsUndefined() {
		js.Global().Call("wasmShowLoading", message)
	}
}
