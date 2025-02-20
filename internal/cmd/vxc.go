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

var (
	vxcName string
)

var vxcsCmd = &cobra.Command{
	Use:   "vxcs",
	Short: "Manage VXCs in the Megaport API",
	Long:  `Manage VXCs in the Megaport API.`,
}

var getVXCCmd = &cobra.Command{
	Use:   "get [vxcUID]",
	Short: "Get details for a single VXC",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		client, err := Login(ctx)
		if err != nil {
			return fmt.Errorf("error logging in: %v", err)
		}

		vxcUID := args[0]
		vxc, err := client.VXCService.GetVXC(ctx, vxcUID)
		if err != nil {
			return fmt.Errorf("error getting VXC: %v", err)
		}

		printVXCs([]*megaport.VXC{vxc}, outputFormat)
		return nil
	},
}

func init() {
	getVXCCmd.Flags().StringVar(&vxcName, "name", "", "Filter by VXC Name")
	vxcsCmd.AddCommand(getVXCCmd)
	rootCmd.AddCommand(vxcsCmd)
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
