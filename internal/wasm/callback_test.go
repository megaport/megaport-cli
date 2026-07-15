//go:build js && wasm

package wasm

import (
	"syscall/js"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestInvokeCallbackRecoversPanic verifies a throwing JS callback is recovered
// and its panic value returned, rather than escaping to the caller.
func TestInvokeCallbackRecoversPanic(t *testing.T) {
	throwing := js.Global().Get("Function").New("throw new Error('callback boom')")

	var r interface{}
	assert.NotPanics(t, func() {
		r = InvokeCallback(throwing)
	})
	assert.NotNil(t, r, "expected the recovered panic value to be returned")
}

// TestInvokeCallbackCleanInvocation verifies a well-behaved callback runs
// normally and InvokeCallback returns nil (no panic to report).
func TestInvokeCallbackCleanInvocation(t *testing.T) {
	var received string
	fn := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			received = args[0].String()
		}
		return nil
	})
	defer fn.Release()

	r := InvokeCallback(fn.Value, "hello")
	assert.Nil(t, r)
	assert.Equal(t, "hello", received)
}
