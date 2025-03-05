package cmd

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
)

var testMCRs = []*megaport.MCR{
	{
		UID:                "mcr-1",
		Name:               "MyMCROne",
		LocationID:         1,
		ProvisioningStatus: "ACTIVE",
	},
	{
		UID:                "mcr-2",
		Name:               "AnotherMCR",
		LocationID:         2,
		ProvisioningStatus: "INACTIVE",
	},
}

func TestPrintMCRs_Table(t *testing.T) {
	output := captureOutput(func() {
		err := printMCRs(testMCRs, "table")
		assert.NoError(t, err)
	})

	expected := `uid     name         location_id   provisioning_status
mcr-1   MyMCROne     1             ACTIVE
mcr-2   AnotherMCR   2             INACTIVE
`
	assert.Equal(t, expected, output)
}

func TestPrintMCRs_JSON(t *testing.T) {
	output := captureOutput(func() {
		err := printMCRs(testMCRs, "json")
		assert.NoError(t, err)
	})

	expected := `[
  {
    "uid": "mcr-1",
    "name": "MyMCROne",
    "location_id": 1,
    "provisioning_status": "ACTIVE"
  },
  {
    "uid": "mcr-2",
    "name": "AnotherMCR",
    "location_id": 2,
    "provisioning_status": "INACTIVE"
  }
]`
	assert.JSONEq(t, expected, output)
}

func TestPrintMCRs_CSV(t *testing.T) {
	output := captureOutput(func() {
		err := printMCRs(testMCRs, "csv")
		assert.NoError(t, err)
	})

	expected := `uid,name,location_id,provisioning_status
mcr-1,MyMCROne,1,ACTIVE
mcr-2,AnotherMCR,2,INACTIVE
`
	assert.Equal(t, expected, output)
}

func TestPrintMCRs_Invalid(t *testing.T) {
	var err error
	output := captureOutput(func() {
		err = printMCRs(testMCRs, "invalid")
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid output format")
	assert.Empty(t, output)
}

func TestPrintMCRs_EmptyAndNilSlice(t *testing.T) {
	tests := []struct {
		name     string
		mcrs     []*megaport.MCR
		format   string
		expected string
	}{
		{
			name:   "empty slice table format",
			mcrs:   []*megaport.MCR{},
			format: "table",
			expected: `uid   name   location_id   provisioning_status
`,
		},
		{
			name:   "empty slice csv format",
			mcrs:   []*megaport.MCR{},
			format: "csv",
			expected: `uid,name,location_id,provisioning_status
`,
		},
		{
			name:     "empty slice json format",
			mcrs:     []*megaport.MCR{},
			format:   "json",
			expected: "[]\n",
		},
		{
			name:   "nil slice table format",
			mcrs:   nil,
			format: "table",
			expected: `uid   name   location_id   provisioning_status
`,
		},
		{
			name:   "nil slice csv format",
			mcrs:   nil,
			format: "csv",
			expected: `uid,name,location_id,provisioning_status
`,
		},
		{
			name:     "nil slice json format",
			mcrs:     nil,
			format:   "json",
			expected: "[]\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureOutput(func() {
				err := printMCRs(tt.mcrs, tt.format)
				assert.NoError(t, err)
			})
			assert.Equal(t, tt.expected, output)
		})
	}
}

