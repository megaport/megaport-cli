package mcr

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/megaport/megaport-cli/internal/base/config"
	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/utils"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// Function to adapt our old tests to work with new wrapCommandFunc signature
func testCommandAdapter(fn func(cmd *cobra.Command, args []string, noColor bool) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return fn(cmd, args, false) // Pass false for noColor in tests
	}
}

func TestGetMCRCmd_WithMockClient(t *testing.T) {
	originalLoginFunc := config.LoginFunc
	defer func() {
		config.LoginFunc = originalLoginFunc
	}()

	tests := []struct {
		name          string
		mcrID         string
		format        string
		setupMock     func(*MockMCRService)
		expectedError string
		expectedOut   []string
	}{
		{
			name:   "get MCR success table format",
			mcrID:  "mcr-123",
			format: "table",
			setupMock: func(m *MockMCRService) {
				m.GetMCRResult = &megaport.MCR{
					UID:                "mcr-123",
					Name:               "Test MCR",
					LocationID:         123,
					ProvisioningStatus: "LIVE",
					Resources: megaport.MCRResources{
						VirtualRouter: megaport.MCRVirtualRouter{
							ASN: 65000,
						},
					},
				}
			},
			expectedOut: []string{"mcr-123", "Test MCR", "LIVE"},
		},
		{
			name:   "get MCR success json format",
			mcrID:  "mcr-123",
			format: "json",
			setupMock: func(m *MockMCRService) {
				m.GetMCRResult = &megaport.MCR{
					UID:                "mcr-123",
					Name:               "Test MCR",
					LocationID:         123,
					ProvisioningStatus: "LIVE",
					Resources: megaport.MCRResources{
						VirtualRouter: megaport.MCRVirtualRouter{
							ASN: 65000,
						},
					},
				}
			},
			expectedOut: []string{`"uid": "mcr-123"`, `"name": "Test MCR"`, `"location_id": 123`},
		},
		{
			name:   "get MCR error",
			mcrID:  "mcr-invalid",
			format: "table",
			setupMock: func(m *MockMCRService) {
				m.GetMCRErr = fmt.Errorf("MCR not found")
			},
			expectedError: "MCR not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMCRService := &MockMCRService{}
			tt.setupMock(mockMCRService)

			config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
				client := &megaport.Client{}
				client.MCRService = mockMCRService
				return client, nil
			}

			var err error
			cmd := &cobra.Command{
				Use: "get-mcr [mcrID]",
				RunE: func(cmd *cobra.Command, args []string) error {
					return GetMCR(cmd, args, false, tt.format)
				},
			}

			// Now add the output flag to the new command
			cmd.Flags().StringP("output", "o", "table", "Output format (json, table)")
			err = cmd.Flags().Set("output", tt.format)
			if err != nil {
				t.Fatalf("Failed to set output format: %v", err)
			}

			output := output.CaptureOutput(func() {
				err = cmd.RunE(cmd, []string{tt.mcrID})
			})

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				for _, expected := range tt.expectedOut {
					assert.Contains(t, output, expected)
				}
			}
		})
	}
}

func TestDeleteMCRCmd_WithMockClient(t *testing.T) {
	// Save original functions and restore after test
	originalLoginFunc := config.LoginFunc
	originalPrompt := utils.Prompt
	defer func() {
		config.LoginFunc = originalLoginFunc
		utils.Prompt = originalPrompt
	}()

	tests := []struct {
		name           string
		mcrID          string
		force          bool
		deleteNow      bool
		promptResponse string
		setupMock      func(*MockMCRService)
		expectedError  string
		expectedOutput string
		expectDeleted  bool
	}{
		{
			name:           "confirm deletion",
			mcrID:          "mcr-to-delete",
			force:          false,
			deleteNow:      false,
			promptResponse: "y",
			setupMock: func(m *MockMCRService) {
				m.DeleteMCRResult = &megaport.DeleteMCRResponse{
					IsDeleting: true,
				}
			},
			expectedOutput: "MCR deleted", // Changed from "MCR mcr-to-delete deleted successfully"
			expectDeleted:  true,
		},
		{
			name:           "confirm immediate deletion",
			mcrID:          "mcr-to-delete-now",
			force:          false,
			deleteNow:      true,
			promptResponse: "y",
			setupMock: func(m *MockMCRService) {
				m.DeleteMCRResult = &megaport.DeleteMCRResponse{
					IsDeleting: true,
				}
			},
			expectedOutput: "MCR deleted", // Changed from "MCR mcr-to-delete-now deleted successfully"
			expectDeleted:  true,
		},
		{
			name:      "force deletion",
			mcrID:     "mcr-force-delete",
			force:     true,
			deleteNow: false,
			setupMock: func(m *MockMCRService) {
				m.DeleteMCRResult = &megaport.DeleteMCRResponse{
					IsDeleting: true,
				}
			},
			expectedOutput: "MCR deleted", // Changed from "MCR mcr-force-delete deleted successfully"
			expectDeleted:  true,
		},
		{
			name:           "cancel deletion",
			mcrID:          "mcr-keep",
			force:          false,
			promptResponse: "n",
			setupMock:      func(m *MockMCRService) {},
			expectedOutput: "Deletion cancelled",
			expectDeleted:  false,
		},
		{
			name:           "deletion error",
			mcrID:          "mcr-error",
			force:          false,
			promptResponse: "y",
			setupMock: func(m *MockMCRService) {
				m.DeleteMCRErr = fmt.Errorf("error deleting MCR")
			},
			expectedError: "error deleting MCR",
			expectDeleted: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock MCR service
			mockMCRService := &MockMCRService{}
			tt.setupMock(mockMCRService)

			// Setup login to return our mock client
			config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
				client := &megaport.Client{}
				client.MCRService = mockMCRService
				return client, nil
			}

			// Setup prompt mock
			utils.Prompt = func(msg string, _ bool) (string, error) {
				return tt.promptResponse, nil
			}

			// Set flags
			cmd := &cobra.Command{
				Use: "delete-mcr [mcrID]",
				RunE: func(cmd *cobra.Command, args []string) error {
					return DeleteMCR(cmd, args, false)
				},
			}
			cmd.Flags().BoolP("force", "f", false, "Force deletion without confirmation")
			cmd.Flags().Bool("now", false, "Delete MCR immediately instead of at end of billing cycle")
			err := cmd.Flags().Set("force", fmt.Sprintf("%v", tt.force))
			if err != nil {
				t.Fatalf("Failed to set force flag: %v", err)
			}
			err = cmd.Flags().Set("now", fmt.Sprintf("%v", tt.deleteNow))
			if err != nil {
				t.Fatalf("Failed to set now flag: %v", err)
			}

			// Execute command and capture output
			output := output.CaptureOutput(func() {
				err = cmd.RunE(cmd, []string{tt.mcrID})
			})

			// Check results
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, output, tt.expectedOutput)

				// Verify the request if deletion was expected
				if tt.expectDeleted {
					assert.NotNil(t, mockMCRService.CapturedDeleteMCRUID)
					assert.Equal(t, tt.mcrID, mockMCRService.CapturedDeleteMCRUID)
				}
			}
		})
	}
}

