package ports

import (
	"github.com/megaport/megaport-cli/internal/base/help"
	"github.com/megaport/megaport-cli/internal/utils"
	"github.com/spf13/cobra"
)

var (
	locationID int
	portSpeed  int
	portName   string
)

// wrapCommandFunc adapts our functions with noColor parameter to standard cobra RunE functions
func wrapCommandFunc(fn func(cmd *cobra.Command, args []string, noColor bool) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// Get noColor value from root command
		noColor, err := cmd.Root().PersistentFlags().GetBool("no-color")
		if err != nil {
			noColor = false // Default to color if flag not found
		}
		return fn(cmd, args, noColor)
	}
}

// portsCmd is the base command for all operations related to ports in the Megaport API.
var portsCmd = &cobra.Command{
	Use:   "ports",
	Short: "Manage ports in the Megaport API",
	RunE:  utils.WrapRunE(GetPortHelp),
}

// buyPortCmd allows you to purchase a port by providing the necessary details.
var buyPortCmd = &cobra.Command{
	Use:   "buy",
	Short: "Buy a port through the Megaport API",
	RunE:  utils.WrapColorAwareRunE(BuyPort),
}

var buyLagCmd = &cobra.Command{
	Use:   "buy-lag",
	Short: "Buy a LAG port through the Megaport API",
	RunE:  utils.WrapColorAwareRunE(BuyLAGPort),
}

// listPortsCmd retrieves and displays all available ports from the Megaport API.
var listPortsCmd = &cobra.Command{
	Use:   "list",
	Short: "List all ports with optional filters",
	RunE:  utils.WrapOutputFormatRunE(ListPorts),
}

// getPortCmd retrieves and displays details for a single port from the Megaport API.
var getPortCmd = &cobra.Command{
	Use:   "get [portUID]",
	Short: "Get details for a single port",
	Args:  cobra.ExactArgs(1),
	RunE:  utils.WrapOutputFormatRunE(GetPort),
}

var updatePortCmd = &cobra.Command{
	Use:   "update [portUID]",
	Short: "Update a port's details",
	Args:  cobra.ExactArgs(1),
	RunE:  utils.WrapColorAwareRunE(UpdatePort),
}

// deletePortCmd deletes a port from the user's account.
var deletePortCmd = &cobra.Command{
	Use:   "delete [portUID]",
	Short: "Delete a port from your account",
	Args:  cobra.ExactArgs(1),
	RunE:  utils.WrapColorAwareRunE(DeletePort),
}

// restorePortCmd restores a previously deleted port.
var restorePortCmd = &cobra.Command{
	Use:   "restore [portUID]",
	Short: "Restore a deleted port",
	Args:  cobra.ExactArgs(1),
	RunE:  utils.WrapColorAwareRunE(RestorePort),
}

// lockPortCmd locks a port in the Megaport API.
var lockPortCmd = &cobra.Command{
	Use:   "lock [portUID]",
	Short: "Lock a port",
	Args:  cobra.ExactArgs(1),
	RunE:  utils.WrapRunE(wrapCommandFunc(LockPort)),
}

// unlockPortCmd unlocks a port in the Megaport API.
var unlockPortCmd = &cobra.Command{
	Use:   "unlock [portUID]",
	Short: "Unlock a port",
	Args:  cobra.ExactArgs(1),
	RunE:  utils.WrapRunE(wrapCommandFunc(UnlockPort)),
}

// checkPortVLANAvailabilityCmd checks if a VLAN is available on a port.
var checkPortVLANAvailabilityCmd = &cobra.Command{
	Use:   "check-vlan [portUID] [vlan]",
	Short: "Check if a VLAN is available on a port",
	Args:  cobra.ExactArgs(2),
	RunE:  utils.WrapRunE(wrapCommandFunc(CheckPortVLANAvailability)),
}

// GetPortHelp is a placeholder function for the ports help command
func GetPortHelp(cmd *cobra.Command, args []string) error {
	return nil
}

