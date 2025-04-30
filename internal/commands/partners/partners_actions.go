package partners

import (
	"context"
	"fmt"
	"time"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/commands/config"
	"github.com/megaport/megaport-cli/internal/utils"
	"github.com/spf13/cobra"
)

func ListPartners(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("error logging in: %v", err)
	}

	spinner := output.PrintResourceListing("partner", noColor)
	partners, err := client.PartnerService.ListPartnerMegaports(ctx)
	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to list partner ports: %v", noColor, err)
		return fmt.Errorf("error listing partners: %v", err)
	}

	productName, _ := cmd.Flags().GetString("product-name")
	connectType, _ := cmd.Flags().GetString("connect-type")
	companyName, _ := cmd.Flags().GetString("company-name")
	locationID, _ := cmd.Flags().GetInt("location-id")
	diversityZone, _ := cmd.Flags().GetString("diversity-zone")

	filteredPartners := filterPartners(partners, productName, connectType, companyName, locationID, diversityZone)

	if len(filteredPartners) == 0 {
		output.PrintWarning("No partner ports found matching the specified filters", noColor)
	} else {
		output.PrintInfo("Found %d partner ports matching the specified filters", noColor, len(filteredPartners))
	}

	err = printPartnersFunc(filteredPartners, outputFormat, noColor)
	if err != nil {
		output.PrintError("Failed to output.Print partner ports: %v", noColor, err)
		return fmt.Errorf("error output.Printing partners: %v", err)
	}
	return nil
}

func FindPartners(cmd *cobra.Command, args []string, noColor bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	output.PrintInfo("Searching for partner ports...", noColor)

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("error logging in: %v", err)
	}

	spinner := output.PrintResourceListing("partner", noColor)
	partners, err := client.PartnerService.ListPartnerMegaports(ctx)
	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to list partner ports: %v", noColor, err)
		return fmt.Errorf("error listing partners: %v", err)
	}

	output.PrintInfo("Filter partner ports - press Enter to skip any filter", noColor)
	output.PrintInfo("----------------------------------------------------", noColor)

	productName, err := utils.Prompt("Product name: ", noColor)
	if err != nil {
		output.PrintError("Failed to get product name: %v", noColor, err)
		return err
	}

	connectType, err := utils.Prompt("Connect type (AWS, AWSHC, AZURE, GOOGLE, ORACLE, IBM, etc): ", noColor)
	if err != nil {
		output.PrintError("Failed to get connect type: %v", noColor, err)
		return err
	}

	companyName, err := utils.Prompt("Company name: ", noColor)
	if err != nil {
		output.PrintError("Failed to get company name: %v", noColor, err)
		return err
	}

	var locationID int
	locationIDStr, err := utils.Prompt("Location ID (numeric): ", noColor)
	if err != nil {
		output.PrintError("Failed to get location ID: %v", noColor, err)
		return err
	}
	if locationIDStr != "" {
		if _, err := fmt.Sscanf(locationIDStr, "%d", &locationID); err != nil {
			output.PrintError("Invalid location ID format: %v", noColor, err)
			return fmt.Errorf("invalid location ID format: %v", err)
		}
	}

	diversityZone, err := utils.Prompt("Diversity zone: ", noColor)
	if err != nil {
		output.PrintError("Failed to get diversity zone: %v", noColor, err)
		return err
	}

	format, err := utils.Prompt("Output format [table/json] (default: table): ", noColor)
	if err != nil {
		output.PrintError("Failed to get output format: %v", noColor, err)
		return err
	}

	selectedFormat := "table"
	if format == "json" {
		selectedFormat = "json"
	}

	filteredPartners := filterPartners(partners, productName, connectType, companyName, locationID, diversityZone)

	output.PrintInfo("Found %d matching partner ports", noColor, len(filteredPartners))

	err = printPartnersFunc(filteredPartners, selectedFormat, noColor)
	if err != nil {
		output.PrintError("Failed to output.Print partner ports: %v", noColor, err)
		return fmt.Errorf("error output.Printing partners: %v", err)
	}
	return nil
}
