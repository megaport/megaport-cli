package cmd

import (
	"context"
	"fmt"
	"time"

	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

func CreateServiceKey(cmd *cobra.Command, args []string) error {
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
			PrintError("Error parsing start date: %v", err)
			return fmt.Errorf("error parsing start date: %v", err)
		}
		endTime, err := time.Parse("2006-01-02", endDate)
		if err != nil {
			PrintError("Error parsing end date: %v", err)
			return fmt.Errorf("error parsing end date: %v", err)
		}
		validFor = &megaport.ValidFor{
			StartTime: &megaport.Time{Time: startTime},
			EndTime:   &megaport.Time{Time: endTime},
		}
	}

	client, err := Login(ctx)
	if err != nil {
		PrintError("Failed to log in: %v", err)
		return fmt.Errorf("error logging in: %v", err)
	}

	PrintInfo("Creating service key...")
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
		PrintError("Failed to create service key: %v", err)
		return fmt.Errorf("error creating service key: %v", err)
	}

	PrintResourceCreated("Service Key", resp.ServiceKeyUID)
	return nil
}

func UpdateServiceKey(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	key, _ := cmd.Flags().GetString("key")
	productUID, _ := cmd.Flags().GetString("product-uid")
	productID, _ := cmd.Flags().GetInt("product-id")
	singleUse, _ := cmd.Flags().GetBool("single-use")
	active, _ := cmd.Flags().GetBool("active")

	client, err := Login(ctx)
	if err != nil {
		PrintError("Failed to log in: %v", err)
		return fmt.Errorf("error logging in: %v", err)
	}

	PrintInfo("Updating service key...")
	req := &megaport.UpdateServiceKeyRequest{
		Key:        key,
		ProductUID: productUID,
		ProductID:  productID,
		SingleUse:  singleUse,
		Active:     active,
	}

	resp, err := client.ServiceKeyService.UpdateServiceKey(ctx, req)
	if err != nil {
		PrintError("Failed to update service key: %v", err)
		return fmt.Errorf("error updating service key: %v", err)
	}

	if resp.IsUpdated {
		PrintInfo("Service key updated successfully")
	} else {
		PrintWarning("Service key update request was not successful")
	}
	return nil
}

func ListServiceKeys(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	client, err := Login(ctx)
	if err != nil {
		PrintError("Failed to log in: %v", err)
		return fmt.Errorf("error logging in: %v", err)
	}

	PrintInfo("Retrieving service keys...")
	req := &megaport.ListServiceKeysRequest{}
	resp, err := client.ServiceKeyService.ListServiceKeys(ctx, req)
	if err != nil {
		PrintError("Failed to list service keys: %v", err)
		return fmt.Errorf("error listing service keys: %v", err)
	}

	if len(resp.ServiceKeys) == 0 {
		PrintWarning("No service keys found")
	}

	outputs := make([]ServiceKeyOutput, 0, len(resp.ServiceKeys))
	for _, sk := range resp.ServiceKeys {
		output, err := ToServiceKeyOutput(sk)
		if err != nil {
			PrintError("Failed to convert service key: %v", err)
			return fmt.Errorf("error converting service key: %v", err)
		}
		outputs = append(outputs, output)
	}

	return printOutput(outputs, outputFormat)
}

func GetServiceKey(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	client, err := Login(ctx)
	if err != nil {
		PrintError("Failed to log in: %v", err)
		return fmt.Errorf("error logging in: %v", err)
	}

	keyID := args[0]
	formattedUID := formatUID(keyID)

	PrintInfo("Retrieving service key %s...", formattedUID)
	resp, err := client.ServiceKeyService.GetServiceKey(ctx, keyID)
	if err != nil {
		PrintError("Failed to get service key: %v", err)
		return fmt.Errorf("error getting service key: %v", err)
	}

	output, err := ToServiceKeyOutput(resp)
	if err != nil {
		PrintError("Failed to convert service key: %v", err)
		return fmt.Errorf("error converting service key: %v", err)
	}
	return printOutput([]ServiceKeyOutput{output}, outputFormat)
}
