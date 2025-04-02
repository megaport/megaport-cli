package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

// BuyMVE handles the purchase of a new Megaport Virtual Edge device
func BuyMVE(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Determine which mode to use
	interactive, _ := cmd.Flags().GetBool("interactive")
	jsonStr, _ := cmd.Flags().GetString("json")
	jsonFile, _ := cmd.Flags().GetString("json-file")

	// Check if any flag-based parameters are provided
	flagsProvided := cmd.Flags().Changed("name") ||
		cmd.Flags().Changed("term") ||
		cmd.Flags().Changed("location-id") ||
		cmd.Flags().Changed("vendor-config") ||
		cmd.Flags().Changed("vnics")

	var req *megaport.BuyMVERequest
	var err error

	// Process input based on mode priority: JSON > Flags > Interactive
	if jsonStr != "" || jsonFile != "" {
		// JSON mode
		req, err = processJSONBuyMVEInput(jsonStr, jsonFile)
		if err != nil {
			return err
		}
	} else if flagsProvided {
		// Flag mode
		req, err = processFlagBuyMVEInput(cmd)
		if err != nil {
			return err
		}
	} else if interactive {
		// Interactive mode
		req, err = promptForBuyMVEDetails()
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("no input provided, use --interactive, --json, or flags to specify MVE details")
	}

	// Call the API to buy the MVE
	client, err := Login(ctx)
	if err != nil {
		return err
	}

	fmt.Println("Validating MVE order...")
	if err := client.MVEService.ValidateMVEOrder(ctx, req); err != nil {
		return fmt.Errorf("validation failed: %v", err)
	}

	fmt.Println("Buying MVE...")
	resp, err := client.MVEService.BuyMVE(ctx, req)
	if err != nil {
		return err
	}

	fmt.Printf("MVE purchased successfully - UID: %s\n", resp.TechnicalServiceUID)
	return nil
}

func UpdateMVE(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Retrieve the MVE UID from command line arguments
	mveUID := args[0]

	// Get the original MVE to compare values later
	client, err := Login(ctx)
	if err != nil {
		return err
	}

	// Fetch original MVE details before update
	originalMVE, err := client.MVEService.GetMVE(ctx, mveUID)
	if err != nil {
		return fmt.Errorf("error retrieving MVE details: %v", err)
	}

	// Determine which mode to use
	interactive, _ := cmd.Flags().GetBool("interactive")
	jsonStr, _ := cmd.Flags().GetString("json")
	jsonFile, _ := cmd.Flags().GetString("json-file")

	// Check if any flag-based parameters are provided
	flagsProvided := cmd.Flags().Changed("name") ||
		cmd.Flags().Changed("cost-centre") ||
		cmd.Flags().Changed("contract-term")

	var req *megaport.ModifyMVERequest

	// Process input based on mode priority: JSON > Flags > Interactive
	if jsonStr != "" || jsonFile != "" {
		// JSON mode
		req, err = processJSONUpdateMVEInput(jsonStr, jsonFile, mveUID)
		if err != nil {
			return err
		}
	} else if flagsProvided {
		// Flag mode
		req, err = processFlagUpdateMVEInput(cmd, mveUID)
		if err != nil {
			return err
		}
	} else if interactive {
		// Interactive mode
		req, err = promptForUpdateMVEDetails(mveUID)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("no input provided, use --interactive, --json, or flags to specify MVE update details")
	}

	// Set common defaults
	req.WaitForUpdate = true
	req.WaitForTime = 10 * time.Minute

	// Call the ModifyMVE method
	fmt.Println("Updating MVE...")
	resp, err := client.MVEService.ModifyMVE(ctx, req)
	if err != nil {
		return fmt.Errorf("error updating MVE: %v", err)
	}

	if !resp.MVEUpdated {
		fmt.Println("MVE update request was not successful")
		return nil
	}

	// Fetch the updated MVE to get the new values
	updatedMVE, err := client.MVEService.GetMVE(ctx, mveUID)
	if err != nil {
		return fmt.Errorf("error retrieving updated MVE details: %v", err)
	}

	// Print detailed success message
	fmt.Println("MVE updated successfully:")
	fmt.Printf("UID:          %s\n", mveUID)

	// Compare and show name changes
	if originalMVE.Name != updatedMVE.Name {
		fmt.Printf("Name:         %s (previously \"%s\")\n", updatedMVE.Name, originalMVE.Name)
	} else {
		fmt.Printf("Name:         %s (unchanged)\n", updatedMVE.Name)
	}

	// Compare and show cost centre changes
	if originalMVE.CostCentre != updatedMVE.CostCentre {
		// Handle empty cost centre specially
		origCC := originalMVE.CostCentre
		if origCC == "" {
			origCC = "none"
		}
		fmt.Printf("Cost Centre:  %s (previously \"%s\")\n", updatedMVE.CostCentre, origCC)
	} else if updatedMVE.CostCentre != "" {
		fmt.Printf("Cost Centre:  %s (unchanged)\n", updatedMVE.CostCentre)
	}

	// Compare and show contract term changes
	if originalMVE.ContractTermMonths != updatedMVE.ContractTermMonths {
		fmt.Printf("Term:         %d months (previously %d months)\n",
			updatedMVE.ContractTermMonths, originalMVE.ContractTermMonths)
	} else {
		fmt.Printf("Term:         %d months (unchanged)\n", updatedMVE.ContractTermMonths)
	}

	return nil
}

