package vxc

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
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

func GetVXC(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	// Set output format for proper JSON mode handling
	output.SetOutputFormat(outputFormat)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("error logging in: %v", err)
	}

	vxcUID := args[0]

	spinner := output.PrintResourceGetting("VXC", vxcUID, noColor)

	vxc, err := client.VXCService.GetVXC(ctx, vxcUID)

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to get VXC: %v", noColor, err)
		return fmt.Errorf("error getting VXC: %v", err)
	}

	err = printVXCs([]*megaport.VXC{vxc}, outputFormat, noColor)
	if err != nil {
		output.PrintError("Failed to print VXCs: %v", noColor, err)
		return fmt.Errorf("error printing VXCs: %v", err)
	}
	return nil
}

var hasUpdateVXCNonInteractiveFlags = func(cmd *cobra.Command) bool {
	flagNames := []string{"name", "rate-limit", "a-end-vlan", "b-end-vlan", "a-end-location", "b-end-location", "locked"}
	for _, name := range flagNames {
		if cmd.Flags().Changed(name) {
			return true
		}
	}
	return false
}

func BuyVXC(cmd *cobra.Command, args []string, noColor bool) error {
	ctx := context.Background()

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("error logging in: %v", err)
	}

	interactive, _ := cmd.Flags().GetBool("interactive")
	jsonStr, _ := cmd.Flags().GetString("json")
	jsonFile, _ := cmd.Flags().GetString("json-file")

	flagsProvided := cmd.Flags().Changed("name") || cmd.Flags().Changed("rate-limit") ||
		cmd.Flags().Changed("term") || cmd.Flags().Changed("a-end-uid") ||
		cmd.Flags().Changed("a-end-vlan") || cmd.Flags().Changed("b-end-uid") ||
		cmd.Flags().Changed("b-end-vlan")

	var req *megaport.BuyVXCRequest

	if jsonStr != "" || jsonFile != "" {
		output.PrintInfo("Using JSON input", noColor)
		req, err = buildVXCRequestFromJSON(jsonStr, jsonFile)
		if err != nil {
			output.PrintError("Failed to process JSON input: %v", noColor, err)
			return err
		}
	} else if flagsProvided {
		output.PrintInfo("Using flag input", noColor)
		req, err = buildVXCRequestFromFlags(cmd, ctx, client.VXCService)
		if err != nil {
			output.PrintError("Failed to process flag input: %v", noColor, err)
			return err
		}
	} else if interactive {
		output.PrintInfo("Starting interactive mode", noColor)
		req, err = buildVXCRequestFromPrompt(ctx, client.VXCService, noColor)
		if err != nil {
			output.PrintError("Interactive input failed: %v", noColor, err)
			return err
		}
	} else {
		output.PrintError("No input provided", noColor)
		return fmt.Errorf("no input provided, use --interactive, --json, or flags to specify VXC details")
	}

	req.WaitForProvision = true
	req.WaitForTime = 10 * time.Minute

	validateSpinner := output.PrintResourceValidating("VXC", noColor)

	err = client.VXCService.ValidateVXCOrder(ctx, req)
	validateSpinner.Stop()

	if err != nil {
		output.PrintError("Failed to validate VXC order: %v", noColor, err)
		return err
	}

	spinner := output.PrintResourceCreating("VXC", req.VXCName, noColor)

	resp, err := buyVXCFunc(ctx, client, req)

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to buy VXC: %v", noColor, err)
		return err
	}

	output.PrintResourceCreated("VXC", resp.TechnicalServiceUID, noColor)
	return nil
}

