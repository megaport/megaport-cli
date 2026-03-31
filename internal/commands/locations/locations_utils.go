package locations

import (
	"context"

	megaport "github.com/megaport/megaportgo"
)

func filterLocations(locations []*megaport.LocationV3, filters map[string]string) []*megaport.LocationV3 {
	var filtered []*megaport.LocationV3
	for _, loc := range locations {
		if metro, ok := filters["metro"]; ok && loc.Metro != metro {
			continue
		}
		if country, ok := filters["country"]; ok && loc.Address.Country != country {
			continue
		}
		if name, ok := filters["name"]; ok && loc.Name != name {
			continue
		}
		if market, ok := filters["market"]; ok && loc.Market != market {
			continue
		}
		if val, ok := filters["mcrAvailable"]; ok && val == "true" && !loc.HasMCRSupport() {
			continue
		}
		filtered = append(filtered, loc)
	}
	return filtered
}

var listCountriesFunc = func(ctx context.Context, client *megaport.Client) ([]*megaport.Country, error) {
	return client.LocationService.ListCountries(ctx)
}

var listMarketCodesFunc = func(ctx context.Context, client *megaport.Client) ([]string, error) {
	return client.LocationService.ListMarketCodes(ctx)
}

var searchLocationsFunc = func(ctx context.Context, client *megaport.Client, search string) ([]*megaport.LocationV3, error) {
	return client.LocationService.GetLocationByNameFuzzyV3(ctx, search)
}

var listLocationsFunc = func(ctx context.Context, client *megaport.Client) ([]*megaport.LocationV3, error) {
	return client.LocationService.ListLocationsV3(ctx)
}
