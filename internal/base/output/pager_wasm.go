//go:build js && wasm
// +build js,wasm

package output

// SetNoPager is a no-op in the WASM build. The browser environment has no
// terminal and cannot spawn pager processes.
func SetNoPager(_ bool) {}

// GetNoPager always returns false in the WASM build; paging is never active.
func GetNoPager() bool { return false }

// RunWithPager in the WASM build simply calls fn directly. There is no
// terminal to detect and no process to spawn.
func RunWithPager(fn func() error) error { return fn() }
