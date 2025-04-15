package mcr

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/commands/config"
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

	// Start a spinner to show progress while getting the MCR details
	spinner := output.PrintResourceGetting("MCR", mcrUID, noColor)

	// Use the API client to get the MCR details based on the provided UID.
	mcr, err := getMCRFunc(ctx, client, mcrUID)

	// Stop the spinner
	spinner.Stop()

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
		confirmation, err := utils.ResourcePrompt("mcr", confirmMsg, noColor)
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

	// Start the spinner to show progress during deletion
	spinner := output.PrintResourceDeleting("MCR", mcrUID, noColor)

	// Delete the MCR
	resp, err := deleteMCRFunc(ctx, client, deleteRequest)

	// Stop the spinner
	spinner.Stop()

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

	// Start a spinner to show progress while listing prefix filter lists
	spinner := output.PrintResourceListing("Prefix filter list", noColor)

	// Call the ListMCRPrefixFilterLists method
	prefixFilterLists, err := listMCRPrefixFilterListsFunc(ctx, client, mcrUID)

	// Stop the spinner
	spinner.Stop()

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

	// Start a spinner to show progress while getting the prefix filter list details
	spinner := output.PrintResourceGetting("Prefix filter list", fmt.Sprintf("%d", prefixFilterListID), noColor)

	// Call the GetMCRPrefixFilterList method
	prefixFilterList, err := getMCRPrefixFilterListFunc(ctx, client, mcrUID, prefixFilterListID)

	// Stop the spinner
	spinner.Stop()

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

	// Start the spinner to show progress during deletion
	spinner := output.PrintResourceDeleting("Prefix filter list", fmt.Sprintf("%d", prefixFilterListID), noColor)

	// Call the DeleteMCRPrefixFilterList method
	resp, err := deleteMCRPrefixFilterListFunc(ctx, client, mcrUID, prefixFilterListID)

	// Stop the spinner
	spinner.Stop()

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

// ListMCRs retrieves and prints a list of MCRs based on the provided filters.
func ListMCRs(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Log into Megaport API
	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("error logging in: %v", err)
	}

	// Get filter values from flags
	locationID, _ := cmd.Flags().GetInt("location-id")
	portSpeed, _ := cmd.Flags().GetInt("port-speed")
	mcrName, _ := cmd.Flags().GetString("mcr-name")
	includeInactive, _ := cmd.Flags().GetBool("include-inactive")

	// Create a ListMCRsRequest
	req := &megaport.ListMCRsRequest{
		IncludeInactive: includeInactive,
	}

	// Start the spinner to show progress while retrieving MCRs
	spinner := output.PrintResourceListing("MCR", noColor)

	// Get all MCRs
	mcrs, err := client.MCRService.ListMCRs(ctx, req)

	// Stop the spinner
	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to list MCRs: %v", noColor, err)
		return fmt.Errorf("error listing MCRs: %v", err)
	}

	// Apply manual filtering for inactive MCRs since our mock doesn't implement this behavior
	var activeMCRs []*megaport.MCR
	if !includeInactive {
		for _, mcr := range mcrs {
			if mcr != nil &&
				mcr.ProvisioningStatus != "DECOMMISSIONED" &&
				mcr.ProvisioningStatus != "CANCELLED" &&
				mcr.ProvisioningStatus != "DECOMMISSIONING" {
				activeMCRs = append(activeMCRs, mcr)
			}
		}
		mcrs = activeMCRs
	}

	// Apply additional filters
	filteredMCRs := filterMCRs(mcrs, locationID, portSpeed, mcrName)

	if len(filteredMCRs) == 0 {
		output.PrintWarning("No MCRs found matching the specified filters", noColor)
	}

	// Print MCRs with current output format
	err = printMCRs(filteredMCRs, outputFormat, noColor)
	if err != nil {
		output.PrintError("Failed to print MCRs: %v", noColor, err)
		return fmt.Errorf("error printing MCRs: %v", err)
	}
	return nil
}

