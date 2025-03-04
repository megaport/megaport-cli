package cmd

import (
	"context"
	"fmt"
	"strconv"
	"time"

	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

func BuyPort(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Prompt for required fields
	name, err := prompt("Enter port name (required): ")
	if err != nil {
		return err
	}
	if name == "" {
		return fmt.Errorf("port name is required")
	}

	termStr, err := prompt("Enter term (1, 12, 24, 36) (required): ")
	if err != nil {
		return err
	}
	term, err := strconv.Atoi(termStr)
	if err != nil || (term != 1 && term != 12 && term != 24 && term != 36) {
		return fmt.Errorf("invalid term, must be one of 1, 12, 24, 36")
	}

	portSpeedStr, err := prompt("Enter port speed (1000, 10000, 100000) (required): ")
	if err != nil {
		return err
	}
	portSpeed, err := strconv.Atoi(portSpeedStr)
	if err != nil || (portSpeed != 1000 && portSpeed != 10000 && portSpeed != 100000) {
		return fmt.Errorf("invalid port speed, must be one of 1000, 10000, 100000")
	}

	locationIDStr, err := prompt("Enter location ID (required): ")
	if err != nil {
		return err
	}
	locationID, err := strconv.Atoi(locationIDStr)
	if err != nil {
		return fmt.Errorf("invalid location ID")
	}

	marketplaceVisibilityStr, err := prompt("Enter marketplace visibility (true/false) (required): ")
	if err != nil {
		return err
	}
	marketplaceVisibility, err := strconv.ParseBool(marketplaceVisibilityStr)
	if err != nil {
		return fmt.Errorf("invalid marketplace visibility, must be true or false")
	}

	// Prompt for optional fields
	diversityZone, err := prompt("Enter diversity zone (optional): ")
	if err != nil {
		return err
	}

	costCentre, err := prompt("Enter cost center (optional): ")
	if err != nil {
		return err
	}

	promoCode, err := prompt("Enter promo code (optional): ")
	if err != nil {
		return err
	}

	// Create the BuyPortRequest
	req := &megaport.BuyPortRequest{
		Name:                  name,
		Term:                  term,
		PortSpeed:             portSpeed,
		LocationId:            locationID,
		MarketPlaceVisibility: marketplaceVisibility,
		DiversityZone:         diversityZone,
		CostCentre:            costCentre,
		PromoCode:             promoCode,
		WaitForProvision:      true,
		WaitForTime:           10 * time.Minute,
	}

	// Call the BuyPort method
	client, err := Login(ctx)
	if err != nil {
		return err
	}
	fmt.Println("Buying port...")
	resp, err := buyPortFunc(ctx, client, req)
	if err != nil {
		return err
	}

	fmt.Printf("Port purchased successfully - UID: %s\n", resp.TechnicalServiceUIDs[0])
	return nil
}

func BuyLAGPort(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Prompt for required fields
	name, err := prompt("Enter port name (required): ")
	if err != nil {
		return err
	}
	if name == "" {
		return fmt.Errorf("port name is required")
	}

	termStr, err := prompt("Enter term (1, 12, 24, 36) (required): ")
	if err != nil {
		return err
	}
	term, err := strconv.Atoi(termStr)
	if err != nil || (term != 1 && term != 12 && term != 24 && term != 36) {
		return fmt.Errorf("invalid term, must be one of 1, 12, 24, 36")
	}

	portSpeedStr, err := prompt("Enter port speed (10000 or 100000) (required): ")
	if err != nil {
		return err
	}
	portSpeed, err := strconv.Atoi(portSpeedStr)
	if err != nil || (portSpeed != 10000 && portSpeed != 100000) {
		return fmt.Errorf("invalid port speed, must be one of 10000 or 100000")
	}

	locationIDStr, err := prompt("Enter location ID (required): ")
	if err != nil {
		return err
	}
	locationID, err := strconv.Atoi(locationIDStr)
	if err != nil {
		return fmt.Errorf("invalid location ID")
	}

	lagCountStr, err := prompt("Enter LAG count (1-8) (required): ")
	if err != nil {
		return err
	}
	lagCount, err := strconv.Atoi(lagCountStr)
	if err != nil || lagCount < 1 || lagCount > 8 {
		return fmt.Errorf("invalid LAG count, must be between 1 and 8")
	}

	marketplaceVisibilityStr, err := prompt("Enter marketplace visibility (true/false) (required): ")
	if err != nil {
		return err
	}
	marketplaceVisibility, err := strconv.ParseBool(marketplaceVisibilityStr)
	if err != nil {
		return fmt.Errorf("invalid marketplace visibility, must be true or false")
	}

	// Prompt for optional fields
	diversityZone, err := prompt("Enter diversity zone (optional): ")
	if err != nil {
		return err
	}

	costCentre, err := prompt("Enter cost center (optional): ")
	if err != nil {
		return err
	}

	promoCode, err := prompt("Enter promo code (optional): ")
	if err != nil {
		return err
	}

	// Create the BuyPortRequest
	req := &megaport.BuyPortRequest{
		Name:                  name,
		Term:                  term,
		PortSpeed:             portSpeed,
		LocationId:            locationID,
		LagCount:              lagCount,
		MarketPlaceVisibility: marketplaceVisibility,
		DiversityZone:         diversityZone,
		CostCentre:            costCentre,
		PromoCode:             promoCode,
		WaitForProvision:      true,
		WaitForTime:           10 * time.Minute,
	}

	// Call the BuyPort method
	client, err := Login(ctx)
	if err != nil {
		return err
	}
	fmt.Println("Buying LAG port...")
	resp, err := buyPortFunc(ctx, client, req)
	if err != nil {
		return err
	}

	fmt.Printf("LAG port purchased successfully - UID: %s\n", resp.TechnicalServiceUIDs[0])
	return nil
}

func ListPorts(cmd *cobra.Command, args []string) error {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Log into Megaport API
	client, err := Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %v", err)
	}

	// Get all ports
	ports, err := client.PortService.ListPorts(ctx)
	if err != nil {
		return fmt.Errorf("error listing ports: %v", err)
	}

	// Get filter values from flags
	locationID, _ := cmd.Flags().GetInt("location-id")
	portSpeed, _ := cmd.Flags().GetInt("port-speed")
	portName, _ := cmd.Flags().GetString("port-name")

	// Apply filters
	filteredPorts := filterPorts(ports, locationID, portSpeed, portName)

	// Print ports with current output format
	return printPorts(filteredPorts, outputFormat)
}

