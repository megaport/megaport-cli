package cmd

import (
	"github.com/spf13/cobra"
)

// mcrCmd is the parent command for all operations related to Megaport Cloud Routers (MCRs).
var mcrCmd = &cobra.Command{
	Use:   "mcr",
	Short: "Manage MCRs in the Megaport API",
	Long: `Manage MCRs in the Megaport API.

This command groups all operations related to Megaport Cloud Routers (MCRs).
MCRs are virtual routing appliances that run in the Megaport network, providing
interconnection between your cloud environments and the Megaport fabric.

With MCRs you can:
- Establish virtual cross-connects (VXCs) to cloud service providers
- Create private network connections between different cloud regions
- Implement hybrid cloud architectures with seamless connectivity
- Peer with other networks using BGP routing

Available operations:
- get: Retrieve details for a single MCR.
- buy: Purchase a new MCR with specified configuration.
- update: Modify an existing MCR's properties.
- delete: Remove an MCR from your account.
- restore: Restore a previously deleted MCR.
- create-prefix-filter-list: Create a prefix filter list on an MCR.
- list-prefix-filter-lists: List all prefix filter lists for a specific MCR.
- get-prefix-filter-list: Retrieve details for a single prefix filter list on an MCR.
- update-prefix-filter-list: Update a prefix filter list on an MCR.
- delete-prefix-filter-list: Delete a prefix filter list on an MCR.

Examples:
  # Get details for a specific MCR
  megaport-cli mcr get [mcrUID]

  # Buy a new MCR
  megaport-cli mcr buy

  # Update an existing MCR
  megaport-cli mcr update [mcrUID]

  # Delete an existing MCR
  megaport-cli mcr delete [mcrUID]
`,
}

// getMCRCmd retrieves and displays detailed information for a single Megaport Cloud Router (MCR).
var getMCRCmd = &cobra.Command{
	Use:   "get [mcrUID]",
	Short: "Get details for a single MCR",
	Long: `Get details for a single MCR.

This command retrieves and displays detailed information for a single Megaport Cloud Router (MCR).
You must provide the unique identifier (UID) of the MCR you wish to retrieve.

The output includes:
  - UID: Unique identifier of the MCR
  - Name: User-defined name of the MCR
  - Location ID: Physical location where the MCR is provisioned
  - Port Speed: Speed of the MCR (e.g., 1000, 2500, 5000, 10000 Mbps)
  - Provisioning Status: Current provisioning status of the MCR (e.g., Active, Inactive, Deleting)
  
Example usage:
  megaport-cli mcr get a1b2c3d4-e5f6-7890-1234-567890abcdef
`,
	Args: cobra.ExactArgs(1),
	RunE: WrapRunE(GetMCR),
}

// listMCRsCmd lists all MCRs.
var listMCRsCmd = &cobra.Command{
	Use:   "list",
	Short: "List all MCRs",
	RunE:  WrapRunE(ListMCRs),
}

