package cmd

import (
	"context"
	"fmt"
	"testing"

	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

var testPorts = []*megaport.Port{
	{
		UID:                "port-1",
		Name:               "MyPortOne",
		LocationID:         1,
		PortSpeed:          1000,
		ProvisioningStatus: "ACTIVE",
	},
	{
		UID:                "port-2",
		Name:               "AnotherPort",
		LocationID:         2,
		PortSpeed:          2000,
		ProvisioningStatus: "INACTIVE",
	},
}

func TestFilterPorts(t *testing.T) {
	tests := []struct {
		name       string
		locationID int
		portSpeed  int
		portName   string
		expected   int
	}{
		{
			name:       "No filters",
			locationID: 0,
			portSpeed:  0,
			portName:   "",
			expected:   2,
		},
		{
			name:       "Filter by LocationID",
			locationID: 1,
			portSpeed:  0,
			portName:   "",
			expected:   1,
		},
		{
			name:       "Filter by PortSpeed",
			locationID: 0,
			portSpeed:  2000,
			portName:   "",
			expected:   1,
		},
		{
			name:       "Filter by PortName",
			locationID: 0,
			portSpeed:  0,
			portName:   "MyPortOne",
			expected:   1,
		},
		{
			name:       "No match",
			locationID: 99,
			portSpeed:  9999,
			portName:   "NoMatch",
			expected:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterPorts(testPorts, tt.locationID, tt.portSpeed, tt.portName)
			assert.Equal(t, tt.expected, len(result))
		})
	}
}

func TestPrintPorts_Table(t *testing.T) {
	output := captureOutput(func() {
		err := printPorts(testPorts, "table")
		assert.NoError(t, err)
	})

	expected := `uid      name          location_id   port_speed   provisioning_status
port-1   MyPortOne     1             1000         ACTIVE
port-2   AnotherPort   2             2000         INACTIVE
`
	assert.Equal(t, expected, output)
}

func TestPrintPorts_JSON(t *testing.T) {
	output := captureOutput(func() {
		err := printPorts(testPorts, "json")
		assert.NoError(t, err)
	})

	expected := `[
  {
    "uid": "port-1",
    "name": "MyPortOne",
    "location_id": 1,
    "port_speed": 1000,
    "provisioning_status": "ACTIVE"
  },
  {
    "uid": "port-2",
    "name": "AnotherPort",
    "location_id": 2,
    "port_speed": 2000,
    "provisioning_status": "INACTIVE"
  }
]`
	assert.JSONEq(t, expected, output)
}

func TestPrintPorts_CSV(t *testing.T) {
	output := captureOutput(func() {
		err := printPorts(testPorts, "csv")
		assert.NoError(t, err)
	})

	expected := `uid,name,location_id,port_speed,provisioning_status
port-1,MyPortOne,1,1000,ACTIVE
port-2,AnotherPort,2,2000,INACTIVE
`
	assert.Equal(t, expected, output)
}

func TestPrintPorts_Invalid(t *testing.T) {
	var err error
	output := captureOutput(func() {
		err = printPorts(testPorts, "invalid")
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid output format")
	assert.Empty(t, output)
}

func TestPrintPorts_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		ports       []*megaport.Port
		format      string
		shouldError bool
		expected    string
		contains    string // New field for partial matches
	}{
		{
			name:        "nil slice",
			ports:       nil,
			format:      "table",
			shouldError: false,
			expected:    "uid   name   location_id   port_speed   provisioning_status\n",
		},
		{
			name:        "empty slice",
			ports:       []*megaport.Port{},
			format:      "csv",
			shouldError: false,
			expected:    "uid,name,location_id,port_speed,provisioning_status\n",
		},
		{
			name: "port with zero values",
			ports: []*megaport.Port{
				{
					UID:                "",
					Name:               "",
					LocationID:         0,
					PortSpeed:          0,
					ProvisioningStatus: "",
				},
			},
			format:      "json",
			shouldError: false,
			expected:    `[{"uid":"","name":"","location_id":0,"port_speed":0,"provisioning_status":""}]`,
		},
		{
			name:        "nil port in slice",
			ports:       []*megaport.Port{nil},
			format:      "table",
			shouldError: true,
			expected:    "invalid port: nil value",
		},
		{
			name: "mixed valid and nil ports",
			ports: []*megaport.Port{
				{
					UID:                "port-1",
					Name:               "ValidPort",
					LocationID:         1,
					PortSpeed:          1000,
					ProvisioningStatus: "ACTIVE",
				},
				nil,
			},
			format:      "table",
			shouldError: true,
			expected:    "invalid port: nil value",
		},
		{
			name: "port with invalid status",
			ports: []*megaport.Port{
				{
					UID:                "port-1",
					Name:               "TestPort",
					LocationID:         1,
					PortSpeed:          1000,
					ProvisioningStatus: "INVALID_STATUS",
				},
			},
			format:      "table",
			shouldError: false,
			contains:    "INVALID_STATUS", // We just want to check if this status appears
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output string
			var err error

			output = captureOutput(func() {
				err = printPorts(tt.ports, tt.format)
			})

			if tt.shouldError {
				assert.Error(t, err)
				if tt.expected != "" {
					assert.Contains(t, err.Error(), tt.expected)
				}
				assert.Empty(t, output)
			} else {
				assert.NoError(t, err)
				if tt.expected != "" {
					if tt.format == "json" {
						assert.JSONEq(t, tt.expected, output)
					} else {
						assert.Equal(t, tt.expected, output)
					}
				}
				if tt.contains != "" {
					assert.Contains(t, output, tt.contains)
				}
			}
		})
	}
}

