package mcr

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/commands/config"
	"github.com/megaport/megaport-cli/internal/testutil"
	"github.com/megaport/megaport-cli/internal/utils"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestGetMCRCmd_WithMockClient(t *testing.T) {
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {})
	defer cleanup()

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

			cmd := testutil.NewCommand("get-mcr [mcrID]", func(cmd *cobra.Command, args []string) error {
				return GetMCR(cmd, args, false, tt.format)
			})
			testutil.SetFlags(t, cmd, map[string]string{"output": tt.format})

			var err error

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
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {})
	defer cleanup()
	originalPrompt := utils.ResourcePrompt
	originalConfirmPrompt := utils.ConfirmPrompt
	defer func() {
		utils.ResourcePrompt = originalPrompt
		utils.ConfirmPrompt = originalConfirmPrompt
	}()

	tests := []struct {
		name           string
		mcrID          string
		force          bool
		deleteNow      bool
		safeDelete     bool
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
			name:           "safe delete flag passed to request",
			mcrID:          "mcr-safe-delete",
			force:          true,
			deleteNow:      false,
			safeDelete:     true,
			promptResponse: "",
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
			expectedError:  "cancelled by user",
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
			cmd.Flags().Bool("safe-delete", false, "Fail if the resource has attached VXCs or other active services")
			flags := map[string]string{
				"force": fmt.Sprintf("%v", tt.force),
				"now":   fmt.Sprintf("%v", tt.deleteNow),
			}
			if tt.safeDelete {
				flags["safe-delete"] = "true"
			}
			testutil.SetFlags(t, cmd, flags)

			var err error
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
					if tt.safeDelete {
						assert.True(t, mockMCRService.CapturedDeleteMCRRequest.SafeDelete)
					}
				}
			}
		})
	}
}

func TestRestoreMCRCmd_WithMockClient(t *testing.T) {
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {})
	defer cleanup()

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
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {})
	defer cleanup()

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
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {})
	defer cleanup()

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
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {})
	defer cleanup()
	originalPrompt := utils.ResourcePrompt
	defer func() {
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
			cmd.Flags().Bool("now", false, "Delete prefix filter list immediately instead of at end of billing cycle")
			testutil.SetFlags(t, cmd, map[string]string{
				"force": fmt.Sprintf("%v", tt.force),
				"now":   fmt.Sprintf("%v", false),
			})

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

func TestBuyMCR_NoWaitFlag(t *testing.T) {
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {})
	defer cleanup()
	originalBuyMCRFunc := buyMCRFunc
	defer func() {
		buyMCRFunc = originalBuyMCRFunc
	}()
	originalBuyConfirmPrompt := utils.BuyConfirmPrompt
	defer func() { utils.BuyConfirmPrompt = originalBuyConfirmPrompt }()
	utils.BuyConfirmPrompt = func(_ string, _ []utils.BuyConfirmDetail, _ bool) bool { return true }

	tests := []struct {
		name                     string
		noWait                   bool
		expectedWaitForProvision bool
	}{
		{
			name:                     "default waits for provisioning",
			noWait:                   false,
			expectedWaitForProvision: true,
		},
		{
			name:                     "no-wait skips provisioning wait",
			noWait:                   true,
			expectedWaitForProvision: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMCRService := &MockMCRService{}
			mockMCRService.BuyMCRResult = &megaport.BuyMCRResponse{
				TechnicalServiceUID: "mcr-uid-123",
			}

			config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
				client := &megaport.Client{}
				client.MCRService = mockMCRService
				return client, nil
			}

			var capturedReq *megaport.BuyMCRRequest
			buyMCRFunc = func(ctx context.Context, client *megaport.Client, req *megaport.BuyMCRRequest) (*megaport.BuyMCRResponse, error) {
				capturedReq = req
				return &megaport.BuyMCRResponse{
					TechnicalServiceUID: "mcr-uid-123",
				}, nil
			}

			cmd := &cobra.Command{Use: "buy"}
			cmd.Flags().BoolP("interactive", "i", false, "")
			cmd.Flags().Bool("no-wait", false, "")
			cmd.Flags().BoolP("yes", "y", false, "")
			cmd.Flags().String("name", "", "")
			cmd.Flags().Int("term", 0, "")
			cmd.Flags().Int("port-speed", 0, "")
			cmd.Flags().Int("location-id", 0, "")
			cmd.Flags().Int("mcr-asn", 0, "")
			cmd.Flags().String("diversity-zone", "", "")
			cmd.Flags().String("cost-centre", "", "")
			cmd.Flags().String("promo-code", "", "")
			cmd.Flags().String("json", "", "")
			cmd.Flags().String("json-file", "", "")

			testutil.SetFlags(t, cmd, map[string]string{
				"name":        "Test MCR",
				"term":        "12",
				"port-speed":  "10000",
				"location-id": "123",
				"mcr-asn":     "65000",
			})
			if tt.noWait {
				assert.NoError(t, cmd.Flags().Set("no-wait", "true"))
			}

			var err error
			output.CaptureOutput(func() {
				err = BuyMCR(cmd, nil, true)
			})

			assert.NoError(t, err)
			assert.NotNil(t, capturedReq)
			assert.Equal(t, tt.expectedWaitForProvision, capturedReq.WaitForProvision)
		})
	}
}

