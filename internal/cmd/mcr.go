package cmd

import (
	"github.com/spf13/cobra"
)

// mcrCmd is the parent command for all operations related to Megaport Cloud Routers (MCRs).
// It serves as a container for subcommands that manage and retrieve information about MCRs.
//
// Example usage:
//
//	megaport mcr get [mcrUID]
var mcrCmd = &cobra.Command{
	Use:   "mcr",
	Short: "Manage MCRs in the Megaport API",
	Long: `Manage MCRs in the Megaport API.

This command groups all operations related to Megaport Cloud Routers (MCRs).
You can use the subcommands to perform actions such as retrieving details for a specific MCR.
For instance, use the "megaport mcr get [mcrUID]" command to fetch details for the MCR with the given UID.

Available subcommands:
  - get: Retrieve details for a single MCR.
  - buy: Purchase an MCR by providing the necessary details.
  - update: Update an existing MCR.
  - delete: Delete an MCR from your account.
  - restore: Restore a previously deleted MCR.
  - create-prefix-filter-list: Create a prefix filter list on an MCR.
  - list-prefix-filter-lists: List all prefix filter lists for a specific MCR.
  - get-prefix-filter-list: Retrieve details for a single prefix filter list on an MCR.
  - update-prefix-filter-list: Update a prefix filter list on an MCR.
  - delete-prefix-filter-list: Delete a prefix filter list on an MCR.
`,
}

// getMCRCmd retrieves and displays detailed information for a single Megaport Cloud Router (MCR).
var getMCRCmd = &cobra.Command{
	Use:   "get [mcrUID]",
	Short: "Get details for a single MCR",
	Args:  cobra.ExactArgs(1),
	RunE:  WrapRunE(GetMCR),
}

// buyMCRCmd allows you to purchase an MCR by providing the necessary details.
var buyMCRCmd = &cobra.Command{
	Use:   "buy",
	Short: "Buy an MCR through the Megaport API",
	Long: `Buy an MCR through the Megaport API.

This command allows you to purchase an MCR by providing the necessary details.
You can provide details in one of three ways:

1. Interactive Mode (with --interactive):
   The command will prompt you for each required and optional field.

2. Flag Mode:
   Provide all required fields as flags:
   --name, --term, --port-speed, --location-id

3. JSON Mode:
   Provide a JSON string or file with all required fields:
   --json <json-string> or --json-file <path>

Required fields:
  - name: The name of the MCR.
  - term: The term of the MCR (1, 12, 24, or 36 months).
  - port_speed: The speed of the MCR (1000, 2500, 5000, or 10000 Mbps).
  - location_id: The ID of the location where the MCR will be provisioned.

Optional fields:
  - mcr_asn: The ASN for the MCR.
  - diversity_zone: The diversity zone for the MCR.
  - cost_centre: The cost center for the MCR.
  - promo_code: A promotional code for the MCR.
  - resource_tags: Key-value tags to associate with the MCR (JSON format).

Example usage:

  # Interactive mode
  megaport mcr buy --interactive

  # Flag mode
  megaport mcr buy --name "My MCR" --term 12 --port-speed 5000 --location-id 123

  # JSON mode
  megaport mcr buy --json '{"name":"My MCR","term":12,"portSpeed":5000,"locationId":123}'
  megaport mcr buy --json-file ./mcr-config.json
`,
	RunE: WrapRunE(BuyMCR),
}

// updateMCRCmd updates an existing Megaport Cloud Router (MCR).
var updateMCRCmd = &cobra.Command{
	Use:   "update [mcrUID]",
	Short: "Update an existing MCR",
	Long: `Update an existing Megaport Cloud Router (MCR).

This command allows you to update the details of an existing MCR.
You can provide details in one of three ways:

1. Interactive Mode (with --interactive):
   The command will prompt you for each field you want to update.

2. Flag Mode:
   Provide fields as flags:
   --name, --cost-centre, --marketplace-visibility, --term

3. JSON Mode:
   Provide a JSON string or file with fields to update:
   --json <json-string> or --json-file <path>

Fields that can be updated:
  - name: The new name of the MCR.
  - cost_centre: The new cost center for the MCR.
  - marketplace_visibility: The new marketplace visibility (true/false).
  - term: The new contract term in months (1, 12, 24, or 36).

Example usage:

  # Interactive mode
  megaport mcr update [mcrUID] --interactive

  # Flag mode
  megaport mcr update [mcrUID] --name "Updated MCR" --marketplace-visibility true

  # JSON mode
  megaport mcr update [mcrUID] --json '{"name":"Updated MCR","marketplaceVisibility":true}'
  megaport mcr update [mcrUID] --json-file ./update-mcr-config.json
`,
	Args: cobra.ExactArgs(1),
	RunE: WrapRunE(UpdateMCR),
}