func TestBuyMCRCmd_WithMockClient(t *testing.T) {
	originalPrompt := prompt
	originalLoginFunc := loginFunc
	defer func() {
		prompt = originalPrompt
		loginFunc = originalLoginFunc
	}()

	tests := []struct {
		name           string
		prompts        []string
		setupMock      func(*MockMCRService)
		expectedError  string
		expectedOutput string
	}{
		{
			name: "successful MCR purchase",
			prompts: []string{
				"Test MCR",  // name
				"12",        // term
				"1000",      // port speed
				"123",       // location ID
				"red",       // diversity zone
				"cost-123",  // cost center
				"PROMO2025", // promo code
			},
			setupMock: func(m *MockMCRService) {
				m.BuyMCRResult = &megaport.BuyMCRResponse{
					TechnicalServiceUID: "mcr-123-abc",
				}
			},
			expectedOutput: "MCR purchased successfully - UID: mcr-123-abc",
		},
		{
			name: "validation error",
			prompts: []string{
				"Test MCR",  // name
				"12",        // term
				"1000",      // port speed
				"123",       // location ID
				"red",       // diversity zone
				"cost-123",  // cost center
				"PROMO2025", // promo code
			},
			setupMock: func(m *MockMCRService) {
				m.ValidateMCROrderErr = fmt.Errorf("validation failed: invalid location")
			},
			expectedError: "validation failed: invalid location",
		},
		{
			name: "API error",
			prompts: []string{
				"Test MCR",  // name
				"12",        // term
				"1000",      // port speed
				"123",       // location ID
				"red",       // diversity zone
				"cost-123",  // cost center
				"PROMO2025", // promo code
			},
			setupMock: func(m *MockMCRService) {
				m.BuyMCRErr = fmt.Errorf("API error: service unavailable")
			},
			expectedError: "API error: service unavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			promptIndex := 0
			prompt = func(msg string) (string, error) {
				if promptIndex < len(tt.prompts) {
					response := tt.prompts[promptIndex]
					promptIndex++
					return response, nil
				}
				return "", fmt.Errorf("unexpected prompt call")
			}

			mockMCRService := &MockMCRService{}
			tt.setupMock(mockMCRService)

			loginFunc = func(ctx context.Context) (*megaport.Client, error) {
				client := &megaport.Client{}
				client.MCRService = mockMCRService
				return client, nil
			}

			cmd := buyMCRCmd
			var err error
			output := captureOutput(func() {
				err = cmd.RunE(cmd, []string{})
			})

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, output, tt.expectedOutput)

				assert.NotNil(t, mockMCRService.CapturedBuyMCRRequest)
				req := mockMCRService.CapturedBuyMCRRequest
				assert.Equal(t, "Test MCR", req.Name)
				assert.Equal(t, 12, req.Term)
				assert.Equal(t, 1000, req.PortSpeed)
				assert.Equal(t, 123, req.LocationID)
				assert.Equal(t, "red", req.DiversityZone)
				assert.Equal(t, "cost-123", req.CostCentre)
				assert.Equal(t, "PROMO2025", req.PromoCode)
			}
		})
	}
}

