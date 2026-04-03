package mve

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
	"github.com/megaport/megaport-cli/internal/validation"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

func exportMVEConfig(mve *megaport.MVE) map[string]interface{} {
	m := map[string]interface{}{
		"name":       mve.Name,
		"term":       mve.ContractTermMonths,
		"locationId": mve.LocationID,
	}
	if mve.DiversityZone != "" {
		m["diversityZone"] = mve.DiversityZone
	}
	if mve.CostCentre != "" {
		m["costCentre"] = mve.CostCentre
	}
	if len(mve.NetworkInterfaces) > 0 {
		vnics := make([]map[string]interface{}, 0, len(mve.NetworkInterfaces))
		for _, ni := range mve.NetworkInterfaces {
			vnic := map[string]interface{}{
				"description": ni.Description,
			}
			if ni.VLAN != 0 {
				vnic["vlan"] = ni.VLAN
			}
			vnics = append(vnics, vnic)
		}
		m["vnics"] = vnics
	}
	return m
}

func ListMVEs(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	output.SetOutputFormat(outputFormat)

	ctx, cancel := utils.ContextFromCmd(cmd)
	defer cancel()

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("error logging in: %v", err)
	}

	locationID, _ := cmd.Flags().GetInt("location-id")
	vendor, _ := cmd.Flags().GetString("vendor")
	name, _ := cmd.Flags().GetString("name")
	includeInactive, _ := cmd.Flags().GetBool("include-inactive")

	req := &megaport.ListMVEsRequest{
		IncludeInactive: includeInactive,
	}

	spinner := output.PrintResourceListing("MVE", noColor)

	mves, err := client.MVEService.ListMVEs(ctx, req)

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to list MVEs: %v", noColor, err)
		return fmt.Errorf("error listing MVEs: %v", err)
	}

	var activeMVEs []*megaport.MVE
	if !includeInactive {
		for _, mve := range mves {
			if mve != nil &&
				mve.ProvisioningStatus != "DECOMMISSIONED" &&
				mve.ProvisioningStatus != "CANCELLED" &&
				mve.ProvisioningStatus != "DECOMMISSIONING" {
				activeMVEs = append(activeMVEs, mve)
			}
		}
		mves = activeMVEs
	}

	filteredMVEs := filterMVEs(mves, locationID, vendor, name)

	limit, _ := cmd.Flags().GetInt("limit")
	if limit < 0 {
		return fmt.Errorf("--limit must be a non-negative integer")
	}
	if limit > 0 && len(filteredMVEs) > limit {
		filteredMVEs = filteredMVEs[:limit]
	}

	if len(filteredMVEs) == 0 {
		if outputFormat == utils.FormatTable {
			output.PrintInfo("No MVEs found. Create one with 'megaport mve buy'.", noColor)
		}
		return nil
	}

	err = printMVEs(filteredMVEs, outputFormat, noColor)
	if err != nil {
		output.PrintError("Failed to print MVEs: %v", noColor, err)
		return fmt.Errorf("error printing MVEs: %v", err)
	}
	return nil
}

func buildMVERequest(cmd *cobra.Command, noColor bool) (*megaport.BuyMVERequest, error) {
	interactive, _ := cmd.Flags().GetBool("interactive")
	jsonStr, _ := cmd.Flags().GetString("json")
	jsonFile, _ := cmd.Flags().GetString("json-file")

	flagsProvided := cmd.Flags().Changed("name") ||
		cmd.Flags().Changed("term") ||
		cmd.Flags().Changed("location-id") ||
		cmd.Flags().Changed("vendor-config") ||
		cmd.Flags().Changed("vnics")

	if jsonStr != "" || jsonFile != "" {
		output.PrintInfo("Using JSON input", noColor)
		req, err := processJSONBuyMVEInput(jsonStr, jsonFile)
		if err != nil {
			output.PrintError("Failed to process JSON input: %v", noColor, err)
			return nil, err
		}
		return req, nil
	} else if flagsProvided {
		output.PrintInfo("Using flag input", noColor)
		req, err := processFlagBuyMVEInput(cmd)
		if err != nil {
			output.PrintError("Failed to process flag input: %v", noColor, err)
			return nil, err
		}
		return req, nil
	} else if interactive {
		output.PrintInfo("Starting interactive mode", noColor)
		req, err := promptForBuyMVEDetails(noColor)
		if err != nil {
			output.PrintError("Interactive input failed: %v", noColor, err)
			return nil, err
		}
		return req, nil
	}
	output.PrintError("No input provided, use --interactive, --json, or flags to specify MVE details", noColor)
	return nil, fmt.Errorf("no input provided, use --interactive, --json, or flags to specify MVE details")
}

