package cmd

import (
	"context"
	"fmt"
	"strconv"
	"time"

	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

func BuyMCR(cmd *cobra.Command, args []string) error {
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
		PrintInfo("Using JSON input")
		req, err = processJSONMCRInput(jsonStr, jsonFile)
		if err != nil {
			PrintError("Failed to process JSON input: %v", err)
			return err
		}
	} else if flagsProvided {
		// Flag mode
		PrintInfo("Using flag input")
		req, err = processFlagMCRInput(cmd)
		if err != nil {
			PrintError("Failed to process flag input: %v", err)
			return err
		}
	} else if interactive {
		// Interactive mode
		PrintInfo("Starting interactive mode")
		req, err = promptForMCRDetails()
		if err != nil {
			PrintError("Interactive input failed: %v", err)
			return err
		}
	} else {
		PrintError("No input provided")
		return fmt.Errorf("no input provided, use --interactive, --json, or flags to specify MCR details")
	}

	// Set common defaults
	req.WaitForProvision = true
	req.WaitForTime = 10 * time.Minute

	// Call the BuyMCR method
	client, err := Login(ctx)
	if err != nil {
		PrintError("Failed to log in: %v", err)
		return err
	}
	PrintInfo("Buying MCR...")
	resp, err := buyMCRFunc(ctx, client, req)
	if err != nil {
		PrintError("Failed to buy MCR: %v", err)
		return err
	}

	PrintResourceCreated("MCR", resp.TechnicalServiceUID)
	return nil
}

func UpdateMCR(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	if len(args) == 0 {
		PrintError("MCR UID is required")
		return fmt.Errorf("mcr UID is required")
	}

	// Get MCR UID from args
	mcrUID := args[0]
	formattedUID := formatUID(mcrUID)

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
		PrintInfo("Using JSON input for MCR %s", formattedUID)
		req, err = processJSONUpdateMCRInput(jsonStr, jsonFile)
		if err != nil {
			PrintError("Failed to process JSON input: %v", err)
			return err
		}
		// Make sure the MCR ID from the command line arguments is set
		req.MCRID = mcrUID
	} else if flagsProvided {
		// Flag mode
		PrintInfo("Using flag input for MCR %s", formattedUID)
		req, err = processFlagUpdateMCRInput(cmd, mcrUID)
		if err != nil {
			PrintError("Failed to process flag input: %v", err)
			return err
		}
	} else if interactive {
		// Interactive mode
		PrintInfo("Starting interactive mode for MCR %s", formattedUID)
		req, err = promptForUpdateMCRDetails(mcrUID)
		if err != nil {
			PrintError("Interactive input failed: %v", err)
			return err
		}
	} else {
		PrintError("No input provided")
		return fmt.Errorf("no input provided, use --interactive, --json, or flags to specify MCR update details")
	}

	// Set common defaults
	req.WaitForUpdate = true
	req.WaitForTime = 10 * time.Minute

	// Call the ModifyMCR method
	client, err := Login(ctx)
	if err != nil {
		PrintError("Failed to log in: %v", err)
		return err
	}
	PrintInfo("Updating MCR %s...", formattedUID)
	resp, err := updateMCRFunc(ctx, client, req)
	if err != nil {
		PrintError("Failed to update MCR: %v", err)
		return err
	}

	if resp.IsUpdated {
		PrintResourceUpdated("MCR", mcrUID)
	} else {
		PrintWarning("MCR update request was not successful")
	}
	return nil
}