// deleteMCRCmd deletes a Megaport Cloud Router (MCR) from the user's account.
var deleteMCRCmd = &cobra.Command{
	Use:   "delete [mcrUID]",
	Short: "Delete an MCR from your account",
	Args:  cobra.ExactArgs(1),
	RunE:  WrapRunE(DeleteMCR),
}

// restoreMCRCmd restores a previously deleted Megaport Cloud Router (MCR).
var restoreMCRCmd = &cobra.Command{
	Use:   "restore [mcrUID]",
	Short: "Restore a deleted MCR",
	Args:  cobra.ExactArgs(1),
	RunE:  WrapRunE(RestoreMCR),
}

// createMCRPrefixFilterListCmd creates a prefix filter list on an MCR.
var createMCRPrefixFilterListCmd = &cobra.Command{
	Use:   "create-prefix-filter-list [mcrUID]",
	Short: "Create a prefix filter list on an MCR",
	Long: `Create a prefix filter list on an MCR.

This command allows you to create a new prefix filter list on an MCR.
You can provide details in one of three ways:

1. Interactive Mode (with --interactive):
   The command will prompt you for each required field.

2. Flag Mode:
   Provide all required fields as flags:
   --description, --address-family, --entries

3. JSON Mode:
   Provide a JSON string or file with all required fields:
   --json <json-string> or --json-file <path>

Required fields:
  - description: The description of the prefix filter list.
  - address_family: The address family (IPv4/IPv6).
  - entries: JSON array of prefix filter entries. Each entry has:
      - action: "permit" or "deny"
      - prefix: CIDR notation (e.g., "192.168.0.0/16")
      - ge (optional): Greater than or equal to value
      - le (optional): Less than or equal to value

Example usage:

  # Interactive mode
  megaport mcr create-prefix-filter-list [mcrUID] --interactive

  # Flag mode
  megaport mcr create-prefix-filter-list [mcrUID] --description "My prefix list" --address-family "IPv4" --entries '[{"action":"permit","prefix":"10.0.0.0/8","ge":24,"le":32}]'

  # JSON mode
  megaport mcr create-prefix-filter-list [mcrUID] --json '{"description":"My prefix list","addressFamily":"IPv4","entries":[{"action":"permit","prefix":"10.0.0.0/8","ge":24,"le":32}]}'
  megaport mcr create-prefix-filter-list [mcrUID] --json-file ./prefix-list-config.json
`,
	Args: cobra.ExactArgs(1),
	RunE: WrapRunE(CreateMCRPrefixFilterList),
}

// listMCRPrefixFilterListsCmd lists all prefix filter lists for a specific MCR.
var listMCRPrefixFilterListsCmd = &cobra.Command{
	Use:   "list-prefix-filter-lists [mcrUID]",
	Short: "List all prefix filter lists for a specific MCR",
	Args:  cobra.ExactArgs(1),
	RunE:  WrapRunE(ListMCRPrefixFilterLists),
}

// getMCRPrefixFilterListCmd retrieves and displays details for a single prefix filter list on an MCR.
var getMCRPrefixFilterListCmd = &cobra.Command{
	Use:   "get-prefix-filter-list [mcrUID] [prefixFilterListID]",
	Short: "Get details for a single prefix filter list on an MCR",
	Args:  cobra.ExactArgs(2),
	RunE:  WrapRunE(GetMCRPrefixFilterList),
}

