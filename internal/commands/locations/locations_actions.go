package locations

import (
	"fmt"
	"strconv"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/commands/config"
	"github.com/megaport/megaport-cli/internal/utils"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

func ListLocations(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	output.SetOutputFormat(outputFormat)

	ctx, cancel := utils.ContextFromCmd(cmd)
	defer cancel()

	client, err := config.NewUnauthenticatedClient()
	if err != nil {
		output.PrintError("Failed to create API client: %v", noColor, err)
		return fmt.Errorf("error creating API client: %w", err)
	}

	spinner := output.PrintResourceListing("Location", noColor)

	locations, err := listLocationsFunc(ctx, client)

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to retrieve locations: %v", noColor, err)
		return fmt.Errorf("error listing locations: %w", err)
	}

	filters := map[string]string{}
	if cmd.Flags().Changed("metro") {
		// Flag read errors are intentionally ignored — flags are registered by the command builder.
		metro, _ := cmd.Flags().GetString("metro")
		filters["metro"] = metro
		output.PrintInfo("Filtering by metro: %s", noColor, metro)
	}
	if cmd.Flags().Changed("country") {
		country, _ := cmd.Flags().GetString("country")
		filters["country"] = country
		output.PrintInfo("Filtering by country: %s", noColor, country)
	}
	if cmd.Flags().Changed("name") {
		name, _ := cmd.Flags().GetString("name")
		filters["name"] = name
		output.PrintInfo("Filtering by name: %s", noColor, name)
	}
	if cmd.Flags().Changed("market-code") {
		marketCode, _ := cmd.Flags().GetString("market-code")
		filters["market"] = marketCode
		output.PrintInfo("Filtering by market code: %s", noColor, marketCode)
	}
	if cmd.Flags().Changed("mcr-available") {
		mcrAvailable, _ := cmd.Flags().GetBool("mcr-available")
		if mcrAvailable {
			filters["mcrAvailable"] = "true"
			output.PrintInfo("Filtering by MCR availability", noColor)
		}
	}

	filteredLocations := filterLocations(locations, filters)

	limit, _ := cmd.Flags().GetInt("limit")
	if limit < 0 {
		return fmt.Errorf("--limit must be a non-negative integer")
	}
	if limit > 0 && len(filteredLocations) > limit {
		filteredLocations = filteredLocations[:limit]
	}

	if len(filteredLocations) == 0 {
		if outputFormat == utils.FormatTable {
			output.PrintInfo("No locations found matching your filters.", noColor)
		}
		return nil
	}

	err = printLocations(filteredLocations, outputFormat, noColor)
	if err != nil {
		output.PrintError("Failed to print locations: %v", noColor, err)
		return fmt.Errorf("error printing locations: %w", err)
	}
	return nil
}

func ListCountries(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	output.SetOutputFormat(outputFormat)

	ctx, cancel := utils.ContextFromCmd(cmd)
	defer cancel()

	client, err := config.NewUnauthenticatedClient()
	if err != nil {
		output.PrintError("Failed to create API client: %v", noColor, err)
		return fmt.Errorf("error creating API client: %w", err)
	}

	spinner := output.PrintResourceListing("Country", noColor)

	countries, err := listCountriesFunc(ctx, client)

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to retrieve countries: %v", noColor, err)
		return fmt.Errorf("error listing countries: %w", err)
	}

	if len(countries) == 0 {
		output.PrintWarning("No countries found", noColor)
	} else {
		output.PrintInfo("Found %d countries", noColor, len(countries))
	}

	err = printCountries(countries, outputFormat, noColor)
	if err != nil {
		output.PrintError("Failed to print countries: %v", noColor, err)
		return fmt.Errorf("error printing countries: %w", err)
	}
	return nil
}

func ListMarketCodes(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	output.SetOutputFormat(outputFormat)

	ctx, cancel := utils.ContextFromCmd(cmd)
	defer cancel()

	client, err := config.NewUnauthenticatedClient()
	if err != nil {
		output.PrintError("Failed to create API client: %v", noColor, err)
		return fmt.Errorf("error creating API client: %w", err)
	}

	spinner := output.PrintResourceListing("Market Code", noColor)

	marketCodes, err := listMarketCodesFunc(ctx, client)

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to retrieve market codes: %v", noColor, err)
		return fmt.Errorf("error listing market codes: %w", err)
	}

	if len(marketCodes) == 0 {
		output.PrintWarning("No market codes found", noColor)
	} else {
		output.PrintInfo("Found %d market codes", noColor, len(marketCodes))
	}

	err = printMarketCodes(marketCodes, outputFormat, noColor)
	if err != nil {
		output.PrintError("Failed to print market codes: %v", noColor, err)
		return fmt.Errorf("error printing market codes: %w", err)
	}
	return nil
}

func SearchLocations(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	output.SetOutputFormat(outputFormat)

	ctx, cancel := utils.ContextFromCmd(cmd)
	defer cancel()

	client, err := config.NewUnauthenticatedClient()
	if err != nil {
		output.PrintError("Failed to create API client: %v", noColor, err)
		return fmt.Errorf("error creating API client: %w", err)
	}

	search := args[0]

	spinner := output.PrintResourceListing("Location", noColor)

	locations, err := searchLocationsFunc(ctx, client, search)

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to search locations: %v", noColor, err)
		return fmt.Errorf("error searching locations: %w", err)
	}

	if len(locations) == 0 {
		output.PrintWarning("No locations found matching '%s'", noColor, search)
	} else {
		output.PrintInfo("Found %d locations matching '%s'", noColor, len(locations), search)
	}

	err = printLocations(locations, outputFormat, noColor)
	if err != nil {
		output.PrintError("Failed to print locations: %v", noColor, err)
		return fmt.Errorf("error printing locations: %w", err)
	}
	return nil
}

func GetLocation(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	output.SetOutputFormat(outputFormat)

	ctx, cancel := utils.ContextFromCmd(cmd)
	defer cancel()

	client, err := config.NewUnauthenticatedClient()
	if err != nil {
		output.PrintError("Failed to create API client: %v", noColor, err)
		return fmt.Errorf("error creating API client: %w", err)
	}

	locationID, err := strconv.Atoi(args[0])
	if err != nil {
		output.PrintError("Invalid location ID: %v", noColor, err)
		return fmt.Errorf("invalid location ID: %w", err)
	}

	spinner := output.PrintResourceGetting("Location", fmt.Sprintf("%d", locationID), noColor)

	locations, err := listLocationsFunc(ctx, client)

	if err != nil {
		spinner.Stop()
		output.PrintError("Failed to retrieve locations: %v", noColor, err)
		return fmt.Errorf("error listing locations: %w", err)
	}

	var targetLocation *megaport.LocationV3
	for _, loc := range locations {
		if loc.ID == locationID {
			targetLocation = loc
			break
		}
	}

	if targetLocation == nil {
		spinner.Stop()
		output.PrintWarning("No location found with ID: %d", noColor, locationID)
		return fmt.Errorf("no location found with ID: %d", locationID)
	}

	spinner.StopWithSuccess(fmt.Sprintf("Found location with ID: %d", locationID))

	err = printLocations([]*megaport.LocationV3{targetLocation}, outputFormat, noColor)
	if err != nil {
		output.PrintError("Failed to print location details: %v", noColor, err)
		return err
	}
	return nil
}
