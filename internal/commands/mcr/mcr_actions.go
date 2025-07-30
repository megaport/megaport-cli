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

	interactive, _ := cmd.Flags().GetBool("interactive")
	jsonStr, _ := cmd.Flags().GetString("json")
	jsonFile, _ := cmd.Flags().GetString("json-file")

	flagsProvided := cmd.Flags().Changed("name") || cmd.Flags().Changed("term") ||
		cmd.Flags().Changed("port-speed") || cmd.Flags().Changed("location-id") ||
		cmd.Flags().Changed("mcr-asn")

	var req *megaport.BuyMCRRequest
	var err error

	if jsonStr != "" || jsonFile != "" {
		output.PrintInfo("Using JSON input", noColor)
		req, err = processJSONMCRInput(jsonStr, jsonFile)
		if err != nil {
			output.PrintError("Failed to process JSON input: %v", noColor, err)
			return err
		}
	} else if flagsProvided {
		output.PrintInfo("Using flag input", noColor)
		req, err = processFlagMCRInput(cmd)
		if err != nil {
			output.PrintError("Failed to process flag input: %v", noColor, err)
			return err
		}
	} else if interactive {
		req, err = promptForMCRDetails(noColor)
		if err != nil {
			return err
		}
	} else {
		output.PrintError("No input provided, use --interactive, --json, or flags to specify MCR details", noColor)
		return fmt.Errorf("no input provided, use --interactive, --json, or flags to specify MCR details")
	}

	req.WaitForProvision = true
	req.WaitForTime = 10 * time.Minute

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Error logging in: %v", noColor, err)
		return err
	}

	spinner := output.PrintResourceValidating("MCR", noColor)
	err = client.MCRService.ValidateMCROrder(ctx, req)
	spinner.Stop()

	if err != nil {
		output.PrintError("Error validating MCR order: %v", noColor, err)
		return err
	}

	buySpinner := output.PrintResourceCreating("MCR", req.Name, noColor)
	resp, err := buyMCRFunc(ctx, client, req)
	buySpinner.Stop()

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

	mcrUID := args[0]

	interactive, _ := cmd.Flags().GetBool("interactive")
	jsonStr, _ := cmd.Flags().GetString("json")
	jsonFile, _ := cmd.Flags().GetString("json-file")

	flagsProvided := cmd.Flags().Changed("name") || cmd.Flags().Changed("cost-centre") ||
		cmd.Flags().Changed("marketplace-visibility") || cmd.Flags().Changed("term")

	var req *megaport.ModifyMCRRequest
	var err error

	if jsonStr != "" || jsonFile != "" {
		output.PrintInfo("Using JSON input", noColor)
		req, err = processJSONUpdateMCRInput(jsonStr, jsonFile)
		if err != nil {
			output.PrintError("Failed to process JSON input: %v", noColor, err)
			return err
		}
		req.MCRID = mcrUID
	} else if flagsProvided {
		output.PrintInfo("Using flag input", noColor)
		req, err = processFlagUpdateMCRInput(cmd, mcrUID)
		if err != nil {
			output.PrintError("Failed to process flag input: %v", noColor, err)
			return err
		}
	} else if interactive {
		req, err = promptForUpdateMCRDetails(mcrUID, noColor)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("at least one field must be updated")
	}

	req.WaitForUpdate = true
	req.WaitForTime = 10 * time.Minute

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
	updateSpinner := output.PrintResourceUpdating("MCR", mcrUID, noColor)
	resp, err := updateMCRFunc(ctx, client, req)
	updateSpinner.Stop()

	if err != nil {
		output.PrintError("Error updating MCR: %v", noColor, err)
		return err
	}

	if !resp.IsUpdated {
		output.PrintError("MCR update request was not successful", noColor)
		return fmt.Errorf("MCR update request was not successful")
	}

	updatedMCR, err := getMCRFunc(ctx, client, mcrUID)
	if err != nil {
		output.PrintError("MCR was updated but failed to retrieve updated details: %v", noColor, err)
		output.PrintResourceUpdated("MCR", mcrUID, noColor)
		return nil
	}

	output.PrintResourceUpdated("MCR", mcrUID, noColor)

	displayMCRChanges(originalMCR, updatedMCR, noColor)

	return nil
}

