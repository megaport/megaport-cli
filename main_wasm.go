//go:build js && wasm
// +build js,wasm

package main

import (
	"embed"
	"fmt"
	"sync"
	"syscall/js"
	"time"

	"github.com/megaport/megaport-cli/cmd/megaport"
	"github.com/megaport/megaport-cli/internal/base/cmdbuilder"
	"github.com/megaport/megaport-cli/internal/base/output"
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

	megaport.ExecuteWithArgs(originalArgs)

	return map[string]interface{}{
		"output": wasm.GetCapturedOutput(),
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

	// asyncCommandTimeout is the maximum time an async command may run before
	// the callback is invoked with a timeout error. The inner goroutine may still
	// be running after the timeout (Go goroutines cannot be forcibly cancelled),
	// but the JS caller will not be left waiting indefinitely.
	const asyncCommandTimeout = 10 * time.Minute

	// once ensures the callback is invoked exactly once even if both the timeout
	// and the normal completion path race.
	var once sync.Once

	go func() {
		defer func() {
			if r := recover(); r != nil {
				once.Do(func() {
					callback.Invoke(map[string]interface{}{
						"error": fmt.Sprintf("Command panicked: %v", r),
					})
				})
			}
		}()

		// Reset all output buffers
		wasm.ResetOutputBuffers()

		// Split the command string into arguments
		cmdArgs := wasm.SplitArgs(cmdString)

		// Create a new slice with the program name
		originalArgs := append([]string{"megaport-cli"}, cmdArgs...)

		// Use our tracing function (no-op when debug mode is off)
		wasm.TraceCommand(cmdString, originalArgs)

		// Ensure Cobra gets all our commands
		megaport.EnsureRootCommandOutput(wasm.WasmOutputBuffer)

		megaport.ExecuteWithArgs(originalArgs)

		result := wasm.GetCapturedOutput()
		once.Do(func() {
			callback.Invoke(map[string]interface{}{
				"output": result,
			})
		})
	}()

	// Fire a timeout so the callback is always invoked within asyncCommandTimeout.
	time.AfterFunc(asyncCommandTimeout, func() {
		once.Do(func() {
			callback.Invoke(map[string]interface{}{
				"error": "command timed out",
			})
		})
	})

	return nil
}

func main() {
	// Wire output state reset into the wasm package. Done here (rather than in
	// internal/wasm) to break the import cycle between wasm and output packages.
	wasm.RegisterOutputStateReset(func() {
		output.SetOutputFields(nil)
		output.SetOutputQuery("")
		output.SetOutputFormat("table")
		output.SetVerbosity("normal")
	})

	// Register the embedded documentation with the cmdbuilder package
	cmdbuilder.RegisterEmbeddedDocs(embeddedDocs)

	// Enable debug mode only when explicitly requested by the host page.
	// Set window.wasmDebugMode = true before the WASM module loads to opt in.
	if js.Global().Get("wasmDebugMode").Truthy() {
		wasm.EnableDebugMode()
	}

	// Register JavaScript functions
	wasm.RegisterJSFunctions()

	// Setup output redirection
	wasm.SetupIO()

	// Initialize the prompt system for interactive mode
	wasm.InitPromptSystem()

	// Export both sync (legacy) and async (preferred) versions
	js.Global().Set("executeMegaportCommand", js.FuncOf(executeMegaportCommand))
	js.Global().Set("executeMegaportCommandAsync", js.FuncOf(executeMegaportCommandAsync))

	// Prevent Go WASM from exiting after main finishes
	<-make(chan bool)
}
