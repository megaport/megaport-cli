package version

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/megaport/megaport-cli/internal/base/cmdbuilder"
	"github.com/megaport/megaport-cli/internal/base/output"
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
	versionCmd := cmdbuilder.NewCommand("version", "Print the version number of Megaport CLI").
		WithLongDesc("All software has versions. This is Megaport CLI's.").
		WithColorAwareRunFunc(func(cmd *cobra.Command, args []string, noColor bool) error {
			// Only shell out to git for source builds; a release binary keeps its
			// ldflags-injected version even when run inside a tagged git repo.
			reportedVersion := version
			if reportedVersion == "dev" {
				if v := GetGitVersion(); v != "" {
					reportedVersion = v
				}
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Megaport CLI Version: %s\n", reportedVersion)

			cacheDir := os.Getenv("MEGAPORT_CONFIG_DIR")
			if cacheDir == "" {
				if home, err := os.UserHomeDir(); err == nil {
					cacheDir = filepath.Join(home, ".megaport")
				}
			}
			if cacheDir != "" {
				client := &http.Client{Timeout: 5 * time.Second}
				if latest, hasUpdate := checkForUpdate(reportedVersion, client, cacheDir); hasUpdate {
					output.PrintInfo(
						"Update available: %s (https://github.com/megaport/megaport-cli/releases/latest)",
						noColor, latest,
					)
				}
			}
			return nil
		}).
		WithExample("megaport-cli version").
		WithRootCmd(rootCmd).
		Build()

	rootCmd.AddCommand(versionCmd)
}
