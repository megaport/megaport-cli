package utils

import (
	"context"
	"errors"
	"testing"

	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
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
