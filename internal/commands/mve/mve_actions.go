package mve

import (
	"context"
	"fmt"
	"time"

	"github.com/megaport/megaport-cli/internal/base/config"
	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/utils"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

// BuyMVE handles the purchase of a new Megaport Virtual Edge device
func BuyMVE(cmd *cobra.Command, args []string, noColor bool) error {
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
		output.PrintInfo("Using JSON input", noColor)
		req, err = processJSONBuyMVEInput(jsonStr, jsonFile)
		if err != nil {
			output.PrintError("Failed to process JSON input: %v", noColor, err)
			return err
		}
	} else if flagsProvided {
		// Flag mode
		output.PrintInfo("Using flag input", noColor)
		req, err = processFlagBuyMVEInput(cmd)
		if err != nil {
			output.PrintError("Failed to process flag input: %v", noColor, err)
			return err
		}
	} else if interactive {
		// Interactive mode
		output.PrintInfo("Starting interactive mode", noColor)
		req, err = promptForBuyMVEDetails(noColor)
		if err != nil {
			output.PrintError("Interactive input failed: %v", noColor, err)
			return err
		}
	} else {
		output.PrintError("No input provided, use --interactive, --json, or flags to specify MVE details", noColor)
		return fmt.Errorf("no input provided, use --interactive, --json, or flags to specify MVE details")
	}

	// Call the API to buy the MVE
	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}

	output.PrintInfo("Validating MVE order...", noColor)
	if err := client.MVEService.ValidateMVEOrder(ctx, req); err != nil {
		output.PrintError("Validation failed: %v", noColor, err)
		return fmt.Errorf("validation failed: %v", err)
	}

	req.WaitForProvision = true
	req.WaitForTime = 10 * time.Minute

	output.PrintInfo("Buying MVE...", noColor)
	resp, err := client.MVEService.BuyMVE(ctx, req)
	if err != nil {
		output.PrintError("Failed to buy MVE: %v", noColor, err)
		return err
	}

	output.PrintResourceCreated("MVE", resp.TechnicalServiceUID, noColor)
	return nil
}

func UpdateMVE(cmd *cobra.Command, args []string, noColor bool) error {
	ctx := context.Background()
	mveUID := args[0]
	formattedUID := output.FormatUID(mveUID, noColor)

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("error logging in: %v", err)
	}

	// Fetch original MVE details before update
	originalMVE, err := client.MVEService.GetMVE(ctx, mveUID)
	if err != nil {
		output.PrintError("Failed to get original MVE details: %v", noColor, err)
		return fmt.Errorf("error getting MVE details: %v", err)
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
			output.PrintError("Failed to process JSON input: %v", noColor, err)
			return fmt.Errorf("error processing JSON input: %v", err)
		}
	} else if flagsProvided {
		// Flag mode
		req, err = processFlagUpdateMVEInput(cmd, mveUID)
		if err != nil {
			output.PrintError("Failed to process flag input: %v", noColor, err)
			return fmt.Errorf("error processing flag input: %v", err)
		}
	} else if interactive {
		// Interactive mode
		output.PrintInfo("Starting interactive mode for MVE %s", noColor, formattedUID)
		req, err = promptForUpdateMVEDetails(mveUID, noColor)
		if err != nil {
			output.PrintError("Failed to get MVE details interactively: %v", noColor, err)
			return fmt.Errorf("error getting MVE details interactively: %v", err)
		}
	} else {
		output.PrintError("No input provided", noColor)
		return fmt.Errorf("no input provided, use --interactive, --json, or flags to specify MVE update details")
	}

	// Set common defaults
	req.WaitForUpdate = true
	req.WaitForTime = 10 * time.Minute

	// Call the ModifyMVE method
	output.PrintInfo("Updating MVE %s...", noColor, formattedUID)
	resp, err := client.MVEService.ModifyMVE(ctx, req)
	if err != nil {
		output.PrintError("Failed to update MVE: %v", noColor, err)
		return err
	}

	if !resp.MVEUpdated {
		output.PrintWarning("MVE update request was not successful", noColor)
		return fmt.Errorf("MVE update request was not successful")
	}

	// Fetch the updated MVE to get the new values
	updatedMVE, err := client.MVEService.GetMVE(ctx, mveUID)
	if err != nil {
		output.PrintError("MVE was updated but failed to retrieve updated details: %v", noColor, err)
		output.PrintResourceUpdated("MVE", mveUID, noColor)
		return nil
	}

	// Output success message
	output.PrintResourceUpdated("MVE", mveUID, noColor)

	// Display changes between original and updated MVE
	displayMVEChanges(originalMVE, updatedMVE, noColor)

	return nil
}

