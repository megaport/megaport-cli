package cmd

import (
	"context"
	"fmt"
	"time"

	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

func GetVXC(cmd *cobra.Command, args []string) error {
	// Create a context with a 30-second timeout for the API call.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Log into the Megaport API.
	client, err := Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %v", err)
	}

	// Retrieve the VXC UID from the command line arguments.
	vxcUID := args[0]

	// Retrieve VXC details using the API client.
	vxc, err := client.VXCService.GetVXC(ctx, vxcUID)
	if err != nil {
		return fmt.Errorf("error getting VXC: %v", err)
	}

	// Print the VXC details using the desired output format.
	err = printVXCs([]*megaport.VXC{vxc}, outputFormat)
	if err != nil {
		return fmt.Errorf("error printing VXCs: %v", err)
	}
	return nil
}

// hasUpdateVXCNonInteractiveFlags checks if any non-interactive flags are set for updating a VXC.
func hasUpdateVXCNonInteractiveFlags(cmd *cobra.Command) bool {
	flagNames := []string{"name", "rate-limit", "a-end-vlan", "b-end-vlan", "a-end-location", "b-end-location", "locked"}
	for _, name := range flagNames {
		if cmd.Flags().Changed(name) {
			return true
		}
	}
	return false
}

// BuyVXC purchases a new Virtual Cross Connect
func BuyVXC(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Determine input mode and build request
	var req *megaport.BuyVXCRequest
	var err error

	interactive, _ := cmd.Flags().GetBool("interactive")
	jsonStr, _ := cmd.Flags().GetString("json")
	jsonFilePath, _ := cmd.Flags().GetString("json-file")

	// Call the BuyVXC method
	client, err := Login(ctx)
	if err != nil {
		return err
	}

	// Check if we have JSON input first
	if jsonStr != "" || jsonFilePath != "" {
		PrintInfo("Using JSON input")
		req, err = buildVXCRequestFromJSON(jsonStr, jsonFilePath)
	} else if interactive {
		PrintInfo("Starting interactive mode")
		req, err = buildVXCRequestFromPrompt(ctx, client.VXCService)
	} else {
		PrintInfo("Using flag input")
		req, err = buildVXCRequestFromFlags(cmd, ctx, client.VXCService)
	}

	if err != nil {
		return err
	}

	if req == nil {
		return fmt.Errorf("no input provided")
	}

	PrintInfo("Buying VXC...")
	if buyVXCFunc == nil {
		return fmt.Errorf("internal error: buyVXCFunc is nil")
	}

	resp, err := buyVXCFunc(ctx, client, req)
	if err != nil {
		PrintError("Failed to purchase VXC: %v", err)
		return err
	}

	PrintResourceCreated("VXC", resp.TechnicalServiceUID)
	return nil
}

// UpdateVXC updates an existing VXC with the provided configuration
func UpdateVXC(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Get the VXC UID from args
	vxcUID := args[0]
	formattedUID := formatUID(vxcUID)

	// Determine input mode and build request
	var req *megaport.UpdateVXCRequest
	var err error

	interactive, _ := cmd.Flags().GetBool("interactive")
	jsonStr, _ := cmd.Flags().GetString("json")
	jsonFilePath, _ := cmd.Flags().GetString("json-file")

	if jsonStr != "" || jsonFilePath != "" {
		PrintInfo("Using JSON input for VXC %s", formattedUID)
		req, err = buildUpdateVXCRequestFromJSON(jsonStr, jsonFilePath)
	} else if interactive || !hasUpdateVXCNonInteractiveFlags(cmd) {
		PrintInfo("Starting interactive mode for VXC %s", formattedUID)
		req, err = buildUpdateVXCRequestFromPrompt(vxcUID)
	} else {
		PrintInfo("Using flag input for VXC %s", formattedUID)
		req, err = buildUpdateVXCRequestFromFlags(cmd)
	}

	if err != nil {
		return err
	}

	if req == nil {
		return fmt.Errorf("no update parameters provided")
	}

	// Set wait for update options if not already set
	if !req.WaitForUpdate {
		req.WaitForUpdate = true
		req.WaitForTime = 5 * time.Minute
	}

	// Login to API
	client, err := Login(ctx)
	if err != nil {
		return err
	}

	// Call update API
	PrintInfo("Updating VXC %s...", formattedUID)
	err = updateVXCFunc(ctx, client, vxcUID, req)
	if err != nil {
		PrintError("Failed to update VXC: %v", err)
		return fmt.Errorf("failed to update VXC: %v", err)
	}

	PrintResourceUpdated("VXC", vxcUID)
	return nil
}

// DeleteVXC deletes a VXC with the provided UID
func DeleteVXC(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	vxcUID := args[0]
	formattedUID := formatUID(vxcUID)

	// Parse command flags
	force, _ := cmd.Flags().GetBool("force")
	deleteNow, _ := cmd.Flags().GetBool("now")

	// If not forced, ask for confirmation
	if !force {
		message := fmt.Sprintf("Are you sure you want to delete VXC %s?", formattedUID)
		if !confirmPrompt(message) {
			PrintInfo("Deletion cancelled")
			return nil
		}
	}

	client, err := loginFunc(ctx)
	if err != nil {
		PrintError("Failed to log in: %v", err)
		return err
	}

	req := &megaport.DeleteVXCRequest{
		DeleteNow: deleteNow,
	}

	PrintInfo("Deleting VXC %s...", formattedUID)
	if err := deleteVXCFunc(ctx, client, vxcUID, req); err != nil {
		PrintError("Failed to delete VXC: %v", err)
		return err
	}

	PrintResourceDeleted("VXC", vxcUID, deleteNow)
	return nil
}
