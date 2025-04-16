package validation

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateVXCEndVLAN(t *testing.T) {
	tests := []struct {
		name    string
		vlan    int
		wantErr bool
	}{
		{"Valid VLAN 100", 100, false},
		{"Valid Min Assignable VLAN", MinAssignableVLAN, false},
		{"Valid Max VLAN", MaxVLAN, false},
		{"Valid Untagged", UntaggedVLAN, false},
		{"Valid Auto Assign", AutoAssignVLAN, false},
		{"Invalid VLAN 1", 1, true}, // Reserved VLAN
		{"Invalid VLAN Too Low", -2, true},
		{"Invalid VLAN Too High", MaxVLAN + 1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVXCEndVLAN(tt.vlan)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVXCEndVLAN() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.wantErr {
				assert.IsType(t, &ValidationError{}, err, "Expected ValidationError type")
			}
		})
	}
}

func TestValidateVXCEndInnerVLAN(t *testing.T) {
	// Inner VLAN validation uses the same logic as outer VLAN
	tests := []struct {
		name    string
		vlan    int
		wantErr bool
	}{
		{"Valid Inner VLAN 200", 200, false},
		{"Valid Inner Min Assignable VLAN", MinAssignableVLAN, false},
		{"Valid Inner Max VLAN", MaxVLAN, false},
		{"Valid Inner Untagged", UntaggedVLAN, false},
		{"Valid Inner Auto Assign", AutoAssignVLAN, false},
		{"Invalid Inner VLAN 1", 1, true}, // Reserved VLAN
		{"Invalid Inner VLAN Too Low", -2, true},
		{"Invalid Inner VLAN Too High", MaxVLAN + 1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVXCEndInnerVLAN(tt.vlan)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVXCEndInnerVLAN() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.wantErr {
				assert.IsType(t, &ValidationError{}, err, "Expected ValidationError type")
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
		errText          string
	}{
		{
			name:             "Valid VXC request",
			vxcName:          "Test VXC",
			term:             12,
			rateLimit:        1000,
			aEndUID:          "a-end-uid",
			bEndUID:          "b-end-uid",
			hasPartnerConfig: false,
			wantErr:          false,
		},
		{
			name:             "Valid VXC request with partner config",
			vxcName:          "Test VXC Partner",
			term:             1,
			rateLimit:        500,
			aEndUID:          "a-end-uid",
			bEndUID:          "", // bEndUID not required if partner config exists
			hasPartnerConfig: true,
			wantErr:          false,
		},
		{
			name:             "Empty name",
			vxcName:          "",
			term:             12,
			rateLimit:        1000,
			aEndUID:          "a-end-uid",
			bEndUID:          "b-end-uid",
			hasPartnerConfig: false,
			wantErr:          true,
			errText:          "Invalid VXC name:  - cannot be empty",
		},
		{
			name:             "Invalid term",
			vxcName:          "Test VXC",
			term:             5, // Invalid term
			rateLimit:        1000,
			aEndUID:          "a-end-uid",
			bEndUID:          "b-end-uid",
			hasPartnerConfig: false,
			wantErr:          true,
			errText:          fmt.Sprintf("Invalid contract term: 5 - must be one of: %v", ValidContractTerms),
		},
		{
			name:             "Invalid rate limit",
			vxcName:          "Test VXC",
			term:             12,
			rateLimit:        0, // Invalid rate limit
			aEndUID:          "a-end-uid",
			bEndUID:          "b-end-uid",
			hasPartnerConfig: false,
			wantErr:          true,
			errText:          "Invalid rate limit: 0 - must be a positive integer",
		},
		{
			name:             "Empty A-End UID",
			vxcName:          "Test VXC",
			term:             12,
			rateLimit:        1000,
			aEndUID:          "", // Empty A-End UID
			bEndUID:          "b-end-uid",
			hasPartnerConfig: false,
			wantErr:          true,
			errText:          "Invalid A-End UID:  - cannot be empty",
		},
		{
			name:             "Empty B-End UID without partner config",
			vxcName:          "Test VXC",
			term:             12,
			rateLimit:        1000,
			aEndUID:          "a-end-uid",
			bEndUID:          "", // Empty B-End UID
			hasPartnerConfig: false,
			wantErr:          true,
			errText:          "Invalid B-End UID:  - cannot be empty when no partner configuration is provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVXCRequest(tt.vxcName, tt.term, tt.rateLimit, tt.aEndUID, tt.bEndUID, tt.hasPartnerConfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVXCRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.wantErr {
				assert.IsType(t, &ValidationError{}, err, "Expected ValidationError type")
				assert.Equal(t, tt.errText, err.Error(), "Error message mismatch")
			}
		})
	}
}

