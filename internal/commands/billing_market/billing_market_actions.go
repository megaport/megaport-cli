package billing_market

import (
	"fmt"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/commands/config"
	"github.com/megaport/megaport-cli/internal/utils"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

func GetBillingMarkets(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	output.SetOutputFormat(outputFormat)
	ctx, cancel := utils.ContextFromCmd(cmd)
	defer cancel()

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("error logging in: %w", err)
	}

	spinner := output.PrintResourceListing("billing market", noColor)

	markets, err := client.BillingMarketService.GetBillingMarkets(ctx)

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to get billing markets: %v", noColor, err)
		return fmt.Errorf("error getting billing markets: %w", err)
	}

	if len(markets) == 0 {
		output.PrintWarning("No billing markets found", noColor)
	}

	outputs := make([]billingMarketOutput, 0, len(markets))
	for _, m := range markets {
		outputs = append(outputs, toBillingMarketOutput(m))
	}

	return output.PrintOutput(outputs, outputFormat, noColor)
}

func SetBillingMarket(cmd *cobra.Command, args []string, noColor bool) error {
	ctx, cancel := utils.ContextFromCmd(cmd)
	defer cancel()

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("error logging in: %w", err)
	}

	req, err := buildSetBillingMarketRequest(cmd)
	if err != nil {
		output.PrintError("Invalid input: %v", noColor, err)
		return fmt.Errorf("error building request: %w", err)
	}

	spinner := output.PrintCustomSpinner("Setting billing market", fmt.Sprintf("%s/%s", req.Country, req.CurrencyEnum), noColor)

	resp, err := client.BillingMarketService.SetBillingMarket(ctx, req)

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to set billing market: %v", noColor, err)
		return fmt.Errorf("error setting billing market: %w", err)
	}

	output.PrintSuccess("Billing market set successfully (Supply ID: %d)", noColor, resp.SupplyID)
	return nil
}

func buildSetBillingMarketRequest(cmd *cobra.Command) (*megaport.SetBillingMarketRequest, error) {
	currency, _ := cmd.Flags().GetString("currency")
	language, _ := cmd.Flags().GetString("language")
	contactName, _ := cmd.Flags().GetString("billing-contact-name")
	contactPhone, _ := cmd.Flags().GetString("billing-contact-phone")
	contactEmail, _ := cmd.Flags().GetString("billing-contact-email")
	address1, _ := cmd.Flags().GetString("address1")
	city, _ := cmd.Flags().GetString("city")
	state, _ := cmd.Flags().GetString("state")
	postcode, _ := cmd.Flags().GetString("postcode")
	country, _ := cmd.Flags().GetString("country")
	firstPartyID, _ := cmd.Flags().GetInt("first-party-id")

	req := &megaport.SetBillingMarketRequest{
		CurrencyEnum:        currency,
		Language:            language,
		BillingContactName:  contactName,
		BillingContactPhone: contactPhone,
		BillingContactEmail: contactEmail,
		Address1:            address1,
		City:                city,
		State:               state,
		Postcode:            postcode,
		Country:             country,
		FirstPartyID:        firstPartyID,
	}

	if firstPartyID <= 0 {
		return nil, fmt.Errorf("first-party-id must be a positive integer")
	}

	// Handle optional fields
	if cmd.Flags().Changed("address2") {
		address2, _ := cmd.Flags().GetString("address2")
		req.Address2 = &address2
	}
	if cmd.Flags().Changed("po-number") {
		poNumber, _ := cmd.Flags().GetString("po-number")
		req.YourPONumber = poNumber
	}
	if cmd.Flags().Changed("tax-number") {
		taxNumber, _ := cmd.Flags().GetString("tax-number")
		req.TaxNumber = taxNumber
	}

	return req, nil
}
