//go:build js && wasm

package output

import (
	"syscall/js"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestPrintErrorJSON_SanitizesMessage verifies that a C1 control character
// carried in the error message (e.g. echoed from an API error response body)
// is stripped from wasmJSONOutput. Go's JSON encoder escapes C0 controls
// (ESC included) as textual "\u00XX" sequences, but passes C1 (0x80-0x9F,
// some terminals' 8-bit CSI introducer) through as a raw byte, so it reaches
// xterm unescaped unless this sanitize step removes it.
func TestPrintErrorJSON_SanitizesMessage(t *testing.T) {
	js.Global().Delete("wasmJSONOutput")

	// hex(0x9b) is the 8-bit CSI introducer some terminals honor directly.
	c1 := string(rune(0x9b))
	PrintErrorJSON(500, "upstream said: acme-corp"+c1+"2Kspoofed")

	output := js.Global().Get("wasmJSONOutput").String()
	assert.NotContains(t, output, c1, "injected C1 control byte must not reach the host")
	assert.Contains(t, output, "acme-corp")
	assert.Contains(t, output, "spoofed")
	assert.Contains(t, output, `"code": 500`)
}
