package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	megaport "github.com/megaport/megaportgo"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create a context with a 30-second timeout.
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Login to the API.
		client, err := Login(ctx)
		if err != nil {
			return fmt.Errorf("error logging in: %v", err)
		}

		// Retrieve partner ports from the API.
		partners, err := client.PartnerService.ListPartnerMegaports(ctx)
		if err != nil {
			return fmt.Errorf("error listing partner ports: %v", err)
		}

		// Perform in-memory filtering using the provided criteria.
		filteredPartners := filterPartners(
			partners,
			productName,
			connectType,
			companyName,
			locationID,
			diversityZone,
		)
		err = printPartners(filteredPartners, outputFormat)
		if err != nil {
			return fmt.Errorf("error printing partner ports: %v", err)
		}
		return nil
	},
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

// PartnerOutput represents the desired fields for JSON output.
type PartnerOutput struct {
	output
	ProductName   string `json:"product_name"`
	ConnectType   string `json:"connect_type"`
	CompanyName   string `json:"company_name"`
	LocationId    int    `json:"location_id"`
	DiversityZone string `json:"diversity_zone"`
	VXCPermitted  bool   `json:"vxc_permitted"`
}

// ToPartnerOutput converts a PartnerMegaport to a PartnerOutput.
func ToPartnerOutput(p *megaport.PartnerMegaport) PartnerOutput {
	return PartnerOutput{
		ProductName:   p.ProductName,
		ConnectType:   p.ConnectType,
		CompanyName:   p.CompanyName,
		LocationId:    p.LocationId,
		DiversityZone: p.DiversityZone,
		VXCPermitted:  p.VXCPermitted,
	}
}

// filterPartners applies basic in-memory filters to a list of partner ports.
func filterPartners(
	partners []*megaport.PartnerMegaport,
	productName, connectType, companyName string,
	locationID int,
	diversityZone string,
) []*megaport.PartnerMegaport {
	var filtered []*megaport.PartnerMegaport
	for _, partner := range partners {
		if productName != "" && !strings.EqualFold(partner.ProductName, productName) {
			continue
		}
		if connectType != "" && !strings.EqualFold(partner.ConnectType, connectType) {
			continue
		}
		if companyName != "" && !strings.EqualFold(partner.CompanyName, companyName) {
			continue
		}
		if locationID != 0 && partner.LocationId != locationID {
			continue
		}
		if diversityZone != "" && !strings.EqualFold(partner.DiversityZone, diversityZone) {
			continue
		}
		filtered = append(filtered, partner)
	}
	return filtered
}

// printPartners prints the partner ports in the specified output format.
func printPartners(partners []*megaport.PartnerMegaport, format string) error {
	// Convert partners to output format
	outputs := make([]PartnerOutput, 0, len(partners))
	for _, partner := range partners {
		outputs = append(outputs, ToPartnerOutput(partner))
	}

	// Use generic printOutput function
	return printOutput(outputs, format)
}
