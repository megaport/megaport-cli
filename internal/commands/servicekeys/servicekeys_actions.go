package servicekeys

import (
	"context"
	"fmt"
	"time"

	"github.com/megaport/megaport-cli/internal/base/config"
	"github.com/megaport/megaport-cli/internal/base/output"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

func CreateServiceKey(cmd *cobra.Command, args []string, noColor bool) error {
	// Create a context with a 30-second timeout for the API call.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	productUID, _ := cmd.Flags().GetString("product-uid")
	productID, _ := cmd.Flags().GetInt("product-id")
	singleUse, _ := cmd.Flags().GetBool("single-use")
	maxSpeed, _ := cmd.Flags().GetInt("max-speed")
	description, _ := cmd.Flags().GetString("description")
	startDate, _ := cmd.Flags().GetString("start-date")
	endDate, _ := cmd.Flags().GetString("end-date")

	var validFor *megaport.ValidFor
	if startDate != "" && endDate != "" {
		startTime, err := time.Parse("2006-01-02", startDate)
		if err != nil {
			output.PrintError("Error parsing start date: %v", noColor, err)
			return fmt.Errorf("error parsing start date: %v", err)
		}
		endTime, err := time.Parse("2006-01-02", endDate)
		if err != nil {
			output.PrintError("Error parsing end date: %v", noColor, err)
			return fmt.Errorf("error parsing end date: %v", err)
		}
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

	output.PrintInfo("Creating service key...", noColor)
	req := &megaport.CreateServiceKeyRequest{
		ProductUID:  productUID,
		ProductID:   productID,
		SingleUse:   singleUse,
		MaxSpeed:    maxSpeed,
		Description: description,
		ValidFor:    validFor,
	}

	resp, err := client.ServiceKeyService.CreateServiceKey(ctx, req)
	if err != nil {
		output.PrintError("Failed to create service key: %v", noColor, err)
		return fmt.Errorf("error creating service key: %v", err)
	}

	output.PrintResourceCreated("Service Key", resp.ServiceKeyUID, noColor)
	return nil
}

func UpdateServiceKey(cmd *cobra.Command, args []string, noColor bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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

	output.PrintInfo("Updating service key...", noColor)
	req := &megaport.UpdateServiceKeyRequest{
		Key:        key,
		ProductUID: productUID,
		ProductID:  productID,
		SingleUse:  singleUse,
		Active:     active,
	}

	resp, err := client.ServiceKeyService.UpdateServiceKey(ctx, req)
	if err != nil {
		output.PrintError("Failed to update service key: %v", noColor, err)
		return fmt.Errorf("error updating service key: %v", err)
	}

	if resp.IsUpdated {
		output.PrintInfo("Service key updated successfully", noColor)
	} else {
		output.PrintWarning("Service key update request was not successful", noColor)
	}
	return nil
}

func ListServiceKeys(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("error logging in: %v", err)
	}

	output.PrintInfo("Retrieving service keys...", noColor)
	req := &megaport.ListServiceKeysRequest{}
	resp, err := client.ServiceKeyService.ListServiceKeys(ctx, req)
	if err != nil {
		output.PrintError("Failed to list service keys: %v", noColor, err)
		return fmt.Errorf("error listing service keys: %v", err)
	}

	if len(resp.ServiceKeys) == 0 {
		output.PrintWarning("No service keys found", noColor)
	}

	outputs := make([]ServiceKeyOutput, 0, len(resp.ServiceKeys))
	for _, sk := range resp.ServiceKeys {
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("error logging in: %v", err)
	}

	keyID := args[0]
	formattedUID := output.FormatUID(keyID, noColor)

	output.PrintInfo("Retrieving service key %s...", noColor, formattedUID)
	resp, err := client.ServiceKeyService.GetServiceKey(ctx, keyID)
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
