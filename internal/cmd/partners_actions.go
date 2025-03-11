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
	return printPartnersFunc(filteredPartners, outputFormat)
}

// FindPartners implements the interactive search functionality for partner ports.
func FindPartners(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println("Searching for partner ports...")

	client, err := Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %v", err)
	}

	// Get all partners first (we'll filter them based on interactive inputs)
	partners, err := client.PartnerService.ListPartnerMegaports(ctx)
	if err != nil {
		return fmt.Errorf("error listing partners: %v", err)
	}

	// Collect filters interactively
	fmt.Println("Filter partner ports - press Enter to skip any filter")
	fmt.Println("----------------------------------------------------")

	productName, err := prompt("Product name: ")
	if err != nil {
		return err
	}

	connectType, err := prompt("Connect type (AWS, AWSHC, AZURE, GOOGLE, ORACLE, IBM, etc): ")
	if err != nil {
		return err
	}

	companyName, err := prompt("Company name: ")
	if err != nil {
		return err
	}

	// Handle numeric input for location ID
	var locationID int
	locationIDStr, err := prompt("Location ID (numeric): ")
	if err != nil {
		return err
	}
	if locationIDStr != "" {
		if _, err := fmt.Sscanf(locationIDStr, "%d", &locationID); err != nil {
			return fmt.Errorf("invalid location ID format: %v", err)
		}
	}

	diversityZone, err := prompt("Diversity zone: ")
	if err != nil {
		return err
	}

	// Prompt for output format
	format, err := prompt("Output format [table/json] (default: table): ")
	if err != nil {
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
	fmt.Printf("\nFound %d matching partner ports\n\n", len(filteredPartners))

	// Print partners with selected output format
	return printPartnersFunc(filteredPartners, selectedFormat)
}
