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

// vxcCmd is the base command for all operations related to Virtual Cross Connects (VXCs).
// It groups subcommands for managing VXCs in the Megaport API.
// Use the "megaport vxc get [vxcUID]" command to retrieve detailed information about a specific VXC.
var vxcCmd = &cobra.Command{
	Use:   "vxc",
	Short: "Manage VXCs in the Megaport API",
	Long: `Manage VXCs in the Megaport API.

This command groups all operations related to Virtual Cross Connects (VXCs).
You can use the subcommands to perform actions such as retrieving details for a specific VXC.
For example, use the "megaport vxc get [vxcUID]" command to fetch details for the VXC identified by its UID.
`,
}

// getVXCCmd retrieves detailed information for a single Virtual Cross Connect (VXC).
// This command requires exactly one argument: the UID of the VXC.
// It establishes a context with a timeout, logs into the Megaport API, and uses the API client
// to obtain and then display the VXC details using the configured output format (JSON or table).
//
// Example usage:
//
//	megaport vxc get VXC12345
var getVXCCmd = &cobra.Command{
	Use:   "get [vxcUID]",
	Short: "Get details for a single VXC",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
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
		printVXCs([]*megaport.VXC{vxc}, outputFormat)
		return nil
	},
}

func init() {
	vxcCmd.AddCommand(getVXCCmd)
	rootCmd.AddCommand(vxcCmd)
}

// VXCOutput represents the desired fields for JSON output.
type VXCOutput struct {
	UID     string `json:"uid"`
	Name    string `json:"name"`
	AEndUID string `json:"a_end_uid"`
	BEndUID string `json:"b_end_uid"`
}

// ToVXCOutput converts a VXC to a VXCOutput.
func ToVXCOutput(v *megaport.VXC) *VXCOutput {
	return &VXCOutput{
		UID:     v.UID,
		Name:    v.Name,
		AEndUID: v.AEndConfiguration.UID,
		BEndUID: v.BEndConfiguration.UID,
	}
}

// printVXCs prints the VXCs in the specified output format.
func printVXCs(vxcs []*megaport.VXC, format string) {
	switch format {
	case "json":
		var outputList []*VXCOutput
		for _, vxc := range vxcs {
			outputList = append(outputList, ToVXCOutput(vxc))
		}
		printed, err := json.Marshal(outputList)
		if err != nil {
			fmt.Println("Error printing VXCs:", err)
			os.Exit(1)
		}
		fmt.Println(string(printed))
	case "table":
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"UID", "Name", "AEndUID", "BEndUID"})

		for _, vxc := range vxcs {
			table.Append([]string{
				vxc.UID,
				vxc.Name,
				vxc.AEndConfiguration.UID,
				vxc.BEndConfiguration.UID,
			})
		}
		table.Render()
	default:
		fmt.Println("Invalid output format. Use 'json' or 'table'")
	}
}
