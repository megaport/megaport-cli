package ports

import (
	"context"
	"fmt"
	"testing"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/commands/config"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func testCommandAdapterOutput(fn func(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		format, _ := cmd.Flags().GetString("output")
		if format == "" {
			format = "table"
		}
		return fn(cmd, args, false, format)
	}
}

func TestGetPortStatus(t *testing.T) {
	originalLoginFunc := config.LoginFunc
	defer func() {
		config.LoginFunc = originalLoginFunc
	}()

	tests := []struct {
		name           string
		portUID        string
		setupMock      func(*MockPortService)
		expectedError  string
		expectedOutput string
		outputFormat   string
	}{
		{
			name:    "successful status retrieval - table format",
			portUID: "port-123abc",
			setupMock: func(m *MockPortService) {
				m.GetPortResult = &megaport.Port{
					UID:                "port-123abc",
					Name:               "Test Port",
					ProvisioningStatus: "CONFIGURED",
					PortSpeed:          10000,
					Type:               "MEGAPORT",
				}
			},
			expectedOutput: "port-123abc",
			outputFormat:   "table",
		},
		{
			name:    "successful status retrieval - json format",
			portUID: "port-123abc",
			setupMock: func(m *MockPortService) {
				m.GetPortResult = &megaport.Port{
					UID:                "port-123abc",
					Name:               "Test Port",
					ProvisioningStatus: "LIVE",
					PortSpeed:          1000,
					Type:               "MEGAPORT",
				}
			},
			expectedOutput: "port-123abc",
			outputFormat:   "json",
		},
		{
			name:    "Port not found",
			portUID: "port-notfound",
			setupMock: func(m *MockPortService) {
				m.GetPortErr = fmt.Errorf("Port not found")
			},
			expectedError: "error getting Port status",
			outputFormat:  "table",
		},
		{
			name:    "API error",
			portUID: "port-error",
			setupMock: func(m *MockPortService) {
				m.GetPortErr = fmt.Errorf("API error")
			},
			expectedError: "API error",
			outputFormat:  "table",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockPortService{}
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
				client := &megaport.Client{}
				client.PortService = mockService
				return client, nil
			}

			cmd := &cobra.Command{
				Use: "status [portUID]",
			}

			var err error
			capturedOutput := output.CaptureOutput(func() {
				err = GetPortStatus(cmd, []string{tt.portUID}, true, tt.outputFormat)
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

func TestGetPortCmd_WithMockClient(t *testing.T) {
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
			mockPortService := &MockPortService{}
			tt.setupMock(mockPortService)

			config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
				client := &megaport.Client{}
				client.PortService = mockPortService
				return client, nil
			}

			cmd := &cobra.Command{
				Use:  "get",
				RunE: testCommandAdapterOutput(GetPort),
			}
			cmd.Flags().String("output", "table", "Output format (table, json)")
			if err := cmd.Flags().Set("output", tt.format); err != nil {
				t.Fatalf("Failed to set output flag: %v", err)
			}

			var err error
			output := output.CaptureOutput(func() {
				cmdWithFormat := &cobra.Command{}
				cmdWithFormat.Flags().String("output", "table", "Output format")
				flagErr := cmdWithFormat.Flags().Set("output", tt.format)
				if flagErr != nil {
					t.Fatalf("Failed to set output flag: %v", err)
				}
				err = testCommandAdapterOutput(GetPort)(cmd, []string{tt.portID})
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

func TestGetPortStatus_NilPort(t *testing.T) {
	originalLoginFunc := config.LoginFunc
	defer func() {
		config.LoginFunc = originalLoginFunc
	}()

	mockService := &MockPortService{ForceNilGetPort: true}
	config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
		client := &megaport.Client{}
		client.PortService = mockService
		return client, nil
	}

	cmd := &cobra.Command{Use: "status"}
	var err error
	output.CaptureOutput(func() {
		err = GetPortStatus(cmd, []string{"port-nil"}, true, "table")
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no port found")
}

func TestBuyPort_EmptyUIDs(t *testing.T) {
	originalLoginFunc := config.LoginFunc
	originalBuyPortFunc := buyPortFunc
	defer func() {
		config.LoginFunc = originalLoginFunc
		buyPortFunc = originalBuyPortFunc
	}()

	mockService := &MockPortService{}
	config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
		client := &megaport.Client{}
		client.PortService = mockService
		return client, nil
	}
	buyPortFunc = func(ctx context.Context, client *megaport.Client, req *megaport.BuyPortRequest) (*megaport.BuyPortResponse, error) {
		return &megaport.BuyPortResponse{
			TechnicalServiceUIDs: []string{},
		}, nil
	}

	cmd := &cobra.Command{Use: "buy"}
	cmd.Flags().Bool("interactive", false, "")
	cmd.Flags().String("json", "", "")
	cmd.Flags().String("json-file", "", "")
	cmd.Flags().String("name", "test-port", "")
	cmd.Flags().Int("term", 12, "")
	cmd.Flags().Int("port-speed", 1000, "")
	cmd.Flags().Int("location-id", 1, "")
	cmd.Flags().String("marketplace-visibility", "private", "")
	cmd.Flags().String("diversity-zone", "", "")
	cmd.Flags().Bool("cost-confirm", true, "")
	_ = cmd.Flags().Set("name", "test-port")
	_ = cmd.Flags().Set("term", "12")
	_ = cmd.Flags().Set("port-speed", "1000")
	_ = cmd.Flags().Set("location-id", "1")
	_ = cmd.Flags().Set("marketplace-visibility", "private")

	var err error
	output.CaptureOutput(func() {
		err = BuyPort(cmd, nil, true)
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no UID returned")
}

func TestBuyLAGPort_EmptyUIDs(t *testing.T) {
	originalLoginFunc := config.LoginFunc
	originalBuyPortFunc := buyPortFunc
	defer func() {
		config.LoginFunc = originalLoginFunc
		buyPortFunc = originalBuyPortFunc
	}()

	mockService := &MockPortService{}
	config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
		client := &megaport.Client{}
		client.PortService = mockService
		return client, nil
	}
	buyPortFunc = func(ctx context.Context, client *megaport.Client, req *megaport.BuyPortRequest) (*megaport.BuyPortResponse, error) {
		return &megaport.BuyPortResponse{
			TechnicalServiceUIDs: []string{},
		}, nil
	}

	cmd := &cobra.Command{Use: "buy-lag"}
	cmd.Flags().Bool("interactive", false, "")
	cmd.Flags().String("json", "", "")
	cmd.Flags().String("json-file", "", "")
	cmd.Flags().String("name", "test-lag", "")
	cmd.Flags().Int("term", 12, "")
	cmd.Flags().Int("port-speed", 10000, "")
	cmd.Flags().Int("location-id", 1, "")
	cmd.Flags().Int("lag-count", 2, "")
	cmd.Flags().String("marketplace-visibility", "private", "")
	cmd.Flags().String("diversity-zone", "", "")
	cmd.Flags().Bool("cost-confirm", true, "")
	_ = cmd.Flags().Set("name", "test-lag")
	_ = cmd.Flags().Set("term", "12")
	_ = cmd.Flags().Set("port-speed", "10000")
	_ = cmd.Flags().Set("location-id", "1")
	_ = cmd.Flags().Set("lag-count", "2")
	_ = cmd.Flags().Set("marketplace-visibility", "private")

	var err error
	output.CaptureOutput(func() {
		err = BuyLAGPort(cmd, nil, true)
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no UID returned")
}

func TestDeletePort_SafeDeleteFlag(t *testing.T) {
	originalLoginFunc := config.LoginFunc
	defer func() {
		config.LoginFunc = originalLoginFunc
	}()

	mockService := &MockPortService{}
	config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
		client := &megaport.Client{}
		client.PortService = mockService
		return client, nil
	}

	cmd := &cobra.Command{
		Use: "delete",
		RunE: func(cmd *cobra.Command, args []string) error {
			return DeletePort(cmd, args, true)
		},
	}
	cmd.Flags().BoolP("force", "f", true, "")
	cmd.Flags().Bool("now", false, "")
	cmd.Flags().Bool("safe-delete", false, "")
	_ = cmd.Flags().Set("safe-delete", "true")

	var err error
	output.CaptureOutput(func() {
		err = cmd.RunE(cmd, []string{"port-uid-123"})
	})

	assert.NoError(t, err)
	assert.NotNil(t, mockService.CapturedDeletePortRequest)
	assert.True(t, mockService.CapturedDeletePortRequest.SafeDelete)
	assert.Equal(t, "port-uid-123", mockService.CapturedDeletePortRequest.PortID)
}
