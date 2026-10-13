package megaport

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// No build tag, so the WASM test job runs this against the browser command tree too.
func TestResourceCommandsAcceptSingularAndPlural(t *testing.T) {
	cases := []struct {
		name  string
		alias string
	}{
		{"ports", "port"},
		{"vxc", "vxcs"},
		{"mcr", "mcrs"},
		{"mve", "mves"},
		{"nat-gateway", "nat-gateways"},
		{"ix", "ixs"},
		{"locations", "location"},
		{"partners", "partner"},
		{"product", "products"},
		{"servicekeys", "servicekey"},
		{"managed-account", "managed-accounts"},
		{"billing-market", "billing-markets"},
		{"users", "user"},
	}

	for _, tc := range cases {
		t.Run(tc.alias, func(t *testing.T) {
			want, _, err := rootCmd.Find([]string{tc.name})
			require.NoError(t, err)
			require.Equal(t, tc.name, want.Name())

			got, _, err := rootCmd.Find([]string{tc.alias})
			require.NoError(t, err)
			assert.Same(t, want, got)
		})
	}
}
