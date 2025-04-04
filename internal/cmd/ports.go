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
//	megaport-cli ports list
//	megaport-cli ports get [portUID]
//	megaport-cli ports buy
//	megaport-cli ports buy-lag
//	megaport-cli ports update [portUID]
//	megaport-cli ports delete [portUID]
//	megaport-cli ports restore [portUID]
//	megaport-cli ports lock [portUID]
//	megaport-cli ports unlock [portUID]
//	megaport-cli ports check-vlan [portUID] [vlan]
var portsCmd = &cobra.Command{
	Use:   "ports",
	Short: "Manage ports in the Megaport API",
	Long: `Manage ports in the Megaport API.

This command groups operations related to ports. You can use the subcommands 
to list all ports, get details for a specific port, buy a new port, buy a LAG port,
update an existing port, delete a port, restore a deleted port, lock a port, unlock a port,
and check VLAN availability on a port.

Examples:
  # List all ports
  megaport-cli ports list

  # Get details for a specific port
  megaport-cli ports get [portUID]

  # Buy a new port
  megaport-cli ports buy --interactive
  megaport-cli ports buy --name "My Port" --term 12 --port-speed 10000 --location-id 123 --marketplace-visibility true
  megaport-cli ports buy --json '{"name":"My Port","term":12,"portSpeed":10000,"locationId":123,"marketPlaceVisibility":true}'
  megaport-cli ports buy --json-file ./port-config.json

  # Buy a LAG port
  megaport-cli ports buy-lag --interactive
  megaport-cli ports buy-lag --name "My LAG Port" --term 12 --port-speed 10000 --location-id 123 --lag-count 2 --marketplace-visibility true
  megaport-cli ports buy-lag --json '{"name":"My LAG Port","term":12,"portSpeed":10000,"locationId":123,"lagCount":2,"marketPlaceVisibility":true}'
  megaport-cli ports buy-lag --json-file ./lag-port-config.json

  # Update a port
  megaport-cli ports update [portUID] --interactive
  megaport-cli ports update [portUID] --name "Updated Port" --marketplace-visibility true
  megaport-cli ports update [portUID] --json '{"name":"Updated Port","marketplaceVisibility":true}'
  megaport-cli ports update [portUID] --json-file ./update-port-config.json

  # Delete a port
  megaport-cli ports delete [portUID] --now

  # Restore a deleted port
  megaport-cli ports restore [portUID]

  # Lock a port
  megaport-cli ports lock [portUID]

  # Unlock a port
  megaport-cli ports unlock [portUID]

  # Check VLAN availability on a port
  megaport-cli ports check-vlan [portUID] [vlan]
`,
}

// buyPortCmd allows you to purchase a port by providing the necessary details.
var buyPortCmd = &cobra.Command{
	Use:   "buy",
	Short: "Buy a port through the Megaport API",
	RunE:  WrapRunE(BuyPort),
}

var buyLagCmd = &cobra.Command{
	Use:   "buy-lag",
	Short: "Buy a LAG port through the Megaport API",
	Long: `Buy a LAG port through the Megaport API.

This command allows you to purchase a LAG port by providing the necessary details.
You can provide details in one of three ways:

1. Interactive Mode (with --interactive):
   The command will prompt you for each required and optional field.

2. Flag Mode:
   Provide all required fields as flags:
   --name, --term, --port-speed, --location-id, --lag-count, --marketplace-visibility

3. JSON Mode:
   Provide a JSON string or file with all required fields:
   --json <json-string> or --json-file <path>

Required fields:
  - name: The name of the port.
  - term: The term of the port (1, 12, or 24 months).
  - port_speed: The speed of the port (10000 or 100000 Mbps).
  - location_id: The ID of the location where the port will be provisioned.
  - lag_count: The number of LAGs (between 1 and 8).
  - marketplace_visibility: Whether the port should be visible in the marketplace (true or false).

Optional fields:
  - diversity_zone: The diversity zone for the port.
  - cost_centre: The cost center for the port.
  - promo_code: A promotional code for the port.

Example usage:

  # Interactive mode
  megaport-cli ports buy-lag --interactive

  # Flag mode
  megaport-cli ports buy-lag --name "My LAG Port" --term 12 --port-speed 10000 --location-id 123 --lag-count 2 --marketplace-visibility true

  # JSON mode
  megaport-cli ports buy-lag --json '{"name":"My LAG Port","term":12,"portSpeed":10000,"locationId":123,"lagCount":2,"marketPlaceVisibility":true}'
  megaport-cli ports buy-lag --json-file ./lag-port-config.json
`,
	RunE: WrapRunE(BuyLAGPort),
}

