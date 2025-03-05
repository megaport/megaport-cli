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

	// Prompt for network interfaces (vnics)
	var vnics []megaport.MVENetworkInterface
	for {
		vnicDescription, err := prompt("Enter VNIC description (leave blank to finish): ")
		if err != nil {
			return err
		}
		if vnicDescription == "" {
			break
		}

		vnicVLANStr, err := prompt("Enter VNIC VLAN (required): ")
		if err != nil {
			return err
		}
		vnicVLAN, err := strconv.Atoi(vnicVLANStr)
		if err != nil {
			return fmt.Errorf("invalid VNIC VLAN")
		}

		vnics = append(vnics, megaport.MVENetworkInterface{
			Description: vnicDescription,
			VLAN:        vnicVLAN,
		})
	}

	// Create the BuyMVERequest
	req := &megaport.BuyMVERequest{
		Name:         name,
		Term:         term,
		LocationID:   locationID,
		VendorConfig: vendorConfig,
		Vnics:        vnics,
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

func UpdateMVE(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	mveUID := args[0]

	// Prompt for fields to update
	name, err := prompt("Enter new MVE name (leave blank to keep current): ")
	if err != nil {
		return err
	}

	costCentre, err := prompt("Enter new cost centre (leave blank to keep current): ")
	if err != nil {
		return err
	}

	contractTermMonthsStr, err := prompt("Enter new contract term in months (leave blank to keep current): ")
	if err != nil {
		return err
	}
	var contractTermMonths *int
	if contractTermMonthsStr != "" {
		term, err := strconv.Atoi(contractTermMonthsStr)
		if err != nil {
			return fmt.Errorf("invalid contract term, must be a number")
		}
		contractTermMonths = &term
	}

	// Create the ModifyMVERequest
	req := &megaport.ModifyMVERequest{
		MVEID:              mveUID,
		Name:               name,
		CostCentre:         costCentre,
		ContractTermMonths: contractTermMonths,
		WaitForUpdate:      true,
		WaitForTime:        10 * time.Minute,
	}

	// Call the ModifyMVE method
	client, err := Login(ctx)
	if err != nil {
		return err
	}
	fmt.Println("Updating MVE...")
	resp, err := client.MVEService.ModifyMVE(ctx, req)
	if err != nil {
		return fmt.Errorf("error updating MVE: %v", err)
	}

	if resp.MVEUpdated {
		fmt.Println("MVE updated successfully")
	} else {
		fmt.Println("MVE update failed")
	}
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

func DeleteMVE(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	mveUID := args[0]

	client, err := Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %v", err)
	}

	fmt.Printf("Deleting MVE with UID: %s...\n", mveUID)
	req := &megaport.DeleteMVERequest{
		MVEID: mveUID,
	}
	resp, err := client.MVEService.DeleteMVE(ctx, req)
	if err != nil {
		return fmt.Errorf("error deleting MVE: %v", err)
	}

	if resp.IsDeleted {
		fmt.Println("MVE deleted successfully")
	} else {
		fmt.Println("MVE delete failed")
	}
	return nil
}
