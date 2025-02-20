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
	mveName string
)

var mvesCmd = &cobra.Command{
	Use:   "mve",
	Short: "Manage MVEs in the Megaport API",
	Long:  `Manage MVEs in the Megaport API.`,
}

var getMVECmd = &cobra.Command{
	Use:   "get [mveUID]",
	Short: "Get details for a single MVE",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		client, err := Login(ctx)
		if err != nil {
			return fmt.Errorf("error logging in: %v", err)
		}

		mveUID := args[0]
		mve, err := client.MVEService.GetMVE(ctx, mveUID)
		if err != nil {
			return fmt.Errorf("error getting MVE: %v", err)
		}

		printMVEs([]*megaport.MVE{mve}, outputFormat)
		return nil
	},
}

func init() {
	getMVECmd.Flags().StringVar(&mveName, "name", "", "Filter by MVE Name")
	mvesCmd.AddCommand(getMVECmd)
	rootCmd.AddCommand(mvesCmd)
}

// MVEOutput represents the desired fields for JSON output.
type MVEOutput struct {
	UID        string `json:"uid"`
	Name       string `json:"name"`
	LocationID int    `json:"location_id"`
}

// ToMVEOutput converts an MVE to an MVEOutput.
func ToMVEOutput(m *megaport.MVE) *MVEOutput {
	return &MVEOutput{
		UID:        m.UID,
		Name:       m.Name,
		LocationID: m.LocationID,
	}
}

// printMVEs prints the MVEs in the specified output format.
func printMVEs(mves []*megaport.MVE, format string) {
	switch format {
	case "json":
		var outputList []*MVEOutput
		for _, mve := range mves {
			outputList = append(outputList, ToMVEOutput(mve))
		}
		printed, err := json.Marshal(outputList)
		if err != nil {
			fmt.Println("Error printing MVEs:", err)
			os.Exit(1)
		}
		fmt.Println(string(printed))
	case "table":
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"UID", "Name", "LocationID"})

		for _, mve := range mves {
			table.Append([]string{
				mve.UID,
				mve.Name,
				fmt.Sprintf("%d", mve.LocationID),
			})
		}
		table.Render()
	default:
		fmt.Println("Invalid output format. Use 'json' or 'table'")
	}
}