func CreateMCRPrefixFilterList(cmd *cobra.Command, args []string, noColor bool) error {
	ctx := context.Background()

	if len(args) == 0 {
		return fmt.Errorf("mcr UID is required")
	}

	mcrUID := args[0]

	interactive, _ := cmd.Flags().GetBool("interactive")
	jsonStr, _ := cmd.Flags().GetString("json")
	jsonFile, _ := cmd.Flags().GetString("json-file")

	flagsProvided := cmd.Flags().Changed("description") || cmd.Flags().Changed("address-family") ||
		cmd.Flags().Changed("entries")

	var req *megaport.CreateMCRPrefixFilterListRequest
	var err error

	if jsonStr != "" || jsonFile != "" {
		output.PrintInfo("Using JSON input", noColor)
		req, err = processJSONPrefixFilterListInput(jsonStr, jsonFile, mcrUID)
		if err != nil {
			output.PrintError("Failed to process JSON input: %v", noColor, err)
			return err
		}
	} else if flagsProvided {
		output.PrintInfo("Using flag input", noColor)
		req, err = processFlagPrefixFilterListInput(cmd, mcrUID)
		if err != nil {
			output.PrintError("Failed to process flag input: %v", noColor, err)
			return err
		}
	} else if interactive {
		req, err = promptForPrefixFilterListDetails(mcrUID, noColor)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("no input provided, use --interactive, --json, or flags to specify prefix filter list details")
	}

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Error logging in: %v", noColor, err)
		return err
	}
	spinner := output.PrintResourceCreating("Prefix Filter List", req.PrefixFilterList.Description, noColor)
	resp, err := createMCRPrefixFilterListFunc(ctx, client, req)
	spinner.Stop()

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

	mcrUID := args[0]
	prefixFilterListID, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid prefix filter list ID: %v", err)
	}

	interactive, _ := cmd.Flags().GetBool("interactive")
	jsonStr, _ := cmd.Flags().GetString("json")
	jsonFile, _ := cmd.Flags().GetString("json-file")

	flagsProvided := cmd.Flags().Changed("description") || cmd.Flags().Changed("address-family") ||
		cmd.Flags().Changed("entries")

	var prefixFilterList *megaport.MCRPrefixFilterList
	var getErr error

	if jsonStr != "" || jsonFile != "" {
		output.PrintInfo("Using JSON input", noColor)
		prefixFilterList, getErr = processJSONUpdatePrefixFilterListInput(jsonStr, jsonFile, mcrUID, prefixFilterListID)
		if getErr != nil {
			output.PrintError("Failed to process JSON input: %v", noColor, getErr)
			return getErr
		}
	} else if flagsProvided {
		output.PrintInfo("Using flag input", noColor)
		prefixFilterList, getErr = processFlagUpdatePrefixFilterListInput(cmd, mcrUID, prefixFilterListID)
		if getErr != nil {
			output.PrintError("Failed to process flag input: %v", noColor, getErr)
			return getErr
		}
	} else if interactive {
		prefixFilterList, getErr = promptForUpdatePrefixFilterListDetails(mcrUID, prefixFilterListID, noColor)
		if getErr != nil {
			return getErr
		}
	} else {
		return fmt.Errorf("at least one field must be updated")
	}

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Error logging in: %v", noColor, err)
		return err
	}
	spinner := output.PrintResourceUpdating("Prefix Filter List", fmt.Sprintf("%d", prefixFilterListID), noColor)
	resp, err := modifyMCRPrefixFilterListFunc(ctx, client, mcrUID, prefixFilterListID, prefixFilterList)
	spinner.Stop()

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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := config.Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %v", err)
	}

	mcrUID := args[0]

	spinner := output.PrintResourceGetting("MCR", mcrUID, noColor)

	mcr, err := getMCRFunc(ctx, client, mcrUID)

	spinner.Stop()

	if err != nil {
		output.PrintError("Error getting MCR: %v", noColor, err)
		return fmt.Errorf("error getting MCR: %v", err)
	}

	err = printMCRs([]*megaport.MCR{mcr}, outputFormat, noColor)
	if err != nil {
		return fmt.Errorf("error printing MCRs: %v", err)
	}
	return nil
}