func BuyMVE(cmd *cobra.Command, args []string, noColor bool) error {
	ctx, cancel := utils.ContextFromCmd(cmd)
	defer cancel()

	req, err := buildMVERequest(cmd, noColor)
	if err != nil {
		return err
	}

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}

	if err := validation.ValidateMVEVendorConfig(req.VendorConfig); err != nil {
		output.PrintError("Validation failed: %v", noColor, err)
		return fmt.Errorf("validation failed: %v", err)
	}

	validationSpinner := output.PrintResourceValidating("MVE", noColor)

	err = client.MVEService.ValidateMVEOrder(ctx, req)

	validationSpinner.Stop()

	if err != nil {
		output.PrintError("Validation failed: %v", noColor, err)
		return fmt.Errorf("validation failed: %v", err)
	}

	output.PrintInfo("Validation successful", noColor)

	jsonStr, _ := cmd.Flags().GetString("json")
	jsonFile, _ := cmd.Flags().GetString("json-file")
	yes, _ := cmd.Flags().GetBool("yes")
	if !yes && jsonStr == "" && jsonFile == "" {
		details := []utils.BuyConfirmDetail{
			{Key: "Name", Value: req.Name},
			{Key: "Term", Value: fmt.Sprintf("%d months", req.Term)},
			{Key: "Location ID", Value: strconv.Itoa(req.LocationID)},
		}
		if !utils.BuyConfirmPrompt("MVE", details, noColor) {
			output.PrintInfo("Purchase cancelled", noColor)
			return exitcodes.New(exitcodes.Cancelled, fmt.Errorf("cancelled by user"))
		}
	}

	noWait, _ := cmd.Flags().GetBool("no-wait")
	if !noWait {
		req.WaitForProvision = true
		req.WaitForTime = 10 * time.Minute
	}

	var spinner *output.Spinner
	if req.WaitForProvision {
		spinner = output.PrintResourceProvisioning("MVE", req.Name, noColor)
	} else {
		spinner = output.PrintResourceCreating("MVE", req.Name, noColor)
	}

	resp, err := client.MVEService.BuyMVE(ctx, req)

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to buy MVE: %v", noColor, err)
		return err
	}

	output.PrintResourceCreated("MVE", resp.TechnicalServiceUID, noColor)
	return nil
}

func ValidateMVE(cmd *cobra.Command, args []string, noColor bool) error {
	ctx, cancel := utils.ContextFromCmd(cmd)
	defer cancel()

	req, err := buildMVERequest(cmd, noColor)
	if err != nil {
		return err
	}

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}

	if err := validation.ValidateMVEVendorConfig(req.VendorConfig); err != nil {
		output.PrintError("Validation failed: %v", noColor, err)
		return fmt.Errorf("validation failed: %v", err)
	}

	spinner := output.PrintResourceValidating("MVE", noColor)
	err = client.MVEService.ValidateMVEOrder(ctx, req)
	spinner.Stop()

	if err != nil {
		output.PrintError("Validation failed: %v", noColor, err)
		return fmt.Errorf("validation failed: %v", err)
	}

	output.PrintSuccess("MVE validation passed", noColor)
	return nil
}

func UpdateMVE(cmd *cobra.Command, args []string, noColor bool) error {
	ctx, cancel := utils.ContextFromCmd(cmd)
	defer cancel()
	mveUID := args[0]
	formattedUID := output.FormatUID(mveUID, noColor)

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("error logging in: %v", err)
	}

	getSpinner := output.PrintResourceGetting("MVE", mveUID, noColor)

	originalMVE, err := client.MVEService.GetMVE(ctx, mveUID)

	getSpinner.Stop()

	if err != nil {
		output.PrintError("Failed to get original MVE details: %v", noColor, err)
		return fmt.Errorf("error getting MVE details: %v", err)
	}

	interactive, _ := cmd.Flags().GetBool("interactive")
	jsonStr, _ := cmd.Flags().GetString("json")
	jsonFile, _ := cmd.Flags().GetString("json-file")

	flagsProvided := cmd.Flags().Changed("name") ||
		cmd.Flags().Changed("cost-centre") ||
		cmd.Flags().Changed("contract-term")

	var req *megaport.ModifyMVERequest

	if jsonStr != "" || jsonFile != "" {
		req, err = processJSONUpdateMVEInput(jsonStr, jsonFile, mveUID)
		if err != nil {
			output.PrintError("Failed to process JSON input: %v", noColor, err)
			return fmt.Errorf("error processing JSON input: %v", err)
		}
	} else if flagsProvided {
		req, err = processFlagUpdateMVEInput(cmd, mveUID)
		if err != nil {
			output.PrintError("Failed to process flag input: %v", noColor, err)
			return fmt.Errorf("error processing flag input: %v", err)
		}
	} else if interactive {
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

	req.WaitForUpdate = true
	req.WaitForTime = 10 * time.Minute

	updateSpinner := output.PrintResourceUpdating("MVE", mveUID, noColor)

	resp, err := client.MVEService.ModifyMVE(ctx, req)

	updateSpinner.Stop()

	if err != nil {
		output.PrintError("Failed to update MVE: %v", noColor, err)
		return err
	}

	if !resp.MVEUpdated {
		output.PrintWarning("MVE update request was not successful", noColor)
		return fmt.Errorf("MVE update request was not successful")
	}

	getUpdatedSpinner := output.PrintResourceGetting("MVE", mveUID, noColor)

	updatedMVE, err := client.MVEService.GetMVE(ctx, mveUID)

	getUpdatedSpinner.Stop()

	if err != nil {
		output.PrintError("MVE was updated but failed to retrieve updated details: %v", noColor, err)
		output.PrintResourceUpdated("MVE", mveUID, noColor)
		return nil
	}

	output.PrintResourceUpdated("MVE", mveUID, noColor)

	displayMVEChanges(originalMVE, updatedMVE, noColor)

	return nil
}

