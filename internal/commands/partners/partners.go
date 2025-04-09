package partners

import (
	"github.com/megaport/megaport-cli/internal/base/cmdbuilder"
	"github.com/spf13/cobra"
)

// AddCommandsTo builds the partners commands and adds them to the root command
func AddCommandsTo(rootCmd *cobra.Command) {
	// Create partners parent command
	partnersCmd := cmdbuilder.NewCommand("partners", "Manage partner ports in the Megaport API").
		WithLongDesc("Manage partner ports in the Megaport API.\n\nThis command groups all operations related to partner ports. You can use its subcommands to list and filter available partner ports based on specific criteria.").
		WithExample("partners find").
		WithExample("partners list").
		WithExample("partners list --product-name \"AWS Partner Port\" --company-name \"AWS\" --location-id 1").
		WithRootCmd(rootCmd).
		Build()

	// Create list partners command
	listPartnersCmd := cmdbuilder.NewCommand("list", "List all partner ports").
		WithLongDesc("List all partner ports available in the Megaport API.\n\nThis command fetches and displays a list of all available partner ports. You can filter the partner ports based on specific criteria.").
		WithOutputFormatRunFunc(ListPartners).
		WithFlag("product-name", "", "Filter partner ports by product name").
		WithFlag("connect-type", "", "Filter partner ports by connect type").
		WithFlag("company-name", "", "Filter partner ports by company name").
		WithIntFlag("location-id", 0, "Filter partner ports by location ID").
		WithFlag("diversity-zone", "", "Filter partner ports by diversity zone").
		WithOptionalFlag("product-name", "Filter partner ports by product name").
		WithOptionalFlag("connect-type", "Filter partner ports by connect type").
		WithOptionalFlag("company-name", "Filter partner ports by company name").
		WithOptionalFlag("location-id", "Filter partner ports by location ID").
		WithOptionalFlag("diversity-zone", "Filter partner ports by diversity zone").
		WithExample("list").
		WithExample("list --product-name \"AWS Partner Port\"").
		WithExample("list --connect-type \"Dedicated Cloud Connection\"").
		WithExample("list --company-name \"Amazon Web Services\"").
		WithExample("list --location-id 1").
		WithExample("list --diversity-zone \"blue\"").
		WithImportantNote("The list can be filtered by multiple criteria at once").
		WithImportantNote("Filtering is case-insensitive and partial matches are supported").
		WithRootCmd(rootCmd).
		Build()

	// Create find partners command
	findPartnersCmd := cmdbuilder.NewCommand("find", "Find partner ports interactively").
		WithLongDesc("Find partner ports using an interactive search with optional filters.\n\nThis command launches an interactive session to help you find partner ports. You'll be prompted for various search criteria, but all prompts are optional. Simply press Enter to skip any filter you don't want to apply.").
		WithColorAwareRunFunc(FindPartners).
		WithExample("find").
		WithImportantNote("You can skip any prompt by pressing Enter without typing anything").
		WithImportantNote("The search is performed after all filters are collected").
		WithImportantNote("Results are displayed in a tabular format").
		WithRootCmd(rootCmd).
		Build()

	// Add commands to their parents
	partnersCmd.AddCommand(listPartnersCmd, findPartnersCmd)
	rootCmd.AddCommand(partnersCmd)
}
