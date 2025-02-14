package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	megaport "github.com/megaport/megaportgo"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// listServiceKeysCmd represents the list command for service keys
var listServiceKeysCmd = &cobra.Command{
	Use:   "list",
	Short: "List all service keys",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := Login()
		if err != nil {
			fmt.Println("Error logging in:", err)
			os.Exit(1)
		}

		ctx := context.Background()

		req := &megaport.ListServiceKeysRequest{}
		resp, err := client.ServiceKeyService.ListServiceKeys(ctx, req)
		if err != nil {
			fmt.Println("Error listing service keys:", err)
			os.Exit(1)
		}

		switch outputFormat {
		case "json":
			printed, err := json.MarshalIndent(resp.ServiceKeys, "", "  ")
			if err != nil {
				fmt.Println("Error printing service keys:", err)
				os.Exit(1)
			}
			fmt.Println(string(printed))
		case "table":
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Key", "Description", "Product UID", "Product ID", "Single Use", "Max Speed", "Active"})

			for _, key := range resp.ServiceKeys {
				table.Append([]string{
					key.Key,
					key.Description,
					key.ProductUID,
					fmt.Sprintf("%d", key.ProductID),
					fmt.Sprintf("%t", key.SingleUse),
					fmt.Sprintf("%d", key.MaxSpeed),
					fmt.Sprintf("%t", key.Active),
				})
			}
			table.Render()
		default:
			fmt.Println("Invalid output format. Use 'json', 'table', or 'csv'.")
		}
	},
}

func init() {
	servicekeysCmd.AddCommand(listServiceKeysCmd)
}
