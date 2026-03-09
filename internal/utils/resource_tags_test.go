package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmdWithFlags(t *testing.T, flags map[string]string) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("json", "", "")
	cmd.Flags().String("json-file", "", "")
	cmd.Flags().String("tags", "", "")
	cmd.Flags().String("tags-file", "", "")
	cmd.Flags().String("resource-tags", "", "")
	for k, v := range flags {
		require.NoError(t, cmd.Flags().Set(k, v), "failed to set flag %q", k)
	}
	return cmd
}

func TestParseResourceTagsInput(t *testing.T) {
	tests := []struct {
		name        string
		flags       map[string]string
		expected    map[string]string
		expectError string
	}{
		{
			name:     "valid JSON string",
			flags:    map[string]string{"json": `{"env":"prod","team":"platform"}`},
			expected: map[string]string{"env": "prod", "team": "platform"},
		},
		{
			name:        "invalid JSON string",
			flags:       map[string]string{"json": `not-json`},
			expectError: "error parsing JSON",
		},
		{
			name:     "empty JSON object",
			flags:    map[string]string{"json": `{}`},
			expected: map[string]string{},
		},
		{
			name:        "no input provided",
			flags:       map[string]string{},
			expectError: "no input provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newCmdWithFlags(t, tt.flags)
			result, err := ParseResourceTagsInput(cmd)
			if tt.expectError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}

	t.Run("valid JSON file", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "tags.json")
		err := os.WriteFile(path, []byte(`{"env":"staging"}`), 0644)
		require.NoError(t, err)

		cmd := newCmdWithFlags(t, map[string]string{"json-file": path})
		result, err := ParseResourceTagsInput(cmd)
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"env": "staging"}, result)
	})

	t.Run("nonexistent JSON file", func(t *testing.T) {
		cmd := newCmdWithFlags(t, map[string]string{"json-file": "/nonexistent/file.json"})
		_, err := ParseResourceTagsInput(cmd)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "error reading JSON file")
	})

	t.Run("invalid JSON file content", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "bad.json")
		err := os.WriteFile(path, []byte(`not-json`), 0644)
		require.NoError(t, err)

		cmd := newCmdWithFlags(t, map[string]string{"json-file": path})
		_, err = ParseResourceTagsInput(cmd)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "error parsing JSON file")
	})

	t.Run("json takes precedence over json-file", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "tags.json")
		err := os.WriteFile(path, []byte(`{"from":"file"}`), 0644)
		require.NoError(t, err)

		cmd := newCmdWithFlags(t, map[string]string{
			"json":      `{"from":"string"}`,
			"json-file": path,
		})
		result, err := ParseResourceTagsInput(cmd)
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"from": "string"}, result)
	})
}

func TestParseResourceTagsInputExtended(t *testing.T) {
	tests := []struct {
		name        string
		flags       map[string]string
		expected    map[string]string
		expectError string
	}{
		{
			name:     "tags flag",
			flags:    map[string]string{"tags": `{"env":"test"}`},
			expected: map[string]string{"env": "test"},
		},
		{
			name:     "resource-tags flag",
			flags:    map[string]string{"resource-tags": `{"env":"dev"}`},
			expected: map[string]string{"env": "dev"},
		},
		{
			name:        "invalid tags flag",
			flags:       map[string]string{"tags": `bad`},
			expectError: "error parsing tags JSON",
		},
		{
			name:        "invalid resource-tags flag",
			flags:       map[string]string{"resource-tags": `bad`},
			expectError: "error parsing resource-tags JSON",
		},
		{
			name:        "no input",
			flags:       map[string]string{},
			expectError: "no input provided",
		},
		{
			name:     "json takes precedence over tags",
			flags:    map[string]string{"json": `{"from":"json"}`, "tags": `{"from":"tags"}`},
			expected: map[string]string{"from": "json"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newCmdWithFlags(t, tt.flags)
			result, err := parseResourceTagsInputExtended(cmd)
			if tt.expectError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}

	t.Run("tags-file flag", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "tags.json")
		err := os.WriteFile(path, []byte(`{"env":"file"}`), 0644)
		require.NoError(t, err)

		cmd := newCmdWithFlags(t, map[string]string{"tags-file": path})
		result, err := parseResourceTagsInputExtended(cmd)
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"env": "file"}, result)
	})

	t.Run("nonexistent tags-file", func(t *testing.T) {
		cmd := newCmdWithFlags(t, map[string]string{"tags-file": "/nonexistent/file.json"})
		_, err := parseResourceTagsInputExtended(cmd)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "error reading tags file")
	})
}
