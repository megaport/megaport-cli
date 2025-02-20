package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
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

var portsCmd = &cobra.Command{
	Use:   "ports",
	Short: "Manage ports in the Megaport API",
	Long:  `Manage ports in the Megaport API.`,
}

var listPortsCmd = &cobra.Command{
	Use:   "list",
	Short: "List all ports",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		client, err := Login(ctx)
		if err != nil {
			return fmt.Errorf("error logging in: %v", err)
		}

		ports, err := client.PortService.ListPorts(ctx)
		if err != nil {
			return fmt.Errorf("error listing ports: %v", err)
		}

		filteredPorts := filterPorts(ports, locationID, portSpeed, portName)
		printPorts(filteredPorts, outputFormat)
		return nil
	},
}

var getPortCmd = &cobra.Command{
	Use:   "get [portUID]",
	Short: "Get details for a single port",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		client, err := Login(ctx)
		if err != nil {
			return fmt.Errorf("error logging in: %v", err)
		}

		portUID := args[0]
		port, err := client.PortService.GetPort(ctx, portUID)
		if err != nil {
			return fmt.Errorf("error getting port: %v", err)
		}

		printPorts([]*megaport.Port{port}, outputFormat)
		return nil
	},
}

func init() {
	listPortsCmd.Flags().IntVar(&locationID, "location-id", 0, "Filter by Location ID")
	listPortsCmd.Flags().IntVar(&portSpeed, "port-speed", 0, "Filter by Port Speed")
	listPortsCmd.Flags().StringVar(&portName, "name", "", "Filter by Port Name")
	listPortsCmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (json, table)")
	getPortCmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (json, table)")
	portsCmd.AddCommand(listPortsCmd)
	portsCmd.AddCommand(getPortCmd)
	rootCmd.AddCommand(portsCmd)
}

// filterPorts applies filters to a list of ports
func filterPorts(ports []*megaport.Port, locationID, portSpeed int, name string) []*megaport.Port {
	var filtered []*megaport.Port
	for _, port := range ports {
		if locationID != 0 && port.LocationID != locationID {
			continue
		}
		if portSpeed != 0 && port.PortSpeed != portSpeed {
			continue
		}
		if name != "" && !strings.Contains(strings.ToLower(port.Name), strings.ToLower(name)) {
			continue
		}
		filtered = append(filtered, port)
	}
	return filtered
}

// PortOutput represents the desired fields for JSON output.
type PortOutput struct {
	UID        string `json:"uid"`
	Name       string `json:"name"`
	LocationID int    `json:"location_id"`
	PortSpeed  int    `json:"port_speed"`
}

// ToPortOutput converts a Port to a PortOutput.
func ToPortOutput(p *megaport.Port) *PortOutput {
	return &PortOutput{
		UID:        p.UID,
		Name:       p.Name,
		LocationID: p.LocationID,
		PortSpeed:  p.PortSpeed,
	}
}

// printPorts prints the ports in the specified output format.
func printPorts(ports []*megaport.Port, format string) {
	switch format {
	case "json":
		var outputList []*PortOutput
		for _, port := range ports {
			outputList = append(outputList, ToPortOutput(port))
		}
		printed, err := json.Marshal(outputList)
		if err != nil {
			fmt.Println("Error printing ports:", err)
			os.Exit(1)
		}
		fmt.Println(string(printed))
	case "table":
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"UID", "Name", "LocationID", "PortSpeed"})

		for _, port := range ports {
			table.Append([]string{
				port.UID,
				port.Name,
				fmt.Sprintf("%d", port.LocationID),
				fmt.Sprintf("%d", port.PortSpeed),
			})
		}
		table.Render()
	default:
		fmt.Println("Invalid output format. Use 'json' or 'table'")
	}
}
