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
You will be prompted to enter the required and optional fields.

Required fields:
  - name: The name of the MCR.
  - term: The term of the MCR (1, 12, 24, or 36 months).
  - port_speed: The speed of the MCR (1000, 2500, 5000, or 10000 Mbps).
  - location_id: The ID of the location where the MCR will be provisioned.

Optional fields:
  - diversity_zone: The diversity zone for the MCR.
  - cost_center: The cost center for the MCR.
  - promo_code: A promotional code for the MCR.

Example usage:

  megaport mcr buy
`,
	RunE: WrapRunE(BuyMCR),
}

// updateMCRCmd updates an existing Megaport Cloud Router (MCR).
var updateMCRCmd = &cobra.Command{
	Use:   "update [mcrUID]",
	Short: "Update an existing MCR",
	Args:  cobra.ExactArgs(1),
	RunE:  WrapRunE(UpdateMCR),
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
	Args:  cobra.ExactArgs(1),
	RunE:  WrapRunE(CreateMCRPrefixFilterList),
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
	Args:  cobra.ExactArgs(2),
	RunE:  WrapRunE(UpdateMCRPrefixFilterList),
}

// deleteMCRPrefixFilterListCmd deletes a prefix filter list on an MCR.
var deleteMCRPrefixFilterListCmd = &cobra.Command{
	Use:   "delete-prefix-filter-list [mcrUID] [prefixFilterListID]",
	Short: "Delete a prefix filter list on an MCR",
	Args:  cobra.ExactArgs(2),
	RunE:  WrapRunE(DeleteMCRPrefixFilterList),
}

func init() {
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
