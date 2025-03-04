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
