package cmd

import (
	"context"
	"fmt"
	"time"

	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

func ListLocations(cmd *cobra.Command, args []string) error {
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
}

func GetLocation(cmd *cobra.Command, args []string) error {
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
}
