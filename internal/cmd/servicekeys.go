package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	megaport "github.com/megaport/megaportgo"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// servicekeysCmd represents the servicekeys command
var servicekeysCmd = &cobra.Command{
	Use:   "servicekeys",
	Short: "Manage service keys in the Megaport API",
	Long:  `Manage service keys in the Megaport API. This command allows you to create, list, update, and get service keys.`,
}

var createServiceKeyCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new service key",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := CreateServiceKey(ctx, cmd); err != nil {
			log.Fatalf("Error creating service key: %v", err)
		}
	},
}

var updateServiceKeyCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an existing service key",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := UpdateServiceKey(ctx, cmd); err != nil {
			log.Fatalf("Error updating service key: %v", err)
		}
	},
}

var listServiceKeysCmd = &cobra.Command{
	Use:   "list",
	Short: "List all service keys",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := ListServiceKeys(ctx, cmd); err != nil {
			log.Fatalf("Error listing service keys: %v", err)
		}
	},
}

var getServiceKeyCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get details of a service key",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := GetServiceKey(ctx, cmd, args[0]); err != nil {
			log.Fatalf("Error getting service key: %v", err)
		}
	},
}

func init() {
	// Add commands to servicekeysCmd
	servicekeysCmd.AddCommand(createServiceKeyCmd)
	servicekeysCmd.AddCommand(updateServiceKeyCmd)
	servicekeysCmd.AddCommand(listServiceKeysCmd)
	servicekeysCmd.AddCommand(getServiceKeyCmd)

	// Create command flags
	createServiceKeyCmd.Flags().String("product-uid", "", "Product UID for the service key")
	createServiceKeyCmd.Flags().Int("product-id", 0, "Product ID for the service key")
	createServiceKeyCmd.Flags().Bool("single-use", false, "Single-use service key")
	createServiceKeyCmd.Flags().Int("max-speed", 0, "Maximum speed for the service key")
	createServiceKeyCmd.Flags().String("description", "", "Description for the service key")
	createServiceKeyCmd.Flags().String("start-date", "", "Start date for the service key (YYYY-MM-DD)")
	createServiceKeyCmd.Flags().String("end-date", "", "End date for the service key (YYYY-MM-DD)")

	// Update command flags
	updateServiceKeyCmd.Flags().String("key", "", "Service key to update")
	updateServiceKeyCmd.Flags().String("product-uid", "", "Product UID for the service key")
	updateServiceKeyCmd.Flags().Int("product-id", 0, "Product ID for the service key")
	updateServiceKeyCmd.Flags().Bool("single-use", false, "Single-use service key")
	updateServiceKeyCmd.Flags().Bool("active", false, "Activate the service key")
	updateServiceKeyCmd.Flags().String("description", "", "Description for the service key")

	rootCmd.AddCommand(servicekeysCmd)
}

func CreateServiceKey(ctx context.Context, cmd *cobra.Command) error {
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
			return fmt.Errorf("error parsing start date: %v", err)
		}
		endTime, err := time.Parse("2006-01-02", endDate)
		if err != nil {
			return fmt.Errorf("error parsing end date: %v", err)
		}
		validFor = &megaport.ValidFor{
			StartTime: &megaport.Time{Time: startTime},
			EndTime:   &megaport.Time{Time: endTime},
		}
	}

	client, err := Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %v", err)
	}

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
		return fmt.Errorf("error creating service key: %v", err)
	}

	fmt.Printf("Service key created: %s\n", resp.ServiceKeyUID)
	return nil
}

func UpdateServiceKey(ctx context.Context, cmd *cobra.Command) error {
	key, _ := cmd.Flags().GetString("key")
	productUID, _ := cmd.Flags().GetString("product-uid")
	productID, _ := cmd.Flags().GetInt("product-id")
	singleUse, _ := cmd.Flags().GetBool("single-use")
	active, _ := cmd.Flags().GetBool("active")

	client, err := Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %v", err)
	}

	req := &megaport.UpdateServiceKeyRequest{
		Key:        key,
		ProductUID: productUID,
		ProductID:  productID,
		SingleUse:  singleUse,
		Active:     active,
	}

	resp, err := client.ServiceKeyService.UpdateServiceKey(ctx, req)
	if err != nil {
		return fmt.Errorf("error updating service key: %v", err)
	}

	fmt.Printf("Service key updated: %v\n", resp.IsUpdated)
	return nil
}

func ListServiceKeys(ctx context.Context, cmd *cobra.Command) error {
	client, err := Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %v", err)
	}

	req := &megaport.ListServiceKeysRequest{}
	resp, err := client.ServiceKeyService.ListServiceKeys(ctx, req)
	if err != nil {
		return fmt.Errorf("error listing service keys: %v", err)
	}

	switch outputFormat {
	case "json":
		printed, err := json.MarshalIndent(resp.ServiceKeys, "", "  ")
		if err != nil {
			return fmt.Errorf("error printing service keys: %v", err)
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
		return fmt.Errorf("invalid output format. Use 'json', 'table', or 'csv'")
	}
	return nil
}

func GetServiceKey(ctx context.Context, cmd *cobra.Command, keyID string) error {
	client, err := Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %v", err)
	}

	resp, err := client.ServiceKeyService.GetServiceKey(ctx, keyID)
	if err != nil {
		return fmt.Errorf("error getting service key: %v", err)
	}

	fmt.Printf("Service Key: %s, Description: %s\n", resp.Key, resp.Description)
	return nil
}
