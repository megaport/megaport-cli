//go:build js && wasm
// +build js,wasm

package main

import (
	"embed"
	"fmt"
	"syscall/js"

	"github.com/megaport/megaport-cli/cmd/megaport"
	"github.com/megaport/megaport-cli/internal/base/cmdbuilder"
	"github.com/megaport/megaport-cli/internal/wasm"
)

//go:embed docs/*.md
var embeddedDocs embed.FS

// executeMegaportCommand runs CLI commands from JavaScript
func executeMegaportCommand(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return map[string]interface{}{
			"error": "No command provided",
		}
	}

	// Get command string from JavaScript
	cmdString := args[0].String()

	// Reset all output buffers
	wasm.ResetOutputBuffers()

	// Split the command string into arguments
	cmdArgs := wasm.SplitArgs(cmdString)

	// Create a new slice with the program name
	originalArgs := append([]string{"megaport-cli"}, cmdArgs...)

	// Use our new tracing function
	wasm.TraceCommand(cmdString, originalArgs)

	// Ensure Cobra gets all our commands
	megaport.EnsureRootCommandOutput(wasm.WasmOutputBuffer)

	// Execute with console timing
	js.Global().Get("console").Call("time", "Command execution")
	megaport.ExecuteWithArgs(originalArgs)
	js.Global().Get("console").Call("timeEnd", "Command execution")

	// Get output and return
	output := wasm.GetCapturedOutput()

	// Log the final output that's being returned to the terminal
	if len(output) > 1000 {
		js.Global().Get("console").Call("log", fmt.Sprintf("📋 Returning output: [first 1000 bytes of %d]:\n%s...",
			len(output), output[:1000]))
	} else {
		js.Global().Get("console").Call("log", fmt.Sprintf("📋 Returning output [%d bytes]:\n%s",
			len(output), output))
	}

	return map[string]interface{}{
		"output": output,
	}
}

func main() {
	// Register the embedded documentation with the cmdbuilder package
	cmdbuilder.RegisterEmbeddedDocs(embeddedDocs)

	// Enable debug mode by default for WASM
	wasm.EnableDebugMode()

	// Log WASM initialization
	js.Global().Get("console").Call("log", "🚀 Megaport CLI WASM initialized with enhanced logging")

	// Register JavaScript functions
	wasm.RegisterJSFunctions()

	// Setup output redirection
	wasm.SetupIO()

	// CRITICAL: Export the main execute function with the exact name expected by the web UI
	js.Global().Set("executeMegaportCommand", js.FuncOf(executeMegaportCommand))

	// Prevent Go WASM from exiting after main finishes
	<-make(chan bool)
}
