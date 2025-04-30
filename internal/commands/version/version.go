package version

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/megaport/megaport-cli/internal/base/cmdbuilder"
	"github.com/spf13/cobra"
)

var (
	version     = "dev"
	execCommand = exec.Command
)

func GetGitVersion() string {
	cmd := execCommand("git", "describe", "--tags", "--abbrev=0")
	out, err := cmd.Output()
	if err != nil {
		commitCmd := execCommand("git", "rev-parse", "--short", "HEAD")
		commitOut, commitErr := commitCmd.Output()
		if commitErr != nil {
			return ""
		}
		return "dev-" + strings.TrimSpace(string(commitOut))
	}
	return strings.TrimSpace(string(out))
}

func AddCommandsTo(rootCmd *cobra.Command) {
	if v := GetGitVersion(); v != "" {
		version = v
	}

	versionCmd := cmdbuilder.NewCommand("version", "Print the version number of Megaport CLI").
		WithLongDesc("All software has versions. This is Megaport CLI's.").
		WithRunFunc(func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(cmd.OutOrStdout(), "Megaport CLI Version: %s\n", version)
			return nil
		}).
		WithExample("megaport-cli version").
		WithRootCmd(rootCmd).
		Build()

	rootCmd.AddCommand(versionCmd)
}
