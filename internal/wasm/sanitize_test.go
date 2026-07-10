//go:build js && wasm

package wasm

import (
	"syscall/js"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeTerminalOutputStripsCursorMoveAndErase(t *testing.T) {
	// A partner/marketplace listing name carrying a cursor-home + erase-line
	// sequence, the exact shape that could repaint a fake "delete? [y/N]"
	// line over real table content.
	evil := "acme-corp\x1b[H\x1b[2Kspoofed"
	got := SanitizeTerminalOutput(evil)

	assert.NotContains(t, got, "\x1b[H")
	assert.NotContains(t, got, "\x1b[2K")
	assert.Contains(t, got, "acme-corp")
	assert.Contains(t, got, "spoofed")
}

func TestSanitizeTerminalOutputStripsOSC(t *testing.T) {
	// OSC 0 (set window title) terminated by BEL.
	evil := "row\x1b]0;pwned\x07value"
	got := SanitizeTerminalOutput(evil)

	assert.NotContains(t, got, "\x1b")
	assert.NotContains(t, got, "\x07")
	assert.Contains(t, got, "row")
	assert.Contains(t, got, "value")
}

func TestSanitizeTerminalOutputStripsC1Control(t *testing.T) {
	// U+009B is the 8-bit CSI introducer some terminals honor directly.
	evil := "port2Kspoofed"
	got := SanitizeTerminalOutput(evil)

	assert.NotContains(t, got, "")
	assert.Contains(t, got, "port")
	assert.Contains(t, got, "spoofed")
}

func TestSanitizeTerminalOutputPreservesSGRColor(t *testing.T) {
	// fatih/color's SprintFunc output: FgRed + Bold, then reset.
	colored := "\x1b[31;1mACTIVE\x1b[0m"
	got := SanitizeTerminalOutput(colored)

	assert.Equal(t, colored, got, "legitimate SGR color sequences must pass through unchanged")
}

func TestSanitizeTerminalOutputPreservesStructuralWhitespace(t *testing.T) {
	got := SanitizeTerminalOutput("line one\nline two\r\ncol1\tcol2")
	assert.Equal(t, "line one\nline two\r\ncol1\tcol2", got)
}

func TestSanitizeTerminalOutputMixedColorAndInjection(t *testing.T) {
	// A colorized table cell whose value carries an injected cursor-move: the
	// SGR wrapper must survive, the injected CSI must not.
	cell := "\x1b[31mport\x1b[2K\x1b[Hspoofed\x1b[0m"
	got := SanitizeTerminalOutput(cell)

	assert.Contains(t, got, "\x1b[31m")
	assert.Contains(t, got, "\x1b[0m")
	assert.NotContains(t, got, "\x1b[2K")
	assert.NotContains(t, got, "\x1b[H")
	assert.Contains(t, got, "port")
	assert.Contains(t, got, "spoofed")
}

func TestSanitizeTerminalTextStripsEverythingIncludingNewlines(t *testing.T) {
	// The prompt-channel variant has no styling to preserve and callers
	// insert their own line breaks after sanitizing, so newlines in a field
	// value are stripped like any other control byte.
	evil := "port\x1b[2K\x1b[Hspoofed\nline2"
	got := SanitizeTerminalText(evil)

	assert.NotContains(t, got, "\x1b")
	assert.NotContains(t, got, "\n")
	assert.Contains(t, got, "port")
	assert.Contains(t, got, "spoofed")
	assert.Contains(t, got, "line2")
}

func TestSanitizeTerminalTextDoesNotAllowlistSGR(t *testing.T) {
	colored := "\x1b[31mred\x1b[0m"
	got := SanitizeTerminalText(colored)

	assert.NotContains(t, got, "\x1b")
	assert.Contains(t, got, "red")
}

// TestDirectOutputBufferWriteSanitizesStreamedAndBufferedOutput drives an
// injected CSI cursor-move sequence, alongside a legitimate SGR-colored
// value, through the actual WASM output path (DirectOutputBuffer.Write ->
// pushOutputChunk) and asserts the pushed chunk and the buffered content read
// back via String() are both clean while the color survives.
func TestDirectOutputBufferWriteSanitizesStreamedAndBufferedOutput(t *testing.T) {
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

	row := "\x1b[31mport-1\x1b[2K\x1b[Hspoofed\x1b[0m\n"
	_, err := WasmOutputBuffer.Write([]byte(row))
	assert.NoError(t, err)

	assert.Len(t, received, 1, "the write should stream exactly one chunk")
	pushed := received[0]
	assert.NotContains(t, pushed, "\x1b[2K", "pushed chunk must not contain the injected erase sequence")
	assert.NotContains(t, pushed, "\x1b[H", "pushed chunk must not contain the injected cursor-home sequence")
	assert.Contains(t, pushed, "\x1b[31m", "pushed chunk must retain the CLI's own SGR color")
	assert.Contains(t, pushed, "port-1")
	assert.Contains(t, pushed, "spoofed")

	buffered := WasmOutputBuffer.String()
	assert.NotContains(t, buffered, "\x1b[2K")
	assert.NotContains(t, buffered, "\x1b[H")
	assert.Contains(t, buffered, "\x1b[31m")
}