func TestValidateAWSPartnerConfig(t *testing.T) {
	tests := []struct {
		name              string
		connectType       string
		ownerAccount      string
		asn               int
		amazonAsn         int
		authKey           string
		customerIPAddress string
		amazonIPAddress   string
		awsName           string
		awsType           string
		wantErr           bool
		errText           string
	}{
		{
			name:              "Valid AWS config",
			connectType:       "AWS",
			ownerAccount:      "123456789012",
			asn:               65000,
			amazonAsn:         64512,
			authKey:           "authkey123",
			customerIPAddress: "192.168.1.1/30",
			amazonIPAddress:   "192.168.1.2/30",
			awsName:           "MyAWSConnection",
			awsType:           "private",
			wantErr:           false,
		},
		{
			name:              "Valid AWSHC config",
			connectType:       "AWSHC",
			ownerAccount:      "123456789012",
			asn:               65000,
			amazonAsn:         64512,
			authKey:           "authkey123",
			customerIPAddress: "192.168.1.1/30",
			amazonIPAddress:   "192.168.1.2/30",
			awsName:           "MyAWSHCConnection",
			wantErr:           false,
		},
		{
			name:         "Empty connect type",
			connectType:  "",
			ownerAccount: "123456789012",
			wantErr:      true,
			errText:      "Invalid AWS connect type:  - cannot be empty",
		},
		{
			name:         "Invalid connect type",
			connectType:  "INVALID",
			ownerAccount: "123456789012",
			wantErr:      true,
			errText:      "Invalid AWS connect type: INVALID - must be 'AWS', 'AWSHC', 'private', or 'public'",
		},
		{
			name:         "Empty owner account",
			connectType:  "AWS",
			ownerAccount: "",
			wantErr:      true,
			errText:      "Invalid AWS owner account:  - cannot be empty",
		},
		{
			name:              "Invalid customer IP CIDR",
			connectType:       "AWS",
			ownerAccount:      "123456789012",
			customerIPAddress: "invalid-ip",
			wantErr:           true,
			errText:           "Invalid AWS customer IP address: invalid-ip - must be a valid IPv4 CIDR",
		},
		{
			name:            "Invalid Amazon IP CIDR",
			connectType:     "AWS",
			ownerAccount:    "123456789012",
			amazonIPAddress: "192.168.1.2/33", // Invalid mask
			wantErr:         true,
			errText:         "Invalid AWS Amazon IP address: 192.168.1.2/33 - must be a valid IPv4 CIDR",
		},
		{
			name:         "AWS name too long",
			connectType:  "AWS",
			ownerAccount: "123456789012",
			awsName:      string(make([]byte, 256)), // 256 chars
			wantErr:      true,
			errText:      "Invalid AWS connection name: ", // Error message includes the long name, truncated here
		},
		{
			name:         "Invalid AWS type for AWS connect type",
			connectType:  "AWS",
			ownerAccount: "123456789012",
			awsType:      "invalid",
			wantErr:      true,
			errText:      "Invalid AWS type: invalid - must be 'private' or 'public' for AWS connect type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAWSPartnerConfig(tt.connectType, tt.ownerAccount, tt.asn, tt.amazonAsn, tt.authKey, tt.customerIPAddress, tt.amazonIPAddress, tt.awsName, tt.awsType)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAWSPartnerConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.wantErr {
				assert.IsType(t, &ValidationError{}, err, "Expected ValidationError type")
				// Use Contains because the long name error message is hard to match exactly
				assert.Contains(t, err.Error(), tt.errText, "Error message mismatch")
			}
		})
	}
}

