package validation

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateContractTerm(t *testing.T) {
	tests := []struct {
		name    string
		term    int
		wantErr bool
		errText string
	}{
		{"Valid term 1", 1, false, ""},
		{"Valid term 12", 12, false, ""},
		{"Valid term 24", 24, false, ""},
		{"Valid term 36", 36, false, ""},
		{"Invalid term 0", 0, true, fmt.Sprintf("Invalid contract term: 0 - must be one of: %v", ValidContractTerms)},
		{"Invalid term -1", -1, true, fmt.Sprintf("Invalid contract term: -1 - must be one of: %v", ValidContractTerms)},
		{"Invalid term 6", 6, true, fmt.Sprintf("Invalid contract term: 6 - must be one of: %v", ValidContractTerms)},
		{"Invalid term 48", 48, true, fmt.Sprintf("Invalid contract term: 48 - must be one of: %v", ValidContractTerms)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateContractTerm(tt.term)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateContractTerm() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.wantErr {
				assert.IsType(t, &ValidationError{}, err, "Expected ValidationError type")
				assert.Equal(t, tt.errText, err.Error(), "Error message mismatch")
			}
		})
	}
}

func TestValidateMCRPortSpeed(t *testing.T) {
	tests := []struct {
		name    string
		speed   int
		wantErr bool
		errText string
	}{
		{"Valid speed 1000", 1000, false, ""},
		{"Valid speed 2500", 2500, false, ""},
		{"Valid speed 5000", 5000, false, ""},
		{"Valid speed 10000", 10000, false, ""},
		{"Valid speed 25000", 25000, false, ""},
		{"Valid speed 50000", 50000, false, ""},
		{"Valid speed 100000", 100000, false, ""},
		{"Invalid speed 0", 0, true, fmt.Sprintf("Invalid MCR port speed: 0 - must be one of: %v", ValidMCRPortSpeeds)},
		{"Invalid speed -1", -1, true, fmt.Sprintf("Invalid MCR port speed: -1 - must be one of: %v", ValidMCRPortSpeeds)},
		{"Invalid speed 500", 500, true, fmt.Sprintf("Invalid MCR port speed: 500 - must be one of: %v", ValidMCRPortSpeeds)},
		{"Invalid speed 150000", 150000, true, fmt.Sprintf("Invalid MCR port speed: 150000 - must be one of: %v", ValidMCRPortSpeeds)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMCRPortSpeed(tt.speed)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMCRPortSpeed() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.wantErr {
				assert.IsType(t, &ValidationError{}, err, "Expected ValidationError type")
				assert.Equal(t, tt.errText, err.Error(), "Error message mismatch")
			}
		})
	}
}

func TestValidatePortSpeed(t *testing.T) {
	tests := []struct {
		name    string
		speed   int
		wantErr bool
		errText string
	}{
		{"Valid speed 1000", 1000, false, ""},
		{"Valid speed 10000", 10000, false, ""},
		{"Valid speed 100000", 100000, false, ""},
		{"Invalid speed 0", 0, true, fmt.Sprintf("Invalid port speed: 0 - must be one of: %v", ValidPortSpeeds)},
		{"Invalid speed 5000", 5000, true, fmt.Sprintf("Invalid port speed: 5000 - must be one of: %v", ValidPortSpeeds)},
		{"Invalid speed -1000", -1000, true, fmt.Sprintf("Invalid port speed: -1000 - must be one of: %v", ValidPortSpeeds)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePortSpeed(tt.speed)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePortSpeed() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.wantErr {
				assert.IsType(t, &ValidationError{}, err, "Expected ValidationError type")
				assert.Equal(t, tt.errText, err.Error(), "Error message mismatch")
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
			if err != nil && tt.wantErr {
				assert.IsType(t, &ValidationError{}, err, "Expected ValidationError type")
				assert.Equal(t, tt.errText, err.Error(), "Error message mismatch")
			}
		})
	}
}

