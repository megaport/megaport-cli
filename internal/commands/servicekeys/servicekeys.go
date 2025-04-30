package servicekeys

import (
	"github.com/megaport/megaport-cli/internal/base/cmdbuilder"
	"github.com/spf13/cobra"
)

func AddCommandsTo(rootCmd *cobra.Command) {
	servicekeysCmd := cmdbuilder.NewCommand("servicekeys", "Manage service keys for the Megaport API").
		WithLongDesc("Manage service keys for the Megaport API.\n\nThis command groups all operations related to service keys. You can use its subcommands to create, update, list, and get details of service keys.").
		WithExample("megaport-cli servicekeys list").
		WithExample("megaport-cli servicekeys get [key]").
		WithExample("megaport-cli servicekeys create --product-uid \"product-uid\" --description \"My service key\"").
		WithExample("megaport-cli servicekeys update [key] --description \"Updated description\"").
		WithRootCmd(rootCmd).
		Build()

	createServiceKeyCmd := cmdbuilder.NewCommand("create", "Create a new service key").
		WithLongDesc("Create a new service key for interacting with the Megaport API.\n\nThis command generates a new service key and displays its details.").
		WithColorAwareRunFunc(CreateServiceKey).
		WithServiceKeyCreateFlags().
		WithExample("megaport-cli servicekeys create --product-uid \"product-uid\" --description \"My service key\"").
		WithExample("megaport-cli servicekeys create --product-uid \"product-uid\" --single-use --max-speed 1000 --description \"Single-use key\"").
		WithExample("megaport-cli servicekeys create --product-uid \"product-uid\" --start-date \"2023-01-01\" --end-date \"2023-12-31\"").
		WithRootCmd(rootCmd).
		Build()

	updateServiceKeyCmd := cmdbuilder.NewCommand("update", "Update an existing service key").
		WithArgs(cobra.ExactArgs(1)).
		WithLongDesc("Update an existing service key for the Megaport API.\n\nThis command allows you to modify the details of an existing service key. You need to specify the key identifier as an argument, and provide any updated values as flags.").
		WithColorAwareRunFunc(UpdateServiceKey).
		WithServiceKeyUpdateFlags().
		WithExample("megaport-cli servicekeys update a1b2c3d4-e5f6-7890-1234-567890abcdef --description \"Updated description\"").
		WithExample("megaport-cli servicekeys update a1b2c3d4-e5f6-7890-1234-567890abcdef --active").
		WithExample("megaport-cli servicekeys update a1b2c3d4-e5f6-7890-1234-567890abcdef --product-uid \"new-product-uid\"").
		WithRootCmd(rootCmd).
		Build()

	listServiceKeysCmd := cmdbuilder.NewCommand("list", "List all service keys").
		WithLongDesc("List all service keys for the Megaport API.\n\nThis command retrieves and displays all service keys along with their details. Use this command to review the keys available in your account.").
		WithOutputFormatRunFunc(ListServiceKeys).
		WithExample("megaport-cli servicekeys list").
		WithRootCmd(rootCmd).
		Build()

	getServiceKeyCmd := cmdbuilder.NewCommand("get", "Get details of a service key").
		WithArgs(cobra.ExactArgs(1)).
		WithLongDesc("Get details of a specific service key.\n\nThis command fetches and displays detailed information about a given service key. You must provide the service key identifier as an argument.").
		WithOutputFormatRunFunc(GetServiceKey).
		WithExample("megaport-cli servicekeys get a1b2c3d4-e5f6-7890-1234-567890abcdef").
		WithRootCmd(rootCmd).
		Build()

	servicekeysCmd.AddCommand(createServiceKeyCmd, updateServiceKeyCmd, listServiceKeysCmd, getServiceKeyCmd)
	rootCmd.AddCommand(servicekeysCmd)
}
