package cmd

import (
	"github.com/spf13/cobra"
)

// locationsCmd is the base command for all operations related to locations in the Megaport API.
var locationsCmd = &cobra.Command{
	Use:   "locations",
	Short: "Manage locations in the Megaport API",
	Long: `Manage locations in the Megaport API.

This command groups all operations related to locations. You can use its subcommands 
to list and get details for specific locations.

Examples:
  megaport-cli locations list
  megaport-cli locations get [locationID]
`,
}

// listLocationsCmd lists all available locations.
var listLocationsCmd = &cobra.Command{
	Use:   "list",
	Short: "List all locations",
	Long: `List all locations available in the Megaport API.

This command fetches and displays a list of all available locations with details such as
location ID, name, country, and metro. You can also filter the locations based on specific criteria.

Available filters:
  - metro: Filter locations by metro area.
  - country: Filter locations by country.
  - name: Filter locations by name.

Example usage:

  megaport-cli locations list
  megaport-cli locations list --metro "San Francisco"
  megaport-cli locations list --country "US"
  megaport-cli locations list --name "Equinix SY1"

Example output:
  ID   Name        Metro       Country
  ---  ----------  ----------  -------
  1    Sydney 1    Sydney      Australia
  2    Melbourne 1 Melbourne   Australia
`,
	RunE: WrapRunE(ListLocations),
}

// getLocationCmd retrieves details for a single location.
var getLocationCmd = &cobra.Command{
	Use:   "get [locationID]",
	Short: "Get details for a single location",
	Long: `Get details for a single location from the Megaport API.

This command fetches and displays detailed information about a specific location.
You need to provide the ID of the location as an argument.

Example usage:

  megaport-cli locations get 1

Example output:
  ID   Name        Metro       Country
  ---  ----------  ----------  -------
  1    Sydney 1    Sydney      Australia
`,
	Args: cobra.ExactArgs(1),
	RunE: WrapRunE(GetLocation),
}

func init() {
	listLocationsCmd.Flags().String("metro", "", "Filter locations by metro area")
	listLocationsCmd.Flags().String("country", "", "Filter locations by country")
	listLocationsCmd.Flags().String("name", "", "Filter locations by name")

	locationsCmd.AddCommand(listLocationsCmd)
	locationsCmd.AddCommand(getLocationCmd)
	rootCmd.AddCommand(locationsCmd)
}
