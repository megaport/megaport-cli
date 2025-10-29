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

// executeMegaportCommand runs CLI commands from JavaScript (LEGACY SYNC VERSION)
// This is kept for backwards compatibility but may not work with async operations
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

// executeMegaportCommandAsync runs CLI commands asynchronously with a callback
// This is the CORRECT way to handle commands that involve async operations (like auth)
func executeMegaportCommandAsync(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		js.Global().Get("console").Call("error", "executeMegaportCommandAsync requires: command (string) and callback (function)")
		return nil
	}

	// Get command string and callback from JavaScript
	cmdString := args[0].String()
	callback := args[1]

	if callback.Type() != js.TypeFunction {
		js.Global().Get("console").Call("error", "Second argument must be a callback function")
		return nil
	}

	js.Global().Get("console").Call("log", fmt.Sprintf("🚀 Starting async command: %s", cmdString))

	// Run the command in a goroutine to allow async operations
	go func() {
		defer func() {
			if r := recover(); r != nil {
				js.Global().Get("console").Call("error", "Panic in async command:", r)
				callback.Invoke(map[string]interface{}{
					"error": fmt.Sprintf("Command panicked: %v", r),
				})
			}
		}()

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
		js.Global().Get("console").Call("time", "Async command execution")
		megaport.ExecuteWithArgs(originalArgs)
		js.Global().Get("console").Call("timeEnd", "Async command execution")

		// Get output
		output := wasm.GetCapturedOutput()

		// Log the output
		if len(output) > 1000 {
			js.Global().Get("console").Call("log", fmt.Sprintf("📋 Async output ready: [first 1000 bytes of %d]:\n%s...",
				len(output), output[:1000]))
		} else {
			js.Global().Get("console").Call("log", fmt.Sprintf("📋 Async output ready [%d bytes]:\n%s",
				len(output), output))
		}

		// Call the callback with the result
		callback.Invoke(map[string]interface{}{
			"output": output,
		})
	}()

	// Return immediately - the callback will be called when done
	js.Global().Get("console").Call("log", "✅ Async command started, returning immediately")
	return nil
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

	// Export both sync (legacy) and async (preferred) versions
	js.Global().Set("executeMegaportCommand", js.FuncOf(executeMegaportCommand))
	js.Global().Set("executeMegaportCommandAsync", js.FuncOf(executeMegaportCommandAsync))
	
	js.Global().Get("console").Call("log", "✅ Registered executeMegaportCommand (sync) and executeMegaportCommandAsync (async)")

	// Prevent Go WASM from exiting after main finishes
	<-make(chan bool)
}
