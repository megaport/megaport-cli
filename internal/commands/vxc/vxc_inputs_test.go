package vxc

import (
	"testing"

	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
)

func TestParsePartnerConfigFromJSON(t *testing.T) {
	tests := []struct {
		name          string
		jsonStr       string
		expectedError string
		validate      func(*testing.T, megaport.VXCPartnerConfiguration)
	}{
		{
			name:    "valid AWS config",
			jsonStr: `{"connectType":"AWS","ownerAccount":"123456"}`,
			validate: func(t *testing.T, cfg megaport.VXCPartnerConfiguration) {
				_, ok := cfg.(*megaport.VXCPartnerConfigAWS)
				assert.True(t, ok, "expected *megaport.VXCPartnerConfigAWS")
			},
		},
		{
			name:    "valid transit config",
			jsonStr: `{"connectType":"TRANSIT"}`,
			validate: func(t *testing.T, cfg megaport.VXCPartnerConfiguration) {
				_, ok := cfg.(*megaport.VXCPartnerConfigTransit)
				assert.True(t, ok, "expected *megaport.VXCPartnerConfigTransit")
			},
		},
		{
			name:    "case insensitive connectType",
			jsonStr: `{"connectType":"aws","ownerAccount":"123456"}`,
			validate: func(t *testing.T, cfg megaport.VXCPartnerConfiguration) {
				_, ok := cfg.(*megaport.VXCPartnerConfigAWS)
				assert.True(t, ok, "expected *megaport.VXCPartnerConfigAWS")
			},
		},
		{
			name:          "invalid JSON",
			jsonStr:       `{invalid}`,
			expectedError: "invalid",
		},
		{
			name:          "missing connectType",
			jsonStr:       `{"ownerAccount":"123456"}`,
			expectedError: "connectType",
		},
		{
			name:          "unsupported connectType",
			jsonStr:       `{"connectType":"UNKNOWN"}`,
			expectedError: "unsupported",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parsePartnerConfigFromJSON(tt.jsonStr)
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, result)
				}
			}
		})
	}
}

func TestParseAWSConfig(t *testing.T) {
	tests := []struct {
		name          string
		config        map[string]interface{}
		expectedError string
		validate      func(*testing.T, *megaport.VXCPartnerConfigAWS)
	}{
		{
			name: "required fields only",
			config: map[string]interface{}{
				"connectType":  "AWS",
				"ownerAccount": "123456",
			},
			validate: func(t *testing.T, cfg *megaport.VXCPartnerConfigAWS) {
				assert.Equal(t, "AWS", cfg.ConnectType)
				assert.Equal(t, "123456", cfg.OwnerAccount)
			},
		},
		{
			name: "all optional fields",
			config: map[string]interface{}{
				"connectType":       "AWS",
				"ownerAccount":      "123456",
				"asn":               65000.0,
				"amazonAsn":         64512.0,
				"authKey":           "key",
				"prefixes":          "10.0.0.0/8",
				"customerIPAddress": "1.2.3.4",
				"amazonIPAddress":   "5.6.7.8",
				"connectionName":    "conn1",
				"type":              "private",
			},
			validate: func(t *testing.T, cfg *megaport.VXCPartnerConfigAWS) {
				assert.Equal(t, "AWS", cfg.ConnectType)
				assert.Equal(t, "123456", cfg.OwnerAccount)
				assert.Equal(t, 65000, cfg.ASN)
				assert.Equal(t, 64512, cfg.AmazonASN)
				assert.Equal(t, "key", cfg.AuthKey)
				assert.Equal(t, "10.0.0.0/8", cfg.Prefixes)
				assert.Equal(t, "1.2.3.4", cfg.CustomerIPAddress)
				assert.Equal(t, "5.6.7.8", cfg.AmazonIPAddress)
				assert.Equal(t, "conn1", cfg.ConnectionName)
				assert.Equal(t, "private", cfg.Type)
			},
		},
		{
			name: "missing ownerAccount",
			config: map[string]interface{}{
				"connectType": "AWS",
			},
			expectedError: "ownerAccount",
		},
		{
			name: "empty ownerAccount",
			config: map[string]interface{}{
				"connectType":  "AWS",
				"ownerAccount": "",
			},
			expectedError: "cannot be empty",
		},
		{
			name: "asn wrong type",
			config: map[string]interface{}{
				"connectType":  "AWS",
				"ownerAccount": "123456",
				"asn":          "string",
			},
			expectedError: "must be a number",
		},
		{
			name: "negative asn",
			config: map[string]interface{}{
				"connectType":  "AWS",
				"ownerAccount": "123456",
				"asn":          -1.0,
			},
			expectedError: "cannot be negative",
		},
		{
			name: "AWSHC connectType",
			config: map[string]interface{}{
				"connectType":  "AWSHC",
				"ownerAccount": "123456",
			},
			validate: func(t *testing.T, cfg *megaport.VXCPartnerConfigAWS) {
				assert.Equal(t, "AWSHC", cfg.ConnectType)
				assert.Equal(t, "123456", cfg.OwnerAccount)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseAWSConfig(tt.config)
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, result)
				}
			}
		})
	}
}

