package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

var updateServiceKeyCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an existing service key",
	Run: func(cmd *cobra.Command, args []string) {
		key, _ := cmd.Flags().GetString("key")
		productUID, _ := cmd.Flags().GetString("product-uid")
		productID, _ := cmd.Flags().GetInt("product-id")
		singleUse, _ := cmd.Flags().GetBool("single-use")
		active, _ := cmd.Flags().GetBool("active")

		client, err := Login()
		if err != nil {
			fmt.Println("Error logging in:", err)
			os.Exit(1)
		}

		ctx := context.Background()

		req := &megaport.UpdateServiceKeyRequest{
			Key:        key,
			ProductUID: productUID,
			ProductID:  productID,
			SingleUse:  singleUse,
			Active:     active,
		}

		resp, err := client.ServiceKeyService.UpdateServiceKey(ctx, req)
		if err != nil {
			log.Fatalf("Error updating service key: %v", err)
		}

		fmt.Printf("Service key updated: %v\n", resp.IsUpdated)
	},
}

func init() {
	// Update command flags
	updateServiceKeyCmd.Flags().String("key", "", "Service key to update")
	updateServiceKeyCmd.Flags().String("product-uid", "", "Product UID for the service key")
	updateServiceKeyCmd.Flags().Int("product-id", 0, "Product ID for the service key")
	updateServiceKeyCmd.Flags().Bool("single-use", false, "Single-use service key")
	updateServiceKeyCmd.Flags().Bool("active", false, "Activate the service key")
	updateServiceKeyCmd.Flags().String("description", "", "Description for the service key")

	servicekeysCmd.AddCommand(updateServiceKeyCmd)
}
