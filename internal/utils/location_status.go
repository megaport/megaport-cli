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

// The lookup asks for every status because the API leaves Extended, New, and
// Expired out of an unfiltered list.
var locationLookupFn = func(ctx context.Context, client *megaport.Client, locationID int) (*megaport.LocationV3, error) {
	locations, err := client.LocationService.ListLocationsV3WithOptions(ctx, &megaport.ListLocationsV3Options{Statuses: LocationStatuses})
	if err != nil {
		return nil, err
	}
	for _, loc := range locations {
		if loc != nil && loc.ID == locationID {
			return loc, nil
		}
	}
	return nil, nil
}

// SetLocationLookup replaces the location lookup that WarnIfLocationNotActive
// uses and returns a function that restores it.
func SetLocationLookup(fn func(context.Context, *megaport.Client, int) (*megaport.LocationV3, error)) func() {
	original := locationLookupFn
	locationLookupFn = fn
	return func() { locationLookupFn = original }
}

// WarnIfLocationNotActive prints a warning when the location isn't Active.
// It never blocks the order. The API decides whether to accept it.
func WarnIfLocationNotActive(ctx context.Context, client *megaport.Client, locationID int, noColor bool) {
	loc, err := locationLookupFn(ctx, client, locationID)
	if err != nil || loc == nil || loc.IsStatusOrderable() {
		return
	}
	output.PrintWarning("Location %d (%s) has status %s, not Active.", noColor, loc.ID, loc.Name, loc.Status)
}
