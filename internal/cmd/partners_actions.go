package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

func ListPartners(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := Login(ctx)
	if err != nil {
		PrintError("Failed to log in: %v", err)
		return fmt.Errorf("error logging in: %v", err)
	}

	PrintInfo("Retrieving partner ports...")
	partners, err := client.PartnerService.ListPartnerMegaports(ctx)
	if err != nil {
		PrintError("Failed to list partner ports: %v", err)
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

	if len(filteredPartners) == 0 {
		PrintWarning("No partner ports found matching the specified filters")
	} else {
		PrintInfo("Found %d partner ports matching the specified filters", len(filteredPartners))
	}

	// Print partners with current output format
	err = printPartnersFunc(filteredPartners, outputFormat)
	if err != nil {
		PrintError("Failed to print partner ports: %v", err)
		return fmt.Errorf("error printing partners: %v", err)
	}
	return nil
}

// FindPartners implements the interactive search functionality for partner ports.
func FindPartners(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	PrintInfo("Searching for partner ports...")

	client, err := Login(ctx)
	if err != nil {
		PrintError("Failed to log in: %v", err)
		return fmt.Errorf("error logging in: %v", err)
	}

	// Get all partners first (we'll filter them based on interactive inputs)
	PrintInfo("Retrieving all partner ports...")
	partners, err := client.PartnerService.ListPartnerMegaports(ctx)
	if err != nil {
		PrintError("Failed to list partner ports: %v", err)
		return fmt.Errorf("error listing partners: %v", err)
	}

	// Collect filters interactively
	PrintInfo("Filter partner ports - press Enter to skip any filter")
	PrintInfo("----------------------------------------------------")

	productName, err := prompt("Product name: ")
	if err != nil {
		PrintError("Failed to get product name: %v", err)
		return err
	}

	connectType, err := prompt("Connect type (AWS, AWSHC, AZURE, GOOGLE, ORACLE, IBM, etc): ")
	if err != nil {
		PrintError("Failed to get connect type: %v", err)
		return err
	}

	companyName, err := prompt("Company name: ")
	if err != nil {
		PrintError("Failed to get company name: %v", err)
		return err
	}

	// Handle numeric input for location ID
	var locationID int
	locationIDStr, err := prompt("Location ID (numeric): ")
	if err != nil {
		PrintError("Failed to get location ID: %v", err)
		return err
	}
	if locationIDStr != "" {
		if _, err := fmt.Sscanf(locationIDStr, "%d", &locationID); err != nil {
			PrintError("Invalid location ID format: %v", err)
			return fmt.Errorf("invalid location ID format: %v", err)
		}
	}

	diversityZone, err := prompt("Diversity zone: ")
	if err != nil {
		PrintError("Failed to get diversity zone: %v", err)
		return err
	}

	// Prompt for output format
	format, err := prompt("Output format [table/json] (default: table): ")
	if err != nil {
		PrintError("Failed to get output format: %v", err)
		return err
	}

	// Set output format - default to table if not specified
	selectedFormat := "table"
	if format == "json" {
		selectedFormat = "json"
	}

	// Apply filters
	filteredPartners := filterPartners(partners, productName, connectType, companyName, locationID, diversityZone)

	// Show count of results
	PrintInfo("Found %d matching partner ports", len(filteredPartners))

	// Print partners with selected output format
	err = printPartnersFunc(filteredPartners, selectedFormat)
	if err != nil {
		PrintError("Failed to print partner ports: %v", err)
		return fmt.Errorf("error printing partners: %v", err)
	}
	return nil
}