func TestParseAzureConfig(t *testing.T) {
	tests := []struct {
		name          string
		config        map[string]interface{}
		expectedError string
		validate      func(*testing.T, *megaport.VXCPartnerConfigAzure)
	}{
		{
			name: "required fields only",
			config: map[string]interface{}{
				"serviceKey": "sk-123",
			},
			validate: func(t *testing.T, cfg *megaport.VXCPartnerConfigAzure) {
				assert.Equal(t, "AZURE", cfg.ConnectType)
				assert.Equal(t, "sk-123", cfg.ServiceKey)
			},
		},
		{
			name: "with peers",
			config: map[string]interface{}{
				"serviceKey": "sk-123",
				"peers": []interface{}{
					map[string]interface{}{
						"type":            "private",
						"peerASN":         "65000",
						"primarySubnet":   "10.0.0.0/30",
						"secondarySubnet": "10.0.0.4/30",
						"prefixes":        "10.1.0.0/16",
						"sharedKey":       "key",
						"vlan":            100.0,
					},
				},
			},
			validate: func(t *testing.T, cfg *megaport.VXCPartnerConfigAzure) {
				assert.Equal(t, "AZURE", cfg.ConnectType)
				assert.Equal(t, "sk-123", cfg.ServiceKey)
				assert.Len(t, cfg.Peers, 1)
				peer := cfg.Peers[0]
				assert.Equal(t, "private", peer.Type)
				assert.Equal(t, "65000", peer.PeerASN)
				assert.Equal(t, "10.0.0.0/30", peer.PrimarySubnet)
				assert.Equal(t, "10.0.0.4/30", peer.SecondarySubnet)
				assert.Equal(t, "10.1.0.0/16", peer.Prefixes)
				assert.Equal(t, "key", peer.SharedKey)
				assert.Equal(t, 100, peer.VLAN)
			},
		},
		{
			name:          "missing serviceKey",
			config:        map[string]interface{}{},
			expectedError: "serviceKey",
		},
		{
			name: "empty serviceKey",
			config: map[string]interface{}{
				"serviceKey": "",
			},
			expectedError: "cannot be empty",
		},
		{
			name: "peers not array",
			config: map[string]interface{}{
				"serviceKey": "sk-123",
				"peers":      "string",
			},
			expectedError: "must be an array",
		},
		{
			name: "invalid vlan in peer",
			config: map[string]interface{}{
				"serviceKey": "sk-123",
				"peers": []interface{}{
					map[string]interface{}{
						"vlan": "string",
					},
				},
			},
			expectedError: "must be a number",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseAzureConfig(tt.config)
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, result)
				}
			}
		})
	}
}

func TestParseGoogleConfig(t *testing.T) {
	tests := []struct {
		name          string
		config        map[string]interface{}
		expectedError string
		validate      func(*testing.T, *megaport.VXCPartnerConfigGoogle)
	}{
		{
			name: "valid",
			config: map[string]interface{}{
				"pairingKey": "pk-123",
			},
			validate: func(t *testing.T, cfg *megaport.VXCPartnerConfigGoogle) {
				assert.Equal(t, "GOOGLE", cfg.ConnectType)
				assert.Equal(t, "pk-123", cfg.PairingKey)
			},
		},
		{
			name:          "missing pairingKey",
			config:        map[string]interface{}{},
			expectedError: "pairingKey",
		},
		{
			name: "empty pairingKey",
			config: map[string]interface{}{
				"pairingKey": "",
			},
			expectedError: "cannot be empty",
		},
		{
			name: "non-string pairingKey",
			config: map[string]interface{}{
				"pairingKey": 123.0,
			},
			expectedError: "must be a string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseGoogleConfig(tt.config)
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, result)
				}
			}
		})
	}
}

