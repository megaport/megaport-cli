package locations

import (
	"context"
	"time"

	"github.com/megaport/megaport-cli/internal/utils"
	megaport "github.com/megaport/megaportgo"
)

// timeNow is a mockable time function for testing default year/month.
var timeNow = time.Now

func filterLocations(locations []*megaport.LocationV3, filters map[string]string) []*megaport.LocationV3 {
	return utils.Filter(locations, func(loc *megaport.LocationV3) bool {
		if metro, ok := filters["metro"]; ok && loc.Metro != metro {
			return false
		}
		if country, ok := filters["country"]; ok && loc.Address.Country != country {
			return false
		}
		if name, ok := filters["name"]; ok && loc.Name != name {
			return false
		}
		if market, ok := filters["market"]; ok && loc.Market != market {
			return false
		}
		if val, ok := filters["mcrAvailable"]; ok && val == "true" && !loc.HasMCRSupport() {
			return false
		}
		return true
	})
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

var getRoundTripTimesFunc = func(ctx context.Context, client *megaport.Client, srcLocationID, year, month int) ([]*megaport.RoundTripTime, error) {
	return client.LocationService.GetRoundTripTimes(ctx, srcLocationID, year, month)
}
