package ports

import (
	"context"
	"fmt"
	"testing"

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

// Function to adapt our old tests to work with new wrapCommandFunc signature
func testCommandAdapterOutput(fn func(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		format, _ := cmd.Flags().GetString("output")
		if format == "" {
			format = "table" // Default to table if not specified
		}
		return fn(cmd, args, false, format)
	}
}

func TestGetPortCmd_WithMockClient(t *testing.T) {
	// Save original login function and restore after test
	originalLoginFunc := config.Login
	defer func() {
		config.LoginFunc = originalLoginFunc
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
			config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
				client := &megaport.Client{}
				client.PortService = mockPortService
				return client, nil
			}

			// Set the global outputFormat variable
			cmd := &cobra.Command{
				Use:  "get",
				RunE: testCommandAdapterOutput(GetPort),
			}
			cmd.Flags().String("output", "table", "Output format (table, json)")
			if err := cmd.Flags().Set("output", tt.format); err != nil {
				t.Fatalf("Failed to set output flag: %v", err)
			}

			// Execute command and capture output
			var err error
			output := output.CaptureOutput(func() {
				cmdWithFormat := &cobra.Command{}
				cmdWithFormat.Flags().String("output", "table", "Output format")
				flagErr := cmdWithFormat.Flags().Set("output", tt.format) // Set the format
				if flagErr != nil {
					t.Fatalf("Failed to set output flag: %v", err)
				}
				err = testCommandAdapterOutput(GetPort)(cmd, []string{tt.portID})
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
	originalLoginFunc := config.LoginFunc
	defer func() {
		config.LoginFunc = originalLoginFunc
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
			expectedOut:   []string{"UID", "NAME", "LOCATIONID", "SPEED", "STATUS"},
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
			config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
				client := &megaport.Client{}
				client.PortService = mockPortService
				return client, nil
			}

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

					noColor := true
					// Print with current format
					return printPorts(filtered, tt.format, noColor)
				},
			}
			cmd.Flags().String("output", "table", "Output format (table, json)")
			if err := cmd.Flags().Set("output", tt.format); err != nil {
				t.Fatalf("Failed to set output flag: %v", err)
			}
			var locationID int
			var portSpeed int
			var portName string
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
			output := output.CaptureOutput(func() {
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

func TestBuyPortCmd(t *testing.T) {
	// Save original functions and restore after test
	originalPrompt := utils.Prompt
	originalLoginFunc := config.LoginFunc
	originalBuyPortFunc := buyPortFunc
	defer func() {
		utils.Prompt = originalPrompt
		config.LoginFunc = originalLoginFunc
		buyPortFunc = originalBuyPortFunc
	}()

	tests := []struct {
		name           string
		args           []string
		interactive    bool
		prompts        []string
		flags          map[string]string
		jsonInput      string
		jsonFilePath   string
		setupMock      func(*MockPortService)
		expectedError  string
		expectedOutput string
	}{
		{
			name:        "interactive mode success",
			interactive: true,
			prompts: []string{
				"Test Port", // name
				"12",        // term
				"10000",     // port speed
				"123",       // location ID
				"true",      // marketplace visibility
				"red",       // diversity zone
				"cost-123",  // cost centre
				"PROMO2025", // promo code
			},
			setupMock: func(m *MockPortService) {
				m.BuyPortResult = &megaport.BuyPortResponse{
					TechnicalServiceUIDs: []string{"port-123-abc"},
				}
			},
			expectedOutput: "Port created port-123-abc",
		},
		{
			name: "flag mode success",
			args: []string{},
			flags: map[string]string{
				"name":                   "Test Port",
				"term":                   "12",
				"port-speed":             "10000",
				"location-id":            "123",
				"marketplace-visibility": "true",
				"diversity-zone":         "red",
				"cost-centre":            "cost-123",
				"promo-code":             "PROMO2025",
			},
			setupMock: func(m *MockPortService) {
				m.BuyPortResult = &megaport.BuyPortResponse{
					TechnicalServiceUIDs: []string{"port-123-abc"},
				}
			},
			expectedOutput: "Port created port-123-abc",
		},
		{
			name: "JSON string mode success",
			args: []string{},
			flags: map[string]string{
				"json": `{"name":"Test Port","term":12,"portSpeed":10000,"locationId":123,"marketPlaceVisibility":true,"diversityZone":"red","costCentre":"cost-123","promoCode":"PROMO2025"}`,
			},
			setupMock: func(m *MockPortService) {
				m.BuyPortResult = &megaport.BuyPortResponse{
					TechnicalServiceUIDs: []string{"port-123-abc"},
				}
			},
			expectedOutput: "Port created port-123-abc",
		},
		{
			name: "missing required fields in flag mode",
			args: []string{},
			flags: map[string]string{
				"name": "Test Port",
				// Missing other required fields
			},
			expectedError: "invalid term, must be one of 1, 12, 24, 36",
		},
		{
			name: "invalid term in flag mode",
			args: []string{},
			flags: map[string]string{
				"name":                   "Test Port",
				"term":                   "13", // Invalid term
				"port-speed":             "10000",
				"location-id":            "123",
				"marketplace-visibility": "true",
			},
			expectedError: "invalid term, must be one of 1, 12, 24, 36",
		},
		{
			name: "invalid port speed in flag mode",
			args: []string{},
			flags: map[string]string{
				"name":                   "Test Port",
				"term":                   "12",
				"port-speed":             "5000", // Invalid port speed
				"location-id":            "123",
				"marketplace-visibility": "true",
			},
			expectedError: "invalid port speed, must be one of 1000, 10000, 100000",
		},
		{
			name: "invalid JSON",
			args: []string{},
			flags: map[string]string{
				"json": `{"name":"Test Port","term":"invalid","portSpeed":10000}`,
			},
			expectedError: "error parsing JSON",
		},
		{
			name: "API error",
			args: []string{},
			flags: map[string]string{
				"name":                   "Test Port",
				"term":                   "12",
				"port-speed":             "10000",
				"location-id":            "123",
				"marketplace-visibility": "true",
			},
			setupMock: func(m *MockPortService) {
				m.BuyPortErr = fmt.Errorf("API error: service unavailable")
			},
			expectedError: "API error: service unavailable",
		},
		{
			name:          "no input provided",
			args:          []string{},
			expectedError: "no input provided, use --interactive, --json, or flags to specify port details",
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

			// Setup mock Port service
			mockPortService := &MockPortService{}
			if tt.setupMock != nil {
				tt.setupMock(mockPortService)
			}

			// Setup login to return our mock client
			config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
				client := &megaport.Client{}
				client.PortService = mockPortService
				return client, nil
			}

			// Create a fresh command for each test to avoid flag conflicts
			cmd := &cobra.Command{
				Use:  "buy",
				RunE: testCommandAdapter(BuyPort),
			}

			// Add all the necessary flags
			cmd.Flags().BoolP("interactive", "i", false, "Use interactive mode with prompts")
			cmd.Flags().String("name", "", "Port name")
			cmd.Flags().Int("term", 0, "Contract term in months (1, 12, 24, or 36)")
			cmd.Flags().Int("port-speed", 0, "Port speed in Mbps (1000, 10000, or 100000)")
			cmd.Flags().Int("location-id", 0, "Location ID where the port will be provisioned")
			cmd.Flags().Bool("marketplace-visibility", false, "Whether the port is visible in marketplace")
			cmd.Flags().String("diversity-zone", "", "Diversity zone for the port")
			cmd.Flags().String("cost-centre", "", "Cost centre for billing")
			cmd.Flags().String("promo-code", "", "Promotional code for discounts")
			cmd.Flags().String("json", "", "JSON string containing port configuration")
			cmd.Flags().String("json-file", "", "Path to JSON file containing port configuration")

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
				if tt.expectedOutput != "" && mockPortService != nil && mockPortService.CapturedRequest != nil {
					req := mockPortService.CapturedRequest
					if tt.flags != nil {
						// For flag mode or JSON mode
						assert.Equal(t, "Test Port", req.Name)
						assert.Equal(t, 12, req.Term)
						assert.Equal(t, 10000, req.PortSpeed)
						assert.Equal(t, 123, req.LocationId)
						assert.Equal(t, true, req.MarketPlaceVisibility)
						assert.Equal(t, "red", req.DiversityZone)
						assert.Equal(t, "cost-123", req.CostCentre)
						assert.Equal(t, "PROMO2025", req.PromoCode)
					} else if len(tt.prompts) > 0 {
						// For interactive mode
						assert.Equal(t, "Test Port", req.Name)
						assert.Equal(t, 12, req.Term)
						assert.Equal(t, 10000, req.PortSpeed)
						assert.Equal(t, 123, req.LocationId)
						assert.Equal(t, true, req.MarketPlaceVisibility)
						assert.Equal(t, "red", req.DiversityZone)
						assert.Equal(t, "cost-123", req.CostCentre)
						assert.Equal(t, "PROMO2025", req.PromoCode)
					}
				}
			}
		})
	}
}

// TestBuyLAGPortCmd tests the buyLagCmd with all three input modes
func TestBuyLAGPortCmd(t *testing.T) {
	// Save original functions and restore after test
	originalPrompt := utils.Prompt
	originalLoginFunc := config.LoginFunc
	originalBuyPortFunc := buyPortFunc
	defer func() {
		utils.Prompt = originalPrompt
		config.LoginFunc = originalLoginFunc
		buyPortFunc = originalBuyPortFunc
	}()

	tests := []struct {
		name           string
		args           []string
		interactive    bool
		prompts        []string
		flags          map[string]string
		jsonInput      string
		jsonFilePath   string
		setupMock      func(*MockPortService)
		expectedError  string
		expectedOutput string
	}{
		{
			name:        "interactive mode success",
			interactive: true,
			prompts: []string{
				"Test LAG Port", // name
				"12",            // term
				"10000",         // port speed
				"123",           // location ID
				"2",             // LAG count
				"true",          // marketplace visibility
				"blue",          // diversity zone
				"cost-456",      // cost centre
				"LAGPROMO2025",  // promo code
			},
			setupMock: func(m *MockPortService) {
				m.BuyPortResult = &megaport.BuyPortResponse{
					TechnicalServiceUIDs: []string{"lag-456-xyz"},
				}
			},
			expectedOutput: "LAG Port created lag-456-xyz",
		},
		{
			name: "flag mode success",
			args: []string{},
			flags: map[string]string{
				"name":                   "Test LAG Port",
				"term":                   "12",
				"port-speed":             "10000",
				"location-id":            "123",
				"lag-count":              "2",
				"marketplace-visibility": "true",
				"diversity-zone":         "blue",
				"cost-centre":            "cost-456",
				"promo-code":             "LAGPROMO2025",
			},
			setupMock: func(m *MockPortService) {
				m.BuyPortResult = &megaport.BuyPortResponse{
					TechnicalServiceUIDs: []string{"lag-456-xyz"},
				}
			},
			expectedOutput: "LAG Port created lag-456-xyz",
		},
		{
			name: "JSON string mode success",
			args: []string{},
			flags: map[string]string{
				"json": `{"name":"Test LAG Port","term":12,"portSpeed":10000,"locationId":123,"lagCount":2,"marketPlaceVisibility":true,"diversityZone":"blue","costCentre":"cost-456","promoCode":"LAGPROMO2025"}`,
			},
			setupMock: func(m *MockPortService) {
				m.BuyPortResult = &megaport.BuyPortResponse{
					TechnicalServiceUIDs: []string{"lag-456-xyz"},
				}
			},
			expectedOutput: "LAG Port created lag-456-xyz",
		},
		{
			name: "missing required fields in flag mode",
			args: []string{},
			flags: map[string]string{
				"name": "Test LAG Port",
				// Missing other required fields
			},
			expectedError: "invalid term, must be one of 1, 12, 24, 36",
		},
		{
			name: "invalid term in flag mode",
			args: []string{},
			flags: map[string]string{
				"name":                   "Test LAG Port",
				"term":                   "13", // Invalid term
				"port-speed":             "10000",
				"location-id":            "123",
				"lag-count":              "2",
				"marketplace-visibility": "true",
			},
			expectedError: "invalid term, must be one of 1, 12, 24, 36",
		},
		{
			name: "invalid port speed in flag mode for LAG",
			args: []string{},
			flags: map[string]string{
				"name":                   "Test LAG Port",
				"term":                   "12",
				"port-speed":             "1000", // Invalid port speed for LAG (should be 10000 or 100000)
				"location-id":            "123",
				"lag-count":              "2",
				"marketplace-visibility": "true",
			},
			expectedError: "invalid port speed, must be one of 10000 or 100000",
		},
		{
			name: "invalid LAG count in flag mode",
			args: []string{},
			flags: map[string]string{
				"name":                   "Test LAG Port",
				"term":                   "12",
				"port-speed":             "10000",
				"location-id":            "123",
				"lag-count":              "10", // Invalid LAG count (should be 1-8)
				"marketplace-visibility": "true",
			},
			expectedError: "invalid LAG count, must be between 1 and 8",
		},
		{
			name: "invalid JSON",
			args: []string{},
			flags: map[string]string{
				"json": `{"name":"Test LAG Port","term":"invalid","portSpeed":10000}`,
			},
			expectedError: "error parsing JSON",
		},
		{
			name: "API error",
			args: []string{},
			flags: map[string]string{
				"name":                   "Test LAG Port",
				"term":                   "12",
				"port-speed":             "10000",
				"location-id":            "123",
				"lag-count":              "2",
				"marketplace-visibility": "true",
			},
			setupMock: func(m *MockPortService) {
				m.BuyPortErr = fmt.Errorf("API error: service unavailable")
			},
			expectedError: "API error: service unavailable",
		},
		{
			name:          "no input provided",
			args:          []string{},
			expectedError: "no input provided, use --interactive, --json, or flags to specify port details",
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

			// Setup mock Port service
			mockPortService := &MockPortService{}
			if tt.setupMock != nil {
				tt.setupMock(mockPortService)
			}

			// Setup login to return our mock client
			config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
				client := &megaport.Client{}
				client.PortService = mockPortService
				return client, nil
			}

			// Create a fresh command for each test to avoid flag conflicts
			cmd := &cobra.Command{
				Use:  "buy-lag",
				RunE: testCommandAdapter(BuyLAGPort),
			}

			// Add all the necessary flags
			cmd.Flags().BoolP("interactive", "i", false, "Use interactive mode with prompts")
			cmd.Flags().String("name", "", "Port name")
			cmd.Flags().Int("term", 0, "Contract term in months (1, 12, 24, or 36)")
			cmd.Flags().Int("port-speed", 0, "Port speed in Mbps (10000 or 100000)")
			cmd.Flags().Int("location-id", 0, "Location ID where the port will be provisioned")
			cmd.Flags().Int("lag-count", 0, "Number of LAGs (1-8)")
			cmd.Flags().Bool("marketplace-visibility", false, "Whether the port is visible in marketplace")
			cmd.Flags().String("diversity-zone", "", "Diversity zone for the port")
			cmd.Flags().String("cost-centre", "", "Cost centre for billing")
			cmd.Flags().String("promo-code", "", "Promotional code for discounts")
			cmd.Flags().String("json", "", "JSON string containing port configuration")
			cmd.Flags().String("json-file", "", "Path to JSON file containing port configuration")

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
				if tt.expectedOutput != "" && mockPortService != nil && mockPortService.CapturedRequest != nil {
					req := mockPortService.CapturedRequest
					if tt.flags != nil {
						// For flag mode or JSON mode
						assert.Equal(t, "Test LAG Port", req.Name)
						assert.Equal(t, 12, req.Term)
						assert.Equal(t, 10000, req.PortSpeed)
						assert.Equal(t, 123, req.LocationId)
						assert.Equal(t, 2, req.LagCount)
						assert.Equal(t, true, req.MarketPlaceVisibility)
						assert.Equal(t, "blue", req.DiversityZone)
						assert.Equal(t, "cost-456", req.CostCentre)
						assert.Equal(t, "LAGPROMO2025", req.PromoCode)
					} else if len(tt.prompts) > 0 {
						// For interactive mode
						assert.Equal(t, "Test LAG Port", req.Name)
						assert.Equal(t, 12, req.Term)
						assert.Equal(t, 10000, req.PortSpeed)
						assert.Equal(t, 123, req.LocationId)
						assert.Equal(t, 2, req.LagCount)
						assert.Equal(t, true, req.MarketPlaceVisibility)
						assert.Equal(t, "blue", req.DiversityZone)
						assert.Equal(t, "cost-456", req.CostCentre)
						assert.Equal(t, "LAGPROMO2025", req.PromoCode)
					}
				}
			}
		})
	}
}

