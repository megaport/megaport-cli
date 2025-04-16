package validation

import (
	"fmt"
	"testing"
)

func TestValidateMCRRequest(t *testing.T) {
	tests := []struct {
		name       string
		mcrName    string
		term       int
		portSpeed  int
		locationID int
		wantErr    bool
		errText    string // Add expected error text for specific cases
	}{
		{
			name:       "Valid MCR request",
			mcrName:    "Test MCR",
			term:       12,
			portSpeed:  5000,
			locationID: 100,
			wantErr:    false,
		},
		{
			name:       "Empty MCR name",
			mcrName:    "",
			term:       12,
			portSpeed:  5000,
			locationID: 100,
			wantErr:    true,
			errText:    "Invalid MCR name:  - cannot be empty", // Use ValidationError format
		},
		{
			name:       "Invalid term",
			mcrName:    "Test MCR",
			term:       5,
			portSpeed:  5000,
			locationID: 100,
			wantErr:    true,
			errText:    fmt.Sprintf("Invalid contract term: 5 - must be one of: %v", ValidContractTerms), // Expect ValidationError message
		},
		{
			name:       "Invalid port speed",
			mcrName:    "Test MCR",
			term:       12,
			portSpeed:  3000,
			locationID: 100,
			wantErr:    true,
			errText:    fmt.Sprintf("Invalid MCR port speed: 3000 - must be one of: %v", ValidMCRPortSpeeds), // Use ValidationError format
		},
		{
			name:       "Invalid location ID",
			mcrName:    "Test MCR",
			term:       12,
			portSpeed:  5000,
			locationID: 0,
			wantErr:    true,
			errText:    "Invalid location ID: 0 - must be a positive integer", // Use ValidationError format
		},
		{
			name:       "Negative location ID",
			mcrName:    "Test MCR",
			term:       12,
			portSpeed:  5000,
			locationID: -1,
			wantErr:    true,
			errText:    "Invalid location ID: -1 - must be a positive integer", // Use ValidationError format
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMCRRequest(tt.mcrName, tt.term, tt.portSpeed, tt.locationID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMCRRequest() error = %v, wantErr %v", err, tt.wantErr)
				return // Avoid panic on nil error
			}
			// Check if the error message matches the expected text when an error is expected
			if err != nil && tt.wantErr {
				if err.Error() != tt.errText {
					t.Errorf("ValidateMCRRequest() error text = %q, want %q", err.Error(), tt.errText)
				}
				// Check if the error type is *ValidationError
				if _, ok := err.(*ValidationError); !ok {
					t.Errorf("ValidateMCRRequest() should return *ValidationError, got %T", err)
				}
			}
		})
	}
}
