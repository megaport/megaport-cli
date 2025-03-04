package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

func BuyMVE(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Prompt for required fields
	name, err := prompt("Enter MVE name (required): ")
	if err != nil {
		return err
	}
	if name == "" {
		return fmt.Errorf("MVE name is required")
	}

	termStr, err := prompt("Enter term (1, 12, 24, 36) (required): ")
	if err != nil {
		return err
	}
	term, err := strconv.Atoi(termStr)
	if err != nil || (term != 1 && term != 12 && term != 24 && term != 36) {
		return fmt.Errorf("invalid term, must be one of 1, 12, 24, 36")
	}

	locationIDStr, err := prompt("Enter location ID (required): ")
	if err != nil {
		return err
	}
	locationID, err := strconv.Atoi(locationIDStr)
	if err != nil {
		return fmt.Errorf("invalid location ID")
	}

	vendor, err := prompt("Enter vendor (required): ")
	if err != nil {
		return err
	}
	if vendor == "" {
		return fmt.Errorf("vendor is required")
	}

	// Prompt for vendor-specific configuration
	var vendorConfig megaport.VendorConfig
	switch strings.ToLower(vendor) {
	case "6wind":
		vendorConfig, err = promptSixwindConfig()
	case "aruba":
		vendorConfig, err = promptArubaConfig()
	case "aviatrix":
		vendorConfig, err = promptAviatrixConfig()
	case "cisco":
		vendorConfig, err = promptCiscoConfig()
	case "fortinet":
		vendorConfig, err = promptFortinetConfig()
	case "paloalto":
		vendorConfig, err = promptPaloAltoConfig()
	case "prisma":
		vendorConfig, err = promptPrismaConfig()
	case "versa":
		vendorConfig, err = promptVersaConfig()
	case "vmware":
		vendorConfig, err = promptVmwareConfig()
	case "meraki":
		vendorConfig, err = promptMerakiConfig()
	default:
		return fmt.Errorf("unsupported vendor: %s", vendor)
	}
	if err != nil {
		return err
	}

	// Create the BuyMVERequest
	req := &megaport.BuyMVERequest{
		Name:         name,
		Term:         term,
		LocationID:   locationID,
		VendorConfig: vendorConfig,
	}

	// Call the BuyMVE method
	client, err := Login(ctx)
	if err != nil {
		return err
	}
	fmt.Println("Buying MVE...")
	if err := client.MVEService.ValidateMVEOrder(ctx, req); err != nil {
		return fmt.Errorf("validation failed: %v", err)
	}

	resp, err := buyMVEFunc(ctx, client, req)
	if err != nil {
		return err
	}

	fmt.Printf("MVE purchased successfully - UID: %s\n", resp.TechnicalServiceUID)
	return nil
}

func GetMVE(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %v", err)
	}

	mveUID := args[0]
	if mveUID == "" {
		return fmt.Errorf("MVE UID cannot be empty")
	}

	mve, err := client.MVEService.GetMVE(ctx, mveUID)
	if err != nil {
		return fmt.Errorf("error getting MVE: %v", err)
	}

	if mve == nil {
		return fmt.Errorf("no MVE found with UID: %s", mveUID)
	}

	err = printMVEs([]*megaport.MVE{mve}, outputFormat)
	if err != nil {
		return fmt.Errorf("error printing MVEs: %v", err)
	}
	return nil
}

func ListMVEImages(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %v", err)
	}

	images, err := client.MVEService.ListMVEImages(ctx)
	if err != nil {
		return fmt.Errorf("error listing MVE images: %v", err)
	}

	if images == nil {
		return fmt.Errorf("no MVE images found")
	}

	// Get filter values from flags
	vendor, _ := cmd.Flags().GetString("vendor")
	productCode, _ := cmd.Flags().GetString("product-code")
	id, _ := cmd.Flags().GetInt("id")
	version, _ := cmd.Flags().GetString("version")
	releaseImage, _ := cmd.Flags().GetBool("release-image")

	// Apply filters
	filteredImages := filterMVEImages(images, vendor, productCode, id, version, releaseImage)

	err = printOutput(filteredImages, outputFormat)
	if err != nil {
		return fmt.Errorf("error printing MVE images: %v", err)
	}
	return nil
}

func ListAvailableMVESizes(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %v", err)
	}

	sizes, err := client.MVEService.ListAvailableMVESizes(ctx)
	if err != nil {
		return fmt.Errorf("error listing MVE sizes: %v", err)
	}

	if sizes == nil {
		return fmt.Errorf("no MVE sizes found")
	}

	err = printOutput(sizes, outputFormat)
	if err != nil {
		return fmt.Errorf("error printing MVE sizes: %v", err)
	}
	return nil
}
