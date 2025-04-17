package validation

import (
	"fmt"
	"testing"

	megaport "github.com/megaport/megaportgo"
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
			errText:    "Invalid port name:  - cannot be empty",
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
			portSpeed:  5000,
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
			req := &megaport.BuyPortRequest{
				Name:       tt.portName,
				Term:       tt.term,
				PortSpeed:  tt.portSpeed,
				LocationId: tt.locationID,
			}
			err := ValidatePortRequest(req)
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
		{"Invalid Auto Assign", AutoAssignVLAN, true, fmt.Sprintf("Invalid VLAN ID: %d - %s", AutoAssignVLAN, baseErrMsg)},
		{"Invalid Untagged", UntaggedVLAN, true, fmt.Sprintf("Invalid VLAN ID: %d - %s", UntaggedVLAN, baseErrMsg)},
		{"Invalid Reserved 1", 1, true, fmt.Sprintf("Invalid VLAN ID: %d - %s", 1, baseErrMsg)},
		{"Invalid Max VLAN", MaxVLAN, true, fmt.Sprintf("Invalid VLAN ID: %d - %s", MaxVLAN, baseErrMsg)},
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
		{"64 character port name", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", false},
		{"65 character port name", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", true},
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
