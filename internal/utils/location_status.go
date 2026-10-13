package utils

import (
	"context"

	"github.com/megaport/megaport-cli/internal/base/output"
	megaport "github.com/megaport/megaportgo"
)

var locationLookupFn = func(ctx context.Context, client *megaport.Client, locationID int) (*megaport.LocationV3, error) {
	return client.LocationService.GetLocationByIDV3(ctx, locationID)
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
