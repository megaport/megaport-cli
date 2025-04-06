package megaport

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var (
	version = "dev" // Fallback version
)

// getGitVersion retrieves the current git tag
func GetGitVersion() string {
	cmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
	out, err := cmd.Output()
	if err != nil {
		// Fall back to commit hash if no tag
		commitCmd := exec.Command("git", "rev-parse", "--short", "HEAD")
		commitOut, commitErr := commitCmd.Output()
		if commitErr != nil {
			return ""
		}
		return "dev-" + strings.TrimSpace(string(commitOut))
	}
	return strings.TrimSpace(string(out))
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Megaport CLI",
	Long:  `All software has versions. This is Megaport CLI's.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Megaport CLI Version:", version)
	},
}

func AddCommandsTo(rootCmd *cobra.Command) {

	if v := GetGitVersion(); v != "" {
		version = v
	}

	rootCmd.AddCommand(versionCmd)
}
