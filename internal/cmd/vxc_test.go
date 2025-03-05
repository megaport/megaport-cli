package cmd

import (
	"context"
	"fmt"
	"testing"

	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
)

var testVXCs = []*megaport.VXC{
	{
		UID:  "vxc-1",
		Name: "MyVXCOne",
		AEndConfiguration: megaport.VXCEndConfiguration{
			UID: "a-end-1",
		},
		BEndConfiguration: megaport.VXCEndConfiguration{
			UID: "b-end-1",
		},
	},
	{
		UID:  "vxc-2",
		Name: "AnotherVXC",
		AEndConfiguration: megaport.VXCEndConfiguration{
			UID: "a-end-2",
		},
		BEndConfiguration: megaport.VXCEndConfiguration{
			UID: "b-end-2",
		},
	},
}

func TestPrintVXCs_Table(t *testing.T) {
	output := captureOutput(func() {
		err := printVXCs(testVXCs, "table")
		assert.NoError(t, err)
	})

	expected := `uid     name         a_end_uid   b_end_uid
vxc-1   MyVXCOne     a-end-1     b-end-1
vxc-2   AnotherVXC   a-end-2     b-end-2
`
	assert.Equal(t, expected, output)
}
func TestPrintVXCs_JSON(t *testing.T) {
	output := captureOutput(func() {
		err := printVXCs(testVXCs, "json")
		assert.NoError(t, err)
	})

	expected := `[
  {
    "uid": "vxc-1",
    "name": "MyVXCOne",
    "a_end_uid": "a-end-1",
    "b_end_uid": "b-end-1"
  },
  {
    "uid": "vxc-2",
    "name": "AnotherVXC",
    "a_end_uid": "a-end-2",
    "b_end_uid": "b-end-2"
  }
]`
	assert.JSONEq(t, expected, output)
}

func TestPrintVXCs_CSV(t *testing.T) {
	output := captureOutput(func() {
		err := printVXCs(testVXCs, "csv")
		assert.NoError(t, err)
	})

	expected := `uid,name,a_end_uid,b_end_uid
vxc-1,MyVXCOne,a-end-1,b-end-1
vxc-2,AnotherVXC,a-end-2,b-end-2
`
	assert.Equal(t, expected, output)
}

func TestPrintVXCs_Invalid(t *testing.T) {
	var err error
	output := captureOutput(func() {
		err = printVXCs(testVXCs, "invalid")
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid output format")
	assert.Empty(t, output)
}
func TestPrintVXCs_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		vxcs        []*megaport.VXC
		format      string
		shouldError bool
		expected    string
	}{
		{
			name:        "nil slice",
			vxcs:        nil,
			format:      "table",
			shouldError: false,
			expected:    "uid   name   a_end_uid   b_end_uid\n",
		},
		{
			name:        "empty slice",
			vxcs:        []*megaport.VXC{},
			format:      "json",
			shouldError: false,
			expected:    "[]",
		},
		{
			name: "nil vxc in slice",
			vxcs: []*megaport.VXC{
				nil,
				{
					UID:  "vxc-1",
					Name: "TestVXC",
				},
			},
			format:      "table",
			shouldError: true,
			expected:    "invalid VXC: nil value",
		},
		{
			name: "nil end configurations",
			vxcs: []*megaport.VXC{
				{
					UID:  "vxc-1",
					Name: "TestVXC",
				},
			},
			format:      "csv",
			shouldError: false,
			expected:    "uid,name,a_end_uid,b_end_uid\nvxc-1,TestVXC,,\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output string
			var err error

			output = captureOutput(func() {
				err = printVXCs(tt.vxcs, tt.format)
			})

			if tt.shouldError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expected)
				assert.Empty(t, output)
			} else {
				assert.NoError(t, err)
				switch tt.format {
				case "json":
					assert.JSONEq(t, tt.expected, output)
				case "table", "csv":
					assert.Equal(t, tt.expected, output)
				}
			}
		})
	}
}