// buyMCRCmd allows you to purchase an MCR by providing the necessary details.
var buyMCRCmd = &cobra.Command{
	Use:   "buy",
	Short: "Buy an MCR through the Megaport API",
	Long: `Buy an MCR through the Megaport API.

This command allows you to purchase an MCR by providing the necessary details.
You can provide details in one of three ways:

1. Interactive Mode (with --interactive):
   The command will prompt you for each required and optional field, guiding you through the configuration process.

2. Flag Mode:
   Provide all required fields as flags:
   --name, --term, --port-speed, --location-id
   Optional fields can also be specified using flags.

3. JSON Mode:
   Provide a JSON string or file with all required and optional fields:
   --json <json-string> or --json-file <path>

Required fields:
  - name: The name of the MCR (1-64 characters).
  - term: The contract term for the MCR (1, 12, 24, or 36 months).
  - port_speed: The speed of the MCR (1000, 2500, 5000, or 10000 Mbps).
  - location_id: The ID of the location where the MCR will be provisioned.

Optional fields:
  - mcr_asn: The ASN for the MCR (64512-65534 for private ASN, or a public ASN). If not provided, a private ASN will be automatically assigned.
  - diversity_zone: The diversity zone for the MCR (if applicable).
  - cost_centre: The cost center for billing purposes.
  - promo_code: A promotional code for discounts (if applicable).

Example usage:

  # Interactive mode
  megaport-cli mcr buy --interactive

  # Flag mode
  megaport-cli mcr buy --name "My MCR" --term 12 --port-speed 5000 --location-id 123 --mcr-asn 65000 --resource-tags '{"environment":"production"}'

  # JSON mode
  megaport-cli mcr buy --json '{"name":"My MCR","term":12,"portSpeed":5000,"locationId":123,"mcrAsn":65000,"resourceTags":{"environment":"production"}}'
  megaport-cli mcr buy --json-file ./mcr-config.json

JSON format example (mcr-config.json):
{
  "name": "My MCR",
  "term": 12,
  "portSpeed": 5000,
  "locationId": 123,
  "mcrAsn": 65000,
  "diversityZone": "zone-a",
  "costCentre": "IT-Networking",
  "promoCode": "SUMMER2024",
}

Notes:
- The location_id must correspond to a valid location in the Megaport API.
- The port_speed must be one of the supported speeds (1000, 2500, 5000, or 10000 Mbps).
- If mcr_asn is not provided, a private ASN will be automatically assigned.
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
   The command will prompt you for each field you want to update, showing the current value and allowing you to modify it.

2. Flag Mode:
   Provide fields as flags:
   --name, --cost-centre, --marketplace-visibility, --term
   Only specified flags will be updated; unspecified fields will remain unchanged.

3. JSON Mode:
   Provide a JSON string or file with fields to update:
   --json <json-string> or --json-file <path>
   Only fields present in the JSON will be updated; unspecified fields will remain unchanged.

Fields that can be updated:
  - name: The new name of the MCR (1-64 characters).
  - cost_centre: The new cost center for the MCR.
  - marketplace_visibility: Whether the MCR is visible in the marketplace (true/false).
  - term: The new contract term in months (1, 12, 24, or 36).

Example usage:

  # Interactive mode
  megaport-cli mcr update [mcrUID] --interactive

  # Flag mode
  megaport-cli mcr update [mcrUID] --name "Updated MCR" --marketplace-visibility true --cost-centre "Finance"

  # JSON mode
  megaport-cli mcr update [mcrUID] --json '{"name":"Updated MCR","marketplaceVisibility":true,"costCentre":"Finance"}'
  megaport-cli mcr update [mcrUID] --json-file ./update-mcr-config.json

JSON format example (update-mcr-config.json):
{
  "name": "Updated MCR",
  "marketplaceVisibility": true,
  "costCentre": "Finance",
  "term": 24
}

Notes:
- The MCR UID cannot be changed.
- Only specified fields will be updated; unspecified fields will remain unchanged.
- Ensure the JSON file is correctly formatted.
`,
	Args: cobra.ExactArgs(1),
	RunE: WrapRunE(UpdateMCR),
}

// deleteMCRCmd deletes a Megaport Cloud Router (MCR) from the user's account.
var deleteMCRCmd = &cobra.Command{
	Use:   "delete [mcrUID]",
	Short: "Delete an MCR from your account",
	Long: `Delete an MCR from your account.

This command allows you to delete an MCR from your account. By default, the MCR
will be scheduled for deletion at the end of the current billing period.

Flags:
  --now: Delete the MCR immediately instead of at the end of the billing period.
  --force, -f: Skip the confirmation prompt and proceed with deletion.

Example usage:
  # Delete MCR at the end of the billing period with confirmation
  megaport-cli mcr delete [mcrUID]

  # Delete MCR immediately with confirmation
  megaport-cli mcr delete [mcrUID] --now

  # Delete MCR immediately without confirmation
  megaport-cli mcr delete [mcrUID] --now --force
`,
	Args: cobra.ExactArgs(1),
	RunE: WrapRunE(DeleteMCR),
}

// restoreMCRCmd restores a previously deleted Megaport Cloud Router (MCR).
var restoreMCRCmd = &cobra.Command{
	Use:   "restore [mcrUID]",
	Short: "Restore a deleted MCR",
	Long: `Restore a previously deleted MCR.

This command allows you to restore a previously deleted MCR, provided it has not
yet been fully decommissioned.

Example usage:
  megaport-cli mcr restore [mcrUID]
`,
	Args: cobra.ExactArgs(1),
	RunE: WrapRunE(RestoreMCR),
}

// createMCRPrefixFilterListCmd creates a prefix filter list on an MCR.
var createMCRPrefixFilterListCmd = &cobra.Command{
	Use:   "create-prefix-filter-list [mcrUID]",
	Short: "Create a prefix filter list on an MCR",
	Long: `Create a prefix filter list on an MCR.

This command allows you to create a new prefix filter list on an MCR.
Prefix filter lists are used to control which routes are accepted or advertised
by the MCR.

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
  - description: The description of the prefix filter list (1-255 characters).
  - address_family: The address family (IPv4 or IPv6).
  - entries: JSON array of prefix filter entries. Each entry has:
      - action: "permit" or "deny"
      - prefix: CIDR notation (e.g., "192.168.0.0/16")
      - ge (optional): Greater than or equal to value (must be less than or equal to the prefix length)
      - le (optional): Less than or equal to value (must be greater than or equal to the prefix length)

Example usage:

  # Interactive mode
  megaport-cli mcr create-prefix-filter-list [mcrUID] --interactive

  # Flag mode
  megaport-cli mcr create-prefix-filter-list [mcrUID] --description "My prefix list" --address-family "IPv4" --entries '[{"action":"permit","prefix":"10.0.0.0/8","ge":24,"le":32}]'

  # JSON mode
  megaport-cli mcr create-prefix-filter-list [mcrUID] --json '{"description":"My prefix list","addressFamily":"IPv4","entries":[{"action":"permit","prefix":"10.0.0.0/8","ge":24,"le":32}]}'
  megaport-cli mcr create-prefix-filter-list [mcrUID] --json-file ./prefix-list-config.json

JSON format example (prefix-list-config.json):
{
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
}

Notes:
- The address_family must be either "IPv4" or "IPv6".
- The entries must be a valid JSON array of prefix filter entries.
- The ge and le values are optional but must be within the range of the prefix length.
`,
	Args: cobra.ExactArgs(1),
	RunE: WrapRunE(CreateMCRPrefixFilterList),
}

