package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	megaport "github.com/megaport/megaportgo"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var (
	productName   string
	connectType   string
	companyName   string
	diversityZone string
)

var partnersCmd = &cobra.Command{
	Use:   "partners",
	Short: "Manage partner ports in the Megaport API",
	Long:  `Manage partner ports in the Megaport API.`,
}

var listPartnersCmd = &cobra.Command{
	Use:   "list",
	Short: "List all partner ports",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		client, err := Login(ctx)
		if err != nil {
			return fmt.Errorf("error logging in: %v", err)
		}

		// Retrieve partner ports from the API
		partners, err := client.PartnerService.ListPartnerMegaports(ctx)
		if err != nil {
			return fmt.Errorf("error listing partner ports: %v", err)
		}

		// Perform in-memory filtering
		filteredPartners := filterPartners(
			partners,
			productName,
			connectType,
			companyName,
			locationID,
			diversityZone,
		)
		printPartners(filteredPartners, outputFormat)
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
	ProductName   string `json:"product_name"`
	ConnectType   string `json:"connect_type"`
	CompanyName   string `json:"company_name"`
	LocationID    int    `json:"location_id"`
	DiversityZone string `json:"diversity_zone"`
	VXCPermitted  bool   `json:"vxc_permitted"`
}

// ToPartnerOutput converts a PartnerMegaport to a PartnerOutput.
func ToPartnerOutput(p *megaport.PartnerMegaport) *PartnerOutput {
	return &PartnerOutput{
		ProductName:   p.ProductName,
		ConnectType:   p.ConnectType,
		CompanyName:   p.CompanyName,
		LocationID:    p.LocationId,
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
func printPartners(partners []*megaport.PartnerMegaport, format string) {
	switch format {
	case "json":
		var outputList []*PartnerOutput
		for _, partner := range partners {
			outputList = append(outputList, ToPartnerOutput(partner))
		}
		printed, err := json.Marshal(outputList)
		if err != nil {
			fmt.Println("Error printing partner ports:", err)
			os.Exit(1)
		}
		fmt.Println(string(printed))
	case "table":
		table := tablewriter.NewWriter(os.Stdout)

		// Disable automatic header formatting so that the headers won't become uppercase.
		table.SetAutoFormatHeaders(false)

		table.SetHeader([]string{
			"ProductName",
			"ConnectType",
			"CompanyName",
			"LocationID",
			"DiversityZone",
			"VXCPermitted",
		})

		for _, partner := range partners {
			table.Append([]string{
				partner.ProductName,
				partner.ConnectType,
				partner.CompanyName,
				fmt.Sprintf("%d", partner.LocationId),
				partner.DiversityZone,
				fmt.Sprintf("%t", partner.VXCPermitted),
			})
		}
		table.Render()
	default:
		fmt.Println("Invalid output format. Use 'json' or 'table'")
	}
}
