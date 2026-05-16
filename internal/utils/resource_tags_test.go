package utils

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/megaport/megaport-cli/internal/base/output"
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
			expectError: "failed to parse JSON",
		},
		{
			name:     "empty JSON object",
			flags:    map[string]string{"json": `{}`},
			expected: map[string]string{},
		},
		{
			name:     "null JSON normalizes to nil map",
			flags:    map[string]string{"json": `null`},
			expected: nil,
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
		assert.Contains(t, err.Error(), "failed to read JSON file")
	})

	t.Run("path traversal rejected", func(t *testing.T) {
		cmd := newCmdWithFlags(t, map[string]string{"json-file": "../../etc/passwd"})
		_, err := ParseResourceTagsInput(cmd)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "path traversal not allowed")
	})

	t.Run("file exceeding size limit rejected", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "big.json")
		big := make([]byte, maxTagsFileSize+1)
		big[0] = '{'
		big[len(big)-1] = '}'
		require.NoError(t, os.WriteFile(path, big, 0644))

		cmd := newCmdWithFlags(t, map[string]string{"json-file": path})
		_, err := ParseResourceTagsInput(cmd)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "exceeds maximum allowed size")
	})

	t.Run("non-regular file rejected", func(t *testing.T) {
		tmpDir := t.TempDir()
		cmd := newCmdWithFlags(t, map[string]string{"json-file": tmpDir})
		_, err := ParseResourceTagsInput(cmd)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not a regular file")
	})

	t.Run("invalid JSON file content", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "bad.json")
		err := os.WriteFile(path, []byte(`not-json`), 0644)
		require.NoError(t, err)

		cmd := newCmdWithFlags(t, map[string]string{"json-file": path})
		_, err = ParseResourceTagsInput(cmd)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse JSON file")
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
			expectError: "failed to parse tags JSON",
		},
		{
			name:        "invalid resource-tags flag",
			flags:       map[string]string{"resource-tags": `bad`},
			expectError: "failed to parse resource-tags JSON",
		},
		{
			name:        "no input",
			flags:       map[string]string{},
			expectError: "no input provided",
		},
		{
			name:     "null tags JSON normalizes to nil map",
			flags:    map[string]string{"tags": `null`},
			expected: nil,
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
		assert.Contains(t, err.Error(), "failed to read tags file")
	})
}

func TestListResourceTags(t *testing.T) {
	// ListResourceTags sets a global output format; restore to "table" after each subtest.
	t.Cleanup(func() { output.SetOutputFormat("table") })

	t.Run("success with tags", func(t *testing.T) {
		listFunc := func(ctx context.Context, uid string) (map[string]string, error) {
			assert.Equal(t, "test-uid-123", uid)
			return map[string]string{"env": "prod", "app": "web"}, nil
		}
		err := ListResourceTags("port", "test-uid-123", true, "json", listFunc)
		assert.NoError(t, err)
	})

	t.Run("empty tags returns no error", func(t *testing.T) {
		listFunc := func(ctx context.Context, uid string) (map[string]string, error) {
			return map[string]string{}, nil
		}
		err := ListResourceTags("port", "uid-1", true, "json", listFunc)
		assert.NoError(t, err)
	})

	t.Run("API error returns error", func(t *testing.T) {
		listFunc := func(ctx context.Context, uid string) (map[string]string, error) {
			return nil, fmt.Errorf("API failure")
		}
		err := ListResourceTags("mcr", "uid-2", true, "table", listFunc)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get resource tags for mcr uid-2")
		assert.Contains(t, err.Error(), "API failure")
	})
}

func newUpdateCmd(t *testing.T, flags map[string]string) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{Use: "update-tags"}
	cmd.Flags().Bool("interactive", false, "")
	cmd.Flags().Bool("force", false, "")
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

