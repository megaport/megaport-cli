//go:build integration

package mve

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestIntegration_MVEProductSizeCandidates is a pure, in-memory guard on the
// capacity-probe size filter: it makes no API calls. It lives under the
// integration tag because productSizeCandidates does, and is named to match the
// provisioning job's run filter so the regression is actually exercised in CI.
func TestIntegration_MVEProductSizeCandidates(t *testing.T) {
	tests := []struct {
		name           string
		availableSizes []string
		want           []string
	}{
		{
			name:           "drops unmappable label, keeps valid in preferred order",
			availableSizes: []string{"MVE 4/16", "MVE 2/8", "MVE 16/64"},
			want:           []string{"MEDIUM", "SMALL"},
		},
		{
			name:           "only unmappable labels falls back to MEDIUM",
			availableSizes: []string{"MVE 16/64", "MVE 32/128"},
			want:           []string{"MEDIUM"},
		},
		{
			name:           "empty falls back to MEDIUM",
			availableSizes: nil,
			want:           []string{"MEDIUM"},
		},
		{
			name:           "programmatic names pass through in preferred order",
			availableSizes: []string{"X_LARGE_12", "LARGE"},
			want:           []string{"LARGE", "X_LARGE_12"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := discoveredImage{AvailableSizes: tt.availableSizes}
			got := img.productSizeCandidates()
			assert.Equal(t, tt.want, got)
			assert.NotContains(t, got, "MVE 16/64", "raw label must never be probed")
		})
	}
}
