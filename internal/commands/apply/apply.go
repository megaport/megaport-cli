package apply

import (
	"github.com/megaport/megaport-cli/internal/base/cmdbuilder"
	"github.com/spf13/cobra"
)

// AddCommandsTo builds the apply command and adds it to the root command.
func AddCommandsTo(rootCmd *cobra.Command) {
	cmd := cmdbuilder.NewCommand("apply", "Provision multiple resources from a config file").
		WithLongDesc("Provision multiple Megaport resources (ports, MCRs, MVEs, VXCs) from a declarative YAML or JSON config file.\n\nResources are provisioned sequentially in dependency order: ports and MCRs first, then MVEs, then VXCs. VXC endpoints can reference previously provisioned resources using {{.type.name}} template syntax.").
		WithOutputFormatRunFunc(ApplyConfig).
		WithFlagP("file", "f", "", "Path to config file (YAML or JSON)").
		WithBoolFlag("dry-run", false, "Validate all orders without provisioning").
		WithBoolFlagP("yes", "y", false, "Skip confirmation prompt").
		WithExample(`megaport apply -f infrastructure.yaml`).
		WithExample(`megaport apply -f infrastructure.yaml --dry-run`).
		WithExample(`megaport apply -f infrastructure.yaml --yes`).
		WithExample(`megaport apply -f infrastructure.json --output json`).
		WithRootCmd(rootCmd).
		Build()

	rootCmd.AddCommand(cmd)
}