func TestUpdateResourceTags(t *testing.T) {
	successListFunc := func(ctx context.Context, uid string) (map[string]string, error) {
		return map[string]string{"existing": "tag"}, nil
	}
	successUpdateFunc := func(ctx context.Context, uid string, tags map[string]string) error {
		return nil
	}

	t.Run("success with JSON input", func(t *testing.T) {
		cmd := newUpdateCmd(t, map[string]string{"json": `{"env":"prod"}`, "force": "true"})
		err := UpdateResourceTags(UpdateTagsOptions{
			ResourceType: "port",
			UID:          "uid-1",
			NoColor:      true,
			Cmd:          cmd,
			ListFunc:     successListFunc,
			UpdateFunc:   successUpdateFunc,
		})
		assert.NoError(t, err)
	})

	t.Run("success with no existing tags skips confirmation", func(t *testing.T) {
		emptyListFunc := func(ctx context.Context, uid string) (map[string]string, error) {
			return map[string]string{}, nil
		}
		cmd := newUpdateCmd(t, map[string]string{"json": `{"env":"prod"}`})
		err := UpdateResourceTags(UpdateTagsOptions{
			ResourceType: "port",
			UID:          "uid-6",
			NoColor:      true,
			Cmd:          cmd,
			ListFunc:     emptyListFunc,
			UpdateFunc:   successUpdateFunc,
		})
		assert.NoError(t, err)
	})

	t.Run("confirmation declined cancels update", func(t *testing.T) {
		original := GetConfirmPrompt()
		SetConfirmPrompt(func(string, bool) bool { return false })
		defer SetConfirmPrompt(original)

		cmd := newUpdateCmd(t, map[string]string{"json": `{"env":"prod"}`})
		err := UpdateResourceTags(UpdateTagsOptions{
			ResourceType: "port",
			UID:          "uid-7",
			NoColor:      true,
			Cmd:          cmd,
			ListFunc:     successListFunc,
			UpdateFunc:   successUpdateFunc,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cancelled by user")
	})

	t.Run("remove all tags uses remove wording in confirmation", func(t *testing.T) {
		var capturedMsg string
		original := GetConfirmPrompt()
		SetConfirmPrompt(func(msg string, noColor bool) bool {
			capturedMsg = msg
			return false
		})
		defer SetConfirmPrompt(original)

		cmd := newUpdateCmd(t, map[string]string{"json": `{}`})
		_ = UpdateResourceTags(UpdateTagsOptions{
			ResourceType: "port",
			UID:          "uid-remove",
			NoColor:      true,
			Cmd:          cmd,
			ListFunc:     successListFunc,
			UpdateFunc:   successUpdateFunc,
		})
		assert.Contains(t, capturedMsg, "remove all")
		assert.NotContains(t, capturedMsg, "replace")
	})

	t.Run("confirmation accepted proceeds with update", func(t *testing.T) {
		original := GetConfirmPrompt()
		SetConfirmPrompt(func(string, bool) bool { return true })
		defer SetConfirmPrompt(original)

		cmd := newUpdateCmd(t, map[string]string{"json": `{"env":"prod"}`})
		err := UpdateResourceTags(UpdateTagsOptions{
			ResourceType: "port",
			UID:          "uid-8",
			NoColor:      true,
			Cmd:          cmd,
			ListFunc:     successListFunc,
			UpdateFunc:   successUpdateFunc,
		})
		assert.NoError(t, err)
	})

	t.Run("failed to parse JSON", func(t *testing.T) {
		cmd := newUpdateCmd(t, map[string]string{"json": `not-json`})
		err := UpdateResourceTags(UpdateTagsOptions{
			ResourceType: "port",
			UID:          "uid-2",
			NoColor:      true,
			Cmd:          cmd,
			ListFunc:     successListFunc,
			UpdateFunc:   successUpdateFunc,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse JSON")
	})

	t.Run("API list error", func(t *testing.T) {
		failListFunc := func(ctx context.Context, uid string) (map[string]string, error) {
			return nil, fmt.Errorf("list failed")
		}
		cmd := newUpdateCmd(t, map[string]string{"json": `{"a":"b"}`})
		err := UpdateResourceTags(UpdateTagsOptions{
			ResourceType: "mcr",
			UID:          "uid-3",
			NoColor:      true,
			Cmd:          cmd,
			ListFunc:     failListFunc,
			UpdateFunc:   successUpdateFunc,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to log in or list existing resource tags")
	})

	t.Run("API update error", func(t *testing.T) {
		failUpdateFunc := func(ctx context.Context, uid string, tags map[string]string) error {
			return fmt.Errorf("update failed")
		}
		cmd := newUpdateCmd(t, map[string]string{"json": `{"env":"staging"}`, "force": "true"})
		err := UpdateResourceTags(UpdateTagsOptions{
			ResourceType: "port",
			UID:          "uid-4",
			NoColor:      true,
			Cmd:          cmd,
			ListFunc:     successListFunc,
			UpdateFunc:   failUpdateFunc,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update resource tags")
	})

	t.Run("no input provided", func(t *testing.T) {
		cmd := newUpdateCmd(t, map[string]string{})
		err := UpdateResourceTags(UpdateTagsOptions{
			ResourceType: "port",
			UID:          "uid-5",
			NoColor:      true,
			Cmd:          cmd,
			ListFunc:     successListFunc,
			UpdateFunc:   successUpdateFunc,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no input provided")
	})
}
