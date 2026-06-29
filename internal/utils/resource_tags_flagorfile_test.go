package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseResourceTagsFlagOrFile(t *testing.T) {
	t.Run("string round-trips", func(t *testing.T) {
		tags, err := ParseResourceTagsFlagOrFile(`{"env":"prod"}`, "")
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"env": "prod"}, tags)
	})

	t.Run("file round-trips", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "tags.json")
		require.NoError(t, os.WriteFile(path, []byte(`{"env":"staging","team":"net"}`), 0o600))

		tags, err := ParseResourceTagsFlagOrFile("", path)
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"env": "staging", "team": "net"}, tags)
	})

	t.Run("string takes precedence over file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "tags.json")
		require.NoError(t, os.WriteFile(path, []byte(`{"from":"file"}`), 0o600))

		tags, err := ParseResourceTagsFlagOrFile(`{"from":"string"}`, path)
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"from": "string"}, tags)
	})

	t.Run("neither set yields nil", func(t *testing.T) {
		tags, err := ParseResourceTagsFlagOrFile("", "")
		require.NoError(t, err)
		assert.Nil(t, tags)
	})

	t.Run("malformed file JSON returns parse error", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "tags.json")
		require.NoError(t, os.WriteFile(path, []byte(`{bad}`), 0o600))

		_, err := ParseResourceTagsFlagOrFile("", path)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse resource tags JSON")
	})

	t.Run("missing file returns read error", func(t *testing.T) {
		_, err := ParseResourceTagsFlagOrFile("", filepath.Join(t.TempDir(), "absent.json"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read resource tags file")
	})

	t.Run("empty key from file rejected", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "tags.json")
		require.NoError(t, os.WriteFile(path, []byte(`{"":"x"}`), 0o600))

		_, err := ParseResourceTagsFlagOrFile("", path)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "tag key must not be empty")
	})

	t.Run("non-string value from file rejected, matching the JSON path", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "tags.json")
		require.NoError(t, os.WriteFile(path, []byte(`{"env":123}`), 0o600))

		_, err := ParseResourceTagsFlagOrFile("", path)
		require.Error(t, err)
		assert.Contains(t, err.Error(), `resourceTags value for key "env" must be a string`)
	})
}