func TestBuyMCRCmd_WithMockClient(t *testing.T) {
	originalPrompt := utils.ResourcePrompt
	originalBuyMCRFunc := buyMCRFunc
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {})
	defer cleanup()
	defer func() {
		utils.ResourcePrompt = originalPrompt
		buyMCRFunc = originalBuyMCRFunc
	}()
	originalBuyConfirmPrompt := utils.BuyConfirmPrompt
	defer func() { utils.BuyConfirmPrompt = originalBuyConfirmPrompt }()
	utils.BuyConfirmPrompt = func(_ string, _ []utils.BuyConfirmDetail, _ bool) bool { return true }

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
		{
			name:        "JSON takes precedence over interactive flag",
			interactive: true,
			flags: map[string]string{
				"json": `{"name":"JSON MCR","term":24,"portSpeed":10000,"locationId":123,"mcrAsn":65000,"diversityZone":"green","costCentre":"cost-789","promoCode":"JSONPROMO"}`,
			},
			setupMock: func(m *MockMCRService) {
				m.BuyMCRResult = &megaport.BuyMCRResponse{
					TechnicalServiceUID: "mcr-json-wins",
				}
			},
			expectedOutput: "MCR created",
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
				RunE: testutil.NoColorAdapter(BuyMCR),
			}

			cmd.Flags().BoolP("interactive", "i", false, "Use interactive mode with prompts")
			cmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
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

			flags := make(map[string]string)
			if tt.interactive {
				flags["interactive"] = "true"
			}
			for k, v := range tt.flags {
				flags[k] = v
			}
			testutil.SetFlags(t, cmd, flags)

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
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {})
	defer cleanup()

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
		{
			name:   "nil MCR returned without error",
			mcrUID: "mcr-nil",
			setupMock: func(m *MockMCRService) {
				m.ForceNilGetMCR = true
			},
			expectedError: "no MCR found",
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

func TestListMCRsCmd_WithMockClient(t *testing.T) {
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {})
	defer cleanup()

	// Sample MCRs for testing
	testMCRs := []*megaport.MCR{
		{
			UID:                "mcr-123",
			Name:               "mcr-demo-01",
			LocationID:         571,
			ProvisioningStatus: "LIVE",
			PortSpeed:          1000,
			Resources: megaport.MCRResources{
				VirtualRouter: megaport.MCRVirtualRouter{
					ASN: 133937,
				},
			},
		},
		{
			UID:                "mcr-456",
			Name:               "mcr-demo0-01",
			LocationID:         558,
			ProvisioningStatus: "LIVE",
			PortSpeed:          1000,
			Resources: megaport.MCRResources{
				VirtualRouter: megaport.MCRVirtualRouter{
					ASN: 133937,
				},
			},
		},
		{
			UID:                "mcr-789",
			Name:               "production-mcr",
			LocationID:         64,
			ProvisioningStatus: "LIVE",
			PortSpeed:          2500,
			Resources: megaport.MCRResources{
				VirtualRouter: megaport.MCRVirtualRouter{
					ASN: 133937,
				},
			},
		},
		{
			UID:                "mcr-abc",
			Name:               "test-mcr-sydney",
			LocationID:         571,
			ProvisioningStatus: "DECOMMISSIONED",
			PortSpeed:          5000,
			Resources: megaport.MCRResources{
				VirtualRouter: megaport.MCRVirtualRouter{
					ASN: 64512,
				},
			},
		},
	}

	tests := []struct {
		name           string
		flags          map[string]string
		setupMock      func(*MockMCRService)
		expectedError  string
		expectedMCRs   []string // Names of MCRs that should be in the output
		unexpectedMCRs []string // Names of MCRs that should NOT be in the output
		outputFormat   string
	}{
		{
			name: "list all active MCRs",
			setupMock: func(m *MockMCRService) {
				m.ListMCRsResult = testMCRs
			},
			expectedMCRs:   []string{"mcr-demo-01", "mcr-demo0-01", "production-mcr"},
			unexpectedMCRs: []string{"test-mcr-sydney"}, // Should be excluded due to DECOMMISSIONED status
			outputFormat:   "table",
		},
		{
			name: "filter by exact name match",
			flags: map[string]string{
				"name": "mcr-demo-01",
			},
			setupMock: func(m *MockMCRService) {
				m.ListMCRsResult = testMCRs
			},
			expectedMCRs:   []string{"mcr-demo-01"},
			unexpectedMCRs: []string{"mcr-demo0-01", "production-mcr", "test-mcr-sydney"},
			outputFormat:   "table",
		},
		{
			name: "filter by partial name match",
			flags: map[string]string{
				"name": "demo",
			},
			setupMock: func(m *MockMCRService) {
				m.ListMCRsResult = testMCRs
			},
			expectedMCRs:   []string{"mcr-demo-01", "mcr-demo0-01"},
			unexpectedMCRs: []string{"production-mcr", "test-mcr-sydney"},
			outputFormat:   "table",
		},
		{
			name: "filter by case insensitive name",
			flags: map[string]string{
				"name": "MCR-DEMO-01",
			},
			setupMock: func(m *MockMCRService) {
				m.ListMCRsResult = testMCRs
			},
			expectedMCRs:   []string{"mcr-demo-01"},
			unexpectedMCRs: []string{"mcr-demo0-01", "production-mcr", "test-mcr-sydney"},
			outputFormat:   "table",
		},
		{
			name: "filter by location ID",
			flags: map[string]string{
				"location-id": "571",
			},
			setupMock: func(m *MockMCRService) {
				m.ListMCRsResult = testMCRs
			},
			expectedMCRs:   []string{"mcr-demo-01"}, // Only active MCR at location 571
			unexpectedMCRs: []string{"mcr-demo0-01", "production-mcr", "test-mcr-sydney"},
			outputFormat:   "table",
		},
		{
			name: "filter by port speed",
			flags: map[string]string{
				"port-speed": "2500",
			},
			setupMock: func(m *MockMCRService) {
				m.ListMCRsResult = testMCRs
			},
			expectedMCRs:   []string{"production-mcr"},
			unexpectedMCRs: []string{"mcr-demo-01", "mcr-demo0-01", "test-mcr-sydney"},
			outputFormat:   "table",
		},
		{
			name: "filter by name and location combined",
			flags: map[string]string{
				"name":        "demo",
				"location-id": "571",
			},
			setupMock: func(m *MockMCRService) {
				m.ListMCRsResult = testMCRs
			},
			expectedMCRs:   []string{"mcr-demo-01"},
			unexpectedMCRs: []string{"mcr-demo0-01", "production-mcr", "test-mcr-sydney"},
			outputFormat:   "table",
		},
		{
			name: "include inactive MCRs",
			flags: map[string]string{
				"include-inactive": "true",
			},
			setupMock: func(m *MockMCRService) {
				m.ListMCRsResult = testMCRs
			},
			expectedMCRs:   []string{"mcr-demo-01", "mcr-demo0-01", "production-mcr", "test-mcr-sydney"},
			unexpectedMCRs: []string{},
			outputFormat:   "table",
		},
		{
			name: "filter with no matches",
			flags: map[string]string{
				"name": "nonexistent-mcr",
			},
			setupMock: func(m *MockMCRService) {
				m.ListMCRsResult = testMCRs
			},
			expectedMCRs:   []string{},
			unexpectedMCRs: []string{"mcr-demo-01", "mcr-demo0-01", "production-mcr", "test-mcr-sydney"},
			outputFormat:   "table",
		},
		{
			name: "JSON output format",
			flags: map[string]string{
				"name": "mcr-demo-01",
			},
			setupMock: func(m *MockMCRService) {
				m.ListMCRsResult = testMCRs
			},
			expectedMCRs:   []string{"mcr-demo-01"},
			unexpectedMCRs: []string{"mcr-demo0-01", "production-mcr"},
			outputFormat:   "json",
		},
		{
			name: "API error",
			setupMock: func(m *MockMCRService) {
				m.ListMCRsErr = fmt.Errorf("API error: service unavailable")
			},
			expectedError: "API error: service unavailable",
			outputFormat:  "table",
		},
		{
			name: "limit results",
			flags: map[string]string{
				"limit": "2",
			},
			setupMock: func(m *MockMCRService) {
				m.ListMCRsResult = testMCRs
			},
			expectedMCRs:   []string{"mcr-demo-01", "mcr-demo0-01"},
			unexpectedMCRs: []string{"production-mcr"},
			outputFormat:   "table",
		},
		{
			name: "negative limit returns error",
			flags: map[string]string{
				"limit": "-1",
			},
			setupMock: func(m *MockMCRService) {
				m.ListMCRsResult = testMCRs
			},
			expectedError: "--limit must be a non-negative integer",
			outputFormat:  "table",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
				Use: "list",
				RunE: func(cmd *cobra.Command, args []string) error {
					return ListMCRs(cmd, args, true, tt.outputFormat) // noColor = true for testing
				},
			}

			// Add flags
			cmd.Flags().String("name", "", "Filter MCRs by name")
			cmd.Flags().Int("location-id", 0, "Filter MCRs by location ID")
			cmd.Flags().Int("port-speed", 0, "Filter MCRs by port speed")
			cmd.Flags().Bool("include-inactive", false, "Include inactive MCRs")
			cmd.Flags().Int("limit", 0, "Maximum number of results to display")

			// Set flag values from test case
			testutil.SetFlags(t, cmd, tt.flags)

			var err error
			capturedOutput := output.CaptureOutput(func() {
				err = cmd.RunE(cmd, []string{})
			})

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)

				// Check that expected MCRs are in the output
				for _, expectedMCR := range tt.expectedMCRs {
					assert.Contains(t, capturedOutput, expectedMCR,
						"Expected MCR '%s' should be in output", expectedMCR)
				}

				// Check that unexpected MCRs are NOT in the output
				for _, unexpectedMCR := range tt.unexpectedMCRs {
					assert.NotContains(t, capturedOutput, unexpectedMCR,
						"Unexpected MCR '%s' should NOT be in output", unexpectedMCR)
				}

				// Verify output format
				switch tt.outputFormat {
				case "json":
					if len(tt.expectedMCRs) > 0 {
						assert.Contains(t, capturedOutput, "\"uid\":")
						assert.Contains(t, capturedOutput, "\"name\":")
					}
				case "table":
					if len(tt.expectedMCRs) > 0 {
						assert.Contains(t, capturedOutput, "UID")
						assert.Contains(t, capturedOutput, "NAME")
					}
				}

				// Check warning message when no results found
				if len(tt.expectedMCRs) == 0 && tt.expectedError == "" {
					assert.Contains(t, capturedOutput, "No MCRs found. Create one with 'megaport mcr buy'.")
				}
			}

			// Verify that the correct request was passed to the mock
			if mockMCRService.CapturedListMCRsRequest != nil {
				if tt.flags["include-inactive"] == "true" {
					assert.True(t, mockMCRService.CapturedListMCRsRequest.IncludeInactive)
				} else {
					assert.False(t, mockMCRService.CapturedListMCRsRequest.IncludeInactive)
				}
			}
		})
	}
}

