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
		req, err = processJSONMCRInput(jsonStr, jsonFile)
		if err != nil {
			return err
		}
	} else if flagsProvided {
		// Flag mode
		req, err = processFlagMCRInput(cmd)
		if err != nil {
			return err
		}
	} else if interactive {
		// Interactive mode
		req, err = promptForMCRDetails()
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("no input provided, use --interactive, --json, or flags to specify MCR details")
	}

	// Set common defaults
	req.WaitForProvision = true
	req.WaitForTime = 10 * time.Minute

	// Call the BuyMCR method
	client, err := Login(ctx)
	if err != nil {
		return err
	}
	fmt.Println("Buying MCR...")
	resp, err := buyMCRFunc(ctx, client, req)
	if err != nil {
		return err
	}

	fmt.Printf("MCR purchased successfully - UID: %s\n", resp.TechnicalServiceUID)
	return nil
}

// ListMCRs lists all MCRs
func ListMCRs(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Log into the Megaport API using the provided credentials
	client, err := Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %v", err)
	}

	// Get the value of the "inactive" flag
	includeInactive, err := cmd.Flags().GetBool("inactive")
	if err != nil {
		return fmt.Errorf("error getting inactive flag: %v", err)
	}
	mcrReq := &megaport.ListMCRsRequest{
		IncludeInactive: includeInactive,
	}

	// Call the ListMCRs method
	mcrs, err := listMCRsFunc(ctx, client, mcrReq)
	if err != nil {
		return fmt.Errorf("error listing mcrs: %v", err)
	}

	// Print the mcrs using the desired output format
	err = printMCRs(mcrs, outputFormat)
	if err != nil {
		return fmt.Errorf("error printing mcrs: %v", err)
	}
	return nil
}

func UpdateMCR(cmd *cobra.Command, args []string) error {
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
		req, err = processJSONUpdateMCRInput(jsonStr, jsonFile)
		if err != nil {
			return err
		}
		// Make sure the MCR ID from the command line arguments is set
		req.MCRID = mcrUID
	} else if flagsProvided {
		// Flag mode
		req, err = processFlagUpdateMCRInput(cmd, mcrUID)
		if err != nil {
			return err
		}
	} else if interactive {
		// Interactive mode
		req, err = promptForUpdateMCRDetails(mcrUID)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("no input provided, use --interactive, --json, or flags to specify MCR update details")
	}

	// Set common defaults
	req.WaitForUpdate = true
	req.WaitForTime = 10 * time.Minute

	// Call the ModifyMCR method
	client, err := Login(ctx)
	if err != nil {
		return err
	}
	fmt.Println("Updating MCR...")
	resp, err := updateMCRFunc(ctx, client, req)
	if err != nil {
		return err
	}

	if resp.IsUpdated {
		fmt.Printf("MCR updated successfully - UID: %s\n", mcrUID)
	} else {
		fmt.Println("MCR update request was not successful")
	}
	return nil
}

func CreateMCRPrefixFilterList(cmd *cobra.Command, args []string) error {
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
		req, err = processJSONPrefixFilterListInput(jsonStr, jsonFile, mcrUID)
		if err != nil {
			return err
		}
	} else if flagsProvided {
		// Flag mode
		req, err = processFlagPrefixFilterListInput(cmd, mcrUID)
		if err != nil {
			return err
		}
	} else if interactive {
		// Interactive mode
		req, err = promptForPrefixFilterListDetails(mcrUID)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("no input provided, use --interactive, --json, or flags to specify prefix filter list details")
	}

	// Call the CreatePrefixFilterList method
	client, err := Login(ctx)
	if err != nil {
		return err
	}
	fmt.Println("Creating prefix filter list...")
	resp, err := createMCRPrefixFilterListFunc(ctx, client, req)
	if err != nil {
		return err
	}

	fmt.Printf("Prefix filter list created successfully - ID: %d\n", resp.PrefixFilterListID)
	return nil
}

