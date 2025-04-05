package mcr

import (
	"github.com/megaport/megaport-cli/internal/base/help"
	"github.com/megaport/megaport-cli/internal/utils"
	"github.com/spf13/cobra"
)

// mcrCmd is the parent command for all operations related to Megaport Cloud Routers (MCRs).
var mcrCmd = &cobra.Command{
	Use:   "mcr",
	Short: "Manage MCRs in the Megaport API",
}

// getMCRCmd retrieves and displays detailed information for a single Megaport Cloud Router (MCR).
var getMCRCmd = &cobra.Command{
	Use:   "get [mcrUID]",
	Short: "Get details for a single MCR",
	Args:  cobra.ExactArgs(1),
	RunE:  utils.WrapOutputFormatRunE(GetMCR),
}

// buyMCRCmd allows you to purchase an MCR by providing the necessary details.
var buyMCRCmd = &cobra.Command{
	Use:   "buy",
	Short: "Buy an MCR through the Megaport API",
	RunE:  utils.WrapColorAwareRunE(BuyMCR),
}

// updateMCRCmd updates an existing Megaport Cloud Router (MCR).
var updateMCRCmd = &cobra.Command{
	Use:   "update [mcrUID]",
	Short: "Update an existing MCR",
	Args:  cobra.ExactArgs(1),
	RunE:  utils.WrapColorAwareRunE(UpdateMCR),
}

// deleteMCRCmd deletes a Megaport Cloud Router (MCR) from the user's account.
var deleteMCRCmd = &cobra.Command{
	Use:   "delete [mcrUID]",
	Short: "Delete an MCR from your account",
	Args:  cobra.ExactArgs(1),
	RunE:  utils.WrapColorAwareRunE(DeleteMCR),
}

// restoreMCRCmd restores a previously deleted Megaport Cloud Router (MCR).
var restoreMCRCmd = &cobra.Command{
	Use:   "restore [mcrUID]",
	Short: "Restore a deleted MCR",
	Args:  cobra.ExactArgs(1),
	RunE:  utils.WrapColorAwareRunE(RestoreMCR),
}

// createMCRPrefixFilterListCmd creates a prefix filter list on an MCR.
var createMCRPrefixFilterListCmd = &cobra.Command{
	Use:   "create-prefix-filter-list [mcrUID]",
	Short: "Create a prefix filter list on an MCR",
	Args:  cobra.ExactArgs(1),
	RunE:  utils.WrapColorAwareRunE(CreateMCRPrefixFilterList),
}

// listMCRPrefixFilterListsCmd lists all prefix filter lists for a specific MCR.
var listMCRPrefixFilterListsCmd = &cobra.Command{
	Use:   "list-prefix-filter-lists [mcrUID]",
	Short: "List all prefix filter lists for a specific MCR",
	Args:  cobra.ExactArgs(1),
	RunE:  utils.WrapOutputFormatRunE(ListMCRPrefixFilterLists),
}

// getMCRPrefixFilterListCmd retrieves and displays details for a single prefix filter list on an MCR.
var getMCRPrefixFilterListCmd = &cobra.Command{
	Use:   "get-prefix-filter-list [mcrUID] [prefixFilterListID]",
	Short: "Get details for a single prefix filter list on an MCR",
	Args:  cobra.ExactArgs(2),
	RunE:  utils.WrapOutputFormatRunE(GetMCRPrefixFilterList),
}

// updateMCRPrefixFilterListCmd updates a prefix filter list on an MCR.
var updateMCRPrefixFilterListCmd = &cobra.Command{
	Use:   "update-prefix-filter-list [mcrUID] [prefixFilterListID]",
	Short: "Update a prefix filter list on an MCR",
	Args:  cobra.ExactArgs(2),
	RunE:  utils.WrapColorAwareRunE(UpdateMCRPrefixFilterList),
}

// deleteMCRPrefixFilterListCmd deletes a prefix filter list on an MCR.
var deleteMCRPrefixFilterListCmd = &cobra.Command{
	Use:   "delete-prefix-filter-list [mcrUID] [prefixFilterListID]",
	Short: "Delete a prefix filter list on an MCR",
	Args:  cobra.ExactArgs(2),
	RunE:  utils.WrapColorAwareRunE(DeleteMCRPrefixFilterList),
}