func GetMVE(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	output.SetOutputFormat(outputFormat)

	ctx, cancel := utils.ContextFromCmd(cmd)
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

	spinner := output.PrintResourceGetting("MVE", formattedUID, noColor)

	mve, err := client.MVEService.GetMVE(ctx, mveUID)

	spinner.Stop()

	if err != nil {
		err = utils.WrapAPIError(err, "MVE", mveUID)
		output.PrintError("Failed to get MVE: %v", noColor, err)
		return fmt.Errorf("error getting MVE: %w", err)
	}

	if mve == nil {
		output.PrintError("No MVE found with UID: %s", noColor, mveUID)
		return fmt.Errorf("no MVE found with UID: %s", mveUID)
	}

	export, _ := cmd.Flags().GetBool("export")
	if export {
		cfg := exportMVEConfig(mve)
		jsonBytes, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return fmt.Errorf("error marshaling export config: %v", err)
		}
		fmt.Println(string(jsonBytes))
		return nil
	}

	err = printMVEs([]*megaport.MVE{mve}, outputFormat, noColor)
	if err != nil {
		output.PrintError("Failed to print MVEs: %v", noColor, err)
		return fmt.Errorf("error printing MVEs: %v", err)
	}
	return nil
}

func ListMVEImages(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	output.SetOutputFormat(outputFormat)
	ctx, cancel := utils.ContextFromCmd(cmd)
	defer cancel()

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("error logging in: %v", err)
	}

	spinner := output.PrintResourceListing("MVE image", noColor)

	images, err := client.MVEService.ListMVEImages(ctx)

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to list MVE images: %v", noColor, err)
		return fmt.Errorf("error listing MVE images: %v", err)
	}

	if images == nil {
		output.PrintWarning("No MVE images found", noColor)
		return fmt.Errorf("no MVE images found")
	}

	vendor, _ := cmd.Flags().GetString("vendor")
	productCode, _ := cmd.Flags().GetString("product-code")
	id, _ := cmd.Flags().GetInt("id")
	version, _ := cmd.Flags().GetString("version")
	releaseImage, _ := cmd.Flags().GetBool("release-image")

	filteredImages := filterMVEImages(images, vendor, productCode, id, version, releaseImage)

	err = output.PrintOutput(filteredImages, outputFormat, noColor)
	if err != nil {
		output.PrintError("Failed to print MVE images: %v", noColor, err)
		return fmt.Errorf("error printing MVE images: %v", err)
	}
	return nil
}

func ListAvailableMVESizes(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	output.SetOutputFormat(outputFormat)
	ctx, cancel := utils.ContextFromCmd(cmd)
	defer cancel()

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("error logging in: %v", err)
	}

	spinner := output.PrintResourceListing("MVE size", noColor)

	sizes, err := client.MVEService.ListAvailableMVESizes(ctx)

	spinner.Stop()

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
	ctx, cancel := utils.ContextFromCmd(cmd)
	defer cancel()
	mveUID := args[0]

	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		output.PrintError("Failed to get force flag: %v", noColor, err)
		return err
	}

	if !force {
		confirmMsg := "Are you sure you want to delete MVE " + mveUID + "? "
		if !utils.ConfirmPrompt(confirmMsg, noColor) {
			output.PrintInfo("Deletion cancelled", noColor)
			return exitcodes.New(exitcodes.Cancelled, fmt.Errorf("cancelled by user"))
		}
	}

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("error logging in: %v", err)
	}

	safeDelete, err := cmd.Flags().GetBool("safe-delete")
	if err != nil {
		output.PrintError("Failed to get safe-delete flag: %v", noColor, err)
		return err
	}

	spinner := output.PrintResourceDeleting("MVE", mveUID, noColor)

	req := &megaport.DeleteMVERequest{
		MVEID:      mveUID,
		SafeDelete: safeDelete,
	}
	resp, err := client.MVEService.DeleteMVE(ctx, req)

	spinner.Stop()

	if err != nil {
		err = utils.WrapAPIError(err, "MVE", mveUID)
		output.PrintError("Failed to delete MVE: %v", noColor, err)
		return fmt.Errorf("error deleting MVE: %w", err)
	}

	if resp.IsDeleted {
		output.PrintResourceDeleted("MVE", mveUID, false, noColor)
	} else {
		output.PrintWarning("MVE delete failed", noColor)
	}
	return nil
}

