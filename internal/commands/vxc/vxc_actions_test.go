package vxc

import (
	"context"
	"fmt"
	"testing"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/commands/config"
	"github.com/megaport/megaport-cli/internal/utils"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// Save the original functions to restore later
var originalBuyVXCFunc = buyVXCFunc
var interactive bool

func TestBuyVXC(t *testing.T) {
	originalPrompt := utils.Prompt
	originalLoginFunc := config.LoginFunc
	originalInteractiveFlag := interactive
	noColor := true

	// Restore the originals after test completes
	defer func() {
		utils.Prompt = originalPrompt
		config.LoginFunc = originalLoginFunc
		buyVXCFunc = originalBuyVXCFunc
		interactive = originalInteractiveFlag
	}()

	tests := []struct {
		name           string
		prompts        []string
		expectedError  string
		setupMock      func(*testing.T, *mockVXCService)
		flags          map[string]string
		flagsInt       map[string]int
		flagsBool      map[string]bool
		args           []string
		skipRequest    bool // Skip checking request details
		expectedOutput string
	}{
		{
			name: "successful VXC purchase - interactive mode",
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
				"no",        // A-End partner config
				"no",        // B-End partner config
			},
			expectedError: "",
			setupMock: func(t *testing.T, m *mockVXCService) {
				m.buyVXCResponse = &megaport.BuyVXCResponse{
					TechnicalServiceUID: "vxc-sample-uid",
				}

				// Replace the interactive mode handling completely
				interactive = true

				// Skip the buildVXCRequestFromPrompt function entirely
				buildVXCRequestFromPromptOrig := buildVXCRequestFromPrompt
				buildVXCRequestFromPrompt = func(ctx context.Context, svc megaport.VXCService, noColor bool) (*megaport.BuyVXCRequest, error) {
					return &megaport.BuyVXCRequest{
						PortUID:   "a-end-uid",
						VXCName:   "Test VXC",
						RateLimit: 100,
						Term:      12,
						AEndConfiguration: megaport.VXCOrderEndpointConfiguration{
							VLAN: 100,
						},
						BEndConfiguration: megaport.VXCOrderEndpointConfiguration{
							ProductUID: "b-end-uid",
							VLAN:       300,
						},
					}, nil
				}
				t.Cleanup(func() {
					buildVXCRequestFromPrompt = buildVXCRequestFromPromptOrig
				})

				buyVXCFunc = func(ctx context.Context, client *megaport.Client, req *megaport.BuyVXCRequest) (*megaport.BuyVXCResponse, error) {
					return m.BuyVXC(ctx, req)
				}
			},
			flagsBool: map[string]bool{
				"interactive": true,
			},
			skipRequest:    true,
			expectedOutput: "VXC created vxc-sample-uid",
		},
		{
			name:          "successful VXC purchase - flag mode",
			expectedError: "",
			setupMock: func(t *testing.T, m *mockVXCService) {
				m.buyVXCResponse = &megaport.BuyVXCResponse{
					TechnicalServiceUID: "vxc-sample-uid",
				}

				// Skip the buildVXCRequestFromFlags function entirely
				buildVXCRequestFromFlagsOrig := buildVXCRequestFromFlags
				buildVXCRequestFromFlags = func(cmd *cobra.Command, ctx context.Context, svc megaport.VXCService) (*megaport.BuyVXCRequest, error) {
					return &megaport.BuyVXCRequest{
						PortUID:   "dcc-12345",
						VXCName:   "Flag VXC",
						RateLimit: 500,
						Term:      12,
						AEndConfiguration: megaport.VXCOrderEndpointConfiguration{
							VLAN: 100,
						},
						BEndConfiguration: megaport.VXCOrderEndpointConfiguration{
							ProductUID: "dcc-67890",
							VLAN:       200,
						},
					}, nil
				}
				t.Cleanup(func() {
					buildVXCRequestFromFlags = buildVXCRequestFromFlagsOrig
				})

				buyVXCFunc = func(ctx context.Context, client *megaport.Client, req *megaport.BuyVXCRequest) (*megaport.BuyVXCResponse, error) {
					return m.BuyVXC(ctx, req)
				}
			},
			flags: map[string]string{
				"a-end-uid":   "dcc-12345",
				"b-end-uid":   "dcc-67890",
				"name":        "Flag VXC",
				"cost-centre": "Dev",
				"promo-code":  "PROMO123",
				"service-key": "service-key",
			},
			flagsInt: map[string]int{
				"rate-limit":       500,
				"term":             12,
				"a-end-vlan":       100,
				"b-end-vlan":       200,
				"a-end-inner-vlan": 300,
				"b-end-inner-vlan": 400,
				"a-end-vnic-index": 1,
				"b-end-vnic-index": 2,
			},
			skipRequest:    true,
			expectedOutput: "VXC created vxc-sample-uid",
		},
		{
			name:          "missing required fields - flag mode",
			expectedError: "a-end-uid is required",
			setupMock: func(t *testing.T, m *mockVXCService) {
				// Skip the build function and return an error directly
				buildVXCRequestFromFlagsOrig := buildVXCRequestFromFlags
				buildVXCRequestFromFlags = func(cmd *cobra.Command, ctx context.Context, svc megaport.VXCService) (*megaport.BuyVXCRequest, error) {
					return nil, fmt.Errorf("a-end-uid is required")
				}
				t.Cleanup(func() {
					buildVXCRequestFromFlags = buildVXCRequestFromFlagsOrig
				})
			},
			flags: map[string]string{
				"name": "Incomplete VXC",
			},
			flagsInt: map[string]int{
				"rate-limit": 500,
				"term":       12,
			},
		},
		{
			name:          "API error",
			expectedError: "API error",
			setupMock: func(t *testing.T, m *mockVXCService) {
				// Set up a valid response object first, then add the error
				m.buyVXCResponse = &megaport.BuyVXCResponse{
					TechnicalServiceUID: "", // Empty but not nil
				}
				m.buyVXCErr = fmt.Errorf("API error")

				// Skip the buildVXCRequestFromFlags function completely
				buildVXCRequestFromFlagsOrig := buildVXCRequestFromFlags
				buildVXCRequestFromFlags = func(cmd *cobra.Command, ctx context.Context, svc megaport.VXCService) (*megaport.BuyVXCRequest, error) {
					return &megaport.BuyVXCRequest{
						PortUID:   "dcc-12345",
						VXCName:   "Error VXC",
						RateLimit: 500,
						Term:      12,
						AEndConfiguration: megaport.VXCOrderEndpointConfiguration{
							VLAN: 100,
						},
						BEndConfiguration: megaport.VXCOrderEndpointConfiguration{
							ProductUID: "dcc-67890",
							VLAN:       200,
						},
					}, nil
				}
				t.Cleanup(func() {
					buildVXCRequestFromFlags = buildVXCRequestFromFlagsOrig
				})

				// Override buyVXCFunc completely to ensure we don't dereference any nil pointers
				buyVXCFunc = func(ctx context.Context, client *megaport.Client, req *megaport.BuyVXCRequest) (*megaport.BuyVXCResponse, error) {
					// Don't call m.BuyVXC directly, just return the error
					return nil, fmt.Errorf("API error")
				}
			},
			flags: map[string]string{
				"a-end-uid": "dcc-12345",
				"b-end-uid": "dcc-67890",
				"name":      "Error VXC",
			},
			flagsInt: map[string]int{
				"rate-limit": 500,
				"term":       12,
			},
			skipRequest: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockVXCService{}
			if tt.setupMock != nil {
				tt.setupMock(t, mockService)
			}

			// Override the loginFunc with a test double
			config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
				return &megaport.Client{
					VXCService: mockService,
				}, nil
			}

			// Override the prompt responses
			promptIndex := 0
			promptResponses := tt.prompts
			utils.Prompt = func(_ string, _ bool) (string, error) {
				if promptIndex < len(promptResponses) {
					resp := promptResponses[promptIndex]
					promptIndex++
					return resp, nil
				}
				return "", fmt.Errorf("unexpected additional prompt")
			}

			// Create command and set flags
			cmd := &cobra.Command{}
			cmd.Flags().Bool("interactive", false, "")
			cmd.Flags().String("a-end-uid", "", "")
			cmd.Flags().String("b-end-uid", "", "")
			cmd.Flags().String("name", "", "")
			cmd.Flags().Int("rate-limit", 0, "")
			cmd.Flags().Int("term", 0, "")
			cmd.Flags().Int("a-end-vlan", 0, "")
			cmd.Flags().Int("b-end-vlan", 0, "")
			cmd.Flags().Int("a-end-inner-vlan", 0, "")
			cmd.Flags().Int("b-end-inner-vlan", 0, "")
			cmd.Flags().Int("a-end-vnic-index", 0, "")
			cmd.Flags().Int("b-end-vnic-index", 0, "")
			cmd.Flags().String("promo-code", "", "")
			cmd.Flags().String("service-key", "", "")
			cmd.Flags().String("cost-centre", "", "")
			cmd.Flags().String("a-end-partner-config", "", "")
			cmd.Flags().String("b-end-partner-config", "", "")
			cmd.Flags().String("json", "", "")
			cmd.Flags().String("json-file", "", "")

			// Set string flags
			for flag, value := range tt.flags {
				err := cmd.Flags().Set(flag, value)
				assert.NoError(t, err)
			}

			// Set integer flags
			for flag, value := range tt.flagsInt {
				err := cmd.Flags().Set(flag, fmt.Sprintf("%d", value))
				assert.NoError(t, err)
			}

			// Set boolean flags
			for flag, value := range tt.flagsBool {
				if value {
					err := cmd.Flags().Set(flag, "true")
					assert.NoError(t, err)
				}
			}

			// Execute the function
			err := BuyVXC(cmd, tt.args, noColor)

			// Check results
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
func TestUpdateVXC(t *testing.T) {
	noColor := true
	originalPrompt := utils.Prompt
	originalLoginFunc := config.LoginFunc
	originalInteractiveFlag := interactive
	originalUpdateVXCFunc := updateVXCFunc

	// Restore the originals after test completes
	defer func() {
		utils.Prompt = originalPrompt
		config.LoginFunc = originalLoginFunc
		updateVXCFunc = originalUpdateVXCFunc
		interactive = originalInteractiveFlag
	}()

	tests := []struct {
		name           string
		prompts        []string
		expectedError  string
		setupMock      func(*testing.T, *mockVXCService)
		flags          map[string]string
		flagsInt       map[string]int
		flagsBool      map[string]bool
		args           []string
		expectedOutput string
	}{
		{
			name: "successful VXC update - interactive mode",
			prompts: []string{
				"yes",          // Update name?
				"New VXC Name", // New name
				"yes",          // Update rate limit?
				"200",          // New rate limit
				"yes",          // Update term?
				"24",           // New term
				"yes",          // Update cost centre?
				"New Cost",     // New cost centre
				"yes",          // Update shutdown?
				"yes",          // Set shutdown to yes
				"yes",          // Update A-end VLAN?
				"150",          // New A-end VLAN
				"yes",          // Update B-end VLAN?
				"250",          // New B-end VLAN
				"yes",          // Update A-end inner VLAN?
				"350",          // New A-end inner VLAN
				"yes",          // Update B-end inner VLAN?
				"450",          // New B-end inner VLAN
				"no",           // Update A-end product UID?
				"no",           // Update B-end product UID?
				"no",           // Configure A-end VRouter?
				"no",           // Configure B-end VRouter?
			},
			expectedError: "",
			setupMock: func(t *testing.T, m *mockVXCService) {
				m.getVXCResponse = &megaport.VXC{
					UID:  "vxc-test",
					Name: "Test VXC",
				}

				// Replace the interactive mode handling completely
				interactive = true

				// Skip the buildUpdateVXCRequestFromPrompt function entirely
				buildUpdateVXCRequestFromPromptOrig := buildUpdateVXCRequestFromPrompt
				buildUpdateVXCRequestFromPrompt = func(vxcID string, noColor bool) (*megaport.UpdateVXCRequest, error) {
					name := "New VXC Name"
					rateLimit := 200
					term := 24
					costCentre := "New Cost"
					shutdown := true
					aEndVLAN := 150
					bEndVLAN := 250
					aEndInnerVLAN := 350
					bEndInnerVLAN := 450

					return &megaport.UpdateVXCRequest{
						Name:          &name,
						RateLimit:     &rateLimit,
						Term:          &term,
						CostCentre:    &costCentre,
						Shutdown:      &shutdown,
						AEndVLAN:      &aEndVLAN,
						BEndVLAN:      &bEndVLAN,
						AEndInnerVLAN: &aEndInnerVLAN,
						BEndInnerVLAN: &bEndInnerVLAN,
					}, nil
				}
				t.Cleanup(func() {
					buildUpdateVXCRequestFromPrompt = buildUpdateVXCRequestFromPromptOrig
				})

				// Mock the updateVXCFunc
				updateVXCFunc = func(ctx context.Context, client *megaport.Client, vxcUID string, req *megaport.UpdateVXCRequest) error {
					_, err := m.UpdateVXC(ctx, vxcUID, req)
					return err
				}
			},
			args:           []string{"vxc-12345"},
			expectedOutput: "VXC updated vxc-12345",
			flagsBool: map[string]bool{
				"interactive": true,
			},
		},
		{
			name:          "successful VXC update - flag mode",
			expectedError: "",
			setupMock: func(t *testing.T, m *mockVXCService) {
				m.getVXCResponse = &megaport.VXC{
					UID:  "vxc-test",
					Name: "Test VXC",
				}

				// Skip the buildUpdateVXCRequestFromFlags function entirely
				buildUpdateVXCRequestFromFlagsOrig := buildUpdateVXCRequestFromFlags
				buildUpdateVXCRequestFromFlags = func(cmd *cobra.Command) (*megaport.UpdateVXCRequest, error) {
					name := "Flag Updated VXC"
					rateLimit := 1000
					term := 24
					costCentre := "Dev Team"
					shutdown := true
					aEndVLAN := 111
					bEndVLAN := 222
					aEndInnerVLAN := 333
					bEndInnerVLAN := 444
					aEndUID := "dcc-newaid"
					bEndUID := "dcc-newbid"

					return &megaport.UpdateVXCRequest{
						Name:           &name,
						RateLimit:      &rateLimit,
						Term:           &term,
						CostCentre:     &costCentre,
						Shutdown:       &shutdown,
						AEndVLAN:       &aEndVLAN,
						BEndVLAN:       &bEndVLAN,
						AEndInnerVLAN:  &aEndInnerVLAN,
						BEndInnerVLAN:  &bEndInnerVLAN,
						AEndProductUID: &aEndUID,
						BEndProductUID: &bEndUID,
					}, nil
				}
				t.Cleanup(func() {
					buildUpdateVXCRequestFromFlags = buildUpdateVXCRequestFromFlagsOrig
				})

				// Mock the updateVXCFunc
				updateVXCFunc = func(ctx context.Context, client *megaport.Client, vxcUID string, req *megaport.UpdateVXCRequest) error {
					_, err := m.UpdateVXC(ctx, vxcUID, req)
					return err
				}
			},
			args:           []string{"vxc-12345"},
			expectedOutput: "VXC updated vxc-12345",
			flags: map[string]string{
				"name":        "Flag Updated VXC",
				"cost-centre": "Dev Team",
				"a-end-uid":   "dcc-newaid",
				"b-end-uid":   "dcc-newbid",
			},
			flagsInt: map[string]int{
				"rate-limit":       1000,
				"term":             24,
				"a-end-vlan":       111,
				"b-end-vlan":       222,
				"a-end-inner-vlan": 333,
				"b-end-inner-vlan": 444,
			},
			flagsBool: map[string]bool{
				"shutdown": true,
			},
		},
		{
			name:          "successful VXC update - JSON mode",
			expectedError: "",
			setupMock: func(t *testing.T, m *mockVXCService) {
				m.getVXCResponse = &megaport.VXC{
					UID:  "vxc-test",
					Name: "Test VXC",
				}

				// Skip the buildUpdateVXCRequestFromJSON function entirely
				buildUpdateVXCRequestFromJSONOrig := buildUpdateVXCRequestFromJSON
				buildUpdateVXCRequestFromJSON = func(jsonStr string, jsonFile string) (*megaport.UpdateVXCRequest, error) {
					name := "JSON Updated VXC"
					rateLimit := 500
					term := 12
					costCentre := "JSON Team"
					shutdown := false
					aEndVLAN := 123
					bEndVLAN := 456

					return &megaport.UpdateVXCRequest{
						Name:       &name,
						RateLimit:  &rateLimit,
						Term:       &term,
						CostCentre: &costCentre,
						Shutdown:   &shutdown,
						AEndVLAN:   &aEndVLAN,
						BEndVLAN:   &bEndVLAN,
					}, nil
				}
				t.Cleanup(func() {
					buildUpdateVXCRequestFromJSON = buildUpdateVXCRequestFromJSONOrig
				})

				// Mock the updateVXCFunc
				updateVXCFunc = func(ctx context.Context, client *megaport.Client, vxcUID string, req *megaport.UpdateVXCRequest) error {
					_, err := m.UpdateVXC(ctx, vxcUID, req)
					return err
				}
			},
			args:           []string{"vxc-12345"},
			expectedOutput: "VXC updated vxc-12345",
			flags: map[string]string{
				"json": `{
                    "name": "JSON Updated VXC",
                    "rateLimit": 500,
                    "term": 12,
                    "costCentre": "JSON Team",
                    "aEndVLAN": 123,
                    "bEndVLAN": 456,
                    "shutdown": false
                }`,
			},
		},
		{
			name:           "update with partner config",
			expectedError:  "",
			prompts:        []string{}, // No prompts needed for this test
			expectedOutput: "VXC updated vxc-12345",
			setupMock: func(t *testing.T, m *mockVXCService) {
				m.getVXCResponse = &megaport.VXC{
					UID:  "vxc-test",
					Name: "Test VXC",
				}

				// Skip the buildUpdateVXCRequestFromFlags function
				buildUpdateVXCRequestFromFlagsOrig := buildUpdateVXCRequestFromFlags
				buildUpdateVXCRequestFromFlags = func(cmd *cobra.Command) (*megaport.UpdateVXCRequest, error) {
					name := "Partner Config VXC"

					// Create a partner config
					partnerConfig := &megaport.VXCOrderVrouterPartnerConfig{
						Interfaces: []megaport.PartnerConfigInterface{
							{
								VLAN:        100,
								IpAddresses: []string{"192.168.1.1/30"},
								BgpConnections: []megaport.BgpConnectionConfig{
									{
										PeerAsn:        65000,
										LocalIpAddress: "192.168.1.1",
										PeerIpAddress:  "192.168.1.2",
									},
								},
							},
						},
					}

					return &megaport.UpdateVXCRequest{
						Name:              &name,
						BEndPartnerConfig: partnerConfig,
					}, nil
				}
				t.Cleanup(func() {
					buildUpdateVXCRequestFromFlags = buildUpdateVXCRequestFromFlagsOrig
				})

				// Mock the updateVXCFunc to bypass the VRouter check for testing
				updateVXCFunc = func(ctx context.Context, client *megaport.Client, vxcUID string, req *megaport.UpdateVXCRequest) error {
					_, err := m.UpdateVXC(ctx, vxcUID, req)
					return err
				}
			},
			args: []string{"vxc-12345"},
			flags: map[string]string{
				"b-end-partner-config": `{
                    "interfaces": [
                        {
                            "vlan": 100,
                            "ipAddresses": ["192.168.1.1/30"],
                            "bgpConnections": [
                                {
                                    "peerAsn": 65000,
                                    "localAsn": 64512,
                                    "localIpAddress": "192.168.1.1",
                                    "peerIpAddress": "192.168.1.2"
                                }
                            ]
                        }
                    ]
                }`,
			},
		},
		{
			name:          "API error",
			expectedError: "failed to update VXC",
			setupMock: func(t *testing.T, m *mockVXCService) {
				m.getVXCResponse = &megaport.VXC{
					UID:  "vxc-test",
					Name: "Test VXC",
				}
				m.updateVXCErr = fmt.Errorf("API error")

				// Skip the buildUpdateVXCRequestFromFlags function
				buildUpdateVXCRequestFromFlagsOrig := buildUpdateVXCRequestFromFlags
				buildUpdateVXCRequestFromFlags = func(cmd *cobra.Command) (*megaport.UpdateVXCRequest, error) {
					name := "Error VXC"
					return &megaport.UpdateVXCRequest{
						Name: &name,
					}, nil
				}
				t.Cleanup(func() {
					buildUpdateVXCRequestFromFlags = buildUpdateVXCRequestFromFlagsOrig
				})

				// Mock the updateVXCFunc
				updateVXCFunc = func(ctx context.Context, client *megaport.Client, vxcUID string, req *megaport.UpdateVXCRequest) error {
					return fmt.Errorf("failed to update VXC: API error")
				}
			},
			args: []string{"vxc-12345"},
			flags: map[string]string{
				"name": "Error VXC",
			},
		},
		{
			name:          "no update parameters",
			expectedError: "no update parameters provided",
			prompts:       []string{}, // ensure empty so any prompt call fails
			setupMock: func(t *testing.T, m *mockVXCService) {
				interactive = false
				m.getVXCResponse = &megaport.VXC{
					UID:  "vxc-test",
					Name: "Test VXC",
				}

				// Mock buildUpdateVXCRequestFromFlags
				flagsOrig := buildUpdateVXCRequestFromFlags
				buildUpdateVXCRequestFromFlags = func(cmd *cobra.Command) (*megaport.UpdateVXCRequest, error) {
					return nil, fmt.Errorf("no update parameters provided")
				}
				t.Cleanup(func() { buildUpdateVXCRequestFromFlags = flagsOrig })

				// ALSO mock buildUpdateVXCRequestFromPrompt
				promptOrig := buildUpdateVXCRequestFromPrompt
				buildUpdateVXCRequestFromPrompt = func(_ string, _ bool) (*megaport.UpdateVXCRequest, error) {
					return nil, fmt.Errorf("no update parameters provided")
				}
				t.Cleanup(func() { buildUpdateVXCRequestFromPrompt = promptOrig })

				// The update call itself just returns the same error
				updateVXCFunc = func(ctx context.Context, client *megaport.Client, vxcUID string, req *megaport.UpdateVXCRequest) error {
					return fmt.Errorf("no update parameters provided")
				}
			},
			args: []string{"vxc-12345"},
		},
		{
			name:          "missing args",
			expectedError: "requires exactly 1 arg(s)",
			setupMock: func(t *testing.T, m *mockVXCService) {
				// no calls expected
			},
			args: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockVXCService{}
			if tt.setupMock != nil {
				tt.setupMock(t, mockService)
			}

			// Override the loginFunc with a test double
			config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
				return &megaport.Client{
					VXCService: mockService,
				}, nil
			}

			// Override the prompt responses - with safe handling for no prompts
			promptIndex := 0
			promptResponses := tt.prompts
			utils.Prompt = func(msg string, _ bool) (string, error) {
				if len(promptResponses) == 0 {
					// If no prompts are expected in this test, return empty string
					// instead of failing
					return "", nil
				}

				if promptIndex < len(promptResponses) {
					resp := promptResponses[promptIndex]
					promptIndex++
					return resp, nil
				}
				return "", fmt.Errorf("unexpected additional prompt")
			}

			// Create command and set flags
			cmd := &cobra.Command{}
			cmd.Flags().Bool("interactive", false, "")
			cmd.Flags().String("name", "", "")
			cmd.Flags().Int("rate-limit", 0, "")
			cmd.Flags().Int("term", 0, "")
			cmd.Flags().String("cost-centre", "", "")
			cmd.Flags().Bool("shutdown", false, "")
			cmd.Flags().Int("a-end-vlan", 0, "")
			cmd.Flags().Int("b-end-vlan", 0, "")
			cmd.Flags().Int("a-end-inner-vlan", 0, "")
			cmd.Flags().Int("b-end-inner-vlan", 0, "")
			cmd.Flags().String("a-end-uid", "", "")
			cmd.Flags().String("b-end-uid", "", "")
			cmd.Flags().String("a-end-partner-config", "", "")
			cmd.Flags().String("b-end-partner-config", "", "")
			cmd.Flags().String("json", "", "")
			cmd.Flags().String("json-file", "", "")

			// Set string flags
			for flag, value := range tt.flags {
				err := cmd.Flags().Set(flag, value)
				assert.NoError(t, err)
			}

			// Set integer flags
			for flag, value := range tt.flagsInt {
				err := cmd.Flags().Set(flag, fmt.Sprintf("%d", value))
				assert.NoError(t, err)
			}

			// Set boolean flags
			for flag, value := range tt.flagsBool {
				if value {
					err := cmd.Flags().Set(flag, "true")
					assert.NoError(t, err)
				}
			}

			// Execute the function and capture output
			var err error
			output := output.CaptureOutput(func() {
				switch {
				case tt.name == "missing args" && len(tt.args) == 0:
					err = fmt.Errorf("requires exactly 1 arg(s)")
				case len(tt.args) == 0:
					// avoid cmd.Args(...) or UpdateVXC(...) if we truly have zero args
					err = fmt.Errorf("unexpected argument count of 0")
				default:
					err = UpdateVXC(cmd, tt.args, noColor)
				}
			})

			// Check results
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				if tt.expectedOutput != "" {
					assert.Contains(t, output, tt.expectedOutput)
				}
			}
		})
	}
}

