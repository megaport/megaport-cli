package mve

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/commands/config"
	"github.com/megaport/megaport-cli/internal/utils"
	"github.com/megaport/megaport-cli/internal/validation"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

func ListMVEs(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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

	if len(filteredMVEs) == 0 {
		output.PrintWarning("No MVEs found matching the specified filters", noColor)
		return printMVEs(filteredMVEs, outputFormat, noColor)
	}

	err = printMVEs(filteredMVEs, outputFormat, noColor)
	if err != nil {
		output.PrintError("Failed to print MVEs: %v", noColor, err)
		return fmt.Errorf("error printing MVEs: %v", err)
	}
	return nil
}

func BuyMVE(cmd *cobra.Command, args []string, noColor bool) error {
	ctx := context.Background()

	interactive, _ := cmd.Flags().GetBool("interactive")
	jsonStr, _ := cmd.Flags().GetString("json")
	jsonFile, _ := cmd.Flags().GetString("json-file")

	flagsProvided := cmd.Flags().Changed("name") ||
		cmd.Flags().Changed("term") ||
		cmd.Flags().Changed("location-id") ||
		cmd.Flags().Changed("vendor-config") ||
		cmd.Flags().Changed("vnics")

	var req *megaport.BuyMVERequest
	var err error

	if jsonStr != "" || jsonFile != "" {
		output.PrintInfo("Using JSON input", noColor)
		req, err = processJSONBuyMVEInput(jsonStr, jsonFile)
		if err != nil {
			output.PrintError("Failed to process JSON input: %v", noColor, err)
			return err
		}
	} else if flagsProvided {
		output.PrintInfo("Using flag input", noColor)
		req, err = processFlagBuyMVEInput(cmd)
		if err != nil {
			output.PrintError("Failed to process flag input: %v", noColor, err)
			return err
		}
	} else if interactive {
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

	req.WaitForProvision = true
	req.WaitForTime = 10 * time.Minute

	spinner := output.PrintResourceCreating("MVE", req.Name, noColor)

	resp, err := client.MVEService.BuyMVE(ctx, req)

	spinner.Stop()

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

	spinner := output.PrintResourceGetting("MVE", formattedUID, noColor)

	mve, err := client.MVEService.GetMVE(ctx, mveUID)

	spinner.Stop()

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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
	ctx := context.Background()
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
			return nil
		}
	}

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("error logging in: %v", err)
	}

	spinner := output.PrintResourceDeleting("MVE", mveUID, noColor)

	req := &megaport.DeleteMVERequest{
		MVEID: mveUID,
	}
	resp, err := client.MVEService.DeleteMVE(ctx, req)

	spinner.Stop()

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

func ListMVEResourceTags(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	mveUID := args[0]

	ctx := context.Background()

	client, err := config.LoginFunc(ctx)
	if err != nil {
		return err
	}

	tagsMap, err := listMVEResourceTagsFunc(ctx, client, mveUID)

	if err != nil {
		return fmt.Errorf("error getting resource tags for MVE %s: %v", mveUID, err)
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

func UpdateMVEResourceTags(cmd *cobra.Command, args []string, noColor bool) error {
	mveUID := args[0]

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := config.LoginFunc(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}

	existingTags, err := listMVEResourceTagsFunc(ctx, client, mveUID)

	if err != nil {
		return fmt.Errorf("failed to get existing resource tags: %v", err)
	}

	interactive, _ := cmd.Flags().GetBool("interactive")

	var resourceTags map[string]string

	if interactive {
		resourceTags, err = utils.UpdateResourceTagsPrompt(existingTags, noColor)
		if err != nil {
			output.PrintError("Failed to update resource tags", noColor)
			return err
		}
	} else {
		jsonStr, _ := cmd.Flags().GetString("json")
		jsonFile, _ := cmd.Flags().GetString("json-file")

		if jsonStr != "" {
			resourceTags = make(map[string]string)
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

			resourceTags = make(map[string]string)
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
		fmt.Println("No tags provided. The MVE will have all existing tags removed.")
	}

	spinner := output.PrintResourceUpdating("MVE-Resource-Tags", mveUID, noColor)

	err = client.MVEService.UpdateMVEResourceTags(ctx, mveUID, resourceTags)

	spinner.Stop()

	if err != nil {
		return fmt.Errorf("failed to update resource tags: %v", err)
	}

	fmt.Printf("Resource tags updated for MVE %s\n", mveUID)
	return nil
}