func GetPort(cmd *cobra.Command, args []string) error {
	// Create a context with a 30-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Log into the Megaport API.
	client, err := Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %v", err)
	}

	// Retrieve the port UID from the command line arguments.
	portUID := args[0]

	// Retrieve port details using the API client.
	port, err := client.PortService.GetPort(ctx, portUID)
	if err != nil {
		return fmt.Errorf("error getting port: %v", err)
	}

	// Print the port details using the desired output format.
	err = printPorts([]*megaport.Port{port}, outputFormat)
	if err != nil {
		return fmt.Errorf("error printing ports: %v", err)
	}
	return nil
}

func UpdatePort(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Retrieve the port UID from the command line arguments.
	portUID := args[0]

	// Prompt for required fields
	name, err := prompt("Enter new port name (required): ")
	if err != nil {
		return err
	}
	if name == "" {
		return fmt.Errorf("port name is required")
	}

	// Prompt for optional fields
	marketplaceVisibilityStr, err := prompt("Enter marketplace visibility (true/false) (optional): ")
	if err != nil {
		return err
	}
	var marketplaceVisibility *bool
	if marketplaceVisibilityStr != "" {
		visibilityValue, err := strconv.ParseBool(marketplaceVisibilityStr)
		if err != nil {
			return fmt.Errorf("invalid marketplace visibility, must be true or false")
		}
		marketplaceVisibility = &visibilityValue
	}

	costCentre, err := prompt("Enter cost center (optional): ")
	if err != nil {
		return err
	}

	termStr, err := prompt("Enter new term (1, 12, 24, 36) (optional): ")
	if err != nil {
		return err
	}
	var term *int
	if termStr != "" {
		termValue, err := strconv.Atoi(termStr)
		if err != nil || (termValue != 1 && termValue != 12 && termValue != 24 && termValue != 36) {
			return fmt.Errorf("invalid term, must be one of 1, 12, 24, 36")
		}
		term = &termValue
	}

	// Create the ModifyPortRequest
	req := &megaport.ModifyPortRequest{
		PortID:                portUID,
		Name:                  name,
		MarketplaceVisibility: marketplaceVisibility,
		CostCentre:            costCentre,
		ContractTermMonths:    term,
		WaitForUpdate:         true,
		WaitForTime:           10 * time.Minute,
	}

	// Call the ModifyPort method
	client, err := Login(ctx)
	if err != nil {
		return err
	}
	fmt.Println("Updating port...")
	resp, err := updatePortFunc(ctx, client, req)
	if err != nil {
		return err
	}

	if resp.IsUpdated {
		fmt.Printf("Port updated successfully - UID: %s\n", portUID)
	} else {
		fmt.Println("Port update request was not successful")
	}
	return nil
}