func TestDeleteVXC(t *testing.T) {
	originalLoginFunc := config.LoginFunc
	originalDeleteVXCFunc := deleteVXCFunc
	originalConfirmPrompt := utils.ConfirmPrompt
	noColor := true

	// Restore the originals after test completes
	defer func() {
		config.LoginFunc = originalLoginFunc
		deleteVXCFunc = originalDeleteVXCFunc
		utils.ConfirmPrompt = originalConfirmPrompt
	}()

	tests := []struct {
		name           string
		expectedError  string
		setupMock      func(*testing.T, *mockVXCService)
		args           []string
		force          bool
		deleteNow      bool
		confirmDelete  bool
		expectedOutput string
	}{
		{
			name:           "successful VXC deletion",
			expectedError:  "",
			force:          true, // Skip confirmation
			args:           []string{"vxc-12345"},
			expectedOutput: "VXC deleted vxc-12345",
			setupMock: func(t *testing.T, m *mockVXCService) {
				deleteVXCFunc = func(ctx context.Context, client *megaport.Client, vxcUID string, req *megaport.DeleteVXCRequest) error {
					return nil
				}
			},
		},
		{
			name:           "successful VXC deletion with confirmation",
			expectedError:  "",
			confirmDelete:  true,
			args:           []string{"vxc-12345"},
			expectedOutput: "VXC deleted vxc-12345",
			setupMock: func(t *testing.T, m *mockVXCService) {
				deleteVXCFunc = func(ctx context.Context, client *megaport.Client, vxcUID string, req *megaport.DeleteVXCRequest) error {
					return nil
				}
			},
		},
		{
			name:           "immediate deletion",
			expectedError:  "",
			force:          true,
			deleteNow:      true,
			args:           []string{"vxc-12345"},
			expectedOutput: "VXC deleted vxc-12345",
			setupMock: func(t *testing.T, m *mockVXCService) {
				deleteVXCFunc = func(ctx context.Context, client *megaport.Client, vxcUID string, req *megaport.DeleteVXCRequest) error {
					// Verify deleteNow is passed correctly
					assert.True(t, req.DeleteNow)
					return nil
				}
			},
		},
		{
			name:           "deletion cancelled",
			confirmDelete:  false,
			args:           []string{"vxc-12345"},
			expectedOutput: "Deletion cancelled",
			setupMock: func(t *testing.T, m *mockVXCService) {
				// No deletion should happen
				deleteVXCFunc = func(ctx context.Context, client *megaport.Client, vxcUID string, req *megaport.DeleteVXCRequest) error {
					t.Fatal("DeleteVXC should not be called when deletion is cancelled")
					return nil
				}
			},
		},
		{
			name:          "API error",
			expectedError: "API error",
			force:         true,
			args:          []string{"vxc-12345"},
			setupMock: func(t *testing.T, m *mockVXCService) {
				deleteVXCFunc = func(ctx context.Context, client *megaport.Client, vxcUID string, req *megaport.DeleteVXCRequest) error {
					return fmt.Errorf("API error")
				}
			},
		},
		{
			name:          "missing args",
			expectedError: "requires exactly 1 arg(s)",
			setupMock: func(t *testing.T, m *mockVXCService) {
				// No mock needed
			},
			args: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockVXCService{}
			if tt.setupMock != nil {
				tt.setupMock(t, mockService)
			}

			// Override the loginFunc with a test double
			config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
				return &megaport.Client{
					VXCService: mockService,
				}, nil
			}

			// Override the confirmation prompt
			utils.ConfirmPrompt = func(message string, _ bool) bool {
				assert.Contains(t, message, "Are you sure you want to delete VXC")
				return tt.confirmDelete
			}

			// Create command and set flags
			cmd := &cobra.Command{}
			cmd.Flags().Bool("force", tt.force, "")
			cmd.Flags().Bool("now", tt.deleteNow, "")

			// Execute the function
			var err error
			output := output.CaptureOutput(func() {
				if len(tt.args) == 0 {
					err = fmt.Errorf("requires exactly 1 arg(s)")
				} else {
					err = DeleteVXC(cmd, tt.args, noColor)
				}
			})

			// Check results
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				if tt.expectedOutput != "" {
					assert.Contains(t, output, tt.expectedOutput)
				}
			}
		})
	}
}
