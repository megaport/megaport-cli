package utils

import (
	"context"
	"errors"
	"strings"
	"testing"

	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWarnIfLocationNotActive(t *testing.T) {
	locations := []*megaport.LocationV3{
		{ID: 42, Name: "Test DC", Status: megaport.LocationStatusRestricted},
		{ID: 43, Name: "Other DC", Status: megaport.LocationStatusDeployment},
		{ID: 44, Name: "Active DC", Status: megaport.LocationStatusActive},
		nil,
	}
	restricted := "Location 42 (Test DC) has status Restricted, not Active."
	deployment := "Location 43 (Other DC) has status Deployment, not Active."

	tests := []struct {
		name      string
		ids       []int
		err       error
		want      []string
		wantCalls int
	}{
		{
			name:      "restricted location warns",
			ids:       []int{42},
			want:      []string{restricted},
			wantCalls: 1,
		},
		{
			name:      "active location is silent",
			ids:       []int{44},
			wantCalls: 1,
		},
		{
			name:      "location missing from the list is silent",
			ids:       []int{99},
			wantCalls: 1,
		},
		{
			name:      "lookup error is silent",
			ids:       []int{42},
			err:       errors.New("lookup failed"),
			wantCalls: 1,
		},
		{
			name:      "one lookup warns once for each location",
			ids:       []int{43, 44, 42, 43},
			want:      []string{deployment, restricted},
			wantCalls: 1,
		},
		{
			name: "no location IDs skips the lookup",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			restore := SetLocationLookup(func(context.Context, *megaport.Client) ([]*megaport.LocationV3, error) {
				calls++
				return locations, tt.err
			})
			defer restore()

			stderr := captureStderr(t, func() {
				WarnIfLocationNotActive(context.Background(), nil, true, tt.ids...)
			})

			assert.Equal(t, tt.wantCalls, calls)
			if len(tt.want) == 0 {
				assert.Empty(t, stderr)
			}
			last := -1
			for _, w := range tt.want {
				assert.Equal(t, 1, strings.Count(stderr, w), w)
				assert.Greater(t, strings.Index(stderr, w), last, "warnings follow the order of the IDs")
				last = strings.Index(stderr, w)
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
		{ID: 42, Name: "Extended DC", Status: megaport.LocationStatusExtended},
	}}
	client := &megaport.Client{LocationService: svc}

	stderr := captureStderr(t, func() {
		WarnIfLocationNotActive(context.Background(), client, true, 42)
	})

	assert.Contains(t, stderr, "Location 42 (Extended DC) has status Extended, not Active.")
	require.NotNil(t, svc.opts)
	assert.ElementsMatch(t, []string{"Deployment", "Restricted", "Extended", "New", "Expired"}, svc.opts.Statuses)
}
