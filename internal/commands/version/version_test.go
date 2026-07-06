package version

import (
	"bytes"
	"os/exec"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			switch args[0] {
			case "describe":
				return exec.Command("false")
			case "rev-parse":
				return exec.Command("echo", "abc1234")
			default:
				return exec.Command("echo", "unexpected command")
			}
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

func TestVersionCommand_InjectedVersionIgnoresGitTag(t *testing.T) {
	// A release binary (version injected via ldflags) run inside a tagged git
	// repo must report the injected version, and must not shell out to git.
	origVersion := version
	origExec := execCommand
	defer func() {
		version = origVersion
		execCommand = origExec
	}()

	gitCalls := 0
	execCommand = func(command string, args ...string) *exec.Cmd {
		gitCalls++
		return exec.Command("echo", "v99.99.99")
	}
	version = "1.2.3"

	t.Setenv("MEGAPORT_CONFIG_DIR", t.TempDir())
	t.Setenv("NO_UPDATE_CHECK", "1")

	rootCmd := &cobra.Command{Use: "test-cli"}
	AddCommandsTo(rootCmd)

	versionCmd, _, err := rootCmd.Find([]string{"version"})
	assert.NoError(t, err)
	require.NotNil(t, versionCmd.RunE)

	buf := new(bytes.Buffer)
	versionCmd.SetOut(buf)
	versionCmd.SetErr(buf)
	assert.NoError(t, versionCmd.RunE(versionCmd, []string{}))

	assert.Contains(t, buf.String(), "Megaport CLI Version: 1.2.3")
	assert.NotContains(t, buf.String(), "v99.99.99")
	assert.Equal(t, 0, gitCalls, "git must not be invoked when version is injected")
}

func TestVersionCommand_DevFallsBackToGit(t *testing.T) {
	origVersion := version
	origExec := execCommand
	defer func() {
		version = origVersion
		execCommand = origExec
	}()

	execCommand = func(command string, args ...string) *exec.Cmd {
		return exec.Command("echo", "v4.5.6")
	}
	version = "dev"

	t.Setenv("MEGAPORT_CONFIG_DIR", t.TempDir())
	t.Setenv("NO_UPDATE_CHECK", "1")

	rootCmd := &cobra.Command{Use: "test-cli"}
	AddCommandsTo(rootCmd)

	versionCmd, _, err := rootCmd.Find([]string{"version"})
	assert.NoError(t, err)
	require.NotNil(t, versionCmd.RunE)

	buf := new(bytes.Buffer)
	versionCmd.SetOut(buf)
	versionCmd.SetErr(buf)
	assert.NoError(t, versionCmd.RunE(versionCmd, []string{}))

	assert.Contains(t, buf.String(), "Megaport CLI Version: v4.5.6")
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