func TestValidateAzurePartnerConfig(t *testing.T) {
	tests := []struct {
		name       string
		serviceKey string
		peers      []map[string]interface{}
		wantErr    bool
		errText    string
	}{
		{
			name:       "Valid Azure config",
			serviceKey: "azure-service-key",
			peers: []map[string]interface{}{
				{
					"type":           "private",
					"peer_asn":       "65001",
					"primary_subnet": "10.0.0.0/30",
				},
			},
			wantErr: false,
		},
		{
			name:       "Valid Azure config with Microsoft peering",
			serviceKey: "azure-service-key",
			peers: []map[string]interface{}{
				{
					"type": "microsoft",
				},
			},
			wantErr: false,
		},
		{
			name:       "Empty service key",
			serviceKey: "",
			wantErr:    true,
			errText:    "Invalid Azure service key:  - cannot be empty",
		},
		{
			name:       "Invalid peer type",
			serviceKey: "azure-service-key",
			peers: []map[string]interface{}{
				{
					"type": "invalid",
				},
			},
			wantErr: true,
			errText: "Invalid Azure peer [0] type: invalid - must be 'private' or 'microsoft'",
		},
		{
			name:       "Invalid peer ASN",
			serviceKey: "azure-service-key",
			peers: []map[string]interface{}{
				{
					"type":     "private",
					"peer_asn": "invalid-asn",
				},
			},
			wantErr: true,
			errText: "Invalid Azure peer [0] ASN: invalid-asn - must be a valid ASN number",
		},
		{
			name:       "Invalid primary subnet",
			serviceKey: "azure-service-key",
			peers: []map[string]interface{}{
				{
					"type":           "private",
					"primary_subnet": "invalid-cidr",
				},
			},
			wantErr: true,
			errText: "Invalid Azure peer [0] primary subnet: invalid-cidr - must be a valid CIDR notation",
		},
		{
			name:       "Invalid secondary subnet",
			serviceKey: "azure-service-key",
			peers: []map[string]interface{}{
				{
					"type":             "private",
					"secondary_subnet": "10.0.0.1", // Not CIDR
				},
			},
			wantErr: true,
			errText: "Invalid Azure peer [0] secondary subnet: 10.0.0.1 - must be a valid CIDR notation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAzurePartnerConfig(tt.serviceKey, tt.peers)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAzurePartnerConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.wantErr {
				assert.IsType(t, &ValidationError{}, err, "Expected ValidationError type")
				assert.Equal(t, tt.errText, err.Error(), "Error message mismatch")
			}
		})
	}
}

func TestValidateGooglePartnerConfig(t *testing.T) {
	tests := []struct {
		name       string
		pairingKey string
		wantErr    bool
		errText    string
	}{
		{"Valid Google config", "google-pairing-key", false, ""},
		{"Empty pairing key", "", true, "Invalid Google pairing key:  - cannot be empty"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGooglePartnerConfig(tt.pairingKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateGooglePartnerConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.wantErr {
				assert.IsType(t, &ValidationError{}, err, "Expected ValidationError type")
				assert.Equal(t, tt.errText, err.Error(), "Error message mismatch")
			}
		})
	}
}

func TestValidateOraclePartnerConfig(t *testing.T) {
	tests := []struct {
		name             string
		virtualCircuitID string
		wantErr          bool
		errText          string
	}{
		{"Valid Oracle config", "ocid1.virtualcircuit.oc1..example", false, ""},
		{"Empty virtual circuit ID", "", true, "Invalid Oracle virtual circuit ID:  - cannot be empty"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOraclePartnerConfig(tt.virtualCircuitID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateOraclePartnerConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.wantErr {
				assert.IsType(t, &ValidationError{}, err, "Expected ValidationError type")
				assert.Equal(t, tt.errText, err.Error(), "Error message mismatch")
			}
		})
	}
}

func TestValidateIBMPartnerConfig(t *testing.T) {
	validAccountID := "abcdef0123456789abcdef0123456789" // 32 hex chars
	tests := []struct {
		name              string
		accountID         string
		customerASN       int
		ibmName           string
		customerIPAddress string
		providerIPAddress string
		wantErr           bool
		errText           string
	}{
		{
			name:              "Valid IBM config",
			accountID:         validAccountID,
			customerASN:       65000,
			ibmName:           "MyIBMConnection",
			customerIPAddress: "10.1.1.1/30",
			providerIPAddress: "10.1.1.2/30",
			wantErr:           false,
		},
		{
			name:      "Empty account ID",
			accountID: "",
			wantErr:   true,
			errText:   "Invalid IBM account ID:  - cannot be empty",
		},
		{
			name:      "Account ID too short",
			accountID: "short",
			wantErr:   true,
			errText:   fmt.Sprintf("Invalid IBM account ID: short - must be exactly %d characters", IBMAccountIDLength),
		},
		{
			name:      "Account ID too long",
			accountID: validAccountID + "extra",
			wantErr:   true,
			errText:   fmt.Sprintf("Invalid IBM account ID: %sextra - must be exactly %d characters", validAccountID, IBMAccountIDLength),
		},
		{
			name:      "Account ID invalid characters",
			accountID: "abcdef0123456789abcdef012345678X", // X is invalid
			wantErr:   true,
			errText:   "Invalid IBM account ID: abcdef0123456789abcdef012345678X - must contain only hexadecimal characters (0-9, a-f, A-F)",
		},
		{
			name:      "Name too long",
			accountID: validAccountID,
			ibmName:   string(make([]byte, MaxIBMNameLength+1)),
			wantErr:   true,
			errText:   "Invalid IBM connection name: ", // Error message includes the long name, truncated here
		},
		{
			name:      "Name invalid characters",
			accountID: validAccountID,
			ibmName:   "My IBM Connection!", // ! is invalid
			wantErr:   true,
			errText:   "Invalid IBM connection name: My IBM Connection! - must only contain characters 0-9, a-z, A-Z, /, -, _, or ,",
		},
		{
			name:              "Invalid customer IP",
			accountID:         validAccountID,
			customerIPAddress: "invalid-ip",
			wantErr:           true,
			errText:           "Invalid IBM customer IP address: invalid-ip - must be a valid CIDR notation",
		},
		{
			name:              "Invalid provider IP",
			accountID:         validAccountID,
			providerIPAddress: "10.1.1.2/33", // Invalid mask
			wantErr:           true,
			errText:           "Invalid IBM provider IP address: 10.1.1.2/33 - must be a valid CIDR notation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIBMPartnerConfig(tt.accountID, tt.customerASN, tt.ibmName, tt.customerIPAddress, tt.providerIPAddress)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateIBMPartnerConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.wantErr {
				assert.IsType(t, &ValidationError{}, err, "Expected ValidationError type")
				assert.Contains(t, err.Error(), tt.errText, "Error message mismatch")
			}
		})
	}
}

func TestValidateVXCPartnerConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
		errText string
	}{
		{
			name: "Valid AWS partner config",
			config: map[string]interface{}{
				"partner": "aws",
				"aws_config": map[string]interface{}{
					"connect_type":  "AWS",
					"owner_account": "123456789012",
				},
			},
			wantErr: false,
		},
		{
			name: "Valid Azure partner config",
			config: map[string]interface{}{
				"partner": "azure",
				"azure_config": map[string]interface{}{
					"service_key": "azure-key",
				},
			},
			wantErr: false,
		},
		{
			name: "Valid Google partner config",
			config: map[string]interface{}{
				"partner": "google",
				"google_config": map[string]interface{}{
					"pairing_key": "google-key",
				},
			},
			wantErr: false,
		},
		{
			name: "Valid Oracle partner config",
			config: map[string]interface{}{
				"partner": "oracle",
				"oracle_config": map[string]interface{}{
					"virtual_circuit_id": "oracle-vcid",
				},
			},
			wantErr: false,
		},
		{
			name: "Valid IBM partner config",
			config: map[string]interface{}{
				"partner": "ibm",
				"ibm_config": map[string]interface{}{
					"account_id": "abcdef0123456789abcdef0123456789",
				},
			},
			wantErr: false,
		},
		{
			name: "Valid vRouter partner config",
			config: map[string]interface{}{
				"partner": "vrouter",
				"vrouter_config": map[string]interface{}{
					"interfaces": []map[string]interface{}{
						{
							"vlan": 100,
							"ip_addresses": []interface{}{
								"192.168.1.1/30",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "Missing partner type",
			config:  map[string]interface{}{},
			wantErr: true,
			errText: "Invalid Partner type:  - cannot be empty", // Updated expected error (<nil> becomes empty string)
		},
		{
			name: "Invalid partner type",
			config: map[string]interface{}{
				"partner": "invalid",
			},
			wantErr: true,
			errText: "Invalid Partner type: invalid - must be one of aws, azure, google, oracle, ibm, or vrouter",
		},
		{
			name: "Missing config for partner type",
			config: map[string]interface{}{
				"partner": "aws",
				// aws_config is missing
			},
			wantErr: true,
			errText: "Invalid Partner configuration: <nil> - no configuration provided for partner type 'aws'",
		},
		{
			name: "Multiple partner configs provided",
			config: map[string]interface{}{
				"partner": "aws",
				"aws_config": map[string]interface{}{
					"connect_type":  "AWS",
					"owner_account": "123456789012",
				},
				"azure_config": map[string]interface{}{
					"service_key": "azure-key",
				},
			},
			wantErr: true,
			errText: "Invalid Azure config: map[service_key:azure-key] - cannot be provided when partner type is not azure", // Updated expected error
		},
		{
			name: "Config provided for wrong partner type",
			config: map[string]interface{}{
				"partner": "azure", // Partner is azure
				"aws_config": map[string]interface{}{ // But aws_config is provided
					"connect_type":  "AWS",
					"owner_account": "123456789012",
				},
			},
			wantErr: true,
			errText: "Invalid AWS config: map[connect_type:AWS owner_account:123456789012] - cannot be provided when partner type is not aws",
		},
		{
			name: "Invalid AWS config details",
			config: map[string]interface{}{
				"partner": "aws",
				"aws_config": map[string]interface{}{
					"connect_type": "", // Invalid connect type
				},
			},
			wantErr: true,
			errText: "Invalid AWS connect type:  - cannot be empty",
		},
		{
			name: "Invalid Azure config details",
			config: map[string]interface{}{
				"partner": "azure",
				"azure_config": map[string]interface{}{
					"service_key": "", // Invalid service key
				},
			},
			wantErr: true,
			errText: "Invalid Azure service key:  - cannot be empty",
		},
		{
			name: "Invalid vRouter config details",
			config: map[string]interface{}{
				"partner": "vrouter",
				"vrouter_config": map[string]interface{}{
					"interfaces": []map[string]interface{}{
						{
							"vlan": 1, // Invalid VLAN
						},
					},
				},
			},
			wantErr: true,
			errText: fmt.Sprintf("Invalid vRouter interface [0] VLAN: 1 - must be between %d-%d (%d is reserved)", MinVLAN, MaxVLAN, ReservedVLAN),
		},
		{
			name: "Deprecated partner_a_end_config (warning only)",
			config: map[string]interface{}{
				"partner":              "vrouter", // Assuming vrouter is the intended type
				"partner_a_end_config": map[string]interface{}{
					// Some config here...
				},
			},
			wantErr: false, // Should not error, just warn
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVXCPartnerConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVXCPartnerConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.wantErr {
				assert.IsType(t, &ValidationError{}, err, "Expected ValidationError type")
				assert.Equal(t, tt.errText, err.Error(), "Error message mismatch")
			}
		})
	}
}

func TestValidateVrouterPartnerConfig(t *testing.T) {
	tests := []struct {
		name       string
		interfaces []map[string]interface{}
		wantErr    bool
		errText    string
	}{
		{
			name: "Valid vRouter config",
			interfaces: []map[string]interface{}{
				{
					"vlan": 100,
					"ip_addresses": []interface{}{
						"192.168.1.1/30",
					},
					"bfd": map[string]interface{}{
						"tx_interval": 500,
						"multiplier":  5,
					},
					"bgp_connections": []interface{}{
						map[string]interface{}{
							"peer_asn":         65001,
							"local_ip_address": "192.168.1.1/30",
							"peer_ip_address":  "192.168.1.2",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:       "No interfaces provided",
			interfaces: []map[string]interface{}{},
			wantErr:    true,
			errText:    "Invalid vRouter interfaces: <nil> - at least one interface must be provided",
		},
		{
			name: "Invalid VLAN",
			interfaces: []map[string]interface{}{
				{
					"vlan": 1, // Reserved VLAN
				},
			},
			wantErr: true,
			errText: fmt.Sprintf("Invalid vRouter interface [0] VLAN: 1 - must be between %d-%d (%d is reserved)", MinVLAN, MaxVLAN, ReservedVLAN),
		},
		{
			name: "Invalid IP address format",
			interfaces: []map[string]interface{}{
				{
					"ip_addresses": []interface{}{
						"invalid-ip",
					},
				},
			},
			wantErr: true,
			errText: "Invalid vRouter interface [0] IP address [0]: invalid-ip - must be a valid CIDR notation",
		},
		{
			name: "Invalid NAT IP address format",
			interfaces: []map[string]interface{}{
				{
					"nat_ip_addresses": []interface{}{
						12345, // Not a string
					},
				},
			},
			wantErr: true,
			errText: "Invalid vRouter interface [0] NAT IP address [0]: 12345 - must be a string in CIDR format",
		},
		{
			name: "Invalid IP route format",
			interfaces: []map[string]interface{}{
				{
					"ip_routes": []interface{}{
						"not-a-map", // Should be map[string]interface{}
					},
				},
			},
			wantErr: true,
			errText: "Invalid vRouter interface [0] IP route [0]: not-a-map - must be a valid route configuration map",
		},
		{
			name: "Invalid BFD config",
			interfaces: []map[string]interface{}{
				{
					"bfd": map[string]interface{}{
						"tx_interval": 100, // Too low
					},
				},
			},
			wantErr: true,
			errText: fmt.Sprintf("Invalid vRouter interface [0] BFD TX interval: 100 - must be between %d-%d milliseconds", MinBFDInterval, MaxBFDInterval),
		},
		{
			name: "Invalid BGP connection format",
			interfaces: []map[string]interface{}{
				{
					"bgp_connections": []interface{}{
						"not-a-map", // Should be map[string]interface{}
					},
				},
			},
			wantErr: true,
			errText: "Invalid vRouter interface [0] BGP connection [0]: not-a-map - must be a valid BGP connection configuration map",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVrouterPartnerConfig(tt.interfaces)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVrouterPartnerConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.wantErr {
				assert.IsType(t, &ValidationError{}, err, "Expected ValidationError type")
				assert.Equal(t, tt.errText, err.Error(), "Error message mismatch")
			}
		})
	}
}

func TestValidateIPRoute(t *testing.T) {
	tests := []struct {
		name       string
		route      map[string]interface{}
		ifaceIndex int
		routeIndex int
		wantErr    bool
		errText    string
	}{
		{
			name: "Valid IP route",
			route: map[string]interface{}{
				"prefix":   "10.0.0.0/8",
				"next_hop": "192.168.1.2",
			},
			ifaceIndex: 0,
			routeIndex: 0,
			wantErr:    false,
		},
		{
			name: "Missing prefix",
			route: map[string]interface{}{
				"next_hop": "192.168.1.2",
			},
			ifaceIndex: 0,
			routeIndex: 0,
			wantErr:    true,
			errText:    "Invalid vRouter interface [0] IP route [0] prefix: <nil> - cannot be empty and must be a string",
		},
		{
			name: "Invalid prefix CIDR",
			route: map[string]interface{}{
				"prefix":   "10.0.0.0/33",
				"next_hop": "192.168.1.2",
			},
			ifaceIndex: 0,
			routeIndex: 0,
			wantErr:    true,
			errText:    "Invalid vRouter interface [0] IP route [0] prefix: 10.0.0.0/33 - must be a valid CIDR notation",
		},
		{
			name: "Missing next hop",
			route: map[string]interface{}{
				"prefix": "10.0.0.0/8",
			},
			ifaceIndex: 0,
			routeIndex: 0,
			wantErr:    true,
			errText:    "Invalid vRouter interface [0] IP route [0] next hop: <nil> - cannot be empty and must be a string",
		},
		{
			name: "Invalid next hop IP",
			route: map[string]interface{}{
				"prefix":   "10.0.0.0/8",
				"next_hop": "invalid-ip",
			},
			ifaceIndex: 0,
			routeIndex: 0,
			wantErr:    true,
			errText:    "Invalid vRouter interface [0] IP route [0] next hop: invalid-ip - must be a valid IPv4 address",
		},
		{
			name: "Next hop is CIDR (invalid)",
			route: map[string]interface{}{
				"prefix":   "10.0.0.0/8",
				"next_hop": "192.168.1.2/30",
			},
			ifaceIndex: 0,
			routeIndex: 0,
			wantErr:    true,
			errText:    "Invalid vRouter interface [0] IP route [0] next hop: 192.168.1.2/30 - must be a valid IPv4 address (not CIDR)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIPRoute(tt.route, tt.ifaceIndex, tt.routeIndex)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateIPRoute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.wantErr {
				assert.IsType(t, &ValidationError{}, err, "Expected ValidationError type")
				assert.Equal(t, tt.errText, err.Error(), "Error message mismatch")
			}
		})
	}
}

func TestValidateBFDConfig(t *testing.T) {
	tests := []struct {
		name       string
		bfd        map[string]interface{}
		ifaceIndex int
		wantErr    bool
		errText    string
	}{
		{
			name: "Valid BFD config",
			bfd: map[string]interface{}{
				"tx_interval": 500,
				"rx_interval": 500,
				"multiplier":  5,
			},
			ifaceIndex: 0,
			wantErr:    false,
		},
		{
			name: "TX interval too low",
			bfd: map[string]interface{}{
				"tx_interval": MinBFDInterval - 1,
			},
			ifaceIndex: 0,
			wantErr:    true,
			errText:    fmt.Sprintf("Invalid vRouter interface [0] BFD TX interval: %d - must be between %d-%d milliseconds", MinBFDInterval-1, MinBFDInterval, MaxBFDInterval),
		},
		{
			name: "TX interval too high",
			bfd: map[string]interface{}{
				"tx_interval": MaxBFDInterval + 1,
			},
			ifaceIndex: 0,
			wantErr:    true,
			errText:    fmt.Sprintf("Invalid vRouter interface [0] BFD TX interval: %d - must be between %d-%d milliseconds", MaxBFDInterval+1, MinBFDInterval, MaxBFDInterval),
		},
		{
			name: "RX interval too low",
			bfd: map[string]interface{}{
				"rx_interval": MinBFDInterval - 1,
			},
			ifaceIndex: 0,
			wantErr:    true,
			errText:    fmt.Sprintf("Invalid vRouter interface [0] BFD RX interval: %d - must be between %d-%d milliseconds", MinBFDInterval-1, MinBFDInterval, MaxBFDInterval),
		},
		{
			name: "RX interval too high",
			bfd: map[string]interface{}{
				"rx_interval": MaxBFDInterval + 1,
			},
			ifaceIndex: 0,
			wantErr:    true,
			errText:    fmt.Sprintf("Invalid vRouter interface [0] BFD RX interval: %d - must be between %d-%d milliseconds", MaxBFDInterval+1, MinBFDInterval, MaxBFDInterval),
		},
		{
			name: "Multiplier too low",
			bfd: map[string]interface{}{
				"multiplier": MinBFDMultiplier - 1,
			},
			ifaceIndex: 0,
			wantErr:    true,
			errText:    fmt.Sprintf("Invalid vRouter interface [0] BFD multiplier: %d - must be between %d-%d", MinBFDMultiplier-1, MinBFDMultiplier, MaxBFDMultiplier),
		},
		{
			name: "Multiplier too high",
			bfd: map[string]interface{}{
				"multiplier": MaxBFDMultiplier + 1,
			},
			ifaceIndex: 0,
			wantErr:    true,
			errText:    fmt.Sprintf("Invalid vRouter interface [0] BFD multiplier: %d - must be between %d-%d", MaxBFDMultiplier+1, MinBFDMultiplier, MaxBFDMultiplier),
		},
		{
			name: "Invalid type for tx_interval",
			bfd: map[string]interface{}{
				"tx_interval": "not-an-int",
			},
			ifaceIndex: 0,
			wantErr:    true,
			errText:    "Invalid vRouter interface [0] BFD TX interval: not-an-int - must be a valid integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBFDConfig(tt.bfd, tt.ifaceIndex)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateBFDConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.wantErr {
				assert.IsType(t, &ValidationError{}, err, "Expected ValidationError type")
				assert.Equal(t, tt.errText, err.Error(), "Error message mismatch")
			}
		})
	}
}

func TestValidateBGPConnection(t *testing.T) {
	tests := []struct {
		name       string
		conn       map[string]interface{}
		ifaceIndex int
		connIndex  int
		wantErr    bool
		errText    string
	}{
		{
			name: "Valid BGP connection",
			conn: map[string]interface{}{
				"peer_asn":              65001,
				"local_ip_address":      "192.168.1.1/30",
				"peer_ip_address":       "192.168.1.2",
				"peer_type":             BGPPeerNonCloud,
				"med_in":                100,
				"as_path_prepend_count": 3,
				"export_policy":         BGPExportPolicyPermit,
			},
			ifaceIndex: 0,
			connIndex:  0,
			wantErr:    false,
		},
		{
			name: "Missing peer ASN",
			conn: map[string]interface{}{
				"local_ip_address": "192.168.1.1/30",
				"peer_ip_address":  "192.168.1.2",
			},
			ifaceIndex: 0,
			connIndex:  0,
			wantErr:    true,
			errText:    "Invalid vRouter interface [0] BGP connection [0] peer ASN: <nil> - is required",
		},
		{
			name: "Invalid peer ASN type",
			conn: map[string]interface{}{
				"peer_asn":         "not-an-int",
				"local_ip_address": "192.168.1.1/30",
				"peer_ip_address":  "192.168.1.2",
			},
			ifaceIndex: 0,
			connIndex:  0,
			wantErr:    true,
			errText:    "Invalid vRouter interface [0] BGP connection [0] peer ASN: not-an-int - must be a valid integer ASN",
		},
		{
			name: "Invalid local ASN type",
			conn: map[string]interface{}{
				"peer_asn":         65001,
				"local_asn":        "not-an-int",
				"local_ip_address": "192.168.1.1/30",
				"peer_ip_address":  "192.168.1.2",
			},
			ifaceIndex: 0,
			connIndex:  0,
			wantErr:    true,
			errText:    "Invalid vRouter interface [0] BGP connection [0] local ASN: not-an-int - must be a valid integer ASN",
		},
		{
			name: "Missing local IP address",
			conn: map[string]interface{}{
				"peer_asn":        65001,
				"peer_ip_address": "192.168.1.2",
			},
			ifaceIndex: 0,
			connIndex:  0,
			wantErr:    true,
			errText:    "Invalid vRouter interface [0] BGP connection [0] local IP address: <nil> - cannot be empty and must be a string",
		},
		{
			name: "Invalid local IP address format",
			conn: map[string]interface{}{
				"peer_asn":         65001,
				"local_ip_address": "invalid",
				"peer_ip_address":  "192.168.1.2",
			},
			ifaceIndex: 0,
			connIndex:  0,
			wantErr:    true,
			errText:    "Invalid vRouter interface [0] BGP connection [0] local IP address: invalid - must be a valid IPv4 address",
		},
		{
			name: "Missing peer IP address",
			conn: map[string]interface{}{
				"peer_asn":         65001,
				"local_ip_address": "192.168.1.1/30",
			},
			ifaceIndex: 0,
			connIndex:  0,
			wantErr:    true,
			errText:    "Invalid vRouter interface [0] BGP connection [0] peer IP address: <nil> - cannot be empty and must be a string",
		},
		{
			name: "Invalid peer IP address format",
			conn: map[string]interface{}{
				"peer_asn":         65001,
				"local_ip_address": "192.168.1.1/30",
				"peer_ip_address":  "invalid",
			},
			ifaceIndex: 0,
			connIndex:  0,
			wantErr:    true,
			errText:    "Invalid vRouter interface [0] BGP connection [0] peer IP address: invalid - must be a valid IPv4 address",
		},
		{
			name: "Invalid peer type",
			conn: map[string]interface{}{
				"peer_asn":         65001,
				"local_ip_address": "192.168.1.1/30",
				"peer_ip_address":  "192.168.1.2",
				"peer_type":        "INVALID_TYPE",
			},
			ifaceIndex: 0,
			connIndex:  0,
			wantErr:    true,
			errText:    fmt.Sprintf("Invalid vRouter interface [0] BGP connection [0] peer type: INVALID_TYPE - must be one of '%s', '%s', or '%s'", BGPPeerNonCloud, BGPPeerPrivCloud, BGPPeerPubCloud),
		},
		{
			name: "MED in too low",
			conn: map[string]interface{}{
				"peer_asn":         65001,
				"local_ip_address": "192.168.1.1/30",
				"peer_ip_address":  "192.168.1.2",
				"med_in":           MinMED - 1,
			},
			ifaceIndex: 0,
			connIndex:  0,
			wantErr:    true,
			errText:    fmt.Sprintf("Invalid vRouter interface [0] BGP connection [0] MED in: %d - must be between %d-%d", MinMED-1, MinMED, MaxMED),
		},
		{
			name: "MED out too high",
			conn: map[string]interface{}{
				"peer_asn":         65001,
				"local_ip_address": "192.168.1.1/30",
				"peer_ip_address":  "192.168.1.2",
				"med_out":          MaxMED + 1,
			},
			ifaceIndex: 0,
			connIndex:  0,
			wantErr:    true,
			// Note: MaxMED is large, so the error message might look strange with large numbers
			errText: fmt.Sprintf("Invalid vRouter interface [0] BGP connection [0] MED out: %d - must be between %d-%d", MaxMED+1, MinMED, MaxMED),
		},
		{
			name: "AS path prepend count too low",
			conn: map[string]interface{}{
				"peer_asn":              65001,
				"local_ip_address":      "192.168.1.1/30",
				"peer_ip_address":       "192.168.1.2",
				"as_path_prepend_count": MinASPathPrependCount - 1,
			},
			ifaceIndex: 0,
			connIndex:  0,
			wantErr:    true,
			errText:    fmt.Sprintf("Invalid vRouter interface [0] BGP connection [0] AS path prepend count: %d - must be between %d-%d", MinASPathPrependCount-1, MinASPathPrependCount, MaxASPathPrependCount),
		},
		{
			name: "AS path prepend count too high",
			conn: map[string]interface{}{
				"peer_asn":              65001,
				"local_ip_address":      "192.168.1.1/30",
				"peer_ip_address":       "192.168.1.2",
				"as_path_prepend_count": MaxASPathPrependCount + 1,
			},
			ifaceIndex: 0,
			connIndex:  0,
			wantErr:    true,
			errText:    fmt.Sprintf("Invalid vRouter interface [0] BGP connection [0] AS path prepend count: %d - must be between %d-%d", MaxASPathPrependCount+1, MinASPathPrependCount, MaxASPathPrependCount),
		},
		{
			name: "Invalid export policy",
			conn: map[string]interface{}{
				"peer_asn":         65001,
				"local_ip_address": "192.168.1.1/30",
				"peer_ip_address":  "192.168.1.2",
				"export_policy":    "invalid",
			},
			ifaceIndex: 0,
			connIndex:  0,
			wantErr:    true,
			errText:    fmt.Sprintf("Invalid vRouter interface [0] BGP connection [0] export policy: invalid - must be '%s' or '%s'", BGPExportPolicyPermit, BGPExportPolicyDeny),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBGPConnection(tt.conn, tt.ifaceIndex, tt.connIndex)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateBGPConnection() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.wantErr {
				assert.IsType(t, &ValidationError{}, err, "Expected ValidationError type")
				assert.Equal(t, tt.errText, err.Error(), "Error message mismatch")
			}
		})
	}
}