func TestValidateRateLimit(t *testing.T) {
	tests := []struct {
		name      string
		rateLimit int
		wantErr   bool
		errText   string
	}{
		{"Valid rate limit 1", 1, false, ""},
		{"Valid rate limit 100", 100, false, ""},
		{"Valid rate limit 1000", 1000, false, ""},
		{"Invalid rate limit 0", 0, true, "Invalid rate limit: 0 - must be a positive integer"},
		{"Invalid rate limit -1", -1, true, "Invalid rate limit: -1 - must be a positive integer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRateLimit(tt.rateLimit)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRateLimit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.wantErr {
				assert.IsType(t, &ValidationError{}, err, "Expected ValidationError type")
				assert.Equal(t, tt.errText, err.Error(), "Error message mismatch")
			}
		})
	}
}

func TestValidateMVEProductSize(t *testing.T) {
	tests := []struct {
		name    string
		size    string
		wantErr bool
		errText string
	}{
		{"Valid SMALL", "SMALL", false, ""},
		{"Valid MEDIUM", "MEDIUM", false, ""},
		{"Valid LARGE", "LARGE", false, ""},
		{"Invalid lowercase", "small", true, fmt.Sprintf("Invalid product size: small - must be one of: %v", ValidMVEProductSizes)},
		{"Invalid value", "XLARGE", true, fmt.Sprintf("Invalid product size: XLARGE - must be one of: %v", ValidMVEProductSizes)},
		{"Empty value", "", true, fmt.Sprintf("Invalid product size:  - must be one of: %v", ValidMVEProductSizes)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMVEProductSize(tt.size)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMVEProductSize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.wantErr {
				assert.IsType(t, &ValidationError{}, err, "Expected ValidationError type")
				assert.Equal(t, tt.errText, err.Error(), "Error message mismatch")
			}
		})
	}
}

func TestValidateDateRange(t *testing.T) {
	tests := []struct {
		name      string
		startDate string
		endDate   string
		wantErr   bool
		errText   string
	}{
		{"Both empty", "", "", false, ""},
		{"Both valid end after start", "2026-01-01", "2026-06-01", false, ""},
		{"Only start provided", "2026-01-01", "", true, "Invalid date range: end-date - both --start-date and --end-date must be provided together"},
		{"Only end provided", "", "2026-06-01", true, "Invalid date range: start-date - both --start-date and --end-date must be provided together"},
		{"Invalid start format", "01-01-2026", "2026-06-01", true, "Invalid start-date: 01-01-2026 - must be in YYYY-MM-DD format"},
		{"Invalid end format", "2026-01-01", "01-06-2026", true, "Invalid end-date: 01-06-2026 - must be in YYYY-MM-DD format"},
		{"End before start", "2026-06-01", "2026-01-01", true, "Invalid date range: 2026-06-01 to 2026-01-01 - end date must be after start date"},
		{"Same date", "2026-06-01", "2026-06-01", true, "Invalid date range: 2026-06-01 to 2026-06-01 - end date must be after start date"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDateRange(tt.startDate, tt.endDate)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDateRange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.wantErr {
				assert.IsType(t, &ValidationError{}, err, "Expected ValidationError type")
				assert.Equal(t, tt.errText, err.Error(), "Error message mismatch")
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
	assert.IsType(t, &ValidationError{}, err, "Expected ValidationError type")
}

func TestFormatIntSlice(t *testing.T) {
	tests := []struct {
		name string
		vals []int
		want string
	}{
		{"empty", nil, ""},
		{"single", []int{1}, "1"},
		{"two", []int{1, 12}, "1 or 12"},
		{"three", []int{1, 12, 24}, "1, 12, or 24"},
		{"four", []int{1, 12, 24, 36}, "1, 12, 24, or 36"},
		{"port speeds", []int{1000, 10000, 100000}, "1000, 10000, or 100000"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, FormatIntSlice(tt.vals))
		})
	}
}

func TestVLANHelpText(t *testing.T) {
	assert.Equal(t, "0=auto-assign, -1=untagged, 2-4094 for specific VLAN (1 is reserved)", VLANHelpText())
}

func TestInnerVLANHelpText(t *testing.T) {
	assert.Equal(t, "0=none, -1=untagged, 2-4094 for specific VLAN (1 is reserved)", InnerVLANHelpText())
}
