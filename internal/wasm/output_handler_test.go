//go:build js && wasm

package wasm

import (
	"syscall/js"
	"testing"

	"github.com/stretchr/testify/assert"
)

// resetOutputStreaming clears the streaming state and buffers between tests.
func resetOutputStreaming() {
	outputCallbackMu.Lock()
	outputCallback = js.Undefined()
	outputCallbackMu.Unlock()
	pushingOutput.Store(false)
	outputStreamed.Store(false)
	WasmOutputBuffer.Reset()
	js.Global().Delete("wasmJSONOutput")
	js.Global().Delete("wasmCSVOutput")
	js.Global().Delete("wasmXMLOutput")
	js.Global().Delete("wasmTableOutput")
}

// TestOutputHandlerReceivesChunksDuringExecution verifies that each write to
// the captured buffer is pushed to a registered handler as it happens, not
// batched until completion.
func TestOutputHandlerReceivesChunksDuringExecution(t *testing.T) {
	resetOutputStreaming()
	defer resetOutputStreaming()

	var received []string
	fn := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			received = append(received, args[0].String())
		}
		return nil
	})
	defer fn.Release()
	RegisterOutputCallback(fn.Value)

	assert.True(t, hasOutputHandler(), "handler should be registered")

	_, _ = WasmOutputBuffer.Write([]byte("step 1\n"))
	assert.Equal(t, []string{"step 1\n"}, received, "chunk should arrive on the first write, not batched")

	_, _ = WasmOutputBuffer.Write([]byte("step 2\n"))
	assert.Equal(t, []string{"step 1\n", "step 2\n"}, received, "chunks stream in write order")
}

// TestOutputHandlerNotRegisteredIsNoop verifies buffer writes work normally
// (and don't panic) when no handler is registered.
func TestOutputHandlerNotRegisteredIsNoop(t *testing.T) {
	resetOutputStreaming()
	defer resetOutputStreaming()

	assert.False(t, hasOutputHandler(), "no handler should be registered")
	assert.NotPanics(t, func() {
		_, _ = WasmOutputBuffer.Write([]byte("no handler here\n"))
	})
	assert.Equal(t, "no handler here\n", WasmOutputBuffer.String(), "write is still buffered")
}

// TestPushOutputChunkEmptyChunkSkipped verifies empty chunks are not delivered.
func TestPushOutputChunkEmptyChunkSkipped(t *testing.T) {
	resetOutputStreaming()
	defer resetOutputStreaming()

	var received []string
	fn := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			received = append(received, args[0].String())
		}
		return nil
	})
	defer fn.Release()
	RegisterOutputCallback(fn.Value)

	pushOutputChunk("")
	assert.Empty(t, received, "empty chunk should not be pushed")
}

// TestRegisterOutputCallbackRejectsNonFunction verifies a non-function value is
// not stored as a handler.
func TestRegisterOutputCallbackRejectsNonFunction(t *testing.T) {
	resetOutputStreaming()
	defer resetOutputStreaming()

	RegisterOutputCallback(js.ValueOf("not a function"))
	assert.False(t, hasOutputHandler(), "a string must not be accepted as a handler")
}

// TestOutputHandlerReentrancyGuard verifies that a write triggered from inside
// the handler is still captured in the buffer but is not pushed again (which
// would recurse into JS).
func TestOutputHandlerReentrancyGuard(t *testing.T) {
	resetOutputStreaming()
	defer resetOutputStreaming()

	var received []string
	wroteInner := false
	fn := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			received = append(received, args[0].String())
		}
		// Simulate a handler that writes back into the buffer exactly once.
		if !wroteInner {
			wroteInner = true
			_, _ = WasmOutputBuffer.Write([]byte("inner"))
		}
		return nil
	})
	defer fn.Release()
	RegisterOutputCallback(fn.Value)

	_, _ = WasmOutputBuffer.Write([]byte("outer"))

	assert.Equal(t, []string{"outer"}, received, "re-entrant write must not be pushed to the handler")
	assert.Equal(t, "outerinner", WasmOutputBuffer.String(), "re-entrant write is still captured in the buffer")
}

