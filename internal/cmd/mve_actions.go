package cmd

import (
	"context"
	"fmt"
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
		PrintInfo("Using JSON input")
		req, err = processJSONBuyMVEInput(jsonStr, jsonFile)
		if err != nil {
			PrintError("Failed to process JSON input: %v", err)
			return err
		}
	} else if flagsProvided {
		// Flag mode
		PrintInfo("Using flag input")
		req, err = processFlagBuyMVEInput(cmd)
		if err != nil {
			PrintError("Failed to process flag input: %v", err)
			return err
		}
	} else if interactive {
		// Interactive mode
		PrintInfo("Starting interactive mode")
		req, err = promptForBuyMVEDetails()
		if err != nil {
			PrintError("Interactive input failed: %v", err)
			return err
		}
	} else {
		PrintError("No input provided")
		return fmt.Errorf("no input provided, use --interactive, --json, or flags to specify MVE details")
	}

	// Call the API to buy the MVE
	client, err := Login(ctx)
	if err != nil {
		PrintError("Failed to log in: %v", err)
		return err
	}

	PrintInfo("Validating MVE order...")
	if err := client.MVEService.ValidateMVEOrder(ctx, req); err != nil {
		PrintError("Validation failed: %v", err)
		return fmt.Errorf("validation failed: %v", err)
	}

	PrintInfo("Buying MVE...")
	resp, err := client.MVEService.BuyMVE(ctx, req)
	if err != nil {
		PrintError("Failed to buy MVE: %v", err)
		return err
	}

	PrintResourceCreated("MVE", resp.TechnicalServiceUID)
	return nil
}

func UpdateMVE(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Retrieve the MVE UID from command line arguments
	mveUID := args[0]
	formattedUID := formatUID(mveUID)

	// Get the original MVE to compare values later
	client, err := Login(ctx)
	if err != nil {
		PrintError("Failed to log in: %v", err)
		return err
	}

	// Fetch original MVE details before update
	originalMVE, err := client.MVEService.GetMVE(ctx, mveUID)
	if err != nil {
		PrintError("Error retrieving MVE details: %v", err)
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
		PrintInfo("Using JSON input for MVE %s", formattedUID)
		req, err = processJSONUpdateMVEInput(jsonStr, jsonFile, mveUID)
		if err != nil {
			PrintError("Failed to process JSON input: %v", err)
			return err
		}
	} else if flagsProvided {
		// Flag mode
		PrintInfo("Using flag input for MVE %s", formattedUID)
		req, err = processFlagUpdateMVEInput(cmd, mveUID)
		if err != nil {
			PrintError("Failed to process flag input: %v", err)
			return err
		}
	} else if interactive {
		// Interactive mode
		PrintInfo("Starting interactive mode for MVE %s", formattedUID)
		req, err = promptForUpdateMVEDetails(mveUID)
		if err != nil {
			PrintError("Interactive input failed: %v", err)
			return err
		}
	} else {
		PrintError("No input provided")
		return fmt.Errorf("no input provided, use --interactive, --json, or flags to specify MVE update details")
	}

	// Set common defaults
	req.WaitForUpdate = true
	req.WaitForTime = 10 * time.Minute

	// Call the ModifyMVE method
	PrintInfo("Updating MVE %s...", formattedUID)
	resp, err := client.MVEService.ModifyMVE(ctx, req)
	if err != nil {
		PrintError("Failed to update MVE: %v", err)
		return fmt.Errorf("error updating MVE: %v", err)
	}

	if !resp.MVEUpdated {
		PrintWarning("MVE update request was not successful")
		return nil
	}

	// Fetch the updated MVE to get the new values
	updatedMVE, err := client.MVEService.GetMVE(ctx, mveUID)
	if err != nil {
		PrintError("Error retrieving updated MVE details: %v", err)
		return fmt.Errorf("error retrieving updated MVE details: %v", err)
	}

	// Print detailed success message
	PrintResourceUpdated("MVE", mveUID)

	// Compare and show name changes
	if originalMVE.Name != updatedMVE.Name {
		PrintInfo("Name:         %s (previously \"%s\")", updatedMVE.Name, originalMVE.Name)
	} else {
		PrintInfo("Name:         %s (unchanged)", updatedMVE.Name)
	}

	// Compare and show cost centre changes
	if originalMVE.CostCentre != updatedMVE.CostCentre {
		// Handle empty cost centre specially
		origCC := originalMVE.CostCentre
		if origCC == "" {
			origCC = "none"
		}
		PrintInfo("Cost Centre:  %s (previously \"%s\")", updatedMVE.CostCentre, origCC)
	} else if updatedMVE.CostCentre != "" {
		PrintInfo("Cost Centre:  %s (unchanged)", updatedMVE.CostCentre)
	}

	// Compare and show contract term changes
	if originalMVE.ContractTermMonths != updatedMVE.ContractTermMonths {
		PrintInfo("Term:         %d months (previously %d months)",
			updatedMVE.ContractTermMonths, originalMVE.ContractTermMonths)
	} else {
		PrintInfo("Term:         %d months (unchanged)", updatedMVE.ContractTermMonths)
	}

	return nil
}

