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

var originalBuyVXCFunc = buyVXCFunc
var interactive bool

func TestBuyVXC(t *testing.T) {
	originalResourcePrompt := utils.ResourcePrompt
	originalLoginFunc := config.LoginFunc
	originalInteractiveFlag := interactive
	noColor := true

	defer func() {
		utils.ResourcePrompt = originalResourcePrompt
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
		skipRequest    bool
		expectedOutput string
	}{
		{
			name: "successful VXC purchase - interactive mode",
			prompts: []string{
				"a-end-uid",
				"b-end-uid",
				"Test VXC",
				"100",
				"12",
				"100",
				"200",
				"0",
				"300",
				"400",
				"1",
				"",
				"",
				"",
				"no",
				"no",
			},
			expectedError: "",
			setupMock: func(t *testing.T, m *mockVXCService) {
				m.buyVXCResponse = &megaport.BuyVXCResponse{
					TechnicalServiceUID: "vxc-sample-uid",
				}

				interactive = true

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
				m.buyVXCResponse = &megaport.BuyVXCResponse{
					TechnicalServiceUID: "",
				}
				m.buyVXCErr = fmt.Errorf("API error")

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

				buyVXCFunc = func(ctx context.Context, client *megaport.Client, req *megaport.BuyVXCRequest) (*megaport.BuyVXCResponse, error) {
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

			config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
				return &megaport.Client{
					VXCService: mockService,
				}, nil
			}

			promptIndex := 0
			promptResponses := tt.prompts

			utils.ResourcePrompt = func(_, _ string, _ bool) (string, error) {
				if promptIndex < len(promptResponses) {
					resp := promptResponses[promptIndex]
					promptIndex++
					return resp, nil
				}
				return "", fmt.Errorf("unexpected additional prompt")
			}

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

			for flag, value := range tt.flags {
				err := cmd.Flags().Set(flag, value)
				assert.NoError(t, err)
			}

			for flag, value := range tt.flagsInt {
				err := cmd.Flags().Set(flag, fmt.Sprintf("%d", value))
				assert.NoError(t, err)
			}

			for flag, value := range tt.flagsBool {
				if value {
					err := cmd.Flags().Set(flag, "true")
					assert.NoError(t, err)
				}
			}

			err := BuyVXC(cmd, tt.args, noColor)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateVXCResourceTagsCmd(t *testing.T) {
	originalLoginFunc := config.LoginFunc
	originalResourcePrompt := utils.UpdateResourceTagsPrompt
	defer func() {
		config.LoginFunc = originalLoginFunc
		utils.UpdateResourceTagsPrompt = originalResourcePrompt
	}()

	tests := []struct {
		name                 string
		vxcUID               string
		interactive          bool
		promptResult         map[string]string
		promptError          error
		jsonInput            string
		jsonFile             string
		existingTags         map[string]string
		setupMock            func(*mockVXCService)
		expectedError        string
		expectedOutput       string
		expectedCapturedTags map[string]string
	}{
		{
			name:        "successful update with interactive mode",
			vxcUID:      "vxc-123",
			interactive: true,
			existingTags: map[string]string{
				"environment": "staging",
			},
			promptResult: map[string]string{
				"environment": "production",
				"team":        "networking",
			},
			setupMock: func(m *mockVXCService) {
				m.ListVXCResourceTagsResult = map[string]string{
					"environment": "staging",
				}
			},
			expectedOutput: "Resource tags updated for VXC vxc-123",
			expectedCapturedTags: map[string]string{
				"environment": "production",
				"team":        "networking",
			},
		},
		{
			name:   "successful update with json",
			vxcUID: "vxc-456",
			jsonInput: `{
				"environment": "production", 
				"team": "networking",
				"project": "cloud-migration"
			}`,
			existingTags: map[string]string{
				"environment": "development",
				"owner":       "john.doe@example.com",
			},
			setupMock: func(m *mockVXCService) {
				m.ListVXCResourceTagsResult = map[string]string{
					"environment": "development",
					"owner":       "john.doe@example.com",
				}
			},
			expectedOutput: "Resource tags updated for VXC vxc-456",
			expectedCapturedTags: map[string]string{
				"environment": "production",
				"team":        "networking",
				"project":     "cloud-migration",
			},
		},
		{
			name:      "error with invalid json",
			vxcUID:    "vxc-789",
			jsonInput: `{invalid json}`,
			setupMock: func(m *mockVXCService) {
				m.ListVXCResourceTagsResult = map[string]string{}
			},
			expectedError: "error parsing JSON",
		},
		{
			name:        "error in interactive mode",
			vxcUID:      "vxc-prompt-error",
			interactive: true,
			existingTags: map[string]string{
				"environment": "staging",
			},
			promptError: fmt.Errorf("user cancelled input"),
			setupMock: func(m *mockVXCService) {
				m.ListVXCResourceTagsResult = map[string]string{
					"environment": "staging",
				}
			},
			expectedError: "user cancelled input",
		},
		{
			name:   "error with API update",
			vxcUID: "vxc-update-error",
			jsonInput: `{
				"environment": "production"
			}`,
			setupMock: func(m *mockVXCService) {
				m.ListVXCResourceTagsResult = map[string]string{}
				m.UpdateVXCResourceTagsErr = fmt.Errorf("API error: unauthorized")
			},
			expectedError: "failed to update resource tags",
		},
		{
			name:   "error with API tag listing",
			vxcUID: "vxc-list-error",
			jsonInput: `{
				"environment": "production"
			}`,
			setupMock: func(m *mockVXCService) {
				m.ListVXCResourceTagsErr = fmt.Errorf("API error: resource not found")
			},
			expectedError: "failed to get existing resource tags",
		},
		{
			name:      "empty tags clear all existing tags",
			vxcUID:    "vxc-clear-tags",
			jsonInput: `{}`,
			existingTags: map[string]string{
				"environment": "staging",
				"team":        "networking",
			},
			setupMock: func(m *mockVXCService) {
				m.ListVXCResourceTagsResult = map[string]string{
					"environment": "staging",
					"team":        "networking",
				}
			},
			expectedOutput:       "Resource tags updated for VXC vxc-clear-tags",
			expectedCapturedTags: map[string]string{},
		},
		{
			name:   "no input provided",
			vxcUID: "vxc-no-input",
			setupMock: func(m *mockVXCService) {
				m.ListVXCResourceTagsResult = map[string]string{}
			},
			expectedError: "no input provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockVXCService{}
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
				client := &megaport.Client{}
				client.VXCService = mockService
				return client, nil
			}

			utils.UpdateResourceTagsPrompt = func(existingTags map[string]string, noColor bool) (map[string]string, error) {
				return tt.promptResult, tt.promptError
			}

			cmd := &cobra.Command{
				Use: "update-tags [vxcUID]",
				RunE: func(cmd *cobra.Command, args []string) error {
					return UpdateVXCResourceTags(cmd, args, false)
				},
			}

			cmd.Flags().Bool("interactive", false, "")
			cmd.Flags().String("json", "", "")
			cmd.Flags().String("json-file", "", "")

			if tt.interactive {
				err := cmd.Flags().Set("interactive", "true")
				assert.NoError(t, err)
			}

			if tt.jsonInput != "" {
				err := cmd.Flags().Set("json", tt.jsonInput)
				assert.NoError(t, err)
			}

			if tt.jsonFile != "" {
				err := cmd.Flags().Set("json-file", tt.jsonFile)
				assert.NoError(t, err)
			}

			var err error
			output := output.CaptureOutput(func() {
				err = cmd.RunE(cmd, []string{tt.vxcUID})
			})

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, output, tt.expectedOutput)

				if tt.expectedCapturedTags != nil {
					assert.Equal(t, tt.expectedCapturedTags, mockService.CapturedUpdateVXCResourceTagsRequest)
				}
			}
		})
	}
}

