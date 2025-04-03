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
	Long: `Manage partner ports in the Megaport API.

This command groups all operations related to partner ports. You can use its subcommands 
to list and filter available partner ports based on specific criteria.

Examples:
  megaport-cli partners find # interactive mode
  megaport-cli partners list
  megaport-cli partners list --product-name "AWS Partner Port" --company-name "AWS" --location-id 1
`,
}

// listPartnersCmd lists all available partner ports and applies filters based on the provided flags.
var listPartnersCmd = &cobra.Command{
	Use:   "list",
	Short: "List all partner ports",
	Long: `List all partner ports available in the Megaport API.

This command fetches and displays a list of all available partner ports. You can filter
the partner ports based on specific criteria.

Available filters:
  - product-name: Filter partner ports by product name.
  - connect-type: Filter partner ports by connect type.
  - company-name: Filter partner ports by company name.
  - location-id: Filter partner ports by location ID.
  - diversity-zone: Filter partner ports by diversity zone.

Example usage:

  megaport-cli partners list
  megaport-cli partners list --product-name "AWS Partner Port"
  megaport-cli partners list --connect-type "Dedicated Cloud Connection"
  megaport-cli partners list --company-name "Amazon Web Services"
  megaport-cli partners list --location-id 1
  megaport-cli partners list --diversity-zone "Zone A"

Example output:
  Product Name        Connect Type              Company Name          Location ID  Diversity Zone
  ------------------  ------------------------  --------------------  -----------  --------------
  AWS Partner Port    Dedicated Cloud Connect   Amazon Web Services             1  Zone A
`,
	RunE: WrapRunE(ListPartners),
}

// findPartnersCmd allows users to interactively search for partner ports.
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
  megaport-cli partners find
`,
	RunE: WrapRunE(FindPartners),
}

func init() {
	listPartnersCmd.Flags().StringVar(&productName, "product-name", "", "Filter by Product Name")
	listPartnersCmd.Flags().StringVar(&connectType, "connect-type", "", "Filter by Connect Type")
	listPartnersCmd.Flags().StringVar(&companyName, "company-name", "", "Filter by Company Name")
	listPartnersCmd.Flags().IntVar(&locationID, "location-id", 0, "Filter by Location ID")
	listPartnersCmd.Flags().StringVar(&diversityZone, "diversity-zone", "", "Filter by Diversity Zone")

	partnersCmd.AddCommand(listPartnersCmd)
	partnersCmd.AddCommand(findPartnersCmd)
	rootCmd.AddCommand(partnersCmd)
}
