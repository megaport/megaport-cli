package cmd

import (
	"context"
	"fmt"
	"testing"

	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
)

var testMVEs = []*megaport.MVE{
	{
		UID:        "mve-1",
		Name:       "MyMVEOne",
		LocationID: 1,
	},
	{
		UID:        "mve-2",
		Name:       "AnotherMVE",
		LocationID: 2,
	},
}

func TestPrintMVEs_Table(t *testing.T) {
	output := captureOutput(func() {
		err := printMVEs(testMVEs, "table")
		assert.NoError(t, err)
	})

	expected := `uid     name         location_id
mve-1   MyMVEOne     1
mve-2   AnotherMVE   2
`
	assert.Equal(t, expected, output)
}

func TestPrintMVEs_JSON(t *testing.T) {
	output := captureOutput(func() {
		err := printMVEs(testMVEs, "json")
		assert.NoError(t, err)
	})

	expected := `[
  {
    "uid": "mve-1",
    "name": "MyMVEOne",
    "location_id": 1
  },
  {
    "uid": "mve-2",
    "name": "AnotherMVE",
    "location_id": 2
  }
]`
	assert.JSONEq(t, expected, output)
}

func TestPrintMVEs_CSV(t *testing.T) {
	output := captureOutput(func() {
		err := printMVEs(testMVEs, "csv")
		assert.NoError(t, err)
	})

	expected := `uid,name,location_id
mve-1,MyMVEOne,1
mve-2,AnotherMVE,2
`
	assert.Equal(t, expected, output)
}