func TestRestoreMCRCmd_WithMockClient(t *testing.T) {
	// Save original login function and restore after test
	originalLoginFunc := config.LoginFunc
	defer func() {
		config.LoginFunc = originalLoginFunc
	}()

	tests := []struct {
		name          string
		mcrID         string
		setupMock     func(*MockMCRService)
		expectedError string
		expectedOut   string
	}{
		{
			name:  "restore MCR success",
			mcrID: "mcr-to-restore",
			setupMock: func(m *MockMCRService) {
				m.RestoreMCRResult = &megaport.RestoreMCRResponse{
					IsRestored: true,
				}
			},
			expectedOut: "MCR mcr-to-restore restored successfully",
		},
		{
			name:  "restore MCR error",
			mcrID: "mcr-error",
			setupMock: func(m *MockMCRService) {
				m.RestoreMCRErr = fmt.Errorf("error restoring MCR")
			},
			expectedError: "error restoring MCR",
		},
		{
			name:  "restore MCR unsuccessful",
			mcrID: "mcr-fail",
			setupMock: func(m *MockMCRService) {
				m.RestoreMCRResult = &megaport.RestoreMCRResponse{
					IsRestored: false,
				}
			},
			expectedOut: "MCR restoration request was not successful",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock MCR service
			mockMCRService := &MockMCRService{}
			tt.setupMock(mockMCRService)

			// Setup login to return our mock client
			config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
				client := &megaport.Client{}
				client.MCRService = mockMCRService
				return client, nil
			}

			// Setup command
			restoreMCRCmd := &cobra.Command{
				Use: "restore-mcr [mcrID]",
				RunE: func(cmd *cobra.Command, args []string) error {
					return RestoreMCR(cmd, args, false)
				},
			}

			// Execute command and capture output
			var err error
			output := output.CaptureOutput(func() {
				err = restoreMCRCmd.RunE(restoreMCRCmd, []string{tt.mcrID})
			})

			// Check results
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, output, tt.expectedOut)
			}
		})
	}
}

func TestGetMCRFunc(t *testing.T) {
	mockMCRService := &MockMCRService{
		GetMCRResult: &megaport.MCR{
			UID:                "mcr-123",
			Name:               "Test MCR",
			LocationID:         123,
			ProvisioningStatus: "LIVE",
		},
	}

	client := &megaport.Client{
		MCRService: mockMCRService,
	}

	ctx := context.Background()
	mcr, err := getMCRFunc(ctx, client, "mcr-123")
	assert.NoError(t, err)
	assert.NotNil(t, mcr)
	assert.Equal(t, "mcr-123", mcr.UID)
	assert.Equal(t, "Test MCR", mcr.Name)
}

func TestBuyMCRFunc(t *testing.T) {
	mockMCRService := &MockMCRService{
		BuyMCRResult: &megaport.BuyMCRResponse{
			TechnicalServiceUID: "mcr-123-abc",
		},
	}

	client := &megaport.Client{
		MCRService: mockMCRService,
	}

	ctx := context.Background()
	req := &megaport.BuyMCRRequest{
		Name:             "Test MCR",
		Term:             12,
		PortSpeed:        1000,
		LocationID:       123,
		DiversityZone:    "red",
		CostCentre:       "cost-123",
		PromoCode:        "PROMO2025",
		WaitForProvision: true,
		WaitForTime:      10 * time.Minute,
	}
	resp, err := buyMCRFunc(ctx, client, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "mcr-123-abc", resp.TechnicalServiceUID)
}

func TestDeleteMCRFunc(t *testing.T) {
	mockMCRService := &MockMCRService{
		DeleteMCRResult: &megaport.DeleteMCRResponse{
			IsDeleting: true,
		},
	}

	client := &megaport.Client{
		MCRService: mockMCRService,
	}

	ctx := context.Background()
	req := &megaport.DeleteMCRRequest{
		MCRID:     "mcr-123",
		DeleteNow: true,
	}
	resp, err := deleteMCRFunc(ctx, client, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.IsDeleting)
}

func TestRestoreMCRFunc(t *testing.T) {
	mockMCRService := &MockMCRService{
		RestoreMCRResult: &megaport.RestoreMCRResponse{
			IsRestored: true,
		},
	}

	client := &megaport.Client{
		MCRService: mockMCRService,
	}

	ctx := context.Background()
	resp, err := restoreMCRFunc(ctx, client, "mcr-123")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.IsRestored)
}

func TestListMCRPrefixFilterListsCmd_WithMockClient(t *testing.T) {
	originalLoginFunc := config.LoginFunc
	defer func() {
		config.LoginFunc = originalLoginFunc
	}()

	tests := []struct {
		name           string
		mcrUID         string
		setupMock      func(*MockMCRService)
		expectedError  string
		expectedOutput string
	}{
		{
			name:   "successful list prefix filter lists",
			mcrUID: "mcr-123",
			setupMock: func(m *MockMCRService) {
				m.ListMCRPrefixFilterListsResult = []*megaport.PrefixFilterList{
					{
						Id:          1,
						Description: "Test Prefix Filter List 1",
					},
					{
						Id:          2,
						Description: "Test Prefix Filter List 2",
					},
				}
			},
			expectedOutput: "Test Prefix Filter List 1",
		},
		{
			name:   "API error",
			mcrUID: "mcr-123",
			setupMock: func(m *MockMCRService) {
				m.ListMCRPrefixFilterListsErr = fmt.Errorf("API error: service unavailable")
			},
			expectedError: "API error: service unavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMCRService := &MockMCRService{}
			tt.setupMock(mockMCRService)

			config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
				client := &megaport.Client{}
				client.MCRService = mockMCRService
				return client, nil
			}

			// Setup command
			listMCRPrefixFilterListsCmd := &cobra.Command{
				Use: "list-mcr-prefix-filter-lists [mcrUID]",
				RunE: func(cmd *cobra.Command, args []string) error {
					return ListMCRPrefixFilterLists(cmd, args, false, "table")
				},
			}

			cmd := listMCRPrefixFilterListsCmd
			var err error
			output := output.CaptureOutput(func() {
				err = cmd.RunE(cmd, []string{tt.mcrUID})
			})

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, output, tt.expectedOutput)
			}
		})
	}
}

