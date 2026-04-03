package product

import (
	"context"
	"fmt"
	"time"

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

func ListProducts(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	output.SetOutputFormat(outputFormat)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("error logging in: %v", err)
	}

	spinner := output.PrintResourceListing("Product", noColor)

	products, err := listProductsFunc(ctx, client)

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to list products: %v", noColor, err)
		return fmt.Errorf("error listing products: %v", err)
	}

	includeInactive, _ := cmd.Flags().GetBool("include-inactive")
	filtered := filterProducts(products, includeInactive)

	limit, _ := cmd.Flags().GetInt("limit")
	if limit < 0 {
		return fmt.Errorf("--limit must be a non-negative integer")
	}
	if limit > 0 && len(filtered) > limit {
		filtered = filtered[:limit]
	}

	if len(filtered) == 0 {
		if outputFormat == utils.FormatTable {
			output.PrintInfo("No products found.", noColor)
		}
		return nil
	}

	err = printProducts(filtered, outputFormat, noColor)
	if err != nil {
		output.PrintError("Failed to print products: %v", noColor, err)
		return fmt.Errorf("error printing products: %v", err)
	}
	return nil
}

func GetProductType(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	output.SetOutputFormat(outputFormat)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("error logging in: %v", err)
	}

	productUID := args[0]

	spinner := output.PrintResourceGetting("Product Type", productUID, noColor)

	productType, err := getProductTypeFunc(ctx, client, productUID)

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to get product type: %v", noColor, err)
		return fmt.Errorf("error getting product type: %v", err)
	}

	outputs := []ProductTypeOutput{
		{
			UID:  productUID,
			Type: productType,
		},
	}

	err = output.PrintOutput(outputs, outputFormat, noColor)
	if err != nil {
		output.PrintError("Failed to print product type: %v", noColor, err)
		return fmt.Errorf("error printing product type: %v", err)
	}
	return nil
}
