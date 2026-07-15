//go:build js && wasm

package wasm

import (
	"sync"
	"sync/atomic"
	"syscall/js"
)

// Live-output bridge. Mirrors the prompt bridge in prompts.go: JavaScript
// registers a single callback that WASM invokes with each chunk written to the
// captured output buffer, so the terminal renders output as it streams rather
// than only when the command completes.
//
// Completion contract (see GetCompletionOutput): when the handler delivers at
// least one chunk without throwing, the narrative is NOT repeated in the async
// command's final result, so the host must not render both. Structured document
// output (JSON/CSV/XML/table) is never streamed; it is returned once at
// completion. If the handler throws or delivers nothing, streaming is disabled
// for the rest of that command and completion falls back to the full captured
// output so nothing is lost (already-streamed chunks may then appear twice).
var (
	// outputCallback is the JavaScript function invoked with each output chunk.
	// Protected by outputCallbackMu.
	outputCallback   js.Value
	outputCallbackMu sync.RWMutex

	// pushingOutput guards against re-entrancy: if the JS callback triggers
	// another buffer write, that nested chunk is still captured in the buffer
	// but is not pushed again (which would recurse into JS). Ordering and
	// no-drop rely on a single writer goroutine (the running command); this
	// guards synchronous re-entrancy from the handler, not concurrent writers.
	// The atomic type is used to satisfy the race detector.
	pushingOutput atomic.Bool

	// outputStreamed records whether any chunk has been delivered to the
	// handler since the last ResetOutputBuffers. Error handling uses it to
	// avoid re-emitting an error that already streamed while still surfacing
	// one that did not (streamed chunks cannot be retracted).
	outputStreamed atomic.Bool

	// outputHandlerFailed latches when the handler throws mid-command. Further
	// chunks then stop being pushed and completion falls back to the full
	// captured buffer, so a handler that streams some chunks and then starts
	// throwing cannot silently drop the rest (including error text). Reset per
	// command by ResetOutputBuffers, so one transient throw does not disable
	// streaming for the remainder of the session.
	outputHandlerFailed atomic.Bool
)

// HasOutputHandler reports whether a usable output handler is registered.
func HasOutputHandler() bool {
	return hasOutputHandler()
}

// DidStreamOutput reports whether any chunk has streamed to the handler since
// the last ResetOutputBuffers.
func DidStreamOutput() bool {
	return outputStreamed.Load()
}

// RegisterOutputCallback stores the JavaScript function that receives streamed
// output chunks. A non-function value is rejected.
func RegisterOutputCallback(callback js.Value) {
	if callback.Type() != js.TypeFunction {
		js.Global().Get("console").Call("error", "Output callback must be a function")
		return
	}

	outputCallbackMu.Lock()
	outputCallback = callback
	outputCallbackMu.Unlock()
	js.Global().Get("console").Call("log", "✅ Output callback registered")
}

// UnregisterOutputCallback clears the registered handler so buffer writes fall
// back to capture-at-completion. Used to reset streaming state between runs so a
// callback does not leak past the lifetime it was registered for.
func UnregisterOutputCallback() {
	outputCallbackMu.Lock()
	outputCallback = js.Undefined()
	outputCallbackMu.Unlock()
}

// hasOutputHandler reports whether a usable output callback is registered.
func hasOutputHandler() bool {
	outputCallbackMu.RLock()
	cb := outputCallback
	outputCallbackMu.RUnlock()
	return !cb.IsUndefined() && !cb.IsNull() && cb.Type() == js.TypeFunction
}

// pushOutputChunk delivers a chunk to the registered output handler. It is a
// no-op when no handler is registered, when the chunk is empty, or when called
// re-entrantly from within the handler itself. A throw from the JS callback is
// recovered so a buggy or torn-down handler cannot abort the running command;
// output fires on every write, so an unrecovered panic here would take down an
// in-flight command mid-execution.
func pushOutputChunk(chunk string) {
	if chunk == "" {
		return
	}

	outputCallbackMu.RLock()
	cb := outputCallback
	outputCallbackMu.RUnlock()
	if cb.IsUndefined() || cb.IsNull() || cb.Type() != js.TypeFunction {
		return
	}

	if !pushingOutput.CompareAndSwap(false, true) {
		return
	}
	defer pushingOutput.Store(false)

	defer func() {
		if r := recover(); r != nil {
			// Latch failure: stop pushing to a broken handler and let completion
			// return the full captured buffer, so later chunks are not lost.
			outputHandlerFailed.Store(true)
			js.Global().Get("console").Call("error", "Output handler threw; disabling live streaming for this command, output will be delivered at completion")
		}
	}()

	cb.Invoke(chunk)
	// Mark streamed only after a clean delivery: if the handler throws (recovered
	// above), the chunk did not reach the terminal, so the error path must still
	// be free to surface a fallback rather than assume it already streamed.
	outputStreamed.Store(true)
}

// registerOutputHandler is the JS-facing wrapper for RegisterOutputCallback.
func registerOutputHandler(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		js.Global().Get("console").Call("error", "registerOutputHandler requires: callback (function)")
		return false
	}

	callback := args[0]
	if callback.Type() != js.TypeFunction {
		js.Global().Get("console").Call("error", "Argument must be a function")
		return false
	}

	RegisterOutputCallback(callback)
	return true
}

// InitOutputSystem registers the JavaScript entry point for live output
// streaming. Call once at startup, alongside InitPromptSystem.
func InitOutputSystem() {
	js.Global().Set("registerOutputHandler", js.FuncOf(registerOutputHandler))
	js.Global().Get("console").Call("log", "✅ WASM Output System initialized")
	js.Global().Get("console").Call("log", "  - registerOutputHandler(callback)")
}