func AddCommandsTo(rootCmd *cobra.Command) {
	// Add all port subcommands to the ports command
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

	// Add flags to each command
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

	// Set up help builders for all commands

	// ports command help
	portsHelp := &help.CommandHelpBuilder{
		CommandName: "megaport-cli ports",
		ShortDesc:   "Manage ports in the Megaport API",
		LongDesc:    "Manage ports in the Megaport API.\n\nThis command groups operations related to ports. You can use the subcommands to list all ports, get details for a specific port, buy a new port, buy a LAG port, update an existing port, delete a port, restore a deleted port, lock a port, unlock a port, and check VLAN availability on a port.",
		Examples: []string{
			"ports list",
			"ports get [portUID]",
			"ports buy --interactive",
			"ports buy-lag --interactive",
			"ports update [portUID] --name \"Updated Port Name\"",
			"ports delete [portUID]",
			"ports restore [portUID]",
			"ports lock [portUID]",
			"ports unlock [portUID]",
			"ports check-vlan [portUID] [vlan]",
		},
	}
	portsCmd.Long = portsHelp.Build(rootCmd)

	// buy port help
	buyPortHelp := &help.CommandHelpBuilder{
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
	buyPortCmd.Long = buyPortHelp.Build(rootCmd)

	// buy LAG port help
	buyLagHelp := &help.CommandHelpBuilder{
		CommandName: "megaport-cli ports buy-lag",
		ShortDesc:   "Buy a LAG port through the Megaport API",
		LongDesc:    "Buy a LAG port through the Megaport API.\n\nThis command allows you to purchase a LAG port by providing the necessary details.",
		RequiredFlags: map[string]string{
			"name":                   "The name of the port (1-64 characters)",
			"term":                   "The term of the port (1, 12, or 24 months)",
			"port-speed":             "The speed of each LAG member port (10000 or 100000 Mbps)",
			"location-id":            "The ID of the location where the port will be provisioned",
			"lag-count":              "The number of LAG members (between 1 and 8)",
			"marketplace-visibility": "Whether the port should be visible in the marketplace (true or false)",
		},
		OptionalFlags: map[string]string{
			"diversity-zone": "The diversity zone for the LAG port",
			"cost-centre":    "The cost center for the LAG port",
			"promo-code":     "A promotional code for the LAG port",
		},
		Examples: []string{
			"buy-lag --interactive",
			"buy-lag --name \"My LAG Port\" --term 12 --port-speed 10000 --location-id 123 --lag-count 2 --marketplace-visibility true",
			"buy-lag --json '{\"name\":\"My LAG Port\",\"term\":12,\"portSpeed\":10000,\"locationId\":123,\"lagCount\":2,\"marketPlaceVisibility\":true}'",
			"buy-lag --json-file ./lag-port-config.json",
		},
		JSONExamples: []string{
			`{
  "name": "My LAG Port",
  "term": 12,
  "portSpeed": 10000,
  "locationId": 123,
  "lagCount": 2,
  "marketPlaceVisibility": true,
  "diversityZone": "A",
  "costCentre": "IT-2023"
}`,
		},
	}
	buyLagCmd.Long = buyLagHelp.Build(rootCmd)

	// list ports help
	listPortsHelp := &help.CommandHelpBuilder{
		CommandName: "megaport-cli ports list",
		ShortDesc:   "List all ports with optional filters",
		LongDesc:    "List all ports available in the Megaport API.\n\nThis command fetches and displays a list of ports with details such as port ID, name, location, speed, and status.",
		OptionalFlags: map[string]string{
			"location-id": "Filter ports by location ID",
			"port-speed":  "Filter ports by port speed",
			"port-name":   "Filter ports by port name",
		},
		Examples: []string{
			"list",
			"list --location-id 1",
			"list --port-speed 10000",
			"list --port-name \"Data Center Primary\"",
			"list --location-id 1 --port-speed 10000 --port-name \"Data Center Primary\"",
		},
	}
	listPortsCmd.Long = listPortsHelp.Build(rootCmd)

	// get port help
	getPortHelp := &help.CommandHelpBuilder{
		CommandName: "megaport-cli ports get",
		ShortDesc:   "Get details for a single port",
		LongDesc:    "Get details for a single port from the Megaport API.\n\nThis command fetches and displays detailed information about a specific port. You need to provide the UID of the port as an argument.",
		Examples: []string{
			"get port-abc123",
			"get 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p",
		},
	}
	getPortCmd.Long = getPortHelp.Build(rootCmd)

	// update port help
	updatePortHelp := &help.CommandHelpBuilder{
		CommandName: "megaport-cli ports update",
		ShortDesc:   "Update a port's details",
		LongDesc:    "Update a port's details in the Megaport API.\n\nThis command allows you to update the details of an existing port by providing the necessary fields.",
		OptionalFlags: map[string]string{
			"name":                   "The new name of the port (1-64 characters)",
			"marketplace-visibility": "Whether the port should be visible in the marketplace (true or false)",
			"cost-centre":            "The cost center for billing purposes",
			"term":                   "The new contract term in months (1, 12, 24, or 36)",
		},
		ImportantNotes: []string{
			"The port UID cannot be changed",
			"Technical specifications (speed, location) cannot be modified",
			"Connectivity (VXCs) will not be affected by these changes",
			"Changing the contract term may affect billing immediately",
		},
		Examples: []string{
			"update 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p --interactive",
			"update 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p --name \"Main Data Center Port\" --marketplace-visibility false",
			"update 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p --json '{\"name\":\"Main Data Center Port\",\"marketplaceVisibility\":false}'",
			"update 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p --json-file ./update-port-config.json",
		},
		JSONExamples: []string{
			`{
  "name": "Main Data Center Port",
  "marketplaceVisibility": false,
  "costCentre": "IT-Network-2023",
  "term": 24
}`,
		},
	}
	updatePortCmd.Long = updatePortHelp.Build(rootCmd)

	// delete port help
	deletePortHelp := &help.CommandHelpBuilder{
		CommandName: "megaport-cli ports delete",
		ShortDesc:   "Delete a port from your account",
		LongDesc:    "Delete a port from your account in the Megaport API.\n\nThis command allows you to delete an existing port by providing the UID of the port as an argument. By default, the port will be scheduled for deletion at the end of the current billing period.",
		OptionalFlags: map[string]string{
			"now":   "Delete the port immediately instead of waiting until the end of the billing period",
			"force": "Skip the confirmation prompt and proceed with deletion",
		},
		ImportantNotes: []string{
			"All VXCs associated with the port must be deleted before the port can be deleted",
			"You can restore a deleted port before it's fully decommissioned using the 'restore' command",
			"Once a port is fully decommissioned, restoration is not possible",
		},
		Examples: []string{
			"delete 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p",
			"delete 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p --now",
			"delete 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p --now --force",
		},
	}
	deletePortCmd.Long = deletePortHelp.Build(rootCmd)

	// restore port help
	restorePortHelp := &help.CommandHelpBuilder{
		CommandName: "megaport-cli ports restore",
		ShortDesc:   "Restore a deleted port",
		LongDesc:    "Restore a previously deleted port in the Megaport API.\n\nThis command allows you to restore a port that has been marked for deletion but not yet fully decommissioned. The port will be reinstated with its original configuration.",
		ImportantNotes: []string{
			"You can only restore ports that are in a \"DECOMMISSIONING\" state",
			"Once a port is fully decommissioned, it cannot be restored",
			"The restoration process is immediate but may take a few minutes to complete",
			"All port attributes will be restored to their pre-deletion state",
			"You will resume being billed for the port according to your original terms",
		},
		Examples: []string{
			"restore 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p",
		},
	}
	restorePortCmd.Long = restorePortHelp.Build(rootCmd)

	// lock port help
	lockPortHelp := &help.CommandHelpBuilder{
		CommandName: "megaport-cli ports lock",
		ShortDesc:   "Lock a port",
		LongDesc:    "Lock a port in the Megaport API.\n\nThis command allows you to lock an existing port, preventing any changes or modifications to the port or its associated VXCs. Locking a port is useful for ensuring critical infrastructure remains stable and preventing accidental changes.",
		ImportantNotes: []string{
			"The port's configuration cannot be modified",
			"New VXCs cannot be created on this port",
			"Existing VXCs cannot be modified or deleted",
			"The port itself cannot be deleted",
			"To reverse this action, use the 'unlock' command",
		},
		Examples: []string{
			"lock 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p",
		},
	}
	lockPortCmd.Long = lockPortHelp.Build(rootCmd)

	// unlock port help
	unlockPortHelp := &help.CommandHelpBuilder{
		CommandName: "megaport-cli ports unlock",
		ShortDesc:   "Unlock a port",
		LongDesc:    "Unlock a port in the Megaport API.\n\nThis command allows you to unlock a previously locked port, re-enabling the ability to make changes to the port and its associated VXCs.",
		ImportantNotes: []string{
			"The port's configuration can be modified",
			"New VXCs can be created on this port",
			"Existing VXCs can be modified or deleted",
			"The port itself can be deleted if needed",
		},
		Examples: []string{
			"unlock 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p",
		},
	}
	unlockPortCmd.Long = unlockPortHelp.Build(rootCmd)

	// check VLAN help
	checkVlanHelp := &help.CommandHelpBuilder{
		CommandName: "megaport-cli ports check-vlan",
		ShortDesc:   "Check if a VLAN is available on a port",
		LongDesc:    "Check if a VLAN is available on a port in the Megaport API.\n\nThis command verifies whether a specific VLAN ID is available for use on a port. This is useful when planning new VXCs to ensure the VLAN ID you want to use is not already in use by another connection.\n\nVLAN ID must be between 2 and 4094 (inclusive).",
		Examples: []string{
			"check-vlan 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p 100",
			"check-vlan port-abc123 500",
		},
	}
	checkPortVLANAvailabilityCmd.Long = checkVlanHelp.Build(rootCmd)

	// Add the ports command to the root command
	rootCmd.AddCommand(portsCmd)
}
