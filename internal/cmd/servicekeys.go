package cmd

import (
	"context"
	"fmt"
	"log"
	"time"

	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

// servicekeysCmd is the parent command for managing service keys in the Megaport API.
// It groups operations that allow you to create, update, list, and get details of service keys.
//
// Example usage:
//
//	megaport servicekeys list
var servicekeysCmd = &cobra.Command{
	Use:   "servicekeys",
	Short: "Manage service keys for the Megaport API",
	Long: `Manage service keys for the Megaport API.

This command groups all operations related to service keys. You can use its subcommands to:
  - Create a new service key.
  - Update an existing service key.
  - List all service keys.
  - Get details of a specific service key.

Example:
  megaport servicekeys list
`,
}

// createServiceKeyCmd creates a new service key.
//
// Example usage:
//
//	megaport servicekeys create --key "my-new-key" --description "My service key"
var createServiceKeyCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new service key",
	Long: `Create a new service key for interacting with the Megaport API.

This command generates a new service key and displays its details.
You may need to provide additional flags or parameters based on your API requirements.

Example:
  megaport servicekeys create --key "my-new-key" --description "My service key"
`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := CreateServiceKey(ctx, cmd); err != nil {
			log.Fatalf("Error creating service key: %v", err)
		}
	},
}

// updateServiceKeyCmd updates an existing service key.
//
// Example usage:
//
//	megaport servicekeys update my-key --description "Updated description"
var updateServiceKeyCmd = &cobra.Command{
	Use:   "update [key]",
	Short: "Update an existing service key",
	Long: `Update an existing service key for the Megaport API.

This command allows you to modify the details of an existing service key.
You need to specify the key identifier as an argument, and provide any updated values as flags.

Example:
  megaport servicekeys update my-key --description "Updated description"
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := UpdateServiceKey(ctx, cmd); err != nil {
			log.Fatalf("Error updating service key: %v", err)
		}
	},
}

// listServiceKeysCmd lists all service keys for the Megaport API.
//
// Example usage:
//
//	megaport servicekeys list
var listServiceKeysCmd = &cobra.Command{
	Use:   "list",
	Short: "List all service keys",
	Long: `List all service keys for the Megaport API.

This command retrieves and displays all service keys along with their details.
Use this command to review the keys available in your account.

Example:
  megaport servicekeys list
`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := ListServiceKeys(ctx, cmd); err != nil {
			log.Fatalf("Error listing service keys: %v", err)
		}
	},
}

// ServiceKeyOutput represents the desired fields for output
type ServiceKeyOutput struct {
	output
	KeyUID      string `json:"key_uid"`
	ProductName string `json:"product_name"`
	ProductUID  string `json:"product_uid"`
	Description string `json:"description"`
	CreateDate  string `json:"create_date"`
}

// ToServiceKeyOutput converts a ServiceKey to ServiceKeyOutput
func ToServiceKeyOutput(sk *megaport.ServiceKey) *ServiceKeyOutput {
	return &ServiceKeyOutput{
		KeyUID:      sk.Key,
		ProductName: sk.ProductName,
		Description: sk.Description,
		ProductUID:  sk.ProductUID,
		CreateDate:  sk.CreateDate.Time.Format(time.RFC3339),
	}
}

// getServiceKeyCmd retrieves details of a specific service key.
//
// Example usage:
//
//	megaport servicekeys get my-key
var getServiceKeyCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get details of a service key",
	Long: `Get details of a specific service key.

This command fetches and displays detailed information about a given service key.
You must provide the service key identifier as an argument.

Example:
  megaport servicekeys get my-key
`,
	Args: cobra.ExactArgs(1),
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

	outputs := make([]*ServiceKeyOutput, 0, len(resp.ServiceKeys))
	for _, sk := range resp.ServiceKeys {
		outputs = append(outputs, ToServiceKeyOutput(sk))
	}

	return printOutput(outputs, outputFormat)
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

	output := ToServiceKeyOutput(resp)
	return printOutput([]*ServiceKeyOutput{output}, outputFormat)
}
