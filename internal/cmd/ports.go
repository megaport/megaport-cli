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

This command groups all operations related to ports. You can use the subcommands 
to list all ports, get details for a specific port, or buy a new port.

Examples:
  megaport ports list
  megaport ports get [portUID]
  megaport ports buy
`,
}

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
		portService := megaport.NewPortService(client)
		fmt.Println("Buying port...")
		resp, err := portService.BuyPort(ctx, req)
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
		// Create a context with a 30-second timeout.
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Log into the Megaport API.
		client, err := Login(ctx)
		if err != nil {
			return fmt.Errorf("error logging in: %v", err)
		}

		// Retrieve the list of ports from the API.
		ports, err := client.PortService.ListPorts(ctx)
		if err != nil {
			return fmt.Errorf("error listing ports: %v", err)
		}

		// Apply filters if provided.
		filteredPorts := filterPorts(ports, locationID, portSpeed, portName)
		printPorts(filteredPorts, outputFormat)
		return nil
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
		printPorts([]*megaport.Port{port}, outputFormat)
		return nil
	},
}

func init() {
	listPortsCmd.Flags().IntVar(&locationID, "location-id", 0, "Filter ports by location ID")
	listPortsCmd.Flags().IntVar(&portSpeed, "port-speed", 0, "Filter ports by port speed")
	listPortsCmd.Flags().StringVar(&portName, "port-name", "", "Filter ports by port name")
	portsCmd.AddCommand(buyPortCmd)
	portsCmd.AddCommand(listPortsCmd)
	portsCmd.AddCommand(getPortCmd)
	rootCmd.AddCommand(portsCmd)
}

// filterPorts filters the provided ports based on the given filters.
func filterPorts(ports []*megaport.Port, locationID, portSpeed int, portName string) []*megaport.Port {
	var filtered []*megaport.Port
	for _, port := range ports {
		if locationID != 0 && port.LocationID != locationID {
			continue
		}
		if portSpeed != 0 && port.PortSpeed != portSpeed {
			continue
		}
		if portName != "" && port.Name != portName {
			continue
		}
		filtered = append(filtered, port)
	}
	return filtered
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
func ToPortOutput(p *megaport.Port) PortOutput {
	return PortOutput{
		UID:                p.UID,
		Name:               p.Name,
		LocationID:         p.LocationID,
		PortSpeed:          p.PortSpeed,
		ProvisioningStatus: p.ProvisioningStatus,
	}
}

func printPorts(ports []*megaport.Port, format string) error {
	outputs := make([]PortOutput, 0, len(ports))
	for _, port := range ports {
		outputs = append(outputs, ToPortOutput(port))
	}
	return printOutput(outputs, format)
}