// ListMCRResourceTags retrieves and displays resource tags for an MCR
func ListMCRResourceTags(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	mcrUID := args[0]

	ctx := context.Background()

	// Login to the Megaport API
	client, err := config.LoginFunc(ctx)
	if err != nil {
		return err
	}

	spinner := output.PrintListingResourceTags("MCR", mcrUID, noColor)

	// Get the resource tags for the MCR
	tagsMap, err := client.MCRService.ListMCRResourceTags(ctx, mcrUID)

	spinner.Stop()
	if err != nil {
		output.PrintError("Error getting resource tags for MCR %s: %v", noColor, mcrUID, err)
		return fmt.Errorf("error getting resource tags for MCR %s: %v", mcrUID, err)
	}

	// Convert map to slice of ResourceTag for output
	tags := make([]output.ResourceTag, 0, len(tagsMap))
	for k, v := range tagsMap {
		tags = append(tags, output.ResourceTag{Key: k, Value: v})
	}

	// Sort tags by key for consistent output
	sort.Slice(tags, func(i, j int) bool {
		return tags[i].Key < tags[j].Key
	})

	// Use the existing PrintOutput function
	return output.PrintOutput(tags, outputFormat, noColor)
}

// UpdateMCRResourceTags updates resource tags for an MCR
func UpdateMCRResourceTags(cmd *cobra.Command, args []string, noColor bool) error {
	mcrUID := args[0]

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Login to the Megaport API
	client, err := config.LoginFunc(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}

	existingTags, err := client.MCRService.ListMCRResourceTags(ctx, mcrUID)
	if err != nil {
		output.PrintError("Failed to get existing resource tags: %v", noColor, err)
		return fmt.Errorf("failed to get existing resource tags: %v", err)
	}

	// Check if we're in interactive mode
	interactive, _ := cmd.Flags().GetBool("interactive")

	// Variables to store tags
	var resourceTags map[string]string

	if interactive {
		// Interactive mode: prompt for tags
		resourceTags, err = utils.UpdateResourceTagsPrompt(existingTags, noColor)
		if err != nil {
			output.PrintError("Failed to update resource tags", noColor, err)
			return err
		}
	} else {
		// Check if we have JSON input (only allowing standard JSON inputs)
		jsonStr, _ := cmd.Flags().GetString("json")
		jsonFile, _ := cmd.Flags().GetString("json-file")

		// Process JSON input if provided
		if jsonStr != "" {
			// Parse JSON string
			if err := json.Unmarshal([]byte(jsonStr), &resourceTags); err != nil {
				output.PrintError("Failed to parse JSON: %v", noColor, err)
				return fmt.Errorf("error parsing JSON: %v", err)
			}
		} else if jsonFile != "" {
			// Read from file
			jsonData, err := os.ReadFile(jsonFile)
			if err != nil {
				output.PrintError("Failed to read JSON file: %v", noColor, err)
				return fmt.Errorf("error reading JSON file: %v", err)
			}

			// Parse JSON from file
			if err := json.Unmarshal(jsonData, &resourceTags); err != nil {
				output.PrintError("Failed to parse JSON file: %v", noColor, err)
				return fmt.Errorf("error parsing JSON file: %v", err)
			}
		} else {
			output.PrintError("No input provided for tags", noColor)
			return fmt.Errorf("no input provided, use --interactive, --json, or --json-file to specify resource tags")
		}
	}

	// If we got here, we have tags to update
	if len(resourceTags) == 0 {
		fmt.Println("No tags provided. The MCR will have all existing tags removed.")
	}

	// Start spinner for updating resource tags
	spinner := output.PrintResourceUpdating("MCR-Resource-Tags", mcrUID, noColor)

	// Update tags
	err = client.MCRService.UpdateMCRResourceTags(ctx, mcrUID, resourceTags)

	// Stop spinner
	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to update resource tags: %v", noColor, err)
		return fmt.Errorf("failed to update resource tags: %v", err)
	}

	fmt.Printf("Resource tags updated for MCR %s\n", mcrUID)
	return nil
}
