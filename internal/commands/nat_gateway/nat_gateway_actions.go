package nat_gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/megaport/megaport-cli/internal/base/exitcodes"
	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/commands/config"
	"github.com/megaport/megaport-cli/internal/utils"
	"github.com/megaport/megaport-cli/internal/validation"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

// CreateNATGateway handles the nat-gateway create command.
func CreateNATGateway(cmd *cobra.Command, args []string, noColor bool) error {
	ctx, cancel := utils.ContextFromCmdWithDefault(cmd, utils.DefaultMutationTimeout)
	defer cancel()

	interactive, _ := cmd.Flags().GetBool("interactive")
	jsonStr, _ := cmd.Flags().GetString("json")
	jsonFile, _ := cmd.Flags().GetString("json-file")
	flagsProvided := cmd.Flags().Changed("name") || cmd.Flags().Changed("term") ||
		cmd.Flags().Changed("speed") || cmd.Flags().Changed("location-id") ||
		cmd.Flags().Changed("session-count") || cmd.Flags().Changed("diversity-zone") ||
		cmd.Flags().Changed("promo-code") || cmd.Flags().Changed("service-level-reference") ||
		cmd.Flags().Changed("auto-renew")

	var req *megaport.CreateNATGatewayRequest
	var err error

	if jsonStr != "" || jsonFile != "" {
		output.PrintInfo("Using JSON input", noColor)
		req, err = processJSONCreateNATGatewayInput(jsonStr, jsonFile)
	} else if flagsProvided {
		output.PrintInfo("Using flag input", noColor)
		req, err = processFlagCreateNATGatewayInput(cmd)
	} else if interactive {
		req, err = promptForCreateNATGatewayDetails(noColor)
	} else {
		return fmt.Errorf("provide --interactive, --json/--json-file, or required flags (--name, --term, --speed, --location-id)")
	}
	if err != nil {
		output.PrintError("Failed to process input: %v", noColor, err)
		return err
	}

	yes, _ := cmd.Flags().GetBool("yes")

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}

	if !yes && jsonStr == "" && jsonFile == "" {
		details := []utils.BuyConfirmDetail{
			{Key: "Name", Value: req.ProductName},
			{Key: "Term", Value: fmt.Sprintf("%d months", req.Term)},
			{Key: "Speed", Value: fmt.Sprintf("%d Mbps", req.Speed)},
			{Key: "Location ID", Value: fmt.Sprintf("%d", req.LocationID)},
		}
		if req.Config.SessionCount > 0 {
			details = append(details, utils.BuyConfirmDetail{Key: "Session Count", Value: fmt.Sprintf("%d", req.Config.SessionCount)})
		}
		if !utils.BuyConfirmPrompt("NAT Gateway", details, noColor) {
			output.PrintInfo("Purchase cancelled", noColor)
			return exitcodes.New(exitcodes.Cancelled, fmt.Errorf("cancelled by user"))
		}
	}

	spinner := output.PrintResourceCreating("NAT Gateway", req.ProductName, noColor)

	var gw *megaport.NATGateway
	err = utils.WithRetry(ctx, func(ctx context.Context) error {
		var e error
		gw, e = createNATGatewayFunc(ctx, client, req)
		return e
	})
	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to create NAT Gateway: %v", noColor, err)
		return err
	}
	if gw == nil || gw.ProductUID == "" {
		err = fmt.Errorf("service returned no NAT Gateway UID")
		output.PrintError("Failed to create NAT Gateway: %v", noColor, err)
		return err
	}

	output.PrintResourceCreated("NAT Gateway", gw.ProductUID, noColor)
	return nil
}