func CreateMCRPrefixFilterList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	if len(args) == 0 {
		PrintError("MCR UID is required")
		return fmt.Errorf("mcr UID is required")
	}

	// Get MCR UID from args
	mcrUID := args[0]
	formattedUID := formatUID(mcrUID)

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
		PrintInfo("Using JSON input for MCR %s", formattedUID)
		req, err = processJSONPrefixFilterListInput(jsonStr, jsonFile, mcrUID)
		if err != nil {
			PrintError("Failed to process JSON input: %v", err)
			return err
		}
	} else if flagsProvided {
		// Flag mode
		PrintInfo("Using flag input for MCR %s", formattedUID)
		req, err = processFlagPrefixFilterListInput(cmd, mcrUID)
		if err != nil {
			PrintError("Failed to process flag input: %v", err)
			return err
		}
	} else if interactive {
		// Interactive mode
		PrintInfo("Starting interactive mode for MCR %s", formattedUID)
		req, err = promptForPrefixFilterListDetails(mcrUID)
		if err != nil {
			PrintError("Interactive input failed: %v", err)
			return err
		}
	} else {
		PrintError("No input provided")
		return fmt.Errorf("no input provided, use --interactive, --json, or flags to specify prefix filter list details")
	}

	// Call the CreatePrefixFilterList method
	client, err := Login(ctx)
	if err != nil {
		PrintError("Failed to log in: %v", err)
		return err
	}
	PrintInfo("Creating prefix filter list for MCR %s...", formattedUID)
	resp, err := createMCRPrefixFilterListFunc(ctx, client, req)
	if err != nil {
		PrintError("Failed to create prefix filter list: %v", err)
		return err
	}

	PrintSuccess("Prefix filter list created successfully - ID: %d", resp.PrefixFilterListID)
	return nil
}

func UpdateMCRPrefixFilterList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	if len(args) < 2 {
		PrintError("MCR UID and prefix filter list ID are required")
		return fmt.Errorf("mcr UID and prefix filter list ID are required")
	}

	// Get MCR UID and prefix filter list ID from args
	mcrUID := args[0]
	formattedUID := formatUID(mcrUID)

	prefixFilterListID, err := strconv.Atoi(args[1])
	if err != nil {
		PrintError("Invalid prefix filter list ID: %v", err)
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
		PrintInfo("Using JSON input for prefix filter list %d", prefixFilterListID)
		prefixFilterList, getErr = processJSONUpdatePrefixFilterListInput(jsonStr, jsonFile, prefixFilterListID)
		if getErr != nil {
			PrintError("Failed to process JSON input: %v", getErr)
			return getErr
		}
	} else if flagsProvided {
		// Flag mode
		PrintInfo("Using flag input for prefix filter list %d", prefixFilterListID)
		prefixFilterList, getErr = processFlagUpdatePrefixFilterListInput(cmd, prefixFilterListID)
		if getErr != nil {
			PrintError("Failed to process flag input: %v", getErr)
			return getErr
		}
	} else if interactive {
		// Interactive mode
		PrintInfo("Starting interactive mode for prefix filter list %d", prefixFilterListID)
		prefixFilterList, getErr = promptForUpdatePrefixFilterListDetails(mcrUID, prefixFilterListID)
		if getErr != nil {
			PrintError("Interactive input failed: %v", getErr)
			return getErr
		}
	} else {
		PrintError("No input provided")
		return fmt.Errorf("no input provided, use --interactive, --json, or flags to specify prefix filter list update details")
	}

	// Call the ModifyMCRPrefixFilterList method
	client, err := Login(ctx)
	if err != nil {
		PrintError("Failed to log in: %v", err)
		return err
	}
	PrintInfo("Updating prefix filter list %d for MCR %s...", prefixFilterListID, formattedUID)
	resp, err := modifyMCRPrefixFilterListFunc(ctx, client, mcrUID, prefixFilterListID, prefixFilterList)
	if err != nil {
		PrintError("Failed to update prefix filter list: %v", err)
		return err
	}

	if resp.IsUpdated {
		PrintSuccess("Prefix filter list updated successfully - ID: %d", prefixFilterListID)
	} else {
		PrintWarning("Prefix filter list update request was not successful")
	}
	return nil
}

