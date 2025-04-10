package locations

import (
	megaport "github.com/megaport/megaportgo"
)

// filterLocations filters the provided locations based on the given filters.
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
