package mcr

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/megaport/megaport-cli/internal/base/exitcodes"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessFlagMCRInput_ResourceTagsFile(t *testing.T) {
	newBuyCmd := func() *cobra.Command {
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("name", "test-mcr", "")
		cmd.Flags().Int("term", 12, "")
		cmd.Flags().Int("port-speed", 5000, "")
		cmd.Flags().Int("location-id", 1, "")
		cmd.Flags().Int("mcr-asn", 0, "")
		cmd.Flags().String("cost-centre", "", "")
		cmd.Flags().String("promo-code", "", "")
		cmd.Flags().String("diversity-zone", "", "")
		cmd.Flags().String("resource-tags", "", "")
		cmd.Flags().String("resource-tags-file", "", "")
		cmd.Flags().Int("ipsec-tunnel-count", 0, "")
		return cmd
	}

	t.Run("tags from file round-trip", func(t *testing.T) {
		cmd := newBuyCmd()
		path := filepath.Join(t.TempDir(), "tags.json")
		require.NoError(t, os.WriteFile(path, []byte(`{"env":"prod","owner":"net"}`), 0o600))
		require.NoError(t, cmd.Flags().Set("resource-tags-file", path))

		req, err := processFlagMCRInput(cmd)
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"env": "prod", "owner": "net"}, req.ResourceTags)
	})

	t.Run("malformed file JSON is a usage error", func(t *testing.T) {
		cmd := newBuyCmd()
		path := filepath.Join(t.TempDir(), "tags.json")
		require.NoError(t, os.WriteFile(path, []byte(`{bad}`), 0o600))
		require.NoError(t, cmd.Flags().Set("resource-tags-file", path))

		_, err := processFlagMCRInput(cmd)
		require.Error(t, err)

		var cliErr *exitcodes.CLIError
		require.True(t, errors.As(err, &cliErr))
		assert.Equal(t, exitcodes.Usage, cliErr.Code)
	})
}