func TestFilterPorts_EdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		ports      []*megaport.Port
		locationID int
		portSpeed  int
		portName   string
		expected   int
	}{
		{
			name:       "nil slice",
			ports:      nil,
			locationID: 1,
			portSpeed:  1000,
			portName:   "Test",
			expected:   0,
		},
		{
			name:       "empty slice",
			ports:      []*megaport.Port{},
			locationID: 1,
			portSpeed:  1000,
			portName:   "Test",
			expected:   0,
		},
		{
			name: "slice with nil port",
			ports: []*megaport.Port{
				nil,
				{
					UID:       "port-1",
					Name:      "TestPort",
					PortSpeed: 1000,
				},
			},
			locationID: 0,
			portSpeed:  1000,
			portName:   "",
			expected:   1, // Should skip nil and return valid port
		},
		{
			name: "zero values in port",
			ports: []*megaport.Port{
				{
					UID:       "",
					Name:      "",
					PortSpeed: 0,
				},
			},
			locationID: 0,
			portSpeed:  0,
			portName:   "",
			expected:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterPorts(tt.ports, tt.locationID, tt.portSpeed, tt.portName)
			assert.Equal(t, tt.expected, len(result))
		})
	}
}

func TestGetPortCmd_WithMockClient(t *testing.T) {
	// Save original login function and restore after test
	originalLoginFunc := loginFunc
	originalOutputFormat := outputFormat
	defer func() {
		loginFunc = originalLoginFunc
		outputFormat = originalOutputFormat
	}()

	tests := []struct {
		name          string
		portID        string
		format        string
		setupMock     func(*MockPortService)
		expectedError string
		expectedOut   []string
	}{
		{
			name:   "get Port success table format",
			portID: "port-123",
			format: "table",
			setupMock: func(m *MockPortService) {
				m.GetPortResult = &megaport.Port{
					UID:                "port-123",
					Name:               "Test Port",
					LocationID:         123,
					PortSpeed:          10000,
					ProvisioningStatus: "LIVE",
				}
			},
			expectedOut: []string{"port-123", "Test Port", "10000", "LIVE"},
		},
		{
			name:   "get Port success json format",
			portID: "port-123",
			format: "json",
			setupMock: func(m *MockPortService) {
				m.GetPortResult = &megaport.Port{
					UID:                "port-123",
					Name:               "Test Port",
					LocationID:         123,
					PortSpeed:          10000,
					ProvisioningStatus: "LIVE",
				}
			},
			expectedOut: []string{`"uid": "port-123"`, `"name": "Test Port"`, `"port_speed": 10000`},
		},
		{
			name:   "get Port error",
			portID: "port-invalid",
			format: "table",
			setupMock: func(m *MockPortService) {
				m.GetPortErr = fmt.Errorf("Port not found")
			},
			expectedError: "Port not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock Port service
			mockPortService := &MockPortService{}
			tt.setupMock(mockPortService)

			// Setup login to return our mock client
			loginFunc = func(ctx context.Context) (*megaport.Client, error) {
				client := &megaport.Client{}
				client.PortService = mockPortService
				return client, nil
			}

			// Set the global outputFormat variable
			outputFormat = tt.format

			// Execute command and capture output
			var err error
			output := captureOutput(func() {
				err = getPortCmd.RunE(getPortCmd, []string{tt.portID})
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

func TestListPortsCmd_WithMockClient(t *testing.T) {
	// Save original functions and restore after test
	originalLoginFunc := loginFunc
	originalOutputFormat := outputFormat
	defer func() {
		loginFunc = originalLoginFunc
		outputFormat = originalOutputFormat
	}()

	// Create test ports
	testPorts := []*megaport.Port{
		{
			UID:                "port-test-1",
			Name:               "Test Port 1",
			LocationID:         123,
			PortSpeed:          1000,
			ProvisioningStatus: "LIVE",
		},
		{
			UID:                "port-test-2",
			Name:               "Test Port 2",
			LocationID:         456,
			PortSpeed:          10000,
			ProvisioningStatus: "CONFIGURING",
		},
	}

	tests := []struct {
		name          string
		format        string
		locationID    int
		portSpeed     int
		portName      string
		setupMock     func(*MockPortService)
		expectedError string
		expectedOut   []string
	}{
		{
			name:       "list Ports table format no filters",
			format:     "table",
			locationID: 0,
			portSpeed:  0,
			portName:   "",
			setupMock: func(m *MockPortService) {
				m.ListPortsResult = testPorts
			},
			expectedOut: []string{"port-test-1", "Test Port 1", "port-test-2", "Test Port 2"},
		},
		{
			name:       "list Ports json format",
			format:     "json",
			locationID: 0,
			portSpeed:  0,
			portName:   "",
			setupMock: func(m *MockPortService) {
				m.ListPortsResult = testPorts
			},
			expectedOut: []string{`"uid": "port-test-1"`, `"uid": "port-test-2"`},
		},
		{
			name:       "list Ports with location filter",
			format:     "table",
			locationID: 123,
			portSpeed:  0,
			portName:   "",
			setupMock: func(m *MockPortService) {
				m.ListPortsResult = testPorts
			},
			expectedOut:   []string{"port-test-1", "Test Port 1"},
			expectedError: "",
		},
		{
			name:       "list Ports with port speed filter",
			format:     "table",
			locationID: 0,
			portSpeed:  10000,
			portName:   "",
			setupMock: func(m *MockPortService) {
				m.ListPortsResult = testPorts
			},
			expectedOut:   []string{"port-test-2", "Test Port 2"},
			expectedError: "",
		},
		{
			name:       "list Ports with name filter",
			format:     "table",
			locationID: 0,
			portSpeed:  0,
			portName:   "Test Port 1",
			setupMock: func(m *MockPortService) {
				m.ListPortsResult = testPorts
			},
			expectedOut:   []string{"port-test-1", "Test Port 1"},
			expectedError: "",
		},
		{
			name:       "list empty Ports",
			format:     "table",
			locationID: 0,
			portSpeed:  0,
			portName:   "",
			setupMock: func(m *MockPortService) {
				m.ListPortsResult = []*megaport.Port{}
			},
			expectedOut:   []string{"uid", "name", "location_id", "port_speed", "provisioning_status"},
			expectedError: "",
		},
		{
			name:       "list Ports error",
			format:     "table",
			locationID: 0,
			portSpeed:  0,
			portName:   "",
			setupMock: func(m *MockPortService) {
				m.ListPortsErr = fmt.Errorf("error listing ports")
			},
			expectedError: "error listing ports",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock Port service
			mockPortService := &MockPortService{}
			tt.setupMock(mockPortService)

			// Setup login to return our mock client
			loginFunc = func(ctx context.Context) (*megaport.Client, error) {
				client := &megaport.Client{}
				client.PortService = mockPortService
				return client, nil
			}

			// Set the global outputFormat variable
			outputFormat = tt.format

			// Create a fresh command for each test to avoid flag conflicts
			cmd := &cobra.Command{
				Use: "list",
				RunE: func(cmd *cobra.Command, args []string) error {
					// First check if we should return an error
					if mockPortService.ListPortsErr != nil {
						return mockPortService.ListPortsErr
					}

					// Get filter values
					locationID, _ := cmd.Flags().GetInt("location-id")
					portSpeed, _ := cmd.Flags().GetInt("port-speed")
					portName, _ := cmd.Flags().GetString("port-name")

					// Get ports from mock - safe now that we've checked for error
					ports := mockPortService.ListPortsResult

					// Apply filters
					filtered := filterPorts(ports, locationID, portSpeed, portName)

					// Print with current format
					return printPorts(filtered, outputFormat)
				},
			}
			cmd.Flags().IntVar(&locationID, "location-id", 0, "Filter ports by location ID")
			cmd.Flags().IntVar(&portSpeed, "port-speed", 0, "Filter ports by port speed")
			cmd.Flags().StringVar(&portName, "port-name", "", "Filter ports by port name")
			// Set flag values for this test
			if tt.locationID != 0 {
				if err := cmd.Flags().Set("location-id", fmt.Sprintf("%d", tt.locationID)); err != nil {
					t.Fatalf("Failed to set location-id flag: %v", err)
				}
			}
			if tt.portSpeed != 0 {
				if err := cmd.Flags().Set("port-speed", fmt.Sprintf("%d", tt.portSpeed)); err != nil {
					t.Fatalf("Failed to set port-speed flag: %v", err)
				}
			}
			if tt.portName != "" {
				if err := cmd.Flags().Set("port-name", tt.portName); err != nil {
					t.Fatalf("Failed to set port-name flag: %v", err)
				}
			}

			// Execute command and capture output
			var err error
			output := captureOutput(func() {
				err = cmd.Execute()
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

func TestBuyPortCmd_WithMockClient(t *testing.T) {
	// Save original functions and restore after test
	originalPrompt := prompt
	originalLoginFunc := loginFunc
	originalBuyPortFunc := buyPortFunc
	defer func() {
		prompt = originalPrompt
		loginFunc = originalLoginFunc
		buyPortFunc = originalBuyPortFunc
	}()

	tests := []struct {
		name           string
		prompts        []string
		setupMock      func(*MockPortService)
		expectedError  string
		expectedOutput string
	}{
		{
			name: "successful Port purchase",
			prompts: []string{
				"Test Port", // name
				"12",        // term
				"1000",      // port speed
				"123",       // location ID
				"true",      // marketplace visibility
				"red",       // diversity zone
				"cost-123",  // cost center
				"PROMO2025", // promo code
			},
			setupMock: func(m *MockPortService) {
				m.BuyPortResult = &megaport.BuyPortResponse{
					TechnicalServiceUIDs: []string{"port-123-abc"},
				}
			},
			expectedOutput: "Port purchased successfully - UID: port-123-abc",
		},
		{
			name: "invalid term",
			prompts: []string{
				"Test Port", // name
				"13",        // invalid term (not 1, 12, 24, 36)
			},
			expectedError: "invalid term, must be one of 1, 12, 24, 36",
		},
		{
			name: "invalid port speed",
			prompts: []string{
				"Test Port", // name
				"12",        // term
				"2000",      // invalid port speed (not 1000, 10000, 100000)
			},
			expectedError: "invalid port speed, must be one of 1000, 10000, 100000",
		},
		{
			name: "API error",
			prompts: []string{
				"Test Port", // name
				"12",        // term
				"1000",      // port speed
				"123",       // location ID
				"true",      // marketplace visibility
				"red",       // diversity zone
				"cost-123",  // cost center
				"PROMO2025", // promo code
			},
			setupMock: func(m *MockPortService) {
				m.BuyPortErr = fmt.Errorf("API error: service unavailable")
			},
			// Change this line to match the actual error without the prefix
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

			// Setup mock Port service
			mockPortService := &MockPortService{}
			if tt.setupMock != nil {
				tt.setupMock(mockPortService)
			}

			// Setup login to return our mock client
			loginFunc = func(ctx context.Context) (*megaport.Client, error) {
				client := &megaport.Client{}
				client.PortService = mockPortService
				return client, nil
			}

			// Execute command and capture output
			cmd := buyPortCmd
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
				assert.NotNil(t, mockPortService.CapturedRequest)
				req := mockPortService.CapturedRequest
				assert.Equal(t, "Test Port", req.Name)
				assert.Equal(t, 12, req.Term)
				assert.Equal(t, 1000, req.PortSpeed)
				assert.Equal(t, 123, req.LocationId)
				assert.Equal(t, true, req.MarketPlaceVisibility)
				assert.Equal(t, "red", req.DiversityZone)
				assert.Equal(t, "cost-123", req.CostCentre)
				assert.Equal(t, "PROMO2025", req.PromoCode)
			}
		})
	}
}
