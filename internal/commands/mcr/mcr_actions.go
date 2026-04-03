package mcr

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/megaport/megaport-cli/internal/base/exitcodes"
	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/commands/config"
	"github.com/megaport/megaport-cli/internal/utils"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

func exportMCRConfig(mcr *megaport.MCR) map[string]interface{} {
	m := map[string]interface{}{
		"name":       mcr.Name,
		"term":       mcr.ContractTermMonths,
		"portSpeed":  mcr.PortSpeed,
		"locationId": mcr.LocationID,
	}
	if mcr.Resources.VirtualRouter.ASN != 0 {
		m["mcrAsn"] = mcr.Resources.VirtualRouter.ASN
	}
	if mcr.DiversityZone != "" {
		m["diversityZone"] = mcr.DiversityZone
	}
	if mcr.CostCentre != "" {
		m["costCentre"] = mcr.CostCentre
	}
	return m
}

func buildMCRRequest(cmd *cobra.Command, noColor bool) (*megaport.BuyMCRRequest, error) {
	interactive, _ := cmd.Flags().GetBool("interactive")
	jsonStr, _ := cmd.Flags().GetString("json")
	jsonFile, _ := cmd.Flags().GetString("json-file")

	flagsProvided := cmd.Flags().Changed("name") || cmd.Flags().Changed("term") ||
		cmd.Flags().Changed("port-speed") || cmd.Flags().Changed("location-id") ||
		cmd.Flags().Changed("mcr-asn")

	if jsonStr != "" || jsonFile != "" {
		output.PrintInfo("Using JSON input", noColor)
		req, err := processJSONMCRInput(jsonStr, jsonFile)
		if err != nil {
			output.PrintError("Failed to process JSON input: %v", noColor, err)
			return nil, err
		}
		return req, nil
	} else if flagsProvided {
		output.PrintInfo("Using flag input", noColor)
		req, err := processFlagMCRInput(cmd)
		if err != nil {
			output.PrintError("Failed to process flag input: %v", noColor, err)
			return nil, err
		}
		return req, nil
	} else if interactive {
		req, err := promptForMCRDetails(noColor)
		if err != nil {
			return nil, err
		}
		return req, nil
	}
	output.PrintError("No input provided, use --interactive, --json, or flags to specify MCR details", noColor)
	return nil, fmt.Errorf("no input provided, use --interactive, --json, or flags to specify MCR details")
}

func BuyMCR(cmd *cobra.Command, args []string, noColor bool) error {
	ctx, cancel := utils.ContextFromCmdWithDefault(cmd, 15*time.Minute)
	defer cancel()

	req, err := buildMCRRequest(cmd, noColor)
	if err != nil {
		return err
	}

	noWait, _ := cmd.Flags().GetBool("no-wait")
	if !noWait {
		req.WaitForProvision = true
		req.WaitForTime = 10 * time.Minute
	}

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

	jsonStr, _ := cmd.Flags().GetString("json")
	jsonFile, _ := cmd.Flags().GetString("json-file")
	yes, _ := cmd.Flags().GetBool("yes")
	if !yes && jsonStr == "" && jsonFile == "" {
		details := []utils.BuyConfirmDetail{
			{Key: "Name", Value: req.Name},
			{Key: "Term", Value: fmt.Sprintf("%d months", req.Term)},
			{Key: "Port Speed", Value: fmt.Sprintf("%d Mbps", req.PortSpeed)},
			{Key: "Location ID", Value: strconv.Itoa(req.LocationID)},
		}
		if req.MCRAsn != 0 {
			details = append(details, utils.BuyConfirmDetail{Key: "ASN", Value: strconv.Itoa(req.MCRAsn)})
		}
		if !utils.BuyConfirmPrompt("MCR", details, noColor) {
			output.PrintInfo("Purchase cancelled", noColor)
			return exitcodes.New(exitcodes.Cancelled, fmt.Errorf("cancelled by user"))
		}
	}

	var buySpinner *output.Spinner
	if req.WaitForProvision {
		buySpinner = output.PrintResourceProvisioning("MCR", req.Name, noColor)
	} else {
		buySpinner = output.PrintResourceCreating("MCR", req.Name, noColor)
	}
	resp, err := buyMCRFunc(ctx, client, req)
	buySpinner.Stop()

	if err != nil {
		output.PrintError("Error buying MCR: %v", noColor, err)
		return err
	}

	output.PrintSuccess("MCR created %s", noColor, resp.TechnicalServiceUID)
	return nil
}

