package utils

import (
	"context"

	"github.com/megaport/megaport-cli/internal/base/output"
	megaport "github.com/megaport/megaportgo"
)

// LocationStatuses lists every status the locations API accepts as a filter.
var LocationStatuses = []string{
	megaport.LocationStatusActive,
	megaport.LocationStatusDeployment,
	megaport.LocationStatusRestricted,
	megaport.LocationStatusExtended,
	megaport.LocationStatusNew,
	megaport.LocationStatusExpired,
}

// The lookup asks only for the non-Active statuses. A location missing from the
// list is Active or doesn't exist, and the warning skips both.
var locationLookupFn = func(ctx context.Context, client *megaport.Client) ([]*megaport.LocationV3, error) {
	return client.LocationService.ListLocationsV3WithOptions(ctx, &megaport.ListLocationsV3Options{Statuses: []string{
		megaport.LocationStatusDeployment,
		megaport.LocationStatusRestricted,
		megaport.LocationStatusExtended,
		megaport.LocationStatusNew,
		megaport.LocationStatusExpired,
	}})
}

// SetLocationLookup replaces the location lookup that WarnIfLocationNotActive
// uses and returns a function that restores it.
func SetLocationLookup(fn func(context.Context, *megaport.Client) ([]*megaport.LocationV3, error)) func() {
	original := locationLookupFn
	locationLookupFn = fn
	return func() { locationLookupFn = original }
}

// WarnIfLocationNotActive prints a warning for each location that isn't Active.
// It never blocks the order. The API decides whether to accept it.
func WarnIfLocationNotActive(ctx context.Context, client *megaport.Client, noColor bool, locationIDs ...int) {
	if len(locationIDs) == 0 {
		return
	}
	locations, err := locationLookupFn(ctx, client)
	if err != nil {
		return
	}
	notActive := make(map[int]*megaport.LocationV3, len(locations))
	for _, loc := range locations {
		if loc != nil && !loc.IsStatusOrderable() {
			notActive[loc.ID] = loc
		}
	}
	for _, id := range locationIDs {
		if loc, ok := notActive[id]; ok {
			output.PrintWarning("Location %d (%s) has status %s, not Active.", noColor, loc.ID, loc.Name, loc.Status)
			delete(notActive, id)
		}
	}
}