// listPortsCmd retrieves and displays all available ports from the Megaport API.
// Optionally, you can filter ports by location ID, port speed, or port name using flags.
//
// Example usage with filtering:
//
//	megaport-cli ports list --location-id 1 --port-speed 10000 --port-name "PortName"
var listPortsCmd = &cobra.Command{
	Use:   "list",
	Short: "List all ports with optional filters",
	Long: `List all ports available in the Megaport API.

This command fetches and displays a list of ports with details such as
port ID, name, location, speed, and status. You can optionally filter the results 
by passing additional flags such as --location-id, --port-speed, and --port-name.

Example:
  megaport-cli ports list --location-id 1 --port-speed 10000 --port-name "PortName"

If no filtering options are provided, all ports will be listed.
`,
	RunE: WrapRunE(ListPorts),
}

// getPortCmd retrieves and displays details for a single port from the Megaport API.
// This command requires exactly one argument: the UID of the port.
//
// Example usage:
//
//	megaport-cli ports get [portUID]
var getPortCmd = &cobra.Command{
	Use:   "get [portUID]",
	Short: "Get details for a single port",
	Long: `Get details for a single port from the Megaport API.

This command fetches and displays detailed information about a specific port.
You need to provide the UID of the port as an argument.

Example:
  megaport-cli ports get [portUID]
`,
	Args: cobra.ExactArgs(1),
	RunE: WrapRunE(GetPort),
}
var updatePortCmd = &cobra.Command{
	Use:   "update [portUID]",
	Short: "Update a port's details",
	Long: `Update a port's details in the Megaport API.

This command allows you to update the details of an existing port by providing the necessary fields.
You can provide details in one of three ways:

1. Interactive Mode (with --interactive):
   The command will prompt you for each updatable field, showing current values
   and allowing you to make changes. Press ENTER to keep the current value.

2. Flag Mode:
   Provide only the fields you want to update as flags:
   --name, --marketplace-visibility, --cost-centre, --term

3. JSON Mode:
   Provide a JSON string or file with the fields you want to update:
   --json <json-string> or --json-file <path>

Fields that can be updated:
- name: The new name of the port (1-64 characters)
- marketplace_visibility: Whether the port should be visible in the marketplace (true or false)
- cost_centre: The cost center for billing purposes (optional)
- term: The new contract term in months (1, 12, 24, or 36)

Important notes:
- The port UID cannot be changed
- Technical specifications (speed, location) cannot be modified
- Connectivity (VXCs) will not be affected by these changes
- Changing the contract term may affect billing immediately

Example usage:

  # Interactive mode
  megaport-cli ports update 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p --interactive

  # Flag mode
  megaport-cli ports update 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p --name "Main Data Center Port" --marketplace-visibility false

  # JSON mode
  megaport-cli ports update 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p --json '{"name":"Main Data Center Port","marketplaceVisibility":false}'
  megaport-cli ports update 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p --json-file ./update-port-config.json

JSON format example (update-port-config.json):
{
  "name": "Main Data Center Port",
  "marketplaceVisibility": false,
  "costCentre": "IT-Network-2023",
  "term": 24
}

Note the JSON property names differ from flag names:
- Flag: --name                      → JSON: "name"
- Flag: --marketplace-visibility    → JSON: "marketplaceVisibility"
- Flag: --cost-centre               → JSON: "costCentre"
- Flag: --term                      → JSON: "term"

Example successful output:
  Port updated successfully - UID: 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p
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
By default, the port will be scheduled for deletion at the end of the current billing period.

Available flags:
  --now    Delete the port immediately instead of waiting until the end of the billing period.
           Note that immediate deletion may affect billing and cannot be undone.
           
  --force, -f  Skip the confirmation prompt and proceed with deletion.
               Use with caution, as this will immediately execute the delete operation.

Important notes:
- All VXCs associated with the port must be deleted before the port can be deleted
- You can restore a deleted port before it's fully decommissioned using the 'restore' command
- Once a port is fully decommissioned, restoration is not possible

Example usage:

  # Delete at the end of the billing period (with confirmation prompt)
  megaport-cli ports delete 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p

  # Delete immediately (with confirmation prompt)
  megaport-cli ports delete 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p --now

  # Delete immediately without confirmation
  megaport-cli ports delete 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p --now --force

Example output:
  Are you sure you want to delete port 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p? (y/n): y
  Port 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p deleted successfully
  The port will be deleted at the end of the current billing period
`,
	Args: cobra.ExactArgs(1),
	RunE: WrapRunE(DeletePort),
}

