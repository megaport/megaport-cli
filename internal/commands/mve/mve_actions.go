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

	ctx, cancel, client, err := utils.LoginClient(cmd, 90*time.Second, config.Login)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}
	defer cancel()

	// Flag read errors are intentionally ignored — flags are registered by the command builder.
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
		return fmt.Errorf("failed to list MVEs: %w", err)
	}

	var activeMVEs []*megaport.MVE
	if !includeInactive {
		for _, mve := range mves {
			if mve != nil &&
				mve.ProvisioningStatus != megaport.STATUS_DECOMMISSIONED &&
				mve.ProvisioningStatus != megaport.STATUS_CANCELLED &&
				mve.ProvisioningStatus != utils.StatusDecommissioning {
				activeMVEs = append(activeMVEs, mve)
			}
		}
		mves = activeMVEs
	}

	filteredMVEs := filterMVEs(mves, locationID, vendor, name)

	tagFilters, _ := cmd.Flags().GetStringArray("tag")
	if len(tagFilters) > 0 {
		var tagged []*megaport.MVE
		for _, m := range filteredMVEs {
			tags, err := listMVEResourceTagsFunc(ctx, client, m.UID)
			if err != nil {
				continue
			}
			if utils.MatchesTagFilters(tags, tagFilters) {
				tagged = append(tagged, m)
			}
		}
		filteredMVEs = tagged
	}

	limit, _ := cmd.Flags().GetInt("limit")
	return utils.ApplyLimitAndPrint(filteredMVEs, limit, outputFormat, noColor,
		"No MVEs found. Create one with 'megaport mve buy'.", printMVEs)
}

func buildMVERequest(cmd *cobra.Command, noColor bool) (*megaport.BuyMVERequest, error) {
	return utils.ResolveInput(utils.InputConfig[*megaport.BuyMVERequest]{
		ResourceName: "MVE",
		Cmd:          cmd,
		NoColor:      noColor,
		FlagsProvided: func() bool {
			return cmd.Flags().Changed("name") || cmd.Flags().Changed("term") ||
				cmd.Flags().Changed("location-id") || cmd.Flags().Changed("vendor-config") ||
				cmd.Flags().Changed("vnics")
		},
		FromJSON:   processJSONBuyMVEInput,
		FromFlags:  func() (*megaport.BuyMVERequest, error) { return processFlagBuyMVEInput(cmd) },
		FromPrompt: func() (*megaport.BuyMVERequest, error) { return promptForBuyMVEDetails(noColor) },
	})
}

func BuyMVE(cmd *cobra.Command, args []string, noColor bool) error {
	ctx, cancel := utils.ContextFromCmdWithDefault(cmd, utils.DefaultMutationTimeout)
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
		return fmt.Errorf("validation failed: %w", err)
	}

	validationSpinner := output.PrintResourceValidating("MVE", noColor)

	err = client.MVEService.ValidateMVEOrder(ctx, req)

	validationSpinner.Stop()

	if err != nil {
		output.PrintError("Validation failed: %v", noColor, err)
		return fmt.Errorf("validation failed: %w", err)
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
		req.WaitForTime = utils.DefaultProvisionTimeout
	}

	var spinner *output.Spinner
	if req.WaitForProvision {
		spinner = output.PrintResourceProvisioning("MVE", req.Name, noColor)
	} else {
		spinner = output.PrintResourceCreating("MVE", req.Name, noColor)
	}

	var resp *megaport.BuyMVEResponse
	err = utils.WithRetry(ctx, func(ctx context.Context) error {
		var e error
		resp, e = client.MVEService.BuyMVE(ctx, req)
		return e
	})

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
		return fmt.Errorf("validation failed: %w", err)
	}

	spinner := output.PrintResourceValidating("MVE", noColor)
	err = client.MVEService.ValidateMVEOrder(ctx, req)
	spinner.Stop()

	if err != nil {
		output.PrintError("Validation failed: %v", noColor, err)
		return fmt.Errorf("validation failed: %w", err)
	}

	output.PrintSuccess("MVE validation passed", noColor)
	return nil
}

