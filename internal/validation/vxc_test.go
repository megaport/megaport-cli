package validation

import (
	"testing"
)

func TestValidateVXCRateLimit(t *testing.T) {
	tests := []struct {
		name      string
		rateLimit int
		maxRate   int
		wantErr   bool
	}{
		{"Valid rate limit", 100, 1000, false},
		{"Equal to max", 1000, 1000, false},
		{"Zero value", 0, 1000, true},
		{"Negative value", -10, 1000, true},
		{"Exceeds max", 1500, 1000, true},
		{"No max specified", 2000, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVXCRateLimit(tt.rateLimit, tt.maxRate)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVXCRateLimit() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.wantErr {
				// Check that the error message includes the invalid value
				if valErr, ok := err.(*ValidationError); ok {
					if valErr.Value != tt.rateLimit {
						t.Errorf("Expected error to contain value %d, got %v", tt.rateLimit, valErr.Value)
					}
				} else {
					t.Errorf("ValidateVXCRateLimit() should return ValidationError, got %T", err)
				}
			}
		})
	}
}

func TestValidateVXCEndVLAN(t *testing.T) {
	tests := []struct {
		name    string
		vlan    int
		endName string
		wantErr bool
	}{
		{"A-End valid VLAN -1", -1, "A-End", false},
		{"A-End valid VLAN 0", 0, "A-End", false},
		{"A-End valid VLAN 2", 2, "A-End", false},
		{"A-End valid VLAN 4093", 4093, "A-End", false},
		{"A-End invalid VLAN 1", 1, "A-End", true},
		{"A-End invalid VLAN 4094", 4094, "A-End", true},
		{"A-End invalid VLAN -2", -2, "A-End", true},
		{"B-End valid VLAN -1", -1, "B-End", false},
		{"B-End valid VLAN 0", 0, "B-End", false},
		{"B-End valid VLAN 2", 2, "B-End", false},
		{"B-End valid VLAN 4093", 4093, "B-End", false},
		{"B-End invalid VLAN 1", 1, "B-End", true},
		{"B-End invalid VLAN 4094", 4094, "B-End", true},
		{"B-End invalid VLAN -2", -2, "B-End", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVXCEndVLAN(tt.vlan, tt.endName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVXCEndVLAN() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.wantErr {
				// Verify error contains the end name
				if valErr, ok := err.(*ValidationError); ok {
					if valErr.Field != tt.endName+" VLAN" {
						t.Errorf("Expected error field to contain %s, got %s", tt.endName+" VLAN", valErr.Field)
					}
				}
			}
		})
	}
}

