package cmd

import (
	"github.com/spf13/cobra"
)

var (
	locationID int
	portSpeed  int
	portName   string
)

// portsCmd is the base command for all operations related to ports in the Megaport API.
// This command serves as a container for subcommands which allow you to list, get details of, and manage ports.
//
// Example usage:
//
//	megaport ports list
//	megaport ports get [portUID]
//	megaport ports buy
//	megaport ports buy-lag
//	megaport ports update [portUID]
//	megaport ports delete [portUID]
//	megaport ports restore [portUID]
//	megaport ports lock [portUID]
//	megaport ports unlock [portUID]
//	megaport ports check-vlan [portUID] [vlan]
var portsCmd = &cobra.Command{
	Use:   "ports",
	Short: "Manage ports in the Megaport API",
	Long: `Manage ports in the Megaport API.

This command groups operations related to ports. You can use the subcommands 
to list all ports, get details for a specific port, buy a new port, buy a LAG port,
update an existing port, delete a port, restore a deleted port, lock a port, unlock a port,
and check VLAN availability on a port.

Examples:
  megaport ports list
  megaport ports get [portUID]
  megaport ports buy
  megaport ports buy-lag
  megaport ports update [portUID]
  megaport ports delete [portUID]
  megaport ports restore [portUID]
  megaport ports lock [portUID]
  megaport ports unlock [portUID]
  megaport ports check-vlan [portUID] [vlan]
`,
}

// buyPortCmd allows you to purchase a port by providing the necessary details.
var buyPortCmd = &cobra.Command{
	Use:   "buy",
	Short: "Buy a port through the Megaport API",
	Long: `Buy a port through the Megaport API.

This command allows you to purchase a port by providing the necessary details.
You will be prompted to enter the required and optional fields.

Required fields:
  - name: The name of the port.
  - term: The term of the port (1, 12, 24, or 36 months).
  - port_speed: The speed of the port (1000, 10000, or 100000 Mbps).
  - location_id: The ID of the location where the port will be provisioned.
  - marketplace_visibility: Whether the port should be visible in the marketplace (true or false).

Optional fields:
  - diversity_zone: The diversity zone for the port.
  - cost_center: The cost center for the port.
  - promo_code: A promotional code for the port.

Example usage:

  megaport ports buy
`,
	RunE: WrapRunE(BuyPort),
}

// buyLagCmd allows you to purchase a LAG port by providing the necessary details.
var buyLagCmd = &cobra.Command{
	Use:   "buy-lag",
	Short: "Buy a LAG port through the Megaport API",
	Long: `Buy a LAG port through the Megaport API.

This command allows you to purchase a LAG port by providing the necessary details.
You will be prompted to enter the required and optional fields.

Required fields:
  - name: The name of the port.
  - term: The term of the port (1, 12, 24, or 36 months).
  - port_speed: The speed of the port (10000 or 100000 Mbps).
  - location_id: The ID of the location where the port will be provisioned.
  - lag_count: The number of LAGs (between 1 and 8).
  - marketplace_visibility: Whether the port should be visible in the marketplace (true or false).

Optional fields:
  - diversity_zone: The diversity zone for the port.
  - cost_center: The cost center for the port.
  - promo_code: A promotional code for the port.

Example usage:

  megaport ports buy-lag
`,
	RunE: WrapRunE(BuyLAGPort),
}

// listPortsCmd retrieves and displays all available ports from the Megaport API.
// Optionally, you can filter ports by location ID, port speed, or port name using flags.
//
// Example usage with filtering:
//
//	megaport ports list --location-id 1 --port-speed 10000 --port-name "PortName"
var listPortsCmd = &cobra.Command{
	Use:   "list",
	Short: "List all ports with optional filters",
	Long: `List all ports available in the Megaport API.

This command fetches and displays a list of ports with details such as
port ID, name, location, speed, and status. You can optionally filter the results 
by passing additional flags such as --location-id, --port-speed, and --port-name.

Example:
  megaport ports list --location-id 1 --port-speed 10000 --port-name "PortName"

If no filtering options are provided, all ports will be listed.
`,
	RunE: WrapRunE(ListPorts),
}

// getPortCmd retrieves and displays details for a single port from the Megaport API.
// This command requires exactly one argument: the UID of the port.
//
// Example usage:
//
//	megaport ports get [portUID]
var getPortCmd = &cobra.Command{
	Use:   "get [portUID]",
	Short: "Get details for a single port",
	Long: `Get details for a single port from the Megaport API.

This command fetches and displays detailed information about a specific port.
You need to provide the UID of the port as an argument.

Example:
  megaport ports get [portUID]
`,
	Args: cobra.ExactArgs(1),
	RunE: WrapRunE(GetPort),
}

