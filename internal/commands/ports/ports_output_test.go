package ports

import (
	"testing"

	op "github.com/megaport/megaport-cli/internal/base/output"
	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
)

func TestDisplayPortChanges(t *testing.T) {
	tests := []struct {
		name             string
		original         *megaport.Port
		updated          *megaport.Port
		expectedContains []string
	}{
		{
			name:     "nil original",
			original: nil,
			updated:  &megaport.Port{},
		},
		{
			name:     "nil updated",
			original: &megaport.Port{},
			updated:  nil,
		},
		{
			name:             "no changes",
			original:         &megaport.Port{Name: "Same", CostCentre: "CC1", ContractTermMonths: 12},
			updated:          &megaport.Port{Name: "Same", CostCentre: "CC1", ContractTermMonths: 12},
			expectedContains: []string{"No changes detected"},
		},
		{
			name:             "name changed",
			original:         &megaport.Port{Name: "Old Name"},
			updated:          &megaport.Port{Name: "New Name"},
			expectedContains: []string{"Name", "Old Name", "New Name"},
		},
		{
			name:             "cost centre changed from empty",
			original:         &megaport.Port{CostCentre: ""},
			updated:          &megaport.Port{CostCentre: "IT-2024"},
			expectedContains: []string{"Cost Centre", "(none)", "IT-2024"},
		},
		{
			name:             "cost centre changed to empty",
			original:         &megaport.Port{CostCentre: "IT-2024"},
			updated:          &megaport.Port{CostCentre: ""},
			expectedContains: []string{"Cost Centre", "IT-2024", "(none)"},
		},
		{
			name:             "contract term changed",
			original:         &megaport.Port{ContractTermMonths: 12},
			updated:          &megaport.Port{ContractTermMonths: 24},
			expectedContains: []string{"Contract Term", "12 months", "24 months"},
		},
		{
			name:             "marketplace visibility changed",
			original:         &megaport.Port{MarketplaceVisibility: false},
			updated:          &megaport.Port{MarketplaceVisibility: true},
			expectedContains: []string{"Marketplace Visibility", "No", "Yes"},
		},
		{
			name:             "admin locked changed",
			original:         &megaport.Port{AdminLocked: false},
			updated:          &megaport.Port{AdminLocked: true},
			expectedContains: []string{"Locked", "No", "Yes"},
		},
		{
			name:             "multiple changes",
			original:         &megaport.Port{Name: "Old", ContractTermMonths: 12, MarketplaceVisibility: false},
			updated:          &megaport.Port{Name: "New", ContractTermMonths: 24, MarketplaceVisibility: true},
			expectedContains: []string{"Name", "Contract Term", "Marketplace Visibility"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := op.CaptureOutput(func() {
				displayPortChanges(tt.original, tt.updated, true)
			})

			for _, expected := range tt.expectedContains {
				assert.Contains(t, output, expected)
			}
		})
	}
}