func ValidateMCR(cmd *cobra.Command, args []string, noColor bool) error {
	ctx, cancel := utils.ContextFromCmd(cmd)
	defer cancel()

	req, err := buildMCRRequest(cmd, noColor)
	if err != nil {
		return err
	}

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

	output.PrintSuccess("MCR validation passed", noColor)
	return nil
}

func UpdateMCR(cmd *cobra.Command, args []string, noColor bool) error {
	ctx, cancel := utils.ContextFromCmdWithDefault(cmd, 15*time.Minute)
	defer cancel()

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
	ctx, cancel := utils.ContextFromCmd(cmd)
	defer cancel()

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
	ctx, cancel := utils.ContextFromCmd(cmd)
	defer cancel()

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
	output.SetOutputFormat(outputFormat)

	watch, _ := cmd.Flags().GetBool("watch")
	if watch {
		return watchGetMCR(cmd, args, noColor, outputFormat)
	}

	ctx, cancel := utils.ContextFromCmd(cmd)
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
		err = utils.WrapAPIError(err, "MCR", mcrUID)
		output.PrintError("Error getting MCR: %v", noColor, err)
		return fmt.Errorf("error getting MCR: %w", err)
	}

	export, _ := cmd.Flags().GetBool("export")
	if export {
		cfg := exportMCRConfig(mcr)
		jsonBytes, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return fmt.Errorf("error marshaling export config: %v", err)
		}
		fmt.Println(string(jsonBytes))
		return nil
	}

	err = printMCRs([]*megaport.MCR{mcr}, outputFormat, noColor)
	if err != nil {
		return fmt.Errorf("error printing MCRs: %v", err)
	}
	return nil
}

func watchGetMCR(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	interval, _ := cmd.Flags().GetDuration("interval")

	ctx := context.Background()
	client, err := config.Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %v", err)
	}

	mcrUID := args[0]
	cfg := utils.WatchConfig{
		Interval:     interval,
		NoColor:      noColor,
		OutputFormat: outputFormat,
		ResourceType: "MCR",
		ResourceUID:  mcrUID,
	}

	return utils.WatchLoop(ctx, cfg, func(pollCtx context.Context) (string, error) {
		mcr, err := getMCRFunc(pollCtx, client, mcrUID)
		if err != nil {
			return "", err
		}
		err = printMCRs([]*megaport.MCR{mcr}, outputFormat, noColor)
		return mcr.ProvisioningStatus, err
	})
}

func DeleteMCR(cmd *cobra.Command, args []string, noColor bool) error {
	ctx, cancel := utils.ContextFromCmd(cmd)
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
		confirmMsg := "Are you sure you want to delete MCR " + mcrUID + "? "
		if !utils.ConfirmPrompt(confirmMsg, noColor) {
			output.PrintInfo("Deletion cancelled", noColor)
			return exitcodes.New(exitcodes.Cancelled, fmt.Errorf("cancelled by user"))
		}
	}

	safeDelete, err := cmd.Flags().GetBool("safe-delete")
	if err != nil {
		return fmt.Errorf("failed to get safe-delete flag: %w", err)
	}

	deleteRequest := &megaport.DeleteMCRRequest{
		MCRID:      mcrUID,
		DeleteNow:  deleteNow,
		SafeDelete: safeDelete,
	}

	spinner := output.PrintResourceDeleting("MCR", mcrUID, noColor)

	resp, err := deleteMCRFunc(ctx, client, deleteRequest)

	spinner.Stop()

	if err != nil {
		err = utils.WrapAPIError(err, "MCR", mcrUID)
		return fmt.Errorf("error deleting MCR: %w", err)
	}

	if resp.IsDeleting {
		output.PrintResourceDeleted("MCR", mcrUID, deleteNow, noColor)
	} else {
		output.PrintError("MCR deletion request was not successful", noColor)
	}

	return nil
}

func RestoreMCR(cmd *cobra.Command, args []string, noColor bool) error {
	ctx, cancel := utils.ContextFromCmd(cmd)
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

func LockMCR(cmd *cobra.Command, args []string, noColor bool) error {
	ctx, cancel := utils.ContextFromCmd(cmd)
	defer cancel()

	client, err := config.Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %v", err)
	}

	mcrUID := args[0]

	output.PrintInfo("Locking MCR %s...", noColor, mcrUID)

	_, err = lockMCRFunc(ctx, client, mcrUID)
	if err != nil {
		return fmt.Errorf("error locking MCR: %v", err)
	}

	output.PrintSuccess("MCR %s locked successfully", noColor, mcrUID)
	return nil
}

