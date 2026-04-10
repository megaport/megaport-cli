package locations

import (
	"github.com/megaport/megaport-cli/internal/base/cmdbuilder"
	"github.com/spf13/cobra"
)

func AddCommandsTo(rootCmd *cobra.Command) {
	locationsCmd := cmdbuilder.NewCommand("locations", "Manage locations in the Megaport API").
		WithLongDesc("Manage locations in the Megaport API.\n\nThis command groups all operations related to locations. You can use its subcommands to list and get details for specific locations.").
		WithExample("megaport-cli locations list").
		WithExample("megaport-cli locations get [locationID]").WithRootCmd(rootCmd).Build()

	listLocationsCmd := cmdbuilder.NewCommand("list", "List all locations").
		WithLongDesc("List all locations available in the Megaport API.\n\nThis command fetches and displays a list of all available locations with details such as location ID, name, country, and metro. You can also filter the locations based on specific criteria.").
		WithLocationsFilterFlags().
		WithOutputFormatRunFunc(ListLocations).
		WithExample("megaport-cli locations list").
		WithExample("megaport-cli locations list --metro \"San Francisco\"").
		WithExample("megaport-cli locations list --country \"US\"").
		WithExample("megaport-cli locations list --name \"Equinix SY1\"").
		WithIntFlag("limit", 0, "Maximum number of results to display (0 = unlimited)").
		WithRootCmd(rootCmd).
		WithAliases([]string{"ls"}).Build()

	getLocationCmd := cmdbuilder.NewCommand("get", "Get details for a specific location by ID").
		WithArgs(cobra.ExactArgs(1)).
		WithLongDesc("Get details for a specific location from the Megaport API.\n\nThis command fetches and displays detailed information about a location using its ID. You must provide the location ID as an argument.").
		WithOutputFormatRunFunc(GetLocation).
		WithExample("megaport-cli locations get 67").
		WithRootCmd(rootCmd).
		WithAliases([]string{"show"}).
		Build()

	listCountriesCmd := cmdbuilder.NewCommand("list-countries", "List all countries with Megaport locations").
		WithLongDesc("List all countries that have Megaport locations.\n\nThis command fetches and displays a list of all countries where Megaport has available locations, including country code, name, prefix, and site count.").
		WithOutputFormatRunFunc(ListCountries).
		WithExample("megaport-cli locations list-countries").
		WithRootCmd(rootCmd).
		Build()

	listMarketCodesCmd := cmdbuilder.NewCommand("list-market-codes", "List all market codes").
		WithLongDesc("List all market codes used to categorize Megaport locations.\n\nThis command fetches and displays a list of all market codes available in the Megaport API.").
		WithOutputFormatRunFunc(ListMarketCodes).
		WithExample("megaport-cli locations list-market-codes").
		WithRootCmd(rootCmd).
		Build()

	searchCmd := cmdbuilder.NewCommand("search", "Search locations by name (fuzzy match)").
		WithArgs(cobra.ExactArgs(1)).
		WithLongDesc("Search for locations by name using fuzzy matching.\n\nThis command searches for locations whose names match the provided search term. The search uses fuzzy matching, so partial or approximate names will return results.").
		WithOutputFormatRunFunc(SearchLocations).
		WithExample("megaport-cli locations search \"Equinix\"").
		WithExample("megaport-cli locations search \"Sydney\"").
		WithRootCmd(rootCmd).
		Build()

	rttCmd := cmdbuilder.NewCommand("rtt", "Query round-trip times between locations").
		WithOutputFormatRunFunc(GetRoundTripTimes).
		WithLongDesc("Query median round-trip times (RTT) between Megaport locations.\n\n"+
			"This command retrieves latency data between a source location and all other "+
			"Megaport locations for a given month. Use this for network planning — choosing "+
			"MCR locations and designing cross-connects based on latency requirements.\n\n"+
			"By default, returns data for the current month. Use --year and --month to query "+
			"historical data.").
		WithIntFlag("src-location", 0, "Source location ID").
		WithDocumentedRequiredFlag("src-location", "Source location ID").
		WithIntFlag("dst-location", 0, "Filter results to a specific destination location ID").
		WithIntFlag("year", 0, "Year for RTT data (default: current year)").
		WithIntFlag("month", 0, "Month for RTT data (1-12, default: current month)").
		WithExample("megaport-cli locations rtt --src-location 67").
		WithExample("megaport-cli locations rtt --src-location 67 --dst-location 3").
		WithExample("megaport-cli locations rtt --src-location 67 --year 2026 --month 3").
		WithExample("megaport-cli locations rtt --src-location 67 --output json").
		WithRootCmd(rootCmd).
		Build()

	locationsCmd.AddCommand(listLocationsCmd, getLocationCmd, listCountriesCmd, listMarketCodesCmd, searchCmd, rttCmd)
	rootCmd.AddCommand(locationsCmd)
}
