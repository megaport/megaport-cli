package ports

import (
	"context"
	"os"
	"testing"

	"github.com/megaport/megaport-cli/internal/utils"
	megaport "github.com/megaport/megaportgo"
)

// The mock client has no LocationService, so this stub returns an empty list
// instead of calling it. A test replaces the stub to return locations.
func TestMain(m *testing.M) {
	utils.SetLocationLookup(func(context.Context, *megaport.Client) ([]*megaport.LocationV3, error) {
		return nil, nil
	})
	os.Exit(m.Run())
}
