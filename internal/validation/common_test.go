package validation

import (
	"fmt"
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
	baseErrMsg := fmt.Sprintf("must be %d, %d, or between %d-%d", AutoAssignVLAN, UntaggedVLAN, MinAssignableVLAN, MaxVLAN)
	tests := []struct {
		name    string
		vlan    int
		wantErr bool
		errText string
	}{
		{"Valid Auto Assign", AutoAssignVLAN, false, ""},
		{"Valid Untagged", UntaggedVLAN, false, ""},
		{"Valid Min Assignable", MinAssignableVLAN, false, ""},
		{"Valid Max Assignable", MaxAssignableVLAN, false, ""},
		{"Valid Max VLAN", MaxVLAN, false, ""},
		{"Valid Mid Range", 1000, false, ""},
		// Update expected error messages to include the prefix
		{"Invalid Too Low", -2, true, fmt.Sprintf("Invalid VLAN ID: %d - %s", -2, baseErrMsg)},
		{"Invalid Reserved 1", 1, true, fmt.Sprintf("Invalid VLAN ID: %d - %s", 1, baseErrMsg)},
		{"Invalid Too High", MaxVLAN + 1, true, fmt.Sprintf("Invalid VLAN ID: %d - %s", MaxVLAN+1, baseErrMsg)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVLAN(tt.vlan)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVLAN() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && err.Error() != tt.errText {
				t.Errorf("ValidateVLAN() error text = %q, want %q", err.Error(), tt.errText)
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
