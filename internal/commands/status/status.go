package status

import (
	"github.com/megaport/megaport-cli/internal/base/cmdbuilder"
	"github.com/spf13/cobra"
)

// AddCommandsTo builds the status command and adds it to the root command.
func AddCommandsTo(rootCmd *cobra.Command) {
	statusCmd := cmdbuilder.NewCommand("status", "Show a dashboard of all Megaport resources").
		WithLongDesc("Display a combined status view of all Megaport resources.\n\nFetches ports, MCRs, MVEs, VXCs, and IXs in parallel and displays them in a single dashboard. By default, only active resources are shown.").
		WithOutputFormatRunFunc(StatusDashboard).
		WithBoolFlag("include-inactive", false, "Include inactive/decommissioned resources").
		WithExample("megaport-cli status").
		WithExample("megaport-cli status --output json").
		WithExample("megaport-cli status --include-inactive").
		WithRootCmd(rootCmd).
		Build()

	rootCmd.AddCommand(statusCmd)
}
