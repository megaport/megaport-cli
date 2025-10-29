package locations

import (
	"context"

	megaport "github.com/megaport/megaportgo"
)

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

// listLocationsFunc now uses the v3 API and converts to legacy format for compatibility
var listLocationsFunc = func(ctx context.Context, client *megaport.Client) ([]*megaport.Location, error) {
	// Use v3 API (recommended)
	locationsV3, err := client.LocationService.ListLocationsV3(ctx)
	if err != nil {
		return nil, err
	}

	// Convert v3 locations to legacy format for backward compatibility
	var legacyLocations []*megaport.Location
	for _, v3Loc := range locationsV3 {
		legacyLocations = append(legacyLocations, v3Loc.ToLegacyLocation())
	}

	return legacyLocations, nil
}
