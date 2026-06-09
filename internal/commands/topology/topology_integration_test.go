//go:build integration

package topology

import (
	"encoding/json"
	"testing"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func integrationTopologyCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "topology"}
	cmd.Flags().Bool("include-inactive", false, "")
	cmd.Flags().String("type", "", "")
	return cmd
}

func TestIntegration_ShowTopology(t *testing.T) {
	client := testutil.SetupIntegrationClient(t)
	defer testutil.LoginWithClient(t, client)()

	cmd := integrationTopologyCmd()

	var err error
	captured := output.CaptureOutput(func() {
		err = ShowTopology(cmd, nil, true, "json")
	})
	require.NoError(t, err)

	var nodes []TopologyNode
	require.NoError(t, json.Unmarshal([]byte(captured), &nodes), "output should be valid JSON")

	if len(nodes) == 0 {
		t.Skip("staging account has no Port/MCR/MVE resources to build a topology from")
	}

	for _, n := range nodes {
		assert.NotEmpty(t, n.UID, "topology node should have a uid")
		assert.NotEmpty(t, n.Type, "topology node should have a type")
	}
}

func TestIntegration_ShowTopology_TypeFilter(t *testing.T) {
	client := testutil.SetupIntegrationClient(t)
	defer testutil.LoginWithClient(t, client)()

	cmd := integrationTopologyCmd()
	require.NoError(t, cmd.Flags().Set("type", "port"))

	var err error
	captured := output.CaptureOutput(func() {
		err = ShowTopology(cmd, nil, true, "json")
	})
	require.NoError(t, err)

	var nodes []TopologyNode
	require.NoError(t, json.Unmarshal([]byte(captured), &nodes), "output should be valid JSON")

	if len(nodes) == 0 {
		t.Skip("staging account has no Port resources to build a topology from")
	}

	for _, n := range nodes {
		assert.Equal(t, "Port", n.Type, "type filter should restrict nodes to Ports")
	}
}