func GetMCR(cmd *cobra.Command, args []string) error {
	// Create a context with a 30-second timeout for the API call.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Log into the Megaport API using the provided credentials.
	client, err := Login(ctx)
	if err != nil {
		PrintError("Failed to log in: %v", err)
		return fmt.Errorf("error logging in: %v", err)
	}

	// Retrieve the MCR UID from the command line arguments.
	mcrUID := args[0]
	formattedUID := formatUID(mcrUID)

	PrintInfo("Retrieving MCR %s...", formattedUID)

	// Use the API client to get the MCR details based on the provided UID.
	mcr, err := getMCRFunc(ctx, client, mcrUID)
	if err != nil {
		PrintError("Failed to retrieve MCR: %v", err)
		return fmt.Errorf("error getting MCR: %v", err)
	}

	// Print the MCR details using the desired output format.
	err = printMCRs([]*megaport.MCR{mcr}, outputFormat)
	if err != nil {
		PrintError("Failed to print MCR details: %v", err)
		return fmt.Errorf("error printing MCRs: %v", err)
	}
	return nil
}

func DeleteMCR(cmd *cobra.Command, args []string) error {
	// Create a context with a 30-second timeout for the API call
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Log into the Megaport API using the provided credentials
	client, err := Login(ctx)
	if err != nil {
		PrintError("Failed to log in: %v", err)
		return fmt.Errorf("error logging in: %v", err)
	}

	// Retrieve the MCR UID from the command line arguments
	mcrUID := args[0]
	formattedUID := formatUID(mcrUID)

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
		confirmMsg := fmt.Sprintf("Are you sure you want to delete MCR %s? ", formattedUID)
		confirmation, err := prompt(FormatConfirmation(confirmMsg))
		if err != nil {
			PrintError("Failed to get confirmation: %v", err)
			return err
		}

		if confirmation != "y" && confirmation != "Y" {
			PrintInfo("Deletion cancelled")
			return nil
		}
	}

	// Create delete request
	deleteRequest := &megaport.DeleteMCRRequest{
		MCRID:     mcrUID,
		DeleteNow: deleteNow,
	}

	PrintInfo("Deleting MCR %s...", formattedUID)

	// Delete the MCR
	resp, err := deleteMCRFunc(ctx, client, deleteRequest)
	if err != nil {
		PrintError("Failed to delete MCR: %v", err)
		return fmt.Errorf("error deleting MCR: %v", err)
	}

	if resp.IsDeleting {
		PrintResourceDeleted("MCR", mcrUID, deleteNow)
	} else {
		PrintWarning("MCR deletion request was not successful")
	}

	return nil
}

func RestoreMCR(cmd *cobra.Command, args []string) error {
	// Create a context with a 30-second timeout for the API call
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Log into the Megaport API using the provided credentials
	client, err := Login(ctx)
	if err != nil {
		PrintError("Failed to log in: %v", err)
		return fmt.Errorf("error logging in: %v", err)
	}

	// Retrieve the MCR UID from the command line arguments
	mcrUID := args[0]
	formattedUID := formatUID(mcrUID)

	PrintInfo("Restoring MCR %s...", formattedUID)

	// Restore the MCR
	resp, err := restoreMCRFunc(ctx, client, mcrUID)
	if err != nil {
		PrintError("Failed to restore MCR: %v", err)
		return fmt.Errorf("error restoring MCR: %v", err)
	}

	if resp.IsRestored {
		PrintResourceSuccess("MCR", "restored", mcrUID)
	} else {
		PrintWarning("MCR restoration request was not successful")
	}

	return nil
}

func ListMCRPrefixFilterLists(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Log into the Megaport API using the provided credentials
	client, err := Login(ctx)
	if err != nil {
		PrintError("Failed to log in: %v", err)
		return fmt.Errorf("error logging in: %v", err)
	}

	// Retrieve the MCR UID from the command line arguments
	mcrUID := args[0]
	formattedUID := formatUID(mcrUID)

	PrintInfo("Retrieving prefix filter lists for MCR %s...", formattedUID)

	// Call the ListMCRPrefixFilterLists method
	prefixFilterLists, err := listMCRPrefixFilterListsFunc(ctx, client, mcrUID)
	if err != nil {
		PrintError("Failed to retrieve prefix filter lists: %v", err)
		return fmt.Errorf("error listing prefix filter lists: %v", err)
	}

	if len(prefixFilterLists) == 0 {
		PrintInfo("No prefix filter lists found for MCR %s", formattedUID)
	} else {
		PrintInfo("Found %d prefix filter lists", len(prefixFilterLists))
	}

	// Print the prefix filter lists using the desired output format
	err = printOutput(prefixFilterLists, outputFormat)
	if err != nil {
		PrintError("Failed to print prefix filter lists: %v", err)
		return fmt.Errorf("error printing prefix filter lists: %v", err)
	}
	return nil
}