// updatePortCmd updates a port's details in the Megaport API.
var updatePortCmd = &cobra.Command{
	Use:   "update [portUID]",
	Short: "Update a port's details",
	Long: `Update a port's details in the Megaport API.

This command allows you to update the details of an existing port by providing the necessary fields.
You need to provide the UID of the port as an argument.

Example usage:

  megaport ports update [portUID]
`,
	Args: cobra.ExactArgs(1),
	RunE: WrapRunE(UpdatePort),
}

// deletePortCmd deletes a port from the user's account.
var deletePortCmd = &cobra.Command{
	Use:   "delete [portUID]",
	Short: "Delete a port from your account",
	Long: `Delete a port from your account in the Megaport API.

This command allows you to delete an existing port by providing the UID of the port as an argument.
You can optionally specify whether to delete the port immediately or at the end of the billing period.

Example usage:

  megaport ports delete [portUID]
`,
	Args: cobra.ExactArgs(1),
	RunE: WrapRunE(DeletePort),
}

// restorePortCmd restores a previously deleted port.
var restorePortCmd = &cobra.Command{
	Use:   "restore [portUID]",
	Short: "Restore a deleted port",
	Long: `Restore a previously deleted port in the Megaport API.

This command allows you to restore a previously deleted port by providing the UID of the port as an argument.

Example usage:

  megaport ports restore [portUID]
`,
	Args: cobra.ExactArgs(1),
	RunE: WrapRunE(RestorePort),
}

// lockPortCmd locks a port in the Megaport API.
var lockPortCmd = &cobra.Command{
	Use:   "lock [portUID]",
	Short: "Lock a port",
	Long: `Lock a port in the Megaport API.

This command allows you to lock an existing port by providing the UID of the port as an argument.

Example usage:

  megaport ports lock [portUID]
`,
	Args: cobra.ExactArgs(1),
	RunE: WrapRunE(LockPort),
}

// unlockPortCmd unlocks a port in the Megaport API.
var unlockPortCmd = &cobra.Command{
	Use:   "unlock [portUID]",
	Short: "Unlock a port",
	Long: `Unlock a port in the Megaport API.

This command allows you to unlock an existing port by providing the UID of the port as an argument.

Example usage:

  megaport ports unlock [portUID]
`,
	Args: cobra.ExactArgs(1),
	RunE: WrapRunE(UnlockPort),
}

// checkPortVLANAvailabilityCmd checks if a VLAN is available on a port.
var checkPortVLANAvailabilityCmd = &cobra.Command{
	Use:   "check-vlan [portUID] [vlan]",
	Short: "Check if a VLAN is available on a port",
	Long: `Check if a VLAN is available on a port in the Megaport API.

This command allows you to check if a specific VLAN is available on an existing port by providing the UID of the port and the VLAN ID as arguments.

Example usage:

  megaport ports check-vlan [portUID] [vlan]
`,
	Args: cobra.ExactArgs(2),
	RunE: WrapRunE(CheckPortVLANAvailability),
}

func init() {
	// Add flags to deletePortCmd
	deletePortCmd.Flags().Bool("now", false, "Delete immediately instead of at the end of the billing period")
	deletePortCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")

	// Add flags to listPortsCmd
	listPortsCmd.Flags().IntVar(&locationID, "location-id", 0, "Filter ports by location ID")
	listPortsCmd.Flags().IntVar(&portSpeed, "port-speed", 0, "Filter ports by port speed")
	listPortsCmd.Flags().StringVar(&portName, "port-name", "", "Filter ports by port name")

	// Add commands to portsCmd
	portsCmd.AddCommand(buyPortCmd)
	portsCmd.AddCommand(buyLagCmd)
	portsCmd.AddCommand(listPortsCmd)
	portsCmd.AddCommand(getPortCmd)
	portsCmd.AddCommand(updatePortCmd)
	portsCmd.AddCommand(deletePortCmd)
	portsCmd.AddCommand(restorePortCmd)
	portsCmd.AddCommand(lockPortCmd)
	portsCmd.AddCommand(unlockPortCmd)
	portsCmd.AddCommand(checkPortVLANAvailabilityCmd)

	// Add portsCmd to rootCmd
	rootCmd.AddCommand(portsCmd)
}
