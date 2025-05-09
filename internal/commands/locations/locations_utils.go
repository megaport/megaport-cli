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

var listLocationsFunc = func(ctx context.Context, client *megaport.Client) ([]*megaport.Location, error) {
	return client.LocationService.ListLocations(ctx)
}
