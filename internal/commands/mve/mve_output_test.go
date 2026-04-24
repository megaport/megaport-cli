package mve

import (
	"testing"

	op "github.com/megaport/megaport-cli/internal/base/output"
	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
)

func TestToMVEOutput_NilMVE(t *testing.T) {
	_, err := toMVEOutput(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nil value")
}

func TestToMVEOutput_Valid(t *testing.T) {
	mve := &megaport.MVE{
		UID:                "mve-123",
		Name:               "Test MVE",
		LocationID:         1,
		ProvisioningStatus: "LIVE",
		Vendor:             "cisco",
		Size:               "MEDIUM",
	}

	out, err := toMVEOutput(mve)
	assert.NoError(t, err)
	assert.Equal(t, "mve-123", out.UID)
	assert.Equal(t, "Test MVE", out.Name)
	assert.Equal(t, "cisco", out.Vendor)
	assert.Equal(t, "MEDIUM", out.Size)
}

func TestDisplayMVEChanges(t *testing.T) {
	tests := []struct {
		name             string
		original         *megaport.MVE
		updated          *megaport.MVE
		expectedContains []string
	}{
		{
			name:     "nil original",
			original: nil,
			updated:  &megaport.MVE{},
		},
		{
			name:     "nil updated",
			original: &megaport.MVE{},
			updated:  nil,
		},
		{
			name:             "no changes",
			original:         &megaport.MVE{Name: "Same", CostCentre: "CC1", ContractTermMonths: 12},
			updated:          &megaport.MVE{Name: "Same", CostCentre: "CC1", ContractTermMonths: 12},
			expectedContains: []string{"No changes detected"},
		},
		{
			name:             "name changed",
			original:         &megaport.MVE{Name: "Old MVE"},
			updated:          &megaport.MVE{Name: "New MVE"},
			expectedContains: []string{"Name", "Old MVE", "New MVE"},
		},
		{
			name:             "cost centre changed from empty",
			original:         &megaport.MVE{CostCentre: ""},
			updated:          &megaport.MVE{CostCentre: "IT-2024"},
			expectedContains: []string{"Cost Centre", "(none)", "IT-2024"},
		},
		{
			name:             "contract term changed",
			original:         &megaport.MVE{ContractTermMonths: 12},
			updated:          &megaport.MVE{ContractTermMonths: 24},
			expectedContains: []string{"Contract Term", "12 months", "24 months"},
		},
		{
			name:             "marketplace visibility changed",
			original:         &megaport.MVE{MarketplaceVisibility: false},
			updated:          &megaport.MVE{MarketplaceVisibility: true},
			expectedContains: []string{"Marketplace Visibility", "No", "Yes"},
		},
		{
			name:             "multiple changes",
			original:         &megaport.MVE{Name: "Old", ContractTermMonths: 12},
			updated:          &megaport.MVE{Name: "New", ContractTermMonths: 24},
			expectedContains: []string{"Name", "Contract Term"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := op.CaptureOutput(func() {
				displayMVEChanges(tt.original, tt.updated, true)
			})

			for _, expected := range tt.expectedContains {
				assert.Contains(t, output, expected)
			}
		})
	}
}
