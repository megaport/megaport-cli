package billing_market

import (
	"github.com/megaport/megaport-cli/internal/base/cmdbuilder"
	"github.com/spf13/cobra"
)

func AddCommandsTo(rootCmd *cobra.Command) {
	billingMarketCmd := cmdbuilder.NewCommand("billing-market", "Manage billing markets for the Megaport API").
		WithLongDesc("Manage billing markets for the Megaport API.\n\nThis command groups all operations related to billing markets. You can use its subcommands to get and set billing market configurations including currency and billing contact details.").
		WithExample("megaport-cli billing-market get").
		WithExample("megaport-cli billing-market set --currency USD --country US --first-party-id 1558 ...").
		WithRootCmd(rootCmd).
		Build()

	getBillingMarketCmd := cmdbuilder.NewCommand("get", "Get billing market configurations").
		WithLongDesc("Get billing market configurations for the current account.\n\nThis command retrieves and displays all billing markets and their contact details.").
		WithOutputFormatRunFunc(GetBillingMarkets).
		WithExample("megaport-cli billing-market get").
		WithExample("megaport-cli billing-market get -o json").
		WithRootCmd(rootCmd).
		Build()

	setBillingMarketCmd := cmdbuilder.NewCommand("set", "Set billing market configuration").
		WithLongDesc("Create or update a billing market configuration.\n\nThis command sets up billing market details including currency, billing contact information, and address.").
		WithColorAwareRunFunc(SetBillingMarket).
		WithBillingMarketSetFlags().
		WithExample("megaport-cli billing-market set --currency USD --language en --billing-contact-name \"John Doe\" --billing-contact-phone \"+1234567890\" --billing-contact-email \"john@example.com\" --address1 \"123 Main St\" --city \"New York\" --state \"NY\" --postcode \"10001\" --country US --first-party-id 1558").
		WithRootCmd(rootCmd).
		Build()

	billingMarketCmd.AddCommand(getBillingMarketCmd, setBillingMarketCmd)
	rootCmd.AddCommand(billingMarketCmd)
}