func UpdateMVE(cmd *cobra.Command, args []string, noColor bool) error {
	ctx, cancel := utils.ContextFromCmdWithDefault(cmd, utils.DefaultMutationTimeout)
	defer cancel()
	mveUID := args[0]
	formattedUID := output.FormatUID(mveUID, noColor)

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}

	getSpinner := output.PrintResourceGetting("MVE", mveUID, noColor)

	originalMVE, err := client.MVEService.GetMVE(ctx, mveUID)

	getSpinner.Stop()

	if err != nil {
		output.PrintError("Failed to get original MVE details: %v", noColor, err)
		return fmt.Errorf("failed to get MVE details: %w", err)
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
			return fmt.Errorf("failed to process JSON input: %w", err)
		}
	} else if flagsProvided {
		req, err = processFlagUpdateMVEInput(cmd, mveUID)
		if err != nil {
			output.PrintError("Failed to process flag input: %v", noColor, err)
			return fmt.Errorf("failed to process flag input: %w", err)
		}
	} else if interactive {
		output.PrintInfo("Starting interactive mode for MVE %s", noColor, formattedUID)
		req, err = promptForUpdateMVEDetails(mveUID, noColor)
		if err != nil {
			output.PrintError("Failed to get MVE details interactively: %v", noColor, err)
			return fmt.Errorf("failed to get MVE details interactively: %w", err)
		}
	} else {
		output.PrintError("No input provided", noColor)
		return fmt.Errorf("no input provided, use --interactive, --json, or flags to specify MVE update details")
	}

	req.WaitForUpdate = true
	req.WaitForTime = utils.DefaultProvisionTimeout

	updateSpinner := output.PrintResourceUpdating("MVE", mveUID, noColor)

	var resp *megaport.ModifyMVEResponse
	err = utils.WithRetry(ctx, func(ctx context.Context) error {
		var e error
		resp, e = client.MVEService.ModifyMVE(ctx, req)
		return e
	})

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

	watch, _ := cmd.Flags().GetBool("watch")
	if watch {
		return watchGetMVE(cmd, args, noColor, outputFormat)
	}

	ctx, cancel, client, err := utils.LoginClient(cmd, 90*time.Second, config.Login)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}
	defer cancel()

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
		return fmt.Errorf("failed to get MVE: %w", err)
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
			return fmt.Errorf("failed to marshal export config: %w", err)
		}
		fmt.Println(string(jsonBytes))
		return nil
	}

	err = printMVEs([]*megaport.MVE{mve}, outputFormat, noColor)
	if err != nil {
		output.PrintError("Failed to print MVEs: %v", noColor, err)
		return fmt.Errorf("failed to print MVEs: %w", err)
	}
	return nil
}

func watchGetMVE(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	mveUID := args[0]
	return utils.WatchResource(cmd, "MVE", mveUID, noColor, outputFormat, config.Login,
		func(ctx context.Context, client *megaport.Client) (string, error) {
			mve, err := client.MVEService.GetMVE(ctx, mveUID)
			if err != nil {
				return "", err
			}
			if mve == nil {
				return "", fmt.Errorf("no MVE found with UID: %s", mveUID)
			}
			err = printMVEs([]*megaport.MVE{mve}, outputFormat, noColor)
			return mve.ProvisioningStatus, err
		})
}

func ListMVEImages(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	output.SetOutputFormat(outputFormat)
	ctx, cancel, client, err := utils.LoginClient(cmd, 90*time.Second, config.Login)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}
	defer cancel()

	spinner := output.PrintResourceListing("MVE image", noColor)

	images, err := client.MVEService.ListMVEImages(ctx)

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to list MVE images: %v", noColor, err)
		return fmt.Errorf("failed to list MVE images: %w", err)
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
		return fmt.Errorf("failed to print MVE images: %w", err)
	}
	return nil
}

