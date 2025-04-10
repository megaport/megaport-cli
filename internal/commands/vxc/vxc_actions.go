package vxc

import (
	"context"
	"fmt"
	"time"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/commands/config"
	"github.com/megaport/megaport-cli/internal/utils"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

func GetVXC(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	// Create a context with a 30-second timeout for the API call.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Log into the Megaport API.
	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("error logging in: %v", err)
	}

	// Retrieve the VXC UID from the command line arguments.
	vxcUID := args[0]

	// Start spinner for getting VXC details
	spinner := output.PrintResourceGetting("VXC", vxcUID, noColor)

	// Retrieve VXC details using the API client.
	vxc, err := client.VXCService.GetVXC(ctx, vxcUID)

	// Stop spinner
	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to get VXC: %v", noColor, err)
		return fmt.Errorf("error getting VXC: %v", err)
	}

	// Print the VXC details using the desired output format.
	err = printVXCs([]*megaport.VXC{vxc}, outputFormat, noColor)
	if err != nil {
		output.PrintError("Failed to print VXCs: %v", noColor, err)
		return fmt.Errorf("error printing VXCs: %v", err)
	}
	return nil
}

// hasUpdateVXCNonInteractiveFlags checks if any non-interactive flags are set for updating a VXC.
var hasUpdateVXCNonInteractiveFlags = func(cmd *cobra.Command) bool {
	flagNames := []string{"name", "rate-limit", "a-end-vlan", "b-end-vlan", "a-end-location", "b-end-location", "locked"}
	for _, name := range flagNames {
		if cmd.Flags().Changed(name) {
			return true
		}
	}
	return false
}

// BuyVXC handles purchasing a new VXC
func BuyVXC(cmd *cobra.Command, args []string, noColor bool) error {
	ctx := context.Background()

	// Determine which mode to use
	interactive, _ := cmd.Flags().GetBool("interactive")
	jsonStr, _ := cmd.Flags().GetString("json")
	jsonFile, _ := cmd.Flags().GetString("json-file")

	// Check if any flag-based parameters are provided
	flagsProvided := cmd.Flags().Changed("name") || cmd.Flags().Changed("rate-limit") ||
		cmd.Flags().Changed("term") || cmd.Flags().Changed("a-end-uid") ||
		cmd.Flags().Changed("a-end-vlan") || cmd.Flags().Changed("b-end-uid") ||
		cmd.Flags().Changed("b-end-vlan")

	var req *megaport.BuyVXCRequest
	var err error

	// Process input based on mode priority: JSON > Flags > Interactive
	if jsonStr != "" || jsonFile != "" {
		// JSON mode
		output.PrintInfo("Using JSON input", noColor)
		req, err = buildVXCRequestFromJSON(jsonStr, jsonFile)
		if err != nil {
			output.PrintError("Failed to process JSON input: %v", noColor, err)
			return err
		}
	} else if flagsProvided {
		// Flag mode
		output.PrintInfo("Using flag input", noColor)
		client, err := config.Login(ctx)
		if err != nil {
			output.PrintError("Failed to log in: %v", noColor, err)
			return err
		}
		req, err = buildVXCRequestFromFlags(cmd, ctx, client.VXCService)
		if err != nil {
			output.PrintError("Failed to process flag input: %v", noColor, err)
			return err
		}
	} else if interactive {
		// Interactive mode
		output.PrintInfo("Starting interactive mode", noColor)
		client, err := config.Login(ctx)
		if err != nil {
			output.PrintError("Failed to log in: %v", noColor, err)
			return err
		}
		req, err = buildVXCRequestFromPrompt(ctx, client.VXCService, noColor)
		if err != nil {
			output.PrintError("Interactive input failed: %v", noColor, err)
			return err
		}
	} else {
		output.PrintError("No input provided", noColor)
		return fmt.Errorf("no input provided, use --interactive, --json, or flags to specify VXC details")
	}

	// Set common defaults
	req.WaitForProvision = true
	req.WaitForTime = 10 * time.Minute

	// Call the BuyVXC method
	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}

	// Start spinner for creating VXC
	spinner := output.PrintResourceCreating("VXC", req.VXCName, noColor)

	// Call the API
	resp, err := buyVXCFunc(ctx, client, req)

	// Stop spinner
	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to buy VXC: %v", noColor, err)
		return err
	}

	output.PrintResourceCreated("VXC", resp.TechnicalServiceUID, noColor)
	return nil
}

// UpdateVXC updates an existing VXC with the provided configuration
func UpdateVXC(cmd *cobra.Command, args []string, noColor bool) error {
	ctx := context.Background()

	// Get the VXC UID from args
	vxcUID := args[0]
	formattedUID := output.FormatUID(vxcUID, noColor)

	// Log into the API first so we can retrieve the original VXC
	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return err
	}

	// Start spinner for getting original VXC details
	getSpinner := output.PrintResourceGetting("VXC", vxcUID, noColor)

	// Retrieve the original VXC for later comparison
	originalVXC, err := client.VXCService.GetVXC(ctx, vxcUID)

	// Stop spinner
	getSpinner.Stop()

	if err != nil {
		output.PrintError("Failed to retrieve original VXC details: %v", noColor, err)
		return fmt.Errorf("failed to retrieve original VXC details: %v", err)
	}

	// Determine input mode and build request
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

	// Start spinner for updating VXC
	updateSpinner := output.PrintResourceUpdating("VXC", vxcUID, noColor)

	// Call update API
	err = updateVXCFunc(ctx, client, vxcUID, req)

	// Stop spinner
	updateSpinner.Stop()

	if err != nil {
		output.PrintError("Failed to update VXC: %v", noColor, err)
		return fmt.Errorf("failed to update VXC: %v", err)
	}

	// Start spinner for getting updated VXC details
	getUpdatedSpinner := output.PrintResourceGetting("VXC", vxcUID, noColor)

	// Retrieve the updated VXC for comparison
	updatedVXC, err := getVXCFunc(ctx, client, vxcUID)

	// Stop spinner
	getUpdatedSpinner.Stop()

	if err != nil {
		output.PrintError("VXC was updated but failed to retrieve updated details: %v", noColor, err)
		output.PrintResourceUpdated("VXC", vxcUID, noColor)
		return nil
	}

	// Print success message
	output.PrintResourceUpdated("VXC", vxcUID, noColor)

	// Show comparison of changes
	displayVXCChanges(originalVXC, updatedVXC, noColor)

	return nil
}

// DeleteVXC deletes a VXC with the provided UID
func DeleteVXC(cmd *cobra.Command, args []string, noColor bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	vxcUID := args[0]
	formattedUID := output.FormatUID(vxcUID, noColor)

	// Parse command flags
	force, _ := cmd.Flags().GetBool("force")
	deleteNow, _ := cmd.Flags().GetBool("now")

	// If not forced, ask for confirmation
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

	// Start spinner for deleting VXC
	spinner := output.PrintResourceDeleting("VXC", vxcUID, noColor)

	// Call the API
	err = deleteVXCFunc(ctx, client, vxcUID, req)

	// Stop spinner
	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to delete VXC: %v", noColor, err)
		return err
	}

	output.PrintResourceDeleted("VXC", vxcUID, deleteNow, noColor)
	return nil
}
