package mcr

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/commands/config"
	"github.com/megaport/megaport-cli/internal/utils"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func testCommandAdapter(fn func(cmd *cobra.Command, args []string, noColor bool) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return fn(cmd, args, false)
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
	originalLoginFunc := config.LoginFunc
	originalPrompt := utils.ResourcePrompt
	originalConfirmPrompt := utils.ConfirmPrompt
	defer func() {
		config.LoginFunc = originalLoginFunc
		utils.ResourcePrompt = originalPrompt
		utils.ConfirmPrompt = originalConfirmPrompt
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
			expectedOutput: "MCR deleted",
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
			expectedOutput: "MCR deleted",
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
			expectedOutput: "MCR deleted",
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
			mockMCRService := &MockMCRService{}
			tt.setupMock(mockMCRService)

			config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
				client := &megaport.Client{}
				client.MCRService = mockMCRService
				return client, nil
			}

			utils.ConfirmPrompt = func(message string, _ bool) bool {
				assert.Contains(t, message, fmt.Sprintf("Are you sure you want to delete MCR %s?", tt.mcrID))
				return tt.promptResponse == "y"
			}

			utils.ResourcePrompt = func(_, msg string, _ bool) (string, error) {
				assert.Contains(t, msg, fmt.Sprintf("Are you sure you want to delete MCR %s?", tt.mcrID))
				return tt.promptResponse, nil
			}

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

			output := output.CaptureOutput(func() {
				err = cmd.RunE(cmd, []string{tt.mcrID})
			})

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, output, tt.expectedOutput)

				if tt.expectDeleted {
					assert.NotNil(t, mockMCRService.CapturedDeleteMCRUID)
					assert.Equal(t, tt.mcrID, mockMCRService.CapturedDeleteMCRUID)
				}
			}
		})
	}
}