func TestGetMCRCmd_WithMockClient(t *testing.T) {
	originalLoginFunc := loginFunc
	originalOutputFormat := outputFormat
	defer func() {
		loginFunc = originalLoginFunc
		outputFormat = originalOutputFormat
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

			loginFunc = func(ctx context.Context) (*megaport.Client, error) {
				client := &megaport.Client{}
				client.MCRService = mockMCRService
				return client, nil
			}

			outputFormat = tt.format

			var err error
			output := captureOutput(func() {
				err = getMCRCmd.RunE(getMCRCmd, []string{tt.mcrID})
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
	originalLoginFunc := loginFunc
	originalPrompt := prompt
	defer func() {
		loginFunc = originalLoginFunc
		prompt = originalPrompt
	}()

	tests := []struct {
		name           string
		mcrID          string
		force          bool
		deleteNow      bool
		promptResponse string
		setupMock      func(*MockMCRService)
		expectedError  string
		expectedOut    string
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
			expectedOut:   "MCR mcr-to-delete deleted successfully",
			expectDeleted: true,
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
			expectedOut:   "MCR mcr-to-delete-now deleted successfully",
			expectDeleted: true,
		},
		{
			name:           "cancel deletion",
			mcrID:          "mcr-keep",
			force:          false,
			promptResponse: "n",
			setupMock:      func(m *MockMCRService) {},
			expectedOut:    "Deletion cancelled",
			expectDeleted:  false,
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
			expectedOut:   "MCR mcr-force-delete deleted successfully",
			expectDeleted: true,
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
			loginFunc = func(ctx context.Context) (*megaport.Client, error) {
				client := &megaport.Client{}
				client.MCRService = mockMCRService
				return client, nil
			}

			// Setup prompt mock
			prompt = func(msg string) (string, error) {
				return tt.promptResponse, nil
			}

			// Set flags
			cmd := deleteMCRCmd
			if err := cmd.Flags().Set("force", fmt.Sprintf("%v", tt.force)); err != nil {
				t.Fatalf("Failed to set force flag: %v", err)
			}
			if err := cmd.Flags().Set("now", fmt.Sprintf("%v", tt.deleteNow)); err != nil {
				t.Fatalf("Failed to set now flag: %v", err)
			}

			// Execute command and capture output
			var err error
			output := captureOutput(func() {
				err = cmd.RunE(cmd, []string{tt.mcrID})
			})

			// Check results
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, output, tt.expectedOut)

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
	originalLoginFunc := loginFunc
	defer func() {
		loginFunc = originalLoginFunc
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
			loginFunc = func(ctx context.Context) (*megaport.Client, error) {
				client := &megaport.Client{}
				client.MCRService = mockMCRService
				return client, nil
			}

			// Execute command and capture output
			var err error
			output := captureOutput(func() {
				err = restoreMCRCmd.RunE(restoreMCRCmd, []string{tt.mcrID})
			})

			// Check results
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, output, tt.expectedOut)

				if strings.Contains(tt.expectedOut, "restored successfully") {
					assert.Equal(t, tt.mcrID, mockMCRService.CapturedRestoreMCRUID)
				}
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

func TestCreateMCRPrefixFilterListCmd_WithMockClient(t *testing.T) {
	originalPrompt := prompt
	originalLoginFunc := loginFunc
	defer func() {
		prompt = originalPrompt
		loginFunc = originalLoginFunc
	}()

	tests := []struct {
		name           string
		mcrUID         string
		prompts        []string
		setupMock      func(*MockMCRService)
		expectedError  string
		expectedOutput string
	}{
		{
			name:   "successful prefix filter list creation",
			mcrUID: "mcr-123",
			prompts: []string{
				"Test Prefix Filter List", // description
				"IPv4",                    // address family
				"permit",                  // action
				"192.168.0.0/16",          // prefix
				"",                        // ge
				"",                        // le
				"",                        // end of entries
			},
			setupMock: func(m *MockMCRService) {
				m.CreateMCRPrefixFilterListResult = &megaport.CreateMCRPrefixFilterListResponse{
					PrefixFilterListID: 456,
				}
			},
			expectedOutput: "Prefix filter list created successfully - ID: 456",
		},
		{
			name:   "validation error",
			mcrUID: "mcr-123",
			prompts: []string{
				"Test Prefix Filter List", // description
				"IPv4",                    // address family
				"permit",                  // action
				"192.168.0.0/16",          // prefix
				"",                        // ge
				"",                        // le
				"",                        // end of entries
			},
			setupMock: func(m *MockMCRService) {
				m.CreateMCRPrefixFilterListErr = fmt.Errorf("validation failed: invalid prefix")
			},
			expectedError: "validation failed: invalid prefix",
		},
		{
			name:   "API error",
			mcrUID: "mcr-123",
			prompts: []string{
				"Test Prefix Filter List", // description
				"IPv4",                    // address family
				"permit",                  // action
				"192.168.0.0/16",          // prefix
				"",                        // ge
				"",                        // le
				"",                        // end of entries
			},
			setupMock: func(m *MockMCRService) {
				m.CreateMCRPrefixFilterListErr = fmt.Errorf("API error: service unavailable")
			},
			expectedError: "API error: service unavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			promptIndex := 0
			prompt = func(msg string) (string, error) {
				if promptIndex < len(tt.prompts) {
					response := tt.prompts[promptIndex]
					promptIndex++
					return response, nil
				}
				return "", fmt.Errorf("unexpected prompt call")
			}

			mockMCRService := &MockMCRService{}
			tt.setupMock(mockMCRService)

			loginFunc = func(ctx context.Context) (*megaport.Client, error) {
				client := &megaport.Client{}
				client.MCRService = mockMCRService
				return client, nil
			}

			cmd := createMCRPrefixFilterListCmd
			var err error
			output := captureOutput(func() {
				err = cmd.RunE(cmd, []string{tt.mcrUID})
			})

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, output, tt.expectedOutput)

				assert.NotNil(t, mockMCRService.CapturedCreateMCRPrefixFilterListRequest)
				req := mockMCRService.CapturedCreateMCRPrefixFilterListRequest
				assert.Equal(t, tt.mcrUID, req.MCRID)
				assert.Equal(t, "Test Prefix Filter List", req.PrefixFilterList.Description)
				assert.Equal(t, "IPv4", req.PrefixFilterList.AddressFamily)
				assert.Len(t, req.PrefixFilterList.Entries, 1)
				assert.Equal(t, "permit", req.PrefixFilterList.Entries[0].Action)
				assert.Equal(t, "192.168.0.0/16", req.PrefixFilterList.Entries[0].Prefix)
			}
		})
	}
}

func TestListMCRPrefixFilterListsCmd_WithMockClient(t *testing.T) {
	originalLoginFunc := loginFunc
	defer func() {
		loginFunc = originalLoginFunc
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

			loginFunc = func(ctx context.Context) (*megaport.Client, error) {
				client := &megaport.Client{}
				client.MCRService = mockMCRService
				return client, nil
			}

			cmd := listMCRPrefixFilterListsCmd
			var err error
			output := captureOutput(func() {
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
	originalLoginFunc := loginFunc
	defer func() {
		loginFunc = originalLoginFunc
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

			loginFunc = func(ctx context.Context) (*megaport.Client, error) {
				client := &megaport.Client{}
				client.MCRService = mockMCRService
				return client, nil
			}

			cmd := getMCRPrefixFilterListCmd
			var err error
			output := captureOutput(func() {
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

func TestUpdateMCRPrefixFilterListCmd_WithMockClient(t *testing.T) {
	originalPrompt := prompt
	originalLoginFunc := loginFunc
	defer func() {
		prompt = originalPrompt
		loginFunc = originalLoginFunc
	}()

	tests := []struct {
		name           string
		mcrUID         string
		prefixListID   int
		prompts        []string
		setupMock      func(*MockMCRService)
		expectedError  string
		expectedOutput string
	}{
		{
			name:         "successful update prefix filter list",
			mcrUID:       "mcr-123",
			prefixListID: 1,
			prompts: []string{
				"Updated Prefix Filter List", // description
				"IPv4",                       // address family
				"permit",                     // action
				"192.168.0.0/16",             // prefix
				"",                           // ge
				"",                           // le
				"",                           // end of entries
			},
			setupMock: func(m *MockMCRService) {
				m.ModifyMCRPrefixFilterListResult = &megaport.ModifyMCRPrefixFilterListResponse{
					IsUpdated: true,
				}
			},
			expectedOutput: "Prefix filter list updated successfully - ID: 1",
		},
		{
			name:         "API error",
			mcrUID:       "mcr-123",
			prefixListID: 1,
			prompts: []string{
				"Updated Prefix Filter List", // description
				"IPv4",                       // address family
				"permit",                     // action
				"192.168.0.0/16",             // prefix
				"",                           // ge
				"",                           // le
				"",                           // end of entries
			},
			setupMock: func(m *MockMCRService) {
				m.ModifyMCRPrefixFilterListErr = fmt.Errorf("API error: service unavailable")
			},
			expectedError: "API error: service unavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			promptIndex := 0
			prompt = func(msg string) (string, error) {
				if promptIndex < len(tt.prompts) {
					response := tt.prompts[promptIndex]
					promptIndex++
					return response, nil
				}
				return "", fmt.Errorf("unexpected prompt call")
			}

			mockMCRService := &MockMCRService{}
			tt.setupMock(mockMCRService)

			loginFunc = func(ctx context.Context) (*megaport.Client, error) {
				client := &megaport.Client{}
				client.MCRService = mockMCRService
				return client, nil
			}

			cmd := updateMCRPrefixFilterListCmd
			var err error
			output := captureOutput(func() {
				err = cmd.RunE(cmd, []string{tt.mcrUID, fmt.Sprintf("%d", tt.prefixListID)})
			})

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, output, tt.expectedOutput)

				assert.NotNil(t, mockMCRService.CapturedModifyMCRPrefixFilterListRequest)
				req := mockMCRService.CapturedModifyMCRPrefixFilterListRequest
				assert.Equal(t, "Updated Prefix Filter List", req.Description)
				assert.Equal(t, "IPv4", req.AddressFamily)
				assert.Len(t, req.Entries, 1)
				assert.Equal(t, "permit", req.Entries[0].Action)
				assert.Equal(t, "192.168.0.0/16", req.Entries[0].Prefix)
			}
		})
	}
}

func TestDeleteMCRPrefixFilterListCmd_WithMockClient(t *testing.T) {
	originalLoginFunc := loginFunc
	defer func() {
		loginFunc = originalLoginFunc
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
			name:         "successful delete prefix filter list",
			mcrUID:       "mcr-123",
			prefixListID: 1,
			setupMock: func(m *MockMCRService) {
				m.DeleteMCRPrefixFilterListResult = &megaport.DeleteMCRPrefixFilterListResponse{
					IsDeleted: true,
				}
			},
			expectedOutput: "Prefix filter list deleted successfully - ID: 1",
		},
		{
			name:         "API error",
			mcrUID:       "mcr-123",
			prefixListID: 1,
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

			loginFunc = func(ctx context.Context) (*megaport.Client, error) {
				client := &megaport.Client{}
				client.MCRService = mockMCRService
				return client, nil
			}

			cmd := deleteMCRPrefixFilterListCmd
			var err error
			output := captureOutput(func() {
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

func TestUpdateMCRCmd_WithMockClient(t *testing.T) {
	originalPrompt := prompt
	originalLoginFunc := loginFunc
	defer func() {
		prompt = originalPrompt
		loginFunc = originalLoginFunc
	}()

	tests := []struct {
		name           string
		mcrUID         string
		prompts        []string
		setupMock      func(*MockMCRService)
		expectedError  string
		expectedOutput string
	}{
		{
			name:   "successful MCR update",
			mcrUID: "mcr-123",
			prompts: []string{
				"Updated MCR", // name
				"new-cost",    // cost centre
				"true",        // marketplace visibility
				"24",          // contract term months
			},
			setupMock: func(m *MockMCRService) {
				m.ModifyMCRResult = &megaport.ModifyMCRResponse{
					IsUpdated: true,
				}
			},
			expectedOutput: "MCR updated successfully",
		},
		{
			name:   "API error",
			mcrUID: "mcr-123",
			prompts: []string{
				"Updated MCR", // name
				"new-cost",    // cost centre
				"true",        // marketplace visibility
				"24",          // contract term months
			},
			setupMock: func(m *MockMCRService) {
				m.ModifyMCRErr = fmt.Errorf("API error: service unavailable")
			},
			expectedError: "API error: service unavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			promptIndex := 0
			prompt = func(msg string) (string, error) {
				if promptIndex < len(tt.prompts) {
					response := tt.prompts[promptIndex]
					promptIndex++
					return response, nil
				}
				return "", fmt.Errorf("unexpected prompt call")
			}

			mockMCRService := &MockMCRService{}
			tt.setupMock(mockMCRService)

			loginFunc = func(ctx context.Context) (*megaport.Client, error) {
				client := &megaport.Client{}
				client.MCRService = mockMCRService
				return client, nil
			}

			cmd := updateMCRCmd
			var err error
			output := captureOutput(func() {
				err = cmd.RunE(cmd, []string{tt.mcrUID})
			})

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, output, tt.expectedOutput)

				assert.NotNil(t, mockMCRService.CapturedModifyMCRRequest)
				req := mockMCRService.CapturedModifyMCRRequest
				assert.Equal(t, tt.mcrUID, req.MCRID)
				assert.Equal(t, "Updated MCR", req.Name)
				assert.Equal(t, "new-cost", req.CostCentre)
				assert.NotNil(t, req.MarketplaceVisibility)
				assert.True(t, *req.MarketplaceVisibility)
				assert.NotNil(t, req.ContractTermMonths)
				assert.Equal(t, 24, *req.ContractTermMonths)
			}
		})
	}
}
