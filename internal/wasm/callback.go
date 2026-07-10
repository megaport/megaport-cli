//go:build js && wasm

package wasm

import "syscall/js"

// InvokeCallback invokes cb with args, recovering from a panic so a throwing
// host JS callback cannot escape into the caller. This mirrors the recover
// that pushOutputChunk (see output_handler.go) already applies around its
// own JS Invoke. It returns the recovered panic value, or nil if the
// invocation completed normally.
//
// Deliberately do not format the returned value into a log or error message:
// a failed Invoke panics with a js.Error wrapping the thrown JS value, and
// js.Error.Error() reads a "message" property back from the JS runtime. If
// that property is a throwing accessor, js.Value.Get itself has no recover
// boundary around it (unlike Invoke), so formatting the panic value can
// crash the whole runtime one call frame past where this recover already
// returned. A fixed, friendly message avoids ever touching the value again.
func InvokeCallback(cb js.Value, args ...interface{}) (panicVal interface{}) {
	defer func() {
		panicVal = recover()
	}()
	cb.Invoke(args...)
	return
}
