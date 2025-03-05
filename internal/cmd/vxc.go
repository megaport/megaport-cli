package cmd

import (
	"github.com/spf13/cobra"
)

// vxcCmd is the base command for all operations related to Virtual Cross Connects (VXCs).
// It groups subcommands for managing VXCs in the Megaport API.
// Use the "megaport vxc get [vxcUID]" command to retrieve detailed information about a specific VXC.
var vxcCmd = &cobra.Command{
	Use:   "vxc",
	Short: "Manage VXCs in the Megaport API",
	Long: `Manage VXCs in the Megaport API.

This command groups all operations related to Virtual Cross Connects (VXCs).
You can use the subcommands to perform actions such as retrieving details for a specific VXC.
For example, use the "megaport vxc get [vxcUID]" command to fetch details for the VXC identified by its UID.
`,
}

// getVXCCmd retrieves detailed information for a single Virtual Cross Connect (VXC).
// This command requires exactly one argument: the UID of the VXC.
// It establishes a context with a timeout, logs into the Megaport API, and uses the API client
// to obtain and then display the VXC details using the configured output format (JSON or table).
//
// Example usage:
//
//	megaport vxc get VXC12345
var getVXCCmd = &cobra.Command{
	Use:   "get [vxcUID]",
	Short: "Get details for a single VXC",
	Args:  cobra.ExactArgs(1),
	RunE:  WrapRunE(GetVXC),
}

var buyVXCCmd = &cobra.Command{
	Use:   "buy",
	Short: "Purchase a new Virtual Cross Connect (VXC)",
	Long: `Purchase a new Virtual Cross Connect (VXC) through the Megaport API.

This command guides you through an interactive process to purchase a VXC by prompting for required and optional details.

Key settings include:

  - A-End Product UID: Source endpoint for your VXC.
  - B-End Product UID: Destination endpoint for your VXC.
  - VXC Name: Descriptive name for your VXC.
  - Rate Limit: Bandwidth in Mbps.
  - Term: Contract length in months.

You'll also be prompted for network settings (VLANs) and optional settings like promo codes.

For connections to AWS, Azure, Google Cloud, or VRouters, the tool will guide you through partner-specific configuration.

Example:

  To purchase a new VXC, simply run:

  megaport vxc buy
`,
	RunE: WrapRunE(BuyVXC),
}

func init() {
	vxcCmd.AddCommand(getVXCCmd)
	vxcCmd.AddCommand(buyVXCCmd)
	rootCmd.AddCommand(vxcCmd)
}