// ListMVEs lists all MVEs
func ListMVEs(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Log into the Megaport API
	client, err := Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %v", err)
	}

	// Get filter values from flags
	includeInactive, _ := cmd.Flags().GetBool("inactive")
	nameFilter, _ := cmd.Flags().GetString("name")
	locationID, _ := cmd.Flags().GetInt("location-id")
	vendorFilter, _ := cmd.Flags().GetString("vendor")

	// Get all MVEs from the API
	req := &megaport.ListMVEsRequest{
		IncludeInactive: includeInactive,
	}

	mves, err := listMVEsFunc(ctx, client, req)
	if err != nil {
		return fmt.Errorf("error listing MVEs: %v", err)
	}

	// Apply filters in the CLI application
	filteredMVEs := filterMVEs(mves, nameFilter, locationID, vendorFilter)

	// Print the filtered results
	err = printMVEs(filteredMVEs, outputFormat)
	if err != nil {
		return fmt.Errorf("error printing MVEs: %v", err)
	}
	return nil
}

// filterMVEs applies the specified filters to a list of MVEs
func filterMVEs(mves []*megaport.MVE, nameFilter string, locationID int, vendorFilter string) []*megaport.MVE {
	// If no filters are set, return the original list
	if nameFilter == "" && locationID == 0 && vendorFilter == "" {
		return mves
	}

	filtered := []*megaport.MVE{}

	for _, mve := range mves {
		// Apply name filter (case-insensitive substring match)
		if nameFilter != "" && !strings.Contains(strings.ToLower(mve.Name), strings.ToLower(nameFilter)) {
			continue
		}

		// Apply location ID filter
		if locationID > 0 && mve.LocationID != locationID {
			continue
		}

		// Apply vendor filter (case-insensitive match)
		if vendorFilter != "" && !strings.EqualFold(mve.Vendor, vendorFilter) {
			continue
		}

		// If we get here, the MVE passed all active filters
		filtered = append(filtered, mve)
	}

	return filtered
}

func GetMVE(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %v", err)
	}

	mveUID := args[0]
	if mveUID == "" {
		return fmt.Errorf("MVE UID cannot be empty")
	}

	mve, err := client.MVEService.GetMVE(ctx, mveUID)
	if err != nil {
		return fmt.Errorf("error getting MVE: %v", err)
	}

	if mve == nil {
		return fmt.Errorf("no MVE found with UID: %s", mveUID)
	}

	err = printMVEs([]*megaport.MVE{mve}, outputFormat)
	if err != nil {
		return fmt.Errorf("error printing MVEs: %v", err)
	}
	return nil
}

func ListMVEImages(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %v", err)
	}

	images, err := client.MVEService.ListMVEImages(ctx)
	if err != nil {
		return fmt.Errorf("error listing MVE images: %v", err)
	}

	if images == nil {
		return fmt.Errorf("no MVE images found")
	}

	// Get filter values from flags
	vendor, _ := cmd.Flags().GetString("vendor")
	productCode, _ := cmd.Flags().GetString("product-code")
	id, _ := cmd.Flags().GetInt("id")
	version, _ := cmd.Flags().GetString("version")
	releaseImage, _ := cmd.Flags().GetBool("release-image")

	// Apply filters
	filteredImages := filterMVEImages(images, vendor, productCode, id, version, releaseImage)

	err = printOutput(filteredImages, outputFormat)
	if err != nil {
		return fmt.Errorf("error printing MVE images: %v", err)
	}
	return nil
}

func ListAvailableMVESizes(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %v", err)
	}

	sizes, err := client.MVEService.ListAvailableMVESizes(ctx)
	if err != nil {
		return fmt.Errorf("error listing MVE sizes: %v", err)
	}

	if sizes == nil {
		return fmt.Errorf("no MVE sizes found")
	}

	err = printOutput(sizes, outputFormat)
	if err != nil {
		return fmt.Errorf("error printing MVE sizes: %v", err)
	}
	return nil
}

func DeleteMVE(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	mveUID := args[0]

	client, err := Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %v", err)
	}

	fmt.Printf("Deleting MVE with UID: %s...\n", mveUID)
	req := &megaport.DeleteMVERequest{
		MVEID: mveUID,
	}
	resp, err := client.MVEService.DeleteMVE(ctx, req)
	if err != nil {
		return fmt.Errorf("error deleting MVE: %v", err)
	}

	if resp.IsDeleted {
		fmt.Println("MVE deleted successfully")
	} else {
		fmt.Println("MVE delete failed")
	}
	return nil
}
