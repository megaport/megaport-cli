package validation

import (
	"testing"
)

func TestValidatePortRequest(t *testing.T) {
	tests := []struct {
		name       string
		portName   string
		term       int
		portSpeed  int
		locationID int
		wantErr    bool
	}{
		{
			name:       "Valid port request",
			portName:   "Test Port",
			term:       12,
			portSpeed:  1000,
			locationID: 100,
			wantErr:    false,
		},
		{
			name:       "Empty port name",
			portName:   "",
			term:       12,
			portSpeed:  1000,
			locationID: 100,
			wantErr:    true,
		},
		{
			name:       "Invalid term",
			portName:   "Test Port",
			term:       5,
			portSpeed:  1000,
			locationID: 100,
			wantErr:    true,
		},
		{
			name:       "Invalid port speed",
			portName:   "Test Port",
			term:       12,
			portSpeed:  500,
			locationID: 100,
			wantErr:    true,
		},
		{
			name:       "Invalid location ID",
			portName:   "Test Port",
			term:       12,
			portSpeed:  1000,
			locationID: 0,
			wantErr:    true,
		},
		{
			name:       "Negative location ID",
			portName:   "Test Port",
			term:       12,
			portSpeed:  1000,
			locationID: -1,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePortRequest(tt.portName, tt.term, tt.portSpeed, tt.locationID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePortRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePortVLANAvailability(t *testing.T) {
	tests := []struct {
		name    string
		vlan    int
		wantErr bool
	}{
		{"Valid VLAN 2", 2, false},
		{"Valid VLAN 1000", 1000, false},
		{"Valid VLAN 4093", 4093, false},
		{"Invalid VLAN 0", 0, true},
		{"Invalid VLAN 1", 1, true},
		{"Invalid VLAN -1", -1, true},
		{"Invalid VLAN 4094", 4094, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePortVLANAvailability(tt.vlan)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePortVLANAvailability() error = %v, wantErr %v", err, tt.wantErr)
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
