package version

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/megaport/megaport-cli/internal/base/cmdbuilder"
	"github.com/spf13/cobra"
)

var (
	version = "dev" // Fallback version
	// Make execCommand package-level so it can be mocked in tests
	execCommand = exec.Command
)

// GetGitVersion retrieves the current git tag
func GetGitVersion() string {
	cmd := execCommand("git", "describe", "--tags", "--abbrev=0")
	out, err := cmd.Output()
	if err != nil {
		// Fall back to commit hash if no tag
		commitCmd := execCommand("git", "rev-parse", "--short", "HEAD")
		commitOut, commitErr := commitCmd.Output()
		if commitErr != nil {
			return ""
		}
		return "dev-" + strings.TrimSpace(string(commitOut))
	}
	return strings.TrimSpace(string(out))
}

// AddCommandsTo adds the version command to the root command
func AddCommandsTo(rootCmd *cobra.Command) {
	// Set version from git tag or commit hash
	if v := GetGitVersion(); v != "" {
		version = v
	}

	// Create version command using command builder
	versionCmd := cmdbuilder.NewCommand("version", "Print the version number of Megaport CLI").
		WithLongDesc("All software has versions. This is Megaport CLI's.").
		WithRunFunc(func(cmd *cobra.Command, args []string) error {
			// Use cmd.OutOrStdout() instead of fmt.Println for testability
			fmt.Fprintf(cmd.OutOrStdout(), "Megaport CLI Version: %s\n", version)
			return nil
		}).
		WithExample("megaport-cli version").
		WithRootCmd(rootCmd).
		Build()

	rootCmd.AddCommand(versionCmd)
}
