package mcr

import (
	"github.com/megaport/megaport-cli/internal/base/cmdbuilder"
	"github.com/spf13/cobra"
)

// AddCommandsTo builds the mcr commands and adds them to the root command
func AddCommandsTo(rootCmd *cobra.Command) {
	// Create mcr parent command
	mcrCmd := cmdbuilder.NewCommand("mcr", "Manage MCRs in the Megaport API").
		WithLongDesc("Manage MCRs in the Megaport API.\n\nThis command groups all operations related to Megaport Cloud Routers (MCRs). MCRs are virtual routing appliances that run in the Megaport network, providing interconnection between your cloud environments and the Megaport fabric.").
		WithExample("mcr get [mcrUID]").
		WithExample("mcr buy").
		WithExample("mcr update [mcrUID]").
		WithExample("mcr delete [mcrUID]").
		WithImportantNote("With MCRs you can establish virtual cross-connects (VXCs) to cloud service providers").
		WithImportantNote("Create private network connections between different cloud regions").
		WithImportantNote("Implement hybrid cloud architectures with seamless connectivity").
		WithImportantNote("Peer with other networks using BGP routing").
		WithRootCmd(rootCmd).
		Build()

	// Create get MCR command
	getMCRCmd := cmdbuilder.NewCommand("get", "Get details for a single MCR").
		WithArgs(cobra.ExactArgs(1)).
		WithOutputFormatRunFunc(GetMCR).
		WithLongDesc("Get details for a single MCR.\n\nThis command retrieves and displays detailed information for a single Megaport Cloud Router (MCR). You must provide the unique identifier (UID) of the MCR you wish to retrieve.").
		WithExample("get a1b2c3d4-e5f6-7890-1234-567890abcdef").
		WithImportantNote("The output includes the MCR's UID, name, location ID, port speed, and provisioning status").
		WithRootCmd(rootCmd).
		Build()

	// Create buy MCR command
	buyMCRCmd := cmdbuilder.NewCommand("buy", "Buy an MCR through the Megaport API").
		WithColorAwareRunFunc(BuyMCR).
		WithMCRCreateFlags().
		WithStandardInputFlags().
		WithLongDesc("Buy an MCR through the Megaport API.\n\nThis command allows you to purchase an MCR by providing the necessary details.").
		WithExample("buy --interactive").
		WithExample("buy --name \"My MCR\" --term 12 --port-speed 5000 --location-id 123 --mcr-asn 65000").
		WithExample("buy --json '{\"name\":\"My MCR\",\"term\":12,\"portSpeed\":5000,\"locationId\":123,\"mcrAsn\":65000}'").
		WithExample("buy --json-file ./mcr-config.json").
		WithJSONExample(`{
  "name": "My MCR",
  "term": 12,
  "portSpeed": 5000,
  "locationId": 123,
  "mcrAsn": 65000,
  "diversityZone": "zone-a",
  "costCentre": "IT-Networking",
  "promoCode": "SUMMER2024"
}`).
		WithImportantNote("The location_id must correspond to a valid location in the Megaport API").
		WithImportantNote("The port_speed must be one of the supported speeds (1000, 2500, 5000, or 10000 Mbps)").
		WithImportantNote("If mcr_asn is not provided, a private ASN will be automatically assigned").
		WithRootCmd(rootCmd).
		Build()

	// Create update MCR command
	updateMCRCmd := cmdbuilder.NewCommand("update", "Update an existing MCR").
		WithArgs(cobra.ExactArgs(1)).
		WithColorAwareRunFunc(UpdateMCR).
		WithStandardInputFlags().
		WithLongDesc("Update an existing Megaport Cloud Router (MCR).\n\nThis command allows you to update the details of an existing MCR.").
		WithOptionalFlag("term", "The new contract term in months (1, 12, 24, or 36)").
		WithExample("update [mcrUID] --interactive").
		WithExample("update [mcrUID] --name \"Updated MCR\" --marketplace-visibility true --cost-centre \"Finance\"").
		WithExample("update [mcrUID] --json '{\"name\":\"Updated MCR\",\"marketplaceVisibility\":true,\"costCentre\":\"Finance\"}'").
		WithExample("update [mcrUID] --json-file ./update-mcr-config.json").
		WithJSONExample(`{
  "name": "Updated MCR",
  "marketplaceVisibility": true,
  "costCentre": "Finance",
  "term": 24
}`).
		WithImportantNote("The MCR UID cannot be changed").
		WithImportantNote("Only specified fields will be updated; unspecified fields will remain unchanged").
		WithImportantNote("Ensure the JSON file is correctly formatted").
		WithRootCmd(rootCmd).
		Build()

	// Create delete MCR command
	deleteMCRCmd := cmdbuilder.NewCommand("delete", "Delete an MCR from your account").
		WithArgs(cobra.ExactArgs(1)).
		WithColorAwareRunFunc(DeleteMCR).
		WithDeleteFlags().
		WithLongDesc("Delete an MCR from your account.\n\nThis command allows you to delete an MCR from your account. By default, the MCR will be scheduled for deletion at the end of the current billing period.").
		WithExample("delete [mcrUID]").
		WithExample("delete [mcrUID] --now").
		WithExample("delete [mcrUID] --now --force").
		WithRootCmd(rootCmd).
		Build()

	// Create restore MCR command
	restoreMCRCmd := cmdbuilder.NewCommand("restore", "Restore a deleted MCR").
		WithArgs(cobra.ExactArgs(1)).
		WithColorAwareRunFunc(RestoreMCR).
		WithLongDesc("Restore a previously deleted MCR.\n\nThis command allows you to restore a previously deleted MCR, provided it has not yet been fully decommissioned.").
		WithExample("restore [mcrUID]").
		WithRootCmd(rootCmd).
		Build()

	// Create prefix filter list command
	createMCRPrefixFilterListCmd := cmdbuilder.NewCommand("create-prefix-filter-list", "Create a prefix filter list on an MCR").
		WithArgs(cobra.ExactArgs(1)).
		WithColorAwareRunFunc(CreateMCRPrefixFilterList).
		WithStandardInputFlags().
		WithLongDesc("Create a prefix filter list on an MCR.\n\nThis command allows you to create a new prefix filter list on an MCR. Prefix filter lists are used to control which routes are accepted or advertised by the MCR.").
		WithCreateMCRPrefixFilterListFlags().
		WithExample("create-prefix-filter-list [mcrUID] --interactive").
		WithExample("create-prefix-filter-list [mcrUID] --description \"My prefix list\" --address-family \"IPv4\" --entries '[{\"action\":\"permit\",\"prefix\":\"10.0.0.0/8\",\"ge\":24,\"le\":32}]'").
		WithExample("create-prefix-filter-list [mcrUID] --json '{\"description\":\"My prefix list\",\"addressFamily\":\"IPv4\",\"entries\":[{\"action\":\"permit\",\"prefix\":\"10.0.0.0/8\",\"ge\":24,\"le\":32}]}'").
		WithExample("create-prefix-filter-list [mcrUID] --json-file ./prefix-list-config.json").
		WithJSONExample(`{
  "description": "My prefix list",
  "addressFamily": "IPv4",
  "entries": [
    {
      "action": "permit",
      "prefix": "10.0.0.0/8",
      "ge": 24,
      "le": 32
    },
    {
      "action": "deny",
      "prefix": "0.0.0.0/0"
    }
  ]
}`).
		WithImportantNote("The address_family must be either \"IPv4\" or \"IPv6\"").
		WithImportantNote("The entries must be a valid JSON array of prefix filter entries").
		WithImportantNote("The ge and le values are optional but must be within the range of the prefix length").
		WithRootCmd(rootCmd).
		Build()

	// List prefix filter lists command
	listMCRPrefixFilterListsCmd := cmdbuilder.NewCommand("list-prefix-filter-lists", "List all prefix filter lists for a specific MCR").
		WithArgs(cobra.ExactArgs(1)).
		WithOutputFormatRunFunc(ListMCRPrefixFilterLists).
		WithLongDesc("List all prefix filter lists for a specific MCR.\n\nThis command retrieves and displays a list of all prefix filter lists configured on the specified MCR.").
		WithExample("list-prefix-filter-lists [mcrUID]").
		WithRootCmd(rootCmd).
		Build()

	// Get prefix filter list command
	getMCRPrefixFilterListCmd := cmdbuilder.NewCommand("get-prefix-filter-list", "Get details for a single prefix filter list on an MCR").
		WithArgs(cobra.ExactArgs(2)).
		WithOutputFormatRunFunc(GetMCRPrefixFilterList).
		WithLongDesc("Get details for a single prefix filter list on an MCR.\n\nThis command retrieves and displays detailed information about a specific prefix filter list on the specified MCR.").
		WithExample("get-prefix-filter-list [mcrUID] [prefixFilterListID]").
		WithRootCmd(rootCmd).
		Build()

	// Update prefix filter list command
	updateMCRPrefixFilterListCmd := cmdbuilder.NewCommand("update-prefix-filter-list", "Update a prefix filter list on an MCR").
		WithArgs(cobra.ExactArgs(2)).
		WithColorAwareRunFunc(UpdateMCRPrefixFilterList).
		WithStandardInputFlags().
		WithPrefixFilterFlags().
		WithLongDesc("Update a prefix filter list on an MCR.\n\nThis command allows you to update the details of an existing prefix filter list on an MCR. You can use this command to modify the description, address family, or entries in the list.").
		WithExample("update-prefix-filter-list [mcrUID] [prefixFilterListID] --interactive").
		WithExample("update-prefix-filter-list [mcrUID] [prefixFilterListID] --description \"Updated prefix list\" --entries '[{\"action\":\"permit\",\"prefix\":\"10.0.0.0/8\",\"ge\":24,\"le\":32}]'").
		WithExample("update-prefix-filter-list [mcrUID] [prefixFilterListID] --json '{\"description\":\"Updated prefix list\",\"entries\":[{\"action\":\"permit\",\"prefix\":\"10.0.0.0/8\",\"ge\":24,\"le\":32}]}'").
		WithExample("update-prefix-filter-list [mcrUID] [prefixFilterListID] --json-file ./update-prefix-list.json").
		WithJSONExample(`{
  "description": "Updated prefix list",
  "addressFamily": "IPv4",
  "entries": [
    {
      "action": "permit",
      "prefix": "10.0.0.0/8",
      "ge": 24,
      "le": 32
    },
    {
      "action": "deny",
      "prefix": "0.0.0.0/0"
    }
  ]
}`).
		WithRootCmd(rootCmd).
		Build()

	// Delete prefix filter list command
	deleteMCRPrefixFilterListCmd := cmdbuilder.NewCommand("delete-prefix-filter-list", "Delete a prefix filter list on an MCR").
		WithArgs(cobra.ExactArgs(2)).
		WithColorAwareRunFunc(DeleteMCRPrefixFilterList).
		WithDeleteFlags().
		WithLongDesc("Delete a prefix filter list on an MCR.\n\nThis command allows you to delete a prefix filter list from the specified MCR.").
		WithExample("delete-prefix-filter-list [mcrUID] [prefixFilterListID]").
		WithExample("delete-prefix-filter-list [mcrUID] [prefixFilterListID] --force").
		WithRootCmd(rootCmd).
		Build()

	// Add commands to their parents
	mcrCmd.AddCommand(
		getMCRCmd,
		buyMCRCmd,
		updateMCRCmd,
		deleteMCRCmd,
		restoreMCRCmd,
		createMCRPrefixFilterListCmd,
		listMCRPrefixFilterListsCmd,
		getMCRPrefixFilterListCmd,
		updateMCRPrefixFilterListCmd,
		deleteMCRPrefixFilterListCmd,
	)
	rootCmd.AddCommand(mcrCmd)
}
