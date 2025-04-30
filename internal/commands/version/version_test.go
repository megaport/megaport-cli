package version

import (
	"bytes"
	"os/exec"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestGetGitVersion_WithMocks(t *testing.T) {
	origExecCommand := execCommand
	defer func() { execCommand = origExecCommand }()

	t.Run("With valid git tag", func(t *testing.T) {
		execCommand = func(command string, args ...string) *exec.Cmd {
			cmd := exec.Command("echo", "v1.2.3")
			return cmd
		}

		result := GetGitVersion()
		assert.Equal(t, "v1.2.3", result)
	})

	t.Run("With no tag, only commit", func(t *testing.T) {
		execCommand = func(command string, args ...string) *exec.Cmd {
			if args[0] == "describe" {
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
		execCommand = func(command string, args ...string) *exec.Cmd {
			return exec.Command("false")
		}

		result := GetGitVersion()
		assert.Equal(t, "", result)
	})
}

func TestAddCommandsTo(t *testing.T) {
	originalVersion := version
	originalExecCommand := execCommand
	defer func() {
		version = originalVersion
		execCommand = originalExecCommand
	}()

	execCommand = func(command string, args ...string) *exec.Cmd {
		return exec.Command("false")
	}

	version = "test-version"

	rootCmd := &cobra.Command{
		Use: "test-cli",
	}

	AddCommandsTo(rootCmd)

	versionCmd, _, err := rootCmd.Find([]string{"version"})
	assert.NoError(t, err)
	assert.NotNil(t, versionCmd)
	assert.Equal(t, "version", versionCmd.Use)

	if runE := versionCmd.RunE; runE != nil {
		buf := new(bytes.Buffer)
		versionCmd.SetOut(buf)
		versionCmd.SetErr(buf)

		err = runE(versionCmd, []string{})
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "Megaport CLI Version: test-version")
	} else {
		t.Error("version command does not have a RunE function")
	}
}

func TestVersionModule(t *testing.T) {
	module := NewModule()

	assert.Equal(t, "version", module.Name())

	rootCmd := &cobra.Command{Use: "test-cli"}
	module.RegisterCommands(rootCmd)

	versionCmd, _, err := rootCmd.Find([]string{"version"})
	assert.NoError(t, err)
	assert.NotNil(t, versionCmd)
}
