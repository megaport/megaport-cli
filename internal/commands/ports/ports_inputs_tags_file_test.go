package ports

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

func newBuyPortTagCmd(t *testing.T) *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("name", "", "")
	cmd.Flags().Int("term", 0, "")
	cmd.Flags().Int("port-speed", 0, "")
	cmd.Flags().Int("location-id", 0, "")
	cmd.Flags().Int("lag-count", 0, "")
	cmd.Flags().Bool("marketplace-visibility", false, "")
	cmd.Flags().String("diversity-zone", "", "")
	cmd.Flags().String("cost-centre", "", "")
	cmd.Flags().String("promo-code", "", "")
	cmd.Flags().String("resource-tags", "", "")
	cmd.Flags().String("resource-tags-file", "", "")
	require.NoError(t, cmd.Flags().Set("name", "test-port"))
	require.NoError(t, cmd.Flags().Set("term", "12"))
	require.NoError(t, cmd.Flags().Set("port-speed", "10000"))
	require.NoError(t, cmd.Flags().Set("location-id", "1"))
	require.NoError(t, cmd.Flags().Set("marketplace-visibility", "true"))
	return cmd
}

func TestProcessFlagPortInput_ResourceTagsFile(t *testing.T) {
	cmd := newBuyPortTagCmd(t)
	path := filepath.Join(t.TempDir(), "tags.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"env":"prod"}`), 0o600))
	require.NoError(t, cmd.Flags().Set("resource-tags-file", path))

	req, err := processFlagPortInput(cmd)
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"env": "prod"}, req.ResourceTags)
}

func TestProcessFlagLAGPortInput_ResourceTagsFile(t *testing.T) {
	cmd := newBuyPortTagCmd(t)
	require.NoError(t, cmd.Flags().Set("lag-count", "2"))
	path := filepath.Join(t.TempDir(), "tags.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"env":"prod"}`), 0o600))
	require.NoError(t, cmd.Flags().Set("resource-tags-file", path))

	req, err := processFlagLAGPortInput(cmd)
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"env": "prod"}, req.ResourceTags)
}

func TestProcessFlagPortInput_ResourceTagsFileMalformed(t *testing.T) {
	cmd := newBuyPortTagCmd(t)
	path := filepath.Join(t.TempDir(), "tags.json")
	require.NoError(t, os.WriteFile(path, []byte(`{bad}`), 0o600))
	require.NoError(t, cmd.Flags().Set("resource-tags-file", path))

	_, err := processFlagPortInput(cmd)
	require.Error(t, err)

	var cliErr *exitcodes.CLIError
	require.True(t, errors.As(err, &cliErr))
	assert.Equal(t, exitcodes.Usage, cliErr.Code)
}
