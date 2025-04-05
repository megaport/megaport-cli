package servicekeys

import (
	"github.com/megaport/megaport-cli/internal/base/help"
	"github.com/megaport/megaport-cli/internal/utils"
	"github.com/spf13/cobra"
)

// servicekeysCmd is the parent command for managing service keys in the Megaport API.
var servicekeysCmd = &cobra.Command{
	Use:   "servicekeys",
	Short: "Manage service keys for the Megaport API",
}

// createServiceKeyCmd creates a new service key.
var createServiceKeyCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new service key",
	RunE:  utils.WrapColorAwareRunE(CreateServiceKey),
}

// updateServiceKeyCmd updates an existing service key.
var updateServiceKeyCmd = &cobra.Command{
	Use:   "update [key]",
	Short: "Update an existing service key",
	Args:  cobra.ExactArgs(1),
	RunE:  utils.WrapColorAwareRunE(UpdateServiceKey),
}

// listServiceKeysCmd lists all service keys for the Megaport API.
var listServiceKeysCmd = &cobra.Command{
	Use:   "list",
	Short: "List all service keys",
	RunE:  utils.WrapOutputFormatRunE(ListServiceKeys),
}

// getServiceKeyCmd retrieves details of a specific service key.
var getServiceKeyCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get details of a service key",
	Args:  cobra.ExactArgs(1),
	RunE:  utils.WrapOutputFormatRunE(GetServiceKey),
}

func AddCommandsTo(rootCmd *cobra.Command) {
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

	// Set up help builders for commands

	// servicekeys command help
	servicekeysHelp := &help.CommandHelpBuilder{
		CommandName: "megaport-cli servicekeys",
		ShortDesc:   "Manage service keys for the Megaport API",
		LongDesc:    "Manage service keys for the Megaport API.\n\nThis command groups all operations related to service keys. You can use its subcommands to create, update, list, and get details of service keys.",
		Examples: []string{
			"servicekeys list",
			"servicekeys get [key]",
			"servicekeys create --product-uid \"product-uid\" --description \"My service key\"",
			"servicekeys update [key] --description \"Updated description\"",
		},
	}
	servicekeysCmd.Long = servicekeysHelp.Build(rootCmd)

	// create servicekey help
	createServiceKeyHelp := &help.CommandHelpBuilder{
		CommandName: "megaport-cli servicekeys create",
		ShortDesc:   "Create a new service key",
		LongDesc:    "Create a new service key for interacting with the Megaport API.\n\nThis command generates a new service key and displays its details.",
		RequiredFlags: map[string]string{
			"product-uid": "Product UID for the service key",
		},
		OptionalFlags: map[string]string{
			"product-id":  "Product ID for the service key",
			"single-use":  "Single-use service key",
			"max-speed":   "Maximum speed for the service key",
			"description": "Description for the service key",
			"start-date":  "Start date for the service key (YYYY-MM-DD)",
			"end-date":    "End date for the service key (YYYY-MM-DD)",
		},
		Examples: []string{
			"create --product-uid \"product-uid\" --description \"My service key\"",
			"create --product-uid \"product-uid\" --single-use --max-speed 1000 --description \"Single-use key\"",
			"create --product-uid \"product-uid\" --start-date \"2023-01-01\" --end-date \"2023-12-31\"",
		},
	}
	createServiceKeyCmd.Long = createServiceKeyHelp.Build(rootCmd)

	// update servicekey help
	updateServiceKeyHelp := &help.CommandHelpBuilder{
		CommandName: "megaport-cli servicekeys update",
		ShortDesc:   "Update an existing service key",
		LongDesc:    "Update an existing service key for the Megaport API.\n\nThis command allows you to modify the details of an existing service key. You need to specify the key identifier as an argument, and provide any updated values as flags.",
		OptionalFlags: map[string]string{
			"product-uid": "Product UID for the service key",
			"product-id":  "Product ID for the service key",
			"single-use":  "Single-use service key",
			"active":      "Activate the service key",
			"description": "Description for the service key",
		},
		Examples: []string{
			"update a1b2c3d4-e5f6-7890-1234-567890abcdef --description \"Updated description\"",
			"update a1b2c3d4-e5f6-7890-1234-567890abcdef --active",
			"update a1b2c3d4-e5f6-7890-1234-567890abcdef --product-uid \"new-product-uid\"",
		},
	}
	updateServiceKeyCmd.Long = updateServiceKeyHelp.Build(rootCmd)

	// list servicekeys help
	listServiceKeysHelp := &help.CommandHelpBuilder{
		CommandName: "megaport-cli servicekeys list",
		ShortDesc:   "List all service keys",
		LongDesc:    "List all service keys for the Megaport API.\n\nThis command retrieves and displays all service keys along with their details. Use this command to review the keys available in your account.",
		Examples: []string{
			"list",
		},
	}
	listServiceKeysCmd.Long = listServiceKeysHelp.Build(rootCmd)

	// get servicekey help
	getServiceKeyHelp := &help.CommandHelpBuilder{
		CommandName: "megaport-cli servicekeys get",
		ShortDesc:   "Get details of a service key",
		LongDesc:    "Get details of a specific service key.\n\nThis command fetches and displays detailed information about a given service key. You must provide the service key identifier as an argument.",
		Examples: []string{
			"get a1b2c3d4-e5f6-7890-1234-567890abcdef",
		},
	}
	getServiceKeyCmd.Long = getServiceKeyHelp.Build(rootCmd)

	rootCmd.AddCommand(servicekeysCmd)
}
