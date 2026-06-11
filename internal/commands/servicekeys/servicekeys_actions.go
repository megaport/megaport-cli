package servicekeys

import (
	"fmt"
	"time"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/commands/config"
	"github.com/megaport/megaport-cli/internal/utils"
	"github.com/megaport/megaport-cli/internal/validation"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

func CreateServiceKey(cmd *cobra.Command, args []string, noColor bool) error {
	// Flag read errors are intentionally ignored — flags are registered by the command builder.
	productUID, _ := cmd.Flags().GetString("product-uid")
	productID, _ := cmd.Flags().GetInt("product-id")
	singleUse, _ := cmd.Flags().GetBool("single-use")
	maxSpeed, _ := cmd.Flags().GetInt("max-speed")
	description, _ := cmd.Flags().GetString("description")
	startDate, _ := cmd.Flags().GetString("start-date")
	endDate, _ := cmd.Flags().GetString("end-date")

	if err := validation.ValidateDateRange(startDate, endDate); err != nil {
		output.PrintError("Failed to validate date range: %v", noColor, err)
		return err
	}

	var validFor *megaport.ValidFor
	if startDate != "" && endDate != "" {
		startTime, err := time.Parse("2006-01-02", startDate)
		if err != nil {
			return fmt.Errorf("invalid start date %q: %w", startDate, err)
		}
		endTime, err := time.Parse("2006-01-02", endDate)
		if err != nil {
			return fmt.Errorf("invalid end date %q: %w", endDate, err)
		}
		validFor = &megaport.ValidFor{
			StartTime: &megaport.Time{Time: startTime},
			EndTime:   &megaport.Time{Time: endTime},
		}
	}

	ctx, cancel, client, err := utils.LoginClient(cmd, 90*time.Second, config.Login)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}
	defer cancel()

	active, _ := cmd.Flags().GetBool("active")
	preApproved, _ := cmd.Flags().GetBool("pre-approved")
	vlan, _ := cmd.Flags().GetInt("vlan")

	req := &megaport.CreateServiceKeyRequest{
		ProductUID:  productUID,
		ProductID:   productID,
		SingleUse:   singleUse,
		MaxSpeed:    maxSpeed,
		Description: description,
		ValidFor:    validFor,
		Active:      active,
		PreApproved: preApproved,
		VLAN:        vlan,
	}

	spinner := output.PrintResourceCreating("Service Key", description, noColor)

	resp, err := client.ServiceKeyService.CreateServiceKey(ctx, req)

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to create service key: %v", noColor, err)
		return fmt.Errorf("failed to create service key: %w", err)
	}

	output.PrintResourceCreated("Service Key", resp.ServiceKeyUID, noColor)
	return nil
}

func UpdateServiceKey(cmd *cobra.Command, args []string, noColor bool) error {
	key := args[0]

	if cmd.Flags().Changed("product-uid") && cmd.Flags().Changed("product-id") {
		return fmt.Errorf("--product-uid and --product-id cannot both be set")
	}

	ctx, cancel, client, err := utils.LoginClient(cmd, 90*time.Second, config.Login)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}
	defer cancel()

	// SingleUse and Active are always serialized by the SDK (no omitempty),
	// so merge from the current key to avoid resetting fields the user
	// didn't ask to change.
	current, err := client.ServiceKeyService.GetServiceKey(ctx, key)
	if err != nil {
		output.PrintError("Failed to fetch current service key: %v", noColor, err)
		return fmt.Errorf("failed to fetch current service key: %w", err)
	}

	req := &megaport.UpdateServiceKeyRequest{
		Key:       key,
		SingleUse: current.SingleUse,
		Active:    current.Active,
	}
	if cmd.Flags().Changed("single-use") {
		req.SingleUse, _ = cmd.Flags().GetBool("single-use")
	}
	if cmd.Flags().Changed("active") {
		req.Active, _ = cmd.Flags().GetBool("active")
	}
	// The SDK rejects requests with both ProductUID and ProductID set, so
	// preserve the current ProductUID only when the user provided neither.
	switch {
	case cmd.Flags().Changed("product-uid"):
		req.ProductUID, _ = cmd.Flags().GetString("product-uid")
	case cmd.Flags().Changed("product-id"):
		req.ProductID, _ = cmd.Flags().GetInt("product-id")
	default:
		req.ProductUID = current.ProductUID
	}

	spinner := output.PrintResourceUpdating("Service Key", key, noColor)

	resp, err := client.ServiceKeyService.UpdateServiceKey(ctx, req)

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to update service key: %v", noColor, err)
		return fmt.Errorf("failed to update service key: %w", err)
	}

	if !resp.IsUpdated {
		output.PrintError("Service key update was not applied", noColor)
		return fmt.Errorf("service key update was not applied")
	}

	output.PrintResourceUpdated("Service Key", key, noColor)
	return nil
}

func ListServiceKeys(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	output.SetOutputFormat(outputFormat)

	ctx, cancel, client, err := utils.LoginClient(cmd, 90*time.Second, config.Login)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}
	defer cancel()

	spinner := output.PrintResourceListing("Service Key", noColor)

	req := &megaport.ListServiceKeysRequest{}
	if cmd.Flags().Changed("product-uid") {
		productUID, _ := cmd.Flags().GetString("product-uid")
		req.ProductUID = &productUID
	}
	resp, err := client.ServiceKeyService.ListServiceKeys(ctx, req)

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to list service keys: %v", noColor, err)
		return fmt.Errorf("failed to list service keys: %w", err)
	}

	serviceKeys := resp.ServiceKeys

	limit, _ := cmd.Flags().GetInt("limit")
	if limit < 0 {
		return fmt.Errorf("--limit must be a non-negative integer")
	}
	if limit > 0 && len(serviceKeys) > limit {
		serviceKeys = serviceKeys[:limit]
	}

	if len(serviceKeys) == 0 {
		if outputFormat == utils.FormatTable {
			output.PrintInfo("No service keys found.", noColor)
		}
		return nil
	}

	outputs := make([]serviceKeyOutput, 0, len(serviceKeys))
	for _, sk := range serviceKeys {
		op, err := toServiceKeyOutput(sk)
		if err != nil {
			output.PrintError("Failed to convert service key: %v", noColor, err)
			return fmt.Errorf("failed to convert service key: %w", err)
		}
		outputs = append(outputs, op)
	}

	return output.PrintOutput(outputs, outputFormat, noColor)
}

func GetServiceKey(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	output.SetOutputFormat(outputFormat)

	ctx, cancel, client, err := utils.LoginClient(cmd, 90*time.Second, config.Login)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}
	defer cancel()

	keyID := args[0]

	spinner := output.PrintResourceGetting("Service Key", keyID, noColor)

	resp, err := client.ServiceKeyService.GetServiceKey(ctx, keyID)

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to get service key: %v", noColor, err)
		return fmt.Errorf("failed to get service key: %w", err)
	}

	op, err := toServiceKeyOutput(resp)
	if err != nil {
		output.PrintError("Failed to convert service key: %v", noColor, err)
		return fmt.Errorf("failed to convert service key: %w", err)
	}
	return output.PrintOutput([]serviceKeyOutput{op}, outputFormat, noColor)
}
