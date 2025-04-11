package locations

import (
	"github.com/megaport/megaport-cli/internal/base/cmdbuilder"
	"github.com/spf13/cobra"
)

// AddCommandsTo builds the locations commands and adds them to the root command
func AddCommandsTo(rootCmd *cobra.Command) {
	// Create locations parent command
	locationsCmd := cmdbuilder.NewCommand("locations", "Manage locations in the Megaport API").
		WithLongDesc("Manage locations in the Megaport API.\n\nThis command groups all operations related to locations. You can use its subcommands to list and get details for specific locations.").
		WithExample("locations list").
		WithExample("locations get [locationID]").WithRootCmd(rootCmd).Build()

	// Create list locations command
	listLocationsCmd := cmdbuilder.NewCommand("list", "List all locations").
		WithLongDesc("List all locations available in the Megaport API.\n\nThis command fetches and displays a list of all available locations with details such as location ID, name, country, and metro. You can also filter the locations based on specific criteria.").
		WithLocationsFilterFlags().
		WithOutputFormatRunFunc(ListLocations).
		WithExample("megaport-cli locations list").
		WithExample("megaport-cli locations list --metro \"San Francisco\"").
		WithExample("megaport-cli locations list --country \"US\"").
		WithExample("megaport-cli locations list --name \"Equinix SY1\"").WithRootCmd(rootCmd).Build()

	// Create get location command
	getLocationCmd := cmdbuilder.NewCommand("get", "Get details for a specific location by ID").
		WithArgs(cobra.ExactArgs(1)).
		WithLongDesc("Get details for a specific location from the Megaport API.\n\nThis command fetches and displays detailed information about a location using its ID. You must provide the location ID as an argument.").
		WithOutputFormatRunFunc(GetLocation).
		WithExample("megaport-cli locations get 67").
		WithRootCmd(rootCmd).
		Build()

	// Add commands to their parents
	locationsCmd.AddCommand(listLocationsCmd, getLocationCmd)
	rootCmd.AddCommand(locationsCmd)
}