func TestPrintMVEs_Invalid(t *testing.T) {
	var err error
	output := captureOutput(func() {
		err = printMVEs(testMVEs, "invalid")
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid output format")
	assert.Empty(t, output)
}

func TestPrintMVEs_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		mves        []*megaport.MVE
		format      string
		shouldError bool
		expected    string
	}{
		{
			name:        "nil slice",
			mves:        nil,
			format:      "table",
			shouldError: false,
			expected:    "uid   name   location_id\n",
		},
		{
			name:        "empty slice",
			mves:        []*megaport.MVE{},
			format:      "json",
			shouldError: false,
			expected:    "[]",
		},
		{
			name: "nil mve in slice",
			mves: []*megaport.MVE{
				nil,
				{
					UID:        "mve-1",
					Name:       "TestMVE",
					LocationID: 1,
				},
			},
			format:      "table",
			shouldError: true,
			expected:    "invalid MVE: nil value",
		},
		{
			name: "zero values",
			mves: []*megaport.MVE{
				{
					UID:        "",
					Name:       "",
					LocationID: 0,
				},
			},
			format:      "csv",
			shouldError: false,
			expected:    "uid,name,location_id\n,,0\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output string
			var err error

			output = captureOutput(func() {
				err = printMVEs(tt.mves, tt.format)
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

func TestToMVEOutput_EdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		mve           *megaport.MVE
		shouldError   bool
		errorContains string
		validateFunc  func(*testing.T, MVEOutput)
	}{
		{
			name:          "nil mve",
			mve:           nil,
			shouldError:   true,
			errorContains: "invalid MVE: nil value",
		},
		{
			name: "zero values",
			mve:  &megaport.MVE{},
			validateFunc: func(t *testing.T, output MVEOutput) {
				assert.Empty(t, output.UID)
				assert.Empty(t, output.Name)
				assert.Zero(t, output.LocationID)
			},
		},
		{
			name: "whitespace values",
			mve: &megaport.MVE{
				UID:        "   ",
				Name:       "   ",
				LocationID: 0,
			},
			validateFunc: func(t *testing.T, output MVEOutput) {
				assert.Equal(t, "   ", output.UID)
				assert.Equal(t, "   ", output.Name)
				assert.Zero(t, output.LocationID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := ToMVEOutput(tt.mve)

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

func TestBuyMVE(t *testing.T) {
	originalPrompt := prompt
	originalLoginFunc := loginFunc
	originalBuyMVEFunc := buyMVEFunc
	defer func() {
		prompt = originalPrompt
		loginFunc = originalLoginFunc
		buyMVEFunc = originalBuyMVEFunc
	}()

	tests := []struct {
		name            string
		prompts         []string
		mockSetup       func(*MockMVEService)
		expectedError   string
		expectedOutput  string
		validateRequest func(*testing.T, *megaport.BuyMVERequest)
	}{
		{
			name: "successful MVE purchase",
			prompts: []string{
				"Test MVE",   // name
				"12",         // term
				"123",        // location ID
				"cisco",      // vendor
				"1",          // image ID
				"large",      // product size
				"label-1",    // MVE label
				"true",       // manage locally
				"admin-ssh",  // admin SSH public key
				"ssh-key",    // SSH public key
				"cloud-init", // cloud init
				"fmc-ip",     // FMC IP address
				"fmc-key",    // FMC registration key
				"fmc-nat",    // FMC NAT ID
			},
			mockSetup: func(m *MockMVEService) {
				m.ValidateMVEOrderError = nil
				m.BuyMVEError = nil
				m.BuyMVEResult = &megaport.BuyMVEResponse{
					TechnicalServiceUID: "mock-mve-uid",
				}
			},
			expectedOutput: "MVE purchased successfully - UID: mock-mve-uid",
			validateRequest: func(t *testing.T, req *megaport.BuyMVERequest) {
				assert.Equal(t, "Test MVE", req.Name)
				assert.Equal(t, 12, req.Term)
				assert.Equal(t, 123, req.LocationID)

				ciscoConfig, ok := req.VendorConfig.(*megaport.CiscoConfig)
				assert.True(t, ok, "Expected a CiscoConfig")
				assert.Equal(t, 1, ciscoConfig.ImageID)
				assert.Equal(t, "large", ciscoConfig.ProductSize)
				assert.Equal(t, "label-1", ciscoConfig.MVELabel)
				assert.True(t, ciscoConfig.ManageLocally)
				assert.Equal(t, "admin-ssh", ciscoConfig.AdminSSHPublicKey)
				assert.Equal(t, "ssh-key", ciscoConfig.SSHPublicKey)
				assert.Equal(t, "cloud-init", ciscoConfig.CloudInit)
				assert.Equal(t, "fmc-ip", ciscoConfig.FMCIPAddress)
				assert.Equal(t, "fmc-key", ciscoConfig.FMCRegistrationKey)
				assert.Equal(t, "fmc-nat", ciscoConfig.FMCNatID)
			},
		},
		{
			name: "validation error",
			prompts: []string{
				"Test MVE",   // name
				"12",         // term
				"123",        // location ID
				"cisco",      // vendor
				"1",          // image ID
				"large",      // product size
				"label-1",    // MVE label
				"true",       // manage locally
				"admin-ssh",  // admin SSH public key
				"ssh-key",    // SSH public key
				"cloud-init", // cloud init
				"fmc-ip",     // FMC IP address
				"fmc-key",    // FMC registration key
				"fmc-nat",    // FMC NAT ID
			},
			mockSetup: func(m *MockMVEService) {
				m.ValidateMVEOrderError = fmt.Errorf("validation failed")
			},
			expectedError: "validation failed",
		},
		{
			name: "purchase error",
			prompts: []string{
				"Test MVE",   // name
				"12",         // term
				"123",        // location ID
				"cisco",      // vendor
				"1",          // image ID
				"large",      // product size
				"label-1",    // MVE label
				"true",       // manage locally
				"admin-ssh",  // admin SSH public key
				"ssh-key",    // SSH public key
				"cloud-init", // cloud init
				"fmc-ip",     // FMC IP address
				"fmc-key",    // FMC registration key
				"fmc-nat",    // FMC NAT ID
			},
			mockSetup: func(m *MockMVEService) {
				m.ValidateMVEOrderError = nil
				m.BuyMVEError = fmt.Errorf("purchase failed")
			},
			expectedError: "purchase failed",
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

			mockService := &MockMVEService{}
			tt.mockSetup(mockService)

			loginFunc = func(ctx context.Context) (*megaport.Client, error) {
				client := &megaport.Client{
					MVEService: mockService,
				}
				return client, nil
			}

			// Use the actual buyMVEFunc to make sure we call the mock service methods
			buyMVEFunc = func(ctx context.Context, client *megaport.Client, req *megaport.BuyMVERequest) (*megaport.BuyMVEResponse, error) {
				return client.MVEService.BuyMVE(ctx, req)
			}

			cmd := buyMVECmd
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

				if tt.validateRequest != nil {
					tt.validateRequest(t, mockService.CapturedBuyMVERequest)
				}
			}
		})
	}
}