func TestGetVXCStatus(t *testing.T) {
	originalLoginFunc := config.LoginFunc
	defer func() {
		config.LoginFunc = originalLoginFunc
	}()

	tests := []struct {
		name           string
		vxcUID         string
		setupMock      func(*mockVXCService)
		expectedError  string
		expectedOutput string
		outputFormat   string
	}{
		{
			name:   "successful status retrieval - table format",
			vxcUID: "vxc-123abc",
			setupMock: func(m *mockVXCService) {
				m.getVXCResponse = &megaport.VXC{
					UID:                "vxc-123abc",
					Name:               "Test VXC",
					ProvisioningStatus: "CONFIGURED",
					AEndConfiguration: megaport.VXCEndConfiguration{
						UID: "port-aend",
					},
					BEndConfiguration: megaport.VXCEndConfiguration{
						UID: "port-bend",
					},
					RateLimit: 1000,
				}
			},
			expectedOutput: "vxc-123abc",
			outputFormat:   "table",
		},
		{
			name:   "successful status retrieval - json format",
			vxcUID: "vxc-123abc",
			setupMock: func(m *mockVXCService) {
				m.getVXCResponse = &megaport.VXC{
					UID:                "vxc-123abc",
					Name:               "Test VXC",
					ProvisioningStatus: "LIVE",
					AEndConfiguration: megaport.VXCEndConfiguration{
						UID: "port-aend",
					},
					BEndConfiguration: megaport.VXCEndConfiguration{
						UID: "port-bend",
					},
					RateLimit: 1000,
				}
			},
			expectedOutput: "vxc-123abc",
			outputFormat:   "json",
		},
		{
			name:   "VXC not found",
			vxcUID: "vxc-notfound",
			setupMock: func(m *mockVXCService) {
				m.getVXCError = fmt.Errorf("VXC not found")
			},
			expectedError: "error getting VXC status",
			outputFormat:  "table",
		},
		{
			name:   "API error",
			vxcUID: "vxc-error",
			setupMock: func(m *mockVXCService) {
				m.getVXCError = fmt.Errorf("API error")
			},
			expectedError: "API error",
			outputFormat:  "table",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockVXCService{}
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
				client := &megaport.Client{}
				client.VXCService = mockService
				return client, nil
			}

			cmd := &cobra.Command{
				Use: "status [vxcUID]",
			}

			var err error
			capturedOutput := output.CaptureOutput(func() {
				err = GetVXCStatus(cmd, []string{tt.vxcUID}, true, tt.outputFormat)
			})

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, capturedOutput, tt.expectedOutput)

				switch tt.outputFormat {
				case "json":
					assert.Contains(t, capturedOutput, "\"uid\":")
					assert.Contains(t, capturedOutput, "\"name\":")
					assert.Contains(t, capturedOutput, "\"status\":")
					assert.Contains(t, capturedOutput, "\"type\":")
				case "table":
					assert.Contains(t, capturedOutput, "UID")
					assert.Contains(t, capturedOutput, "NAME")
					assert.Contains(t, capturedOutput, "STATUS")
					assert.Contains(t, capturedOutput, "TYPE")
				}
			}
		})
	}
}
