package output

import (
	"encoding/json"
	"regexp"
	"strings"
)

var ansiRegexp = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]|\x1b\[K`)

// ExtractJSON strips ANSI escape sequences from captured output and extracts
// the first complete JSON value (array or object) using json.Decoder.
// This is useful in tests where spinner output may contaminate stdout.
func ExtractJSON(s string) string {
	clean := ansiRegexp.ReplaceAllString(s, "")
	// Scan forward through all '[' and '{' positions, attempting to decode
	// a complete JSON value at each one. This handles cases where non-JSON
	// brackets appear before the actual JSON (e.g., "deploy [a b] [{"uid":...}]").
	remaining := clean
	for {
		start := strings.IndexAny(remaining, "[{")
		if start == -1 {
			return clean
		}
		dec := json.NewDecoder(strings.NewReader(remaining[start:]))
		var raw json.RawMessage
		if err := dec.Decode(&raw); err == nil {
			return string(raw)
		}
		// Move past this bracket and try the next one
		remaining = remaining[start+1:]
	}
}
