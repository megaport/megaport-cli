package cmd

import (
	"context"
	"fmt"
	"strconv"
	"time"

	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

var (
	locationID int
	portSpeed  int
	portName   string
)

// portsCmd is the base command for all operations related to ports in the Megaport API.
// This command serves as a container for subcommands which allow you to list and get details of ports.
//
// Example usage:
//
//	megaport ports list
//	megaport ports get [portUID]
//	megaport ports buy
var portsCmd = &cobra.Command{
	Use:   "ports",
	Short: "Manage ports in the Megaport API",
	Long: `Manage ports in the Megaport API.

This command groups operations related to ports. You can use the subcommands 
to list all ports, get details for a specific port, or buy a new port.

Examples:
  megaport ports list
  megaport ports get [portUID]
  megaport ports buy
`,
}

// Define the functions for easier testing
var updatePortFunc = func(ctx context.Context, client *megaport.Client, req *megaport.ModifyPortRequest) (*megaport.ModifyPortResponse, error) {
	return client.PortService.ModifyPort(ctx, req)
}

var deletePortFunc = func(ctx context.Context, client *megaport.Client, req *megaport.DeletePortRequest) (*megaport.DeletePortResponse, error) {
	return client.PortService.DeletePort(ctx, req)
}

var restorePortFunc = func(ctx context.Context, client *megaport.Client, portUID string) (*megaport.RestorePortResponse, error) {
	return client.PortService.RestorePort(ctx, portUID)
}

var lockPortFunc = func(ctx context.Context, client *megaport.Client, portUID string) (*megaport.LockPortResponse, error) {
	return client.PortService.LockPort(ctx, portUID)
}

var unlockPortFunc = func(ctx context.Context, client *megaport.Client, portUID string) (*megaport.UnlockPortResponse, error) {
	return client.PortService.UnlockPort(ctx, portUID)
}

var checkPortVLANAvailabilityFunc = func(ctx context.Context, client *megaport.Client, portUID string, vlan int) (bool, error) {
	return client.PortService.CheckPortVLANAvailability(ctx, portUID, vlan)
}

var buyPortFunc = func(ctx context.Context, client *megaport.Client, req *megaport.BuyPortRequest) (*megaport.BuyPortResponse, error) {
	return client.PortService.BuyPort(ctx, req)
}

// buyPortCmd allows you to purchase a port by providing the necessary details.
var buyPortCmd = &cobra.Command{
	Use:   "buy",
	Short: "Buy a port through the Megaport API",
	Long: `Buy a port through the Megaport API.

This command allows you to purchase a port by providing the necessary details.
You will be prompted to enter the required and optional fields.

Required fields:
  - name: The name of the port.
  - term: The term of the port (1, 12, 24, or 36 months).
  - port_speed: The speed of the port (1000, 10000, or 100000 Mbps).
  - location_id: The ID of the location where the port will be provisioned.
  - marketplace_visibility: Whether the port should be visible in the marketplace (true or false).

Optional fields:
  - diversity_zone: The diversity zone for the port.
  - cost_center: The cost center for the port.
  - promo_code: A promotional code for the port.

Example usage:

  megaport ports buy
`,
	RunE: func(cmd *cobra.Command, args []string) error {
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
	},
}

// listPortsCmd retrieves and displays all available ports from the Megaport API.
// Optionally, you can filter ports by location ID, port speed, or port name using flags.
//
// Example usage with filtering:
//
//	megaport ports list --location-id 1 --port-speed 10000 --port-name "PortName"
var listPortsCmd = &cobra.Command{
	Use:   "list",
	Short: "List all ports with optional filters",
	Long: `List all ports available in the Megaport API.

This command fetches and displays a list of ports with details such as
port ID, name, location, speed, and status. You can optionally filter the results 
by passing additional flags such as --location-id, --port-speed, and --port-name.

Example:
  megaport ports list --location-id 1 --port-speed 10000 --port-name "PortName"

If no filtering options are provided, all ports will be listed.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
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
	},
}

// getPortCmd retrieves and displays details for a single port from the Megaport API.
// This command requires exactly one argument: the UID of the port.
//
// Example usage:
//
//	megaport ports get [portUID]
var getPortCmd = &cobra.Command{
	Use:   "get [portUID]",
	Short: "Get details for a single port",
	Long: `Get details for a single port from the Megaport API.

This command fetches and displays detailed information about a specific port.
You need to provide the UID of the port as an argument.

Example:
  megaport ports get [portUID]
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
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
	},
}

// updatePortCmd updates a port's details in the Megaport API.
var updatePortCmd = &cobra.Command{
	Use:   "update [portUID]",
	Short: "Update a port's details",
	Long: `Update a port's details in the Megaport API.

This command allows you to update the details of an existing port by providing the necessary fields.
You need to provide the UID of the port as an argument.

Example usage:

  megaport ports update [portUID]
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
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
	},
}

// deletePortCmd deletes a port from the user's account.
var deletePortCmd = &cobra.Command{
	Use:   "delete [portUID]",
	Short: "Delete a port from your account",
	Long: `Delete a port from your account in the Megaport API.

This command allows you to delete an existing port by providing the UID of the port as an argument.
You can optionally specify whether to delete the port immediately or at the end of the billing period.

Example usage:

  megaport ports delete [portUID]
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
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
	},
}