func DeleteMCR(cmd *cobra.Command, args []string, noColor bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := config.Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %v", err)
	}

	mcrUID := args[0]

	deleteNow, err := cmd.Flags().GetBool("now")
	if err != nil {
		return err
	}

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

	deleteRequest := &megaport.DeleteMCRRequest{
		MCRID:     mcrUID,
		DeleteNow: deleteNow,
	}

	spinner := output.PrintResourceDeleting("MCR", mcrUID, noColor)

	resp, err := deleteMCRFunc(ctx, client, deleteRequest)

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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := config.Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %v", err)
	}

	mcrUID := args[0]

	output.PrintInfo("Restoring MCR %s...", noColor, mcrUID)

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

	client, err := config.Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %v", err)
	}

	mcrUID := args[0]

	spinner := output.PrintResourceListing("Prefix filter list", noColor)

	prefixFilterLists, err := listMCRPrefixFilterListsFunc(ctx, client, mcrUID)

	spinner.Stop()

	if err != nil {
		return fmt.Errorf("error listing prefix filter lists: %v", err)
	}

	err = output.PrintOutput(prefixFilterLists, outputFormat, noColor)
	if err != nil {
		return fmt.Errorf("error printing prefix filter lists: %v", err)
	}
	return nil
}

func GetMCRPrefixFilterList(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	ctx := context.Background()

	client, err := config.Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %v", err)
	}

	mcrUID := args[0]
	prefixFilterListID, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid prefix filter list ID: %v", err)
	}

	spinner := output.PrintResourceGetting("Prefix filter list", fmt.Sprintf("%d", prefixFilterListID), noColor)

	prefixFilterList, err := getMCRPrefixFilterListFunc(ctx, client, mcrUID, prefixFilterListID)

	spinner.Stop()

	if err != nil {
		return fmt.Errorf("error getting prefix filter list: %v", err)
	}

	op, err := ToPrefixFilterListOutput(prefixFilterList)
	if err != nil {
		return fmt.Errorf("error converting prefix filter list: %v", err)
	}

	err = output.PrintOutput([]PrefixFilterListOutput{op}, outputFormat, noColor)
	if err != nil {
		return fmt.Errorf("error printing prefix filter list: %v", err)
	}
	return nil
}

func DeleteMCRPrefixFilterList(cmd *cobra.Command, args []string, noColor bool) error {
	ctx := context.Background()

	client, err := config.Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %v", err)
	}

	mcrUID := args[0]
	prefixFilterListID, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid prefix filter list ID: %v", err)
	}

	spinner := output.PrintResourceDeleting("Prefix filter list", fmt.Sprintf("%d", prefixFilterListID), noColor)

	resp, err := deleteMCRPrefixFilterListFunc(ctx, client, mcrUID, prefixFilterListID)

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

func ListMCRs(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("error logging in: %v", err)
	}

	locationID, _ := cmd.Flags().GetInt("location-id")
	portSpeed, _ := cmd.Flags().GetInt("port-speed")
	mcrName, _ := cmd.Flags().GetString("name")
	includeInactive, _ := cmd.Flags().GetBool("include-inactive")

	req := &megaport.ListMCRsRequest{
		IncludeInactive: includeInactive,
	}

	spinner := output.PrintResourceListing("MCR", noColor)

	mcrs, err := client.MCRService.ListMCRs(ctx, req)

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to list MCRs: %v", noColor, err)
		return fmt.Errorf("error listing MCRs: %v", err)
	}

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

	filteredMCRs := filterMCRs(mcrs, locationID, portSpeed, mcrName)

	if len(filteredMCRs) == 0 {
		output.PrintWarning("No MCRs found matching the specified filters", noColor)
	}

	err = printMCRs(filteredMCRs, outputFormat, noColor)
	if err != nil {
		output.PrintError("Failed to print MCRs: %v", noColor, err)
		return fmt.Errorf("error printing MCRs: %v", err)
	}
	return nil
}

