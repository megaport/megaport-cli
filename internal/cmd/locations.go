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

// locationsCmd represents the locations command
var locationsCmd = &cobra.Command{
	Use:   "locations",
	Short: "List all available locations",
	Long: `The locations command provides a list of all available locations 
where services can be provisioned. This command can be used to get 
detailed information about each location, including its name, 
region, and availability. For example:

mp1 locations`,
}

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

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		ListLocations(ctx, filters, outputFormat)
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
	rootCmd.AddCommand(locationsCmd)
}

func ListLocations(ctx context.Context, filters map[string]string, outputFormat string) {
	client, err := Login(ctx)
	if err != nil {
		fmt.Println("Error logging in:", err)
		os.Exit(1)
	}

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
			fmt.Println("Invalid output format. Use 'json', 'table', or 'csv'")
		}
	} else {
		fmt.Println("No locations found")
	}
}
