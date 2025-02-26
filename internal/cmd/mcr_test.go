package cmd

import (
	"context"
	"fmt"
	"strings"
	"testing"

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

func TestPrintMCRs_EmptySlice(t *testing.T) {
	var emptyMCRs []*megaport.MCR

	tests := []struct {
		name     string
		format   string
		expected string
	}{
		{
			name:   "table format",
			format: "table",
			expected: `uid   name   location_id   provisioning_status
`,
		},
		{
			name:   "csv format",
			format: "csv",
			expected: `uid,name,location_id,provisioning_status
`,
		},
		{
			name:     "json format",
			format:   "json",
			expected: "[]\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureOutput(func() {
				err := printMCRs(emptyMCRs, tt.format)
				assert.NoError(t, err)
			})
			assert.Equal(t, tt.expected, output)
		})
	}
}

func TestPrintMCRs_NilSlice(t *testing.T) {
	var nilMCRs []*megaport.MCR = nil

	tests := []struct {
		name     string
		format   string
		expected string
	}{
		{
			name:   "table format",
			format: "table",
			expected: `uid   name   location_id   provisioning_status
`,
		},
		{
			name:   "csv format",
			format: "csv",
			expected: `uid,name,location_id,provisioning_status
`,
		},
		{
			name:     "json format",
			format:   "json",
			expected: "[]\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureOutput(func() {
				err := printMCRs(nilMCRs, tt.format)
				assert.NoError(t, err)
			})
			assert.Equal(t, tt.expected, output)
		})
	}
}

func TestPrintMCRs_InvalidMCR(t *testing.T) {
	invalidMCRs := []*megaport.MCR{
		{
			UID:                "",
			Name:               "",
			LocationID:         0,
			ProvisioningStatus: "",
		},
	}

	tests := []struct {
		name        string
		format      string
		mcrs        []*megaport.MCR
		shouldError bool
		expected    string
	}{
		{
			name:        "table format with zero values",
			format:      "table",
			mcrs:        invalidMCRs,
			shouldError: false,
			expected:    "   0             ",
		},
		{
			name:        "csv format with zero values",
			format:      "csv",
			mcrs:        invalidMCRs,
			shouldError: false,
			expected:    ",,0,",
		},
		{
			name:        "json format with zero values",
			format:      "json",
			mcrs:        invalidMCRs,
			shouldError: false,
			expected:    `[{"uid":"","name":"","location_id":0,"provisioning_status":""}]`,
		},
		{
			name:        "table format with nil MCR",
			format:      "table",
			mcrs:        []*megaport.MCR{nil},
			shouldError: true,
			expected:    "invalid MCR: nil value",
		},
		{
			name:        "csv format with nil MCR",
			format:      "csv",
			mcrs:        []*megaport.MCR{nil},
			shouldError: true,
			expected:    "invalid MCR: nil value",
		},
		{
			name:        "json format with nil MCR",
			format:      "json",
			mcrs:        []*megaport.MCR{nil},
			shouldError: true,
			expected:    "invalid MCR: nil value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output string
			var err error

			output = captureOutput(func() {
				err = printMCRs(tt.mcrs, tt.format)
			})

			if tt.shouldError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expected)
				assert.Empty(t, output)
			} else {
				assert.NoError(t, err)
				if tt.format == "json" {
					assert.JSONEq(t, tt.expected, output)
				} else {
					assert.Contains(t, output, tt.expected)
				}
			}
		})
	}
}

func TestBuyMCRCmd_WithMockClient(t *testing.T) {
	// Save original functions and restore after test
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
			// Setup mock prompt
			promptIndex := 0
			prompt = func(msg string) (string, error) {
				if promptIndex < len(tt.prompts) {
					response := tt.prompts[promptIndex]
					promptIndex++
					return response, nil
				}
				return "", fmt.Errorf("unexpected prompt call")
			}

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
			cmd := buyMCRCmd
			var err error
			output := captureOutput(func() {
				err = cmd.RunE(cmd, []string{})
			})

			// Check results
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, output, tt.expectedOutput)

				// Verify request details
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
	// Save original login function and restore after test
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
			// Update these to match the actual JSON formatting with spaces after colons
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
			// Setup mock MCR service
			mockMCRService := &MockMCRService{}
			tt.setupMock(mockMCRService)

			// Setup login to return our mock client
			loginFunc = func(ctx context.Context) (*megaport.Client, error) {
				client := &megaport.Client{}
				client.MCRService = mockMCRService
				return client, nil
			}

			// Set the global outputFormat variable
			outputFormat = tt.format

			// Execute command and capture output
			var err error
			output := captureOutput(func() {
				err = getMCRCmd.RunE(getMCRCmd, []string{tt.mcrID})
			})

			// Check results
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