func ListAvailableMVESizes(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	output.SetOutputFormat(outputFormat)
	ctx, cancel, client, err := utils.LoginClient(cmd, 90*time.Second, config.Login)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}
	defer cancel()

	spinner := output.PrintResourceListing("MVE size", noColor)

	sizes, err := client.MVEService.ListAvailableMVESizes(ctx)

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to list MVE sizes: %v", noColor, err)
		return fmt.Errorf("failed to list MVE sizes: %w", err)
	}

	if sizes == nil {
		output.PrintWarning("No MVE sizes found", noColor)
		return fmt.Errorf("no MVE sizes found")
	}

	err = output.PrintOutput(sizes, outputFormat, noColor)
	if err != nil {
		output.PrintError("Failed to print MVE sizes: %v", noColor, err)
		return fmt.Errorf("failed to print MVE sizes: %w", err)
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
		return err
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
	var resp *megaport.DeleteMVEResponse
	err = utils.WithRetry(ctx, func(ctx context.Context) error {
		var e error
		resp, e = client.MVEService.DeleteMVE(ctx, req)
		return e
	})

	spinner.Stop()

	if err != nil {
		err = utils.WrapAPIError(err, "MVE", mveUID)
		output.PrintError("Failed to delete MVE: %v", noColor, err)
		return fmt.Errorf("failed to delete MVE: %w", err)
	}

	// MVEs always delete immediately (SDK hardcodes DeleteNow: true), so the
	// --now flag has no API effect. Always report as immediate deletion.
	if resp.IsDeleted {
		output.PrintResourceDeleted("MVE", mveUID, true, noColor)
	} else {
		output.PrintWarning("MVE delete failed", noColor)
	}
	return nil
}

func ListMVEResourceTags(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	mveUID := args[0]
	return utils.ListResourceTags("MVE", mveUID, noColor, outputFormat, func(ctx context.Context, uid string) (map[string]string, error) {
		client, err := config.Login(ctx)
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
		client, err = config.Login(ctx)
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

	watch, _ := cmd.Flags().GetBool("watch")
	if watch {
		return watchMVEStatus(cmd, args, noColor, outputFormat)
	}

	ctx, cancel, client, err := utils.LoginClient(cmd, 90*time.Second, config.Login)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}
	defer cancel()

	mveUID := args[0]

	spinner := output.PrintResourceGetting("MVE", mveUID, noColor)

	mve, err := client.MVEService.GetMVE(ctx, mveUID)

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to get MVE status: %v", noColor, err)
		return fmt.Errorf("failed to get MVE status: %w", err)
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

func watchMVEStatus(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	mveUID := args[0]
	return utils.WatchResource(cmd, "MVE", mveUID, noColor, outputFormat, config.Login,
		func(ctx context.Context, client *megaport.Client) (string, error) {
			mve, err := client.MVEService.GetMVE(ctx, mveUID)
			if err != nil {
				return "", err
			}
			if mve == nil {
				return "", fmt.Errorf("no MVE found with UID: %s", mveUID)
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
			err = output.PrintOutput(status, outputFormat, noColor)
			return mve.ProvisioningStatus, err
		})
}

func LockMVE(cmd *cobra.Command, args []string, noColor bool) error {
	ctx, cancel, client, err := utils.LoginClient(cmd, 90*time.Second, config.Login)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}
	defer cancel()

	mveUID := args[0]

	output.PrintInfo("Locking MVE %s...", noColor, mveUID)

	err = utils.WithRetry(ctx, func(ctx context.Context) error {
		_, e := lockMVEFunc(ctx, client, mveUID)
		return e
	})
	if err != nil {
		return fmt.Errorf("failed to lock MVE: %w", err)
	}

	output.PrintSuccess("MVE %s locked successfully", noColor, mveUID)
	return nil
}

func UnlockMVE(cmd *cobra.Command, args []string, noColor bool) error {
	ctx, cancel, client, err := utils.LoginClient(cmd, 90*time.Second, config.Login)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}
	defer cancel()

	mveUID := args[0]

	output.PrintInfo("Unlocking MVE %s...", noColor, mveUID)

	err = utils.WithRetry(ctx, func(ctx context.Context) error {
		_, e := unlockMVEFunc(ctx, client, mveUID)
		return e
	})
	if err != nil {
		return fmt.Errorf("failed to unlock MVE: %w", err)
	}

	output.PrintSuccess("MVE %s unlocked successfully", noColor, mveUID)
	return nil
}

func RestoreMVE(cmd *cobra.Command, args []string, noColor bool) error {
	ctx, cancel, client, err := utils.LoginClient(cmd, 90*time.Second, config.Login)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}
	defer cancel()

	mveUID := args[0]

	output.PrintInfo("Restoring MVE %s...", noColor, mveUID)

	err = utils.WithRetry(ctx, func(ctx context.Context) error {
		_, e := restoreMVEFunc(ctx, client, mveUID)
		return e
	})
	if err != nil {
		return fmt.Errorf("failed to restore MVE: %w", err)
	}

	output.PrintSuccess("MVE %s restored successfully", noColor, mveUID)
	return nil
}
