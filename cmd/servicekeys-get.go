package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var getServiceKeyCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get details of a service key",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		keyID := args[0]

		client, err := Login()
		if err != nil {
			fmt.Println("Error logging in:", err)
			os.Exit(1)
		}

		ctx := context.Background()

		resp, err := client.ServiceKeyService.GetServiceKey(ctx, keyID)
		if err != nil {
			log.Fatalf("Error getting service key: %v", err)
		}

		fmt.Printf("Service Key: %s, Description: %s\n", resp.Key, resp.Description)
	},
}

func init() {
	servicekeysCmd.AddCommand(getServiceKeyCmd)
}