func TestFilterMCRsFunction(t *testing.T) {
	testMCRs := []*megaport.MCR{
		{
			UID:        "mcr-123",
			Name:       "mcr-demo-01",
			LocationID: 571,
			PortSpeed:  1000,
		},
		{
			UID:        "mcr-456",
			Name:       "mcr-demo0-01",
			LocationID: 558,
			PortSpeed:  1000,
		},
		{
			UID:        "mcr-789",
			Name:       "Production-MCR",
			LocationID: 64,
			PortSpeed:  2500,
		},
		{
			UID:        "mcr-abc",
			Name:       "test-mcr-sydney",
			LocationID: 571,
			PortSpeed:  5000,
		},
	}

	tests := []struct {
		name         string
		locationID   int
		portSpeed    int
		mcrName      string
		expectedUIDs []string
	}{
		{
			name:         "no filters - return all",
			expectedUIDs: []string{"mcr-123", "mcr-456", "mcr-789", "mcr-abc"},
		},
		{
			name:         "filter by exact name",
			mcrName:      "mcr-demo-01",
			expectedUIDs: []string{"mcr-123"},
		},
		{
			name:         "filter by partial name",
			mcrName:      "demo",
			expectedUIDs: []string{"mcr-123", "mcr-456"},
		},
		{
			name:         "filter by case insensitive name",
			mcrName:      "PRODUCTION",
			expectedUIDs: []string{"mcr-789"},
		},
		{
			name:         "filter by location ID",
			locationID:   571,
			expectedUIDs: []string{"mcr-123", "mcr-abc"},
		},
		{
			name:         "filter by port speed",
			portSpeed:    1000,
			expectedUIDs: []string{"mcr-123", "mcr-456"},
		},
		{
			name:         "filter by name and location",
			mcrName:      "demo",
			locationID:   571,
			expectedUIDs: []string{"mcr-123"},
		},
		{
			name:         "filter by all parameters",
			mcrName:      "mcr-demo-01",
			locationID:   571,
			portSpeed:    1000,
			expectedUIDs: []string{"mcr-123"},
		},
		{
			name:         "no matches",
			mcrName:      "nonexistent",
			expectedUIDs: []string{},
		},
		{
			name:         "empty name filter - return all",
			mcrName:      "",
			expectedUIDs: []string{"mcr-123", "mcr-456", "mcr-789", "mcr-abc"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterMCRs(testMCRs, tt.locationID, tt.portSpeed, tt.mcrName)

			// Check that we got the expected number of results
			assert.Equal(t, len(tt.expectedUIDs), len(result),
				"Expected %d MCRs, got %d", len(tt.expectedUIDs), len(result))

			// Check that all expected UIDs are present
			resultUIDs := make([]string, len(result))
			for i, mcr := range result {
				resultUIDs[i] = mcr.UID
			}

			for _, expectedUID := range tt.expectedUIDs {
				assert.Contains(t, resultUIDs, expectedUID,
					"Expected UID '%s' should be in filtered results", expectedUID)
			}
		})
	}
}

func TestFilterMCRs_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		mcrs        []*megaport.MCR
		locationID  int
		portSpeed   int
		mcrName     string
		expectedLen int
	}{
		{
			name:        "nil slice",
			mcrs:        nil,
			expectedLen: 0,
		},
		{
			name:        "empty slice",
			mcrs:        []*megaport.MCR{},
			expectedLen: 0,
		},
		{
			name: "slice with nil elements",
			mcrs: []*megaport.MCR{
				{UID: "mcr-123", Name: "test-mcr"},
				nil,
				{UID: "mcr-456", Name: "another-mcr"},
			},
			expectedLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterMCRs(tt.mcrs, tt.locationID, tt.portSpeed, tt.mcrName)
			assert.Equal(t, tt.expectedLen, len(result))
		})
	}
}

func TestListMCRResourceTagsCmd(t *testing.T) {
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {})
	defer cleanup()

	tests := []struct {
		name           string
		mcrUID         string
		setupMock      func(*MockMCRService)
		expectedError  string
		expectedOutput string
	}{
		{
			name:   "successful list",
			mcrUID: "mcr-123",
			setupMock: func(m *MockMCRService) {
				m.ListMCRResourceTagsResult = map[string]string{
					"environment": "production",
					"team":        "networking",
				}
			},
			expectedOutput: "environment",
		},
		{
			name:   "empty tags",
			mcrUID: "mcr-empty",
			setupMock: func(m *MockMCRService) {
				m.ListMCRResourceTagsResult = map[string]string{}
			},
		},
		{
			name:   "API error",
			mcrUID: "mcr-error",
			setupMock: func(m *MockMCRService) {
				m.ListMCRResourceTagsErr = fmt.Errorf("API error: not found")
			},
			expectedError: "error getting resource tags",
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
				Use: "list-tags [mcrUID]",
				RunE: func(cmd *cobra.Command, args []string) error {
					return ListMCRResourceTags(cmd, args, false, "table")
				},
			}

			var err error
			capturedOutput := output.CaptureOutput(func() {
				err = cmd.RunE(cmd, []string{tt.mcrUID})
			})

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				if tt.expectedOutput != "" {
					assert.Contains(t, capturedOutput, tt.expectedOutput)
				}
			}
		})
	}
}