// restorePortCmd restores a previously deleted port.
var restorePortCmd = &cobra.Command{
	Use:   "restore [portUID]",
	Short: "Restore a deleted port",
	Long: `Restore a previously deleted port in the Megaport API.

This command allows you to restore a port that has been marked for deletion but not yet
fully decommissioned. The port will be reinstated with its original configuration.

Important notes:
- You can only restore ports that are in a "DECOMMISSIONING" state
- Once a port is fully decommissioned, it cannot be restored
- The restoration process is immediate but may take a few minutes to complete
- All port attributes will be restored to their pre-deletion state
- You will resume being billed for the port according to your original terms

Example usage:

  megaport-cli ports restore 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p

Example output:
  Port 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p restored successfully
`,
	Args: cobra.ExactArgs(1),
	RunE: WrapRunE(RestorePort),
}

// lockPortCmd locks a port in the Megaport API.
var lockPortCmd = &cobra.Command{
	Use:   "lock [portUID]",
	Short: "Lock a port",
	Long: `Lock a port in the Megaport API.

This command allows you to lock an existing port, preventing any changes or
modifications to the port or its associated VXCs. Locking a port is useful for
ensuring critical infrastructure remains stable and preventing accidental changes.

When a port is locked:
- The port's configuration cannot be modified
- New VXCs cannot be created on this port
- Existing VXCs cannot be modified or deleted
- The port itself cannot be deleted

To reverse this action, use the 'unlock' command.

Example usage:

  megaport-cli ports lock 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p

Example output:
  Port 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p locked successfully
`,
	Args: cobra.ExactArgs(1),
	RunE: WrapRunE(LockPort),
}

// unlockPortCmd unlocks a port in the Megaport API.
var unlockPortCmd = &cobra.Command{
	Use:   "unlock [portUID]",
	Short: "Unlock a port",
	Long: `Unlock a port in the Megaport API.

This command allows you to unlock a previously locked port, re-enabling the ability
to make changes to the port and its associated VXCs.

When a port is unlocked:
- The port's configuration can be modified
- New VXCs can be created on this port
- Existing VXCs can be modified or deleted
- The port itself can be deleted if needed

Example usage:

  megaport-cli ports unlock 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p

Example output:
  Port 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p unlocked successfully
`,
	Args: cobra.ExactArgs(1),
	RunE: WrapRunE(UnlockPort),
}

// checkPortVLANAvailabilityCmd checks if a VLAN is available on a port.
var checkPortVLANAvailabilityCmd = &cobra.Command{
	Use:   "check-vlan [portUID] [vlan]",
	Short: "Check if a VLAN is available on a port",
	Long: `Check if a VLAN is available on a port in the Megaport API.

This command verifies whether a specific VLAN ID is available for use on a port.
This is useful when planning new VXCs to ensure the VLAN ID you want to use is not
already in use by another connection.

VLAN ID must be between 2 and 4094 (inclusive).

Example usage:

  megaport-cli ports check-vlan 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p 100

Example outputs:
  VLAN 100 is available on port 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p
  
  VLAN 100 is not available on port 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p
`,
	Args: cobra.ExactArgs(2),
	RunE: WrapRunE(CheckPortVLANAvailability),
}