func TestToVXCOutput_EdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		vxc           *megaport.VXC
		shouldError   bool
		errorContains string
		validateFunc  func(*testing.T, VXCOutput)
	}{
		{
			name:          "nil vxc",
			vxc:           nil,
			shouldError:   true,
			errorContains: "invalid VXC: nil value",
		},
		{
			name: "zero values",
			vxc:  &megaport.VXC{},
			validateFunc: func(t *testing.T, output VXCOutput) {
				assert.Empty(t, output.UID)
				assert.Empty(t, output.Name)
				assert.Empty(t, output.AEndUID)
				assert.Empty(t, output.BEndUID)
			},
		},
		{
			name: "nil end configurations",
			vxc: &megaport.VXC{
				UID:  "vxc-1",
				Name: "TestVXC",
			},
			validateFunc: func(t *testing.T, output VXCOutput) {
				assert.Equal(t, "vxc-1", output.UID)
				assert.Equal(t, "TestVXC", output.Name)
				assert.Empty(t, output.AEndUID)
				assert.Empty(t, output.BEndUID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := ToVXCOutput(tt.vxc)

			if tt.shouldError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				if tt.validateFunc != nil {
					tt.validateFunc(t, output)
				}
			}
		})
	}
}

func TestBuyVXC(t *testing.T) {
	originalPrompt := prompt
	originalLoginFunc := loginFunc

	// Restore the originals after test completes
	defer func() {
		prompt = originalPrompt
		loginFunc = originalLoginFunc
	}()

	tests := []struct {
		name          string
		prompts       []string
		expectedError string
		setupMock     func(*testing.T, *mockVXCService)
	}{
		{
			name: "successful VXC purchase",
			prompts: []string{
				"a-end-uid", // A-End Product UID
				"b-end-uid", // B-End Product UID
				"Test VXC",  // VXC name
				"100",       // Rate limit
				"12",        // Term
				"100",       // A-End VLAN
				"200",       // A-End Inner VLAN
				"0",         // A-End vNIC index
				"300",       // B-End VLAN
				"400",       // B-End Inner VLAN
				"1",         // B-End vNIC index
				"",          // Promo code
				"",          // Service key
				"",          // Cost centre
				"",          // A-End partner config
				"",          // B-End partner config
			},
			expectedError: "",
			setupMock: func(t *testing.T, m *mockVXCService) {
				m.buyVXCResponse = &megaport.BuyVXCResponse{
					TechnicalServiceUID: "vxc-sample-uid",
				}
			},
		},
		{
			name: "missing A-End Product UID",
			prompts: []string{
				"",          // A-End Product UID (missing)
				"b-end-uid", // B-End Product UID
				"Test VXC",  // VXC name
				"100",       // Rate limit
				"12",        // Term
				"100",       // A-End VLAN
				"200",       // A-End Inner VLAN
				"0",         // A-End vNIC index
				"300",       // B-End VLAN
				"400",       // B-End Inner VLAN
				"1",         // B-End vNIC index
				"",          // Promo code
				"",          // Service key
				"",          // Cost centre
				"",          // A-End partner config
				"",          // B-End partner config
			},
			expectedError: "A-End Product UID is required",
			setupMock: func(t *testing.T, m *mockVXCService) {
				// Should not call Login if required fields are missing
			},
		},
		{
			name: "invalid rate limit",
			prompts: []string{
				"a-end-uid", // A-End Product UID
				"b-end-uid", // B-End Product UID
				"Test VXC",  // VXC name
				"invalid",   // Bad rate limit
			},
			expectedError: "invalid rate limit",
			setupMock: func(t *testing.T, m *mockVXCService) {
				// Should not call Login if rate limit is invalid
			},
		},
		{
			name: "AWS partner config",
			prompts: []string{
				"a-end-uid",       // A-End Product UID
				"b-end-uid",       // B-End Product UID
				"AWS VXC",         // VXC name
				"200",             // Rate limit
				"12",              // Term
				"150",             // A-End VLAN
				"250",             // A-End Inner VLAN
				"0",               // A-End vNIC index
				"350",             // B-End VLAN
				"450",             // B-End Inner VLAN
				"1",               // B-End vNIC index
				"",                // Promo code
				"",                // Service key
				"",                // Cost centre
				"aws",             // A-End partner config
				"direct",          // Connect type
				"account-id",      // AWS Account ID
				"12345",           // ASN
				"67890",           // Amazon ASN
				"auth-key",        // Auth key
				"192.168.1.0/24",  // Prefixes
				"192.168.1.1",     // Customer IP address
				"192.168.1.2",     // Amazon IP address
				"connection-name", // Connection name
				"",                // B-End partner config
			},
			expectedError: "",
			setupMock: func(t *testing.T, m *mockVXCService) {
				m.buyVXCResponse = &megaport.BuyVXCResponse{
					TechnicalServiceUID: "vxc-aws-uid",
				}
			},
		},
		{
			name: "Azure partner config",
			prompts: []string{
				"a-end-uid",         // A-End Product UID
				"b-end-uid",         // B-End Product UID
				"Azure VXC",         // VXC name
				"300",               // Rate limit
				"12",                // Term
				"160",               // A-End VLAN
				"260",               // A-End Inner VLAN
				"1",                 // A-End vNIC index
				"360",               // B-End VLAN
				"460",               // B-End Inner VLAN
				"2",                 // B-End vNIC index
				"",                  // Promo code
				"",                  // Service key
				"",                  // Cost centre
				"azure",             // A-End partner config
				"azure-service-key", // Azure Service Key
				"yes",               // Add a peering config?
				"type1",             // Peering type
				"peer-asn",          // Peer ASN
				"primary-subnet",    // Primary Subnet
				"secondary-subnet",  // Secondary Subnet
				"prefixes",          // Prefixes
				"shared-key",        // Shared Key
				"100",               // VLAN
				"no",                // Add another peering config?
				"",                  // B-End partner config
			},
			expectedError: "",
			setupMock: func(t *testing.T, m *mockVXCService) {
				m.buyVXCResponse = &megaport.BuyVXCResponse{
					TechnicalServiceUID: "vxc-azure-uid",
				}
			},
		},
		{
			name: "VRouter partner config",
			prompts: []string{
				"a-end-uid",              // A-End Product UID
				"b-end-uid",              // B-End Product UID
				"VRouter BGP Connection", // VXC name
				"500",                    // Rate limit
				"12",                     // Term
				"170",                    // A-End VLAN
				"270",                    // A-End Inner VLAN
				"0",                      // A-End vNIC index
				"370",                    // B-End VLAN
				"470",                    // B-End Inner VLAN
				"0",                      // B-End vNIC index
				"",                       // Promo code
				"",                       // Service key
				"",                       // Cost centre
				"vrouter",                // A-End partner config type
				"yes",                    // Add an interface?
				"100",                    // VLAN
				"192.168.1.1/24",         // IP address
				"",                       // Finish IP addresses
				"10.0.0.0/16",            // IP route prefix
				"description",            // IP route description
				"192.168.1.1",            // IP route next hop
				"",                       // Finish IP routes
				"10.0.0.1",               // NAT IP address
				"",                       // Finish NAT IP addresses
				"true",                   // Enable BFD?
				"300",                    // BFD TxInterval
				"300",                    // BFD RxInterval
				"3",                      // BFD Multiplier
				"yes",                    // Add a BGP connection?
				"65000",                  // Peer ASN
				"12345",                  // Local ASN
				"192.168.1.1",            // Local IP Address
				"192.168.1.2",            // Peer IP Address
				"password",               // Password
				"false",                  // Shutdown
				"description",            // Description
				"100",                    // MED In
				"200",                    // MED Out
				"true",                   // BFD Enabled
				"export-policy",          // Export Policy
				"permit1,permit2",        // Permit Export To
				"deny1,deny2",            // Deny Export To
				"10",                     // Import Whitelist
				"20",                     // Import Blacklist
				"30",                     // Export Whitelist
				"40",                     // Export Blacklist
				"5",                      // AS Path Prepend Count
				"NON_CLOUD",              // Peer Type
				"no",                     // Finish BGP connections
				"no",                     // Add another interface?
				"",                       // B-End partner config
			},
			expectedError: "",
			setupMock: func(t *testing.T, m *mockVXCService) {
				m.buyVXCResponse = &megaport.BuyVXCResponse{
					TechnicalServiceUID: "vxc-vrouter-uid",
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockVXCService{}
			if tt.setupMock != nil {
				tt.setupMock(t, mockService)
			}

			// Override the loginFunc with a test double
			loginFunc = func(ctx context.Context) (*megaport.Client, error) {
				if mockService == nil {
					t.Fatal("mockService is nil")
				}
				return &megaport.Client{
					VXCService: mockService,
				}, nil
			}

			// Override the prompt responses
			promptIndex := 0
			promptResponses := tt.prompts
			prompt = func(_ string) (string, error) {
				if promptIndex < len(promptResponses) {
					resp := promptResponses[promptIndex]
					promptIndex++
					return resp, nil
				}
				// If test tries to prompt again beyond expected answers
				return "", fmt.Errorf("unexpected additional prompt")
			}

			// Execute the BuyVXC command
			cmd := buyVXCCmd
			cmd.SetArgs([]string{}) // No command-line args
			err := BuyVXC(cmd, []string{})

			// Check results
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			// Run post-execution validation if provided
			if mockService.postExecutionCheck != nil {
				mockService.postExecutionCheck()
			}
		})
	}
}