func UpdateVXC(cmd *cobra.Command, args []string, noColor bool) error {
	ctx := context.Background()

	vxcUID := args[0]
	formattedUID := output.FormatUID(vxcUID, noColor)

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}

	getSpinner := output.PrintResourceGetting("VXC", vxcUID, noColor)

	originalVXC, err := client.VXCService.GetVXC(ctx, vxcUID)

	getSpinner.Stop()

	if err != nil {
		output.PrintError("Failed to retrieve original VXC details: %v", noColor, err)
		return fmt.Errorf("failed to retrieve original VXC details: %v", err)
	}

	var req *megaport.UpdateVXCRequest
	var buildErr error

	interactive, _ := cmd.Flags().GetBool("interactive")
	jsonStr, _ := cmd.Flags().GetString("json")
	jsonFilePath, _ := cmd.Flags().GetString("json-file")

	if jsonStr != "" || jsonFilePath != "" {
		output.PrintInfo("Using JSON input for VXC %s", noColor, formattedUID)
		req, buildErr = buildUpdateVXCRequestFromJSON(jsonStr, jsonFilePath)
	} else if interactive || !hasUpdateVXCNonInteractiveFlags(cmd) {
		output.PrintInfo("Starting interactive mode for VXC %s", noColor, formattedUID)
		req, buildErr = buildUpdateVXCRequestFromPrompt(vxcUID, noColor)
	} else {
		output.PrintInfo("Using flag input for VXC %s", noColor, formattedUID)
		req, buildErr = buildUpdateVXCRequestFromFlags(cmd)
	}

	if buildErr != nil {
		output.PrintError("Failed to build update request: %v", noColor, buildErr)
		return buildErr
	}

	if req == nil {
		output.PrintError("No update parameters provided", noColor)
		return fmt.Errorf("no update parameters provided")
	}

	req.WaitForUpdate = true
	req.WaitForTime = 10 * time.Minute

	updateSpinner := output.PrintResourceUpdating("VXC", vxcUID, noColor)

	err = updateVXCFunc(ctx, client, vxcUID, req)

	updateSpinner.Stop()

	if err != nil {
		output.PrintError("Failed to update VXC: %v", noColor, err)
		return fmt.Errorf("failed to update VXC: %v", err)
	}

	getUpdatedSpinner := output.PrintResourceGetting("VXC", vxcUID, noColor)

	updatedVXC, err := getVXCFunc(ctx, client, vxcUID)

	getUpdatedSpinner.Stop()

	if err != nil {
		output.PrintError("VXC was updated but failed to retrieve updated details: %v", noColor, err)
		output.PrintResourceUpdated("VXC", vxcUID, noColor)
		return nil
	}

	output.PrintResourceUpdated("VXC", vxcUID, noColor)

	displayVXCChanges(originalVXC, updatedVXC, noColor)

	return nil
}

func DeleteVXC(cmd *cobra.Command, args []string, noColor bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	vxcUID := args[0]
	formattedUID := output.FormatUID(vxcUID, noColor)

	force, _ := cmd.Flags().GetBool("force")
	deleteNow, _ := cmd.Flags().GetBool("now")

	if !force {
		message := fmt.Sprintf("Are you sure you want to delete VXC %s?", formattedUID)
		if !utils.ConfirmPrompt(message, noColor) {
			output.PrintInfo("Deletion cancelled", noColor)
			return nil
		}
	}

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}

	req := &megaport.DeleteVXCRequest{
		DeleteNow: deleteNow,
	}

	spinner := output.PrintResourceDeleting("VXC", vxcUID, noColor)

	err = deleteVXCFunc(ctx, client, vxcUID, req)

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to delete VXC: %v", noColor, err)
		return err
	}

	output.PrintResourceDeleted("VXC", vxcUID, deleteNow, noColor)
	return nil
}

func ListVXCResourceTags(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	// Set output format for proper JSON mode handling
	output.SetOutputFormat(outputFormat)

	vxcUID := args[0]

	ctx := context.Background()

	client, err := config.LoginFunc(ctx)
	if err != nil {
		return err
	}

	tagsMap, err := client.VXCService.ListVXCResourceTags(ctx, vxcUID)

	if err != nil {
		output.PrintError("Error getting resource tags for VXC %s: %v", noColor, vxcUID, err)
		return fmt.Errorf("error getting resource tags for VXC %s: %v", vxcUID, err)
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

func UpdateVXCResourceTags(cmd *cobra.Command, args []string, noColor bool) error {
	vxcUID := args[0]

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := config.LoginFunc(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}

	existingTags, err := client.VXCService.ListVXCResourceTags(ctx, vxcUID)

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
		cmd.PrintErrln("No tags provided. The VXC will have all existing tags removed")
	}

	spinner := output.PrintResourceUpdating("VXC-Resource-Tags", vxcUID, noColor)

	err = client.VXCService.UpdateVXCResourceTags(ctx, vxcUID, resourceTags)

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to update resource tags: %v", noColor, err)
		return fmt.Errorf("failed to update resource tags: %v", err)
	}

	fmt.Printf("Resource tags updated for VXC %s\n", vxcUID)
	return nil
}

func GetVXCStatus(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	// Set output format for proper JSON mode handling
	output.SetOutputFormat(outputFormat)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("error logging in: %v", err)
	}

	vxcUID := args[0]

	spinner := output.PrintResourceGetting("VXC", vxcUID, noColor)

	vxc, err := client.VXCService.GetVXC(ctx, vxcUID)

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to get VXC status: %v", noColor, err)
		return fmt.Errorf("error getting VXC status: %v", err)
	}

	status := []VXCStatus{
		{
			UID:    vxc.UID,
			Name:   vxc.Name,
			Status: vxc.ProvisioningStatus,
			Type:   vxc.Type,
		},
	}

	return output.PrintOutput(status, outputFormat, noColor)
}
