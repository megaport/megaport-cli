package validation

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateMVERequest(t *testing.T) {
	tests := []struct {
		name        string
		productName string
		term        int
		locationID  int
		wantErr     bool
		errText     string
	}{
		{
			name:        "Valid MVE request",
			productName: "Test MVE",
			term:        12,
			locationID:  123,
			wantErr:     false,
		},
		{
			name:        "Empty name",
			productName: "",
			term:        12,
			locationID:  123,
			wantErr:     true,
			errText:     "Invalid MVE name:  - cannot be empty",
		},
		{
			name:        "Invalid term",
			productName: "Test MVE",
			term:        5, // Not in the valid set
			locationID:  123,
			wantErr:     true,
			errText:     fmt.Sprintf("Invalid contract term: 5 - must be one of: %v", ValidContractTerms),
		},
		{
			name:        "Invalid location ID",
			productName: "Test MVE",
			term:        12,
			locationID:  0,
			wantErr:     true,
			errText:     "Invalid location ID: 0 - must be a positive integer",
		},
		{
			name:        "Name too long",
			productName: "This name is way too long and should exceed the 64 character limit for MVE product names which will cause validation to fail",
			term:        12,
			locationID:  123,
			wantErr:     true,
			errText:     "Invalid MVE name: This name is way too long and should exceed the 64 character limit for MVE product names which will cause validation to fail - cannot exceed 64 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMVERequest(tt.productName, tt.term, tt.locationID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMVERequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.wantErr {
				assert.IsType(t, &ValidationError{}, err, "Expected ValidationError type")
				assert.Equal(t, tt.errText, err.Error(), "Error message mismatch")
			}
		})
	}
}

func TestValidateMVEVendor(t *testing.T) {
	tests := []struct {
		name    string
		vendor  string
		wantErr bool
		errText string
	}{
		{
			name:    "Valid vendor lowercase",
			vendor:  "cisco",
			wantErr: false,
		},
		{
			name:    "Valid vendor uppercase",
			vendor:  "CISCO",
			wantErr: false,
		},
		{
			name:    "Valid vendor mixed case",
			vendor:  "CiScO",
			wantErr: false,
		},
		{
			name:    "Invalid vendor",
			vendor:  "invalid_vendor",
			wantErr: true,
			errText: fmt.Sprintf("Invalid MVE vendor: invalid_vendor - must be one of: %v", ValidMVEVendors),
		},
		{
			name:    "Empty vendor",
			vendor:  "",
			wantErr: true,
			errText: fmt.Sprintf("Invalid MVE vendor:  - must be one of: %v", ValidMVEVendors), // Adjusted for empty value check if ValidateStringOneOf handles it
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMVEVendor(tt.vendor)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMVEVendor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.wantErr {
				assert.IsType(t, &ValidationError{}, err, "Expected ValidationError type")
				assert.Equal(t, tt.errText, err.Error(), "Error message mismatch")
			}
		})
	}
}

func TestValidateMVENetworkInterfaces(t *testing.T) {
	tests := []struct {
		name    string
		vnics   []string
		wantErr bool
		errText string
	}{
		{
			name:    "Valid single vNIC",
			vnics:   []string{"Interface 1"},
			wantErr: false,
		},
		{
			name:    "Valid multiple vNICs",
			vnics:   []string{"Interface 1", "Interface 2", "Interface 3"},
			wantErr: false,
		},
		{
			name:    "Maximum vNICs",
			vnics:   []string{"Interface 1", "Interface 2", "Interface 3", "Interface 4", "Interface 5"},
			wantErr: false,
		},
		{
			name:    "Too many vNICs",
			vnics:   []string{"Interface 1", "Interface 2", "Interface 3", "Interface 4", "Interface 5", "Interface 6"},
			wantErr: true,
			errText: "Invalid network interfaces: 6 - cannot exceed 5 vNICs",
		},
		{
			name:    "Empty description",
			vnics:   []string{"Interface 1", ""},
			wantErr: true,
			errText: "Invalid network interface 2:  - Invalid network interface description:  - cannot be empty", // Updated expected error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMVENetworkInterfaces(tt.vnics)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMVENetworkInterfaces() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.wantErr {
				assert.IsType(t, &ValidationError{}, err, "Expected ValidationError type")
				assert.Equal(t, tt.errText, err.Error(), "Error message mismatch")
			}
		})
	}
}