func TestParseOracleConfig(t *testing.T) {
	tests := []struct {
		name          string
		config        map[string]interface{}
		expectedError string
		validate      func(*testing.T, *megaport.VXCPartnerConfigOracle)
	}{
		{
			name: "valid",
			config: map[string]interface{}{
				"virtualCircuitId": "vc-123",
			},
			validate: func(t *testing.T, cfg *megaport.VXCPartnerConfigOracle) {
				assert.Equal(t, "ORACLE", cfg.ConnectType)
				assert.Equal(t, "vc-123", cfg.VirtualCircuitId)
			},
		},
		{
			name:          "missing",
			config:        map[string]interface{}{},
			expectedError: "virtualCircuitId",
		},
		{
			name: "empty",
			config: map[string]interface{}{
				"virtualCircuitId": "",
			},
			expectedError: "cannot be empty",
		},
		{
			name: "non-string",
			config: map[string]interface{}{
				"virtualCircuitId": 123.0,
			},
			expectedError: "must be a string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseOracleConfig(tt.config)
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, result)
				}
			}
		})
	}
}

func TestParseIBMConfig(t *testing.T) {
	tests := []struct {
		name          string
		config        map[string]interface{}
		expectedError string
		validate      func(*testing.T, *megaport.VXCPartnerConfigIBM)
	}{
		{
			name: "required fields only",
			config: map[string]interface{}{
				"accountID": "acct-123",
			},
			validate: func(t *testing.T, cfg *megaport.VXCPartnerConfigIBM) {
				assert.Equal(t, "IBM", cfg.ConnectType)
				assert.Equal(t, "acct-123", cfg.AccountID)
			},
		},
		{
			name: "all fields",
			config: map[string]interface{}{
				"accountID":         "acct-123",
				"customerASN":       65000.0,
				"customerIPAddress": "1.2.3.4",
				"providerIPAddress": "5.6.7.8",
				"name":              "my-ibm",
			},
			validate: func(t *testing.T, cfg *megaport.VXCPartnerConfigIBM) {
				assert.Equal(t, "IBM", cfg.ConnectType)
				assert.Equal(t, "acct-123", cfg.AccountID)
				assert.Equal(t, 65000, cfg.CustomerASN)
				assert.Equal(t, "1.2.3.4", cfg.CustomerIPAddress)
				assert.Equal(t, "5.6.7.8", cfg.ProviderIPAddress)
				assert.Equal(t, "my-ibm", cfg.Name)
			},
		},
		{
			name:          "missing accountID",
			config:        map[string]interface{}{},
			expectedError: "accountID",
		},
		{
			name: "empty accountID",
			config: map[string]interface{}{
				"accountID": "",
			},
			expectedError: "cannot be empty",
		},
		{
			name: "customerASN wrong type",
			config: map[string]interface{}{
				"accountID":   "acct-123",
				"customerASN": "string",
			},
			expectedError: "must be a number",
		},
		{
			name: "negative customerASN",
			config: map[string]interface{}{
				"accountID":   "acct-123",
				"customerASN": -1.0,
			},
			expectedError: "cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseIBMConfig(tt.config)
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, result)
				}
			}
		})
	}
}