func TestUpdateMCRResourceTagsCmd(t *testing.T) {
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {})
	defer cleanup()
	originalResourcePrompt := utils.UpdateResourceTagsPrompt
	defer func() {
		utils.UpdateResourceTagsPrompt = originalResourcePrompt
	}()

	tests := []struct {
		name                 string
		mcrUID               string
		interactive          bool
		promptResult         map[string]string
		promptError          error
		jsonInput            string
		setupMock            func(*MockMCRService)
		expectedError        string
		expectedOutput       string
		expectedCapturedTags map[string]string
	}{
		{
			name:        "successful update with interactive mode",
			mcrUID:      "mcr-123",
			interactive: true,
			promptResult: map[string]string{
				"environment": "production",
			},
			setupMock: func(m *MockMCRService) {
				m.ListMCRResourceTagsResult = map[string]string{"environment": "staging"}
			},
			expectedOutput:       "Resource tags updated for MCR mcr-123",
			expectedCapturedTags: map[string]string{"environment": "production"},
		},
		{
			name:      "successful update with json",
			mcrUID:    "mcr-456",
			jsonInput: `{"environment": "production", "team": "networking"}`,
			setupMock: func(m *MockMCRService) {
				m.ListMCRResourceTagsResult = map[string]string{}
			},
			expectedOutput: "Resource tags updated for MCR mcr-456",
			expectedCapturedTags: map[string]string{
				"environment": "production",
				"team":        "networking",
			},
		},
		{
			name:      "error with invalid json",
			mcrUID:    "mcr-789",
			jsonInput: `{invalid json}`,
			setupMock: func(m *MockMCRService) {
				m.ListMCRResourceTagsResult = map[string]string{}
			},
			expectedError: "error parsing JSON",
		},
		{
			name:      "error with API tag listing",
			mcrUID:    "mcr-list-error",
			jsonInput: `{"environment": "production"}`,
			setupMock: func(m *MockMCRService) {
				m.ListMCRResourceTagsErr = fmt.Errorf("API error: resource not found")
			},
			expectedError: "failed to login or list existing resource tags",
		},
		{
			name:      "error with API update",
			mcrUID:    "mcr-update-error",
			jsonInput: `{"environment": "production"}`,
			setupMock: func(m *MockMCRService) {
				m.ListMCRResourceTagsResult = map[string]string{}
				m.UpdateMCRResourceTagsErr = fmt.Errorf("API error: unauthorized")
			},
			expectedError: "failed to update resource tags",
		},
		{
			name:      "empty tags clear all existing tags",
			mcrUID:    "mcr-clear",
			jsonInput: `{}`,
			setupMock: func(m *MockMCRService) {
				m.ListMCRResourceTagsResult = map[string]string{"env": "staging"}
			},
			expectedOutput:       "Resource tags updated for MCR mcr-clear",
			expectedCapturedTags: map[string]string{},
		},
		{
			name:   "no input provided",
			mcrUID: "mcr-no-input",
			setupMock: func(m *MockMCRService) {
				m.ListMCRResourceTagsResult = map[string]string{}
			},
			expectedError: "no input provided",
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

			utils.UpdateResourceTagsPrompt = func(existingTags map[string]string, noColor bool) (map[string]string, error) {
				return tt.promptResult, tt.promptError
			}

			cmd := &cobra.Command{
				Use: "update-tags [mcrUID]",
				RunE: func(cmd *cobra.Command, args []string) error {
					return UpdateMCRResourceTags(cmd, args, false)
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

			var err error
			capturedOutput := output.CaptureOutput(func() {
				err = cmd.RunE(cmd, []string{tt.mcrUID})
			})

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				if tt.expectedOutput != "" {
					assert.Contains(t, capturedOutput, tt.expectedOutput)
				}
				if tt.expectedCapturedTags != nil {
					assert.Equal(t, tt.expectedCapturedTags, mockService.CapturedUpdateMCRResourceTagsRequest)
				}
			}
		})
	}
}