func TestGetMCRPrefixFilterListCmd_WithMockClient(t *testing.T) {
	originalLoginFunc := config.LoginFunc
	defer func() {
		config.LoginFunc = originalLoginFunc
	}()

	tests := []struct {
		name           string
		mcrUID         string
		prefixListID   int
		setupMock      func(*MockMCRService)
		expectedError  string
		expectedOutput string
	}{
		{
			name:         "successful get prefix filter list",
			mcrUID:       "mcr-123",
			prefixListID: 1,
			setupMock: func(m *MockMCRService) {
				m.GetMCRPrefixFilterListResult = &megaport.MCRPrefixFilterList{
					ID:          1,
					Description: "Test Prefix Filter List",
				}
			},
			expectedOutput: "Test Prefix Filter List",
		},
		{
			name:         "API error",
			mcrUID:       "mcr-123",
			prefixListID: 1,
			setupMock: func(m *MockMCRService) {
				m.GetMCRPrefixFilterListErr = fmt.Errorf("API error: service unavailable")
			},
			expectedError: "API error: service unavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMCRService := &MockMCRService{}
			tt.setupMock(mockMCRService)

			config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
				client := &megaport.Client{}
				client.MCRService = mockMCRService
				return client, nil
			}

			cmd := &cobra.Command{
				Use:  "get-mcr-prefix-filter-list [mcrUID] [prefixListID]",
				Args: cobra.ExactArgs(2),
				RunE: func(cmd *cobra.Command, args []string) error {
					return GetMCRPrefixFilterList(cmd, args, false, "table")
				},
			}
			var err error
			output := output.CaptureOutput(func() {
				err = cmd.RunE(cmd, []string{tt.mcrUID, fmt.Sprintf("%d", tt.prefixListID)})
			})

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, output, tt.expectedOutput)
			}
		})
	}
}

func TestDeleteMCRPrefixFilterListCmd_WithMockClient(t *testing.T) {
	originalLoginFunc := config.LoginFunc
	originalPrompt := utils.Prompt
	defer func() {
		config.LoginFunc = originalLoginFunc
		utils.Prompt = originalPrompt
	}()

	tests := []struct {
		name           string
		mcrUID         string
		prefixListID   int
		force          bool
		promptResponse string
		setupMock      func(*MockMCRService)
		expectedError  string
		expectedOutput string
	}{
		{
			name:           "successful delete prefix filter list",
			mcrUID:         "mcr-123",
			prefixListID:   1,
			force:          false,
			promptResponse: "y",
			setupMock: func(m *MockMCRService) {
				m.DeleteMCRPrefixFilterListResult = &megaport.DeleteMCRPrefixFilterListResponse{
					IsDeleted: true,
				}
			},
			expectedOutput: "Prefix filter list deleted successfully - ID: 1",
		},
		{
			name:           "API error",
			mcrUID:         "mcr-123",
			prefixListID:   1,
			force:          false,
			promptResponse: "y",
			setupMock: func(m *MockMCRService) {
				m.DeleteMCRPrefixFilterListErr = fmt.Errorf("API error: service unavailable")
			},
			expectedError: "API error: service unavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMCRService := &MockMCRService{}
			tt.setupMock(mockMCRService)

			config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
				client := &megaport.Client{}
				client.MCRService = mockMCRService
				return client, nil
			}

			// Mock the prompt function
			utils.Prompt = func(msg string, _ bool) (string, error) {
				return tt.promptResponse, nil
			}

			cmd := &cobra.Command{
				Use: "delete-mcr-prefix-filter-list [mcrUID] [prefixListID]",
				RunE: func(cmd *cobra.Command, args []string) error {
					return DeleteMCRPrefixFilterList(cmd, args, false)
				},
			}
			cmd.Flags().BoolP("force", "f", false, "Force deletion without confirmation")
			err := cmd.Flags().Set("force", fmt.Sprintf("%v", tt.force))
			if err != nil {
				t.Fatalf("Failed to set force flag: %v", err)
			}
			cmd.Flags().Bool("now", false, "Delete prefix filter list immediately instead of at end of billing cycle")
			err = cmd.Flags().Set("now", fmt.Sprintf("%v", false))
			if err != nil {
				t.Fatalf("Failed to set now flag: %v", err)
			}

			output := output.CaptureOutput(func() {
				err = cmd.RunE(cmd, []string{tt.mcrUID, fmt.Sprintf("%d", tt.prefixListID)})
			})

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, output, tt.expectedOutput)
			}
		})
	}
}

