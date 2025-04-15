package ports

import (
	"github.com/megaport/megaport-cli/internal/base/cmdbuilder"
	"github.com/spf13/cobra"
)

// AddCommandsTo builds the ports commands and adds them to the root command
func AddCommandsTo(rootCmd *cobra.Command) {
	// Create ports parent command
	portsCmd := cmdbuilder.NewCommand("ports", "Manage ports in the Megaport API").
		WithLongDesc("Manage ports in the Megaport API.\n\nThis command groups operations related to ports. You can use the subcommands to list all ports, get details for a specific port, buy a new port, buy a LAG port, update an existing port, delete a port, restore a deleted port, lock a port, unlock a port, and check VLAN availability on a port.").
		WithExample("megaport-cli ports list").
		WithExample("megaport-cli ports get [portUID]").
		WithExample("megaport-cli ports buy --interactive").
		WithExample("megaport-cli ports buy-lag --interactive").
		WithExample("megaport-cli ports update [portUID] --name \"Updated Port Name\"").
		WithExample("megaport-cli ports delete [portUID]").
		WithExample("megaport-cli ports restore [portUID]").
		WithExample("megaport-cli ports lock [portUID]").
		WithExample("megaport-cli ports unlock [portUID]").
		WithExample("megaport-cli ports check-vlan [portUID] [vlan]").
		WithRootCmd(rootCmd).
		Build()

	buyPortCmd := cmdbuilder.NewCommand("buy", "Buy a port through the Megaport API").
		WithColorAwareRunFunc(BuyPort).
		WithInteractiveFlag().
		WithPortCreationFlags().
		WithJSONConfigFlags().
		WithLongDesc("Buy a port through the Megaport API.\n\nThis command allows you to purchase a port by providing the necessary details.").
		WithDocumentedRequiredFlag("name", "The name of the port (1-64 characters)").
		WithDocumentedRequiredFlag("term", "The term of the port (1, 12, 24, or 36 months)").
		WithDocumentedRequiredFlag("port-speed", "The speed of the port (1000, 10000, or 100000 Mbps)").
		WithDocumentedRequiredFlag("location-id", "The ID of the location where the port will be provisioned").
		WithDocumentedRequiredFlag("marketplace-visibility", "Whether the port should be visible in the marketplace (true or false)").
		WithOptionalFlag("diversity-zone", "The diversity zone for the port").
		WithOptionalFlag("cost-centre", "The cost centre for the port").
		WithOptionalFlag("promo-code", "A promotional code for the port").
		WithExample("megaport-cli ports buy --interactive").
		WithExample(`megaport-cli ports buy --name "My Port" --term 12 --port-speed 10000 --location-id 123 --marketplace-visibility true`).
		WithExample(`megaport-cli ports buy --name "My Port" --term 12 --port-speed 10000 --location-id 123 --marketplace-visibility true --resource-tags '{"env":"prod","owner":"network-team"}'`).
		WithExample("megaport-cli ports buy --json '{\"name\":\"My Port\",\"term\":12,\"portSpeed\":10000,\"locationId\":123,\"marketPlaceVisibility\":true}'").
		WithExample("megaport-cli ports buy --json-file ./port-config.json").
		WithJSONExample(`{
  "name": "My Port",
  "term": 12,
  "portSpeed": 10000,
  "locationId": 123,
  "marketPlaceVisibility": true,
  "diversityZone": "blue",
  "costCentre": "IT-2023",
  "resourceTags": {
    "environment": "production",
    "department": "networking",
    "project": "cloud-migration",
    "owner": "john.doe@example.com"
  }
}`).
		WithImportantNote("Resource tags allow you to categorize resources for organization, filtering, and billing purposes").
		WithImportantNote("Required flags (name, term, port-speed, location-id, marketplace-visibility) can be skipped when using --interactive, --json, or --json-file").
		WithRootCmd(rootCmd).
		WithConditionalRequirements("name", "term", "port-speed", "location-id", "marketplace-visibility").
		Build()

	// Create buy LAG port command
	buyLagCmd := cmdbuilder.NewCommand("buy-lag", "Buy a LAG port through the Megaport API").
		WithColorAwareRunFunc(BuyLAGPort).
		WithInteractiveFlag().
		WithPortLAGFlags().
		WithJSONConfigFlags().
		WithLongDesc("Buy a LAG port through the Megaport API.\n\nThis command allows you to purchase a LAG port by providing the necessary details.").
		WithDocumentedRequiredFlag("name", "The name of the port (1-64 characters)").
		WithDocumentedRequiredFlag("term", "The term of the port (1, 12, or 24 months)").
		WithDocumentedRequiredFlag("port-speed", "The speed of each LAG member port (10000 or 100000 Mbps)").
		WithDocumentedRequiredFlag("location-id", "The ID of the location where the port will be provisioned").
		WithDocumentedRequiredFlag("lag-count", "The number of LAG members (between 1 and 8)").
		WithDocumentedRequiredFlag("marketplace-visibility", "Whether the port should be visible in the marketplace (true or false)").
		WithOptionalFlag("diversity-zone", "The diversity zone for the LAG port").
		WithOptionalFlag("cost-centre", "The cost centre for the LAG port").
		WithOptionalFlag("promo-code", "A promotional code for the LAG port").
		WithExample("megaport-cli ports buy-lag --interactive").
		WithExample(`megaport-cli ports buy-lag --name "My LAG Port" --term 12 --port-speed 10000 --location-id 123 --lag-count 2 --marketplace-visibility true`).
		WithExample(`megaport-cli ports buy-lag --name "My LAG Port" --term 12 --port-speed 10000 --location-id 123 --lag-count 2 --marketplace-visibility true --resource-tags '{"env":"prod","owner":"network-team"}'`).
		WithExample("megaport-cli ports buy-lag --json '{\"name\":\"My LAG Port\",\"term\":12,\"portSpeed\":10000,\"locationId\":123,\"lagCount\":2,\"marketPlaceVisibility\":true}'").
		WithJSONExample(`{
  "name": "My LAG Port",
  "term": 12,
  "portSpeed": 10000,
  "locationId": 123,
  "lagCount": 2,
  "marketPlaceVisibility": true,
  "diversityZone": "blue",
  "costCentre": "IT-2023",
  "resourceTags": {
    "environment": "production",
    "department": "networking",
    "application": "database-cluster"
  }
}`).
		WithImportantNote("Resource tags allow you to categorize resources for organization, filtering, and billing purposes").
		WithImportantNote("Required flags can be skipped when using --interactive, --json, or --json-file").
		WithRootCmd(rootCmd).
		WithConditionalRequirements("name", "term", "port-speed", "location-id", "lag-count", "marketplace-visibility").
		Build()

	// Update port command
	updatePortCmd := cmdbuilder.NewCommand("update", "Update a port's details").
		WithArgs(cobra.ExactArgs(1)).
		WithColorAwareRunFunc(UpdatePort).
		WithInteractiveFlag().
		WithPortUpdateFlags().
		WithJSONConfigFlags().
		WithLongDesc("Update a port's details in the Megaport API.\n\nThis command allows you to update the details of an existing port by providing the necessary fields.").
		WithOptionalFlag("name", "The new name of the port (1-64 characters)").
		WithOptionalFlag("marketplace-visibility", "Whether the port should be visible in the marketplace (true or false)").
		WithOptionalFlag("cost-centre", "The cost centre for billing purposes").
		WithOptionalFlag("term", "The new contract term in months (1, 12, 24, or 36)").
		// At least one flag must be provided when not using interactive or JSON
		WithImportantNote("At least one update flag must be provided when not using --interactive, --json, or --json-file").
		WithRootCmd(rootCmd).
		// Ensure at least one flag is set, but only when not using interactive/JSON
		WithConditionalRequirements("at_least_one:name,marketplace-visibility,cost-centre,term").
		Build()

		// Create list ports command
	listPortsCmd := cmdbuilder.NewCommand("list", "List all ports with optional filters").
		WithOutputFormatRunFunc(ListPorts).
		WithPortFilterFlags().
		WithLongDesc("List all ports available in the Megaport API.\n\nThis command fetches and displays a list of ports with details such as port ID, name, location, speed, and status. By default, only active ports are shown.").
		WithOptionalFlag("location-id", "Filter ports by location ID").
		WithOptionalFlag("port-speed", "Filter ports by port speed").
		WithOptionalFlag("port-name", "Filter ports by port name").
		WithOptionalFlag("lag-only", "Show only LAG ports").
		WithOptionalFlag("include-inactive", "Include ports in CANCELLED, DECOMMISSIONED, or DECOMMISSIONING states").
		WithExample("megaport-cli ports list").
		WithExample("megaport-cli ports list --location-id 1").
		WithExample("megaport-cli ports list --port-speed 10000").
		WithExample("megaport-cli ports list --port-name \"Data Center Primary\"").
		WithExample("megaport-cli ports list --include-inactive").
		WithExample("megaport-cli ports list --location-id 1 --port-speed 10000 --port-name \"Data Center Primary\"").
		WithRootCmd(rootCmd).
		Build()

	// Create get port command
	getPortCmd := cmdbuilder.NewCommand("get", "Get details for a single port").
		WithArgs(cobra.ExactArgs(1)).
		WithOutputFormatRunFunc(GetPort).
		WithLongDesc("Get details for a single port from the Megaport API.\n\nThis command fetches and displays detailed information about a specific port. You need to provide the UID of the port as an argument.").
		WithExample("megaport-cli ports get port-abc123").
		WithExample("megaport-cli ports get 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p").
		WithRootCmd(rootCmd).
		Build()

	// Create delete port command
	deletePortCmd := cmdbuilder.NewCommand("delete", "Delete a port from your account").
		WithArgs(cobra.ExactArgs(1)).
		WithColorAwareRunFunc(DeletePort).
		WithDeleteFlags().
		WithLongDesc("Delete a port from your account in the Megaport API.\n\nThis command allows you to delete an existing port by providing the UID of the port as an argument. By default, the port will be scheduled for deletion at the end of the current billing period.").
		WithOptionalFlag("now", "Delete the port immediately instead of waiting until the end of the billing period").
		WithOptionalFlag("force", "Skip the confirmation prompt and proceed with deletion").
		WithImportantNote("All VXCs associated with the port must be deleted before the port can be deleted").
		WithImportantNote("You can restore a deleted port before it's fully decommissioned using the 'restore' command").
		WithImportantNote("Once a port is fully decommissioned, restoration is not possible").
		WithExample("megaport-cli ports delete 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p").
		WithExample("megaport-cli ports delete 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p --now").
		WithExample("megaport-cli ports delete 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p --now --force").
		WithRootCmd(rootCmd).
		Build()

	// Create restore port command
	restorePortCmd := cmdbuilder.NewCommand("restore", "Restore a deleted port").
		WithArgs(cobra.ExactArgs(1)).
		WithColorAwareRunFunc(RestorePort).
		WithLongDesc("Restore a previously deleted port in the Megaport API.\n\nThis command allows you to restore a port that has been marked for deletion but not yet fully decommissioned. The port will be reinstated with its original configuration.").
		WithImportantNote("You can only restore ports that are in a \"DECOMMISSIONING\" state").
		WithImportantNote("Once a port is fully decommissioned, it cannot be restored").
		WithImportantNote("The restoration process is immediate but may take a few minutes to complete").
		WithImportantNote("All port attributes will be restored to their pre-deletion state").
		WithImportantNote("You will resume being billed for the port according to your original terms").
		WithExample("megaport-cli ports restore 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p").
		WithRootCmd(rootCmd).
		Build()

	// Create lock port command
	lockPortCmd := cmdbuilder.NewCommand("lock", "Lock a port").
		WithArgs(cobra.ExactArgs(1)).
		WithColorAwareRunFunc(LockPort).
		WithLongDesc("Lock a port in the Megaport API.\n\nThis command allows you to lock an existing port, preventing any changes or modifications to the port or its associated VXCs. Locking a port is useful for ensuring critical infrastructure remains stable and preventing accidental changes.").
		WithImportantNote("The port's configuration cannot be modified").
		WithImportantNote("New VXCs cannot be created on this port").
		WithImportantNote("Existing VXCs cannot be modified or deleted").
		WithImportantNote("The port itself cannot be deleted").
		WithImportantNote("To reverse this action, use the 'unlock' command").
		WithExample("megaport-cli ports lock 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p").
		WithRootCmd(rootCmd).
		Build()

	// Create unlock port command
	unlockPortCmd := cmdbuilder.NewCommand("unlock", "Unlock a port").
		WithArgs(cobra.ExactArgs(1)).
		WithColorAwareRunFunc(UnlockPort).
		WithLongDesc("Unlock a port in the Megaport API.\n\nThis command allows you to unlock a previously locked port, re-enabling the ability to make changes to the port and its associated VXCs.").
		WithImportantNote("The port's configuration can be modified").
		WithImportantNote("New VXCs can be created on this port").
		WithImportantNote("Existing VXCs can be modified or deleted").
		WithImportantNote("The port itself can be deleted if needed").
		WithExample("megaport-cli ports unlock 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p").
		WithRootCmd(rootCmd).
		Build()

	// Create check VLAN command
	checkPortVLANAvailabilityCmd := cmdbuilder.NewCommand("check-vlan", "Check if a VLAN is available on a port").
		WithArgs(cobra.ExactArgs(2)).
		WithColorAwareRunFunc(CheckPortVLANAvailability).
		WithLongDesc("Check if a VLAN is available on a port in the Megaport API.\n\nThis command verifies whether a specific VLAN ID is available for use on a port. This is useful when planning new VXCs to ensure the VLAN ID you want to use is not already in use by another connection.\n\nVLAN ID must be between 2 and 4094 (inclusive).").
		WithExample("megaport-cli ports check-vlan 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p 100").
		WithExample("megaport-cli ports check-vlan port-abc123 500").
		WithRootCmd(rootCmd).
		Build()

	// Add list-tags command
	listTagsCmd := cmdbuilder.NewCommand("list-tags", "List resource tags on a specific port.").
		WithLongDesc("Lists all resource tags associated with a specific port").
		WithArgs(cobra.ExactArgs(1)).
		WithOutputFormatRunFunc(ListPortResourceTags).
		WithExample("megaport port list-tags port-abc123").
		Build()

	// Add update-tags command
	updateTagsCmd := cmdbuilder.NewCommand("update-tags", "Update resource tags on a specific port").
		WithLongDesc("Update resource tags associated with a specific port. Tags can be provided via interactive prompts, JSON string, or JSON file.").
		WithArgs(cobra.ExactArgs(1)).
		WithColorAwareRunFunc(UpdatePortResourceTags).
		WithStandardInputFlags().
		WithExample("megaport-cli ports update-tags port-abc123 --interactive").
		WithExample("megaport-cli ports update-tags port-abc123 --json '{\"env\":\"production\",\"team\":\"network\"}'").
		WithExample("megaport-cli ports update-tags port-abc123 --json-file ./tags.json").
		WithImportantNote("All existing tags will be replaced with the provided tags. To clear all tags, provide an empty tag set.").
		Build()

	// Add commands to their parents
	portsCmd.AddCommand(
		buyPortCmd,
		buyLagCmd,
		listPortsCmd,
		getPortCmd,
		updatePortCmd,
		deletePortCmd,
		restorePortCmd,
		lockPortCmd,
		unlockPortCmd,
		checkPortVLANAvailabilityCmd,
		listTagsCmd,
		updateTagsCmd,
	)
	rootCmd.AddCommand(portsCmd)
}
