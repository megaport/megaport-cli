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

func BuyVXC(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Determine input mode and build request
	var req *megaport.BuyVXCRequest
	var err error

	interactive, _ := cmd.Flags().GetBool("interactive")
	jsonStr, _ := cmd.Flags().GetString("json")
	jsonFilePath, _ := cmd.Flags().GetString("json-file")

	// Check if we have JSON input first
	if jsonStr != "" || jsonFilePath != "" {
		// JSON mode
		req, err = buildVXCRequestFromJSON(jsonStr, jsonFilePath)
	} else if interactive || !hasNonInteractiveFlags(cmd) {
		// Interactive mode if explicitly requested or if no other input mode is provided
		req, err = buildVXCRequestFromPrompt()
	} else {
		// Flag mode - use when interactive is false and flags are provided
		req, err = buildVXCRequestFromFlags(cmd)
	}

	if err != nil {
		return err
	}

	if req == nil {
		return fmt.Errorf("no input provided")
	}

	// Call the BuyVXC method
	client, err := Login(ctx)
	if err != nil {
		return err
	}

	fmt.Println("Buying VXC...")
	if buyVXCFunc == nil {
		return fmt.Errorf("internal error: buyVXCFunc is nil")
	}

	resp, err := buyVXCFunc(ctx, client, req)
	if err != nil {
		return err
	}

	fmt.Printf("VXC purchased successfully - UID: %s\n", resp.TechnicalServiceUID)
	return nil
}

// hasNonInteractiveFlags checks if any of the non-interactive flags were set
func hasNonInteractiveFlags(cmd *cobra.Command) bool {
	// List of flags that indicate non-interactive mode
	nonInteractiveFlags := []string{
		"a-end-uid", "b-end-uid", "name", "rate-limit", "term",
		"a-end-vlan", "b-end-vlan", "a-end-inner-vlan", "b-end-inner-vlan",
		"a-end-vnic-index", "b-end-vnic-index", "promo-code",
		"service-key", "cost-centre", "a-end-partner-config", "b-end-partner-config",
	}

	// Check if any of these flags were explicitly set
	for _, flag := range nonInteractiveFlags {
		if cmd.Flags().Changed(flag) {
			return true
		}
	}

	return false
}

// UpdateVXC updates an existing VXC with the provided configuration
func UpdateVXC(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Get the VXC UID from args
	vxcUID := args[0]

	// Determine input mode and build request
	var req *megaport.UpdateVXCRequest
	var err error

	interactive, _ := cmd.Flags().GetBool("interactive")
	jsonStr, _ := cmd.Flags().GetString("json")
	jsonFilePath, _ := cmd.Flags().GetString("json-file")

	if jsonStr != "" || jsonFilePath != "" {
		// JSON mode
		req, err = buildUpdateVXCRequestFromJSON(jsonStr, jsonFilePath)
	} else if interactive || !hasUpdateVXCNonInteractiveFlags(cmd) {
		// Interactive mode
		req, err = buildUpdateVXCRequestFromPrompt(vxcUID)
	} else {
		// Flag mode
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
	fmt.Println("Updating VXC...")
	err = updateVXCFunc(ctx, client, vxcUID, req)
	if err != nil {
		return fmt.Errorf("failed to update VXC: %v", err)
	}

	fmt.Println("VXC updated successfully.")
	return nil
}

// hasUpdateVXCNonInteractiveFlags checks if any of the non-interactive flags for VXC update were set
func hasUpdateVXCNonInteractiveFlags(cmd *cobra.Command) bool {
	// List of flags that indicate non-interactive mode for VXC update
	nonInteractiveFlags := []string{
		"name", "rate-limit", "term", "cost-centre", "shutdown",
		"a-end-vlan", "b-end-vlan", "a-end-inner-vlan", "b-end-inner-vlan",
		"a-end-uid", "b-end-uid", "a-end-partner-config", "b-end-partner-config",
	}

	// Check if any of these flags were explicitly set
	for _, flag := range nonInteractiveFlags {
		if cmd.Flags().Changed(flag) {
			return true
		}
	}

	return false
}

func DeleteVXC(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	client, err := loginFunc(ctx)
	if err != nil {
		return err
	}

	vxcUID := args[0]
	req := &megaport.DeleteVXCRequest{
		DeleteNow: true,
	}

	// The key fix - check error BEFORE printing success message
	if err := deleteVXCFunc(ctx, client, vxcUID, req); err != nil {
		return err // Return error without printing success message
	}

	_, err = consolePrintf("VXC %s deleted successfully\n", vxcUID)
	if err != nil {
		return err
	}
	return nil
}
