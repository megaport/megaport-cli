package vxc

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildVXCRequestFromFlags_ResourceTagsFile(t *testing.T) {
	newBuyCmd := func(t *testing.T) *cobra.Command {
		cmd := &cobra.Command{Use: "buy"}
		cmd.Flags().String("a-end-uid", "", "")
		cmd.Flags().String("b-end-uid", "", "")
		cmd.Flags().String("name", "", "")
		cmd.Flags().Int("rate-limit", 0, "")
		cmd.Flags().Int("term", 0, "")
		cmd.Flags().Int("a-end-vlan", 0, "")
		cmd.Flags().Int("b-end-vlan", 0, "")
		cmd.Flags().Int("a-end-inner-vlan", 0, "")
		cmd.Flags().Int("b-end-inner-vlan", 0, "")
		cmd.Flags().Int("a-end-vnic-index", 0, "")
		cmd.Flags().Int("b-end-vnic-index", 0, "")
		cmd.Flags().String("promo-code", "", "")
		cmd.Flags().String("service-key", "", "")
		cmd.Flags().String("cost-centre", "", "")
		cmd.Flags().String("a-end-partner-config", "", "")
		cmd.Flags().String("b-end-partner-config", "", "")
		cmd.Flags().String("resource-tags", "", "")
		cmd.Flags().String("resource-tags-file", "", "")
		require.NoError(t, cmd.Flags().Set("name", "Test VXC"))
		require.NoError(t, cmd.Flags().Set("a-end-uid", "a-end-uid-123"))
		require.NoError(t, cmd.Flags().Set("b-end-uid", "b-end-uid-123"))
		require.NoError(t, cmd.Flags().Set("rate-limit", "100"))
		require.NoError(t, cmd.Flags().Set("term", "1"))
		return cmd
	}

	t.Run("tags from file round-trip", func(t *testing.T) {
		cmd := newBuyCmd(t)
		path := filepath.Join(t.TempDir(), "tags.json")
		require.NoError(t, os.WriteFile(path, []byte(`{"env":"prod","team":"networking"}`), 0o600))
		require.NoError(t, cmd.Flags().Set("resource-tags-file", path))

		req, err := buildVXCRequestFromFlags(cmd, context.Background(), &MockVXCService{})
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"env": "prod", "team": "networking"}, req.ResourceTags)
	})

	t.Run("flag string takes precedence over file", func(t *testing.T) {
		cmd := newBuyCmd(t)
		path := filepath.Join(t.TempDir(), "tags.json")
		require.NoError(t, os.WriteFile(path, []byte(`{"from":"file"}`), 0o600))
		require.NoError(t, cmd.Flags().Set("resource-tags", `{"from":"string"}`))
		require.NoError(t, cmd.Flags().Set("resource-tags-file", path))

		req, err := buildVXCRequestFromFlags(cmd, context.Background(), &MockVXCService{})
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"from": "string"}, req.ResourceTags)
	})
}
