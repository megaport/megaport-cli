package product

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/megaport/megaport-cli/internal/base/exitcodes"
	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/commands/config"
	"github.com/megaport/megaport-cli/internal/utils"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

var listProductsFunc = func(ctx context.Context, client *megaport.Client) ([]megaport.Product, error) {
	return client.ProductService.ListProducts(ctx)
}

var getProductTypeFunc = func(ctx context.Context, client *megaport.Client, productUID string) (string, error) {
	return client.ProductService.GetProductType(ctx, productUID)
}

var getProductPricingFunc = func(ctx context.Context, client *megaport.Client, req megaport.PriceBookRequest) (*megaport.PriceBookDTO, error) {
	return client.ProductService.GetProductPricing(ctx, req)
}

var getProductPricingForCompanyFunc = func(ctx context.Context, client *megaport.Client, req *megaport.GetProductPricingRequest) (*megaport.PriceBookDTO, error) {
	return client.ProductService.GetProductPricingForCompany(ctx, req)
}

func ListProducts(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	output.SetOutputFormat(outputFormat)
	ctx, cancel := utils.ContextFromCmd(cmd)
	defer cancel()

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("failed to log in: %w", err)
	}

	spinner := output.PrintResourceListing("Product", noColor)

	products, err := listProductsFunc(ctx, client)

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to list products: %v", noColor, err)
		return fmt.Errorf("failed to list products: %w", err)
	}

	// Flag read errors are intentionally ignored — flags are registered by the command builder.
	includeInactive, _ := cmd.Flags().GetBool("include-inactive")
	filtered := filterProducts(products, includeInactive)

	limit, _ := cmd.Flags().GetInt("limit")
	return utils.ApplyLimitAndPrint(filtered, limit, outputFormat, noColor,
		"No products found.", printProducts)
}

func GetProductType(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	output.SetOutputFormat(outputFormat)
	ctx, cancel := utils.ContextFromCmd(cmd)
	defer cancel()

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("failed to log in: %w", err)
	}

	productUID := args[0]

	spinner := output.PrintResourceGetting("Product Type", productUID, noColor)

	productType, err := getProductTypeFunc(ctx, client, productUID)

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to get product type: %v", noColor, err)
		return fmt.Errorf("failed to get product type: %w", err)
	}

	outputs := []productTypeOutput{
		{
			UID:  productUID,
			Type: productType,
		},
	}

	err = output.PrintOutput(outputs, outputFormat, noColor)
	if err != nil {
		output.PrintError("Failed to print product type: %v", noColor, err)
		return fmt.Errorf("failed to print product type: %w", err)
	}
	return nil
}

func GetPortPricing(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	req := &megaport.MegaportPriceBookRequest{}
	return getProductPricing(cmd, noColor, outputFormat, "Port", req, func() {
		// Flag read errors are ignored: the command builder registers these flags.
		req.LocationID, _ = cmd.Flags().GetInt("location-id")
		req.Speed, _ = cmd.Flags().GetInt("speed")
		req.Term, _ = cmd.Flags().GetInt("term")
		req.Currency, _ = cmd.Flags().GetString("currency")
		req.ProductUID, _ = cmd.Flags().GetString("product-uid")
	})
}

func GetVXCPricing(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	req := &megaport.VXCPriceBookRequest{}
	return getProductPricing(cmd, noColor, outputFormat, "VXC", req, func() {
		req.ALocationID, _ = cmd.Flags().GetInt("a-location-id")
		req.BLocationID, _ = cmd.Flags().GetInt("b-location-id")
		req.Speed, _ = cmd.Flags().GetInt("speed")
		req.Term, _ = cmd.Flags().GetInt("term")
		req.AEndProductType, _ = cmd.Flags().GetString("a-end-product-type")
		req.ConnectType, _ = cmd.Flags().GetString("connect-type")
		req.Currency, _ = cmd.Flags().GetString("currency")
		req.ProductUID, _ = cmd.Flags().GetString("product-uid")
	})
}

// getProductPricing fills req from --json or --json-file when either is set,
// and from fromFlags otherwise.
func getProductPricing(cmd *cobra.Command, noColor bool, outputFormat, productName string, req megaport.PriceBookRequest, fromFlags func()) error {
	output.SetOutputFormat(outputFormat)
	ctx, cancel := utils.ContextFromCmd(cmd)
	defer cancel()

	jsonStr, _ := cmd.Flags().GetString("json")
	jsonFile, _ := cmd.Flags().GetString("json-file")
	if jsonStr != "" || jsonFile != "" {
		jsonData, err := utils.ReadJSONInput(jsonStr, jsonFile)
		if err != nil {
			output.PrintError("Failed to process JSON input: %v", noColor, err)
			return exitcodes.NewUsageError(err)
		}
		if err := json.Unmarshal(jsonData, req); err != nil {
			output.PrintError("Failed to process JSON input: %v", noColor, err)
			return exitcodes.NewUsageError(fmt.Errorf("failed to parse JSON: %w", err))
		}
	} else {
		fromFlags()
	}
	companyUID, _ := cmd.Flags().GetString("company-uid")

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("failed to log in: %w", err)
	}

	spinner := output.PrintCustomSpinner("Getting pricing for", productName, noColor)

	var price *megaport.PriceBookDTO
	if companyUID != "" {
		price, err = getProductPricingForCompanyFunc(ctx, client, &megaport.GetProductPricingRequest{Req: req, CompanyUID: companyUID})
	} else {
		price, err = getProductPricingFunc(ctx, client, req)
	}

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to get %s pricing: %v", noColor, productName, err)
		return fmt.Errorf("failed to get %s pricing: %w", productName, err)
	}

	if err := printProductPricing(price, outputFormat, noColor); err != nil {
		output.PrintError("Failed to print %s pricing: %v", noColor, productName, err)
		return fmt.Errorf("failed to print %s pricing: %w", productName, err)
	}
	return nil
}
