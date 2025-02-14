/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"

	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listLocationsCmd = &cobra.Command{
	Use:   "list",
	Short: "List all locations with optional filters",
	Long:  `List all locations with optional filters. You can filter by metro area, country, or name.`,
	Run: func(cmd *cobra.Command, args []string) {
		filters := map[string]string{}
		if cmd.Flags().Changed("metro") {
			metro, _ := cmd.Flags().GetString("metro")
			filters["metro"] = metro
		}
		if cmd.Flags().Changed("country") {
			country, _ := cmd.Flags().GetString("country")
			filters["country"] = country
		}
		if cmd.Flags().Changed("name") {
			name, _ := cmd.Flags().GetString("name")
			filters["name"] = name
		}

		ListLocations(filters, outputFormat)
	},
}

// LocationOutput represents the desired fields for JSON output.
type LocationOutput struct {
	ID        int     `json:"id"`
	Name      string  `json:"name"`
	Country   string  `json:"country"`
	Metro     string  `json:"metro"`
	SiteCode  string  `json:"site_code"`
	Market    string  `json:"market"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Status    string  `json:"status"`
}

// ToLocationOutput converts a Location to a LocationOutput.
func ToLocationOutput(l *megaport.Location) *LocationOutput {
	return &LocationOutput{
		ID:        l.ID,
		Name:      l.Name,
		Country:   l.Country,
		Metro:     l.Metro,
		SiteCode:  l.SiteCode,
		Market:    l.Market,
		Latitude:  l.Latitude,
		Longitude: l.Longitude,
		Status:    l.Status,
	}
}

func init() {
	listLocationsCmd.PersistentFlags().String("metro", "", "Metro area to filter by")
	listLocationsCmd.PersistentFlags().String("country", "", "Country to filter by")
	listLocationsCmd.PersistentFlags().String("name", "", "Name to filter by, does not need to be exact")
	locationsCmd.AddCommand(listLocationsCmd)
}
func ListLocations(filters map[string]string, outputFormat string) {
	client, err := Login()
	if err != nil {
		fmt.Println("Error logging in:", err)
		os.Exit(1)
	}

	ctx := context.Background()

	locations, err := client.LocationService.ListLocations(ctx)
	if err != nil {
		fmt.Println("Error listing locations:", err)
		os.Exit(1)
	}

	var locationList []*megaport.Location
	filtered := []*megaport.Location{}

	if len(filters) > 0 {
		for _, location := range locations {
			if filters["metro"] != "" && location.Metro != filters["metro"] {
				continue
			}
			if filters["country"] != "" && location.Country != filters["country"] {
				continue
			}
			if filters["name"] != "" && !strings.Contains(location.Name, filters["name"]) {
				continue
			}
			filtered = append(filtered, location)
		}
		locationList = filtered
	} else {
		locationList = locations
	}

	if len(locationList) > 0 {
		switch outputFormat {
		case "json":
			var outputList []*LocationOutput
			for _, location := range locationList {
				outputList = append(outputList, ToLocationOutput(location))
			}
			printed, err := json.Marshal(outputList)
			if err != nil {
				fmt.Println("Error printing locations:", err)
				os.Exit(1)
			}
			fmt.Println(string(printed))
		case "table":
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"ID", "Name", "Metro", "Country", "Site Code", "Market", "Latitude", "Longitude", "VRouter Available", "Status"})

			for _, location := range locationList {
				table.Append([]string{
					fmt.Sprintf("%d", location.ID),
					location.Name,
					location.Metro,
					location.Country,
					location.SiteCode,
					location.Market,
					fmt.Sprintf("%f", location.Latitude),
					fmt.Sprintf("%f", location.Longitude),
					fmt.Sprintf("%t", location.VRouterAvailable),
					location.Status,
				})
			}
			table.Render()
		default:
			fmt.Println("Invalid output format. Use 'json', 'table', 'csv', 'yaml', 'xml', or 'html'.")
		}
	} else {
		fmt.Println("No locations found")
	}
}
