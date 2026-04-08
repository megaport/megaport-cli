package servicekeys

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/commands/config"
	"github.com/megaport/megaport-cli/internal/testutil"
	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
)

func TestCreateServiceKey_FlagsPropagated(t *testing.T) {
	mockService := &MockServiceKeyService{}
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {
		c.ServiceKeyService = mockService
	})
	defer cleanup()

	cmd := testutil.NewCommand("create", testutil.NoColorAdapter(CreateServiceKey))
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

	testutil.SetFlags(t, cmd, map[string]string{
		"product-uid":  "prod-uid-123",
		"product-id":   "42",
		"single-use":   "true",
		"max-speed":    "1000",
		"description":  "test key",
		"active":       "true",
		"pre-approved": "true",
		"vlan":         "100",
	})

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

func TestCreateServiceKey_InvalidDateParsing(t *testing.T) {
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {})
	defer cleanup()

	tests := []struct {
		name          string
		startDate     string
		endDate       string
		expectedError string
	}{
		{
			name:          "invalid start date",
			startDate:     "not-a-date",
			endDate:       "2025-12-31",
			expectedError: "Invalid start-date",
		},
		{
			name:          "invalid end date",
			startDate:     "2025-01-01",
			endDate:       "not-a-date",
			expectedError: "Invalid end-date",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := testutil.NewCommand("create", testutil.NoColorAdapter(CreateServiceKey))
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

			testutil.SetFlags(t, cmd, map[string]string{
				"product-uid": "prod-123",
				"start-date":  tt.startDate,
				"end-date":    tt.endDate,
			})

			var err error
			output.CaptureOutput(func() {
				err = cmd.RunE(cmd, []string{})
			})

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

func TestListServiceKeys_ProductUIDFilter(t *testing.T) {
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {})
	defer cleanup()

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
			config.SetLoginFunc(func(ctx context.Context) (*megaport.Client, error) {
				client := &megaport.Client{}
				client.ServiceKeyService = mockService
				return client, nil
			})

			cmd := testutil.NewCommand("list", testutil.OutputAdapter(ListServiceKeys))
			cmd.Flags().String("product-uid", "", "")

			if tt.setProductUID {
				testutil.SetFlags(t, cmd, map[string]string{
					"product-uid": tt.productUID,
				})
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
			expectedErr: "failed to update service key",
		},
		{
			name:        "login error",
			mockService: &MockServiceKeyService{},
			loginErr:    fmt.Errorf("login failure"),
			expectedErr: "login failure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := testutil.SetupLogin(func(c *megaport.Client) {})
			defer cleanup()

			config.SetLoginFunc(func(ctx context.Context) (*megaport.Client, error) {
				if tt.loginErr != nil {
					return nil, tt.loginErr
				}
				client := &megaport.Client{}
				client.ServiceKeyService = tt.mockService
				return client, nil
			})

			cmd := testutil.NewCommand("update", testutil.NoColorAdapter(UpdateServiceKey))
			cmd.Flags().String("product-uid", "", "")
			cmd.Flags().Int("product-id", 0, "")
			cmd.Flags().Bool("single-use", false, "")
			cmd.Flags().Bool("active", false, "")

			testutil.SetFlags(t, cmd, map[string]string{
				"active": "true",
			})

			var err error
			capturedOutput := output.CaptureOutput(func() {
				err = cmd.RunE(cmd, []string{"test-key-123"})
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
			expectedErr:  "failed to get service key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := testutil.SetupLogin(func(c *megaport.Client) {
				c.ServiceKeyService = tt.mockService
			})
			defer cleanup()

			cmd := testutil.NewCommand("get", testutil.OutputAdapter(GetServiceKey))
			testutil.SetFlags(t, cmd, map[string]string{
				"output": tt.outputFormat,
			})

			defer output.SetOutputFormat("table")

			var err error
			capturedOutput := output.CaptureOutput(func() {
				err = cmd.RunE(cmd, []string{"test-key-id"})
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

func TestListServiceKeys_EmptyResult(t *testing.T) {
	tests := []struct {
		name           string
		outputFormat   string
		expectedOutput string
		notExpected    string
	}{
		{
			name:           "table format shows info message",
			outputFormat:   "table",
			expectedOutput: "No service keys found.",
		},
		{
			name:         "json format returns empty array without message",
			outputFormat: "json",
			notExpected:  "No service keys found.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockServiceKeyService{
				ListServiceKeysResult: &megaport.ListServiceKeysResponse{
					ServiceKeys: []*megaport.ServiceKey{},
				},
			}
			config.SetLoginFunc(func(ctx context.Context) (*megaport.Client, error) {
				client := &megaport.Client{}
				client.ServiceKeyService = mockService
				return client, nil
			})

			cmd := testutil.NewCommand("list", testutil.OutputAdapter(ListServiceKeys))
			cmd.Flags().String("product-uid", "", "")
			if tt.outputFormat != "" && tt.outputFormat != "table" {
				testutil.SetFlags(t, cmd, map[string]string{"output": tt.outputFormat})
			}

			capturedOutput := output.CaptureOutput(func() {
				_ = cmd.RunE(cmd, []string{})
			})

			if tt.expectedOutput != "" {
				assert.Contains(t, capturedOutput, tt.expectedOutput)
			}
			if tt.notExpected != "" {
				assert.NotContains(t, capturedOutput, tt.notExpected)
			}
		})
	}
}

func TestListServiceKeys_Limit(t *testing.T) {
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {})
	defer cleanup()

	mockService := &MockServiceKeyService{
		ListServiceKeysResult: &megaport.ListServiceKeysResponse{
			ServiceKeys: []*megaport.ServiceKey{
				{Key: "key-1", ProductName: "Product A", ProductUID: "prod-1", Description: "First key"},
				{Key: "key-2", ProductName: "Product B", ProductUID: "prod-2", Description: "Second key"},
				{Key: "key-3", ProductName: "Product C", ProductUID: "prod-3", Description: "Third key"},
			},
		},
	}
	config.SetLoginFunc(func(ctx context.Context) (*megaport.Client, error) {
		client := &megaport.Client{}
		client.ServiceKeyService = mockService
		return client, nil
	})

	cmd := testutil.NewCommand("list", testutil.OutputAdapter(ListServiceKeys))
	cmd.Flags().String("product-uid", "", "")
	cmd.Flags().Int("limit", 0, "")
	testutil.SetFlags(t, cmd, map[string]string{"limit": "2"})

	capturedOutput := output.CaptureOutput(func() {
		err := cmd.RunE(cmd, []string{})
		assert.NoError(t, err)
	})

	assert.Contains(t, capturedOutput, "key-1")
	assert.Contains(t, capturedOutput, "key-2")
	assert.NotContains(t, capturedOutput, "key-3")
}

func TestListServiceKeys_NegativeLimit(t *testing.T) {
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {})
	defer cleanup()

	mockService := &MockServiceKeyService{
		ListServiceKeysResult: &megaport.ListServiceKeysResponse{
			ServiceKeys: []*megaport.ServiceKey{
				{Key: "key-1", ProductName: "Product A", ProductUID: "prod-1"},
			},
		},
	}
	config.SetLoginFunc(func(ctx context.Context) (*megaport.Client, error) {
		client := &megaport.Client{}
		client.ServiceKeyService = mockService
		return client, nil
	})

	cmd := testutil.NewCommand("list", testutil.OutputAdapter(ListServiceKeys))
	cmd.Flags().String("product-uid", "", "")
	cmd.Flags().Int("limit", 0, "")
	testutil.SetFlags(t, cmd, map[string]string{"limit": "-1"})

	err := cmd.RunE(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--limit must be a non-negative integer")
}