func UnlockMCR(cmd *cobra.Command, args []string, noColor bool) error {
	ctx, cancel := utils.ContextFromCmd(cmd)
	defer cancel()

	client, err := config.Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %v", err)
	}

	mcrUID := args[0]

	output.PrintInfo("Unlocking MCR %s...", noColor, mcrUID)

	_, err = unlockMCRFunc(ctx, client, mcrUID)
	if err != nil {
		return fmt.Errorf("error unlocking MCR: %v", err)
	}

	output.PrintSuccess("MCR %s unlocked successfully", noColor, mcrUID)
	return nil
}

func ListMCRPrefixFilterLists(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	// Set output format for proper JSON mode handling
	output.SetOutputFormat(outputFormat)

	ctx, cancel := utils.ContextFromCmd(cmd)
	defer cancel()

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
	// Set output format for proper JSON mode handling
	output.SetOutputFormat(outputFormat)

	ctx, cancel := utils.ContextFromCmd(cmd)
	defer cancel()

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
	ctx, cancel := utils.ContextFromCmd(cmd)
	defer cancel()

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
	output.SetOutputFormat(outputFormat)
	ctx, cancel := utils.ContextFromCmd(cmd)
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

	limit, _ := cmd.Flags().GetInt("limit")
	if limit < 0 {
		return fmt.Errorf("--limit must be a non-negative integer")
	}
	if limit > 0 && len(filteredMCRs) > limit {
		filteredMCRs = filteredMCRs[:limit]
	}

	if len(filteredMCRs) == 0 {
		if outputFormat == utils.FormatTable {
			output.PrintInfo("No MCRs found. Create one with 'megaport mcr buy'.", noColor)
		}
		return nil
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
	return utils.ListResourceTags("MCR", mcrUID, noColor, outputFormat, func(ctx context.Context, uid string) (map[string]string, error) {
		client, err := config.LoginFunc(ctx)
		if err != nil {
			return nil, err
		}
		return client.MCRService.ListMCRResourceTags(ctx, uid)
	})
}

func UpdateMCRResourceTags(cmd *cobra.Command, args []string, noColor bool) error {
	mcrUID := args[0]
	var client *megaport.Client
	login := func(ctx context.Context) error {
		var err error
		client, err = config.LoginFunc(ctx)
		return err
	}
	return utils.UpdateResourceTags(utils.UpdateTagsOptions{
		ResourceType: "MCR",
		UID:          mcrUID,
		NoColor:      noColor,
		Cmd:          cmd,
		ListFunc: func(ctx context.Context, uid string) (map[string]string, error) {
			if err := login(ctx); err != nil {
				return nil, err
			}
			return client.MCRService.ListMCRResourceTags(ctx, uid)
		},
		UpdateFunc: func(ctx context.Context, uid string, tags map[string]string) error {
			return client.MCRService.UpdateMCRResourceTags(ctx, uid, tags)
		},
	})
}

func GetMCRStatus(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	output.SetOutputFormat(outputFormat)

	watch, _ := cmd.Flags().GetBool("watch")
	if watch {
		return watchMCRStatus(cmd, args, noColor, outputFormat)
	}

	ctx, cancel := utils.ContextFromCmd(cmd)
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

	if mcr == nil {
		output.PrintError("No MCR found with UID: %s", noColor, mcrUID)
		return fmt.Errorf("no MCR found with UID: %s", mcrUID)
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

func watchMCRStatus(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	interval, _ := cmd.Flags().GetDuration("interval")

	ctx := context.Background()
	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("error logging in: %v", err)
	}

	mcrUID := args[0]
	cfg := utils.WatchConfig{
		Interval:     interval,
		NoColor:      noColor,
		OutputFormat: outputFormat,
		ResourceType: "MCR",
		ResourceUID:  mcrUID,
	}

	return utils.WatchLoop(ctx, cfg, func(pollCtx context.Context) (string, error) {
		mcr, err := client.MCRService.GetMCR(pollCtx, mcrUID)
		if err != nil {
			return "", err
		}
		if mcr == nil {
			return "", fmt.Errorf("no MCR found with UID: %s", mcrUID)
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
		err = output.PrintOutput(status, outputFormat, noColor)
		return mcr.ProvisioningStatus, err
	})
}
