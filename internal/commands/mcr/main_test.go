package mcr

import (
	"context"
	"os"
	"testing"

	"github.com/megaport/megaport-cli/internal/utils"
	megaport "github.com/megaport/megaportgo"
)

// The buy tests use a mock client with no LocationService, so the location
// status lookup returns an Active location unless a test replaces it.
func TestMain(m *testing.M) {
	utils.SetLocationLookup(func(context.Context, *megaport.Client, int) (*megaport.LocationV3, error) {
		return &megaport.LocationV3{Status: megaport.LocationStatusActive}, nil
	})
	os.Exit(m.Run())
}