func TestBuyMCRCmd_WithMockClient(t *testing.T) {
	// Save original functions and restore after test
	originalPrompt := utils.Prompt
	originalLoginFunc := config.LoginFunc
	originalBuyMCRFunc := buyMCRFunc
	defer func() {
		utils.Prompt = originalPrompt
		config.LoginFunc = originalLoginFunc
		buyMCRFunc = originalBuyMCRFunc
	}()

	tests := []struct {
		name           string
		args           []string
		interactive    bool
		prompts        []string
		flags          map[string]string
		jsonInput      string
		jsonFilePath   string
		setupMock      func(*MockMCRService)
		expectedError  string
		expectedOutput string
	}{
		{
			name:        "interactive mode success",
			interactive: true,
			prompts: []string{
				"Test MCR",     // name
				"12",           // term
				"10000",        // port speed
				"123",          // location ID
				"65000",        // ASN
				"red",          // diversity zone
				"cost-123",     // cost centre
				"MCRPROMO2025", // promo code
			},
			setupMock: func(m *MockMCRService) {
				m.BuyMCRResult = &megaport.BuyMCRResponse{
					TechnicalServiceUID: "mcr-123-abc",
				}
			},
			expectedOutput: "MCR created",
		},
		{
			name: "flag mode success",
			flags: map[string]string{
				"name":           "Flag MCR",
				"term":           "12",
				"port-speed":     "10000",
				"location-id":    "123",
				"mcr-asn":        "65000",
				"diversity-zone": "blue",
				"cost-centre":    "cost-456",
				"promo-code":     "FLAGPROMO",
			},
			setupMock: func(m *MockMCRService) {
				m.BuyMCRResult = &megaport.BuyMCRResponse{
					TechnicalServiceUID: "mcr-456-def",
				}
			},
			expectedOutput: "MCR created",
		},
		{
			name: "JSON string mode success",
			flags: map[string]string{
				"json": `{"name":"JSON MCR","term":24,"portSpeed":10000,"locationId":123,"mcrAsn":65000,"diversityZone":"green","costCentre":"cost-789","promoCode":"JSONPROMO"}`,
			},
			setupMock: func(m *MockMCRService) {
				m.BuyMCRResult = &megaport.BuyMCRResponse{
					TechnicalServiceUID: "mcr-789-ghi",
				}
			},
			expectedOutput: "MCR created",
		},
		{
			name: "missing required fields in flag mode",
			flags: map[string]string{
				"name": "Test MCR",
			},
			expectedError: "term is required",
		},
		{
			name: "invalid term in flag mode",
			flags: map[string]string{
				"name":        "Test MCR",
				"term":        "13", // Invalid term
				"port-speed":  "10000",
				"location-id": "123",
				"mcr-asn":     "65000",
			},
			expectedError: "invalid term",
		},
		{
			name: "invalid JSON",
			flags: map[string]string{
				"json": `{"name":"Test MCR","term":"invalid"}`,
			},
			expectedError: "error parsing JSON",
		},
		{
			name: "API error",
			flags: map[string]string{
				"name":        "Test MCR",
				"term":        "12",
				"port-speed":  "10000",
				"location-id": "123",
				"mcr-asn":     "65000",
			},
			setupMock: func(m *MockMCRService) {
				m.BuyMCRErr = fmt.Errorf("API error: service unavailable")
			},
			expectedError: "API error: service unavailable",
		},
		{
			name:          "no input provided",
			expectedError: "no input provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock prompt
			if len(tt.prompts) > 0 {
				promptIndex := 0
				utils.Prompt = func(msg string, _ bool) (string, error) {
					if promptIndex < len(tt.prompts) {
						response := tt.prompts[promptIndex]
						promptIndex++
						return response, nil
					}
					return "", fmt.Errorf("unexpected prompt call")
				}
			}

			// Setup mock MCR service
			mockMCRService := &MockMCRService{}
			if tt.setupMock != nil {
				tt.setupMock(mockMCRService)
			}

			// Setup login to return our mock client
			config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
				client := &megaport.Client{}
				client.MCRService = mockMCRService
				return client, nil
			}

			// Create a fresh command for each test to avoid flag conflicts
			cmd := &cobra.Command{
				Use:  "buy",
				RunE: testCommandAdapter(BuyMCR),
			}

			// Add all the necessary flags
			cmd.Flags().BoolP("interactive", "i", false, "Use interactive mode with prompts")
			cmd.Flags().String("name", "", "MCR name")
			cmd.Flags().Int("term", 0, "Contract term in months (1, 12, 24, or 36)")
			cmd.Flags().Int("port-speed", 0, "Port speed in Mbps")
			cmd.Flags().Int("location-id", 0, "Location ID where the MCR will be provisioned")
			cmd.Flags().Int("mcr-asn", 0, "ASN for the MCR")
			cmd.Flags().String("diversity-zone", "", "Diversity zone for the MCR")
			cmd.Flags().String("cost-centre", "", "Cost centre for billing")
			cmd.Flags().String("promo-code", "", "Promotional code for discounts")
			cmd.Flags().String("json", "", "JSON string containing MCR configuration")
			cmd.Flags().String("json-file", "", "Path to JSON file containing MCR configuration")

			// Set interactive flag if needed
			if tt.interactive {
				if err := cmd.Flags().Set("interactive", "true"); err != nil {
					t.Fatalf("Failed to set interactive flag: %v", err)
				}
			}

			// Set flag values for this test
			for flagName, flagValue := range tt.flags {
				err := cmd.Flags().Set(flagName, flagValue)
				if err != nil {
					t.Fatalf("Failed to set %s flag: %v", flagName, err)
				}
			}

			// Execute command and capture output
			var err error
			output := output.CaptureOutput(func() {
				err = cmd.RunE(cmd, tt.args)
			})

			// Check results
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, output, tt.expectedOutput)

				// Verify request details if applicable
				if tt.expectedOutput != "" && mockMCRService != nil && mockMCRService.CapturedBuyMCRRequest != nil {
					req := mockMCRService.CapturedBuyMCRRequest

					if tt.flags != nil && tt.flags["json"] != "" {
						// For JSON mode
						assert.Equal(t, "JSON MCR", req.Name)
						assert.Equal(t, 24, req.Term)
						assert.Equal(t, 10000, req.PortSpeed)
						assert.Equal(t, 123, req.LocationID)
						assert.Equal(t, 65000, req.MCRAsn)
						assert.Equal(t, "green", req.DiversityZone)
						assert.Equal(t, "cost-789", req.CostCentre)
						assert.Equal(t, "JSONPROMO", req.PromoCode)
					} else if tt.flags != nil {
						// For flag mode
						assert.Equal(t, "Flag MCR", req.Name)
						assert.Equal(t, 12, req.Term)
						assert.Equal(t, 10000, req.PortSpeed)
						assert.Equal(t, 123, req.LocationID)
						assert.Equal(t, 65000, req.MCRAsn)
						assert.Equal(t, "blue", req.DiversityZone)
						assert.Equal(t, "cost-456", req.CostCentre)
						assert.Equal(t, "FLAGPROMO", req.PromoCode)
					} else if len(tt.prompts) > 0 {
						// For interactive mode
						assert.Equal(t, "Test MCR", req.Name)
						assert.Equal(t, 12, req.Term)
						assert.Equal(t, 10000, req.PortSpeed)
						assert.Equal(t, 123, req.LocationID)
						assert.Equal(t, 65000, req.MCRAsn)
						assert.Equal(t, "red", req.DiversityZone)
						assert.Equal(t, "cost-123", req.CostCentre)
						assert.Equal(t, "MCRPROMO2025", req.PromoCode)
					}
				}
			}
		})
	}
}