// GetNATGateway handles the nat-gateway get command.
func GetNATGateway(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	output.SetOutputFormat(outputFormat)

	watch, _ := cmd.Flags().GetBool("watch")
	if watch {
		return watchGetNATGateway(cmd, args, noColor, outputFormat)
	}

	ctx, cancel, client, err := utils.LoginClient(cmd, 90*time.Second, config.Login)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}
	defer cancel()

	uid := args[0]
	spinner := output.PrintResourceGetting("NAT Gateway", uid, noColor)
	gw, err := getNATGatewayFunc(ctx, client, uid)
	spinner.Stop()

	if err != nil {
		err = utils.WrapAPIError(err, "NAT Gateway", uid)
		output.PrintError("Failed to get NAT Gateway: %v", noColor, err)
		return fmt.Errorf("failed to get NAT Gateway: %w", err)
	}
	if gw == nil {
		output.PrintError("No NAT Gateway found with UID: %s", noColor, uid)
		return fmt.Errorf("no NAT Gateway found with UID: %s", uid)
	}

	export, _ := cmd.Flags().GetBool("export")
	if export {
		cfg := exportNATGatewayConfig(gw)
		jsonBytes, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal export config: %w", err)
		}
		fmt.Println(string(jsonBytes))
		return nil
	}

	return printNATGateways([]*megaport.NATGateway{gw}, outputFormat, noColor)
}

func watchGetNATGateway(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	uid := args[0]
	return utils.WatchResource(cmd, "NAT Gateway", uid, noColor, outputFormat, config.Login,
		func(pollCtx context.Context, client *megaport.Client) (string, error) {
			gw, err := getNATGatewayFunc(pollCtx, client, uid)
			if err != nil {
				return "", err
			}
			if gw == nil {
				return "", fmt.Errorf("no NAT Gateway found with UID: %s", uid)
			}
			err = printNATGateways([]*megaport.NATGateway{gw}, outputFormat, noColor)
			return gw.ProvisioningStatus, err
		})
}

// ListNATGateways handles the nat-gateway list command.
func ListNATGateways(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	output.SetOutputFormat(outputFormat)

	ctx, cancel, client, err := utils.LoginClient(cmd, 90*time.Second, config.Login)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}
	defer cancel()

	locationID, _ := cmd.Flags().GetInt("location-id")
	name, _ := cmd.Flags().GetString("name")
	includeInactive, _ := cmd.Flags().GetBool("include-inactive")

	spinner := output.PrintResourceListing("NAT Gateway", noColor)
	gateways, err := listNATGatewaysFunc(ctx, client)
	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to list NAT Gateways: %v", noColor, err)
		return fmt.Errorf("failed to list NAT Gateways: %w", err)
	}

	if !includeInactive {
		var active []*megaport.NATGateway
		for _, gw := range gateways {
			if gw != nil &&
				gw.ProvisioningStatus != megaport.STATUS_DECOMMISSIONED &&
				gw.ProvisioningStatus != megaport.STATUS_CANCELLED &&
				gw.ProvisioningStatus != utils.StatusDecommissioning {
				active = append(active, gw)
			}
		}
		gateways = active
	}

	filtered := filterNATGateways(gateways, locationID, name)
	limit, _ := cmd.Flags().GetInt("limit")
	return utils.ApplyLimitAndPrint(filtered, limit, outputFormat, noColor,
		"No NAT Gateways found. Create one with 'megaport-cli nat-gateway create'.", printNATGateways)
}

