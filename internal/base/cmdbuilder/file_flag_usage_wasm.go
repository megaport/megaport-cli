//go:build js && wasm

package cmdbuilder

// fileInputUsage annotates a file-path input flag's usage in the browser build,
// where reading from an OS filesystem is unsupported (see the platform helpers
// in internal/utils/input_file_wasm.go). The flag stays registered so its input
// path still surfaces the clear runtime error; inlineFlag names the inline
// alternative to point the user at.
func fileInputUsage(base, inlineFlag string) string {
	return base + " (not available in the browser; use " + inlineFlag + " instead)"
}