func TestRestoreMCRCmd_WithMockClient(t *testing.T) {
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
			mockMCRService := &MockMCRService{}
			tt.setupMock(mockMCRService)

			config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
				client := &megaport.Client{}
				client.MCRService = mockMCRService
				return client, nil
			}

			restoreMCRCmd := &cobra.Command{
				Use: "restore-mcr [mcrID]",
				RunE: func(cmd *cobra.Command, args []string) error {
					return RestoreMCR(cmd, args, false)
				},
			}

			var err error
			output := output.CaptureOutput(func() {
				err = restoreMCRCmd.RunE(restoreMCRCmd, []string{tt.mcrID})
			})

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
	originalPrompt := utils.ResourcePrompt
	defer func() {
		config.LoginFunc = originalLoginFunc
		utils.ResourcePrompt = originalPrompt
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

			utils.ResourcePrompt = func(_, msg string, _ bool) (string, error) {
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
	originalPrompt := utils.ResourcePrompt
	originalLoginFunc := config.LoginFunc
	originalBuyMCRFunc := buyMCRFunc
	defer func() {
		utils.ResourcePrompt = originalPrompt
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
				"Test MCR",
				"12",
				"10000",
				"123",
				"65000",
				"red",
				"cost-123",
				"MCRPROMO2025",
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
			expectedError: "Invalid contract term: 0 - must be one of: [1 12 24 36]",
		},
		{
			name: "invalid term in flag mode",
			flags: map[string]string{
				"name":        "Test MCR",
				"term":        "13",
				"port-speed":  "10000",
				"location-id": "123",
				"mcr-asn":     "65000",
			},
			expectedError: "Invalid contract term: 13 - must be one of: [1 12 24 36]",
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
			if len(tt.prompts) > 0 {
				promptIndex := 0
				utils.ResourcePrompt = func(_, msg string, _ bool) (string, error) {
					if promptIndex < len(tt.prompts) {
						response := tt.prompts[promptIndex]
						promptIndex++
						return response, nil
					}
					return "", fmt.Errorf("unexpected prompt call")
				}
			}

			mockMCRService := &MockMCRService{}
			if tt.setupMock != nil {
				tt.setupMock(mockMCRService)
			}

			config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
				client := &megaport.Client{}
				client.MCRService = mockMCRService
				return client, nil
			}

			cmd := &cobra.Command{
				Use:  "buy",
				RunE: testCommandAdapter(BuyMCR),
			}

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

			if tt.interactive {
				if err := cmd.Flags().Set("interactive", "true"); err != nil {
					t.Fatalf("Failed to set interactive flag: %v", err)
				}
			}

			for flagName, flagValue := range tt.flags {
				err := cmd.Flags().Set(flagName, flagValue)
				if err != nil {
					t.Fatalf("Failed to set %s flag: %v", flagName, err)
				}
			}

			var err error
			output := output.CaptureOutput(func() {
				err = cmd.RunE(cmd, tt.args)
			})

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, output, tt.expectedOutput)

				if tt.expectedOutput != "" && mockMCRService != nil && mockMCRService.CapturedBuyMCRRequest != nil {
					req := mockMCRService.CapturedBuyMCRRequest

					if tt.flags != nil && tt.flags["json"] != "" {
						assert.Equal(t, "JSON MCR", req.Name)
						assert.Equal(t, 24, req.Term)
						assert.Equal(t, 10000, req.PortSpeed)
						assert.Equal(t, 123, req.LocationID)
						assert.Equal(t, 65000, req.MCRAsn)
						assert.Equal(t, "green", req.DiversityZone)
						assert.Equal(t, "cost-789", req.CostCentre)
						assert.Equal(t, "JSONPROMO", req.PromoCode)
					} else if tt.flags != nil {
						assert.Equal(t, "Flag MCR", req.Name)
						assert.Equal(t, 12, req.Term)
						assert.Equal(t, 10000, req.PortSpeed)
						assert.Equal(t, 123, req.LocationID)
						assert.Equal(t, 65000, req.MCRAsn)
						assert.Equal(t, "blue", req.DiversityZone)
						assert.Equal(t, "cost-456", req.CostCentre)
						assert.Equal(t, "FLAGPROMO", req.PromoCode)
					} else if len(tt.prompts) > 0 {
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

func TestGetMCRStatus(t *testing.T) {
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
		outputFormat   string
	}{
		{
			name:   "successful status retrieval - table format",
			mcrUID: "mcr-123abc",
			setupMock: func(m *MockMCRService) {
				m.GetMCRResult = &megaport.MCR{
					UID:                "mcr-123abc",
					Name:               "Test MCR",
					ProvisioningStatus: "CONFIGURED",
					PortSpeed:          10000,
					Resources: megaport.MCRResources{
						VirtualRouter: megaport.MCRVirtualRouter{
							ASN: 64512,
						},
					},
				}
			},
			expectedOutput: "mcr-123abc",
			outputFormat:   "table",
		},
		{
			name:   "successful status retrieval - json format",
			mcrUID: "mcr-123abc",
			setupMock: func(m *MockMCRService) {
				m.GetMCRResult = &megaport.MCR{
					UID:                "mcr-123abc",
					Name:               "Test MCR",
					ProvisioningStatus: "LIVE",
					PortSpeed:          1000,
					Resources: megaport.MCRResources{
						VirtualRouter: megaport.MCRVirtualRouter{
							ASN: 64513,
						},
					},
				}
			},
			expectedOutput: "mcr-123abc",
			outputFormat:   "json",
		},
		{
			name:   "MCR not found",
			mcrUID: "mcr-notfound",
			setupMock: func(m *MockMCRService) {
				m.GetMCRErr = fmt.Errorf("MCR not found")
			},
			expectedError: "error getting MCR status",
			outputFormat:  "table",
		},
		{
			name:   "API error",
			mcrUID: "mcr-error",
			setupMock: func(m *MockMCRService) {
				m.GetMCRErr = fmt.Errorf("API error")
			},
			expectedError: "API error",
			outputFormat:  "table",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockMCRService{}
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
				client := &megaport.Client{}
				client.MCRService = mockService
				return client, nil
			}

			cmd := &cobra.Command{
				Use: "status [mcrUID]",
			}

			var err error
			capturedOutput := output.CaptureOutput(func() {
				err = GetMCRStatus(cmd, []string{tt.mcrUID}, true, tt.outputFormat)
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
					assert.Contains(t, capturedOutput, "\"asn\":")
				case "table":
					assert.Contains(t, capturedOutput, "UID")
					assert.Contains(t, capturedOutput, "NAME")
					assert.Contains(t, capturedOutput, "STATUS")
					assert.Contains(t, capturedOutput, "ASN")
				}
			}
		})
	}
}
