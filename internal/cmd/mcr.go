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
	mcrName string
)

var mcrsCmd = &cobra.Command{
	Use:   "mcr",
	Short: "Manage MCRs in the Megaport API",
	Long:  `Manage MCRs in the Megaport API.`,
}

var getMCRCmd = &cobra.Command{
	Use:   "get [mcrUID]",
	Short: "Get details for a single MCR",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		client, err := Login(ctx)
		if err != nil {
			return fmt.Errorf("error logging in: %v", err)
		}

		mcrUID := args[0]
		mcr, err := client.MCRService.GetMCR(ctx, mcrUID)
		if err != nil {
			return fmt.Errorf("error getting MCR: %v", err)
		}

		printMCRs([]*megaport.MCR{mcr}, outputFormat)
		return nil
	},
}

func init() {
	getMCRCmd.Flags().StringVar(&mcrName, "name", "", "Filter by MCR Name")
	mcrsCmd.AddCommand(getMCRCmd)
	rootCmd.AddCommand(mcrsCmd)
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