func GetMVE(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := Login(ctx)
	if err != nil {
		PrintError("Failed to log in: %v", err)
		return fmt.Errorf("error logging in: %v", err)
	}

	mveUID := args[0]
	formattedUID := formatUID(mveUID)
	if mveUID == "" {
		PrintError("MVE UID cannot be empty")
		return fmt.Errorf("mVE UID cannot be empty")
	}

	PrintInfo("Retrieving MVE %s...", formattedUID)
	mve, err := client.MVEService.GetMVE(ctx, mveUID)
	if err != nil {
		PrintError("Failed to get MVE: %v", err)
		return fmt.Errorf("error getting MVE: %v", err)
	}

	if mve == nil {
		PrintError("No MVE found with UID: %s", mveUID)
		return fmt.Errorf("no MVE found with UID: %s", mveUID)
	}

	err = printMVEs([]*megaport.MVE{mve}, outputFormat)
	if err != nil {
		PrintError("Failed to print MVEs: %v", err)
		return fmt.Errorf("error printing MVEs: %v", err)
	}
	return nil
}

func ListMVEImages(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := Login(ctx)
	if err != nil {
		PrintError("Failed to log in: %v", err)
		return fmt.Errorf("error logging in: %v", err)
	}

	PrintInfo("Retrieving MVE images...")
	images, err := client.MVEService.ListMVEImages(ctx)
	if err != nil {
		PrintError("Failed to list MVE images: %v", err)
		return fmt.Errorf("error listing MVE images: %v", err)
	}

	if images == nil {
		PrintWarning("No MVE images found")
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
		PrintError("Failed to print MVE images: %v", err)
		return fmt.Errorf("error printing MVE images: %v", err)
	}
	return nil
}

func ListAvailableMVESizes(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := Login(ctx)
	if err != nil {
		PrintError("Failed to log in: %v", err)
		return fmt.Errorf("error logging in: %v", err)
	}

	PrintInfo("Retrieving available MVE sizes...")
	sizes, err := client.MVEService.ListAvailableMVESizes(ctx)
	if err != nil {
		PrintError("Failed to list MVE sizes: %v", err)
		return fmt.Errorf("error listing MVE sizes: %v", err)
	}

	if sizes == nil {
		PrintWarning("No MVE sizes found")
		return fmt.Errorf("no MVE sizes found")
	}

	err = printOutput(sizes, outputFormat)
	if err != nil {
		PrintError("Failed to print MVE sizes: %v", err)
		return fmt.Errorf("error printing MVE sizes: %v", err)
	}
	return nil
}

func DeleteMVE(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	mveUID := args[0]
	formattedUID := formatUID(mveUID)

	// Confirm deletion unless force flag is set
	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		PrintError("Failed to get force flag: %v", err)
		return err
	}

	if !force {
		confirmMsg := "Are you sure you want to delete MVE " + mveUID + "? "
		if !confirmPrompt(confirmMsg) {
			PrintInfo("Deletion cancelled")
			return nil
		}
	}

	client, err := Login(ctx)
	if err != nil {
		PrintError("Failed to log in: %v", err)
		return fmt.Errorf("error logging in: %v", err)
	}

	PrintInfo("Deleting MVE %s...", formattedUID)
	req := &megaport.DeleteMVERequest{
		MVEID: mveUID,
	}
	resp, err := client.MVEService.DeleteMVE(ctx, req)
	if err != nil {
		PrintError("Failed to delete MVE: %v", err)
		return fmt.Errorf("error deleting MVE: %v", err)
	}

	if resp.IsDeleted {
		PrintResourceDeleted("MVE", mveUID, false)
	} else {
		PrintWarning("MVE delete failed")
	}
	return nil
}