func TestParseVRouterConfig(t *testing.T) {
	tests := []struct {
		name          string
		config        map[string]interface{}
		expectedError string
		validate      func(*testing.T, *megaport.VXCOrderVrouterPartnerConfig)
	}{
		{
			name: "no interfaces key",
			config: map[string]interface{}{
				"connectType": "VROUTER",
			},
			validate: func(t *testing.T, cfg *megaport.VXCOrderVrouterPartnerConfig) {
				assert.Empty(t, cfg.Interfaces)
			},
		},
		{
			name: "single interface",
			config: map[string]interface{}{
				"connectType": "VROUTER",
				"interfaces": []interface{}{
					map[string]interface{}{
						"vlan":        100.0,
						"ipAddresses": []interface{}{"10.0.0.1/30"},
					},
				},
			},
			validate: func(t *testing.T, cfg *megaport.VXCOrderVrouterPartnerConfig) {
				assert.Len(t, cfg.Interfaces, 1)
				iface := cfg.Interfaces[0]
				assert.Equal(t, 100, iface.VLAN)
				assert.Equal(t, []string{"10.0.0.1/30"}, iface.IpAddresses)
			},
		},
		{
			name: "interface with routes and bfd",
			config: map[string]interface{}{
				"connectType": "VROUTER",
				"interfaces": []interface{}{
					map[string]interface{}{
						"vlan":        100.0,
						"ipAddresses": []interface{}{"10.0.0.1/30"},
						"ipRoutes": []interface{}{
							map[string]interface{}{
								"prefix":      "10.1.0.0/24",
								"nextHop":     "10.0.0.2",
								"description": "route1",
							},
						},
						"bfd": map[string]interface{}{
							"txInterval": 300.0,
							"rxInterval": 300.0,
							"multiplier": 3.0,
						},
					},
				},
			},
			validate: func(t *testing.T, cfg *megaport.VXCOrderVrouterPartnerConfig) {
				assert.Len(t, cfg.Interfaces, 1)
				iface := cfg.Interfaces[0]
				assert.Equal(t, 100, iface.VLAN)
				assert.Equal(t, []string{"10.0.0.1/30"}, iface.IpAddresses)
				assert.Len(t, iface.IpRoutes, 1)
				route := iface.IpRoutes[0]
				assert.Equal(t, "10.1.0.0/24", route.Prefix)
				assert.Equal(t, "10.0.0.2", route.NextHop)
				assert.Equal(t, "route1", route.Description)
				assert.Equal(t, 300, iface.Bfd.TxInterval)
				assert.Equal(t, 300, iface.Bfd.RxInterval)
				assert.Equal(t, 3, iface.Bfd.Multiplier)
			},
		},
		{
			name: "interfaces not array",
			config: map[string]interface{}{
				"connectType": "VROUTER",
				"interfaces":  "string",
			},
			expectedError: "interfaces must be an array",
		},
		{
			name: "interface not object",
			config: map[string]interface{}{
				"connectType": "VROUTER",
				"interfaces":  []interface{}{123},
			},
			expectedError: "must be an object",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseVRouterConfig(tt.config)
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, result)
				}
			}
		})
	}
}

func TestParseVXCEndpointConfig(t *testing.T) {
	tests := []struct {
		name          string
		config        map[string]interface{}
		expectedError string
		validate      func(*testing.T, megaport.VXCOrderEndpointConfiguration)
	}{
		{
			name: "basic fields",
			config: map[string]interface{}{
				"productUID":    "uid-1",
				"vlan":          100.0,
				"diversityZone": "blue",
			},
			validate: func(t *testing.T, cfg megaport.VXCOrderEndpointConfiguration) {
				assert.Equal(t, "uid-1", cfg.ProductUID)
				assert.Equal(t, 100, cfg.VLAN)
				assert.Equal(t, "blue", cfg.DiversityZone)
			},
		},
		{
			name: "with partner config",
			config: map[string]interface{}{
				"productUID": "uid-1",
				"partnerConfig": map[string]interface{}{
					"connectType": "GOOGLE",
					"pairingKey":  "pk-1",
				},
			},
			validate: func(t *testing.T, cfg megaport.VXCOrderEndpointConfiguration) {
				assert.Equal(t, "uid-1", cfg.ProductUID)
				_, ok := cfg.PartnerConfig.(*megaport.VXCPartnerConfigGoogle)
				assert.True(t, ok, "expected *megaport.VXCPartnerConfigGoogle")
			},
		},
		{
			name: "with MVE config",
			config: map[string]interface{}{
				"innerVlan": 200.0,
				"vNicIndex": 1.0,
			},
			validate: func(t *testing.T, cfg megaport.VXCOrderEndpointConfiguration) {
				assert.NotNil(t, cfg.VXCOrderMVEConfig)
				assert.Equal(t, 200, cfg.InnerVLAN)
				assert.Equal(t, 1, cfg.NetworkInterfaceIndex)
			},
		},
		{
			name:   "empty map",
			config: map[string]interface{}{},
			validate: func(t *testing.T, cfg megaport.VXCOrderEndpointConfiguration) {
				assert.Equal(t, "", cfg.ProductUID)
				assert.Equal(t, 0, cfg.VLAN)
				assert.Nil(t, cfg.VXCOrderMVEConfig)
			},
		},
		{
			name: "invalid partner config",
			config: map[string]interface{}{
				"partnerConfig": map[string]interface{}{
					"ownerAccount": "123",
				},
			},
			expectedError: "partner config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseVXCEndpointConfig(tt.config, "test")
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, result)
				}
			}
		})
	}
}

