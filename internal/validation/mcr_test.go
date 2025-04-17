package validation

import (
	"fmt"
	"testing"

	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
)

func TestValidateMCRRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *megaport.BuyMCRRequest
		wantErr bool
		errText string // Add expected error text for specific cases
	}{
		{
			name: "Valid MCR request",
			req: &megaport.BuyMCRRequest{
				Name:       "Test MCR",
				Term:       12,
				PortSpeed:  5000,
				LocationID: 100,
			},
			wantErr: false,
		},
		{
			name: "Empty MCR name",
			req: &megaport.BuyMCRRequest{
				Name:       "",
				Term:       12,
				PortSpeed:  5000,
				LocationID: 100,
			},
			wantErr: true,
			errText: "Invalid MCR name:  - cannot be empty", // Use ValidationError format
		},
		{
			name: "Invalid term",
			req: &megaport.BuyMCRRequest{
				Name:       "Test MCR",
				Term:       5,
				PortSpeed:  5000,
				LocationID: 100,
			},
			wantErr: true,
			errText: fmt.Sprintf("Invalid contract term: 5 - must be one of: %v", ValidContractTerms), // Expect ValidationError message
		},
		{
			name: "Invalid port speed",
			req: &megaport.BuyMCRRequest{
				Name:       "Test MCR",
				Term:       12,
				PortSpeed:  3000,
				LocationID: 100,
			},
			wantErr: true,
			errText: fmt.Sprintf("Invalid MCR port speed: 3000 - must be one of: %v", ValidMCRPortSpeeds), // Use ValidationError format
		},
		{
			name: "Invalid location ID",
			req: &megaport.BuyMCRRequest{
				Name:       "Test MCR",
				Term:       12,
				PortSpeed:  5000,
				LocationID: 0,
			},
			wantErr: true,
			errText: "Invalid location ID: 0 - must be a positive integer", // Use ValidationError format
		},
		{
			name: "Negative location ID",
			req: &megaport.BuyMCRRequest{
				Name:       "Test MCR",
				Term:       12,
				PortSpeed:  5000,
				LocationID: -1,
			},
			wantErr: true,
			errText: "Invalid location ID: -1 - must be a positive integer", // Use ValidationError format
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMCRRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMCRRequest() error = %v, wantErr %v", err, tt.wantErr)
				return // Avoid panic on nil error
			}
			// Check if the error message matches the expected text when an error is expected
			if err != nil && tt.wantErr {
				// Check if the error type is *ValidationError
				assert.IsType(t, &ValidationError{}, err, "Expected ValidationError type")
				// Check the error message
				assert.Equal(t, tt.errText, err.Error(), "Error message mismatch")
			}
		})
	}
}
