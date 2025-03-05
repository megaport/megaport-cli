package cmd

import (
	"context"
	"fmt"
	"strconv"
	"time"

	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

func GetVXC(cmd *cobra.Command, args []string) error {
	// Create a context with a 30-second timeout for the API call.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Log into the Megaport API.
	client, err := Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %v", err)
	}

	// Retrieve the VXC UID from the command line arguments.
	vxcUID := args[0]

	// Retrieve VXC details using the API client.
	vxc, err := client.VXCService.GetVXC(ctx, vxcUID)
	if err != nil {
		return fmt.Errorf("error getting VXC: %v", err)
	}

	// Print the VXC details using the desired output format.
	err = printVXCs([]*megaport.VXC{vxc}, outputFormat)
	if err != nil {
		return fmt.Errorf("error printing VXCs: %v", err)
	}
	return nil
}

func BuyVXC(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Prompt for required fields
	aEndProductUID, err := prompt("Enter A-End Product UID (required): ")
	if err != nil {
		return err
	}
	if aEndProductUID == "" {
		return fmt.Errorf("A-End Product UID is required")
	}

	bEndProductUID, err := prompt("Enter B-End Product UID (optional): ")
	if err != nil {
		return err
	}

	vxcName, err := prompt("Enter VXC name (required): ")
	if err != nil {
		return err
	}
	if vxcName == "" {
		return fmt.Errorf("VXC name is required")
	}

	rateLimitStr, err := prompt("Enter rate limit in Mbps (required): ")
	if err != nil {
		return err
	}
	rateLimit, err := strconv.Atoi(rateLimitStr)
	if err != nil {
		return fmt.Errorf("invalid rate limit")
	}

	termStr, err := prompt("Enter term (1, 12, 24, 36) (required): ")
	if err != nil {
		return err
	}
	term, err := strconv.Atoi(termStr)
	if err != nil || (term != 1 && term != 12 && term != 24 && term != 36) {
		return fmt.Errorf("invalid term, must be one of 1, 12, 24, 36")
	}

	aEndVLANStr, err := prompt("Enter A-End VLAN (optional, default 0): ")
	if err != nil {
		return err
	}
	aEndVLAN, err := strconv.Atoi(aEndVLANStr)
	if err != nil || aEndVLANStr == "" {
		aEndVLAN = 0
	}
	if (aEndVLAN < -1 || aEndVLAN > 4093) || aEndVLAN == 1 {
		return fmt.Errorf("invalid A-End VLAN, must be between -1 and 4093, but not 1")
	}

	aEndInnerVLANStr, err := prompt("Enter A-End Inner VLAN (optional, default 0): ")
	if err != nil {
		return err
	}
	aEndInnerVLAN, err := strconv.Atoi(aEndInnerVLANStr)
	if err != nil || aEndInnerVLANStr == "" {
		aEndInnerVLAN = 0
	}
	if aEndInnerVLAN < -1 {
		return fmt.Errorf("invalid A-End Inner VLAN, must be -1 or higher")
	}

	aEndVNICIndexStr, err := prompt("Enter A-End vNIC index (optional, default 0): ")
	if err != nil {
		return err
	}
	aEndVNICIndex, err := strconv.Atoi(aEndVNICIndexStr)
	if err != nil || aEndVNICIndexStr == "" {
		aEndVNICIndex = 0
	}

	bEndVLANStr, err := prompt("Enter B-End VLAN (optional, default 0): ")
	if err != nil {
		return err
	}
	bEndVLAN, err := strconv.Atoi(bEndVLANStr)
	if err != nil || bEndVLANStr == "" {
		bEndVLAN = 0
	}
	if (bEndVLAN < -1 || bEndVLAN > 4093) || bEndVLAN == 1 {
		return fmt.Errorf("invalid B-End VLAN, must be between -1 and 4093, but not 1")
	}

	bEndInnerVLANStr, err := prompt("Enter B-End Inner VLAN (optional, default 0): ")
	if err != nil {
		return err
	}
	bEndInnerVLAN, err := strconv.Atoi(bEndInnerVLANStr)
	if err != nil || bEndInnerVLANStr == "" {
		bEndInnerVLAN = 0
	}
	if bEndInnerVLAN < -1 {
		return fmt.Errorf("invalid B-End Inner VLAN, must be -1 or higher")
	}

	bEndVNICIndexStr, err := prompt("Enter B-End vNIC index (optional, default 0): ")
	if err != nil {
		return err
	}
	bEndVNICIndex, err := strconv.Atoi(bEndVNICIndexStr)
	if err != nil || bEndVNICIndexStr == "" {
		bEndVNICIndex = 0
	}

	// Prompt for optional fields
	promoCode, err := prompt("Enter promo code (optional): ")
	if err != nil {
		return err
	}

	serviceKey, err := prompt("Enter service key (optional): ")
	if err != nil {
		return err
	}

	costCentre, err := prompt("Enter cost centre (optional): ")
	if err != nil {
		return err
	}

	// Prompt for A-End partner configuration
	aEndPartnerConfig, err := promptPartnerConfig("A-End")
	if err != nil {
		return err
	}

	// Prompt for B-End partner configuration
	bEndPartnerConfig, err := promptPartnerConfig("B-End")
	if err != nil {
		return err
	}

	// Create the BuyVXCRequest
	req := &megaport.BuyVXCRequest{
		PortUID:    aEndProductUID,
		VXCName:    vxcName,
		RateLimit:  rateLimit,
		Term:       term,
		PromoCode:  promoCode,
		ServiceKey: serviceKey,
		CostCentre: costCentre,
		AEndConfiguration: megaport.VXCOrderEndpointConfiguration{
			VLAN: aEndVLAN,
			VXCOrderMVEConfig: &megaport.VXCOrderMVEConfig{
				InnerVLAN:             aEndInnerVLAN,
				NetworkInterfaceIndex: aEndVNICIndex,
			},
			PartnerConfig: aEndPartnerConfig,
		},
		BEndConfiguration: megaport.VXCOrderEndpointConfiguration{
			ProductUID: bEndProductUID,
			VLAN:       bEndVLAN,
			VXCOrderMVEConfig: &megaport.VXCOrderMVEConfig{
				InnerVLAN:             bEndInnerVLAN,
				NetworkInterfaceIndex: bEndVNICIndex,
			},
			PartnerConfig: bEndPartnerConfig,
		},
	}

	// Call the BuyVXC method
	client, err := Login(ctx)
	if err != nil {
		return err
	}
	fmt.Println("Buying VXC...")
	if buyVXCFunc == nil {
		return fmt.Errorf("internal error: buyVXCFunc is nil")
	}
	resp, err := buyVXCFunc(ctx, client, req)
	if err != nil {
		return err
	}

	fmt.Printf("VXC purchased successfully - UID: %s\n", resp.TechnicalServiceUID)
	return nil
}
