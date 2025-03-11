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
// This command serves as a container for subcommands which allow you to list and filter partner ports
// based on various criteria such as product name, connect type, company name, location ID, and diversity zone.
//
// Example usage:
//
//	megaport partners list
var partnersCmd = &cobra.Command{
	Use:   "partners",
	Short: "Manage partner ports in the Megaport API",
	Long: `Manage partner ports in the Megaport API.

This command groups all operations related to partner ports. You can use its subcommands 
to list and filter available partner ports based on specific criteria.

Examples:
  megaport partners list
  megaport partners list --product-name "AWS Partner Port" --company-name "AWS" --location-id 1
`,
}

// listPartnersCmd lists all available partner ports and applies filters based on the provided flags.
// The filtering criteria include product name, connect type, company name, location ID, and diversity zone.
// The results are printed in the output format specified by the global flag (either JSON or table).
//
// Example usage:
//
//	megaport partners list --product-name "Enterprise" --connect-type "Fiber" --company-name "Acme Corp" --location-id 2 --diversity-zone "ZoneA"
var listPartnersCmd = &cobra.Command{
	Use:   "list",
	Short: "List all partner ports",
	Long: `List all partner ports available in the Megaport API.

This command fetches and displays a list of all available partner ports with details such as
product name, connect type, company name, location ID, and diversity zone. You can also filter
the partner ports based on specific criteria.

Available filters:
  - product-name: Filter partner ports by product name.
  - connect-type: Filter partner ports by connect type.
  - company-name: Filter partner ports by company name.
  - location-id: Filter partner ports by location ID.
  - diversity-zone: Filter partner ports by diversity zone.

Example usage:

  megaport partners list
  megaport partners list --product-name "AWS Partner Port"
  megaport partners list --connect-type "AWS"
  megaport partners list --company-name "AWS"
  megaport partners list --location-id 67
  megaport partners list --diversity-zone "blue"
`,
	RunE: WrapRunE(ListPartners),
}

// Add this new command definition after listPartnersCmd
var findPartnersCmd = &cobra.Command{
	Use:   "find",
	Short: "Find partner ports interactively",
	Long: `Find partner ports using an interactive search with optional filters.

This command launches an interactive session to help you find partner ports.
You'll be prompted for various search criteria, but all prompts are optional.
Simply press Enter to skip any filter you don't want to apply.

Available filters:
  - Product name
  - Connect type
  - Company name
  - Location ID
  - Diversity zone

Example usage:
  megaport partners find
`,
	RunE: WrapRunE(FindPartners),
}

func init() {
	listPartnersCmd.Flags().StringVar(&productName, "product-name", "", "Filter by Product Name")
	listPartnersCmd.Flags().StringVar(&connectType, "connect-type", "", "Filter by Connect Type")
	listPartnersCmd.Flags().StringVar(&companyName, "company-name", "", "Filter by Company Name")
	listPartnersCmd.Flags().IntVar(&locationID, "location-id", 0, "Filter by Location ID")
	listPartnersCmd.Flags().StringVar(&diversityZone, "diversity-zone", "", "Filter by Diversity Zone")
	partnersCmd.AddCommand(findPartnersCmd)
	partnersCmd.AddCommand(listPartnersCmd)
	rootCmd.AddCommand(partnersCmd)
}