func TestUpdateMCRCmd_WithMockClient(t *testing.T) {
	// Save original functions and restore after test
	originalPrompt := utils.Prompt
	originalLoginFunc := config.LoginFunc
	originalUpdateMCRFunc := updateMCRFunc
	defer func() {
		utils.Prompt = originalPrompt
		config.LoginFunc = originalLoginFunc
		updateMCRFunc = originalUpdateMCRFunc
	}()

	tests := []struct {
		name                  string
		mcrUID                string
		interactive           bool
		prompts               []string
		flags                 map[string]string
		jsonInput             string
		jsonFilePath          string
		setupMock             func(*MockMCRService)
		expectedError         string
		expectedOutput        string
		skipRequestValidation bool
	}{
		{
			name:        "interactive mode success",
			mcrUID:      "mcr-123",
			interactive: true,
			prompts: []string{
				"Updated MCR", // name
				"cost-123",    // cost centre
				"24",          // term
			},
			setupMock: func(m *MockMCRService) {
				m.ModifyMCRResult = &megaport.ModifyMCRResponse{
					IsUpdated: true,
				}
			},
			expectedOutput: "MCR updated",
		},
		{
			name:   "flag mode success",
			mcrUID: "mcr-456",
			flags: map[string]string{
				"name":                   "Updated Flag MCR",
				"cost-centre":            "cost-456",
				"marketplace-visibility": "true",
				"term":                   "36",
			},
			setupMock: func(m *MockMCRService) {
				m.ModifyMCRResult = &megaport.ModifyMCRResponse{
					IsUpdated: true,
				}
			},
			expectedOutput: "MCR updated",
		},
		{
			name:   "JSON string mode success",
			mcrUID: "mcr-789",
			flags: map[string]string{
				"json": `{"name":"Updated JSON MCR","costCentre":"cost-789","marketplaceVisibility":true,"contractTermMonths":36}`,
			},
			setupMock: func(m *MockMCRService) {
				m.ModifyMCRResult = &megaport.ModifyMCRResponse{
					IsUpdated: true,
				}
			},
			expectedOutput: "MCR updated",
		},
		{
			name: "missing mcrUID",
			flags: map[string]string{
				"name": "Updated MCR",
			},
			expectedError:         "mcr UID is required",
			skipRequestValidation: true,
		},
		{
			name:   "missing required fields in flag mode",
			mcrUID: "mcr-123",
			flags: map[string]string{
				// Missing name
				"cost-centre": "cost-123",
			},
			expectedError:         "name is required",
			skipRequestValidation: true,
		},
		{
			name:   "invalid term in flag mode",
			mcrUID: "mcr-123",
			flags: map[string]string{
				"name": "Updated MCR",
				"term": "13", // Invalid term
			},
			expectedError:         "invalid term",
			skipRequestValidation: true,
		},
		{
			name:   "invalid JSON",
			mcrUID: "mcr-123",
			flags: map[string]string{
				"json": `{"name":"Updated MCR","marketplaceVisibility":"invalid-boolean"}`,
			},
			expectedError:         "error parsing JSON",
			skipRequestValidation: true,
		},
		{
			name:   "API error",
			mcrUID: "mcr-123",
			flags: map[string]string{
				"name": "Updated MCR",
			},
			setupMock: func(m *MockMCRService) {
				m.ModifyMCRErr = fmt.Errorf("API error: service unavailable")
			},
			expectedError:         "API error: service unavailable",
			skipRequestValidation: true,
		},
		{
			name:                  "no input provided",
			mcrUID:                "mcr-123",
			expectedError:         "no input provided",
			skipRequestValidation: true,
		},
		{
			name:        "update unsuccessful",
			mcrUID:      "mcr-123",
			interactive: true,
			prompts: []string{
				"Updated MCR", // name
				"cost-123",    // cost centre
				"24",          // term
			},
			setupMock: func(m *MockMCRService) {
				m.ModifyMCRResult = &megaport.ModifyMCRResponse{
					IsUpdated: false,
				}
			},
			expectedOutput: "MCR update request was not successful",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock prompt
			if len(tt.prompts) > 0 {
				promptIndex := 0
				utils.Prompt = func(msg string, _ bool) (string, error) {
					if promptIndex < len(tt.prompts) {
						response := tt.prompts[promptIndex]
						promptIndex++
						return response, nil
					}
					return "", fmt.Errorf("unexpected prompt call")
				}
			}

			// Setup mock MCR service
			mockMCRService := &MockMCRService{}
			if tt.setupMock != nil {
				tt.setupMock(mockMCRService)
			}

			// Setup login to return our mock client
			config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
				client := &megaport.Client{}
				client.MCRService = mockMCRService
				return client, nil
			}

			// Create a fresh command for each test to avoid flag conflicts
			cmd := &cobra.Command{
				Use:  "update [mcrUID]",
				Args: cobra.ExactArgs(1),
				RunE: testCommandAdapter(UpdateMCR),
			}

			// Add all the necessary flags
			cmd.Flags().BoolP("interactive", "i", false, "Use interactive mode with prompts")
			cmd.Flags().String("name", "", "New MCR name")
			cmd.Flags().String("cost-centre", "", "Cost centre for billing")
			cmd.Flags().Bool("marketplace-visibility", false, "Whether the MCR is visible in marketplace")
			cmd.Flags().Int("term", 0, "New contract term in months (1, 12, 24, or 36)")
			cmd.Flags().String("json", "", "JSON string containing MCR configuration")
			cmd.Flags().String("json-file", "", "Path to JSON file containing MCR configuration")

			// Set interactive flag if needed
			if tt.interactive {
				if err := cmd.Flags().Set("interactive", "true"); err != nil {
					t.Fatalf("Failed to set interactive flag: %v", err)
				}
			}

			// Set flag values for this test
			for flagName, flagValue := range tt.flags {
				err := cmd.Flags().Set(flagName, flagValue)
				if err != nil {
					t.Fatalf("Failed to set %s flag: %v", flagName, err)
				}
			}

			// Execute command and capture output
			var err error
			var args []string
			if tt.mcrUID != "" {
				args = []string{tt.mcrUID}
			}

			output := output.CaptureOutput(func() {
				err = cmd.RunE(cmd, args)
			})

			// Check results
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, output, tt.expectedOutput)

				// Verify request details if applicable
				if !tt.skipRequestValidation && mockMCRService.CapturedModifyMCRRequest != nil {
					req := mockMCRService.CapturedModifyMCRRequest
					assert.Equal(t, tt.mcrUID, req.MCRID)

					if tt.flags != nil && tt.flags["json"] != "" {
						// For JSON mode
						assert.Equal(t, "Updated JSON MCR", req.Name)
						assert.Equal(t, "cost-789", req.CostCentre)
						assert.NotNil(t, req.MarketplaceVisibility)
						assert.True(t, *req.MarketplaceVisibility)
						assert.NotNil(t, req.ContractTermMonths)
						assert.Equal(t, 36, *req.ContractTermMonths)
					} else if tt.flags != nil && !tt.interactive {
						// For flag mode
						assert.Equal(t, "Updated Flag MCR", req.Name)
						assert.Equal(t, "cost-456", req.CostCentre)
						if req.MarketplaceVisibility != nil {
							assert.True(t, *req.MarketplaceVisibility)
						} else {
							assert.Fail(t, "MarketplaceVisibility should not be nil")
						}
						if req.ContractTermMonths != nil {
							assert.Equal(t, 36, *req.ContractTermMonths)
						} else {
							assert.Fail(t, "ContractTermMonths should not be nil")
						}
					} else if len(tt.prompts) > 0 {
						// For interactive mode
						assert.Equal(t, "Updated MCR", req.Name)
						assert.Equal(t, "cost-123", req.CostCentre)
						assert.NotNil(t, req.ContractTermMonths)
						assert.Equal(t, 24, *req.ContractTermMonths)
					}
				}
			}
		})
	}
}

