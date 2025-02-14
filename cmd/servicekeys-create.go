package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

var createServiceKeyCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new service key",
	Run: func(cmd *cobra.Command, args []string) {
		productUID, _ := cmd.Flags().GetString("product-uid")
		productID, _ := cmd.Flags().GetInt("product-id")
		singleUse, _ := cmd.Flags().GetBool("single-use")
		maxSpeed, _ := cmd.Flags().GetInt("max-speed")
		description, _ := cmd.Flags().GetString("description")
		startDate, _ := cmd.Flags().GetString("start-date")
		endDate, _ := cmd.Flags().GetString("end-date")

		var validFor *megaport.ValidFor
		if startDate != "" && endDate != "" {
			startTime, err := time.Parse("2006-01-02", startDate)
			if err != nil {
				log.Fatalf("Error parsing start date: %v", err)
			}
			endTime, err := time.Parse("2006-01-02", endDate)
			if err != nil {
				log.Fatalf("Error parsing end date: %v", err)
			}
			validFor = &megaport.ValidFor{
				StartTime: &megaport.Time{Time: startTime},
				EndTime:   &megaport.Time{Time: endTime},
			}
		}

		client, err := Login()
		if err != nil {
			fmt.Println("Error logging in:", err)
			os.Exit(1)
		}

		ctx := context.Background()

		req := &megaport.CreateServiceKeyRequest{
			ProductUID:  productUID,
			ProductID:   productID,
			SingleUse:   singleUse,
			MaxSpeed:    maxSpeed,
			Description: description,
			ValidFor:    validFor,
		}

		resp, err := client.ServiceKeyService.CreateServiceKey(ctx, req)
		if err != nil {
			log.Fatalf("Error creating service key: %v", err)
		}

		fmt.Printf("Service key created: %s\n", resp.ServiceKeyUID)
	},
}

func init() {
	// Create command flags
	createServiceKeyCmd.Flags().String("product-uid", "", "Product UID for the service key")
	createServiceKeyCmd.Flags().Int("product-id", 0, "Product ID for the service key")
	createServiceKeyCmd.Flags().Bool("single-use", false, "Single-use service key")
	createServiceKeyCmd.Flags().Int("max-speed", 0, "Maximum speed for the service key")
	createServiceKeyCmd.Flags().String("description", "", "Description for the service key")
	createServiceKeyCmd.Flags().String("start-date", "", "Start date for the service key (YYYY-MM-DD)")
	createServiceKeyCmd.Flags().String("end-date", "", "End date for the service key (YYYY-MM-DD)")

	servicekeysCmd.AddCommand(createServiceKeyCmd)
}
