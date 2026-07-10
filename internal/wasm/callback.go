//go:build js && wasm

package wasm

import "syscall/js"

// InvokeCallback invokes cb with args, recovering from a panic so a throwing
// host JS callback cannot escape into the caller. This mirrors the recover
// pushOutputChunk (see output_handler.go) already applies around its own JS
// Invoke. It returns the recovered panic value, or nil if the invocation
// completed normally.
func InvokeCallback(cb js.Value, args ...interface{}) (panicVal interface{}) {
	defer func() {
		panicVal = recover()
	}()
	cb.Invoke(args...)
	return nil
}