func TestUpdateMCR(t *testing.T) {
	originalLoginFunc := config.LoginFunc
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {})
	defer cleanup()
	originalGetMCRFunc := getMCRFunc
	originalUpdateMCRFunc := updateMCRFunc
	defer func() {
		getMCRFunc = originalGetMCRFunc
		updateMCRFunc = originalUpdateMCRFunc
	}()

	tests := []struct {
		name           string
		args           []string
		flags          map[string]string
		setupLogin     func()
		setupGetMCR    func()
		setupUpdateMCR func()
		expectedError  string
		expectedOutput string
	}{
		{
			name: "success with flags",
			args: []string{"mcr-123"},
			flags: map[string]string{
				"name": "Updated MCR",
			},
			setupLogin: func() {
				config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
					client := &megaport.Client{}
					client.MCRService = &MockMCRService{}
					return client, nil
				}
			},
			setupGetMCR: func() {
				getMCRFunc = func(ctx context.Context, client *megaport.Client, mcrUID string) (*megaport.MCR, error) {
					return &megaport.MCR{
						UID:                mcrUID,
						Name:               "Original MCR",
						ProvisioningStatus: "LIVE",
					}, nil
				}
			},
			setupUpdateMCR: func() {
				updateMCRFunc = func(ctx context.Context, client *megaport.Client, req *megaport.ModifyMCRRequest) (*megaport.ModifyMCRResponse, error) {
					return &megaport.ModifyMCRResponse{IsUpdated: true}, nil
				}
			},
			expectedOutput: "MCR updated mcr-123",
		},
		{
			name: "success with JSON",
			args: []string{"mcr-456"},
			flags: map[string]string{
				"json": `{"name":"JSON Updated MCR"}`,
			},
			setupLogin: func() {
				config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
					client := &megaport.Client{}
					client.MCRService = &MockMCRService{}
					return client, nil
				}
			},
			setupGetMCR: func() {
				getMCRFunc = func(ctx context.Context, client *megaport.Client, mcrUID string) (*megaport.MCR, error) {
					return &megaport.MCR{
						UID:                mcrUID,
						Name:               "JSON Updated MCR",
						ProvisioningStatus: "LIVE",
					}, nil
				}
			},
			setupUpdateMCR: func() {
				updateMCRFunc = func(ctx context.Context, client *megaport.Client, req *megaport.ModifyMCRRequest) (*megaport.ModifyMCRResponse, error) {
					return &megaport.ModifyMCRResponse{IsUpdated: true}, nil
				}
			},
			expectedOutput: "MCR updated mcr-456",
		},
		{
			name:          "missing UID",
			args:          []string{},
			expectedError: "mcr UID is required",
		},
		{
			name: "login error",
			args: []string{"mcr-123"},
			flags: map[string]string{
				"name": "Updated MCR",
			},
			setupLogin: func() {
				config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
					return nil, fmt.Errorf("authentication failed")
				}
			},
			expectedError: "authentication failed",
		},
		{
			name: "get original MCR error",
			args: []string{"mcr-123"},
			flags: map[string]string{
				"name": "Updated MCR",
			},
			setupLogin: func() {
				config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
					client := &megaport.Client{}
					client.MCRService = &MockMCRService{}
					return client, nil
				}
			},
			setupGetMCR: func() {
				getMCRFunc = func(ctx context.Context, client *megaport.Client, mcrUID string) (*megaport.MCR, error) {
					return nil, fmt.Errorf("MCR not found")
				}
			},
			expectedError: "MCR not found",
		},
		{
			name: "update API error",
			args: []string{"mcr-123"},
			flags: map[string]string{
				"name": "Updated MCR",
			},
			setupLogin: func() {
				config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
					client := &megaport.Client{}
					client.MCRService = &MockMCRService{}
					return client, nil
				}
			},
			setupGetMCR: func() {
				getMCRFunc = func(ctx context.Context, client *megaport.Client, mcrUID string) (*megaport.MCR, error) {
					return &megaport.MCR{
						UID:                mcrUID,
						Name:               "Original MCR",
						ProvisioningStatus: "LIVE",
					}, nil
				}
			},
			setupUpdateMCR: func() {
				updateMCRFunc = func(ctx context.Context, client *megaport.Client, req *megaport.ModifyMCRRequest) (*megaport.ModifyMCRResponse, error) {
					return nil, fmt.Errorf("API error: service unavailable")
				}
			},
			expectedError: "API error: service unavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset to defaults
			getMCRFunc = originalGetMCRFunc
			updateMCRFunc = originalUpdateMCRFunc
			config.LoginFunc = originalLoginFunc

			if tt.setupLogin != nil {
				tt.setupLogin()
			}
			if tt.setupGetMCR != nil {
				tt.setupGetMCR()
			}
			if tt.setupUpdateMCR != nil {
				tt.setupUpdateMCR()
			}

			cmd := &cobra.Command{
				Use:  "update",
				RunE: testutil.NoColorAdapter(UpdateMCR),
			}

			cmd.Flags().Bool("interactive", false, "")
			cmd.Flags().String("json", "", "")
			cmd.Flags().String("json-file", "", "")
			cmd.Flags().String("name", "", "")
			cmd.Flags().String("cost-centre", "", "")
			cmd.Flags().Bool("marketplace-visibility", false, "")
			cmd.Flags().Int("term", 0, "")

			testutil.SetFlags(t, cmd, tt.flags)

			var err error
			capturedOutput := output.CaptureOutput(func() {
				err = cmd.RunE(cmd, tt.args)
			})

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				if tt.expectedOutput != "" {
					assert.Contains(t, capturedOutput, tt.expectedOutput)
				}
			}
		})
	}
}

func TestCreateMCRPrefixFilterList(t *testing.T) {
	originalLoginFunc := config.LoginFunc
	originalCreateFunc := createMCRPrefixFilterListFunc
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {})
	defer cleanup()
	defer func() {
		createMCRPrefixFilterListFunc = originalCreateFunc
	}()

	tests := []struct {
		name           string
		args           []string
		flags          map[string]string
		setupLogin     func()
		setupCreate    func()
		expectedError  string
		expectedOutput string
	}{
		{
			name: "success with flags",
			args: []string{"mcr-123"},
			flags: map[string]string{
				"description":    "Test Prefix List",
				"address-family": "IPv4",
				"entries":        `[{"action":"permit","prefix":"10.0.0.0/8"}]`,
			},
			setupLogin: func() {
				config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
					client := &megaport.Client{}
					client.MCRService = &MockMCRService{}
					return client, nil
				}
			},
			setupCreate: func() {
				createMCRPrefixFilterListFunc = func(ctx context.Context, client *megaport.Client, req *megaport.CreateMCRPrefixFilterListRequest) (*megaport.CreateMCRPrefixFilterListResponse, error) {
					return &megaport.CreateMCRPrefixFilterListResponse{PrefixFilterListID: 42}, nil
				}
			},
			expectedOutput: "Prefix filter list created successfully",
		},
		{
			name: "success with JSON",
			args: []string{"mcr-456"},
			flags: map[string]string{
				"json": `{"description":"JSON Prefix List","addressFamily":"IPv4","entries":[{"action":"deny","prefix":"192.168.0.0/16"}]}`,
			},
			setupLogin: func() {
				config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
					client := &megaport.Client{}
					client.MCRService = &MockMCRService{}
					return client, nil
				}
			},
			setupCreate: func() {
				createMCRPrefixFilterListFunc = func(ctx context.Context, client *megaport.Client, req *megaport.CreateMCRPrefixFilterListRequest) (*megaport.CreateMCRPrefixFilterListResponse, error) {
					return &megaport.CreateMCRPrefixFilterListResponse{PrefixFilterListID: 99}, nil
				}
			},
			expectedOutput: "Prefix filter list created successfully",
		},
		{
			name:          "missing UID",
			args:          []string{},
			expectedError: "mcr UID is required",
		},
		{
			name: "login error",
			args: []string{"mcr-123"},
			flags: map[string]string{
				"description":    "Test Prefix List",
				"address-family": "IPv4",
				"entries":        `[{"action":"permit","prefix":"10.0.0.0/8"}]`,
			},
			setupLogin: func() {
				config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
					return nil, fmt.Errorf("authentication failed")
				}
			},
			expectedError: "authentication failed",
		},
		{
			name: "API error",
			args: []string{"mcr-123"},
			flags: map[string]string{
				"description":    "Test Prefix List",
				"address-family": "IPv4",
				"entries":        `[{"action":"permit","prefix":"10.0.0.0/8"}]`,
			},
			setupLogin: func() {
				config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
					client := &megaport.Client{}
					client.MCRService = &MockMCRService{}
					return client, nil
				}
			},
			setupCreate: func() {
				createMCRPrefixFilterListFunc = func(ctx context.Context, client *megaport.Client, req *megaport.CreateMCRPrefixFilterListRequest) (*megaport.CreateMCRPrefixFilterListResponse, error) {
					return nil, fmt.Errorf("API error: prefix filter list creation failed")
				}
			},
			expectedError: "API error: prefix filter list creation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset to defaults
			config.LoginFunc = originalLoginFunc
			createMCRPrefixFilterListFunc = originalCreateFunc

			if tt.setupLogin != nil {
				tt.setupLogin()
			}
			if tt.setupCreate != nil {
				tt.setupCreate()
			}

			cmd := &cobra.Command{
				Use:  "create-prefix-filter-list",
				RunE: testutil.NoColorAdapter(CreateMCRPrefixFilterList),
			}

			cmd.Flags().Bool("interactive", false, "")
			cmd.Flags().String("json", "", "")
			cmd.Flags().String("json-file", "", "")
			cmd.Flags().String("description", "", "")
			cmd.Flags().String("address-family", "", "")
			cmd.Flags().String("entries", "", "")

			testutil.SetFlags(t, cmd, tt.flags)

			var err error
			capturedOutput := output.CaptureOutput(func() {
				err = cmd.RunE(cmd, tt.args)
			})

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				if tt.expectedOutput != "" {
					assert.Contains(t, capturedOutput, tt.expectedOutput)
				}
			}
		})
	}
}

