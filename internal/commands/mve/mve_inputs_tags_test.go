package mve

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

func TestProcessFlagBuyMVEInput_ResourceTags(t *testing.T) {
	const validVendorConfig = `{"vendor":"aruba","productSize":"MEDIUM","imageId":1,"accountName":"test","accountKey":"test","systemTag":"test"}`

	newBuyCmd := func(t *testing.T) *cobra.Command {
		cmd := &cobra.Command{Use: "buy"}
		cmd.Flags().String("name", "", "")
		cmd.Flags().Int("term", 0, "")
		cmd.Flags().Int("location-id", 0, "")
		cmd.Flags().String("diversity-zone", "", "")
		cmd.Flags().String("promo-code", "", "")
		cmd.Flags().String("cost-centre", "", "")
		cmd.Flags().String("vendor-config", "", "")
		cmd.Flags().String("vnics", "", "")
		cmd.Flags().String("resource-tags", "", "")
		cmd.Flags().String("resource-tags-file", "", "")
		require.NoError(t, cmd.Flags().Set("name", "Test MVE"))
		require.NoError(t, cmd.Flags().Set("term", "12"))
		require.NoError(t, cmd.Flags().Set("location-id", "123"))
		require.NoError(t, cmd.Flags().Set("vendor-config", validVendorConfig))
		return cmd
	}

	t.Run("tags round-trip through the flags path", func(t *testing.T) {
		cmd := newBuyCmd(t)
		require.NoError(t, cmd.Flags().Set("resource-tags", `{"env":"prod","owner":"netops"}`))

		req, err := processFlagBuyMVEInput(cmd)
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"env": "prod", "owner": "netops"}, req.ResourceTags)
	})

	t.Run("tags from file round-trip", func(t *testing.T) {
		cmd := newBuyCmd(t)
		path := filepath.Join(t.TempDir(), "tags.json")
		require.NoError(t, os.WriteFile(path, []byte(`{"env":"staging"}`), 0o600))
		require.NoError(t, cmd.Flags().Set("resource-tags-file", path))

		req, err := processFlagBuyMVEInput(cmd)
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"env": "staging"}, req.ResourceTags)
	})

	t.Run("no tag flags leaves tags nil", func(t *testing.T) {
		cmd := newBuyCmd(t)

		req, err := processFlagBuyMVEInput(cmd)
		require.NoError(t, err)
		assert.Nil(t, req.ResourceTags)
	})

	t.Run("malformed JSON is a usage error", func(t *testing.T) {
		cmd := newBuyCmd(t)
		require.NoError(t, cmd.Flags().Set("resource-tags", `{bad}`))

		_, err := processFlagBuyMVEInput(cmd)
		require.Error(t, err)

		var cliErr *exitcodes.CLIError
		require.True(t, errors.As(err, &cliErr))
		assert.Equal(t, exitcodes.Usage, cliErr.Code)
	})
}

func TestProcessJSONBuyMVEInput_ResourceTags(t *testing.T) {
	const baseBody = `{
		"name": "Test MVE",
		"term": 12,
		"locationId": 123,
		"vendorConfig": {"vendor":"aruba","productSize":"MEDIUM","imageId":1,"accountName":"test","accountKey":"test","systemTag":"test"}`

	t.Run("tags from JSON reach the request", func(t *testing.T) {
		body := baseBody + `,"resourceTags":{"env":"prod","owner":"netops"}}`

		req, err := processJSONBuyMVEInput(body, "")
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"env": "prod", "owner": "netops"}, req.ResourceTags)
	})

	t.Run("no resourceTags leaves tags nil", func(t *testing.T) {
		req, err := processJSONBuyMVEInput(baseBody+`}`, "")
		require.NoError(t, err)
		assert.Nil(t, req.ResourceTags)
	})

	t.Run("empty tag key is rejected as a usage error", func(t *testing.T) {
		body := baseBody + `,"resourceTags":{"":"prod"}}`

		_, err := processJSONBuyMVEInput(body, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "tag key must not be empty")

		var cliErr *exitcodes.CLIError
		require.True(t, errors.As(err, &cliErr))
		assert.Equal(t, exitcodes.Usage, cliErr.Code)
	})

	t.Run("non-string tag value is rejected as a usage error", func(t *testing.T) {
		body := baseBody + `,"resourceTags":{"env":123}}`

		_, err := processJSONBuyMVEInput(body, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be a string")

		var cliErr *exitcodes.CLIError
		require.True(t, errors.As(err, &cliErr))
		assert.Equal(t, exitcodes.Usage, cliErr.Code)
	})

	t.Run("non-object resourceTags is rejected as a usage error", func(t *testing.T) {
		body := baseBody + `,"resourceTags":"not-an-object"}`

		_, err := processJSONBuyMVEInput(body, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "resourceTags must be an object")

		var cliErr *exitcodes.CLIError
		require.True(t, errors.As(err, &cliErr))
		assert.Equal(t, exitcodes.Usage, cliErr.Code)
	})
}