// TestOutputHandlerThrowingCallbackRecovered verifies that a handler which
// throws does not abort the write or the running command: the panic from
// cb.Invoke is recovered and the buffer write still succeeds.
func TestOutputHandlerThrowingCallbackRecovered(t *testing.T) {
	resetOutputStreaming()
	defer resetOutputStreaming()

	throwing := js.Global().Get("Function").New("throw new Error('handler boom')")
	RegisterOutputCallback(throwing)
	assert.True(t, hasOutputHandler(), "a throwing handler still registers")

	var n int
	var err error
	assert.NotPanics(t, func() {
		n, err = WasmOutputBuffer.Write([]byte("still written\n"))
	}, "a throwing handler must not panic out of the write")

	assert.NoError(t, err)
	assert.Equal(t, len("still written\n"), n)
	assert.Equal(t, "still written\n", WasmOutputBuffer.String(),
		"write is still captured even when the handler throws")

	assert.False(t, pushingOutput.Load(), "re-entrancy guard is reset after a recovered throw")
	assert.False(t, DidStreamOutput(),
		"a chunk the handler threw on did not reach the terminal, so it must not count as streamed")
}

// TestDidStreamOutputTracksDelivery verifies the streamed-since-reset marker
// used by the error path to avoid double-rendering: it flips true once a chunk
// is delivered and is cleared by ResetOutputBuffers.
func TestDidStreamOutputTracksDelivery(t *testing.T) {
	resetOutputStreaming()
	defer resetOutputStreaming()

	assert.False(t, DidStreamOutput(), "nothing streamed before any write")

	fn := js.FuncOf(func(this js.Value, args []js.Value) interface{} { return nil })
	defer fn.Release()
	RegisterOutputCallback(fn.Value)
	assert.True(t, HasOutputHandler(), "handler should be registered")

	_, _ = WasmOutputBuffer.Write([]byte("something\n"))
	assert.True(t, DidStreamOutput(), "a delivered chunk marks output as streamed")

	ResetOutputBuffers()
	assert.False(t, DidStreamOutput(), "reset clears the streamed marker")
}

// TestGetCompletionOutputNoDoubleRender verifies the completion contract: with a
// handler registered, narrative is not returned again (it was streamed), while
// structured document output still is.
func TestGetCompletionOutputNoDoubleRender(t *testing.T) {
	t.Run("handler registered, narrative only, returns empty", func(t *testing.T) {
		resetOutputStreaming()
		defer resetOutputStreaming()

		fn := js.FuncOf(func(this js.Value, args []js.Value) interface{} { return nil })
		defer fn.Release()
		RegisterOutputCallback(fn.Value)

		_, _ = WasmOutputBuffer.Write([]byte("streamed narrative\n"))
		assert.Equal(t, "", GetCompletionOutput(),
			"streamed narrative must not be returned again at completion")
	})

	t.Run("handler registered, structured output still returned", func(t *testing.T) {
		resetOutputStreaming()
		defer resetOutputStreaming()

		fn := js.FuncOf(func(this js.Value, args []js.Value) interface{} { return nil })
		defer fn.Release()
		RegisterOutputCallback(fn.Value)

		js.Global().Set("wasmTableOutput", "PORT TABLE")
		_, _ = WasmOutputBuffer.Write([]byte("streamed narrative\n"))
		assert.Equal(t, "PORT TABLE", GetCompletionOutput(),
			"structured document output is delivered once at completion")
	})

	t.Run("no handler falls back to full captured output", func(t *testing.T) {
		resetOutputStreaming()
		defer resetOutputStreaming()

		_, _ = WasmOutputBuffer.Write([]byte("full output\n"))
		assert.Equal(t, "full output\n", GetCompletionOutput(),
			"without a handler the host still receives the full output")
	})
}