// UpdateNATGateway handles the nat-gateway update command.
// It fetches the original resource first and merges partial input so users
// can update individual fields without providing every required field.
func UpdateNATGateway(cmd *cobra.Command, args []string, noColor bool) error {
	ctx, cancel := utils.ContextFromCmdWithDefault(cmd, utils.DefaultMutationTimeout)
	defer cancel()

	uid := args[0]
	interactive, _ := cmd.Flags().GetBool("interactive")
	jsonStr, _ := cmd.Flags().GetString("json")
	jsonFile, _ := cmd.Flags().GetString("json-file")
	flagsProvided := cmd.Flags().Changed("name") || cmd.Flags().Changed("term") ||
		cmd.Flags().Changed("speed") || cmd.Flags().Changed("location-id") ||
		cmd.Flags().Changed("session-count") || cmd.Flags().Changed("diversity-zone") ||
		cmd.Flags().Changed("promo-code") || cmd.Flags().Changed("service-level-reference") ||
		cmd.Flags().Changed("auto-renew")

	if jsonStr == "" && jsonFile == "" && !flagsProvided && !interactive {
		return fmt.Errorf("at least one field must be updated")
	}

	// Login and fetch the original gateway to use as defaults for unset fields.
	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}

	originalGW, err := getNATGatewayFunc(ctx, client, uid)
	if err != nil {
		output.PrintError("Failed to get original NAT Gateway: %v", noColor, err)
		return err
	}
	if originalGW == nil {
		output.PrintError("NAT Gateway %s not found", noColor, uid)
		return fmt.Errorf("NAT Gateway %s not found", uid)
	}

	// Build the request from user input.
	var req *megaport.UpdateNATGatewayRequest

	if jsonStr != "" || jsonFile != "" {
		output.PrintInfo("Using JSON input", noColor)
		req, err = processJSONUpdateNATGatewayInput(jsonStr, jsonFile, uid)
	} else if flagsProvided {
		output.PrintInfo("Using flag input", noColor)
		req, err = processFlagUpdateNATGatewayInput(cmd, uid)
	} else if interactive {
		req, err = promptForUpdateNATGatewayDetails(uid, noColor)
	}
	if err != nil {
		output.PrintError("Failed to process input: %v", noColor, err)
		return err
	}

	// Merge: fill in unset fields from the original gateway so partial
	// updates work without requiring every field.
	mergeUpdateDefaults(req, originalGW)

	if err := validation.ValidateUpdateNATGatewayRequest(req); err != nil {
		output.PrintError("Validation failed: %v", noColor, err)
		return err
	}

	spinner := output.PrintResourceUpdating("NAT Gateway", uid, noColor)
	var updatedGW *megaport.NATGateway
	err = utils.WithRetry(ctx, func(ctx context.Context) error {
		var e error
		updatedGW, e = updateNATGatewayFunc(ctx, client, req)
		return e
	})
	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to update NAT Gateway: %v", noColor, err)
		return err
	}

	output.PrintResourceUpdated("NAT Gateway", uid, noColor)
	displayNATGatewayChanges(originalGW, updatedGW, noColor)
	return nil
}

// mergeUpdateDefaults fills zero-valued fields in the update request with
// values from the original NAT Gateway, enabling partial updates.
func mergeUpdateDefaults(req *megaport.UpdateNATGatewayRequest, original *megaport.NATGateway) {
	if original == nil {
		return
	}
	if req.ProductName == "" {
		req.ProductName = original.ProductName
	}
	if req.LocationID == 0 {
		req.LocationID = original.LocationID
	}
	if req.Speed == 0 {
		req.Speed = original.Speed
	}
	if req.Term == 0 {
		req.Term = original.Term
	}
	if req.Config.SessionCount == 0 {
		req.Config.SessionCount = original.Config.SessionCount
	}
	if req.Config.DiversityZone == "" {
		req.Config.DiversityZone = original.Config.DiversityZone
	}
	if req.Config.ASN == 0 {
		req.Config.ASN = original.Config.ASN
	}
	// AutoRenewTerm and BGPShutdownDefault have no omitempty — false is always
	// serialised. Inherit from the original so partial updates don't silently
	// disable these. Users who need to explicitly set them to false can do so
	// via JSON input with the field present.
	if !req.AutoRenewTerm {
		req.AutoRenewTerm = original.AutoRenewTerm
	}
	if !req.Config.BGPShutdownDefault {
		req.Config.BGPShutdownDefault = original.Config.BGPShutdownDefault
	}
}

