package locations

import (
	"context"
	"strings"
	"time"

	"github.com/megaport/megaport-cli/internal/utils"
	megaport "github.com/megaport/megaportgo"
)

// timeNow is a mockable time function for testing default year/month.
var timeNow = time.Now

func filterLocations(locations []*megaport.LocationV3, filters map[string]string) []*megaport.LocationV3 {
	return utils.Filter(locations, func(loc *megaport.LocationV3) bool {
		if loc == nil {
			return false
		}
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

// listLocationsFunc returns Active, Deployment, and Restricted locations when
// statuses is empty. The API applies that default.
var listLocationsFunc = func(ctx context.Context, client *megaport.Client, statuses []string) ([]*megaport.LocationV3, error) {
	return client.LocationService.ListLocationsV3WithOptions(ctx, &megaport.ListLocationsV3Options{Statuses: statuses})
}

// normalizeLocationStatuses fixes the case of each known status, since the API
// rejects "active". It passes an unknown value through for the API to reject.
func normalizeLocationStatuses(values []string) []string {
	statuses := make([]string, 0, len(values))
	for _, v := range values {
		status := v
		for _, known := range utils.LocationStatuses {
			if strings.EqualFold(v, known) {
				status = known
				break
			}
		}
		statuses = append(statuses, status)
	}
	return statuses
}

var getRoundTripTimesFunc = func(ctx context.Context, client *megaport.Client, srcLocationID, year, month int) ([]*megaport.RoundTripTime, error) {
	return client.LocationService.GetRoundTripTimes(ctx, srcLocationID, year, month)
}