func GetMCRPrefixFilterList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Log into the Megaport API using the provided credentials
	client, err := Login(ctx)
	if err != nil {
		PrintError("Failed to log in: %v", err)
		return fmt.Errorf("error logging in: %v", err)
	}

	// Retrieve the MCR UID and prefix filter list ID from the command line arguments
	mcrUID := args[0]
	formattedUID := formatUID(mcrUID)

	prefixFilterListID, err := strconv.Atoi(args[1])
	if err != nil {
		PrintError("Invalid prefix filter list ID: %v", err)
		return fmt.Errorf("invalid prefix filter list ID: %v", err)
	}

	PrintInfo("Retrieving prefix filter list %d for MCR %s...", prefixFilterListID, formattedUID)

	// Call the GetMCRPrefixFilterList method
	prefixFilterList, err := getMCRPrefixFilterListFunc(ctx, client, mcrUID, prefixFilterListID)
	if err != nil {
		PrintError("Failed to retrieve prefix filter list: %v", err)
		return fmt.Errorf("error getting prefix filter list: %v", err)
	}

	// Convert the prefix filter list to the custom output format
	output, err := ToPrefixFilterListOutput(prefixFilterList)
	if err != nil {
		PrintError("Failed to convert prefix filter list: %v", err)
		return fmt.Errorf("error converting prefix filter list: %v", err)
	}

	// Print the prefix filter list details using the desired output format
	err = printOutput([]PrefixFilterListOutput{output}, outputFormat)
	if err != nil {
		PrintError("Failed to print prefix filter list: %v", err)
		return fmt.Errorf("error printing prefix filter list: %v", err)
	}
	return nil
}

func DeleteMCRPrefixFilterList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Log into the Megaport API using the provided credentials
	client, err := Login(ctx)
	if err != nil {
		PrintError("Failed to log in: %v", err)
		return fmt.Errorf("error logging in: %v", err)
	}

	// Retrieve the MCR UID and prefix filter list ID from the command line arguments
	mcrUID := args[0]
	formattedUID := formatUID(mcrUID)

	prefixFilterListID, err := strconv.Atoi(args[1])
	if err != nil {
		PrintError("Invalid prefix filter list ID: %v", err)
		return fmt.Errorf("invalid prefix filter list ID: %v", err)
	}

	// Confirm deletion unless force flag is set
	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return err
	}

	if !force {
		confirmMsg := fmt.Sprintf("Are you sure you want to delete prefix filter list %d for MCR %s? ",
			prefixFilterListID, formattedUID)
		confirmation, err := prompt(FormatConfirmation(confirmMsg))
		if err != nil {
			PrintError("Failed to get confirmation: %v", err)
			return err
		}

		if confirmation != "y" && confirmation != "Y" {
			PrintInfo("Deletion cancelled")
			return nil
		}
	}

	PrintInfo("Deleting prefix filter list %d for MCR %s...", prefixFilterListID, formattedUID)

	// Call the DeleteMCRPrefixFilterList method
	resp, err := deleteMCRPrefixFilterListFunc(ctx, client, mcrUID, prefixFilterListID)
	if err != nil {
		PrintError("Failed to delete prefix filter list: %v", err)
		return fmt.Errorf("error deleting prefix filter list: %v", err)
	}

	if resp.IsDeleted {
		PrintSuccess("Prefix filter list %d deleted successfully", prefixFilterListID)
	} else {
		PrintWarning("Prefix filter list deletion request was not successful")
	}

	return nil
}
