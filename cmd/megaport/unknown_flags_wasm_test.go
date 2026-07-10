//go:build js && wasm

package megaport

import (
	"testing"

	"github.com/megaport/megaport-cli/internal/wasm"
	"github.com/stretchr/testify/assert"
)

// TestExecuteWithArgs_RejectsUnknownFlags is the ESD-1634 regression test:
// the WASM entrypoint must report an unknown flag as an error, matching the
// native build, instead of silently dropping it and either proceeding or
// failing later with a misleading "required flag not set" message. Cases are
// help-only or otherwise non-mutating by design so no network call happens
// and no real order is placed.
func TestExecuteWithArgs_RejectsUnknownFlags(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{"unknown flag on a leaf command", []string{"megaport-cli", "locations", "list", "--totally-bogus-flag"}},
		{"unknown flag before the subcommand", []string{"megaport-cli", "--totally-bogus-flag", "locations", "list"}},
		{"unknown flag between subcommand levels", []string{"megaport-cli", "locations", "--totally-bogus-flag", "list"}},
		{"ticket repro: typo'd boolean flag on a buy command", []string{"megaport-cli", "ports", "buy", "--marketplace-visability", "false"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wasm.ResetOutputBuffers()
			ExecuteWithArgs(tc.args)
			out := wasm.GetCapturedOutput()
			assert.Contains(t, out, "unknown flag", "%v should report the unknown flag", tc.args)
			assert.NotContains(t, out, "required flag(s)", "%v should not fall through to a required-flag error", tc.args)
		})
	}
}

// TestExecuteWithArgs_ValidTraversalStillResolves guards the other half of
// ESD-1634: removing the dead UnknownFlags allowlist must not break
// subcommand resolution when a persistent flag precedes the subcommand, or
// when help is requested at a nested command. Cases are help-only so no
// network call is made.
func TestExecuteWithArgs_ValidTraversalStillResolves(t *testing.T) {
	cases := []struct {
		name string
		args []string
		// usage anchors on the "Usage:" block so a parent's "Example usage"
		// prose (e.g. the root help's Examples list) can't satisfy the match
		// if traversal actually failed and root help was shown instead.
		usage string
	}{
		{"persistent flag before subcommand", []string{"megaport-cli", "--no-color", "locations", "list", "--help"}, "Usage:\n  megaport-cli locations list"},
		{"persistent flag between subcommand levels", []string{"megaport-cli", "locations", "--no-color", "list", "--help"}, "Usage:\n  megaport-cli locations list"},
		{"deeply nested subcommand", []string{"megaport-cli", "ports", "buy", "--help"}, "Usage:\n  megaport-cli ports buy"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wasm.ResetOutputBuffers()
			ExecuteWithArgs(tc.args)
			out := wasm.GetCapturedOutput()
			assert.NotContains(t, out, "unknown flag", "%v should not be treated as an unknown flag", tc.args)
			assert.NotContains(t, out, "unknown command", "%v should be a registered command", tc.args)
			assert.Contains(t, out, tc.usage, "%v help should show its usage path", tc.args)
		})
	}
}
