package mcr

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/megaport/megaport-cli/internal/base/config"
	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/utils"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

func BuyMCR(cmd *cobra.Command, args []string, noColor bool) error {
	ctx := context.Background()

	// Determine which mode to use
	interactive, _ := cmd.Flags().GetBool("interactive")
	jsonStr, _ := cmd.Flags().GetString("json")
	jsonFile, _ := cmd.Flags().GetString("json-file")

	// Check if any flag-based parameters are provided
	flagsProvided := cmd.Flags().Changed("name") || cmd.Flags().Changed("term") ||
		cmd.Flags().Changed("port-speed") || cmd.Flags().Changed("location-id") ||
		cmd.Flags().Changed("mcr-asn")

	var req *megaport.BuyMCRRequest
	var err error

	// Process input based on mode priority: JSON > Flags > Interactive
	if jsonStr != "" || jsonFile != "" {
		// JSON mode
		output.PrintInfo("Using JSON input", noColor)
		req, err = processJSONMCRInput(jsonStr, jsonFile)
		if err != nil {
			output.PrintError("Failed to process JSON input: %v", noColor, err)
			return err
		}
	} else if flagsProvided {
		// Flag mode
		output.PrintInfo("Using flag input", noColor)
		req, err = processFlagMCRInput(cmd)
		if err != nil {
			output.PrintError("Failed to process flag input: %v", noColor, err)
			return err
		}
	} else if interactive {
		// Interactive mode
		req, err = promptForMCRDetails(noColor)
		if err != nil {
			return err
		}
	} else {
		// No input provided
		output.PrintError("No input provided, use --interactive, --json, or flags to specify MCR details", noColor)
		return fmt.Errorf("no input provided, use --interactive, --json, or flags to specify MCR details")
	}

	// Set common defaults
	req.WaitForProvision = true
	req.WaitForTime = 10 * time.Minute

	// Call the BuyMCR method
	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Error logging in: %v", noColor, err)
		return err
	}
	// Validate MCR Order
	output.PrintInfo("Validating MCR order...", noColor)
	err = client.MCRService.ValidateMCROrder(ctx, req)
	if err != nil {
		output.PrintError("Error validating MCR order: %v", noColor, err)
		return err
	}
	// Buy MCR
	output.PrintInfo("Buying MCR...", noColor)
	resp, err := buyMCRFunc(ctx, client, req)
	if err != nil {
		output.PrintError("Error buying MCR: %v", noColor, err)
		return err
	}

	output.PrintSuccess("MCR created %s", noColor, resp.TechnicalServiceUID)
	return nil
}