func ListMCRResourceTags(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	mcrUID := args[0]

	ctx := context.Background()

	client, err := config.LoginFunc(ctx)
	if err != nil {
		return err
	}

	tagsMap, err := client.MCRService.ListMCRResourceTags(ctx, mcrUID)

	if err != nil {
		output.PrintError("Error getting resource tags for MCR %s: %v", noColor, mcrUID, err)
		return fmt.Errorf("error getting resource tags for MCR %s: %v", mcrUID, err)
	}

	tags := make([]output.ResourceTag, 0, len(tagsMap))
	for k, v := range tagsMap {
		tags = append(tags, output.ResourceTag{Key: k, Value: v})
	}

	sort.Slice(tags, func(i, j int) bool {
		return tags[i].Key < tags[j].Key
	})

	return output.PrintOutput(tags, outputFormat, noColor)
}

func UpdateMCRResourceTags(cmd *cobra.Command, args []string, noColor bool) error {
	mcrUID := args[0]

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

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

	interactive, _ := cmd.Flags().GetBool("interactive")

	var resourceTags map[string]string

	if interactive {
		resourceTags, err = utils.UpdateResourceTagsPrompt(existingTags, noColor)
		if err != nil {
			output.PrintError("Failed to update resource tags", noColor, err)
			return err
		}
	} else {
		jsonStr, _ := cmd.Flags().GetString("json")
		jsonFile, _ := cmd.Flags().GetString("json-file")

		if jsonStr != "" {
			if err := json.Unmarshal([]byte(jsonStr), &resourceTags); err != nil {
				output.PrintError("Failed to parse JSON: %v", noColor, err)
				return fmt.Errorf("error parsing JSON: %v", err)
			}
		} else if jsonFile != "" {
			jsonData, err := os.ReadFile(jsonFile)
			if err != nil {
				output.PrintError("Failed to read JSON file: %v", noColor, err)
				return fmt.Errorf("error reading JSON file: %v", err)
			}

			if err := json.Unmarshal(jsonData, &resourceTags); err != nil {
				output.PrintError("Failed to parse JSON file: %v", noColor, err)
				return fmt.Errorf("error parsing JSON file: %v", err)
			}
		} else {
			output.PrintError("No input provided for tags", noColor)
			return fmt.Errorf("no input provided, use --interactive, --json, or --json-file to specify resource tags")
		}
	}

	if len(resourceTags) == 0 {
		fmt.Println("No tags provided. The MCR will have all existing tags removed.")
	}

	spinner := output.PrintResourceUpdating("MCR-Resource-Tags", mcrUID, noColor)

	err = client.MCRService.UpdateMCRResourceTags(ctx, mcrUID, resourceTags)

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to update resource tags: %v", noColor, err)
		return fmt.Errorf("failed to update resource tags: %v", err)
	}

	fmt.Printf("Resource tags updated for MCR %s\n", mcrUID)
	return nil
}

func GetMCRStatus(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("error logging in: %v", err)
	}

	mcrUID := args[0]

	spinner := output.PrintResourceGetting("MCR", mcrUID, noColor)

	mcr, err := client.MCRService.GetMCR(ctx, mcrUID)

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to get MCR status: %v", noColor, err)
		return fmt.Errorf("error getting MCR status: %v", err)
	}

	status := []MCRStatus{
		{
			UID:    mcr.UID,
			Name:   mcr.Name,
			Status: mcr.ProvisioningStatus,
			ASN:    mcr.Resources.VirtualRouter.ASN,
			Speed:  mcr.PortSpeed,
		},
	}

	return output.PrintOutput(status, outputFormat, noColor)
}
