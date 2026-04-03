package vxc

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/megaport/megaport-cli/internal/base/exitcodes"
	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/commands/config"
	"github.com/megaport/megaport-cli/internal/utils"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

func exportVXCConfig(vxc *megaport.VXC) map[string]interface{} {
	m := map[string]interface{}{
		"portUid":   vxc.AEndConfiguration.UID,
		"vxcName":   vxc.Name,
		"rateLimit": vxc.RateLimit,
		"term":      vxc.ContractTermMonths,
	}
	if vxc.CostCentre != "" {
		m["costCentre"] = vxc.CostCentre
	}

	aEnd := map[string]interface{}{}
	if vxc.AEndConfiguration.VLAN != 0 {
		aEnd["vlan"] = vxc.AEndConfiguration.VLAN
	}
	if vxc.AEndConfiguration.InnerVLAN != 0 {
		aEnd["innerVlan"] = vxc.AEndConfiguration.InnerVLAN
	}
	if len(aEnd) > 0 {
		m["aEndConfiguration"] = aEnd
	}

	bEnd := map[string]interface{}{
		"productUID": vxc.BEndConfiguration.UID,
	}
	if vxc.BEndConfiguration.VLAN != 0 {
		bEnd["vlan"] = vxc.BEndConfiguration.VLAN
	}
	if vxc.BEndConfiguration.InnerVLAN != 0 {
		bEnd["innerVlan"] = vxc.BEndConfiguration.InnerVLAN
	}
	m["bEndConfiguration"] = bEnd

	return m
}

func ListVXCs(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	output.SetOutputFormat(outputFormat)

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("error logging in: %v", err)
	}

	name, _ := cmd.Flags().GetString("name")
	nameContains, _ := cmd.Flags().GetString("name-contains")
	rateLimit, _ := cmd.Flags().GetInt("rate-limit")
	aEndUID, _ := cmd.Flags().GetString("a-end-uid")
	bEndUID, _ := cmd.Flags().GetString("b-end-uid")
	includeInactive, _ := cmd.Flags().GetBool("include-inactive")
	statusStr, _ := cmd.Flags().GetString("status")

	// Determine server-side name filter: --name-contains takes precedence, else --name
	serverNameContains := nameContains
	if serverNameContains == "" {
		serverNameContains = name
	}

	// Parse comma-separated status filter
	var statusFilter []string
	if statusStr != "" {
		for _, s := range strings.Split(statusStr, ",") {
			s = strings.TrimSpace(s)
			if s != "" {
				statusFilter = append(statusFilter, s)
			}
		}
	}

	req := &megaport.ListVXCsRequest{
		IncludeInactive: includeInactive,
		NameContains:    serverNameContains,
		AEndProductUID:  aEndUID,
		BEndProductUID:  bEndUID,
		RateLimit:       rateLimit,
		Status:          statusFilter,
	}

	spinner := output.PrintResourceListing("VXC", noColor)

	vxcs, err := client.VXCService.ListVXCs(ctx, req)

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to list VXCs: %v", noColor, err)
		return fmt.Errorf("error listing VXCs: %v", err)
	}

	var activeVXCs []*megaport.VXC
	if !includeInactive {
		for _, vxc := range vxcs {
			if vxc != nil &&
				vxc.ProvisioningStatus != megaport.STATUS_DECOMMISSIONED &&
				vxc.ProvisioningStatus != megaport.STATUS_CANCELLED &&
				vxc.ProvisioningStatus != "DECOMMISSIONING" {
				activeVXCs = append(activeVXCs, vxc)
			}
		}
		vxcs = activeVXCs
	}

	filteredVXCs := filterVXCs(vxcs, name)

	if len(filteredVXCs) == 0 {
		if outputFormat == utils.FormatTable {
			output.PrintInfo("No VXCs found. Create one with 'megaport vxc buy'.", noColor)
		}
		return nil
	}

	err = printVXCs(filteredVXCs, outputFormat, noColor)
	if err != nil {
		output.PrintError("Failed to print VXCs: %v", noColor, err)
		return fmt.Errorf("error printing VXCs: %v", err)
	}
	return nil
}

func GetVXC(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	output.SetOutputFormat(outputFormat)

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
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
		err = utils.WrapAPIError(err, "VXC", vxcUID)
		output.PrintError("Failed to get VXC: %v", noColor, err)
		return fmt.Errorf("error getting VXC: %w", err)
	}

	export, _ := cmd.Flags().GetBool("export")
	if export {
		cfg := exportVXCConfig(vxc)
		jsonBytes, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return fmt.Errorf("error marshaling export config: %v", err)
		}
		fmt.Println(string(jsonBytes))
		return nil
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

func buildVXCRequest(cmd *cobra.Command, ctx context.Context, client *megaport.Client, noColor bool) (*megaport.BuyVXCRequest, error) {
	interactive, _ := cmd.Flags().GetBool("interactive")
	jsonStr, _ := cmd.Flags().GetString("json")
	jsonFile, _ := cmd.Flags().GetString("json-file")

	flagsProvided := cmd.Flags().Changed("name") || cmd.Flags().Changed("rate-limit") ||
		cmd.Flags().Changed("term") || cmd.Flags().Changed("a-end-uid") ||
		cmd.Flags().Changed("a-end-vlan") || cmd.Flags().Changed("b-end-uid") ||
		cmd.Flags().Changed("b-end-vlan")

	if jsonStr != "" || jsonFile != "" {
		output.PrintInfo("Using JSON input", noColor)
		req, err := buildVXCRequestFromJSON(jsonStr, jsonFile)
		if err != nil {
			output.PrintError("Failed to process JSON input: %v", noColor, err)
			return nil, err
		}
		return req, nil
	} else if flagsProvided {
		output.PrintInfo("Using flag input", noColor)
		req, err := buildVXCRequestFromFlags(cmd, ctx, client.VXCService)
		if err != nil {
			output.PrintError("Failed to process flag input: %v", noColor, err)
			return nil, err
		}
		return req, nil
	} else if interactive {
		output.PrintInfo("Starting interactive mode", noColor)
		req, err := buildVXCRequestFromPrompt(ctx, client.VXCService, noColor)
		if err != nil {
			output.PrintError("Interactive input failed: %v", noColor, err)
			return nil, err
		}
		return req, nil
	}
	output.PrintError("No input provided", noColor)
	return nil, fmt.Errorf("no input provided, use --interactive, --json, or flags to specify VXC details")
}