func ListMVEResourceTags(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	mveUID := args[0]
	return utils.ListResourceTags("MVE", mveUID, noColor, outputFormat, func(ctx context.Context, uid string) (map[string]string, error) {
		client, err := config.LoginFunc(ctx)
		if err != nil {
			return nil, err
		}
		return listMVEResourceTagsFunc(ctx, client, uid)
	})
}

func UpdateMVEResourceTags(cmd *cobra.Command, args []string, noColor bool) error {
	mveUID := args[0]
	var client *megaport.Client
	login := func(ctx context.Context) error {
		var err error
		client, err = config.LoginFunc(ctx)
		return err
	}
	return utils.UpdateResourceTags(utils.UpdateTagsOptions{
		ResourceType: "MVE",
		UID:          mveUID,
		NoColor:      noColor,
		Cmd:          cmd,
		ListFunc: func(ctx context.Context, uid string) (map[string]string, error) {
			if err := login(ctx); err != nil {
				return nil, err
			}
			return listMVEResourceTagsFunc(ctx, client, uid)
		},
		UpdateFunc: func(ctx context.Context, uid string, tags map[string]string) error {
			return client.MVEService.UpdateMVEResourceTags(ctx, uid, tags)
		},
	})
}

// GetMVEStatus retrieves only the provisioning status of an MVE without all details
func GetMVEStatus(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	output.SetOutputFormat(outputFormat)

	ctx, cancel := utils.ContextFromCmd(cmd)
	defer cancel()

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("error logging in: %v", err)
	}

	mveUID := args[0]

	spinner := output.PrintResourceGetting("MVE", mveUID, noColor)

	mve, err := client.MVEService.GetMVE(ctx, mveUID)

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to get MVE status: %v", noColor, err)
		return fmt.Errorf("error getting MVE status: %v", err)
	}

	if mve == nil {
		output.PrintError("No MVE found with UID: %s", noColor, mveUID)
		return fmt.Errorf("no MVE found with UID: %s", mveUID)
	}

	status := []MVEStatus{
		{
			UID:    mve.UID,
			Name:   mve.Name,
			Status: mve.ProvisioningStatus,
			Vendor: mve.Vendor,
			Size:   mve.Size,
		},
	}

	return output.PrintOutput(status, outputFormat, noColor)
}

func LockMVE(cmd *cobra.Command, args []string, noColor bool) error {
	ctx, cancel := utils.ContextFromCmd(cmd)
	defer cancel()

	client, err := config.Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %v", err)
	}

	mveUID := args[0]

	output.PrintInfo("Locking MVE %s...", noColor, mveUID)

	_, err = lockMVEFunc(ctx, client, mveUID)
	if err != nil {
		return fmt.Errorf("error locking MVE: %v", err)
	}

	output.PrintSuccess("MVE %s locked successfully", noColor, mveUID)
	return nil
}

func UnlockMVE(cmd *cobra.Command, args []string, noColor bool) error {
	ctx, cancel := utils.ContextFromCmd(cmd)
	defer cancel()

	client, err := config.Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %v", err)
	}

	mveUID := args[0]

	output.PrintInfo("Unlocking MVE %s...", noColor, mveUID)

	_, err = unlockMVEFunc(ctx, client, mveUID)
	if err != nil {
		return fmt.Errorf("error unlocking MVE: %v", err)
	}

	output.PrintSuccess("MVE %s unlocked successfully", noColor, mveUID)
	return nil
}

func RestoreMVE(cmd *cobra.Command, args []string, noColor bool) error {
	ctx, cancel := utils.ContextFromCmd(cmd)
	defer cancel()

	client, err := config.Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %v", err)
	}

	mveUID := args[0]

	output.PrintInfo("Restoring MVE %s...", noColor, mveUID)

	_, err = restoreMVEFunc(ctx, client, mveUID)
	if err != nil {
		return fmt.Errorf("error restoring MVE: %v", err)
	}

	output.PrintSuccess("MVE %s restored successfully", noColor, mveUID)
	return nil
}
