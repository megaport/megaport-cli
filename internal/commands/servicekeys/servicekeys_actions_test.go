package servicekeys

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/commands/config"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestCreateServiceKey_FlagsPropagated(t *testing.T) {
	originalLoginFunc := config.LoginFunc
	defer func() {
		config.LoginFunc = originalLoginFunc
	}()

	mockService := &MockServiceKeyService{}
	config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
		client := &megaport.Client{}
		client.ServiceKeyService = mockService
		return client, nil
	}

	cmd := &cobra.Command{
		Use: "create",
		RunE: func(cmd *cobra.Command, args []string) error {
			return CreateServiceKey(cmd, args, true)
		},
	}
	cmd.Flags().String("product-uid", "", "")
	cmd.Flags().Int("product-id", 0, "")
	cmd.Flags().Bool("single-use", false, "")
	cmd.Flags().Int("max-speed", 0, "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("start-date", "", "")
	cmd.Flags().String("end-date", "", "")
	cmd.Flags().Bool("active", false, "")
	cmd.Flags().Bool("pre-approved", false, "")
	cmd.Flags().Int("vlan", 0, "")

	_ = cmd.Flags().Set("product-uid", "prod-uid-123")
	_ = cmd.Flags().Set("product-id", "42")
	_ = cmd.Flags().Set("single-use", "true")
	_ = cmd.Flags().Set("max-speed", "1000")
	_ = cmd.Flags().Set("description", "test key")
	_ = cmd.Flags().Set("active", "true")
	_ = cmd.Flags().Set("pre-approved", "true")
	_ = cmd.Flags().Set("vlan", "100")

	var err error
	output.CaptureOutput(func() {
		err = cmd.RunE(cmd, []string{})
	})

	assert.NoError(t, err)
	assert.NotNil(t, mockService.CapturedCreateServiceKeyRequest)

	req := mockService.CapturedCreateServiceKeyRequest
	assert.Equal(t, "prod-uid-123", req.ProductUID)
	assert.Equal(t, 42, req.ProductID)
	assert.True(t, req.SingleUse)
	assert.Equal(t, 1000, req.MaxSpeed)
	assert.Equal(t, "test key", req.Description)
	assert.True(t, req.Active)
	assert.True(t, req.PreApproved)
	assert.Equal(t, 100, req.VLAN)
}

func TestListServiceKeys_ProductUIDFilter(t *testing.T) {
	originalLoginFunc := config.LoginFunc
	defer func() {
		config.LoginFunc = originalLoginFunc
	}()

	tests := []struct {
		name              string
		setProductUID     bool
		productUID        string
		expectFilterSet   bool
		expectedFilterVal string
	}{
		{
			name:              "with product-uid filter",
			setProductUID:     true,
			productUID:        "prod-uid-456",
			expectFilterSet:   true,
			expectedFilterVal: "prod-uid-456",
		},
		{
			name:            "without product-uid filter",
			setProductUID:   false,
			expectFilterSet: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockServiceKeyService{}
			config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
				client := &megaport.Client{}
				client.ServiceKeyService = mockService
				return client, nil
			}

			cmd := &cobra.Command{
				Use: "list",
				RunE: func(cmd *cobra.Command, args []string) error {
					return ListServiceKeys(cmd, args, true, "table")
				},
			}
			cmd.Flags().String("product-uid", "", "")

			if tt.setProductUID {
				_ = cmd.Flags().Set("product-uid", tt.productUID)
			}

			var err error
			output.CaptureOutput(func() {
				err = cmd.RunE(cmd, []string{})
			})

			assert.NoError(t, err)
			assert.NotNil(t, mockService.CapturedListServiceKeysRequest)

			req := mockService.CapturedListServiceKeysRequest
			if tt.expectFilterSet {
				assert.NotNil(t, req.ProductUID)
				assert.Equal(t, tt.expectedFilterVal, *req.ProductUID)
			} else {
				assert.Nil(t, req.ProductUID)
			}
		})
	}
}