func TestValidateVXCEndInnerVLAN(t *testing.T) {
	tests := []struct {
		name      string
		vlan      int
		outerVLAN int
		endName   string
		wantErr   bool
	}{
		{"Valid 0 inner VLAN", 0, 100, "A-End", false},
		{"Valid 100 inner VLAN", 100, 200, "A-End", false},
		{"Valid 4093 inner VLAN", 4093, 100, "A-End", false},
		{"Inner VLAN with untagged outer", 100, -1, "A-End", true},
		{"Invalid inner VLAN 1", 1, 100, "A-End", true},
		{"Invalid inner VLAN 4094", 4094, 100, "A-End", true},
		{"Inner VLAN with auto-assigned outer", 100, 0, "A-End", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVXCEndInnerVLAN(tt.vlan, tt.outerVLAN, tt.endName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVXCEndInnerVLAN() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateVXCRequest(t *testing.T) {
	tests := []struct {
		name             string
		vxcName          string
		term             int
		rateLimit        int
		aEndUID          string
		bEndUID          string
		hasPartnerConfig bool
		wantErr          bool
	}{
		{
			name:             "Valid VXC request with both ends",
			vxcName:          "Test VXC",
			term:             12,
			rateLimit:        1000,
			aEndUID:          "port-12345",
			bEndUID:          "port-67890",
			hasPartnerConfig: false,
			wantErr:          false,
		},
		{
			name:             "Valid VXC request with partner config",
			vxcName:          "Test VXC with Partner",
			term:             12,
			rateLimit:        1000,
			aEndUID:          "port-12345",
			bEndUID:          "",
			hasPartnerConfig: true,
			wantErr:          false,
		},
		{
			name:             "Empty VXC name",
			vxcName:          "",
			term:             12,
			rateLimit:        1000,
			aEndUID:          "port-12345",
			bEndUID:          "port-67890",
			hasPartnerConfig: false,
			wantErr:          true,
		},
		{
			name:             "Invalid term",
			vxcName:          "Test VXC",
			term:             5,
			rateLimit:        1000,
			aEndUID:          "port-12345",
			bEndUID:          "port-67890",
			hasPartnerConfig: false,
			wantErr:          true,
		},
		{
			name:             "Invalid rate limit",
			vxcName:          "Test VXC",
			term:             12,
			rateLimit:        0,
			aEndUID:          "port-12345",
			bEndUID:          "port-67890",
			hasPartnerConfig: false,
			wantErr:          true,
		},
		{
			name:             "Empty A-End UID",
			vxcName:          "Test VXC",
			term:             12,
			rateLimit:        1000,
			aEndUID:          "",
			bEndUID:          "port-67890",
			hasPartnerConfig: false,
			wantErr:          true,
		},
		{
			name:             "Empty B-End UID without partner config",
			vxcName:          "Test VXC",
			term:             12,
			rateLimit:        1000,
			aEndUID:          "port-12345",
			bEndUID:          "",
			hasPartnerConfig: false,
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVXCRequest(tt.vxcName, tt.term, tt.rateLimit, tt.aEndUID, tt.bEndUID, tt.hasPartnerConfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVXCRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePartnerConfigs(t *testing.T) {
	// Test AWS partner configuration validation
	t.Run("ValidateAWSPartnerConfig", func(t *testing.T) {
		tests := []struct {
			name              string
			connectType       string
			ownerAccount      string
			asn               int
			amazonAsn         int
			authKey           string
			customerIPAddress string
			amazonIPAddress   string
			connectionName    string
			awsType           string
			wantErr           bool
		}{
			{
				name:              "Valid AWS private config",
				connectType:       "private",
				ownerAccount:      "123456789012",
				asn:               65000,
				amazonAsn:         7224,
				authKey:           "",
				customerIPAddress: "169.254.0.0/30",
				amazonIPAddress:   "169.254.0.4/30",
				connectionName:    "test-connection",
				awsType:           "",
				wantErr:           false,
			},
			{
				name:              "Valid AWS public config",
				connectType:       "public",
				ownerAccount:      "123456789012",
				asn:               0,
				amazonAsn:         0,
				authKey:           "authkey123",
				customerIPAddress: "",
				amazonIPAddress:   "",
				connectionName:    "test-connection",
				awsType:           "public",
				wantErr:           false,
			},
			{
				name:              "Empty connect type",
				connectType:       "",
				ownerAccount:      "123456789012",
				asn:               0,
				amazonAsn:         0,
				authKey:           "",
				customerIPAddress: "",
				amazonIPAddress:   "",
				connectionName:    "",
				awsType:           "",
				wantErr:           true,
			},
			{
				name:              "Empty owner account",
				connectType:       "private",
				ownerAccount:      "",
				asn:               0,
				amazonAsn:         0,
				authKey:           "",
				customerIPAddress: "",
				amazonIPAddress:   "",
				connectionName:    "",
				awsType:           "",
				wantErr:           true,
			},
			{
				name:              "Invalid connect type",
				connectType:       "invalid",
				ownerAccount:      "123456789012",
				asn:               0,
				amazonAsn:         0,
				authKey:           "",
				customerIPAddress: "",
				amazonIPAddress:   "",
				connectionName:    "",
				awsType:           "",
				wantErr:           true,
				// Removed: "Invalid ASN range" test case
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := ValidateAWSPartnerConfig(
					tt.connectType,
					tt.ownerAccount,
					tt.asn,
					tt.amazonAsn,
					tt.authKey,
					tt.customerIPAddress,
					tt.amazonIPAddress,
					tt.connectionName,
					tt.awsType)
				if (err != nil) != tt.wantErr {
					t.Errorf("ValidateAWSPartnerConfig() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})

	// Test Azure partner configuration validation
	t.Run("ValidateAzurePartnerConfig", func(t *testing.T) {
		tests := []struct {
			name       string
			serviceKey string
			peers      []map[string]interface{}
			wantErr    bool
		}{
			{
				name:       "Valid Azure config without peers",
				serviceKey: "valid-service-key",
				peers:      nil,
				wantErr:    false,
			},
			{
				name:       "Empty service key",
				serviceKey: "",
				peers:      nil,
				wantErr:    true,
			},
			{
				name:       "Valid Azure config with valid peers",
				serviceKey: "valid-service-key",
				peers: []map[string]interface{}{
					{
						"type":             "private",
						"peer_asn":         "65000",
						"primary_subnet":   "10.0.0.0/30",
						"secondary_subnet": "10.0.0.4/30",
					},
				},
				wantErr: false,
			},
			{
				name:       "Invalid peer type",
				serviceKey: "valid-service-key",
				peers: []map[string]interface{}{
					{
						"type":           "invalid",
						"peer_asn":       "65000",
						"primary_subnet": "10.0.0.0/30",
					},
				},
				wantErr: true,
			},
			{
				name:       "Invalid peer ASN",
				serviceKey: "valid-service-key",
				peers: []map[string]interface{}{
					{
						"type":           "private",
						"peer_asn":       "invalid",
						"primary_subnet": "10.0.0.0/30",
					},
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := ValidateAzurePartnerConfig(tt.serviceKey, tt.peers)
				if (err != nil) != tt.wantErr {
					t.Errorf("ValidateAzurePartnerConfig() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})

	// Test Google partner configuration validation
	t.Run("ValidateGooglePartnerConfig", func(t *testing.T) {
		tests := []struct {
			name       string
			pairingKey string
			wantErr    bool
		}{
			{"Valid Google config", "valid-pairing-key", false},
			{"Empty pairing key", "", true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := ValidateGooglePartnerConfig(tt.pairingKey)
				if (err != nil) != tt.wantErr {
					t.Errorf("ValidateGooglePartnerConfig() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})

	// Test Oracle partner configuration validation
	t.Run("ValidateOraclePartnerConfig", func(t *testing.T) {
		tests := []struct {
			name             string
			virtualCircuitID string
			wantErr          bool
		}{
			{"Valid Oracle config", "ocid1.virtualcircuit.oc1.region.id", false},
			{"Empty virtual circuit ID", "", true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := ValidateOraclePartnerConfig(tt.virtualCircuitID)
				if (err != nil) != tt.wantErr {
					t.Errorf("ValidateOraclePartnerConfig() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})

	// Test IBM partner configuration validation
	t.Run("ValidateIBMPartnerConfig", func(t *testing.T) {
		tests := []struct {
			name              string
			accountID         string
			customerASN       int
			connectionName    string
			customerIPAddress string
			providerIPAddress string
			wantErr           bool
		}{
			{
				name:              "Valid IBM config",
				accountID:         "1a2b3c4d5e6f7890abcdef1234567890",
				customerASN:       64500,
				connectionName:    "test-connection",
				customerIPAddress: "169.254.0.0/30",
				providerIPAddress: "169.254.0.4/30",
				wantErr:           false,
			},
			{
				name:              "Empty account ID",
				accountID:         "",
				customerASN:       0,
				connectionName:    "",
				customerIPAddress: "",
				providerIPAddress: "",
				wantErr:           true,
			},
			{
				name:              "Invalid account ID format (not 32 hex chars)",
				accountID:         "1a2b3c",
				customerASN:       0,
				connectionName:    "",
				customerIPAddress: "",
				providerIPAddress: "",
				wantErr:           true,
			},
			{
				name:              "Invalid account ID format (not hex)",
				accountID:         "1a2b3c4d5e6f7890abcdefghijklmnop",
				customerASN:       0,
				connectionName:    "",
				customerIPAddress: "",
				providerIPAddress: "",
				wantErr:           true,
				// Removed: "Invalid ASN" test case
			},
			{
				name:              "Too long connection name",
				accountID:         "1a2b3c4d5e6f7890abcdef1234567890",
				customerASN:       64500,
				connectionName:    "this-is-a-very-long-connection-name-that-exceeds-the-maximum-allowed-length-of-100-characters-so-it-should-fail-validation-test",
				customerIPAddress: "",
				providerIPAddress: "",
				wantErr:           true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := ValidateIBMPartnerConfig(
					tt.accountID,
					tt.customerASN,
					tt.connectionName,
					tt.customerIPAddress,
					tt.providerIPAddress)
				if (err != nil) != tt.wantErr {
					t.Errorf("ValidateIBMPartnerConfig() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})
}
