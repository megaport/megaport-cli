package ix

import (
	"context"
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

func ListIXs(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	output.SetOutputFormat(outputFormat)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("error logging in: %v", err)
	}

	name, _ := cmd.Flags().GetString("name")
	asn, _ := cmd.Flags().GetInt("asn")
	vlan, _ := cmd.Flags().GetInt("vlan")
	networkServiceType, _ := cmd.Flags().GetString("network-service-type")
	locationID, _ := cmd.Flags().GetInt("location-id")
	rateLimit, _ := cmd.Flags().GetInt("rate-limit")
	includeInactive, _ := cmd.Flags().GetBool("include-inactive")

	req := &megaport.ListIXsRequest{
		IncludeInactive: includeInactive,
	}

	spinner := output.PrintResourceListing("IX", noColor)

	ixs, err := client.IXService.ListIXs(ctx, req)

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to list IXs: %v", noColor, err)
		return fmt.Errorf("error listing IXs: %v", err)
	}

	if !includeInactive {
		var activeIXs []*megaport.IX
		for _, ix := range ixs {
			if ix != nil &&
				ix.ProvisioningStatus != megaport.STATUS_DECOMMISSIONED &&
				ix.ProvisioningStatus != megaport.STATUS_CANCELLED &&
				ix.ProvisioningStatus != "DECOMMISSIONING" {
				activeIXs = append(activeIXs, ix)
			}
		}
		ixs = activeIXs
	}

	filteredIXs := filterIXs(ixs, name, networkServiceType, asn, vlan, locationID, rateLimit)

	if len(filteredIXs) == 0 {
		if outputFormat == utils.FormatTable {
			output.PrintInfo("No IX connections found. Create one with 'megaport ix buy'.", noColor)
		}
		return nil
	}

	err = printIXs(filteredIXs, outputFormat, noColor)
	if err != nil {
		output.PrintError("Failed to print IXs: %v", noColor, err)
		return fmt.Errorf("error printing IXs: %v", err)
	}
	return nil
}

func GetIX(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	output.SetOutputFormat(outputFormat)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	client, err := config.Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %v", err)
	}

	ixUID := args[0]

	spinner := output.PrintResourceGetting("IX", ixUID, noColor)

	ix, err := getIXFunc(ctx, client, ixUID)

	spinner.Stop()

	if err != nil {
		err = utils.WrapAPIError(err, "IX", ixUID)
		output.PrintError("Error getting IX: %v", noColor, err)
		return fmt.Errorf("error getting IX: %w", err)
	}

	err = printIXs([]*megaport.IX{ix}, outputFormat, noColor)
	if err != nil {
		return fmt.Errorf("error printing IXs: %v", err)
	}
	return nil
}

func GetIXStatus(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	output.SetOutputFormat(outputFormat)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("error logging in: %v", err)
	}

	ixUID := args[0]

	spinner := output.PrintResourceGetting("IX", ixUID, noColor)

	ix, err := client.IXService.GetIX(ctx, ixUID)

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to get IX status: %v", noColor, err)
		return fmt.Errorf("error getting IX status: %v", err)
	}

	if ix == nil {
		output.PrintError("No IX found with UID: %s", noColor, ixUID)
		return fmt.Errorf("no IX found with UID: %s", ixUID)
	}

	status := []IXStatus{
		{
			UID:    ix.ProductUID,
			Name:   ix.ProductName,
			Status: ix.ProvisioningStatus,
			Type:   ix.NetworkServiceType,
		},
	}

	return output.PrintOutput(status, outputFormat, noColor)
}

func buildIXRequest(cmd *cobra.Command, noColor bool) (*megaport.BuyIXRequest, error) {
	interactive, _ := cmd.Flags().GetBool("interactive")
	jsonStr, _ := cmd.Flags().GetString("json")
	jsonFile, _ := cmd.Flags().GetString("json-file")

	flagsProvided := cmd.Flags().Changed("name") || cmd.Flags().Changed("product-uid") ||
		cmd.Flags().Changed("network-service-type") || cmd.Flags().Changed("asn") ||
		cmd.Flags().Changed("mac-address") || cmd.Flags().Changed("rate-limit") ||
		cmd.Flags().Changed("vlan")

	if jsonStr != "" || jsonFile != "" {
		output.PrintInfo("Using JSON input", noColor)
		req, err := buildIXRequestFromJSON(jsonStr, jsonFile)
		if err != nil {
			output.PrintError("Failed to process JSON input: %v", noColor, err)
			return nil, err
		}
		return req, nil
	} else if flagsProvided {
		output.PrintInfo("Using flag input", noColor)
		req, err := buildIXRequestFromFlags(cmd)
		if err != nil {
			output.PrintError("Failed to process flag input: %v", noColor, err)
			return nil, err
		}
		return req, nil
	} else if interactive {
		ctx := context.Background()
		req, err := buildIXRequestFromPrompt(ctx, noColor)
		if err != nil {
			return nil, err
		}
		return req, nil
	}
	output.PrintError("No input provided, use --interactive, --json, or flags to specify IX details", noColor)
	return nil, fmt.Errorf("no input provided, use --interactive, --json, or flags to specify IX details")
}

