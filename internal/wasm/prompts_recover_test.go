//go:build js && wasm

package wasm

import (
	"syscall/js"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestPromptForInputThrowingCallbackRecovered verifies a throwing prompt
// callback is recovered: PromptForInput returns an error instead of
// panicking, and the pendingPrompts entry it registered before invoking the
// callback is cleaned up rather than leaked. Cleanup normally only happens in
// the response/timeout select in PromptForInput, which a panicking Invoke
// never reaches, so this exercises the dedicated cleanup path.
func TestPromptForInputThrowingCallbackRecovered(t *testing.T) {
	pendingMutex.Lock()
	pendingPrompts = make(map[string]*PromptRequest)
	pendingMutex.Unlock()

	throwing := js.Global().Get("Function").New("throw new Error('prompt handler boom')")
	promptCallback = throwing
	defer func() { promptCallback = js.Undefined() }()

	var result string
	var err error
	assert.NotPanics(t, func() {
		result, err = PromptForInput("Enter name:", "text", "")
	})

	assert.Empty(t, result)
	assert.Error(t, err)

	pendingMutex.Lock()
	remaining := len(pendingPrompts)
	pendingMutex.Unlock()
	assert.Equal(t, 0, remaining, "pendingPrompts entry must not leak when the callback throws")
}