func TestUpdateMCRPrefixFilterList(t *testing.T) {
	originalLoginFunc := config.LoginFunc
	originalModifyFunc := modifyMCRPrefixFilterListFunc
	originalGetPrefixFunc := getMCRPrefixFilterListFunc
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {})
	defer cleanup()
	defer func() {
		modifyMCRPrefixFilterListFunc = originalModifyFunc
		getMCRPrefixFilterListFunc = originalGetPrefixFunc
	}()

	tests := []struct {
		name             string
		args             []string
		flags            map[string]string
		setupLogin       func()
		setupModify      func()
		setupGetPrefixFL func()
		expectedError    string
		expectedOutput   string
	}{
		{
			name: "success with flags",
			args: []string{"mcr-123", "456"},
			flags: map[string]string{
				"description": "Updated Prefix List",
			},
			setupLogin: func() {
				config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
					client := &megaport.Client{}
					client.MCRService = &MockMCRService{}
					return client, nil
				}
			},
			setupGetPrefixFL: func() {
				getMCRPrefixFilterListFunc = func(ctx context.Context, client *megaport.Client, mcrUID string, prefixFilterListID int) (*megaport.MCRPrefixFilterList, error) {
					return &megaport.MCRPrefixFilterList{
						ID:            456,
						Description:   "Original Prefix List",
						AddressFamily: "IPv4",
						Entries: []*megaport.MCRPrefixListEntry{
							{Action: "permit", Prefix: "10.0.0.0/8"},
						},
					}, nil
				}
			},
			setupModify: func() {
				modifyMCRPrefixFilterListFunc = func(ctx context.Context, client *megaport.Client, mcrID string, prefixFilterListID int, prefixFilterList *megaport.MCRPrefixFilterList) (*megaport.ModifyMCRPrefixFilterListResponse, error) {
					return &megaport.ModifyMCRPrefixFilterListResponse{IsUpdated: true}, nil
				}
			},
			expectedOutput: "Prefix filter list updated successfully",
		},
		{
			name: "success with JSON",
			args: []string{"mcr-789", "123"},
			flags: map[string]string{
				"json": `{"description":"JSON Updated Prefix List","entries":[{"action":"deny","prefix":"172.16.0.0/12"}]}`,
			},
			setupLogin: func() {
				config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
					client := &megaport.Client{}
					client.MCRService = &MockMCRService{}
					return client, nil
				}
			},
			setupGetPrefixFL: func() {
				getMCRPrefixFilterListFunc = func(ctx context.Context, client *megaport.Client, mcrUID string, prefixFilterListID int) (*megaport.MCRPrefixFilterList, error) {
					return &megaport.MCRPrefixFilterList{
						ID:            123,
						Description:   "Original Prefix List",
						AddressFamily: "IPv4",
						Entries: []*megaport.MCRPrefixListEntry{
							{Action: "permit", Prefix: "10.0.0.0/8"},
						},
					}, nil
				}
			},
			setupModify: func() {
				modifyMCRPrefixFilterListFunc = func(ctx context.Context, client *megaport.Client, mcrID string, prefixFilterListID int, prefixFilterList *megaport.MCRPrefixFilterList) (*megaport.ModifyMCRPrefixFilterListResponse, error) {
					return &megaport.ModifyMCRPrefixFilterListResponse{IsUpdated: true}, nil
				}
			},
			expectedOutput: "Prefix filter list updated successfully",
		},
		{
			name:          "missing args",
			args:          []string{"mcr-123"},
			expectedError: "mcr UID and prefix filter list ID are required",
		},
		{
			name:          "invalid prefix filter list ID",
			args:          []string{"mcr-123", "abc"},
			expectedError: "invalid prefix filter list ID",
		},
		{
			name: "API error",
			args: []string{"mcr-123", "456"},
			flags: map[string]string{
				"description": "Updated Prefix List",
			},
			setupLogin: func() {
				config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
					client := &megaport.Client{}
					client.MCRService = &MockMCRService{}
					return client, nil
				}
			},
			setupGetPrefixFL: func() {
				getMCRPrefixFilterListFunc = func(ctx context.Context, client *megaport.Client, mcrUID string, prefixFilterListID int) (*megaport.MCRPrefixFilterList, error) {
					return &megaport.MCRPrefixFilterList{
						ID:            456,
						Description:   "Original Prefix List",
						AddressFamily: "IPv4",
						Entries: []*megaport.MCRPrefixListEntry{
							{Action: "permit", Prefix: "10.0.0.0/8"},
						},
					}, nil
				}
			},
			setupModify: func() {
				modifyMCRPrefixFilterListFunc = func(ctx context.Context, client *megaport.Client, mcrID string, prefixFilterListID int, prefixFilterList *megaport.MCRPrefixFilterList) (*megaport.ModifyMCRPrefixFilterListResponse, error) {
					return nil, fmt.Errorf("API error: update failed")
				}
			},
			expectedError: "API error: update failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset to defaults
			config.LoginFunc = originalLoginFunc
			modifyMCRPrefixFilterListFunc = originalModifyFunc
			getMCRPrefixFilterListFunc = originalGetPrefixFunc

			if tt.setupLogin != nil {
				tt.setupLogin()
			}
			if tt.setupGetPrefixFL != nil {
				tt.setupGetPrefixFL()
			}
			if tt.setupModify != nil {
				tt.setupModify()
			}

			cmd := &cobra.Command{
				Use:  "update-prefix-filter-list",
				RunE: testutil.NoColorAdapter(UpdateMCRPrefixFilterList),
			}

			cmd.Flags().Bool("interactive", false, "")
			cmd.Flags().String("json", "", "")
			cmd.Flags().String("json-file", "", "")
			cmd.Flags().String("description", "", "")
			cmd.Flags().String("address-family", "", "")
			cmd.Flags().String("entries", "", "")

			testutil.SetFlags(t, cmd, tt.flags)

			var err error
			capturedOutput := output.CaptureOutput(func() {
				err = cmd.RunE(cmd, tt.args)
			})

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				if tt.expectedOutput != "" {
					assert.Contains(t, capturedOutput, tt.expectedOutput)
				}
			}
		})
	}
}