func BuyIX(cmd *cobra.Command, args []string, noColor bool) error {
	ctx := context.Background()

	req, err := buildIXRequest(cmd, noColor)
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

	spinner := output.PrintResourceValidating("IX", noColor)
	err = client.IXService.ValidateIXOrder(ctx, req)
	spinner.Stop()

	if err != nil {
		output.PrintError("Error validating IX order: %v", noColor, err)
		return err
	}

	jsonStr, _ := cmd.Flags().GetString("json")
	jsonFile, _ := cmd.Flags().GetString("json-file")
	yes, _ := cmd.Flags().GetBool("yes")
	if !yes && jsonStr == "" && jsonFile == "" {
		details := []utils.BuyConfirmDetail{
			{Key: "Name", Value: req.Name},
			{Key: "Network Service Type", Value: req.NetworkServiceType},
			{Key: "Rate Limit", Value: fmt.Sprintf("%d Mbps", req.RateLimit)},
			{Key: "ASN", Value: strconv.Itoa(req.ASN)},
		}
		if !utils.BuyConfirmPrompt("IX", details, noColor) {
			output.PrintInfo("Purchase cancelled", noColor)
			return exitcodes.New(exitcodes.Cancelled, fmt.Errorf("cancelled by user"))
		}
	}

	buySpinner := output.PrintResourceCreating("IX", req.Name, noColor)
	resp, err := buyIXFunc(ctx, client, req)
	buySpinner.Stop()

	if err != nil {
		output.PrintError("Error buying IX: %v", noColor, err)
		return err
	}

	output.PrintSuccess("IX created %s", noColor, resp.TechnicalServiceUID)
	return nil
}

func ValidateIX(cmd *cobra.Command, args []string, noColor bool) error {
	ctx := context.Background()

	req, err := buildIXRequest(cmd, noColor)
	if err != nil {
		return err
	}

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Error logging in: %v", noColor, err)
		return err
	}

	spinner := output.PrintResourceValidating("IX", noColor)
	err = client.IXService.ValidateIXOrder(ctx, req)
	spinner.Stop()

	if err != nil {
		output.PrintError("Error validating IX order: %v", noColor, err)
		return err
	}

	output.PrintSuccess("IX validation passed", noColor)
	return nil
}

func UpdateIX(cmd *cobra.Command, args []string, noColor bool) error {
	ctx := context.Background()

	if len(args) == 0 {
		return fmt.Errorf("IX UID is required")
	}

	ixUID := args[0]

	interactive, _ := cmd.Flags().GetBool("interactive")
	jsonStr, _ := cmd.Flags().GetString("json")
	jsonFile, _ := cmd.Flags().GetString("json-file")

	flagsProvided := cmd.Flags().Changed("name") || cmd.Flags().Changed("rate-limit") ||
		cmd.Flags().Changed("cost-centre") || cmd.Flags().Changed("vlan") ||
		cmd.Flags().Changed("mac-address") || cmd.Flags().Changed("asn") ||
		cmd.Flags().Changed("password") || cmd.Flags().Changed("public-graph") ||
		cmd.Flags().Changed("reverse-dns") || cmd.Flags().Changed("a-end-product-uid") ||
		cmd.Flags().Changed("shutdown")

	var req *megaport.UpdateIXRequest
	var err error

	if jsonStr != "" || jsonFile != "" {
		output.PrintInfo("Using JSON input", noColor)
		req, err = buildUpdateIXRequestFromJSON(jsonStr, jsonFile)
		if err != nil {
			output.PrintError("Failed to process JSON input: %v", noColor, err)
			return err
		}
	} else if flagsProvided {
		output.PrintInfo("Using flag input", noColor)
		req, err = buildUpdateIXRequestFromFlags(cmd)
		if err != nil {
			output.PrintError("Failed to process flag input: %v", noColor, err)
			return err
		}
	} else if interactive {
		req, err = buildUpdateIXRequestFromPrompt(ixUID, noColor)
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

	originalIX, err := getIXFunc(ctx, client, ixUID)
	if err != nil {
		output.PrintError("Error getting original IX: %v", noColor, err)
		return err
	}

	updateSpinner := output.PrintResourceUpdating("IX", ixUID, noColor)
	updatedIX, err := updateIXFunc(ctx, client, ixUID, req)
	updateSpinner.Stop()

	if err != nil {
		output.PrintError("Error updating IX: %v", noColor, err)
		return err
	}

	output.PrintResourceUpdated("IX", ixUID, noColor)

	displayIXChanges(originalIX, updatedIX, noColor)

	return nil
}

func DeleteIX(cmd *cobra.Command, args []string, noColor bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	client, err := config.Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %v", err)
	}

	ixUID := args[0]

	deleteNow, err := cmd.Flags().GetBool("now")
	if err != nil {
		return err
	}

	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return err
	}

	if !force {
		confirmMsg := "Are you sure you want to delete IX " + ixUID + "? (y/n): "
		confirmation, err := utils.ResourcePrompt("ix", confirmMsg, noColor)
		if err != nil {
			return err
		}

		if confirmation != "y" && confirmation != "Y" {
			output.PrintInfo("Deletion cancelled", noColor)
			return exitcodes.New(exitcodes.Cancelled, fmt.Errorf("cancelled by user"))
		}
	}

	deleteRequest := &megaport.DeleteIXRequest{
		DeleteNow: deleteNow,
	}

	spinner := output.PrintResourceDeleting("IX", ixUID, noColor)

	err = deleteIXFunc(ctx, client, ixUID, deleteRequest)

	spinner.Stop()

	if err != nil {
		err = utils.WrapAPIError(err, "IX", ixUID)
		return fmt.Errorf("error deleting IX: %w", err)
	}

	output.PrintResourceDeleted("IX", ixUID, deleteNow, noColor)

	return nil
}
