package cmd

import (
	"github.com/spf13/cobra"
)

var (
	productName   string
	connectType   string
	companyName   string
	diversityZone string
)

// partnersCmd is the base command for all operations related to partner ports in the Megaport API.
var partnersCmd = &cobra.Command{
	Use:   "partners",
	Short: "Manage partner ports in the Megaport API",
}

// listPartnersCmd lists all available partner ports and applies filters based on the provided flags.
var listPartnersCmd = &cobra.Command{
	Use:   "list",
	Short: "List all partner ports",
	RunE:  WrapRunE(ListPartners),
}

// findPartnersCmd allows users to interactively search for partner ports.
var findPartnersCmd = &cobra.Command{
	Use:   "find",
	Short: "Find partner ports interactively",
	RunE:  WrapRunE(FindPartners),
}

func init() {
	// Add filters to list command
	listPartnersCmd.Flags().StringVar(&productName, "product-name", "", "Filter by Product Name")
	listPartnersCmd.Flags().StringVar(&connectType, "connect-type", "", "Filter by Connect Type")
	listPartnersCmd.Flags().StringVar(&companyName, "company-name", "", "Filter by Company Name")
	listPartnersCmd.Flags().IntVar(&locationID, "location-id", 0, "Filter by Location ID")
	listPartnersCmd.Flags().StringVar(&diversityZone, "diversity-zone", "", "Filter by Diversity Zone")

	// Add subcommands to partners command
	partnersCmd.AddCommand(listPartnersCmd)
	partnersCmd.AddCommand(findPartnersCmd)

	// Set up help builders for commands

	// partners command help
	partnersHelp := &CommandHelpBuilder{
		CommandName: "megaport-cli partners",
		ShortDesc:   "Manage partner ports in the Megaport API",
		LongDesc:    "Manage partner ports in the Megaport API.\n\nThis command groups all operations related to partner ports. You can use its subcommands to list and filter available partner ports based on specific criteria.",
		Examples: []string{
			"partners find",
			"partners list",
			"partners list --product-name \"AWS Partner Port\" --company-name \"AWS\" --location-id 1",
		},
	}
	partnersCmd.Long = partnersHelp.Build()

	// list partners help
	listPartnersHelp := &CommandHelpBuilder{
		CommandName: "megaport-cli partners list",
		ShortDesc:   "List all partner ports",
		LongDesc:    "List all partner ports available in the Megaport API.\n\nThis command fetches and displays a list of all available partner ports. You can filter the partner ports based on specific criteria.",
		OptionalFlags: map[string]string{
			"product-name":   "Filter partner ports by product name",
			"connect-type":   "Filter partner ports by connect type",
			"company-name":   "Filter partner ports by company name",
			"location-id":    "Filter partner ports by location ID",
			"diversity-zone": "Filter partner ports by diversity zone",
		},
		Examples: []string{
			"list",
			"list --product-name \"AWS Partner Port\"",
			"list --connect-type \"Dedicated Cloud Connection\"",
			"list --company-name \"Amazon Web Services\"",
			"list --location-id 1",
			"list --diversity-zone \"Zone A\"",
		},
		ImportantNotes: []string{
			"The list can be filtered by multiple criteria at once",
			"Filtering is case-insensitive and partial matches are supported",
		},
	}
	listPartnersCmd.Long = listPartnersHelp.Build()

	// find partners help
	findPartnersHelp := &CommandHelpBuilder{
		CommandName: "megaport-cli partners find",
		ShortDesc:   "Find partner ports interactively",
		LongDesc:    "Find partner ports using an interactive search with optional filters.\n\nThis command launches an interactive session to help you find partner ports. You'll be prompted for various search criteria, but all prompts are optional. Simply press Enter to skip any filter you don't want to apply.",
		ImportantNotes: []string{
			"You can skip any prompt by pressing Enter without typing anything",
			"The search is performed after all filters are collected",
			"Results are displayed in a tabular format",
		},
		Examples: []string{
			"find",
		},
	}
	findPartnersCmd.Long = findPartnersHelp.Build()

	rootCmd.AddCommand(partnersCmd)
}
