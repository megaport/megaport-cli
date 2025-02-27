package cmd

import (
	"context"
	"fmt"
	"time"

	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

var (
	metroFilter    string
	countryFilter  string
	nameFilter     string
	idFilter       int
	siteCodeFilter string
)

// locationsCmd represents the locations command.
// It provides a list of all available locations where services can be provisioned,
// along with detailed information including name, ID, country, metropolitan area, site code, and availability status.
var locationsCmd = &cobra.Command{
	Use:   "locations",
	Short: "Manage locations in the Megaport API",
	Long: `Manage locations in the Megaport API.

This command groups operations related to locations. You can use the subcommands 
to list all locations, get details for a specific location, and filter locations 
based on various criteria such as metro area, country, and name.

Examples:
  megaport locations list
  megaport locations get --id 123
  megaport locations get --site-code "EQX-ASH"
  megaport locations get --name "Equinix Ashburn"
`,
}

// listLocationsCmd retrieves and displays all available locations from the Megaport API.
// Optionally, you can filter locations by metro area, country, or name using flags.
//
// Example usage with filtering:
//
//	megaport locations list --metro "Ashburn" --country "USA" --name "Equinix"
var listLocationsCmd = &cobra.Command{
	Use:   "list",
	Short: "List all locations with optional filters",
	Long: `List all locations available in the Megaport API.

This command fetches and displays a list of locations with details such as
location ID, name, country, metropolitan area, site code, and availability status. 
You can optionally filter the results by passing additional flags such as --metro, --country, and --name.

Example:
  megaport locations list --metro "Ashburn" --country "USA" --name "Equinix"

If no filtering options are provided, all locations will be listed.
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

		// Retrieve the list of locations from the API.
		locations, err := client.LocationService.ListLocations(ctx)
		if err != nil {
			return fmt.Errorf("error listing locations: %v", err)
		}

		// Apply filters if provided.
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

		// Filter locations based on the provided flags.
		filteredLocations := filterLocations(locations, filters)
		err = printLocations(filteredLocations, outputFormat)
		if err != nil {
			return fmt.Errorf("error printing locations: %v", err)
		}
		return nil
	},
}

// getLocationCmd retrieves and displays details for a single location from the Megaport API.
// You can specify the location by ID, site code, or exact name.
//
// Example usage:
//
//	megaport locations get --id 123
//	megaport locations get --site-code "EQX-ASH"
//	megaport locations get --name "Equinix Ashburn"
var getLocationCmd = &cobra.Command{
	Use:   "get",
	Short: "Get details for a single location",
	Long: `Get details for a single location from the Megaport API.

This command fetches and displays detailed information about a specific location.
You can specify the location by ID, site code, or exact name.

Example:
  megaport locations get --id 67
  megaport locations get --site-code "ash-eq2"
  megaport locations get --name "Equinix DC4"
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

		// Retrieve the list of locations from the API.
		locations, err := client.LocationService.ListLocations(ctx)
		if err != nil {
			return fmt.Errorf("error listing locations: %v", err)
		}

		// Filter locations based on the provided flags.
		var filteredLocations []*megaport.Location
		if cmd.Flags().Changed("id") {
			id, _ := cmd.Flags().GetInt("id")
			for _, loc := range locations {
				if loc.ID == id {
					filteredLocations = append(filteredLocations, loc)
					break
				}
			}
		} else if cmd.Flags().Changed("site-code") {
			siteCode, _ := cmd.Flags().GetString("site-code")
			for _, loc := range locations {
				if loc.SiteCode == siteCode {
					filteredLocations = append(filteredLocations, loc)
					break
				}
			}
		} else if cmd.Flags().Changed("name") {
			name, _ := cmd.Flags().GetString("name")
			for _, loc := range locations {
				if loc.Name == name {
					filteredLocations = append(filteredLocations, loc)
					break
				}
			}
		} else {
			return fmt.Errorf("please specify one of the following flags: --id, --site-code, --name")
		}

		// Print the filtered location.
		err = printLocations(filteredLocations, outputFormat)
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	// Add flags to listLocationsCmd
	listLocationsCmd.Flags().StringVar(&metroFilter, "metro", "", "Filter locations by metro area")
	listLocationsCmd.Flags().StringVar(&countryFilter, "country", "", "Filter locations by country")
	listLocationsCmd.Flags().StringVar(&nameFilter, "name", "", "Filter locations by name")

	// Add flags to getLocationCmd
	getLocationCmd.Flags().IntVar(&idFilter, "id", 0, "Get location by ID")
	getLocationCmd.Flags().StringVar(&siteCodeFilter, "site-code", "", "Get location by site code")
	getLocationCmd.Flags().StringVar(&nameFilter, "name", "", "Get location by exact name")

	// Add commands to locationsCmd
	locationsCmd.AddCommand(listLocationsCmd)
	locationsCmd.AddCommand(getLocationCmd)

	// Add locationsCmd to rootCmd
	rootCmd.AddCommand(locationsCmd)
}

// filterLocations filters the provided locations based on the given filters.
func filterLocations(locations []*megaport.Location, filters map[string]string) []*megaport.Location {
	var filtered []*megaport.Location
	for _, loc := range locations {
		if metro, ok := filters["metro"]; ok && loc.Metro != metro {
			continue
		}
		if country, ok := filters["country"]; ok && loc.Country != country {
			continue
		}
		if name, ok := filters["name"]; ok && loc.Name != name {
			continue
		}
		filtered = append(filtered, loc)
	}
	return filtered
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
func ToLocationOutput(l *megaport.Location) LocationOutput {
	return LocationOutput{
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

// printLocations prints the locations in the specified output format.
func printLocations(locations []*megaport.Location, format string) error {
	outputs := make([]LocationOutput, 0, len(locations))
	for _, loc := range locations {
		outputs = append(outputs, ToLocationOutput(loc))
	}
	return printOutput(outputs, format)
}