func TestLockMCRCmd_WithMockClient(t *testing.T) {
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {})
	defer cleanup()

	tests := []struct {
		name          string
		mcrID         string
		lockErr       error
		expectedError string
		expectedOut   string
	}{
		{
			name:        "lock MCR success",
			mcrID:       "mcr-to-lock",
			expectedOut: "MCR mcr-to-lock locked successfully",
		},
		{
			name:          "lock MCR error",
			mcrID:         "mcr-error",
			lockErr:       fmt.Errorf("error locking MCR"),
			expectedError: "error locking MCR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origFunc := lockMCRFunc
			defer func() { lockMCRFunc = origFunc }()

			lockMCRFunc = func(ctx context.Context, client *megaport.Client, mcrUID string) (*megaport.ManageProductLockResponse, error) {
				if tt.lockErr != nil {
					return nil, tt.lockErr
				}
				return &megaport.ManageProductLockResponse{}, nil
			}

			lockMCRCmd := &cobra.Command{
				Use: "lock [mcrUID]",
				RunE: func(cmd *cobra.Command, args []string) error {
					return LockMCR(cmd, args, false)
				},
			}

			var err error
			capturedOutput := output.CaptureOutput(func() {
				err = lockMCRCmd.RunE(lockMCRCmd, []string{tt.mcrID})
			})

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, capturedOutput, tt.expectedOut)
			}
		})
	}
}

func TestUnlockMCRCmd_WithMockClient(t *testing.T) {
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {})
	defer cleanup()

	tests := []struct {
		name          string
		mcrID         string
		unlockErr     error
		expectedError string
		expectedOut   string
	}{
		{
			name:        "unlock MCR success",
			mcrID:       "mcr-to-unlock",
			expectedOut: "MCR mcr-to-unlock unlocked successfully",
		},
		{
			name:          "unlock MCR error",
			mcrID:         "mcr-error",
			unlockErr:     fmt.Errorf("error unlocking MCR"),
			expectedError: "error unlocking MCR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origFunc := unlockMCRFunc
			defer func() { unlockMCRFunc = origFunc }()

			unlockMCRFunc = func(ctx context.Context, client *megaport.Client, mcrUID string) (*megaport.ManageProductLockResponse, error) {
				if tt.unlockErr != nil {
					return nil, tt.unlockErr
				}
				return &megaport.ManageProductLockResponse{}, nil
			}

			unlockMCRCmd := &cobra.Command{
				Use: "unlock [mcrUID]",
				RunE: func(cmd *cobra.Command, args []string) error {
					return UnlockMCR(cmd, args, false)
				},
			}

			var err error
			capturedOutput := output.CaptureOutput(func() {
				err = unlockMCRCmd.RunE(unlockMCRCmd, []string{tt.mcrID})
			})

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, capturedOutput, tt.expectedOut)
			}
		})
	}
}

func TestLockMCRCmd_LoginError(t *testing.T) {
	config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
		return nil, fmt.Errorf("login failed")
	}
	defer func() { config.LoginFunc = nil }()

	cmd := &cobra.Command{}
	err := LockMCR(cmd, []string{"mcr-123"}, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error logging in")
}

func TestUnlockMCRCmd_LoginError(t *testing.T) {
	config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
		return nil, fmt.Errorf("login failed")
	}
	defer func() { config.LoginFunc = nil }()

	cmd := &cobra.Command{}
	err := UnlockMCR(cmd, []string{"mcr-123"}, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error logging in")
}

func TestBuyMCR_Confirmation(t *testing.T) {
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {})
	defer cleanup()

	originalBuyMCRFunc := buyMCRFunc
	defer func() { buyMCRFunc = originalBuyMCRFunc }()

	originalBuyConfirmPrompt := utils.BuyConfirmPrompt
	defer func() { utils.BuyConfirmPrompt = originalBuyConfirmPrompt }()

	tests := []struct {
		name                 string
		flags                map[string]string
		jsonInput            string
		confirmResult        bool
		expectBuyCalled      bool
		expectedOutput       string
		expectedError        string
		promptShouldBeCalled bool
	}{
		{
			name: "confirmation accepted",
			flags: map[string]string{
				"name":        "Test MCR",
				"term":        "12",
				"port-speed":  "10000",
				"location-id": "123",
				"mcr-asn":     "65000",
			},
			confirmResult:        true,
			expectBuyCalled:      true,
			expectedOutput:       "MCR created",
			promptShouldBeCalled: true,
		},
		{
			name: "confirmation denied",
			flags: map[string]string{
				"name":        "Test MCR",
				"term":        "12",
				"port-speed":  "10000",
				"location-id": "123",
				"mcr-asn":     "65000",
			},
			confirmResult:        false,
			expectBuyCalled:      false,
			expectedError:        "cancelled by user",
			promptShouldBeCalled: true,
		},
		{
			name: "yes flag skips confirmation",
			flags: map[string]string{
				"name":        "Test MCR",
				"term":        "12",
				"port-speed":  "10000",
				"location-id": "123",
				"mcr-asn":     "65000",
				"yes":         "true",
			},
			confirmResult:        false,
			expectBuyCalled:      true,
			expectedOutput:       "MCR created",
			promptShouldBeCalled: false,
		},
		{
			name: "json input skips confirmation",
			flags: map[string]string{
				"json": `{"name":"JSON MCR","term":12,"portSpeed":10000,"locationId":123,"mcrAsn":65000}`,
			},
			confirmResult:        false,
			expectBuyCalled:      true,
			expectedOutput:       "MCR created",
			promptShouldBeCalled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMCRService := &MockMCRService{
				BuyMCRResult: &megaport.BuyMCRResponse{
					TechnicalServiceUID: "mcr-confirm-123",
				},
			}

			config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
				client := &megaport.Client{}
				client.MCRService = mockMCRService
				return client, nil
			}

			buyCalled := false
			buyMCRFunc = func(ctx context.Context, client *megaport.Client, req *megaport.BuyMCRRequest) (*megaport.BuyMCRResponse, error) {
				buyCalled = true
				return &megaport.BuyMCRResponse{
					TechnicalServiceUID: "mcr-confirm-123",
				}, nil
			}

			promptCalled := false
			utils.BuyConfirmPrompt = func(_ string, _ []utils.BuyConfirmDetail, _ bool) bool {
				promptCalled = true
				return tt.confirmResult
			}

			cmd := &cobra.Command{
				Use:  "buy",
				RunE: testutil.NoColorAdapter(BuyMCR),
			}

			cmd.Flags().BoolP("interactive", "i", false, "")
			cmd.Flags().BoolP("yes", "y", false, "")
			cmd.Flags().Bool("no-wait", false, "")
			cmd.Flags().String("name", "", "")
			cmd.Flags().Int("term", 0, "")
			cmd.Flags().Int("port-speed", 0, "")
			cmd.Flags().Int("location-id", 0, "")
			cmd.Flags().Int("mcr-asn", 0, "")
			cmd.Flags().String("diversity-zone", "", "")
			cmd.Flags().String("cost-centre", "", "")
			cmd.Flags().String("promo-code", "", "")
			cmd.Flags().String("json", "", "")
			cmd.Flags().String("json-file", "", "")

			testutil.SetFlags(t, cmd, tt.flags)

			var err error
			capturedOutput := output.CaptureOutput(func() {
				err = cmd.RunE(cmd, nil)
			})

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, capturedOutput, tt.expectedOutput)
			}

			assert.Equal(t, tt.expectBuyCalled, buyCalled, "buy function called mismatch")
			assert.Equal(t, tt.promptShouldBeCalled, promptCalled, "confirmation prompt called mismatch")
		})
	}
}