func BuyVXC(cmd *cobra.Command, args []string, noColor bool) error {
	ctx := context.Background()

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("error logging in: %v", err)
	}

	req, err := buildVXCRequest(cmd, ctx, client, noColor)
	if err != nil {
		return err
	}

	noWait, _ := cmd.Flags().GetBool("no-wait")
	if !noWait {
		req.WaitForProvision = true
		req.WaitForTime = 10 * time.Minute
	}

	validateSpinner := output.PrintResourceValidating("VXC", noColor)

	err = client.VXCService.ValidateVXCOrder(ctx, req)
	validateSpinner.Stop()

	if err != nil {
		output.PrintError("Failed to validate VXC order: %v", noColor, err)
		return err
	}

	jsonStr, _ := cmd.Flags().GetString("json")
	jsonFile, _ := cmd.Flags().GetString("json-file")
	yes, _ := cmd.Flags().GetBool("yes")
	if !yes && jsonStr == "" && jsonFile == "" {
		details := []utils.BuyConfirmDetail{
			{Key: "Name", Value: req.VXCName},
			{Key: "Term", Value: fmt.Sprintf("%d months", req.Term)},
			{Key: "Rate Limit", Value: fmt.Sprintf("%d Mbps", req.RateLimit)},
			{Key: "A-End Port UID", Value: req.PortUID},
		}
		if !utils.BuyConfirmPrompt("VXC", details, noColor) {
			output.PrintInfo("Purchase cancelled", noColor)
			return exitcodes.New(exitcodes.Cancelled, fmt.Errorf("cancelled by user"))
		}
	}

	var spinner *output.Spinner
	if req.WaitForProvision {
		spinner = output.PrintResourceProvisioning("VXC", req.VXCName, noColor)
	} else {
		spinner = output.PrintResourceCreating("VXC", req.VXCName, noColor)
	}

	resp, err := buyVXCFunc(ctx, client, req)

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to buy VXC: %v", noColor, err)
		return err
	}

	output.PrintResourceCreated("VXC", resp.TechnicalServiceUID, noColor)
	return nil
}

func ValidateVXC(cmd *cobra.Command, args []string, noColor bool) error {
	ctx := context.Background()

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("error logging in: %v", err)
	}

	req, err := buildVXCRequest(cmd, ctx, client, noColor)
	if err != nil {
		return err
	}

	spinner := output.PrintResourceValidating("VXC", noColor)
	err = client.VXCService.ValidateVXCOrder(ctx, req)
	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to validate VXC order: %v", noColor, err)
		return err
	}

	output.PrintSuccess("VXC validation passed", noColor)
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
	} else if hasUpdateVXCNonInteractiveFlags(cmd) {
		output.PrintInfo("Using flag input for VXC %s", noColor, formattedUID)
		req, buildErr = buildUpdateVXCRequestFromFlags(cmd)
	} else if interactive {
		output.PrintInfo("Starting interactive mode for VXC %s", noColor, formattedUID)
		req, buildErr = buildUpdateVXCRequestFromPrompt(vxcUID, noColor)
	} else {
		return fmt.Errorf("at least one field must be updated")
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
			return exitcodes.New(exitcodes.Cancelled, fmt.Errorf("cancelled by user"))
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
		err = utils.WrapAPIError(err, "VXC", vxcUID)
		output.PrintError("Failed to delete VXC: %v", noColor, err)
		return err
	}

	output.PrintResourceDeleted("VXC", vxcUID, deleteNow, noColor)
	return nil
}

func ListVXCResourceTags(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	vxcUID := args[0]
	return utils.ListResourceTags("VXC", vxcUID, noColor, outputFormat, func(ctx context.Context, uid string) (map[string]string, error) {
		client, err := config.LoginFunc(ctx)
		if err != nil {
			return nil, err
		}
		return client.VXCService.ListVXCResourceTags(ctx, uid)
	})
}

func UpdateVXCResourceTags(cmd *cobra.Command, args []string, noColor bool) error {
	vxcUID := args[0]
	var client *megaport.Client
	login := func(ctx context.Context) error {
		var err error
		client, err = config.LoginFunc(ctx)
		return err
	}
	return utils.UpdateResourceTags(utils.UpdateTagsOptions{
		ResourceType: "VXC",
		UID:          vxcUID,
		NoColor:      noColor,
		Cmd:          cmd,
		ListFunc: func(ctx context.Context, uid string) (map[string]string, error) {
			if err := login(ctx); err != nil {
				return nil, err
			}
			return client.VXCService.ListVXCResourceTags(ctx, uid)
		},
		UpdateFunc: func(ctx context.Context, uid string, tags map[string]string) error {
			return client.VXCService.UpdateVXCResourceTags(ctx, uid, tags)
		},
	})
}

func GetVXCStatus(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	output.SetOutputFormat(outputFormat)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
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

	if vxc == nil {
		output.PrintError("No VXC found with UID: %s", noColor, vxcUID)
		return fmt.Errorf("no VXC found with UID: %s", vxcUID)
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