func UpdateMCR(cmd *cobra.Command, args []string, noColor bool) error {
	ctx := context.Background()

	if len(args) == 0 {
		return fmt.Errorf("mcr UID is required")
	}

	// Get MCR UID from args
	mcrUID := args[0]

	// Determine which mode to use
	interactive, _ := cmd.Flags().GetBool("interactive")
	jsonStr, _ := cmd.Flags().GetString("json")
	jsonFile, _ := cmd.Flags().GetString("json-file")

	// Check if any flag-based parameters are provided
	flagsProvided := cmd.Flags().Changed("name") || cmd.Flags().Changed("cost-centre") ||
		cmd.Flags().Changed("marketplace-visibility") || cmd.Flags().Changed("term")

	var req *megaport.ModifyMCRRequest
	var err error

	// Process input based on mode priority: JSON > Flags > Interactive
	if jsonStr != "" || jsonFile != "" {
		// JSON mode
		output.PrintInfo("Using JSON input", noColor)
		req, err = processJSONUpdateMCRInput(jsonStr, jsonFile)
		if err != nil {
			output.PrintError("Failed to process JSON input: %v", noColor, err)
			return err
		}
		// Make sure the MCR ID from the command line arguments is set
		req.MCRID = mcrUID
	} else if flagsProvided {
		// Flag mode
		output.PrintInfo("Using flag input", noColor)
		req, err = processFlagUpdateMCRInput(cmd, mcrUID)
		if err != nil {
			output.PrintError("Failed to process flag input: %v", noColor, err)
			return err
		}
	} else if interactive {
		// Interactive mode
		req, err = promptForUpdateMCRDetails(mcrUID, noColor)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("at least one field must be updated")
	}

	// Set common defaults
	req.WaitForUpdate = true
	req.WaitForTime = 10 * time.Minute

	// Call the ModifyMCR method
	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Error logging in: %v", noColor, err)
		return err
	}

	originalMCR, err := getMCRFunc(ctx, client, mcrUID)
	if err != nil {
		output.PrintError("Error getting original MCR: %v", noColor, err)
		return err
	}
	output.PrintInfo("Updating MCR...", noColor)
	_, err = updateMCRFunc(ctx, client, req)
	if err != nil {
		output.PrintError("Error updating MCR: %v", noColor, err)
		return err
	}

	// Call the ModifyMCR method
	output.PrintInfo("Updating MCR %s...", noColor, mcrUID)
	resp, err := updateMCRFunc(ctx, client, req)
	if err != nil {
		output.PrintError("Failed to update MCR: %v", noColor, err)
		return err
	}

	if !resp.IsUpdated {
		output.PrintError("MCR update request was not successful", noColor)
		return fmt.Errorf("MCR update request was not successful")
	}

	// Retrieve the updated MCR for comparison
	updatedMCR, err := getMCRFunc(ctx, client, mcrUID)
	if err != nil {
		output.PrintError("MCR was updated but failed to retrieve updated details: %v", noColor, err)
		output.PrintResourceUpdated("MCR", mcrUID, noColor)
		return nil
	}

	// Print success message
	output.PrintResourceUpdated("MCR", mcrUID, noColor)

	// Display changes between original and updated MCR
	displayMCRChanges(originalMCR, updatedMCR, noColor)

	return nil
}

func CreateMCRPrefixFilterList(cmd *cobra.Command, args []string, noColor bool) error {
	ctx := context.Background()

	if len(args) == 0 {
		return fmt.Errorf("mcr UID is required")
	}

	// Get MCR UID from args
	mcrUID := args[0]

	// Determine which mode to use
	interactive, _ := cmd.Flags().GetBool("interactive")
	jsonStr, _ := cmd.Flags().GetString("json")
	jsonFile, _ := cmd.Flags().GetString("json-file")

	// Check if any flag-based parameters are provided
	flagsProvided := cmd.Flags().Changed("description") || cmd.Flags().Changed("address-family") ||
		cmd.Flags().Changed("entries")

	var req *megaport.CreateMCRPrefixFilterListRequest
	var err error

	// Process input based on mode priority: JSON > Flags > Interactive
	if jsonStr != "" || jsonFile != "" {
		// JSON mode
		output.PrintInfo("Using JSON input", noColor)
		req, err = processJSONPrefixFilterListInput(jsonStr, jsonFile, mcrUID)
		if err != nil {
			output.PrintError("Failed to process JSON input: %v", noColor, err)
			return err
		}
	} else if flagsProvided {
		// Flag mode
		output.PrintInfo("Using flag input", noColor)
		req, err = processFlagPrefixFilterListInput(cmd, mcrUID)
		if err != nil {
			output.PrintError("Failed to process flag input: %v", noColor, err)
			return err
		}
	} else if interactive {
		// Interactive mode
		req, err = promptForPrefixFilterListDetails(mcrUID, noColor)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("no input provided, use --interactive, --json, or flags to specify prefix filter list details")
	}

	// Call the CreatePrefixFilterList method
	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Error logging in: %v", noColor, err)
		return err
	}
	output.PrintInfo("Creating prefix filter list...", noColor)
	resp, err := createMCRPrefixFilterListFunc(ctx, client, req)
	if err != nil {
		output.PrintError("Error creating prefix filter list: %v", noColor, err)
		return err
	}

	output.PrintSuccess("Prefix filter list created successfully - ID: %d", noColor, resp.PrefixFilterListID)
	return nil
}

