package locations

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/commands/config"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

func ListLocations(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	// Create a context with a 30-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Log into the Megaport API.
	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("error logging in: %v", err)
	}

	// Start a spinner to show progress while listing locations
	spinner := output.PrintResourceListing("Location", noColor)

	// Retrieve the list of locations from the API.
	locations, err := listLocationsFunc(ctx, client)

	// Stop the spinner
	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to retrieve locations: %v", noColor, err)
		return fmt.Errorf("error listing locations: %v", err)
	}

	// Apply filters if provided.
	filters := map[string]string{}
	if cmd.Flags().Changed("metro") {
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

	// Filter locations based on the provided flags.
	filteredLocations := filterLocations(locations, filters)

	if len(filteredLocations) == 0 {
		output.PrintWarning("No locations found matching the specified filters", noColor)
	} else {
		output.PrintInfo("Found %d locations", noColor, len(filteredLocations))
	}

	err = printLocations(filteredLocations, outputFormat, noColor)
	if err != nil {
		output.PrintError("Failed to print locations: %v", noColor, err)
		return fmt.Errorf("error printing locations: %v", err)
	}
	return nil
}

func GetLocation(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	// Create a context with a 30-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Log into the Megaport API.
	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("error logging in: %v", err)
	}

	// Parse location ID from args
	locationID, err := strconv.Atoi(args[0])
	if err != nil {
		output.PrintError("Invalid location ID: %v", noColor, err)
		return fmt.Errorf("invalid location ID: %v", err)
	}

	// Start a spinner to show progress while retrieving the location
	spinner := output.PrintResourceGetting("Location", fmt.Sprintf("%d", locationID), noColor)

	// Retrieve the list of locations from the API.
	locations, err := listLocationsFunc(ctx, client)

	if err != nil {
		spinner.Stop() // Make sure to stop spinner on error
		output.PrintError("Failed to retrieve locations: %v", noColor, err)
		return fmt.Errorf("error listing locations: %v", err)
	}

	// Find the location with the matching ID
	var targetLocation *megaport.Location
	for _, loc := range locations {
		if loc.ID == locationID {
			targetLocation = loc
			break
		}
	}

	if targetLocation == nil {
		spinner.Stop() // Stop the spinner before showing error
		output.PrintWarning("No location found with ID: %d", noColor, locationID)
		return fmt.Errorf("no location found with ID: %d", locationID)
	}

	// Stop the spinner with success message
	spinner.StopWithSuccess(fmt.Sprintf("Found location with ID: %d", locationID))

	// Print location details
	err = printLocations([]*megaport.Location{targetLocation}, outputFormat, noColor)
	if err != nil {
		output.PrintError("Failed to print location details: %v", noColor, err)
		return err
	}
	return nil
}