func GetMVE(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("error logging in: %v", err)
	}

	mveUID := args[0]
	formattedUID := output.FormatUID(mveUID, noColor)
	if mveUID == "" {
		output.PrintError("MVE UID cannot be empty", noColor)
		return fmt.Errorf("MVE UID cannot be empty")
	}

	output.PrintInfo("Retrieving MVE %s...", noColor, formattedUID)
	mve, err := client.MVEService.GetMVE(ctx, mveUID)
	if err != nil {
		output.PrintError("Failed to get MVE: %v", noColor, err)
		return fmt.Errorf("error getting MVE: %v", err)
	}

	if mve == nil {
		output.PrintError("No MVE found with UID: %s", noColor, mveUID)
		return fmt.Errorf("no MVE found with UID: %s", mveUID)
	}

	err = output.PrintOutput([]*megaport.MVE{mve}, outputFormat, noColor)
	if err != nil {
		output.PrintError("Failed to print MVEs: %v", noColor, err)
		return fmt.Errorf("error printing MVEs: %v", err)
	}
	return nil
}

func ListMVEImages(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("error logging in: %v", err)
	}

	output.PrintInfo("Retrieving MVE images...", noColor)
	images, err := client.MVEService.ListMVEImages(ctx)
	if err != nil {
		output.PrintError("Failed to list MVE images: %v", noColor, err)
		return fmt.Errorf("error listing MVE images: %v", err)
	}

	if images == nil {
		output.PrintWarning("No MVE images found", noColor)
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

	err = output.PrintOutput(filteredImages, outputFormat, noColor)
	if err != nil {
		output.PrintError("Failed to print MVE images: %v", noColor, err)
		return fmt.Errorf("error printing MVE images: %v", err)
	}
	return nil
}

func ListAvailableMVESizes(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("error logging in: %v", err)
	}

	output.PrintInfo("Retrieving available MVE sizes...", noColor)
	sizes, err := client.MVEService.ListAvailableMVESizes(ctx)
	if err != nil {
		output.PrintError("Failed to list MVE sizes: %v", noColor, err)
		return fmt.Errorf("error listing MVE sizes: %v", err)
	}

	if sizes == nil {
		output.PrintWarning("No MVE sizes found", noColor)
		return fmt.Errorf("no MVE sizes found")
	}

	err = output.PrintOutput(sizes, outputFormat, noColor)
	if err != nil {
		output.PrintError("Failed to print MVE sizes: %v", noColor, err)
		return fmt.Errorf("error printing MVE sizes: %v", err)
	}
	return nil
}

func DeleteMVE(cmd *cobra.Command, args []string, noColor bool) error {
	ctx := context.Background()
	mveUID := args[0]
	formattedUID := output.FormatUID(mveUID, noColor)

	// Confirm deletion unless force flag is set
	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		output.PrintError("Failed to get force flag: %v", noColor, err)
		return err
	}

	if !force {
		confirmMsg := "Are you sure you want to delete MVE " + mveUID + "? "
		if !utils.ConfirmPrompt(confirmMsg, noColor) {
			output.PrintInfo("Deletion cancelled", noColor)
			return nil
		}
	}

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("error logging in: %v", err)
	}

	output.PrintInfo("Deleting MVE %s...", noColor, formattedUID)
	req := &megaport.DeleteMVERequest{
		MVEID: mveUID,
	}
	resp, err := client.MVEService.DeleteMVE(ctx, req)
	if err != nil {
		output.PrintError("Failed to delete MVE: %v", noColor, err)
		return fmt.Errorf("error deleting MVE: %v", err)
	}

	if resp.IsDeleted {
		output.PrintResourceDeleted("MVE", mveUID, false, noColor)
	} else {
		output.PrintWarning("MVE delete failed", noColor)
	}
	return nil
}
