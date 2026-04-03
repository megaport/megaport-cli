package topology

import (
	"github.com/megaport/megaport-cli/internal/base/cmdbuilder"
	"github.com/spf13/cobra"
)

// AddCommandsTo builds the topology command and adds it to the root command
func AddCommandsTo(rootCmd *cobra.Command) {
	topologyCmd := cmdbuilder.NewCommand("topology", "Show resource relationship tree").
		WithOutputFormatRunFunc(ShowTopology).
		WithBoolFlag("include-inactive", false, "Include deprovisioned resources in the tree").
		WithFlag("type", "", "Filter by resource type: port, mcr, or mve").
		WithLongDesc("Show a tree view of Megaport resources and their VXC connections.\n\nThis command fetches all Ports, MCRs, and MVEs and renders each with its associated Virtual Cross Connects (VXCs) as a tree. The B-End destination of each VXC is shown to illustrate connectivity.\n\nDefault output is a human-readable ASCII tree. Use --output json for structured output.").
		WithExample("megaport-cli topology").
		WithExample("megaport-cli topology --output json").
		WithExample("megaport-cli topology --type mcr").
		WithExample("megaport-cli topology --include-inactive").
		WithImportantNote("Each VXC is shown once, under its A-End parent resource").
		WithImportantNote("CSV and XML output formats are not supported for hierarchical topology data").
		WithRootCmd(rootCmd).
		Build()

	rootCmd.AddCommand(topologyCmd)
}