func UpdateMCRPrefixFilterList(cmd *cobra.Command, args []string) error {
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
		prefixFilterList, getErr = processJSONUpdatePrefixFilterListInput(jsonStr, jsonFile, prefixFilterListID)
		if getErr != nil {
			return getErr
		}
	} else if flagsProvided {
		// Flag mode
		prefixFilterList, getErr = processFlagUpdatePrefixFilterListInput(cmd, prefixFilterListID)
		if getErr != nil {
			return getErr
		}
	} else if interactive {
		// Interactive mode
		prefixFilterList, getErr = promptForUpdatePrefixFilterListDetails(mcrUID, prefixFilterListID)
		if getErr != nil {
			return getErr
		}
	} else {
		return fmt.Errorf("no input provided, use --interactive, --json, or flags to specify prefix filter list update details")
	}

	// Call the ModifyMCRPrefixFilterList method
	client, err := Login(ctx)
	if err != nil {
		return err
	}
	fmt.Println("Updating prefix filter list...")
	resp, err := modifyMCRPrefixFilterListFunc(ctx, client, mcrUID, prefixFilterListID, prefixFilterList)
	if err != nil {
		return err
	}

	if resp.IsUpdated {
		fmt.Printf("Prefix filter list updated successfully - ID: %d\n", prefixFilterListID)
	} else {
		fmt.Println("Prefix filter list update request was not successful")
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
	err = printMCRs([]*megaport.MCR{mcr}, outputFormat)
	if err != nil {
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
		confirmation, err := prompt(confirmMsg)
		if err != nil {
			return err
		}

		if confirmation != "y" && confirmation != "Y" {
			fmt.Println("Deletion cancelled")
			return nil
		}
	}

	// Create delete request
	deleteRequest := &megaport.DeleteMCRRequest{
		MCRID:     mcrUID,
		DeleteNow: deleteNow,
	}

	// Delete the MCR
	resp, err := deleteMCRFunc(ctx, client, deleteRequest)
	if err != nil {
		return fmt.Errorf("error deleting MCR: %v", err)
	}

	if resp.IsDeleting {
		fmt.Printf("MCR %s deleted successfully\n", mcrUID)
		if deleteNow {
			fmt.Println("The MCR will be deleted immediately")
		} else {
			fmt.Println("The MCR will be deleted at the end of the current billing period")
		}
	} else {
		fmt.Println("MCR deletion request was not successful")
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
		return fmt.Errorf("error logging in: %v", err)
	}

	// Retrieve the MCR UID from the command line arguments
	mcrUID := args[0]

	// Restore the MCR
	resp, err := restoreMCRFunc(ctx, client, mcrUID)
	if err != nil {
		return fmt.Errorf("error restoring MCR: %v", err)
	}

	if resp.IsRestored {
		fmt.Printf("MCR %s restored successfully\n", mcrUID)
	} else {
		fmt.Println("MCR restoration request was not successful")
	}

	return nil
}

func ListMCRPrefixFilterLists(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Log into the Megaport API using the provided credentials
	client, err := Login(ctx)
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
	err = printOutput(prefixFilterLists, outputFormat)
	if err != nil {
		return fmt.Errorf("error printing prefix filter lists: %v", err)
	}
	return nil
}

func GetMCRPrefixFilterList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Log into the Megaport API using the provided credentials
	client, err := Login(ctx)
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
	output, err := ToPrefixFilterListOutput(prefixFilterList)
	if err != nil {
		return fmt.Errorf("error converting prefix filter list: %v", err)
	}

	// Print the prefix filter list details using the desired output format
	err = printOutput([]PrefixFilterListOutput{output}, outputFormat)
	if err != nil {
		return fmt.Errorf("error printing prefix filter list: %v", err)
	}
	return nil
}

func DeleteMCRPrefixFilterList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Log into the Megaport API using the provided credentials
	client, err := Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %v", err)
	}

	// Retrieve the MCR UID and prefix filter list ID from the command line arguments
	mcrUID := args[0]
	prefixFilterListID, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid prefix filter list ID: %v", err)
	}

	// Call the DeleteMCRPrefixFilterList method
	resp, err := deleteMCRPrefixFilterListFunc(ctx, client, mcrUID, prefixFilterListID)
	if err != nil {
		return fmt.Errorf("error deleting prefix filter list: %v", err)
	}

	if resp.IsDeleted {
		fmt.Printf("Prefix filter list deleted successfully - ID: %d\n", prefixFilterListID)
	} else {
		fmt.Println("Prefix filter list deletion request was not successful")
	}

	return nil
}
