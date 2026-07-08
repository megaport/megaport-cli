//go:build !js || !wasm

package cmdbuilder

// fileInputUsage returns the usage string for a file-path input flag. On native
// builds the flag works normally, so the base text is used unchanged (inlineFlag
// only matters in the browser annotation).
func fileInputUsage(base, _ string) string { return base }
