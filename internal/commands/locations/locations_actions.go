package locations

import (
	"context"
	"fmt"
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

	output.PrintInfo("Retrieving locations...", noColor)

	// Retrieve the list of locations from the API.
	locations, err := client.LocationService.ListLocations(ctx)
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

	output.PrintInfo("Retrieving locations...", noColor)

	// Retrieve the list of locations from the API.
	locations, err := client.LocationService.ListLocations(ctx)
	if err != nil {
		output.PrintError("Failed to retrieve locations: %v", noColor, err)
		return fmt.Errorf("error listing locations: %v", err)
	}

	// Filter locations based on the provided flags.
	var filteredLocations []*megaport.Location
	var searchCriteria string

	if cmd.Flags().Changed("id") {
		id, _ := cmd.Flags().GetInt("id")
		searchCriteria = fmt.Sprintf("ID: %d", id)
		for _, loc := range locations {
			if loc.ID == id {
				filteredLocations = append(filteredLocations, loc)
				break
			}
		}
	} else if cmd.Flags().Changed("site-code") {
		siteCode, _ := cmd.Flags().GetString("site-code")
		searchCriteria = fmt.Sprintf("site code: %s", siteCode)
		for _, loc := range locations {
			if loc.SiteCode == siteCode {
				filteredLocations = append(filteredLocations, loc)
				break
			}
		}
	} else if cmd.Flags().Changed("name") {
		name, _ := cmd.Flags().GetString("name")
		searchCriteria = fmt.Sprintf("name: %s", name)
		for _, loc := range locations {
			if loc.Name == name {
				filteredLocations = append(filteredLocations, loc)
				break
			}
		}
	} else {
		output.PrintError("Missing search criteria", noColor)
		return fmt.Errorf("please specify one of the following flags: --id, --site-code, --name")
	}

	// output.Print the filtered location.
	if len(filteredLocations) == 0 {
		output.PrintWarning("No location found with %s", noColor, searchCriteria)
		return fmt.Errorf("no location found with %s", searchCriteria)
	}

	output.PrintSuccess("Found location with %s", noColor, searchCriteria)

	err = printLocations(filteredLocations, outputFormat, noColor)
	if err != nil {
		output.PrintError("Failed to print location details: %v", noColor, err)
		return err
	}
	return nil
}
