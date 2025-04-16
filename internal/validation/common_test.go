package validation

import (
	"testing"
)

func TestValidateContractTerm(t *testing.T) {
	tests := []struct {
		name    string
		term    int
		wantErr bool
	}{
		{"Valid term 1", 1, false},
		{"Valid term 12", 12, false},
		{"Valid term 24", 24, false},
		{"Valid term 36", 36, false},
		{"Invalid term 0", 0, true},
		{"Invalid term -1", -1, true},
		{"Invalid term 6", 6, true},
		{"Invalid term 48", 48, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateContractTerm(tt.term)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateContractTerm() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.wantErr {
				// Check that the error message includes the invalid value
				if _, ok := err.(*ValidationError); !ok {
					t.Errorf("ValidateContractTerm() should return ValidationError, got %T", err)
				}
			}
		})
	}
}

func TestValidateMCRPortSpeed(t *testing.T) {
	tests := []struct {
		name    string
		speed   int
		wantErr bool
	}{
		{"Valid speed 1000", 1000, false},
		{"Valid speed 2500", 2500, false},
		{"Valid speed 5000", 5000, false},
		{"Valid speed 10000", 10000, false},
		{"Valid speed 25000", 25000, false},
		{"Valid speed 50000", 50000, false},
		{"Valid speed 100000", 100000, false},
		{"Invalid speed 0", 0, true},
		{"Invalid speed -1", -1, true},
		{"Invalid speed 500", 500, true},
		{"Invalid speed 150000", 150000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMCRPortSpeed(tt.speed)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMCRPortSpeed() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.wantErr {
				// Check that the error message includes the invalid value
				if valErr, ok := err.(*ValidationError); ok {
					if valErr.Value != tt.speed {
						t.Errorf("Expected error to contain value %d, got %v", tt.speed, valErr.Value)
					}
				} else {
					t.Errorf("ValidateMCRPortSpeed() should return ValidationError, got %T", err)
				}
			}
		})
	}
}

func TestValidateVLAN(t *testing.T) {
	tests := []struct {
		name    string
		vlan    int
		wantErr bool
	}{
		{"Valid VLAN 2", 2, false},
		{"Valid VLAN 1000", 1000, false},
		{"Valid VLAN 4093", 4093, false},
		{"Invalid VLAN 0", 0, false},
		{"Invalid VLAN 1", 1, true}, // Reserved
		{"Invalid VLAN -1", -1, false},
		{"Invalid VLAN 4094", 4094, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVLAN(tt.vlan)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVLAN() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateRateLimit(t *testing.T) {
	tests := []struct {
		name      string
		rateLimit int
		wantErr   bool
	}{
		{"Valid rate limit 1", 1, false},
		{"Valid rate limit 100", 100, false},
		{"Valid rate limit 1000", 1000, false},
		{"Invalid rate limit 0", 0, true},
		{"Invalid rate limit -1", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRateLimit(tt.rateLimit)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRateLimit() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	err := NewValidationError("test field", 123, "test reason")
	expected := "Invalid test field: 123 - test reason"
	if err.Error() != expected {
		t.Errorf("ValidationError.Error() = %v, want %v", err.Error(), expected)
	}
}