func TestValidateMVEVendorConfig(t *testing.T) {
	// Test cases for ValidateMVEVendorConfig
	tests := []struct {
		name    string
		vendor  string
		config  map[string]interface{}
		wantErr bool
		errText string
	}{
		{
			name:   "Valid 6wind config",
			vendor: "6wind",
			config: map[string]interface{}{
				"image_id":       123,
				"product_size":   "MEDIUM",
				"mve_label":      "6wind-mve",
				"ssh_public_key": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC....",
			},
			wantErr: false,
		},
		{
			name:   "Invalid 6wind config - missing ssh key",
			vendor: "6wind",
			config: map[string]interface{}{
				"image_id":     123,
				"product_size": "MEDIUM",
				"mve_label":    "6wind-mve",
			},
			wantErr: true,
			errText: "Invalid SSH public key:  - cannot be empty",
		},
		{
			name:   "Valid Cisco config",
			vendor: "cisco",
			config: map[string]interface{}{
				"image_id":             123,
				"product_size":         "MEDIUM",
				"mve_label":            "cisco-mve",
				"admin_ssh_public_key": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC....",
				"ssh_public_key":       "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC....",
				"cloud_init":           "base64_encoded_data",
				"manage_locally":       true,
			},
			wantErr: false,
		},
		{
			name:   "Invalid Cisco config - missing FMC data when not managing locally",
			vendor: "cisco",
			config: map[string]interface{}{
				"image_id":             123,
				"product_size":         "MEDIUM",
				"mve_label":            "cisco-mve",
				"admin_ssh_public_key": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC....",
				"ssh_public_key":       "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC....",
				"cloud_init":           "base64_encoded_data",
				"manage_locally":       false, // Requires FMC fields
			},
			wantErr: true,
			errText: "Invalid FMC IP address:  - cannot be empty when not managing locally",
		},
		{
			name:   "Invalid product size",
			vendor: "cisco",
			config: map[string]interface{}{
				"image_id":             123,
				"product_size":         "INVALID_SIZE",
				"mve_label":            "cisco-mve",
				"admin_ssh_public_key": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC....",
				"ssh_public_key":       "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC....",
			},
			wantErr: true,
			errText: fmt.Sprintf("Invalid MVE product size: INVALID_SIZE - must be one of: %v", ValidMVEProductSizes),
		},
		{
			name:   "Invalid vendor",
			vendor: "invalid_vendor",
			config: map[string]interface{}{
				"image_id":     123,
				"product_size": "MEDIUM",
			},
			wantErr: true,
			errText: fmt.Sprintf("Invalid MVE vendor: invalid_vendor - must be one of: %v", ValidMVEVendors),
		},
		{
			name:   "Missing image ID",
			vendor: "cisco",
			config: map[string]interface{}{
				// image_id missing
				"product_size": "MEDIUM",
			},
			wantErr: true,
			errText: "Invalid image ID: <nil> - must be a valid integer",
		},
		{
			name:   "Invalid image ID type",
			vendor: "cisco",
			config: map[string]interface{}{
				"image_id":     "not-an-int",
				"product_size": "MEDIUM",
			},
			wantErr: true,
			errText: "Invalid image ID: not-an-int - must be a valid integer",
		},
		{
			name:   "Missing product size",
			vendor: "cisco",
			config: map[string]interface{}{
				"image_id": 123,
				// product_size missing
			},
			wantErr: true,
			errText: "Invalid product size: <nil> - must be a valid string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMVEVendorConfig(tt.vendor, tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMVEVendorConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.wantErr {
				assert.IsType(t, &ValidationError{}, err, "Expected ValidationError type")
				assert.Equal(t, tt.errText, err.Error(), "Error message mismatch")
			}
		})
	}
}