func init() {
	deletePortCmd.Flags().Bool("now", false, "Delete immediately instead of at the end of the billing period")
	deletePortCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")

	listPortsCmd.Flags().IntVar(&locationID, "location-id", 0, "Filter ports by location ID")
	listPortsCmd.Flags().IntVar(&portSpeed, "port-speed", 0, "Filter ports by port speed")
	listPortsCmd.Flags().StringVar(&portName, "port-name", "", "Filter ports by port name")

	buyPortCmd.Flags().BoolP("interactive", "i", false, "Use interactive mode with prompts")
	buyPortCmd.Flags().String("name", "", "Port name")
	buyPortCmd.Flags().Int("term", 0, "Contract term in months (1, 12, 24, or 36)")
	buyPortCmd.Flags().Int("port-speed", 0, "Port speed in Mbps (1000, 10000, or 100000)")
	buyPortCmd.Flags().Int("location-id", 0, "Location ID where the port will be provisioned")
	buyPortCmd.Flags().Bool("marketplace-visibility", false, "Whether the port is visible in marketplace")
	buyPortCmd.Flags().String("diversity-zone", "", "Diversity zone for the port")
	buyPortCmd.Flags().String("cost-centre", "", "Cost centre for billing")
	buyPortCmd.Flags().String("promo-code", "", "Promotional code for discounts")
	buyPortCmd.Flags().String("json", "", "JSON string containing port configuration")
	buyPortCmd.Flags().String("json-file", "", "Path to JSON file containing port configuration")

	buyLagCmd.Flags().BoolP("interactive", "i", false, "Use interactive mode with prompts")
	buyLagCmd.Flags().String("name", "", "Port name")
	buyLagCmd.Flags().Int("term", 0, "Contract term in months (1, 12, 24, or 36)")
	buyLagCmd.Flags().Int("port-speed", 0, "Port speed in Mbps (10000 or 100000)")
	buyLagCmd.Flags().Int("location-id", 0, "Location ID where the port will be provisioned")
	buyLagCmd.Flags().Int("lag-count", 0, "Number of LAGs (1-8)")
	buyLagCmd.Flags().Bool("marketplace-visibility", false, "Whether the port is visible in marketplace")
	buyLagCmd.Flags().String("diversity-zone", "", "Diversity zone for the port")
	buyLagCmd.Flags().String("cost-centre", "", "Cost centre for billing")
	buyLagCmd.Flags().String("promo-code", "", "Promotional code for discounts")
	buyLagCmd.Flags().String("json", "", "JSON string containing port configuration")
	buyLagCmd.Flags().String("json-file", "", "Path to JSON file containing port configuration")

	updatePortCmd.Flags().BoolP("interactive", "i", false, "Use interactive mode with prompts")
	updatePortCmd.Flags().String("name", "", "New port name")
	updatePortCmd.Flags().Bool("marketplace-visibility", false, "Whether the port is visible in marketplace")
	updatePortCmd.Flags().String("cost-centre", "", "Cost centre for billing")
	updatePortCmd.Flags().Int("term", 0, "New contract term in months (1, 12, 24, or 36)")
	updatePortCmd.Flags().String("json", "", "JSON string containing port configuration")
	updatePortCmd.Flags().String("json-file", "", "Path to JSON file containing port configuration")

	buyPortHelp := &CommandHelpBuilder{
		CommandName: "megaport-cli ports buy",
		ShortDesc:   "Buy a port through the Megaport API",
		LongDesc:    "Buy a port through the Megaport API.\n\nThis command allows you to purchase a port by providing the necessary details.",
		RequiredFlags: map[string]string{
			"name":                   "The name of the port (1-64 characters)",
			"term":                   "The term of the port (1, 12, 24, or 36 months)",
			"port-speed":             "The speed of the port (1000, 10000, or 100000 Mbps)",
			"location-id":            "The ID of the location where the port will be provisioned",
			"marketplace-visibility": "Whether the port should be visible in the marketplace (true or false)",
		},
		OptionalFlags: map[string]string{
			"diversity-zone": "The diversity zone for the port",
			"cost-centre":    "The cost center for the port",
			"promo-code":     "A promotional code for the port",
		},
		Examples: []string{
			"buy --interactive",
			"buy --name \"My Port\" --term 12 --port-speed 10000 --location-id 123 --marketplace-visibility true",
			"buy --json '{\"name\":\"My Port\",\"term\":12,\"portSpeed\":10000,\"locationId\":123,\"marketPlaceVisibility\":true}'",
			"buy --json-file ./port-config.json",
		},
		JSONExamples: []string{
			`{
  "name": "My Port",
  "term": 12,
  "portSpeed": 10000,
  "locationId": 123,
  "marketPlaceVisibility": true,
  "diversityZone": "A",
  "costCentre": "IT-2023"
}`,
		},
	}
	buyPortCmd.Long = buyPortHelp.Build()

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

	rootCmd.AddCommand(portsCmd)
}
