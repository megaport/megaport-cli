//go:build js && wasm
// +build js,wasm

package output

// SetNoPager stores the no-pager flag in the shared OutputConfig. The stored
// value has no effect in the WASM build because RunWithPager always calls fn
// directly — there is no terminal and no pager process to spawn.
func SetNoPager(v bool) {
	cfg := GetOutputConfig()
	cfg.NoPager = v
	ApplyOutputConfig(cfg)
}

// GetNoPager returns the stored no-pager setting. It may be true if the
// --no-pager flag was passed, but RunWithPager ignores it in WASM builds.
func GetNoPager() bool { return GetOutputConfig().NoPager }

// RunWithPager in the WASM build simply calls fn directly. There is no
// terminal to detect and no process to spawn.
func RunWithPager(fn func() error) error { return fn() }
