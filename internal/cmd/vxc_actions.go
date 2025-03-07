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
