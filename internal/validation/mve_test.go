package validation

import (
	"testing"
)

func TestValidateMVERequest(t *testing.T) {
	tests := []struct {
		name        string
		productName string
		term        int
		locationID  int
		wantErr     bool
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
		},
		{
			name:        "Invalid term",
			productName: "Test MVE",
			term:        5, // Not in the valid set
			locationID:  123,
			wantErr:     true,
		},
		{
			name:        "Invalid location ID",
			productName: "Test MVE",
			term:        12,
			locationID:  0,
			wantErr:     true,
		},
		{
			name:        "Name too long",
			productName: "This name is way too long and should exceed the 64 character limit for MVE product names which will cause validation to fail",
			term:        12,
			locationID:  123,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMVERequest(tt.productName, tt.term, tt.locationID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMVERequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateMVEVendor(t *testing.T) {
	tests := []struct {
		name    string
		vendor  string
		wantErr bool
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
		},
		{
			name:    "Empty vendor",
			vendor:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMVEVendor(tt.vendor)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMVEVendor() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateMVENetworkInterfaces(t *testing.T) {
	tests := []struct {
		name    string
		vnics   []string
		wantErr bool
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
		},
		{
			name:    "Empty description",
			vnics:   []string{"Interface 1", ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMVENetworkInterfaces(tt.vnics)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMVENetworkInterfaces() error = %v, wantErr %v", err, tt.wantErr)
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
				"manage_locally":       false,
			},
			wantErr: true,
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
		},
		{
			name:   "Invalid vendor",
			vendor: "invalid_vendor",
			config: map[string]interface{}{
				"image_id":     123,
				"product_size": "MEDIUM",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMVEVendorConfig(tt.vendor, tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMVEVendorConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