// TestCreateMCRPrefixFilterListCmd tests the createMCRPrefixFilterListCmd with all three input modes
func TestCreateMCRPrefixFilterListCmd(t *testing.T) {
	// Save original functions and restore after test
	originalPrompt := utils.Prompt
	originalLoginFunc := config.LoginFunc
	originalCreateMCRPrefixFilterListFunc := createMCRPrefixFilterListFunc
	defer func() {
		utils.Prompt = originalPrompt
		config.LoginFunc = originalLoginFunc
		createMCRPrefixFilterListFunc = originalCreateMCRPrefixFilterListFunc
	}()

	tests := []struct {
		name                  string
		args                  []string
		interactive           bool
		prompts               []string
		flags                 map[string]string
		jsonInput             string
		jsonFilePath          string
		setupMock             func(*MockMCRService)
		expectedError         string
		expectedOutput        string
		skipRequestValidation bool
	}{
		{
			name:        "interactive mode success",
			args:        []string{"mcr-123"},
			interactive: true,
			prompts: []string{
				"Test List",      // description
				"IPv4",           // address family
				"192.168.0.0/24", // prefix
				"permit",         // action
				"24",             // ge
				"30",             // le
				"",               // end entries
			},
			setupMock: func(m *MockMCRService) {
				m.CreateMCRPrefixFilterListResult = &megaport.CreateMCRPrefixFilterListResponse{
					IsCreated:          true,
					PrefixFilterListID: 101,
				}
			},
			expectedOutput: "Prefix filter list created successfully - ID: 101",
		},
		{
			name: "flag mode success",
			args: []string{"mcr-123"},
			flags: map[string]string{
				"description":    "Flag List",
				"address-family": "IPv4",
				"entries":        `[{"action":"permit","prefix":"10.0.0.0/8","ge":16,"le":24},{"action":"deny","prefix":"172.16.0.0/12"}]`,
			},
			setupMock: func(m *MockMCRService) {
				m.CreateMCRPrefixFilterListResult = &megaport.CreateMCRPrefixFilterListResponse{
					IsCreated:          true,
					PrefixFilterListID: 102,
				}
			},
			expectedOutput: "Prefix filter list created successfully - ID: 102",
		},
		{
			name: "JSON string mode success",
			args: []string{"mcr-123"},
			flags: map[string]string{
				"json": `{"description":"JSON List","addressFamily":"IPv6","entries":[{"action":"permit","prefix":"2001:db8::/32","ge":48,"le":64}]}`,
			},
			setupMock: func(m *MockMCRService) {
				m.CreateMCRPrefixFilterListResult = &megaport.CreateMCRPrefixFilterListResponse{
					IsCreated:          true,
					PrefixFilterListID: 103,
				}
			},
			expectedOutput: "Prefix filter list created successfully - ID: 103",
		},
		{
			name: "missing MCR UID",
			args: []string{},
			flags: map[string]string{
				"description":    "Test List",
				"address-family": "IPv4",
				"entries":        `[{"action":"permit","prefix":"10.0.0.0/8"}]`,
			},
			expectedError:         "mcr UID is required",
			skipRequestValidation: true,
		},
		{
			name: "missing required fields in flag mode",
			args: []string{"mcr-123"},
			flags: map[string]string{
				"description": "Test List",
				// Missing address family
			},
			expectedError:         "address family is required",
			skipRequestValidation: true,
		},
		{
			name: "invalid address family in flag mode",
			args: []string{"mcr-123"},
			flags: map[string]string{
				"description":    "Test List",
				"address-family": "IPvX", // Invalid
				"entries":        `[{"action":"permit","prefix":"10.0.0.0/8"}]`,
			},
			expectedError:         "invalid address family, must be IPv4 or IPv6",
			skipRequestValidation: true,
		},
		{
			name: "invalid JSON in entries",
			args: []string{"mcr-123"},
			flags: map[string]string{
				"description":    "Test List",
				"address-family": "IPv4",
				"entries":        `[{"action":"INVALID","prefix":"10.0.0.0/8"}]`,
			},
			expectedError:         "invalid action",
			skipRequestValidation: true,
		},
		{
			name: "invalid JSON format",
			args: []string{"mcr-123"},
			flags: map[string]string{
				"json": `{"description":"Test List","addressFamily":123}`,
			},
			expectedError:         "error parsing JSON",
			skipRequestValidation: true,
		},
		{
			name: "API error",
			args: []string{"mcr-123"},
			flags: map[string]string{
				"description":    "Test List",
				"address-family": "IPv4",
				"entries":        `[{"action":"permit","prefix":"10.0.0.0/8"}]`,
			},
			setupMock: func(m *MockMCRService) {
				m.CreateMCRPrefixFilterListErr = fmt.Errorf("API error: service unavailable")
			},
			expectedError:         "API error: service unavailable",
			skipRequestValidation: true,
		},
		{
			name:                  "no input provided",
			args:                  []string{"mcr-123"},
			expectedError:         "no input provided, use --interactive, --json, or flags to specify prefix filter list details",
			skipRequestValidation: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock prompt
			if len(tt.prompts) > 0 {
				promptIndex := 0
				utils.Prompt = func(msg string, _ bool) (string, error) {
					if promptIndex < len(tt.prompts) {
						response := tt.prompts[promptIndex]
						promptIndex++
						return response, nil
					}
					return "", fmt.Errorf("unexpected prompt call")
				}
			}

			// Setup mock MCR service
			mockMCRService := &MockMCRService{}
			if tt.setupMock != nil {
				tt.setupMock(mockMCRService)
			}

			// Setup login to return our mock client
			config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
				client := &megaport.Client{}
				client.MCRService = mockMCRService
				return client, nil
			}

			// Create a fresh command for each test to avoid flag conflicts
			cmd := &cobra.Command{
				Use:  "create-prefix-filter-list [mcrUID]",
				Args: cobra.ExactArgs(1),
				RunE: testCommandAdapter(CreateMCRPrefixFilterList),
			}

			// Add all the necessary flags
			cmd.Flags().BoolP("interactive", "i", false, "Use interactive mode with prompts")
			cmd.Flags().String("description", "", "Description of the prefix filter list")
			cmd.Flags().String("address-family", "", "Address family (IPv4 or IPv6)")
			cmd.Flags().String("entries", "", "JSON string array of prefix filter entries")
			cmd.Flags().String("json", "", "JSON string containing prefix filter list configuration")
			cmd.Flags().String("json-file", "", "Path to JSON file containing prefix filter list configuration")

			// Set interactive flag if needed
			if tt.interactive {
				if err := cmd.Flags().Set("interactive", "true"); err != nil {
					t.Fatalf("Failed to set interactive flag: %v", err)
				}
			}

			// Set flag values for this test
			for flagName, flagValue := range tt.flags {
				err := cmd.Flags().Set(flagName, flagValue)
				if err != nil {
					t.Fatalf("Failed to set %s flag: %v", flagName, err)
				}
			}

			// Execute command and capture output
			var err error
			output := output.CaptureOutput(func() {
				err = cmd.RunE(cmd, tt.args)
			})

			// Check results
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, output, tt.expectedOutput)

				// Verify request details if applicable
				if !tt.skipRequestValidation && mockMCRService.CapturedCreatePrefixFilterListRequest != nil {
					req := mockMCRService.CapturedCreatePrefixFilterListRequest
					assert.Equal(t, tt.args[0], req.MCRID)

					if tt.flags != nil && tt.flags["json"] != "" {
						// For JSON mode
						assert.Equal(t, "JSON List", req.PrefixFilterList.Description)
						assert.Equal(t, "IPv6", req.PrefixFilterList.AddressFamily)
						assert.Len(t, req.PrefixFilterList.Entries, 1)
						assert.Equal(t, "permit", req.PrefixFilterList.Entries[0].Action)
						assert.Equal(t, "2001:db8::/32", req.PrefixFilterList.Entries[0].Prefix)
						assert.Equal(t, 48, req.PrefixFilterList.Entries[0].Ge)
						assert.Equal(t, 64, req.PrefixFilterList.Entries[0].Le)
					} else if tt.flags != nil && !tt.interactive {
						// For flag mode
						assert.Equal(t, "Flag List", req.PrefixFilterList.Description)
						assert.Equal(t, "IPv4", req.PrefixFilterList.AddressFamily)
						assert.Len(t, req.PrefixFilterList.Entries, 2)
						assert.Equal(t, "permit", req.PrefixFilterList.Entries[0].Action)
						assert.Equal(t, "10.0.0.0/8", req.PrefixFilterList.Entries[0].Prefix)
						assert.Equal(t, 16, req.PrefixFilterList.Entries[0].Ge)
						assert.Equal(t, 24, req.PrefixFilterList.Entries[0].Le)
						assert.Equal(t, "deny", req.PrefixFilterList.Entries[1].Action)
						assert.Equal(t, "172.16.0.0/12", req.PrefixFilterList.Entries[1].Prefix)
					} else if len(tt.prompts) > 0 {
						// For interactive mode
						assert.Equal(t, "Test List", req.PrefixFilterList.Description)
						assert.Equal(t, "IPv4", req.PrefixFilterList.AddressFamily)
						assert.Len(t, req.PrefixFilterList.Entries, 1)
						assert.Equal(t, "permit", req.PrefixFilterList.Entries[0].Action)
						assert.Equal(t, "192.168.0.0/24", req.PrefixFilterList.Entries[0].Prefix)
						assert.Equal(t, 24, req.PrefixFilterList.Entries[0].Ge)
						assert.Equal(t, 30, req.PrefixFilterList.Entries[0].Le)
					}
				}
			}
		})
	}
}