// listMCRPrefixFilterListsCmd lists all prefix filter lists for a specific MCR.
var listMCRPrefixFilterListsCmd = &cobra.Command{
	Use:   "list-prefix-filter-lists [mcrUID]",
	Short: "List all prefix filter lists for a specific MCR",
	Long: `List all prefix filter lists for a specific MCR.

This command retrieves and displays a list of all prefix filter lists configured
on the specified MCR.

Example usage:
  megaport-cli mcr list-prefix-filter-lists [mcrUID]
`,
	Args: cobra.ExactArgs(1),
	RunE: WrapRunE(ListMCRPrefixFilterLists),
}

// getMCRPrefixFilterListCmd retrieves and displays details for a single prefix filter list on an MCR.
var getMCRPrefixFilterListCmd = &cobra.Command{
	Use:   "get-prefix-filter-list [mcrUID] [prefixFilterListID]",
	Short: "Get details for a single prefix filter list on an MCR",
	Long: `Get details for a single prefix filter list on an MCR.

This command retrieves and displays detailed information about a specific prefix
filter list on the specified MCR.

Example usage:
  megaport-cli mcr get-prefix-filter-list [mcrUID] [prefixFilterListID]
`,
	Args: cobra.ExactArgs(2),
	RunE: WrapRunE(GetMCRPrefixFilterList),
}

// updateMCRPrefixFilterListCmd updates a prefix filter list on an MCR.
var updateMCRPrefixFilterListCmd = &cobra.Command{
	Use:   "update-prefix-filter-list [mcrUID] [prefixFilterListID]",
	Short: "Update a prefix filter list on an MCR",
	Long: `Update a prefix filter list on an MCR.

This command allows you to update the details of an existing prefix filter list on an MCR.
You can use this command to modify the description, address family, or entries in the list.

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
  - description: The new description of the prefix filter list (1-255 characters).
  - address_family: The new address family (IPv4 or IPv6).
  - entries: JSON array of prefix filter entries. Each entry has:
      - action: "permit" or "deny"
      - prefix: CIDR notation (e.g., "192.168.0.0/16")
      - ge (optional): Greater than or equal to value (must be less than or equal to the prefix length)
      - le (optional): Less than or equal to value (must be greater than or equal to the prefix length)

Example usage:

  # Interactive mode
  megaport-cli mcr update-prefix-filter-list [mcrUID] [prefixFilterListID] --interactive

  # Flag mode
  megaport-cli mcr update-prefix-filter-list [mcrUID] [prefixFilterListID] --description "Updated prefix list" --entries '[{"action":"permit","prefix":"10.0.0.0/8","ge":24,"le":32}]'

  # JSON mode
  megaport-cli mcr update-prefix-filter-list [mcrUID] [prefixFilterListID] --json '{"description":"Updated prefix list","entries":[{"action":"permit","prefix":"10.0.0.0/8","ge":24,"le":32}]}'
  megaport-cli mcr update-prefix-filter-list [mcrUID] [prefixFilterListID] --json-file ./update-prefix-list.json

JSON format example (update-prefix-list.json):
{
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
}

Notes:
- The address_family must be either "IPv4" or "IPv6".
- The entries must be a valid JSON array of prefix filter entries.
- The ge and le values are optional but must be within the range of the prefix length.
`,
	Args: cobra.ExactArgs(2),
	RunE: WrapRunE(UpdateMCRPrefixFilterList),
}

// deleteMCRPrefixFilterListCmd deletes a prefix filter list on an MCR.
var deleteMCRPrefixFilterListCmd = &cobra.Command{
	Use:   "delete-prefix-filter-list [mcrUID] [prefixFilterListID]",
	Short: "Delete a prefix filter list on an MCR",
	Long: `Delete a prefix filter list on an MCR.

This command allows you to delete a prefix filter list from the specified MCR.

Example usage:
  megaport-cli mcr delete-prefix-filter-list [mcrUID] [prefixFilterListID]
`,
	Args: cobra.ExactArgs(2),
	RunE: WrapRunE(DeleteMCRPrefixFilterList),
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

	listMCRsCmd.Flags().Bool("inactive", false, "Include inactive MCRs in the list")

	mcrCmd.AddCommand(getMCRCmd)
	mcrCmd.AddCommand(listMCRsCmd)
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
