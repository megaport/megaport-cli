package cmd

import (
	"context"
	"fmt"
	"time"

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
  megaport partners list --product-name "Enterprise" --company-name "Acme Corp" --location-id 1
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
  megaport partners list --product-name "Enterprise"
  megaport partners list --connect-type "Fiber"
  megaport partners list --company-name "Acme Corp"
  megaport partners list --location-id 2
  megaport partners list --diversity-zone "ZoneA"
`,
	RunE: WrapRunE(ListPartners),
}

func ListPartners(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %v", err)
	}

	partners, err := client.PartnerService.ListPartnerMegaports(ctx)
	if err != nil {
		return fmt.Errorf("error listing partners: %v", err)
	}

	// Get filter values from flags
	productName, _ := cmd.Flags().GetString("product-name")
	connectType, _ := cmd.Flags().GetString("connect-type")
	companyName, _ := cmd.Flags().GetString("company-name")
	locationID, _ := cmd.Flags().GetInt("location-id")
	diversityZone, _ := cmd.Flags().GetString("diversity-zone")

	// Apply filters
	filteredPartners := filterPartners(partners, productName, connectType, companyName, locationID, diversityZone)

	// Print partners with current output format
	return printPartners(filteredPartners, outputFormat)
}

func init() {
	listPartnersCmd.Flags().StringVar(&productName, "product-name", "", "Filter by Product Name")
	listPartnersCmd.Flags().StringVar(&connectType, "connect-type", "", "Filter by Connect Type")
	listPartnersCmd.Flags().StringVar(&companyName, "company-name", "", "Filter by Company Name")
	listPartnersCmd.Flags().IntVar(&locationID, "location-id", 0, "Filter by Location ID")
	listPartnersCmd.Flags().StringVar(&diversityZone, "diversity-zone", "", "Filter by Diversity Zone")
	partnersCmd.AddCommand(listPartnersCmd)
	rootCmd.AddCommand(partnersCmd)
}