// TestUpdateMCRPrefixFilterListCmd tests the updateMCRPrefixFilterListCmd with all three input modes
func TestUpdateMCRPrefixFilterListCmd(t *testing.T) {
	// Save original functions and restore after test
	originalPrompt := utils.Prompt
	originalLoginFunc := config.LoginFunc
	originalModifyMCRPrefixFilterListFunc := modifyMCRPrefixFilterListFunc
	originalGetMCRPrefixFilterListFunc := getMCRPrefixFilterListFunc
	defer func() {
		utils.Prompt = originalPrompt
		config.LoginFunc = originalLoginFunc
		modifyMCRPrefixFilterListFunc = originalModifyMCRPrefixFilterListFunc
		getMCRPrefixFilterListFunc = originalGetMCRPrefixFilterListFunc
	}()

	tests := []struct {
		name                  string
		args                  []string
		interactive           bool
		prompts               []string
		flags                 map[string]string
		jsonInput             string
		jsonFilePath          string
		setupMock             func(*MockMCRService)
		expectedError         string
		expectedOutput        string
		skipRequestValidation bool
	}{
		{
			name:        "interactive mode success",
			args:        []string{"mcr-123", "101"},
			interactive: true,
			prompts: []string{
				"Updated List",   // description
				"IPv4",           // address family
				"yes",            // modify existing entries
				"yes",            // keep entry 1
				"yes",            // modify entry 1
				"192.168.1.0/24", // new prefix
				"permit",         // new action
				"24",             // new ge
				"30",             // new le
				"yes",            // add new entries
				"172.16.0.0/12",  // prefix for new entry
				"deny",           // action for new entry
				"16",             // ge for new entry
				"24",             // le for new entry
				"",               // end new entries
			},
			setupMock: func(m *MockMCRService) {
				m.GetMCRPrefixFilterListResult = &megaport.MCRPrefixFilterList{
					ID:            101,
					Description:   "Test List",
					AddressFamily: "IPv4",
					Entries: []*megaport.MCRPrefixListEntry{
						{
							Action: "permit",
							Prefix: "192.168.0.0/24",
							Ge:     24,
							Le:     30,
						},
					},
				}
				m.ModifyMCRPrefixFilterListResult = &megaport.ModifyMCRPrefixFilterListResponse{
					IsUpdated: true,
				}
			},
			expectedOutput: "Prefix filter list updated successfully - ID: 101",
		},
		{
			name: "flag mode success",
			args: []string{"mcr-123", "102"},
			flags: map[string]string{
				"description":    "Updated Flag List",
				"address-family": "IPv4",
				"entries":        `[{"action":"permit","prefix":"10.0.0.0/8","ge":16,"le":24},{"action":"deny","prefix":"172.16.0.0/12"}]`,
			},
			setupMock: func(m *MockMCRService) {
				m.ModifyMCRPrefixFilterListResult = &megaport.ModifyMCRPrefixFilterListResponse{
					IsUpdated: true,
				}
			},
			expectedOutput: "Prefix filter list updated successfully - ID: 102",
		},
		{
			name: "JSON string mode success",
			args: []string{"mcr-123", "103"},
			flags: map[string]string{
				"json": `{"description":"Updated JSON List","addressFamily":"IPv6","entries":[{"action":"permit","prefix":"2001:db8::/32","ge":48,"le":64}]}`,
			},
			setupMock: func(m *MockMCRService) {
				m.ModifyMCRPrefixFilterListResult = &megaport.ModifyMCRPrefixFilterListResponse{
					IsUpdated: true,
				}
			},
			expectedOutput: "Prefix filter list updated successfully - ID: 103",
		},
		{
			name: "missing arguments",
			args: []string{"mcr-123"},
			flags: map[string]string{
				"description":    "Updated List",
				"address-family": "IPv4",
				"entries":        `[{"action":"permit","prefix":"10.0.0.0/8"}]`,
			},
			expectedError:         "mcr UID and prefix filter list ID are required",
			skipRequestValidation: true,
		},
		{
			name: "missing required fields in flag mode",
			args: []string{"mcr-123", "101"},
			flags: map[string]string{
				"description": "Updated List",
				// Missing address family
			},
			expectedError:         "address family is required",
			skipRequestValidation: true,
		},
		{
			name: "invalid address family in flag mode",
			args: []string{"mcr-123", "101"},
			flags: map[string]string{
				"description":    "Updated List",
				"address-family": "IPvX", // Invalid
				"entries":        `[{"action":"permit","prefix":"10.0.0.0/8"}]`,
			},
			expectedError:         "invalid address family, must be IPv4 or IPv6",
			skipRequestValidation: true,
		},
		{
			name: "invalid JSON format for entries",
			args: []string{"mcr-123", "101"},
			flags: map[string]string{
				"description":    "Updated List",
				"address-family": "IPv4",
				"entries":        `[{"action":"INVALID","prefix":"10.0.0.0/8"}]`,
			},
			expectedError:         "invalid action",
			skipRequestValidation: true,
		},
		{
			name: "API error",
			args: []string{"mcr-123", "101"},
			flags: map[string]string{
				"description":    "Updated List",
				"address-family": "IPv4",
				"entries":        `[{"action":"permit","prefix":"10.0.0.0/8"}]`,
			},
			setupMock: func(m *MockMCRService) {
				m.ModifyMCRPrefixFilterListErr = fmt.Errorf("API error: service unavailable")
			},
			expectedError:         "API error: service unavailable",
			skipRequestValidation: true,
		},
		{
			name:                  "no input provided",
			args:                  []string{"mcr-123", "101"},
			expectedError:         "no input provided, use --interactive, --json, or flags to specify prefix filter list update details",
			skipRequestValidation: true,
		},
		{
			name:        "update unsuccessful",
			args:        []string{"mcr-123", "101"},
			interactive: true,
			prompts: []string{
				"Updated List", // description
				"IPv4",         // address family
				"no",           // don't modify existing entries
				"no",           // don't add new entries
			},
			setupMock: func(m *MockMCRService) {
				m.GetMCRPrefixFilterListResult = &megaport.MCRPrefixFilterList{
					ID:            101,
					Description:   "Test List",
					AddressFamily: "IPv4",
					Entries: []*megaport.MCRPrefixListEntry{
						{
							Action: "permit",
							Prefix: "192.168.0.0/24",
							Ge:     24,
							Le:     30,
						},
					},
				}
				m.ModifyMCRPrefixFilterListResult = &megaport.ModifyMCRPrefixFilterListResponse{
					IsUpdated: false,
				}
			},
			expectedOutput:        "Prefix filter list update request was not successful",
			skipRequestValidation: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock prompt
			if len(tt.prompts) > 0 {
				promptIndex := 0
				utils.Prompt = func(msg string, _ bool) (string, error) {
					if promptIndex < len(tt.prompts) {
						response := tt.prompts[promptIndex]
						promptIndex++
						return response, nil
					}
					return "", fmt.Errorf("unexpected prompt call")
				}
			}

			// Setup mock MCR service
			mockMCRService := &MockMCRService{}
			if tt.setupMock != nil {
				tt.setupMock(mockMCRService)
			}

			// Setup login to return our mock client
			config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
				client := &megaport.Client{}
				client.MCRService = mockMCRService
				return client, nil
			}

			// Create a fresh command for each test to avoid flag conflicts
			cmd := &cobra.Command{
				Use:  "update-prefix-filter-list [mcrUID] [prefixFilterListID]",
				Args: cobra.ExactArgs(2),
				RunE: testCommandAdapter(UpdateMCRPrefixFilterList),
			}

			// Add all the necessary flags
			cmd.Flags().BoolP("interactive", "i", false, "Use interactive mode with prompts")
			cmd.Flags().String("description", "", "Description of the prefix filter list")
			cmd.Flags().String("address-family", "", "Address family (IPv4 or IPv6)")
			cmd.Flags().String("entries", "", "JSON string array of prefix filter entries")
			cmd.Flags().String("json", "", "JSON string containing prefix filter list configuration")
			cmd.Flags().String("json-file", "", "Path to JSON file containing prefix filter list configuration")

			// Set interactive flag if needed
			if tt.interactive {
				if err := cmd.Flags().Set("interactive", "true"); err != nil {
					t.Fatalf("Failed to set interactive flag: %v", err)
				}
			}

			// Set flag values for this test
			for flagName, flagValue := range tt.flags {
				err := cmd.Flags().Set(flagName, flagValue)
				if err != nil {
					t.Fatalf("Failed to set %s flag: %v", flagName, err)
				}
			}

			// Execute command and capture output
			var err error
			output := output.CaptureOutput(func() {
				err = cmd.RunE(cmd, tt.args)
			})

			// Check results
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, output, tt.expectedOutput)

				// Verify request details if applicable
				if !tt.skipRequestValidation && mockMCRService.CapturedModifyPrefixFilterListMCRID != "" {
					assert.Equal(t, tt.args[0], mockMCRService.CapturedModifyPrefixFilterListMCRID)

					prefixIDInt, _ := strconv.Atoi(tt.args[1])
					assert.Equal(t, prefixIDInt, mockMCRService.CapturedModifyPrefixFilterListID)

					if mockMCRService.CapturedModifyPrefixFilterList != nil {
						prefixList := mockMCRService.CapturedModifyPrefixFilterList

						if tt.flags != nil && tt.flags["json"] != "" {
							// For JSON mode
							assert.Equal(t, "Updated JSON List", prefixList.Description)
							assert.Equal(t, "IPv6", prefixList.AddressFamily)
							assert.Len(t, prefixList.Entries, 1)
							assert.Equal(t, "permit", prefixList.Entries[0].Action)
							assert.Equal(t, "2001:db8::/32", prefixList.Entries[0].Prefix)
							assert.Equal(t, 48, prefixList.Entries[0].Ge)
							assert.Equal(t, 64, prefixList.Entries[0].Le)
						} else if tt.flags != nil && !tt.interactive {
							// For flag mode
							assert.Equal(t, "Updated Flag List", prefixList.Description)
							assert.Equal(t, "IPv4", prefixList.AddressFamily)
							assert.Len(t, prefixList.Entries, 2)
							assert.Equal(t, "permit", prefixList.Entries[0].Action)
							assert.Equal(t, "10.0.0.0/8", prefixList.Entries[0].Prefix)
							assert.Equal(t, 16, prefixList.Entries[0].Ge)
							assert.Equal(t, 24, prefixList.Entries[0].Le)
							assert.Equal(t, "deny", prefixList.Entries[1].Action)
							assert.Equal(t, "172.16.0.0/12", prefixList.Entries[1].Prefix)
						} else if len(tt.prompts) > 0 && !tt.skipRequestValidation {
							// For interactive mode (only when it should be successful)
							assert.Equal(t, "Updated List", prefixList.Description)
							assert.Equal(t, "IPv4", prefixList.AddressFamily)
							assert.GreaterOrEqual(t, len(prefixList.Entries), 1)

							// First entry should be modified
							assert.Equal(t, "permit", prefixList.Entries[0].Action)
							assert.Equal(t, "192.168.1.0/24", prefixList.Entries[0].Prefix)
							assert.Equal(t, 24, prefixList.Entries[0].Ge)
							assert.Equal(t, 30, prefixList.Entries[0].Le)

							// Second entry should be the newly added one
							if len(prefixList.Entries) > 1 {
								assert.Equal(t, "deny", prefixList.Entries[1].Action)
								assert.Equal(t, "172.16.0.0/12", prefixList.Entries[1].Prefix)
								assert.Equal(t, 16, prefixList.Entries[1].Ge)
								assert.Equal(t, 24, prefixList.Entries[1].Le)
							}
						}
					}
				}
			}
		})
	}
}
