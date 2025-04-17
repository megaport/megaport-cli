package validation

import (
	"fmt"
	"testing"

	megaport "github.com/megaport/megaportgo"
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
				t.Errorf("ValidateVXCEndVLAN() error = %v, wantErr:%v", err, tt.wantErr)
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
				t.Errorf("ValidateVXCEndInnerVLAN() error = %v, wantErr:%v", err, tt.wantErr)
			}
			if err != nil && tt.wantErr {
				assert.IsType(t, &ValidationError{}, err, "Expected ValidationError type")
			}
		})
	}
}

func TestValidateVXCRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *megaport.BuyVXCRequest
		wantErr bool
		errText string
	}{
		{
			name: "Valid VXC request",
			req: &megaport.BuyVXCRequest{
				VXCName:   "Test VXC",
				Term:      12,
				RateLimit: 1000,
				PortUID:   "a-end-uid",
				BEndConfiguration: megaport.VXCOrderEndpointConfiguration{
					ProductUID: "b-end-uid",
				},
			},
			wantErr: false,
		},
		{
			name: "Valid VXC request with partner config",
			req: &megaport.BuyVXCRequest{
				VXCName:   "Test VXC Partner",
				Term:      1,
				RateLimit: 500,
				PortUID:   "a-end-uid",
				BEndConfiguration: megaport.VXCOrderEndpointConfiguration{
					PartnerConfig: &megaport.VXCPartnerConfigAWS{
						ConnectType:  "AWS",
						OwnerAccount: "12345",
						ASN:          65000,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Empty name",
			req: &megaport.BuyVXCRequest{
				VXCName:   "",
				Term:      12,
				RateLimit: 1000,
				PortUID:   "a-end-uid",
				BEndConfiguration: megaport.VXCOrderEndpointConfiguration{
					ProductUID: "b-end-uid",
				},
			},
			wantErr: true,
			errText: "Invalid VXC name:  - cannot be empty",
		},
		{
			name: "Invalid term",
			req: &megaport.BuyVXCRequest{
				VXCName:   "Test VXC",
				Term:      5, // Invalid term
				RateLimit: 1000,
				PortUID:   "a-end-uid",
				BEndConfiguration: megaport.VXCOrderEndpointConfiguration{
					ProductUID: "b-end-uid",
				},
			},
			wantErr: true,
			errText: fmt.Sprintf("Invalid contract term: 5 - must be one of: %v", ValidContractTerms),
		},
		{
			name: "Invalid rate limit",
			req: &megaport.BuyVXCRequest{
				VXCName:   "Test VXC",
				Term:      12,
				RateLimit: 0, // Invalid rate limit
				PortUID:   "a-end-uid",
				BEndConfiguration: megaport.VXCOrderEndpointConfiguration{
					ProductUID: "b-end-uid",
				},
			},
			wantErr: true,
			errText: "Invalid rate limit: 0 - must be a positive integer",
		},
		{
			name: "Empty A-End UID",
			req: &megaport.BuyVXCRequest{
				VXCName:   "Test VXC",
				Term:      12,
				RateLimit: 1000,
				PortUID:   "", // Empty A-End UID
				BEndConfiguration: megaport.VXCOrderEndpointConfiguration{
					ProductUID: "b-end-uid",
				},
			},
			wantErr: true,
			errText: "Invalid A-End UID (PortUID):  - cannot be empty",
		},
		{
			name: "Empty B-End UID without partner config",
			req: &megaport.BuyVXCRequest{
				VXCName:   "Test VXC",
				Term:      12,
				RateLimit: 1000,
				PortUID:   "a-end-uid",
				BEndConfiguration: megaport.VXCOrderEndpointConfiguration{
					ProductUID: "", // Empty B-End UID
				},
			},
			wantErr: true,
			errText: "Invalid B-End UID:  - cannot be empty when no partner configuration is provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVXCRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVXCRequest() error = %v, wantErr:%v", err, tt.wantErr)
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
			connectType:       "private",
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
			connectType:       "public",
			ownerAccount:      "123456789012",
			asn:               65001,
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
			asn:          65000,
			wantErr:      true,
			errText:      "Invalid AWS connect type:  - cannot be empty",
		},
		{
			connectType:  "INVALID",
			ownerAccount: "123456789012",
			asn:          65000,
			wantErr:      true,
			errText:      "Invalid AWS connect type: INVALID - must be 'AWS', or 'AWSHC'",
		},
		{
			name:         "Empty owner account",
			connectType:  "AWS",
			ownerAccount: "",
			asn:          65000,
			wantErr:      true,
			errText:      "Invalid AWS owner account:  - cannot be empty",
		},
		{
			name:              "Invalid customer IP CIDR",
			connectType:       "AWS",
			ownerAccount:      "123456789012",
			asn:               65000,
			customerIPAddress: "invalid-ip",
			wantErr:           true,
			errText:           "Invalid AWS customer IP address: invalid-ip - must be a valid CIDR notation", // Updated error message
		},
		{
			name:            "Invalid Amazon IP CIDR",
			connectType:     "AWS",
			ownerAccount:    "123456789012",
			asn:             65000,
			amazonIPAddress: "192.168.1.2/33", // Invalid mask
			wantErr:         true,
			errText:         "Invalid AWS Amazon IP address: 192.168.1.2/33 - must be a valid CIDR notation", // Updated error message
		},
		{
			name:         "AWS name too long",
			connectType:  "AWS",
			ownerAccount: "123456789012",
			asn:          65000,
			awsName:      string(make([]byte, 256)), // 256 chars
			wantErr:      true,
			errText:      "Invalid AWS connection name: ", // Error message includes the long name, truncated here
		},
		{
			name:         "Invalid AWS type for AWS connect type",
			connectType:  "AWS",
			ownerAccount: "123456789012",
			asn:          65000,
			awsType:      "invalid",
			wantErr:      true,
			errText:      "Invalid AWS type: invalid - must be 'private' or 'public' for AWS connect type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			awsConfig := &megaport.VXCPartnerConfigAWS{
				ConnectType:       tt.connectType,
				OwnerAccount:      tt.ownerAccount,
				CustomerIPAddress: tt.customerIPAddress,
				AmazonIPAddress:   tt.amazonIPAddress,
				ConnectionName:    tt.awsName,
				Type:              tt.awsType,
				ASN:               tt.asn,
				AmazonASN:         tt.amazonAsn,
			}
			err := ValidateAWSPartnerConfig(awsConfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAWSPartnerConfig() error = %v, wantErr:%v", err, tt.wantErr)
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
			googleConfig := &megaport.VXCPartnerConfigGoogle{
				PairingKey: tt.pairingKey,
			}
			err := ValidateGooglePartnerConfig(googleConfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateGooglePartnerConfig() error = %v, wantErr:%v", err, tt.wantErr)
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
			// Fix: Use the correct struct for ValidateOraclePartnerConfig
			oracleConfig := &megaport.VXCPartnerConfigOracle{
				VirtualCircuitId: tt.virtualCircuitID,
			}
			err := ValidateOraclePartnerConfig(oracleConfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateOraclePartnerConfig() error = %v, wantErr:%v", err, tt.wantErr)
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
			// Fix: Use the correct struct for ValidateIBMPartnerConfig
			ibmConfig := &megaport.VXCPartnerConfigIBM{
				AccountID:         tt.accountID,
				Name:              tt.ibmName,
				CustomerIPAddress: tt.customerIPAddress,
				ProviderIPAddress: tt.providerIPAddress,
			}
			err := ValidateIBMPartnerConfig(ibmConfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateIBMPartnerConfig() error = %v, wantErr:%v", err, tt.wantErr)
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
		config  megaport.VXCPartnerConfiguration // Use interface type
		wantErr bool
		errText string
	}{
		{
			name: "Valid AWS partner config",
			config: &megaport.VXCPartnerConfigAWS{ // Use struct pointer
				ConnectType:  "AWS",
				OwnerAccount: "123456789012",
				ASN:          65000,
			},
			wantErr: false,
		},
		{
			name: "Valid Azure partner config",
			config: &megaport.VXCPartnerConfigAzure{ // Use struct pointer
				ConnectType: "AZURE", // Assuming ConnectType is needed
				ServiceKey:  "azure-key",
			},
			wantErr: false,
		},
		{
			name: "Valid Google partner config",
			config: &megaport.VXCPartnerConfigGoogle{ // Use struct pointer
				ConnectType: "GOOGLE", // Assuming ConnectType is needed
				PairingKey:  "google-key",
			},
			wantErr: false,
		},
		{
			name: "Valid Oracle partner config",
			config: &megaport.VXCPartnerConfigOracle{ // Use struct pointer
				ConnectType:      "ORACLE", // Assuming ConnectType is needed
				VirtualCircuitId: "oracle-vcid",
			},
			wantErr: false,
		},
		{
			name: "Valid IBM partner config",
			config: &megaport.VXCPartnerConfigIBM{ // Use struct pointer
				ConnectType: "IBM", // Assuming ConnectType is needed
				AccountID:   "abcdef0123456789abcdef0123456789",
			},
			wantErr: false,
		},
		{
			name: "Valid vRouter partner config",
			config: &megaport.VXCOrderVrouterPartnerConfig{ // Use struct pointer
				Interfaces: []megaport.PartnerConfigInterface{
					{
						VLAN: 100,
						IpAddresses: []string{
							"192.168.1.1/30",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "Missing partner type (nil config)", // Test case for nil config
			config:  nil,
			wantErr: true,
			errText: "Invalid Partner configuration type: <nil> - is not supported", // Adjusted error based on test output
		},
		{
			name: "Invalid partner type in AWS config",
			config: &megaport.VXCPartnerConfigAWS{
				ConnectType:  "invalid", // Invalid type set in a specific struct
				OwnerAccount: "123456789012",
				ASN:          65000,
			},
			wantErr: true,
			errText: "Invalid AWS connect type: invalid - must be 'AWS', or 'AWSHC'", // Adjusted error based on test output
		},
		{
			name:    "Missing config details (handled by specific validators)",
			config:  &megaport.VXCPartnerConfigAWS{}, // Empty AWS config
			wantErr: true,
			errText: "Invalid AWS connect type:  - cannot be empty", // Error from ValidateAWSPartnerConfig
		},
		{
			name: "Invalid AWS config details",
			config: &megaport.VXCPartnerConfigAWS{
				ConnectType: "", // Invalid connect type
				ASN:         65000,
			},
			wantErr: true,
			errText: "Invalid AWS connect type:  - cannot be empty",
		},
		{
			name: "Invalid Azure config details",
			config: &megaport.VXCPartnerConfigAzure{
				ServiceKey: "", // Invalid service key
			},
			wantErr: true,
			errText: "Invalid Azure service key:  - cannot be empty",
		},
		{
			name: "Invalid vRouter config details",
			config: &megaport.VXCOrderVrouterPartnerConfig{
				Interfaces: []megaport.PartnerConfigInterface{
					{
						VLAN: 1, // Invalid VLAN
					},
				},
			},
			wantErr: true,
			errText: fmt.Sprintf("Invalid vRouter interface [0] VLAN: 1 - must be between %d-%d (%d is reserved)", AutoAssignVLAN, MaxVLAN, ReservedVLAN),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVXCPartnerConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVXCPartnerConfig() error = %v, wantErr:%v", err, tt.wantErr)
				return
			}
			if err != nil && tt.wantErr {
				assert.IsType(t, &ValidationError{}, err, "Expected ValidationError type")
				assert.Equal(t, tt.errText, err.Error(), "Error message mismatch")
			}
		})
	}
}

func TestValidateAzurePartnerConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *megaport.VXCPartnerConfigAzure
		wantErr bool
		errText string
	}{
		{
			name: "Valid Azure config without peers",
			config: &megaport.VXCPartnerConfigAzure{
				ServiceKey: "valid-service-key",
			},
			wantErr: false,
		},
		{
			name: "Valid Azure config with valid peer - primary subnet",
			config: &megaport.VXCPartnerConfigAzure{
				ServiceKey: "valid-service-key",
				Peers: []megaport.PartnerOrderAzurePeeringConfig{
					{
						PrimarySubnet: "10.0.0.0/24",
						VLAN:          100,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Valid Azure config with valid peer - secondary subnet",
			config: &megaport.VXCPartnerConfigAzure{
				ServiceKey: "valid-service-key",
				Peers: []megaport.PartnerOrderAzurePeeringConfig{
					{
						SecondarySubnet: "10.0.1.0/24",
						VLAN:            200, // Added valid VLAN
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Valid Azure config with valid peer - both subnets",
			config: &megaport.VXCPartnerConfigAzure{
				ServiceKey: "valid-service-key",
				Peers: []megaport.PartnerOrderAzurePeeringConfig{
					{
						PrimarySubnet:   "10.0.0.0/24",
						SecondarySubnet: "10.0.1.0/24",
						VLAN:            300, // Added valid VLAN
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Valid Azure config with auto-assign VLAN",
			config: &megaport.VXCPartnerConfigAzure{
				ServiceKey: "valid-service-key",
				Peers: []megaport.PartnerOrderAzurePeeringConfig{
					{
						PrimarySubnet: "10.0.0.0/24",
						VLAN:          0, // Auto-assign
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Valid Azure config with untagged VLAN",
			config: &megaport.VXCPartnerConfigAzure{
				ServiceKey: "valid-service-key",
				Peers: []megaport.PartnerOrderAzurePeeringConfig{
					{
						PrimarySubnet: "10.0.0.0/24",
						VLAN:          -1, // Untagged
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Empty service key",
			config: &megaport.VXCPartnerConfigAzure{
				ServiceKey: "",
			},
			wantErr: true,
			errText: "Invalid Azure service key:  - cannot be empty",
		},
		{
			name:    "Nil config",
			config:  nil,
			wantErr: true,
			errText: "Invalid Azure partner config: <nil> - cannot be nil",
		},
		{
			name: "Invalid peer - missing subnets",
			config: &megaport.VXCPartnerConfigAzure{
				ServiceKey: "valid-service-key",
				Peers: []megaport.PartnerOrderAzurePeeringConfig{
					{
						VLAN: 100, // Valid VLAN but missing subnets
					},
				},
			},
			wantErr: true,
			errText: "Invalid Azure peer [0] subnet: <nil> - at least one of primary_subnet or secondary_subnet must be provided",
		},
		{
			name: "Invalid peer - invalid VLAN (reserved value)",
			config: &megaport.VXCPartnerConfigAzure{
				ServiceKey: "valid-service-key",
				Peers: []megaport.PartnerOrderAzurePeeringConfig{
					{
						PrimarySubnet: "10.0.0.0/24",
						VLAN:          1, // Reserved VLAN
					},
				},
			},
			wantErr: true,
			errText: fmt.Sprintf("Invalid Azure peer [0] VLAN: 1 - must be valid (0 for auto-assign, -1 for untagged, or %d-%d except %d)",
				MinAssignableVLAN, MaxVLAN, ReservedVLAN),
		},
		{
			name: "Invalid peer - VLAN too high",
			config: &megaport.VXCPartnerConfigAzure{
				ServiceKey: "valid-service-key",
				Peers: []megaport.PartnerOrderAzurePeeringConfig{
					{
						PrimarySubnet: "10.0.0.0/24",
						VLAN:          4095, // Too high
					},
				},
			},
			wantErr: true,
			errText: fmt.Sprintf("Invalid Azure peer [0] VLAN: 4095 - must be valid (0 for auto-assign, -1 for untagged, or %d-%d except %d)",
				MinAssignableVLAN, MaxVLAN, ReservedVLAN),
		},
		{
			name: "Multiple peers with one invalid",
			config: &megaport.VXCPartnerConfigAzure{
				ServiceKey: "valid-service-key",
				Peers: []megaport.PartnerOrderAzurePeeringConfig{
					{
						PrimarySubnet: "10.0.0.0/24", // Valid
						VLAN:          100,           // Valid VLAN
					},
					{
						// Invalid - both subnets missing
						VLAN: 200, // Valid VLAN
					},
				},
			},
			wantErr: true,
			errText: "Invalid Azure peer [1] subnet: <nil> - at least one of primary_subnet or secondary_subnet must be provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAzurePartnerConfig(tt.config)
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