// restorePortCmd restores a previously deleted port.
var restorePortCmd = &cobra.Command{
	Use:   "restore [portUID]",
	Short: "Restore a deleted port",
	Long: `Restore a previously deleted port in the Megaport API.

This command allows you to restore a previously deleted port by providing the UID of the port as an argument.

Example usage:

  megaport ports restore [portUID]
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
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
	},
}

// lockPortCmd locks a port in the Megaport API.
var lockPortCmd = &cobra.Command{
	Use:   "lock [portUID]",
	Short: "Lock a port",
	Long: `Lock a port in the Megaport API.

This command allows you to lock an existing port by providing the UID of the port as an argument.

Example usage:

  megaport ports lock [portUID]
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
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
	},
}

// unlockPortCmd unlocks a port in the Megaport API.
var unlockPortCmd = &cobra.Command{
	Use:   "unlock [portUID]",
	Short: "Unlock a port",
	Long: `Unlock a port in the Megaport API.

This command allows you to unlock an existing port by providing the UID of the port as an argument.

Example usage:

  megaport ports unlock [portUID]
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
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
	},
}

// checkPortVLANAvailabilityCmd checks if a VLAN is available on a port.
var checkPortVLANAvailabilityCmd = &cobra.Command{
	Use:   "check-vlan [portUID] [vlan]",
	Short: "Check if a VLAN is available on a port",
	Long: `Check if a VLAN is available on a port in the Megaport API.

This command allows you to check if a specific VLAN is available on an existing port by providing the UID of the port and the VLAN ID as arguments.

Example usage:

  megaport ports check-vlan [portUID] [vlan]
`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
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
	},
}

func init() {
	// Add flags to deletePortCmd
	deletePortCmd.Flags().Bool("now", false, "Delete immediately instead of at the end of the billing period")
	deletePortCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")

	// Add flags to listPortsCmd
	listPortsCmd.Flags().IntVar(&locationID, "location-id", 0, "Filter ports by location ID")
	listPortsCmd.Flags().IntVar(&portSpeed, "port-speed", 0, "Filter ports by port speed")
	listPortsCmd.Flags().StringVar(&portName, "port-name", "", "Filter ports by port name")

	// Add commands to portsCmd
	portsCmd.AddCommand(buyPortCmd)
	portsCmd.AddCommand(listPortsCmd)
	portsCmd.AddCommand(getPortCmd)
	portsCmd.AddCommand(updatePortCmd)
	portsCmd.AddCommand(deletePortCmd)
	portsCmd.AddCommand(restorePortCmd)
	portsCmd.AddCommand(lockPortCmd)
	portsCmd.AddCommand(unlockPortCmd)
	portsCmd.AddCommand(checkPortVLANAvailabilityCmd)

	// Add portsCmd to rootCmd
	rootCmd.AddCommand(portsCmd)
}

// filterPorts filters the provided ports based on the given filters.
func filterPorts(ports []*megaport.Port, locationID int, portSpeed int, portName string) []*megaport.Port {
	if ports == nil {
		return []*megaport.Port{}
	}

	filteredPorts := make([]*megaport.Port, 0)

	for _, port := range ports {
		if port == nil {
			continue
		}

		// Apply location ID filter
		if locationID != 0 && port.LocationID != locationID {
			continue
		}

		// Apply port speed filter
		if portSpeed != 0 && port.PortSpeed != portSpeed {
			continue
		}

		// Apply port name filter
		if portName != "" && port.Name != portName {
			continue
		}

		// Port passed all filters
		filteredPorts = append(filteredPorts, port)
	}

	return filteredPorts
}

// PortOutput represents the desired fields for JSON output.
type PortOutput struct {
	UID                string `json:"uid"`
	Name               string `json:"name"`
	LocationID         int    `json:"location_id"`
	PortSpeed          int    `json:"port_speed"`
	ProvisioningStatus string `json:"provisioning_status"`
}

// ToPortOutput converts a *megaport.Port to our PortOutput struct.
func ToPortOutput(port *megaport.Port) (PortOutput, error) {
	if port == nil {
		return PortOutput{}, fmt.Errorf("invalid port: nil value")
	}

	return PortOutput{
		UID:                port.UID,
		Name:               port.Name,
		LocationID:         port.LocationID,
		PortSpeed:          port.PortSpeed,
		ProvisioningStatus: port.ProvisioningStatus,
	}, nil
}

func printPorts(ports []*megaport.Port, format string) error {
	outputs := make([]PortOutput, 0, len(ports))
	for _, port := range ports {
		output, err := ToPortOutput(port)
		if err != nil {
			return err
		}
		outputs = append(outputs, output)
	}
	return printOutput(outputs, format)
}