// updateMCRPrefixFilterListCmd updates a prefix filter list on an MCR.
var updateMCRPrefixFilterListCmd = &cobra.Command{
	Use:   "update-prefix-filter-list [mcrUID] [prefixFilterListID]",
	Short: "Update a prefix filter list on an MCR",
	Long: `Update a prefix filter list on an MCR.

This command allows you to update the details of an existing prefix filter list on an MCR.
You can provide details in one of three ways:

1. Interactive Mode (with --interactive):
   The command will prompt you for each field you want to update.

2. Flag Mode:
   Provide fields as flags:
   --description, --address-family, --entries

3. JSON Mode:
   Provide a JSON string or file with fields to update:
   --json <json-string> or --json-file <path>

Fields that can be updated:
  - description: The new description of the prefix filter list.
  - address_family: The new address family (IPv4/IPv6).
  - entries: JSON array of prefix filter entries. Each entry has:
      - action: "permit" or "deny"
      - prefix: CIDR notation (e.g., "192.168.0.0/16")
      - ge (optional): Greater than or equal to value
      - le (optional): Less than or equal to value

Example usage:

  # Interactive mode
  megaport mcr update-prefix-filter-list [mcrUID] [prefixFilterListID] --interactive

  # Flag mode
  megaport mcr update-prefix-filter-list [mcrUID] [prefixFilterListID] --description "Updated prefix list" --entries '[{"action":"permit","prefix":"10.0.0.0/8","ge":24,"le":32}]'

  # JSON mode
  megaport mcr update-prefix-filter-list [mcrUID] [prefixFilterListID] --json '{"description":"Updated prefix list","entries":[{"action":"permit","prefix":"10.0.0.0/8","ge":24,"le":32}]}'
  megaport mcr update-prefix-filter-list [mcrUID] [prefixFilterListID] --json-file ./update-prefix-list.json
`,
	Args: cobra.ExactArgs(2),
	RunE: WrapRunE(UpdateMCRPrefixFilterList),
}

// deleteMCRPrefixFilterListCmd deletes a prefix filter list on an MCR.
var deleteMCRPrefixFilterListCmd = &cobra.Command{
	Use:   "delete-prefix-filter-list [mcrUID] [prefixFilterListID]",
	Short: "Delete a prefix filter list on an MCR",
	Args:  cobra.ExactArgs(2),
	RunE:  WrapRunE(DeleteMCRPrefixFilterList),
}

func init() {
	buyMCRCmd.Flags().BoolP("interactive", "i", false, "Use interactive mode with prompts")
	buyMCRCmd.Flags().String("name", "", "MCR name")
	buyMCRCmd.Flags().Int("term", 0, "Contract term in months (1, 12, 24, or 36)")
	buyMCRCmd.Flags().Int("port-speed", 0, "Port speed in Mbps (1000, 2500, 5000, or 10000)")
	buyMCRCmd.Flags().Int("location-id", 0, "Location ID where the MCR will be provisioned")
	buyMCRCmd.Flags().Int("mcr-asn", 0, "ASN for the MCR (optional)")
	buyMCRCmd.Flags().String("diversity-zone", "", "Diversity zone for the MCR")
	buyMCRCmd.Flags().String("cost-centre", "", "Cost centre for billing")
	buyMCRCmd.Flags().String("promo-code", "", "Promotional code for discounts")
	buyMCRCmd.Flags().String("resource-tags", "", "JSON string of key-value resource tags")
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

	updateMCRPrefixFilterListCmd.Flags().BoolP("interactive", "i", false, "Use interactive mode with prompts")
	updateMCRPrefixFilterListCmd.Flags().String("description", "", "New description of the prefix filter list")
	updateMCRPrefixFilterListCmd.Flags().String("address-family", "", "New address family (IPv4 or IPv6)")
	updateMCRPrefixFilterListCmd.Flags().String("entries", "", "JSON array of prefix filter entries")
	updateMCRPrefixFilterListCmd.Flags().String("json", "", "JSON string containing prefix filter list configuration")
	updateMCRPrefixFilterListCmd.Flags().String("json-file", "", "Path to JSON file containing prefix filter list configuration")

	mcrCmd.AddCommand(getMCRCmd)
	mcrCmd.AddCommand(buyMCRCmd)
	mcrCmd.AddCommand(updateMCRCmd)
	deleteMCRCmd.Flags().Bool("now", false, "Delete immediately instead of at the end of the billing period")
	deleteMCRCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	mcrCmd.AddCommand(deleteMCRCmd)
	mcrCmd.AddCommand(restoreMCRCmd)
	mcrCmd.AddCommand(createMCRPrefixFilterListCmd)
	mcrCmd.AddCommand(listMCRPrefixFilterListsCmd)
	mcrCmd.AddCommand(getMCRPrefixFilterListCmd)
	mcrCmd.AddCommand(updateMCRPrefixFilterListCmd)
	mcrCmd.AddCommand(deleteMCRPrefixFilterListCmd)
	rootCmd.AddCommand(mcrCmd)
}