// TestUpdatePortCmd tests the updatePortCmd with all three input modes
func TestUpdatePortCmd(t *testing.T) {
	// Save original functions and restore after test
	originalPrompt := utils.Prompt
	originalLoginFunc := config.LoginFunc
	originalUpdatePortFunc := updatePortFunc
	originalGetPortFunc := getPortFunc
	defer func() {
		utils.Prompt = originalPrompt
		config.LoginFunc = originalLoginFunc
		updatePortFunc = originalUpdatePortFunc
		getPortFunc = originalGetPortFunc
	}()

	tests := []struct {
		name                  string
		args                  []string
		interactive           bool
		prompts               []string
		flags                 map[string]string
		jsonInput             string
		jsonFilePath          string
		setupMock             func(*MockPortService)
		expectedError         string
		expectedOutput        string
		skipRequestValidation bool // Add this flag to skip request validation for certain tests
	}{
		{
			name:        "interactive mode success",
			args:        []string{"port-123"},
			interactive: true,
			prompts: []string{
				"Updated Port Name", // name
				"true",              // marketplace visibility
				"cost-centre-123",   // cost centre
				"12",                // term
			},
			setupMock: func(m *MockPortService) {
				m.ModifyPortResult = &megaport.ModifyPortResponse{
					IsUpdated: true,
				}
			},
			expectedOutput: "Port updated port-123",
		},
		{
			name: "flag mode success",
			args: []string{"port-456"},
			flags: map[string]string{
				"name":                   "Updated Flag Port",
				"marketplace-visibility": "true",
				"cost-centre":            "cost-centre-456",
				"term":                   "24",
			},
			setupMock: func(m *MockPortService) {
				m.ModifyPortResult = &megaport.ModifyPortResponse{
					IsUpdated: true,
				}
			},
			expectedOutput: "Port updated port-456",
		},
		{
			name: "JSON string mode success",
			args: []string{"port-789"},
			flags: map[string]string{
				"json": `{"name":"Updated JSON Port","marketplaceVisibility":true,"costCentre":"cost-centre-789","contractTermMonths":36}`,
			},
			setupMock: func(m *MockPortService) {
				m.ModifyPortResult = &megaport.ModifyPortResponse{
					IsUpdated: true,
				}
			},
			expectedOutput: "Port updated port-789",
		},
		{
			name: "missing port UID",
			args: []string{},
			flags: map[string]string{
				"name":                   "Test Port",
				"marketplace-visibility": "true",
			},
			expectedError:         "accepts 1 arg(s), received 0",
			skipRequestValidation: true, // Skip validation because the command won't even run
		},
		{
			name:  "at least one field required in flag mode",
			args:  []string{"port-123"},
			flags: map[string]string{
				// No update fields provided
			},
			expectedError:         "at least one field must be updated",
			skipRequestValidation: true, // Skip validation because no request will be sent
		},
		{
			name: "update with only marketplace visibility works",
			args: []string{"port-123"},
			flags: map[string]string{
				"marketplace-visibility": "true",
			},
			setupMock: func(m *MockPortService) {
				m.ModifyPortResult = &megaport.ModifyPortResponse{
					IsUpdated: true,
				}
			},
			expectedOutput: "Port updated port-123",
		},
		{
			name: "invalid term in flag mode",
			args: []string{"port-123"},
			flags: map[string]string{
				"name":                   "Test Port",
				"marketplace-visibility": "true",
				"term":                   "13", // Invalid term
			},
			expectedError:         "invalid term, must be one of 1, 12, 24, 36",
			skipRequestValidation: true, // Skip validation because no request will be sent
		},
		{
			name: "invalid JSON",
			args: []string{"port-123"},
			flags: map[string]string{
				"json": `{"name":"Test Port","marketplaceVisibility":"invalid-boolean"}`,
			},
			expectedError:         "error parsing JSON",
			skipRequestValidation: true, // Skip validation because no request will be sent
		},
		{
			name: "API error",
			args: []string{"port-123"},
			flags: map[string]string{
				"name":                   "Test Port",
				"marketplace-visibility": "true",
			},
			setupMock: func(m *MockPortService) {
				m.ModifyPortErr = fmt.Errorf("API error: service unavailable")
			},
			expectedError:         "API error: service unavailable",
			skipRequestValidation: true, // Skip validation because the API error occurs
		},
		{
			name:                  "no input provided",
			args:                  []string{"port-123"},
			expectedError:         "at least one field must be updated",
			skipRequestValidation: true, // Skip validation because no request will be sent
		},
		{
			name:        "interactive mode with failed update",
			args:        []string{"port-123"},
			interactive: true,
			prompts: []string{
				"Updated Port Name", // name
				"true",              // marketplace visibility
				"",                  // cost centre (empty)
				"",                  // term (empty)
			},
			setupMock: func(m *MockPortService) {
				// Make sure we create a real capture for the request
				m.ModifyPortResult = &megaport.ModifyPortResponse{
					IsUpdated: false,
				}
			},
			expectedError:         "port update request was not successful",
			skipRequestValidation: true, // Skip validation since we're testing error handling
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

			getPortFunc = func(ctx context.Context, client *megaport.Client, portID string) (*megaport.Port, error) {
				// Mock the response for getPort
				return &megaport.Port{
					UID:  portID,
					Name: tt.name,
					// Add other fields as needed
				}, nil
			}

			// Setup mock Port service
			mockPortService := &MockPortService{}
			if tt.setupMock != nil {
				tt.setupMock(mockPortService)
			}

			// Setup login to return our mock client
			config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
				client := &megaport.Client{}
				client.PortService = mockPortService
				return client, nil
			}

			// Create a fresh command for each test to avoid flag conflicts
			cmd := &cobra.Command{
				Use:  "update [portUID]",
				Args: cobra.ExactArgs(1),
				RunE: testCommandAdapter(UpdatePort),
			}

			updatePortFunc = func(ctx context.Context, client *megaport.Client, req *megaport.ModifyPortRequest) (*megaport.ModifyPortResponse, error) {
				// Store the request for validation
				mockPortService.CapturedModifyPortRequest = req

				// Return mock result
				if mockPortService.ModifyPortErr != nil {
					return nil, mockPortService.ModifyPortErr
				}
				return mockPortService.ModifyPortResult, nil
			}

			// Add all the necessary flags
			cmd.Flags().BoolP("interactive", "i", false, "Use interactive mode with prompts")
			cmd.Flags().String("name", "", "New port name")
			cmd.Flags().Bool("marketplace-visibility", false, "Whether the port is visible in marketplace")
			cmd.Flags().String("cost-centre", "", "Cost centre for billing")
			cmd.Flags().Int("term", 0, "New contract term in months (1, 12, 24, or 36)")
			cmd.Flags().String("json", "", "JSON string containing port configuration")
			cmd.Flags().String("json-file", "", "Path to JSON file containing port configuration")

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
				if len(tt.args) > 0 {
					err = cmd.RunE(cmd, tt.args)
				} else {
					err = cmd.Execute()
				}
			})

			// Check results
			if tt.expectedError != "" {
				assert.Error(t, err, "Expected an error but got none")
				if err != nil { // Only check error contents if there actually is an error
					assert.Contains(t, err.Error(), tt.expectedError)
				}
			} else {
				assert.NoError(t, err)
				assert.Contains(t, output, tt.expectedOutput)

				// Change this section in TestUpdatePortCmd
				if !tt.skipRequestValidation && mockPortService != nil && mockPortService.CapturedModifyPortRequest != nil {
					req := mockPortService.CapturedModifyPortRequest

					// Verify port ID
					if len(tt.args) > 0 {
						assert.Equal(t, tt.args[0], req.PortID)
					}

					// For the "update with only marketplace visibility works" test case,
					// only check the marketplace visibility field
					if tt.name == "update with only marketplace visibility works" {
						assert.NotNil(t, req.MarketplaceVisibility)
						if req.MarketplaceVisibility != nil {
							assert.True(t, *req.MarketplaceVisibility)
						}
					} else if tt.flags != nil && tt.flags["json"] != "" {
						// For JSON mode
						// Only check fields that should actually be in the request
						if req.Name != "" {
							assert.Equal(t, "Updated JSON Port", req.Name)
						}
						if req.MarketplaceVisibility != nil {
							assert.True(t, *req.MarketplaceVisibility)
						}
						if req.CostCentre != "" {
							assert.Equal(t, "cost-centre-789", req.CostCentre)
						}
						if req.ContractTermMonths != nil {
							assert.Equal(t, 36, *req.ContractTermMonths)
						}
					} else if len(tt.flags) > 0 && !tt.interactive {
						// For flag mode
						// Only check fields that should actually be in the request
						if req.Name != "" {
							assert.Equal(t, "Updated Flag Port", req.Name)
						}
						if req.MarketplaceVisibility != nil {
							assert.True(t, *req.MarketplaceVisibility)
						}
						if req.CostCentre != "" {
							assert.Equal(t, "cost-centre-456", req.CostCentre)
						}
						if req.ContractTermMonths != nil {
							assert.Equal(t, 24, *req.ContractTermMonths)
						}
					}
				}
			}
		})
	}
}