func TestBuildVXCRequestFromJSON(t *testing.T) {
	tests := []struct {
		name          string
		jsonStr       string
		jsonFilePath  string
		expectedError string
		validate      func(*testing.T, *megaport.BuyVXCRequest)
	}{
		{
			name:    "valid full JSON",
			jsonStr: `{"portUid":"port-1","vxcName":"Test VXC","rateLimit":1000,"term":12,"promoCode":"PROMO","costCentre":"CC","aEndConfiguration":{"vlan":100},"bEndConfiguration":{"productUID":"port-2","vlan":200}}`,
			validate: func(t *testing.T, req *megaport.BuyVXCRequest) {
				assert.Equal(t, "port-1", req.PortUID)
				assert.Equal(t, "Test VXC", req.VXCName)
				assert.Equal(t, 1000, req.RateLimit)
				assert.Equal(t, 12, req.Term)
				assert.Equal(t, "PROMO", req.PromoCode)
				assert.Equal(t, "CC", req.CostCentre)
				assert.Equal(t, 100, req.AEndConfiguration.VLAN)
				assert.Equal(t, "port-2", req.BEndConfiguration.ProductUID)
				assert.Equal(t, 200, req.BEndConfiguration.VLAN)
			},
		},
		{
			name:    "with resource tags",
			jsonStr: `{"portUid":"port-1","vxcName":"Test VXC","rateLimit":1000,"term":12,"resourceTags":{"env":"prod"},"bEndConfiguration":{"productUID":"port-2"}}`,
			validate: func(t *testing.T, req *megaport.BuyVXCRequest) {
				assert.NotNil(t, req.ResourceTags)
				assert.Equal(t, "prod", req.ResourceTags["env"])
			},
		},
		{
			name:          "missing portUid",
			jsonStr:       `{"vxcName":"Test VXC","rateLimit":1000,"term":12}`,
			expectedError: "portUid",
		},
		{
			name:          "invalid JSON",
			jsonStr:       `{bad}`,
			expectedError: "error parsing JSON",
		},
		{
			name:          "empty inputs",
			jsonStr:       "",
			jsonFilePath:  "",
			expectedError: "either json or json-file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := buildVXCRequestFromJSON(tt.jsonStr, tt.jsonFilePath)
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, result)
				}
			}
		})
	}
}

func TestBuildUpdateVXCRequestFromJSON_PartnerConfigs(t *testing.T) {
	tests := []struct {
		name          string
		json          string
		expectedError string
		validate      func(*testing.T, *megaport.UpdateVXCRequest)
	}{
		{
			name: "valid VRouter A-End partner config",
			json: `{"aEndPartnerConfig":{"connectType":"VROUTER","interfaces":[{"vlan":100,"ipAddresses":["10.0.0.1/30"]}]}}`,
			validate: func(t *testing.T, req *megaport.UpdateVXCRequest) {
				assert.NotNil(t, req.AEndPartnerConfig)
				vrouterCfg, ok := req.AEndPartnerConfig.(*megaport.VXCOrderVrouterPartnerConfig)
				assert.True(t, ok, "expected *megaport.VXCOrderVrouterPartnerConfig")
				assert.Len(t, vrouterCfg.Interfaces, 1)
			},
		},
		{
			name:          "non-VRouter A-End partner config",
			json:          `{"aEndPartnerConfig":{"connectType":"AWS","ownerAccount":"123"}}`,
			expectedError: "only VRouter",
		},
		{
			name:          "non-VRouter B-End partner config",
			json:          `{"bEndPartnerConfig":{"connectType":"IBM","accountID":"123"}}`,
			expectedError: "only VRouter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := buildUpdateVXCRequestFromJSON(tt.json, "")
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, result)
				}
			}
		})
	}
}
