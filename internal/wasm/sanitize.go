//go:build js && wasm

package wasm

import "unicode"

// SanitizeTerminalOutput strips control sequences from WASM data output
// (tables, JSON/CSV/XML, narrative messages, errors) before it reaches the
// host terminal. The host writes this text to xterm via terminal.write()
// without escaping, so a crafted resource name or API response field
// carrying ESC/CSI/OSC bytes could otherwise move the cursor or erase
// unrelated lines and repaint a fake prompt over real content. Some of
// these fields are third-party controlled (partner/marketplace listings
// carry other tenants' names; error paths echo API response bodies).
//
// The CLI's own SGR color sequences (ESC '[' <digits/';'> 'm', emitted by
// fatih/color for table and status styling) are allowlisted so intended
// styling survives; every other C0 control, DEL, C1 control, and non-SGR
// CSI/OSC sequence is removed. \n, \r, and \t pass through unchanged since
// they are structural formatting in a multi-line document.
//
// This does not distinguish a CLI-emitted SGR sequence from attacker text
// that merely happens to match SGR syntax once everything is one string.
// That is an accepted gap: SGR only changes subsequent text color/style, it
// cannot move the cursor or erase the screen, so the residual risk is
// cosmetic, not the cursor-hijack/erase attack this sanitizer closes.
func SanitizeTerminalOutput(s string) string {
	return sanitizeControlSequences(s, true)
}

// SanitizeTerminalText strips all C0/C1 controls, DEL, and ESC from a single
// field value bound for the live prompt channel (see the WASM prompt bridge
// in internal/utils/prompts_wasm.go). Prompt messages carry no color, so
// nothing needs to be allowlisted, and callers insert their own structural
// line breaks after sanitizing each field, so \n/\r/\t are stripped like any
// other control byte.
func SanitizeTerminalText(s string) string {
	return sanitizeControlSequences(s, false)
}

// sanitizeControlSequences removes terminal control bytes from s. When
// allowSGR is true, \n, \r, and \t are preserved as structural bytes and an
// ESC that introduces a well-formed CSI SGR sequence (ESC '[' <digits/';'>*
// 'm') is copied through verbatim. Any other ESC is dropped by itself: this
// leaves the rest of a CSI or OSC sequence (e.g. "[2K" or "]0;evil") as inert
// printable text, since the terminal only interprets it as a command when
// the introducing ESC (or its C1 equivalent) is present. All other Unicode
// control runes, including DEL and the C1 range (some terminals treat
// 0x80-0x9f as 8-bit CSI/OSC introducers, e.g. 0x9b as CSI), are dropped
// outright.
func sanitizeControlSequences(s string, allowSGR bool) string {
	runes := []rune(s)
	out := make([]rune, 0, len(runes))
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		switch {
		case r == 0x1b:
			if allowSGR {
				if end, ok := sgrSequenceEnd(runes, i); ok {
					out = append(out, runes[i:end+1]...)
					i = end
					continue
				}
			}
			// Not an allowlisted sequence: drop just the ESC byte so any
			// following text is re-evaluated as plain characters, not
			// re-armed as part of a control sequence.
		case allowSGR && (r == '\n' || r == '\r' || r == '\t'):
			out = append(out, r)
		case unicode.IsControl(r):
			// Drop C0/C1 controls and DEL.
		default:
			out = append(out, r)
		}
	}
	return string(out)
}

// sgrSequenceEnd reports whether the ESC at runes[start] begins a CSI SGR
// sequence and, if so, returns the index of the final 'm'. Only SGR is
// allowlisted: it is the only CSI class the CLI's own color output emits.
// Cursor movement, erase, and every other CSI/OSC final byte fall through
// and are stripped by the caller.
func sgrSequenceEnd(runes []rune, start int) (int, bool) {
	if start+1 >= len(runes) || runes[start+1] != '[' {
		return 0, false
	}
	for i := start + 2; i < len(runes); i++ {
		switch r := runes[i]; {
		case r >= '0' && r <= '9', r == ';':
			continue
		case r == 'm':
			return i, true
		default:
			return 0, false
		}
	}
	return 0, false
}
