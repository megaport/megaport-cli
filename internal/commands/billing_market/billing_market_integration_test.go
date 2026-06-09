//go:build integration

package billing_market

import (
	"encoding/json"
	"testing"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_GetBillingMarkets(t *testing.T) {
	client := testutil.SetupIntegrationClient(t)
	defer testutil.LoginWithClient(t, client)()

	cmd := &cobra.Command{Use: "get"}

	var err error
	captured := output.CaptureOutput(func() {
		err = GetBillingMarkets(cmd, nil, true, "json")
	})
	require.NoError(t, err)

	var markets []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(captured), &markets), "output should be valid JSON")

	if len(markets) == 0 {
		t.Skip("staging account has no billing markets configured")
	}

	for _, m := range markets {
		assert.Contains(t, m, "id", "billing market should have an id field")
		assert.Contains(t, m, "currency", "billing market should have a currency field")
		assert.Contains(t, m, "country", "billing market should have a country field")
	}
}
