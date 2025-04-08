package version

import (
	"bytes"
	"os/exec"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// TestGetGitVersion_WithMocks tests GetGitVersion using mocked git commands
func TestGetGitVersion_WithMocks(t *testing.T) {
	// Save original exec.Command function
	origExecCommand := execCommand
	defer func() { execCommand = origExecCommand }()

	t.Run("With valid git tag", func(t *testing.T) {
		// Create a simple mock that always returns v1.2.3
		execCommand = func(command string, args ...string) *exec.Cmd {
			cmd := exec.Command("echo", "v1.2.3")
			return cmd
		}

		result := GetGitVersion()
		assert.Equal(t, "v1.2.3", result)
	})

	t.Run("With no tag, only commit", func(t *testing.T) {
		// Mock git describe failure but git rev-parse success
		execCommand = func(command string, args ...string) *exec.Cmd {
			if args[0] == "describe" {
				// This will fail when Output() is called
				return exec.Command("false")
			} else if args[0] == "rev-parse" {
				return exec.Command("echo", "abc1234")
			}
			return exec.Command("echo", "unexpected command")
		}

		result := GetGitVersion()
		assert.Equal(t, "dev-abc1234", result)
	})

	t.Run("With no git info", func(t *testing.T) {
		// Mock both git commands to fail
		execCommand = func(command string, args ...string) *exec.Cmd {
			return exec.Command("false")
		}

		result := GetGitVersion()
		assert.Equal(t, "", result)
	})
}

// TestAddCommandsTo tests the AddCommandsTo function
func TestAddCommandsTo(t *testing.T) {
	// Save current version and restore after test
	originalVersion := version
	originalExecCommand := execCommand
	defer func() {
		version = originalVersion
		execCommand = originalExecCommand
	}()

	// Force git commands to fail so version won't be updated
	execCommand = func(command string, args ...string) *exec.Cmd {
		return exec.Command("false")
	}

	// Set known version for testing
	version = "test-version"

	// Create a test root command
	rootCmd := &cobra.Command{
		Use: "test-cli",
	}

	// Add version command to root
	AddCommandsTo(rootCmd)

	// Check that version command exists
	versionCmd, _, err := rootCmd.Find([]string{"version"})
	assert.NoError(t, err)
	assert.NotNil(t, versionCmd)
	assert.Equal(t, "version", versionCmd.Use)

	// Test by directly invoking the command's run function
	if runE := versionCmd.RunE; runE != nil {
		// Create a buffer to capture the output
		buf := new(bytes.Buffer)
		versionCmd.SetOut(buf)
		versionCmd.SetErr(buf)

		// Execute the command's run function directly
		err = runE(versionCmd, []string{})
		assert.NoError(t, err)

		// Check the output
		output := buf.String()
		assert.Contains(t, output, "Megaport CLI Version: test-version")
	} else {
		t.Error("version command does not have a RunE function")
	}
}

// TestVersionModule tests the version module
func TestVersionModule(t *testing.T) {
	module := NewModule()

	assert.Equal(t, "version", module.Name())

	// Test RegisterCommands
	rootCmd := &cobra.Command{Use: "test-cli"}
	module.RegisterCommands(rootCmd)

	// Check that the version command was registered
	versionCmd, _, err := rootCmd.Find([]string{"version"})
	assert.NoError(t, err)
	assert.NotNil(t, versionCmd)
}
