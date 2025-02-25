package cmd

import (
	"context"
	"fmt"
	"time"

	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

// mcrCmd is the parent command for all operations related to Megaport Cloud Routers (MCRs).
// It serves as a container for subcommands that manage and retrieve information about MCRs.
//
// Example usage:
//
//	megaport mcr get [mcrUID]
var mcrCmd = &cobra.Command{
	Use:   "mcr",
	Short: "Manage MCRs in the Megaport API",
	Long: `Manage MCRs in the Megaport API.

This command groups all operations related to Megaport Cloud Routers (MCRs).
You can use the subcommands to perform actions such as retrieving details for a specific MCR.
For instance, use the "megaport mcr get [mcrUID]" command to fetch details for the MCR with the given UID.
`,
}

// getMCRCmd retrieves and displays detailed information for a single Megaport Cloud Router (MCR).
// This command requires exactly one argument: the UID of the MCR.
//
// It establishes a context with a timeout, logs into the Megaport API, and then uses the API client
// to get the MCR details. The retrieved information is printed using the configured output format (table/json).
//
// Example usage:
//
//	megaport mcr get MCR12345
var getMCRCmd = &cobra.Command{
	Use:   "get [mcrUID]",
	Short: "Get details for a single MCR",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create a context with a 30-second timeout for the API call.
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Log into the Megaport API using the provided credentials.
		client, err := Login(ctx)
		if err != nil {
			return fmt.Errorf("error logging in: %v", err)
		}

		// Retrieve the MCR UID from the command line arguments.
		mcrUID := args[0]

		// Use the API client to get the MCR details based on the provided UID.
		mcr, err := client.MCRService.GetMCR(ctx, mcrUID)
		if err != nil {
			return fmt.Errorf("error getting MCR: %v", err)
		}

		// Print the MCR details using the desired output format.
		err = printMCRs([]*megaport.MCR{mcr}, outputFormat)
		if err != nil {
			return fmt.Errorf("error printing MCRs: %v", err)
		}
		return nil
	},
}

func init() {
	mcrCmd.AddCommand(getMCRCmd)
	rootCmd.AddCommand(mcrCmd)
}

type MCROutput struct {
	output
	UID                string `json:"uid"`
	Name               string `json:"name"`
	LocationID         int    `json:"location_id"`
	ProvisioningStatus string `json:"provisioning_status"`
}

func ToMCROutput(mcr *megaport.MCR) (MCROutput, error) {
	if mcr == nil {
		return MCROutput{}, fmt.Errorf("invalid MCR: nil value")
	}

	return MCROutput{
		UID:                mcr.UID,
		Name:               mcr.Name,
		LocationID:         mcr.LocationID,
		ProvisioningStatus: mcr.ProvisioningStatus,
	}, nil
}

func printMCRs(mcrs []*megaport.MCR, format string) error {
	outputs := make([]MCROutput, 0, len(mcrs))
	for _, mcr := range mcrs {
		output, err := ToMCROutput(mcr)
		if err != nil {
			return err
		}
		outputs = append(outputs, output)
	}
	return printOutput(outputs, format)
}
