package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	megaport "github.com/megaport/megaportgo"
	"github.com/olekukonko/tablewriter"
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
var portsCmd = &cobra.Command{
	Use:   "ports",
	Short: "Manage ports in the Megaport API",
	Long: `Manage ports in the Megaport API.

This command groups all operations related to ports. You can use the subcommands 
to list all ports or get details for a specific port.

Examples:
  megaport ports list
  megaport ports get [portUID]
`,
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
func ToPortOutput(p *megaport.Port) *PortOutput {
	return &PortOutput{
		UID:                p.UID,
		Name:               p.Name,
		LocationID:         p.LocationID,
		PortSpeed:          p.PortSpeed,
		ProvisioningStatus: p.ProvisioningStatus,
	}
}

// printPorts prints the ports in either JSON or table format.
func printPorts(ports []*megaport.Port, format string) {
	switch format {
	case "json":
		var output []*PortOutput
		for _, p := range ports {
			output = append(output, ToPortOutput(p))
		}
		data, err := json.Marshal(output)
		if err != nil {
			fmt.Println("Error marshalling ports:", err)
			os.Exit(1)
		}
		fmt.Println(string(data))
	case "table":
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"UID", "Name", "LocationID", "PortSpeed", "ProvisioningStatus"})
		table.SetAutoFormatHeaders(false)

		for _, p := range ports {
			table.Append([]string{
				p.UID,
				p.Name,
				fmt.Sprintf("%d", p.LocationID),
				fmt.Sprintf("%d", p.PortSpeed),
				p.ProvisioningStatus,
			})
		}
		table.Render()
	default:
		fmt.Println("Invalid output format. Use 'json' or 'table'")
	}
}