func AddCommandsTo(rootCmd *cobra.Command) {
	buyMCRCmd.Flags().BoolP("interactive", "i", false, "Use interactive mode with prompts")
	buyMCRCmd.Flags().String("name", "", "MCR name")
	buyMCRCmd.Flags().Int("term", 0, "Contract term in months (1, 12, 24, or 36)")
	buyMCRCmd.Flags().Int("port-speed", 0, "Port speed in Mbps (1000, 2500, 5000, or 10000)")
	buyMCRCmd.Flags().Int("location-id", 0, "Location ID where the MCR will be provisioned")
	buyMCRCmd.Flags().Int("mcr-asn", 0, "ASN for the MCR (optional)")
	buyMCRCmd.Flags().String("diversity-zone", "", "Diversity zone for the MCR")
	buyMCRCmd.Flags().String("cost-centre", "", "Cost centre for billing")
	buyMCRCmd.Flags().String("promo-code", "", "Promotional code for discounts")
	buyMCRCmd.Flags().String("json", "", "JSON string containing MCR configuration")
	buyMCRCmd.Flags().String("json-file", "", "Path to JSON file containing MCR configuration")

	updateMCRCmd.Flags().BoolP("interactive", "i", false, "Use interactive mode with prompts")
	updateMCRCmd.Flags().String("name", "", "New MCR name")
	updateMCRCmd.Flags().String("cost-centre", "", "Cost centre for billing")
	updateMCRCmd.Flags().Bool("marketplace-visibility", false, "Whether the MCR is visible in marketplace")
	updateMCRCmd.Flags().Int("term", 0, "New contract term in months (1, 12, 24, or 36)")
	updateMCRCmd.Flags().String("json", "", "JSON string containing MCR configuration")
	updateMCRCmd.Flags().String("json-file", "", "Path to JSON file containing MCR configuration")

	createMCRPrefixFilterListCmd.Flags().BoolP("interactive", "i", false, "Use interactive mode with prompts")
	createMCRPrefixFilterListCmd.Flags().String("description", "", "Description of the prefix filter list")
	createMCRPrefixFilterListCmd.Flags().String("address-family", "", "Address family (IPv4 or IPv6)")
	createMCRPrefixFilterListCmd.Flags().String("entries", "", "JSON array of prefix filter entries")
	createMCRPrefixFilterListCmd.Flags().String("json", "", "JSON string containing prefix filter list configuration")
	createMCRPrefixFilterListCmd.Flags().String("json-file", "", "Path to JSON file containing prefix filter list configuration")

	deleteMCRCmd.Flags().BoolP("force", "f", false, "Force deletion without confirmation")
	deleteMCRCmd.Flags().Bool("now", false, "Delete MCR immediately instead of at end of billing cycle")

	updateMCRPrefixFilterListCmd.Flags().BoolP("interactive", "i", false, "Use interactive mode with prompts")
	updateMCRPrefixFilterListCmd.Flags().String("description", "", "New description of the prefix filter list")
	updateMCRPrefixFilterListCmd.Flags().String("address-family", "", "New address family (IPv4 or IPv6)")
	updateMCRPrefixFilterListCmd.Flags().String("entries", "", "JSON array of prefix filter entries")
	updateMCRPrefixFilterListCmd.Flags().String("json", "", "JSON string containing prefix filter list configuration")
	updateMCRPrefixFilterListCmd.Flags().String("json-file", "", "Path to JSON file containing prefix filter list configuration")

	deleteMCRPrefixFilterListCmd.Flags().Bool("force", false, "Force deletion without confirmation")

	// Set up help builders for commands

	// mcr command help
	mcrHelp := &help.CommandHelpBuilder{
		CommandName: "megaport-cli mcr",
		ShortDesc:   "Manage MCRs in the Megaport API",
		LongDesc:    "Manage MCRs in the Megaport API.\n\nThis command groups all operations related to Megaport Cloud Routers (MCRs). MCRs are virtual routing appliances that run in the Megaport network, providing interconnection between your cloud environments and the Megaport fabric.",
		Examples: []string{
			"mcr get [mcrUID]",
			"mcr buy",
			"mcr update [mcrUID]",
			"mcr delete [mcrUID]",
		},
		ImportantNotes: []string{
			"With MCRs you can establish virtual cross-connects (VXCs) to cloud service providers",
			"Create private network connections between different cloud regions",
			"Implement hybrid cloud architectures with seamless connectivity",
			"Peer with other networks using BGP routing",
		},
	}
	mcrCmd.Long = mcrHelp.Build(rootCmd)

	// get MCR help
	getMCRHelp := &help.CommandHelpBuilder{
		CommandName: "megaport-cli mcr get",
		ShortDesc:   "Get details for a single MCR",
		LongDesc:    "Get details for a single MCR.\n\nThis command retrieves and displays detailed information for a single Megaport Cloud Router (MCR). You must provide the unique identifier (UID) of the MCR you wish to retrieve.",
		Examples: []string{
			"get a1b2c3d4-e5f6-7890-1234-567890abcdef",
		},
		ImportantNotes: []string{
			"The output includes the MCR's UID, name, location ID, port speed, and provisioning status",
		},
	}
	getMCRCmd.Long = getMCRHelp.Build(rootCmd)

	// buy MCR help
	buyMCRHelp := &help.CommandHelpBuilder{
		CommandName: "megaport-cli mcr buy",
		ShortDesc:   "Buy an MCR through the Megaport API",
		LongDesc:    "Buy an MCR through the Megaport API.\n\nThis command allows you to purchase an MCR by providing the necessary details.",
		RequiredFlags: map[string]string{
			"name":        "The name of the MCR (1-64 characters)",
			"term":        "The contract term for the MCR (1, 12, 24, or 36 months)",
			"port-speed":  "The speed of the MCR (1000, 2500, 5000, or 10000 Mbps)",
			"location-id": "The ID of the location where the MCR will be provisioned",
		},
		OptionalFlags: map[string]string{
			"mcr-asn":        "The ASN for the MCR (64512-65534 for private ASN, or a public ASN)",
			"diversity-zone": "The diversity zone for the MCR",
			"cost-centre":    "The cost center for billing purposes",
			"promo-code":     "A promotional code for discounts",
		},
		Examples: []string{
			"buy --interactive",
			"buy --name \"My MCR\" --term 12 --port-speed 5000 --location-id 123 --mcr-asn 65000",
			"buy --json '{\"name\":\"My MCR\",\"term\":12,\"portSpeed\":5000,\"locationId\":123,\"mcrAsn\":65000}'",
			"buy --json-file ./mcr-config.json",
		},
		JSONExamples: []string{
			`{
  "name": "My MCR",
  "term": 12,
  "portSpeed": 5000,
  "locationId": 123,
  "mcrAsn": 65000,
  "diversityZone": "zone-a",
  "costCentre": "IT-Networking",
  "promoCode": "SUMMER2024"
}`,
		},
		ImportantNotes: []string{
			"The location_id must correspond to a valid location in the Megaport API",
			"The port_speed must be one of the supported speeds (1000, 2500, 5000, or 10000 Mbps)",
			"If mcr_asn is not provided, a private ASN will be automatically assigned",
		},
	}
	buyMCRCmd.Long = buyMCRHelp.Build(rootCmd)

	// update MCR help
	updateMCRHelp := &help.CommandHelpBuilder{
		CommandName: "megaport-cli mcr update",
		ShortDesc:   "Update an existing MCR",
		LongDesc:    "Update an existing Megaport Cloud Router (MCR).\n\nThis command allows you to update the details of an existing MCR.",
		OptionalFlags: map[string]string{
			"name":                   "The new name of the MCR (1-64 characters)",
			"cost-centre":            "The new cost center for the MCR",
			"marketplace-visibility": "Whether the MCR is visible in the marketplace (true/false)",
			"term":                   "The new contract term in months (1, 12, 24, or 36)",
		},
		Examples: []string{
			"update [mcrUID] --interactive",
			"update [mcrUID] --name \"Updated MCR\" --marketplace-visibility true --cost-centre \"Finance\"",
			"update [mcrUID] --json '{\"name\":\"Updated MCR\",\"marketplaceVisibility\":true,\"costCentre\":\"Finance\"}'",
			"update [mcrUID] --json-file ./update-mcr-config.json",
		},
		JSONExamples: []string{
			`{
  "name": "Updated MCR",
  "marketplaceVisibility": true,
  "costCentre": "Finance",
  "term": 24
}`,
		},
		ImportantNotes: []string{
			"The MCR UID cannot be changed",
			"Only specified fields will be updated; unspecified fields will remain unchanged",
			"Ensure the JSON file is correctly formatted",
		},
	}
	updateMCRCmd.Long = updateMCRHelp.Build(rootCmd)

	// delete MCR help
	deleteMCRHelp := &help.CommandHelpBuilder{
		CommandName: "megaport-cli mcr delete",
		ShortDesc:   "Delete an MCR from your account",
		LongDesc:    "Delete an MCR from your account.\n\nThis command allows you to delete an MCR from your account. By default, the MCR will be scheduled for deletion at the end of the current billing period.",
		OptionalFlags: map[string]string{
			"now":   "Delete the MCR immediately instead of at the end of the billing period",
			"force": "Skip the confirmation prompt and proceed with deletion",
		},
		Examples: []string{
			"delete [mcrUID]",
			"delete [mcrUID] --now",
			"delete [mcrUID] --now --force",
		},
	}
	deleteMCRCmd.Long = deleteMCRHelp.Build(rootCmd)

	// restore MCR help
	restoreMCRHelp := &help.CommandHelpBuilder{
		CommandName: "megaport-cli mcr restore",
		ShortDesc:   "Restore a deleted MCR",
		LongDesc:    "Restore a previously deleted MCR.\n\nThis command allows you to restore a previously deleted MCR, provided it has not yet been fully decommissioned.",
		Examples: []string{
			"restore [mcrUID]",
		},
	}
	restoreMCRCmd.Long = restoreMCRHelp.Build(rootCmd)

	// create prefix filter list help
	createPrefixFilterListHelp := &help.CommandHelpBuilder{
		CommandName: "megaport-cli mcr create-prefix-filter-list",
		ShortDesc:   "Create a prefix filter list on an MCR",
		LongDesc:    "Create a prefix filter list on an MCR.\n\nThis command allows you to create a new prefix filter list on an MCR. Prefix filter lists are used to control which routes are accepted or advertised by the MCR.",
		RequiredFlags: map[string]string{
			"description":    "The description of the prefix filter list (1-255 characters)",
			"address-family": "The address family (IPv4 or IPv6)",
			"entries":        "JSON array of prefix filter entries",
		},
		Examples: []string{
			"create-prefix-filter-list [mcrUID] --interactive",
			"create-prefix-filter-list [mcrUID] --description \"My prefix list\" --address-family \"IPv4\" --entries '[{\"action\":\"permit\",\"prefix\":\"10.0.0.0/8\",\"ge\":24,\"le\":32}]'",
			"create-prefix-filter-list [mcrUID] --json '{\"description\":\"My prefix list\",\"addressFamily\":\"IPv4\",\"entries\":[{\"action\":\"permit\",\"prefix\":\"10.0.0.0/8\",\"ge\":24,\"le\":32}]}'",
			"create-prefix-filter-list [mcrUID] --json-file ./prefix-list-config.json",
		},
		JSONExamples: []string{
			`{
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
}`,
		},
		ImportantNotes: []string{
			"The address_family must be either \"IPv4\" or \"IPv6\"",
			"The entries must be a valid JSON array of prefix filter entries",
			"The ge and le values are optional but must be within the range of the prefix length",
		},
	}
	createMCRPrefixFilterListCmd.Long = createPrefixFilterListHelp.Build(rootCmd)

	// list prefix filter lists help
	listPrefixFilterListsHelp := &help.CommandHelpBuilder{
		CommandName: "megaport-cli mcr list-prefix-filter-lists",
		ShortDesc:   "List all prefix filter lists for a specific MCR",
		LongDesc:    "List all prefix filter lists for a specific MCR.\n\nThis command retrieves and displays a list of all prefix filter lists configured on the specified MCR.",
		Examples: []string{
			"list-prefix-filter-lists [mcrUID]",
		},
	}
	listMCRPrefixFilterListsCmd.Long = listPrefixFilterListsHelp.Build(rootCmd)

	// get prefix filter list help
	getPrefixFilterListHelp := &help.CommandHelpBuilder{
		CommandName: "megaport-cli mcr get-prefix-filter-list",
		ShortDesc:   "Get details for a single prefix filter list on an MCR",
		LongDesc:    "Get details for a single prefix filter list on an MCR.\n\nThis command retrieves and displays detailed information about a specific prefix filter list on the specified MCR.",
		Examples: []string{
			"get-prefix-filter-list [mcrUID] [prefixFilterListID]",
		},
	}
	getMCRPrefixFilterListCmd.Long = getPrefixFilterListHelp.Build(rootCmd)

	// update prefix filter list help
	updatePrefixFilterListHelp := &help.CommandHelpBuilder{
		CommandName: "megaport-cli mcr update-prefix-filter-list",
		ShortDesc:   "Update a prefix filter list on an MCR",
		LongDesc:    "Update a prefix filter list on an MCR.\n\nThis command allows you to update the details of an existing prefix filter list on an MCR. You can use this command to modify the description, address family, or entries in the list.",
		OptionalFlags: map[string]string{
			"description":    "The new description of the prefix filter list (1-255 characters)",
			"address-family": "The new address family (IPv4 or IPv6)",
			"entries":        "JSON array of prefix filter entries",
		},
		Examples: []string{
			"update-prefix-filter-list [mcrUID] [prefixFilterListID] --interactive",
			"update-prefix-filter-list [mcrUID] [prefixFilterListID] --description \"Updated prefix list\" --entries '[{\"action\":\"permit\",\"prefix\":\"10.0.0.0/8\",\"ge\":24,\"le\":32}]'",
			"update-prefix-filter-list [mcrUID] [prefixFilterListID] --json '{\"description\":\"Updated prefix list\",\"entries\":[{\"action\":\"permit\",\"prefix\":\"10.0.0.0/8\",\"ge\":24,\"le\":32}]}'",
			"update-prefix-filter-list [mcrUID] [prefixFilterListID] --json-file ./update-prefix-list.json",
		},
		JSONExamples: []string{
			`{
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
}`,
		},
	}
	updateMCRPrefixFilterListCmd.Long = updatePrefixFilterListHelp.Build(rootCmd)

	// delete prefix filter list help
	deletePrefixFilterListHelp := &help.CommandHelpBuilder{
		CommandName: "megaport-cli mcr delete-prefix-filter-list",
		ShortDesc:   "Delete a prefix filter list on an MCR",
		LongDesc:    "Delete a prefix filter list on an MCR.\n\nThis command allows you to delete a prefix filter list from the specified MCR.",
		OptionalFlags: map[string]string{
			"force": "Force deletion without confirmation",
		},
		Examples: []string{
			"delete-prefix-filter-list [mcrUID] [prefixFilterListID]",
			"delete-prefix-filter-list [mcrUID] [prefixFilterListID] --force",
		},
	}
	deleteMCRPrefixFilterListCmd.Long = deletePrefixFilterListHelp.Build(rootCmd)

	mcrCmd.AddCommand(getMCRCmd)
	mcrCmd.AddCommand(buyMCRCmd)
	mcrCmd.AddCommand(updateMCRCmd)
	mcrCmd.AddCommand(deleteMCRCmd)
	mcrCmd.AddCommand(restoreMCRCmd)
	mcrCmd.AddCommand(createMCRPrefixFilterListCmd)
	mcrCmd.AddCommand(listMCRPrefixFilterListsCmd)
	mcrCmd.AddCommand(getMCRPrefixFilterListCmd)
	mcrCmd.AddCommand(updateMCRPrefixFilterListCmd)
	mcrCmd.AddCommand(deleteMCRPrefixFilterListCmd)
	rootCmd.AddCommand(mcrCmd)
}
