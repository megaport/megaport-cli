package servicekeys

import (
	"context"
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
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	productUID, _ := cmd.Flags().GetString("product-uid")
	productID, _ := cmd.Flags().GetInt("product-id")
	singleUse, _ := cmd.Flags().GetBool("single-use")
	maxSpeed, _ := cmd.Flags().GetInt("max-speed")
	description, _ := cmd.Flags().GetString("description")
	startDate, _ := cmd.Flags().GetString("start-date")
	endDate, _ := cmd.Flags().GetString("end-date")

	if err := validation.ValidateDateRange(startDate, endDate); err != nil {
		output.PrintError(fmt.Sprintf("%v", err), noColor)
		return err
	}

	var validFor *megaport.ValidFor
	if startDate != "" && endDate != "" {
		startTime, _ := time.Parse("2006-01-02", startDate)
		endTime, _ := time.Parse("2006-01-02", endDate)
		validFor = &megaport.ValidFor{
			StartTime: &megaport.Time{Time: startTime},
			EndTime:   &megaport.Time{Time: endTime},
		}
	}

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("error logging in: %v", err)
	}

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
		return fmt.Errorf("error creating service key: %v", err)
	}

	output.PrintResourceCreated("Service Key", resp.ServiceKeyUID, noColor)
	return nil
}

func UpdateServiceKey(cmd *cobra.Command, args []string, noColor bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	key, _ := cmd.Flags().GetString("key")
	productUID, _ := cmd.Flags().GetString("product-uid")
	productID, _ := cmd.Flags().GetInt("product-id")
	singleUse, _ := cmd.Flags().GetBool("single-use")
	active, _ := cmd.Flags().GetBool("active")

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("error logging in: %v", err)
	}

	req := &megaport.UpdateServiceKeyRequest{
		Key:        key,
		ProductUID: productUID,
		ProductID:  productID,
		SingleUse:  singleUse,
		Active:     active,
	}

	spinner := output.PrintResourceUpdating("Service Key", key, noColor)

	resp, err := client.ServiceKeyService.UpdateServiceKey(ctx, req)

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to update service key: %v", noColor, err)
		return fmt.Errorf("error updating service key: %v", err)
	}

	if resp.IsUpdated {
		output.PrintResourceUpdated("Service Key", key, noColor)
	} else {
		output.PrintWarning("Service key update request was not successful", noColor)
	}
	return nil
}

func ListServiceKeys(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	output.SetOutputFormat(outputFormat)

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("error logging in: %v", err)
	}

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
		return fmt.Errorf("error listing service keys: %v", err)
	}

	serviceKeys := resp.ServiceKeys

	limit, _ := cmd.Flags().GetInt("limit")
	if limit > 0 && len(serviceKeys) > limit {
		serviceKeys = serviceKeys[:limit]
	}

	if len(serviceKeys) == 0 {
		if outputFormat == utils.FormatTable {
			output.PrintInfo("No service keys found.", noColor)
		}
		return nil
	}

	outputs := make([]ServiceKeyOutput, 0, len(serviceKeys))
	for _, sk := range serviceKeys {
		op, err := ToServiceKeyOutput(sk)
		if err != nil {
			output.PrintError("Failed to convert service key: %v", noColor, err)
			return fmt.Errorf("error converting service key: %v", err)
		}
		outputs = append(outputs, op)
	}

	return output.PrintOutput(outputs, outputFormat, noColor)
}

func GetServiceKey(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	output.SetOutputFormat(outputFormat)

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("error logging in: %v", err)
	}

	keyID := args[0]

	spinner := output.PrintResourceGetting("Service Key", keyID, noColor)

	resp, err := client.ServiceKeyService.GetServiceKey(ctx, keyID)

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to get service key: %v", noColor, err)
		return fmt.Errorf("error getting service key: %v", err)
	}

	op, err := ToServiceKeyOutput(resp)
	if err != nil {
		output.PrintError("Failed to convert service key: %v", noColor, err)
		return fmt.Errorf("error converting service key: %v", err)
	}
	return output.PrintOutput([]ServiceKeyOutput{op}, outputFormat, noColor)
}
