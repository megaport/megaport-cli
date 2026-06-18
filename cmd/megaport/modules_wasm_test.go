//go:build js && wasm

package megaport

import (
	"testing"

	"github.com/megaport/megaport-cli/internal/wasm"
	"github.com/stretchr/testify/assert"
)

// TestReadOnlyModulesRegistered verifies the read-only status, topology, and
// product modules are wired into the WASM command tree (ESD-1283). Each is
// invoked with --help so no network call is made; an unregistered command would
// instead surface an "unknown command" error. Cases are help-only by design so
// they share the global rootCmd without leaking flag state between subtests.
func TestReadOnlyModulesRegistered(t *testing.T) {
	cases := []struct {
		name string
		args []string
		// usage is the command path cobra echoes in the "Usage:" block; it
		// tracks the command structure rather than drift-prone help prose.
		usage string
	}{
		{"status", []string{"megaport-cli", "status", "--help"}, "megaport-cli status"},
		{"topology", []string{"megaport-cli", "topology", "--help"}, "megaport-cli topology"},
		{"product", []string{"megaport-cli", "product", "--help"}, "megaport-cli product"},
		{"product list", []string{"megaport-cli", "product", "list", "--help"}, "megaport-cli product list"},
		{"product get-type", []string{"megaport-cli", "product", "get-type", "--help"}, "megaport-cli product get-type"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wasm.ResetOutputBuffers()
			ExecuteWithArgs(tc.args)
			out := wasm.GetCapturedOutput()
			assert.NotContains(t, out, "unknown command", "%v should be a registered command", tc.args)
			assert.Contains(t, out, tc.usage, "%v help should show its usage path", tc.args)
		})
	}
}