// DeleteNATGateway handles the nat-gateway delete command.
func DeleteNATGateway(cmd *cobra.Command, args []string, noColor bool) error {
	ctx, cancel, client, err := utils.LoginClient(cmd, 90*time.Second, config.Login)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}
	defer cancel()

	uid := args[0]
	force, _ := cmd.Flags().GetBool("force")

	if !force {
		if !utils.ConfirmPrompt("Are you sure you want to delete NAT Gateway "+uid+"? ", noColor) {
			output.PrintInfo("Deletion cancelled", noColor)
			return exitcodes.New(exitcodes.Cancelled, fmt.Errorf("cancelled by user"))
		}
	}

	spinner := output.PrintResourceDeleting("NAT Gateway", uid, noColor)
	err = utils.WithRetry(ctx, func(ctx context.Context) error {
		return deleteNATGatewayFunc(ctx, client, uid)
	})
	spinner.Stop()

	if err != nil {
		err = utils.WrapAPIError(err, "NAT Gateway", uid)
		output.PrintError("Failed to delete NAT Gateway: %v", noColor, err)
		return fmt.Errorf("failed to delete NAT Gateway: %w", err)
	}

	output.PrintResourceDeleted("NAT Gateway", uid, true, noColor)
	return nil
}

// ListNATGatewaySessions handles the nat-gateway list-sessions command.
func ListNATGatewaySessions(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	output.SetOutputFormat(outputFormat)

	ctx, cancel, client, err := utils.LoginClient(cmd, 90*time.Second, config.Login)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}
	defer cancel()

	spinner := output.PrintResourceListing("NAT Gateway sessions", noColor)
	sessions, err := listNATGatewaySessionsFunc(ctx, client)
	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to list NAT Gateway sessions: %v", noColor, err)
		return fmt.Errorf("failed to list NAT Gateway sessions: %w", err)
	}

	if len(sessions) == 0 {
		output.PrintInfo("No NAT Gateway session options found", noColor)
		return nil
	}

	return printNATGatewaySessions(sessions, outputFormat, noColor)
}

// GetNATGatewayTelemetry handles the nat-gateway telemetry command.
func GetNATGatewayTelemetry(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	output.SetOutputFormat(outputFormat)

	ctx, cancel, client, err := utils.LoginClient(cmd, 90*time.Second, config.Login)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}
	defer cancel()

	uid := args[0]
	typesStr, _ := cmd.Flags().GetString("types")
	types := parseTelemetryTypes(typesStr)
	if len(types) == 0 {
		return fmt.Errorf("--types is required (e.g. --types BITS,PACKETS,SPEED)")
	}

	req := &megaport.GetNATGatewayTelemetryRequest{
		ProductUID: uid,
		Types:      types,
	}

	if cmd.Flags().Changed("days") {
		dInt, _ := cmd.Flags().GetInt("days")
		if dInt < 1 || dInt > 180 {
			return fmt.Errorf("--days must be between 1 and 180")
		}
		d := int32(dInt) //nolint:gosec // validated above
		req.Days = &d
	} else if cmd.Flags().Changed("from") || cmd.Flags().Changed("to") {
		fromStr, _ := cmd.Flags().GetString("from")
		toStr, _ := cmd.Flags().GetString("to")
		if fromStr == "" || toStr == "" {
			return fmt.Errorf("--from and --to must both be provided together")
		}
		from, err := time.Parse(time.RFC3339, fromStr)
		if err != nil {
			return fmt.Errorf("invalid --from time (use RFC3339 format, e.g. 2024-01-01T00:00:00Z): %w", err)
		}
		to, err := time.Parse(time.RFC3339, toStr)
		if err != nil {
			return fmt.Errorf("invalid --to time (use RFC3339 format, e.g. 2024-01-07T00:00:00Z): %w", err)
		}
		req.From = &from
		req.To = &to
	}

	spinner := output.PrintResourceGetting("NAT Gateway telemetry", uid, noColor)
	resp, err := getNATGatewayTelemetryFunc(ctx, client, req)
	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to get NAT Gateway telemetry: %v", noColor, err)
		return fmt.Errorf("failed to get NAT Gateway telemetry: %w", err)
	}
	if resp == nil {
		output.PrintError("NAT Gateway telemetry: empty response received", noColor)
		return fmt.Errorf("NAT Gateway telemetry: empty response received")
	}

	return printNATGatewayTelemetry(resp, outputFormat, noColor)
}
