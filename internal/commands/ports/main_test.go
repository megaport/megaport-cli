package ports

import (
	"context"
	"os"
	"testing"

	"github.com/megaport/megaport-cli/internal/utils"
	megaport "github.com/megaport/megaportgo"
)

// The buy tests use a mock client with no LocationService, so the location
// status lookup returns no non-Active locations unless a test replaces it.
func TestMain(m *testing.M) {
	utils.SetLocationLookup(func(context.Context, *megaport.Client) ([]*megaport.LocationV3, error) {
		return nil, nil
	})
	os.Exit(m.Run())
}