func TestExportMCRConfig(t *testing.T) {
	mcr := &megaport.MCR{
		UID:                "mcr-should-not-appear",
		Name:               "My MCR",
		ContractTermMonths: 12,
		PortSpeed:          1000,
		LocationID:         99,
		DiversityZone:      "red",
		CostCentre:         "NetOps",
		ProvisioningStatus: "LIVE",
		Resources: megaport.MCRResources{
			VirtualRouter: megaport.MCRVirtualRouter{
				ASN: 65000,
			},
		},
	}
	m := exportMCRConfig(mcr)

	assert.Equal(t, "My MCR", m["name"])
	assert.Equal(t, 12, m["term"])
	assert.Equal(t, 1000, m["portSpeed"])
	assert.Equal(t, 99, m["locationId"])
	assert.Equal(t, "red", m["diversityZone"])
	assert.Equal(t, "NetOps", m["costCentre"])
	assert.Equal(t, 65000, m["mcrAsn"])

	_, hasUID := m["productUid"]
	assert.False(t, hasUID, "export should not include productUid")
	_, hasStatus := m["provisioningStatus"]
	assert.False(t, hasStatus, "export should not include provisioningStatus")
}

func TestGetMCR_Export(t *testing.T) {
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {})
	defer cleanup()

	mockService := &MockMCRService{
		GetMCRResult: &megaport.MCR{
			UID:                "mcr-export-123",
			Name:               "Export MCR",
			ContractTermMonths: 12,
			PortSpeed:          1000,
			LocationID:         42,
			ProvisioningStatus: "LIVE",
		},
	}
	config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
		client := &megaport.Client{}
		client.MCRService = mockService
		return client, nil
	}

	cmd := &cobra.Command{Use: "get"}
	cmd.Flags().Bool("export", false, "")
	assert.NoError(t, cmd.Flags().Set("export", "true"))

	var err error
	capturedOutput := output.CaptureOutput(func() {
		err = GetMCR(cmd, []string{"mcr-export-123"}, true, "table")
	})

	assert.NoError(t, err)
	var parsed map[string]interface{}
	assert.NoError(t, json.Unmarshal([]byte(capturedOutput), &parsed), "export output must be valid JSON")
	assert.Equal(t, "Export MCR", parsed["name"])
	assert.Equal(t, float64(42), parsed["locationId"])
	_, hasUID := parsed["productUid"]
	assert.False(t, hasUID, "export should not include productUid")
}

func TestValidateMCR(t *testing.T) {
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {})
	defer cleanup()

	tests := []struct {
		name             string
		flags            map[string]string
		jsonInput        string
		jsonFileContent  string
		setupMock        func(*MockMCRService)
		loginError       error
		expectedError    string
		expectedContains string
	}{
		{
			name: "success with flags",
			flags: map[string]string{
				"name":                   "test-mcr",
				"term":                   "12",
				"port-speed":             "5000",
				"location-id":            "1",
				"marketplace-visibility": "true",
			},
			setupMock:        func(m *MockMCRService) {},
			expectedContains: "validation passed",
		},
		{
			name:             "success with JSON",
			jsonInput:        `{"name":"json-mcr","term":12,"portSpeed":5000,"locationId":1,"marketplaceVisibility":true}`,
			setupMock:        func(m *MockMCRService) {},
			expectedContains: "validation passed",
		},
		{
			name: "validation error",
			flags: map[string]string{
				"name":                   "test-mcr",
				"term":                   "12",
				"port-speed":             "5000",
				"location-id":            "1",
				"marketplace-visibility": "true",
			},
			setupMock: func(m *MockMCRService) {
				m.ValidateMCROrderErr = fmt.Errorf("invalid MCR configuration")
			},
			expectedError: "invalid MCR configuration",
		},
		{
			name:          "no input provided",
			flags:         map[string]string{},
			setupMock:     func(m *MockMCRService) {},
			expectedError: "no input provided",
		},
		{
			name: "login error",
			flags: map[string]string{
				"name":                   "test-mcr",
				"term":                   "12",
				"port-speed":             "5000",
				"location-id":            "1",
				"marketplace-visibility": "true",
			},
			setupMock:     func(m *MockMCRService) {},
			loginError:    fmt.Errorf("authentication failed"),
			expectedError: "authentication failed",
		},
		{
			name:          "invalid JSON input",
			jsonInput:     `{invalid json}`,
			setupMock:     func(m *MockMCRService) {},
			expectedError: "error parsing JSON",
		},
		{
			name:             "success with JSON file",
			jsonFileContent:  `{"name":"file-mcr","term":12,"portSpeed":5000,"locationId":1,"marketplaceVisibility":true}`,
			setupMock:        func(m *MockMCRService) {},
			expectedContains: "validation passed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockMCRService{}
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			if tt.loginError != nil {
				config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
					return nil, tt.loginError
				}
			} else {
				config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
					client := &megaport.Client{}
					client.MCRService = mockService
					return client, nil
				}
			}

			cmd := &cobra.Command{Use: "validate"}
			cmd.Flags().Bool("interactive", false, "")
			cmd.Flags().String("json", "", "")
			cmd.Flags().String("json-file", "", "")
			cmd.Flags().String("name", "", "")
			cmd.Flags().Int("term", 0, "")
			cmd.Flags().Int("port-speed", 0, "")
			cmd.Flags().Int("location-id", 0, "")
			cmd.Flags().Bool("marketplace-visibility", false, "")
			cmd.Flags().Int("mcr-asn", 0, "")
			cmd.Flags().String("diversity-zone", "", "")
			cmd.Flags().String("cost-centre", "", "")
			cmd.Flags().String("promo-code", "", "")
			cmd.Flags().String("resource-tags", "", "")

			if tt.jsonInput != "" {
				assert.NoError(t, cmd.Flags().Set("json", tt.jsonInput))
			}
			if tt.jsonFileContent != "" {
				tmpFile, tmpErr := os.CreateTemp("", "mcr-validate-*.json")
				assert.NoError(t, tmpErr)
				defer os.Remove(tmpFile.Name())
				_, tmpErr = tmpFile.WriteString(tt.jsonFileContent)
				assert.NoError(t, tmpErr)
				tmpFile.Close()
				assert.NoError(t, cmd.Flags().Set("json-file", tmpFile.Name()))
			}
			for k, v := range tt.flags {
				assert.NoError(t, cmd.Flags().Set(k, v))
			}

			var err error
			capturedOutput := output.CaptureOutput(func() {
				err = ValidateMCR(cmd, nil, true)
			})

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				if tt.expectedContains != "" {
					assert.Contains(t, capturedOutput, tt.expectedContains)
				}
			}
		})
	}
}