func UpdateMCRPrefixFilterList(cmd *cobra.Command, args []string, noColor bool) error {
	ctx := context.Background()

	if len(args) < 2 {
		return fmt.Errorf("mcr UID and prefix filter list ID are required")
	}

	// Get MCR UID and prefix filter list ID from args
	mcrUID := args[0]
	prefixFilterListID, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid prefix filter list ID: %v", err)
	}

	// Determine which mode to use
	interactive, _ := cmd.Flags().GetBool("interactive")
	jsonStr, _ := cmd.Flags().GetString("json")
	jsonFile, _ := cmd.Flags().GetString("json-file")

	// Check if any flag-based parameters are provided
	flagsProvided := cmd.Flags().Changed("description") || cmd.Flags().Changed("address-family") ||
		cmd.Flags().Changed("entries")

	var prefixFilterList *megaport.MCRPrefixFilterList
	var getErr error

	// Process input based on mode priority: JSON > Flags > Interactive
	if jsonStr != "" || jsonFile != "" {
		// JSON mode
		output.PrintInfo("Using JSON input", noColor)
		prefixFilterList, getErr = processJSONUpdatePrefixFilterListInput(jsonStr, jsonFile, mcrUID, prefixFilterListID)
		if getErr != nil {
			output.PrintError("Failed to process JSON input: %v", noColor, getErr)
			return getErr
		}
	} else if flagsProvided {
		// Flag mode
		output.PrintInfo("Using flag input", noColor)
		prefixFilterList, getErr = processFlagUpdatePrefixFilterListInput(cmd, mcrUID, prefixFilterListID)
		if getErr != nil {
			output.PrintError("Failed to process flag input: %v", noColor, getErr)
			return getErr
		}
	} else if interactive {
		// Interactive mode
		prefixFilterList, getErr = promptForUpdatePrefixFilterListDetails(mcrUID, prefixFilterListID, noColor)
		if getErr != nil {
			return getErr
		}
	} else {
		return fmt.Errorf("at least one field must be updated")
	}

	// Call the ModifyMCRPrefixFilterList method
	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Error logging in: %v", noColor, err)
		return err
	}
	output.PrintInfo("Updating prefix filter list...", noColor)
	resp, err := modifyMCRPrefixFilterListFunc(ctx, client, mcrUID, prefixFilterListID, prefixFilterList)
	if err != nil {
		output.PrintError("Error updating prefix filter list: %v", noColor, err)
		return err
	}

	if resp.IsUpdated {
		output.PrintSuccess("Prefix filter list updated successfully - ID: %d", noColor, prefixFilterListID)
	} else {
		output.PrintError("Prefix filter list update request was not successful", noColor)
	}
	return nil
}

func GetMCR(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	// Create a context with a 30-second timeout for the API call.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Log into the Megaport API using the provided credentials.
	client, err := config.Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %v", err)
	}

	// Retrieve the MCR UID from the command line arguments.
	mcrUID := args[0]

	// Use the API client to get the MCR details based on the provided UID.
	mcr, err := getMCRFunc(ctx, client, mcrUID)
	if err != nil {
		return fmt.Errorf("error getting MCR: %v", err)
	}

	// Print the MCR details using the desired output format.
	err = printMCRs([]*megaport.MCR{mcr}, outputFormat, noColor)
	if err != nil {
		return fmt.Errorf("error printing MCRs: %v", err)
	}
	return nil
}

func DeleteMCR(cmd *cobra.Command, args []string, noColor bool) error {
	// Create a context with a 30-second timeout for the API call
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Log into the Megaport API using the provided credentials
	client, err := config.Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %v", err)
	}

	// Retrieve the MCR UID from the command line arguments
	mcrUID := args[0]

	// Get delete now flag
	deleteNow, err := cmd.Flags().GetBool("now")
	if err != nil {
		return err
	}

	// Confirm deletion unless force flag is set
	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return err
	}

	if !force {
		confirmMsg := "Are you sure you want to delete MCR " + mcrUID + "? (y/n): "
		confirmation, err := utils.Prompt(confirmMsg, noColor)
		if err != nil {
			return err
		}

		if confirmation != "y" && confirmation != "Y" {
			output.PrintInfo("Deletion cancelled", noColor)
			return nil
		}
	}

	// Create delete request
	deleteRequest := &megaport.DeleteMCRRequest{
		MCRID:     mcrUID,
		DeleteNow: deleteNow,
	}

	output.PrintInfo("Deleting MCR %s...", noColor, mcrUID)

	// Delete the MCR
	resp, err := deleteMCRFunc(ctx, client, deleteRequest)
	if err != nil {
		return fmt.Errorf("error deleting MCR: %v", err)
	}

	if resp.IsDeleting {
		output.PrintResourceDeleted("MCR", mcrUID, deleteNow, noColor)
	} else {
		output.PrintError("MCR deletion request was not successful", noColor)
	}

	return nil
}

