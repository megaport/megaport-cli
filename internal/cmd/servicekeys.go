package cmd

import (
	"github.com/spf13/cobra"
)

// servicekeysCmd is the parent command for managing service keys in the Megaport API.
var servicekeysCmd = &cobra.Command{
	Use:   "servicekeys",
	Short: "Manage service keys for the Megaport API",
	Long: `Manage service keys for the Megaport API.

This command groups all operations related to service keys. You can use its subcommands to:
  - Create a new service key.
  - Update an existing service key.
  - List all service keys.
  - Get details of a specific service key.

Examples:
  megaport-cli servicekeys list
  megaport-cli servicekeys get [key]
  megaport-cli servicekeys create --product-uid "product-uid" --description "My service key"
  megaport-cli servicekeys update [key] --description "Updated description"
`,
}

// createServiceKeyCmd creates a new service key.
var createServiceKeyCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new service key",
	Long: `Create a new service key for interacting with the Megaport API.

This command generates a new service key and displays its details.
You may need to provide additional flags or parameters based on your API requirements.

Example:
  megaport-cli servicekeys create --product-uid "product-uid" --description "My service key"

Example output:
  Key: a1b2c3d4-e5f6-7890-1234-567890abcdef  Product UID: product-uid  Description: My service key
`,
	RunE: WrapRunE(CreateServiceKey),
}

// updateServiceKeyCmd updates an existing service key.
var updateServiceKeyCmd = &cobra.Command{
	Use:   "update [key]",
	Short: "Update an existing service key",
	Long: `Update an existing service key for the Megaport API.

This command allows you to modify the details of an existing service key.
You need to specify the key identifier as an argument, and provide any updated values as flags.

Example:
  megaport-cli servicekeys update a1b2c3d4-e5f6-7890-1234-567890abcdef --description "Updated description"

Example output:
  Key: a1b2c3d4-e5f6-7890-1234-567890abcdef  Description: Updated description
`,
	Args: cobra.ExactArgs(1),
	RunE: WrapRunE(UpdateServiceKey),
}

// listServiceKeysCmd lists all service keys for the Megaport API.
var listServiceKeysCmd = &cobra.Command{
	Use:   "list",
	Short: "List all service keys",
	Long: `List all service keys for the Megaport API.

This command retrieves and displays all service keys along with their details.
Use this command to review the keys available in your account.

Example:
  megaport-cli servicekeys list

Example output:
  Key                                     Product UID          Description
  --------------------------------------  -------------------  --------------------
  a1b2c3d4-e5f6-7890-1234-567890abcdef    product-uid          My service key
`,
	RunE: WrapRunE(ListServiceKeys),
}

// getServiceKeyCmd retrieves details of a specific service key.
var getServiceKeyCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get details of a service key",
	Long: `Get details of a specific service key.

This command fetches and displays detailed information about a given service key.
You must provide the service key identifier as an argument.

Example:
  megaport-cli servicekeys get a1b2c3d4-e5f6-7890-1234-567890abcdef

Example output:
  Key: a1b2c3d4-e5f6-7890-1234-567890abcdef  Product UID: product-uid  Description: My service key
`,
	Args: cobra.ExactArgs(1),
	RunE: WrapRunE(GetServiceKey),
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
	updateServiceKeyCmd.Flags().String("product-uid", "", "Product UID for the service key")
	updateServiceKeyCmd.Flags().Int("product-id", 0, "Product ID for the service key")
	updateServiceKeyCmd.Flags().Bool("single-use", false, "Single-use service key")
	updateServiceKeyCmd.Flags().Bool("active", false, "Activate the service key")
	updateServiceKeyCmd.Flags().String("description", "", "Description for the service key")

	rootCmd.AddCommand(servicekeysCmd)
}