func DeletePort(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Retrieve the port UID from the command line arguments.
	portUID := args[0]

	// Get delete now flag
	deleteNow, err := cmd.Flags().GetBool("now")
	if err != nil {
		return err
	}

	// Confirm deletion unless force flag is set
	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return err
	}

	if !force {
		confirmMsg := "Are you sure you want to delete port " + portUID + "? (y/n): "
		confirmation, err := prompt(confirmMsg)
		if err != nil {
			return err
		}

		if confirmation != "y" && confirmation != "Y" {
			fmt.Println("Deletion cancelled")
			return nil
		}
	}

	// Create delete request
	deleteRequest := &megaport.DeletePortRequest{
		PortID:    portUID,
		DeleteNow: deleteNow,
	}

	// Delete the port
	client, err := Login(ctx)
	if err != nil {
		return err
	}
	resp, err := deletePortFunc(ctx, client, deleteRequest)
	if err != nil {
		return err
	}

	if resp.IsDeleting {
		fmt.Printf("Port %s deleted successfully\n", portUID)
		if deleteNow {
			fmt.Println("The port will be deleted immediately")
		} else {
			fmt.Println("The port will be deleted at the end of the current billing period")
		}
	} else {
		fmt.Println("Port deletion request was not successful")
	}
	return nil
}

func RestorePort(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Retrieve the port UID from the command line arguments.
	portUID := args[0]

	// Restore the port
	client, err := Login(ctx)
	if err != nil {
		return err
	}
	resp, err := restorePortFunc(ctx, client, portUID)
	if err != nil {
		return err
	}

	if resp.IsRestored {
		fmt.Printf("Port %s restored successfully\n", portUID)
	} else {
		fmt.Println("Port restoration request was not successful")
	}
	return nil
}

func LockPort(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Retrieve the port UID from the command line arguments.
	portUID := args[0]

	// Lock the port
	client, err := Login(ctx)
	if err != nil {
		return err
	}
	resp, err := lockPortFunc(ctx, client, portUID)
	if err != nil {
		return err
	}

	if resp.IsLocking {
		fmt.Printf("Port %s locked successfully\n", portUID)
	} else {
		fmt.Println("Port lock request was not successful")
	}
	return nil
}

func UnlockPort(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Retrieve the port UID from the command line arguments.
	portUID := args[0]

	// Unlock the port
	client, err := Login(ctx)
	if err != nil {
		return err
	}
	resp, err := unlockPortFunc(ctx, client, portUID)
	if err != nil {
		return err
	}

	if resp.IsUnlocking {
		fmt.Printf("Port %s unlocked successfully\n", portUID)
	} else {
		fmt.Println("Port unlock request was not successful")
	}
	return nil
}

func CheckPortVLANAvailability(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Retrieve the port UID and VLAN ID from the command line arguments.
	portUID := args[0]
	vlan, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid VLAN ID")
	}

	// Check VLAN availability
	client, err := Login(ctx)
	if err != nil {
		return err
	}
	available, err := checkPortVLANAvailabilityFunc(ctx, client, portUID, vlan)
	if err != nil {
		return err
	}

	if available {
		fmt.Printf("VLAN %d is available on port %s\n", vlan, portUID)
	} else {
		fmt.Printf("VLAN %d is not available on port %s\n", vlan, portUID)
	}
	return nil
}