func RestoreMCR(cmd *cobra.Command, args []string, noColor bool) error {
	// Create a context with a 30-second timeout for the API call
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Log into the Megaport API using the provided credentials
	client, err := config.Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %v", err)
	}

	// Retrieve the MCR UID from the command line arguments
	mcrUID := args[0]

	output.PrintInfo("Restoring MCR %s...", noColor, mcrUID)

	// Restore the MCR
	resp, err := restoreMCRFunc(ctx, client, mcrUID)
	if err != nil {
		return fmt.Errorf("error restoring MCR: %v", err)
	}

	if resp.IsRestored {
		output.PrintSuccess("MCR %s restored successfully", noColor, mcrUID)
	} else {
		output.PrintError("MCR restoration request was not successful", noColor)
	}

	return nil
}

func ListMCRPrefixFilterLists(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	ctx := context.Background()

	// Log into the Megaport API using the provided credentials
	client, err := config.Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %v", err)
	}

	// Retrieve the MCR UID from the command line arguments
	mcrUID := args[0]

	// Call the ListMCRPrefixFilterLists method
	prefixFilterLists, err := listMCRPrefixFilterListsFunc(ctx, client, mcrUID)
	if err != nil {
		return fmt.Errorf("error listing prefix filter lists: %v", err)
	}

	// Print the prefix filter lists using the desired output format
	err = output.PrintOutput(prefixFilterLists, outputFormat, noColor)
	if err != nil {
		return fmt.Errorf("error printing prefix filter lists: %v", err)
	}
	return nil
}

func GetMCRPrefixFilterList(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	ctx := context.Background()

	// Log into the Megaport API using the provided credentials
	client, err := config.Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %v", err)
	}

	// Retrieve the MCR UID and prefix filter list ID from the command line arguments
	mcrUID := args[0]
	prefixFilterListID, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid prefix filter list ID: %v", err)
	}

	// Call the GetMCRPrefixFilterList method
	prefixFilterList, err := getMCRPrefixFilterListFunc(ctx, client, mcrUID, prefixFilterListID)
	if err != nil {
		return fmt.Errorf("error getting prefix filter list: %v", err)
	}

	// Convert the prefix filter list to the custom output format
	op, err := ToPrefixFilterListOutput(prefixFilterList)
	if err != nil {
		return fmt.Errorf("error converting prefix filter list: %v", err)
	}

	// Print the prefix filter list details using the desired output format
	err = output.PrintOutput([]PrefixFilterListOutput{op}, outputFormat, noColor)
	if err != nil {
		return fmt.Errorf("error printing prefix filter list: %v", err)
	}
	return nil
}

func DeleteMCRPrefixFilterList(cmd *cobra.Command, args []string, noColor bool) error {
	ctx := context.Background()

	// Log into the Megaport API using the provided credentials
	client, err := config.Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %v", err)
	}

	// Retrieve the MCR UID and prefix filter list ID from the command line arguments
	mcrUID := args[0]
	prefixFilterListID, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid prefix filter list ID: %v", err)
	}

	output.PrintInfo("Deleting prefix filter list %d...", noColor, prefixFilterListID)

	// Call the DeleteMCRPrefixFilterList method
	resp, err := deleteMCRPrefixFilterListFunc(ctx, client, mcrUID, prefixFilterListID)
	if err != nil {
		return fmt.Errorf("error deleting prefix filter list: %v", err)
	}

	if resp.IsDeleted {
		output.PrintSuccess("Prefix filter list deleted successfully - ID: %d", noColor, prefixFilterListID)
	} else {
		output.PrintError("Prefix filter list deletion request was not successful", noColor)
	}

	return nil
}
