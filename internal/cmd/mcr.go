package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	megaport "github.com/megaport/megaportgo"
	"github.com/olekukonko/tablewriter"
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
		printMCRs([]*megaport.MCR{mcr}, outputFormat)
		return nil
	},
}

func init() {
	mcrCmd.AddCommand(getMCRCmd)
	rootCmd.AddCommand(mcrCmd)
}

// MCROutput represents the desired fields for JSON output.
type MCROutput struct {
	UID        string `json:"uid"`
	Name       string `json:"name"`
	LocationID int    `json:"location_id"`
}

// ToMCROutput converts an MCR to an MCROutput.
func ToMCROutput(m *megaport.MCR) *MCROutput {
	return &MCROutput{
		UID:        m.UID,
		Name:       m.Name,
		LocationID: m.LocationID,
	}
}

// printMCRs prints the MCRs in the specified output format.
func printMCRs(mcrs []*megaport.MCR, format string) {
	switch format {
	case "json":
		var outputList []*MCROutput
		for _, mcr := range mcrs {
			outputList = append(outputList, ToMCROutput(mcr))
		}
		printed, err := json.Marshal(outputList)
		if err != nil {
			fmt.Println("Error printing MCRs:", err)
			os.Exit(1)
		}
		fmt.Println(string(printed))
	case "table":
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"UID", "Name", "LocationID"})

		for _, mcr := range mcrs {
			table.Append([]string{
				mcr.UID,
				mcr.Name,
				fmt.Sprintf("%d", mcr.LocationID),
			})
		}
		table.Render()
	default:
		fmt.Println("Invalid output format. Use 'json' or 'table'")
	}
}
