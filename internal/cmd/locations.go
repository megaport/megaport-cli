package cmd

import (
	"github.com/spf13/cobra"
)

// locationsCmd is the base command for all operations related to locations in the Megaport API.
var locationsCmd = &cobra.Command{
	Use:   "locations",
	Short: "Manage locations in the Megaport API",
}

// listLocationsCmd lists all available locations.
var listLocationsCmd = &cobra.Command{
	Use:   "list",
	Short: "List all locations",
	RunE:  WrapRunE(ListLocations),
}

// getLocationCmd retrieves details for a single location.
var getLocationCmd = &cobra.Command{
	Use:   "get [locationID]",
	Short: "Get details for a single location",
	Args:  cobra.ExactArgs(1),
	RunE:  WrapRunE(GetLocation),
}

func init() {
	// Add flags to the list command
	listLocationsCmd.Flags().String("metro", "", "Filter locations by metro area")
	listLocationsCmd.Flags().String("country", "", "Filter locations by country")
	listLocationsCmd.Flags().String("name", "", "Filter locations by name")

	// Set up help builders for commands

	// locations command help
	locationsHelp := &CommandHelpBuilder{
		CommandName: "megaport-cli locations",
		ShortDesc:   "Manage locations in the Megaport API",
		LongDesc:    "Manage locations in the Megaport API.\n\nThis command groups all operations related to locations. You can use its subcommands to list and get details for specific locations.",
		Examples: []string{
			"locations list",
			"locations get [locationID]",
		},
	}
	locationsCmd.Long = locationsHelp.Build()

	// list locations help
	listLocationsHelp := &CommandHelpBuilder{
		CommandName: "megaport-cli locations list",
		ShortDesc:   "List all locations",
		LongDesc:    "List all locations available in the Megaport API.\n\nThis command fetches and displays a list of all available locations with details such as location ID, name, country, and metro. You can also filter the locations based on specific criteria.",
		OptionalFlags: map[string]string{
			"metro":   "Filter locations by metro area",
			"country": "Filter locations by country",
			"name":    "Filter locations by name",
		},
		Examples: []string{
			"list",
			"list --metro \"San Francisco\"",
			"list --country \"US\"",
			"list --name \"Equinix SY1\"",
		},
	}
	listLocationsCmd.Long = listLocationsHelp.Build()

	// get location help
	getLocationHelp := &CommandHelpBuilder{
		CommandName: "megaport-cli locations get",
		ShortDesc:   "Get details for a single location",
		LongDesc:    "Get details for a single location from the Megaport API.\n\nThis command fetches and displays detailed information about a specific location. You need to provide the ID of the location as an argument.",
		Examples: []string{
			"get 1",
		},
	}
	getLocationCmd.Long = getLocationHelp.Build()

	// Add commands to their parents
	locationsCmd.AddCommand(listLocationsCmd)
	locationsCmd.AddCommand(getLocationCmd)
	rootCmd.AddCommand(locationsCmd)
}
