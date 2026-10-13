package utils

import (
	"context"
	"errors"
	"testing"

	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWarnIfLocationNotActive(t *testing.T) {
	tests := []struct {
		name     string
		location *megaport.LocationV3
		err      error
		want     string
	}{
		{
			name:     "restricted location warns",
			location: &megaport.LocationV3{ID: 42, Name: "Test DC", Status: megaport.LocationStatusRestricted},
			want:     "Location 42 (Test DC) has status Restricted, not Active.",
		},
		{
			name:     "active location is silent",
			location: &megaport.LocationV3{ID: 42, Name: "Test DC", Status: megaport.LocationStatusActive},
		},
		{
			name: "lookup error is silent",
			err:  errors.New("not found"),
		},
		{
			name: "nil location is silent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			restore := SetLocationLookup(func(context.Context, *megaport.Client, int) (*megaport.LocationV3, error) {
				return tt.location, tt.err
			})
			defer restore()

			stderr := captureStderr(t, func() {
				WarnIfLocationNotActive(context.Background(), nil, 42, true)
			})

			if tt.want == "" {
				assert.Empty(t, stderr)
			} else {
				assert.Contains(t, stderr, tt.want)
			}
		})
	}
}

type fakeLocationService struct {
	megaport.LocationService
	locations []*megaport.LocationV3
	opts      *megaport.ListLocationsV3Options
}

func (f *fakeLocationService) ListLocationsV3WithOptions(_ context.Context, opts *megaport.ListLocationsV3Options) ([]*megaport.LocationV3, error) {
	f.opts = opts
	return f.locations, nil
}

func TestWarnIfLocationNotActive_ExtendedLocation(t *testing.T) {
	svc := &fakeLocationService{locations: []*megaport.LocationV3{
		{ID: 1, Name: "Active DC", Status: megaport.LocationStatusActive},
		{ID: 42, Name: "Extended DC", Status: megaport.LocationStatusExtended},
	}}
	client := &megaport.Client{LocationService: svc}

	stderr := captureStderr(t, func() {
		WarnIfLocationNotActive(context.Background(), client, 42, true)
	})

	assert.Contains(t, stderr, "Location 42 (Extended DC) has status Extended, not Active.")
	require.NotNil(t, svc.opts)
	assert.ElementsMatch(t, []string{"Active", "Deployment", "Restricted", "Extended", "New", "Expired"}, svc.opts.Statuses)
}
