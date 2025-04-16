package validation

import (
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
		},
		{
			name:       "Invalid term",
			mcrName:    "Test MCR",
			term:       5,
			portSpeed:  5000,
			locationID: 100,
			wantErr:    true,
		},
		{
			name:       "Invalid port speed",
			mcrName:    "Test MCR",
			term:       12,
			portSpeed:  3000,
			locationID: 100,
			wantErr:    true,
		},
		{
			name:       "Invalid location ID",
			mcrName:    "Test MCR",
			term:       12,
			portSpeed:  5000,
			locationID: 0,
			wantErr:    true,
		},
		{
			name:       "Negative location ID",
			mcrName:    "Test MCR",
			term:       12,
			portSpeed:  5000,
			locationID: -1,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMCRRequest(tt.mcrName, tt.term, tt.portSpeed, tt.locationID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMCRRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
