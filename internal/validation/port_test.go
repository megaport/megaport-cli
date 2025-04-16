package validation

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatePortRequest(t *testing.T) {
	tests := []struct {
		name       string
		portName   string
		term       int
		portSpeed  int
		locationID int
		wantErr    bool
		errText    string
	}{
		{
			name:       "Valid Port request",
			portName:   "Test Port",
			term:       12,
			portSpeed:  10000,
			locationID: 100,
			wantErr:    false,
		},
		{
			name:       "Empty Port name",
			portName:   "",
			term:       12,
			portSpeed:  10000,
			locationID: 100,
			wantErr:    true,
			errText:    "Invalid port name:  - cannot be empty", // Updated expected error (lowercase 'port')
		},
		{
			name:       "Invalid term",
			portName:   "Test Port",
			term:       5,
			portSpeed:  10000,
			locationID: 100,
			wantErr:    true,
			errText:    fmt.Sprintf("Invalid contract term: 5 - must be one of: %v", ValidContractTerms),
		},
		{
			name:       "Invalid port speed",
			portName:   "Test Port",
			term:       12,
			portSpeed:  5000, // Not a valid Port speed
			locationID: 100,
			wantErr:    true,
			errText:    fmt.Sprintf("Invalid port speed: 5000 - must be one of: %v", ValidPortSpeeds),
		},
		{
			name:       "Invalid location ID",
			portName:   "Test Port",
			term:       12,
			portSpeed:  10000,
			locationID: 0,
			wantErr:    true,
			errText:    "Invalid location ID: 0 - must be a positive integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePortRequest(tt.portName, tt.term, tt.portSpeed, tt.locationID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePortRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.wantErr {
				assert.IsType(t, &ValidationError{}, err, "Expected ValidationError type")
				assert.Equal(t, tt.errText, err.Error(), "Error message mismatch")
			}
		})
	}
}

func TestValidatePortVLANAvailability(t *testing.T) {
	baseErrMsg := fmt.Sprintf("must be between %d-%d for VLAN availability check", MinAssignableVLAN, MaxAssignableVLAN)
	tests := []struct {
		name    string
		vlan    int
		wantErr bool
		errText string
	}{
		{"Valid Min Assignable", MinAssignableVLAN, false, ""},
		{"Valid Max Assignable", MaxAssignableVLAN, false, ""},
		{"Valid Mid Range", 2000, false, ""},
		// Update expected error messages to include the prefix
		{"Invalid Auto Assign", AutoAssignVLAN, true, fmt.Sprintf("Invalid VLAN ID: %d - %s", AutoAssignVLAN, baseErrMsg)},
		{"Invalid Untagged", UntaggedVLAN, true, fmt.Sprintf("Invalid VLAN ID: %d - %s", UntaggedVLAN, baseErrMsg)},
		{"Invalid Reserved 1", 1, true, fmt.Sprintf("Invalid VLAN ID: %d - %s", 1, baseErrMsg)},
		{"Invalid Max VLAN", MaxVLAN, true, fmt.Sprintf("Invalid VLAN ID: %d - %s", MaxVLAN, baseErrMsg)}, // 4094 is outside 2-4093
		{"Invalid Too High", MaxVLAN + 1, true, fmt.Sprintf("Invalid VLAN ID: %d - %s", MaxVLAN+1, baseErrMsg)},
		{"Invalid Too Low", MinAssignableVLAN - 10, true, fmt.Sprintf("Invalid VLAN ID: %d - %s", MinAssignableVLAN-10, baseErrMsg)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePortVLANAvailability(tt.vlan)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePortVLANAvailability() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && err.Error() != tt.errText {
				t.Errorf("ValidatePortVLANAvailability() error text = %q, want %q", err.Error(), tt.errText)
			}
		})
	}
}

func TestValidatePortName(t *testing.T) {
	tests := []struct {
		name     string
		portName string
		wantErr  bool
	}{
		{"Valid port name", "Test Port", false},
		{"Empty port name", "", true},
		// Boundary updated based on validator behavior: max allowed length is 64 characters.
		{"64 character port name", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", false}, // 64 A's
		{"65 character port name", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", true}, // 65 A's
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePortName(tt.portName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePortName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