func TestUpdateServiceKey(t *testing.T) {
	tests := []struct {
		name        string
		mockService *MockServiceKeyService
		loginErr    error
		expectedErr string
		expectWarn  bool
	}{
		{
			name: "success - IsUpdated true",
			mockService: &MockServiceKeyService{
				UpdateServiceKeyResult: &megaport.UpdateServiceKeyResponse{
					IsUpdated: true,
				},
			},
		},
		{
			name: "IsUpdated false",
			mockService: &MockServiceKeyService{
				UpdateServiceKeyResult: &megaport.UpdateServiceKeyResponse{
					IsUpdated: false,
				},
			},
			expectWarn: true,
		},
		{
			name: "API error",
			mockService: &MockServiceKeyService{
				UpdateServiceKeyError: fmt.Errorf("API failure"),
			},
			expectedErr: "error updating service key",
		},
		{
			name:        "login error",
			mockService: &MockServiceKeyService{},
			loginErr:    fmt.Errorf("login failure"),
			expectedErr: "error logging in",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalLoginFunc := config.LoginFunc
			defer func() {
				config.LoginFunc = originalLoginFunc
			}()

			config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
				if tt.loginErr != nil {
					return nil, tt.loginErr
				}
				client := &megaport.Client{}
				client.ServiceKeyService = tt.mockService
				return client, nil
			}

			cmd := &cobra.Command{
				Use: "update",
			}
			cmd.Flags().String("key", "", "")
			cmd.Flags().String("product-uid", "", "")
			cmd.Flags().Int("product-id", 0, "")
			cmd.Flags().Bool("single-use", false, "")
			cmd.Flags().Bool("active", false, "")

			_ = cmd.Flags().Set("key", "test-key-123")
			_ = cmd.Flags().Set("active", "true")

			var err error
			capturedOutput := output.CaptureOutput(func() {
				err = UpdateServiceKey(cmd, []string{}, true)
			})

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)
				if tt.expectWarn {
					assert.Contains(t, capturedOutput, "update request was not successful")
				}
			}
		})
	}
}

func TestGetServiceKey(t *testing.T) {
	tests := []struct {
		name         string
		mockService  *MockServiceKeyService
		outputFormat string
		expectedErr  string
		checkOutput  func(t *testing.T, capturedOutput string)
	}{
		{
			name: "success table format",
			mockService: &MockServiceKeyService{
				GetServiceKeyResult: &megaport.ServiceKey{
					Key:         "sk-123",
					Description: "Test Key",
					ProductUID:  "prod-uid-1",
					ProductName: "Test Product",
				},
			},
			outputFormat: "table",
			checkOutput: func(t *testing.T, capturedOutput string) {
				assert.Contains(t, capturedOutput, "sk-123")
			},
		},
		{
			name: "success JSON format",
			mockService: &MockServiceKeyService{
				GetServiceKeyResult: &megaport.ServiceKey{
					Key:         "sk-456",
					Description: "JSON Key",
					ProductUID:  "prod-uid-2",
					ProductName: "JSON Product",
				},
			},
			outputFormat: "json",
			checkOutput: func(t *testing.T, capturedOutput string) {
				var parsed []map[string]interface{}
				assert.NoError(t, json.Unmarshal([]byte(capturedOutput), &parsed), "JSON output should be valid JSON")
				if assert.NotEmpty(t, parsed) {
					assert.Equal(t, "sk-456", parsed[0]["key_uid"])
				}
			},
		},
		{
			name: "API error",
			mockService: &MockServiceKeyService{
				GetServiceKeyError: fmt.Errorf("service key not found"),
			},
			outputFormat: "table",
			expectedErr:  "error getting service key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalLoginFunc := config.LoginFunc
			defer func() {
				config.LoginFunc = originalLoginFunc
			}()

			config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
				client := &megaport.Client{}
				client.ServiceKeyService = tt.mockService
				return client, nil
			}

			cmd := &cobra.Command{
				Use: "get",
			}

			output.SetOutputFormat(tt.outputFormat)
			defer output.SetOutputFormat("")

			var err error
			capturedOutput := output.CaptureOutput(func() {
				err = GetServiceKey(cmd, []string{"test-key-id"}, true, tt.outputFormat)
			})

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)
				if tt.checkOutput != nil {
					tt.checkOutput(t, capturedOutput)
				}
			}
		})
	}
}
