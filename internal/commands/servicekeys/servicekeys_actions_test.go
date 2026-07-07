package servicekeys

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/commands/config"
	"github.com/megaport/megaport-cli/internal/testutil"
	"github.com/megaport/megaport-cli/internal/utils"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCreateServiceKeyCmd() *cobra.Command {
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
	cmd.Flags().BoolP("interactive", "i", false, "")
	cmd.Flags().String("json", "", "")
	cmd.Flags().String("json-file", "", "")
	return cmd
}

func TestCreateServiceKey_FlagsPropagated(t *testing.T) {
	mockService := &MockServiceKeyService{}
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {
		c.ServiceKeyService = mockService
	})
	defer cleanup()

	cmd := newCreateServiceKeyCmd()

	testutil.SetFlags(t, cmd, map[string]string{
		"product-uid":  "prod-uid-123",
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
	assert.Equal(t, 0, req.ProductID)
	assert.True(t, req.SingleUse)
	assert.Equal(t, 1000, req.MaxSpeed)
	assert.Equal(t, "test key", req.Description)
	assert.True(t, req.Active)
	assert.True(t, req.PreApproved)
	assert.Equal(t, 100, req.VLAN)
}

func TestCreateServiceKey_BothProductFlagsRejected(t *testing.T) {
	mockService := &MockServiceKeyService{}
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {
		c.ServiceKeyService = mockService
	})
	defer cleanup()

	cmd := newCreateServiceKeyCmd()
	testutil.SetFlags(t, cmd, map[string]string{
		"product-uid": "prod-uid-123",
		"product-id":  "42",
	})

	var err error
	output.CaptureOutput(func() {
		err = cmd.RunE(cmd, []string{})
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot both be set")
	assert.Nil(t, mockService.CapturedCreateServiceKeyRequest)
}

func TestCreateServiceKey_NilResponse(t *testing.T) {
	mockService := &MockServiceKeyService{CreateServiceKeyReturnNil: true}
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {
		c.ServiceKeyService = mockService
	})
	defer cleanup()

	cmd := newCreateServiceKeyCmd()

	testutil.SetFlags(t, cmd, map[string]string{
		"product-uid": "prod-uid-123",
		"description": "test key",
	})

	var err error
	output.CaptureOutput(func() {
		err = cmd.RunE(cmd, []string{})
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty response from API")
}

func TestGetServiceKey_NilResponse(t *testing.T) {
	mockService := &MockServiceKeyService{GetServiceKeyReturnNil: true}
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {
		c.ServiceKeyService = mockService
	})
	defer cleanup()

	cmd := testutil.NewCommand("get", testutil.OutputAdapter(GetServiceKey))
	defer output.SetOutputFormat("table")

	var err error
	output.CaptureOutput(func() {
		err = cmd.RunE(cmd, []string{"test-key-id"})
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty response from API")
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
			cmd := newCreateServiceKeyCmd()

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

func TestCreateServiceKey_JSONMode(t *testing.T) {
	tests := []struct {
		name          string
		json          string
		useFile       bool
		expectedError string
		check         func(t *testing.T, req *megaport.CreateServiceKeyRequest)
	}{
		{
			name: "json string success",
			json: `{"productUid":"json-prod-uid","description":"JSON key","singleUse":true,"maxSpeed":500,"active":true,"preApproved":true,"vlan":50,"startDate":"2025-01-01","endDate":"2025-12-31"}`,
			check: func(t *testing.T, req *megaport.CreateServiceKeyRequest) {
				assert.Equal(t, "json-prod-uid", req.ProductUID)
				assert.Equal(t, "JSON key", req.Description)
				assert.True(t, req.SingleUse)
				assert.Equal(t, 500, req.MaxSpeed)
				assert.True(t, req.Active)
				assert.True(t, req.PreApproved)
				assert.Equal(t, 50, req.VLAN)
				if assert.NotNil(t, req.ValidFor) {
					assert.Equal(t, "2025-01-01", req.ValidFor.StartTime.Format("2006-01-02"))
					assert.Equal(t, "2025-12-31", req.ValidFor.EndTime.Format("2006-01-02"))
				}
			},
		},
		{
			name:    "json file success",
			json:    `{"productUid":"json-file-prod-uid","description":"JSON file key"}`,
			useFile: true,
			check: func(t *testing.T, req *megaport.CreateServiceKeyRequest) {
				assert.Equal(t, "json-file-prod-uid", req.ProductUID)
				assert.Equal(t, "JSON file key", req.Description)
			},
		},
		{
			name:          "invalid JSON syntax",
			json:          `{invalid}`,
			expectedError: "failed to parse JSON",
		},
		{
			name:          "product uid and product id both set",
			json:          `{"productUid":"json-prod-uid","productId":42}`,
			expectedError: "productUid and productId cannot both be set",
		},
		{
			name: "raw validFor key is ignored",
			json: `{"productUid":"json-prod-uid","validFor":{"start":111,"end":222}}`,
			check: func(t *testing.T, req *megaport.CreateServiceKeyRequest) {
				assert.Equal(t, "json-prod-uid", req.ProductUID)
				assert.Nil(t, req.ValidFor)
				assert.Nil(t, req.OrderValidFor)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockServiceKeyService{}
			cleanup := testutil.SetupLogin(func(c *megaport.Client) {
				c.ServiceKeyService = mockService
			})
			defer cleanup()

			cmd := newCreateServiceKeyCmd()

			if tt.useFile {
				tmpFile, tmpErr := os.CreateTemp("", "servicekey-create-*.json")
				require.NoError(t, tmpErr)
				defer os.Remove(tmpFile.Name())
				_, tmpErr = tmpFile.WriteString(tt.json)
				require.NoError(t, tmpErr)
				require.NoError(t, tmpFile.Close())
				testutil.SetFlags(t, cmd, map[string]string{"json-file": tmpFile.Name()})
			} else {
				testutil.SetFlags(t, cmd, map[string]string{"json": tt.json})
			}

			var err error
			output.CaptureOutput(func() {
				err = cmd.RunE(cmd, []string{})
			})

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			assert.NoError(t, err)
			if assert.NotNil(t, mockService.CapturedCreateServiceKeyRequest) {
				tt.check(t, mockService.CapturedCreateServiceKeyRequest)
			}
		})
	}
}

func TestCreateServiceKey_InteractiveMode(t *testing.T) {
	originalResourcePrompt := utils.GetResourcePrompt()
	originalConfirmPrompt := utils.GetConfirmPrompt()
	defer func() {
		utils.SetResourcePrompt(originalResourcePrompt)
		utils.SetConfirmPrompt(originalConfirmPrompt)
	}()

	mockService := &MockServiceKeyService{
		CreateServiceKeyResult: &megaport.CreateServiceKeyResponse{ServiceKeyUID: "sk-interactive"},
	}
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {
		c.ServiceKeyService = mockService
	})
	defer cleanup()

	resourcePrompts := []string{
		"prod-uid-interactive", // product UID
		"1500",                 // max speed
		"Interactive key",      // description
		"2025-01-01",           // start date
		"2025-12-31",           // end date
		"200",                  // VLAN
	}
	promptIndex := 0
	utils.SetResourcePrompt(func(_, _ string, _ bool) (string, error) {
		resp := resourcePrompts[promptIndex]
		promptIndex++
		return resp, nil
	})

	confirmResponses := []bool{true, true, false} // single-use, active, pre-approved
	confirmIndex := 0
	utils.SetConfirmPrompt(func(_ string, _ bool) bool {
		resp := confirmResponses[confirmIndex]
		confirmIndex++
		return resp
	})

	cmd := newCreateServiceKeyCmd()
	testutil.SetFlags(t, cmd, map[string]string{"interactive": "true"})

	var err error
	output.CaptureOutput(func() {
		err = cmd.RunE(cmd, []string{})
	})

	assert.NoError(t, err)
	req := mockService.CapturedCreateServiceKeyRequest
	if assert.NotNil(t, req) {
		assert.Equal(t, "prod-uid-interactive", req.ProductUID)
		assert.True(t, req.SingleUse)
		assert.Equal(t, 1500, req.MaxSpeed)
		assert.Equal(t, "Interactive key", req.Description)
		assert.True(t, req.Active)
		assert.False(t, req.PreApproved)
		assert.Equal(t, 200, req.VLAN)
		if assert.NotNil(t, req.ValidFor) {
			assert.Equal(t, "2025-01-01", req.ValidFor.StartTime.Format("2006-01-02"))
			assert.Equal(t, "2025-12-31", req.ValidFor.EndTime.Format("2006-01-02"))
		}
	}
}

func TestCreateServiceKey_InteractiveConflict(t *testing.T) {
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {})
	defer cleanup()

	cmd := newCreateServiceKeyCmd()
	testutil.SetFlags(t, cmd, map[string]string{
		"interactive": "true",
		"json":        `{"productUid":"prod-uid"}`,
	})

	var err error
	output.CaptureOutput(func() {
		err = cmd.RunE(cmd, []string{})
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be combined with")
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

func newUpdateServiceKeyCmd() *cobra.Command {
	cmd := testutil.NewCommand("update", testutil.NoColorAdapter(UpdateServiceKey))
	cmd.Flags().String("product-uid", "", "")
	cmd.Flags().Int("product-id", 0, "")
	cmd.Flags().Bool("single-use", false, "")
	cmd.Flags().Bool("active", false, "")
	cmd.Flags().BoolP("interactive", "i", false, "")
	cmd.Flags().String("json", "", "")
	cmd.Flags().String("json-file", "", "")
	return cmd
}

func TestUpdateServiceKey(t *testing.T) {
	tests := []struct {
		name        string
		mockService *MockServiceKeyService
		loginErr    error
		expectedErr string
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
			name: "IsUpdated false returns error",
			mockService: &MockServiceKeyService{
				UpdateServiceKeyResult: &megaport.UpdateServiceKeyResponse{
					IsUpdated: false,
				},
			},
			expectedErr: "service key update was not applied",
		},
		{
			name: "API error",
			mockService: &MockServiceKeyService{
				UpdateServiceKeyError: fmt.Errorf("API failure"),
			},
			expectedErr: "failed to update service key",
		},
		{
			name: "get current key error",
			mockService: &MockServiceKeyService{
				GetServiceKeyError: fmt.Errorf("key not found"),
			},
			expectedErr: "failed to fetch current service key",
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

			cmd := newUpdateServiceKeyCmd()
			testutil.SetFlags(t, cmd, map[string]string{
				"active": "true",
			})

			var err error
			output.CaptureOutput(func() {
				err = cmd.RunE(cmd, []string{"test-key-123"})
			})

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateServiceKey_NilResponse(t *testing.T) {
	mockService := &MockServiceKeyService{GetServiceKeyReturnNil: true}
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {
		c.ServiceKeyService = mockService
	})
	defer cleanup()

	cmd := newUpdateServiceKeyCmd()

	var err error
	output.CaptureOutput(func() {
		err = cmd.RunE(cmd, []string{"test-key-123"})
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty response from API")
	assert.Nil(t, mockService.CapturedUpdateServiceKeyRequest)
}

func TestUpdateServiceKey_NilUpdateResponse(t *testing.T) {
	mockService := &MockServiceKeyService{UpdateServiceKeyReturnNil: true}
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {
		c.ServiceKeyService = mockService
	})
	defer cleanup()

	cmd := newUpdateServiceKeyCmd()
	testutil.SetFlags(t, cmd, map[string]string{"active": "true"})

	var err error
	output.CaptureOutput(func() {
		err = cmd.RunE(cmd, []string{"test-key-123"})
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty response from API")
}

func TestUpdateServiceKey_BothProductFlagsRejected(t *testing.T) {
	mockService := &MockServiceKeyService{}
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {
		c.ServiceKeyService = mockService
	})
	defer cleanup()

	cmd := newUpdateServiceKeyCmd()
	testutil.SetFlags(t, cmd, map[string]string{
		"product-uid": "prod-uid",
		"product-id":  "42",
	})

	var err error
	output.CaptureOutput(func() {
		err = cmd.RunE(cmd, []string{"test-key-123"})
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot both be set")
	assert.Nil(t, mockService.CapturedUpdateServiceKeyRequest)
}

func TestUpdateServiceKey_MergesUnsetFlags(t *testing.T) {
	currentKey := &megaport.ServiceKey{
		Key:        "test-key-123",
		ProductUID: "current-prod-uid",
		SingleUse:  true,
		Active:     true,
	}

	tests := []struct {
		name              string
		flags             map[string]string
		expectedSingleUse bool
		expectedActive    bool
		expectedUID       string
		expectedID        int
	}{
		{
			name:              "no flags preserves current values",
			flags:             map[string]string{},
			expectedSingleUse: true,
			expectedActive:    true,
			expectedUID:       "current-prod-uid",
		},
		{
			name:              "product-uid only preserves bools",
			flags:             map[string]string{"product-uid": "new-prod-uid"},
			expectedSingleUse: true,
			expectedActive:    true,
			expectedUID:       "new-prod-uid",
		},
		{
			name:              "explicit active=false is honored",
			flags:             map[string]string{"active": "false"},
			expectedSingleUse: true,
			expectedActive:    false,
			expectedUID:       "current-prod-uid",
		},
		{
			name:              "explicit single-use=false is honored",
			flags:             map[string]string{"single-use": "false"},
			expectedSingleUse: false,
			expectedActive:    true,
			expectedUID:       "current-prod-uid",
		},
		{
			name:              "product-id replaces current product-uid",
			flags:             map[string]string{"product-id": "42"},
			expectedSingleUse: true,
			expectedActive:    true,
			expectedUID:       "",
			expectedID:        42,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockServiceKeyService{
				GetServiceKeyResult: currentKey,
			}
			cleanup := testutil.SetupLogin(func(c *megaport.Client) {
				c.ServiceKeyService = mockService
			})
			defer cleanup()

			cmd := newUpdateServiceKeyCmd()
			testutil.SetFlags(t, cmd, tt.flags)

			var err error
			output.CaptureOutput(func() {
				err = cmd.RunE(cmd, []string{"test-key-123"})
			})

			assert.NoError(t, err)
			req := mockService.CapturedUpdateServiceKeyRequest
			if assert.NotNil(t, req) {
				assert.Equal(t, "test-key-123", req.Key)
				assert.Equal(t, tt.expectedSingleUse, req.SingleUse)
				assert.Equal(t, tt.expectedActive, req.Active)
				assert.Equal(t, tt.expectedUID, req.ProductUID)
				assert.Equal(t, tt.expectedID, req.ProductID)
			}
		})
	}
}

func TestUpdateServiceKey_JSONMode(t *testing.T) {
	currentKey := &megaport.ServiceKey{
		Key:        "test-key-123",
		ProductUID: "current-prod-uid",
		SingleUse:  true,
		Active:     true,
	}

	tests := []struct {
		name              string
		json              string
		expectedError     string
		expectedSingleUse bool
		expectedActive    bool
		expectedUID       string
		expectedID        int
	}{
		{
			name:              "empty object preserves current values",
			json:              `{}`,
			expectedSingleUse: true,
			expectedActive:    true,
			expectedUID:       "current-prod-uid",
		},
		{
			name:              "product uid only preserves bools",
			json:              `{"productUid":"new-prod-uid"}`,
			expectedSingleUse: true,
			expectedActive:    true,
			expectedUID:       "new-prod-uid",
		},
		{
			name:              "explicit active false is honored",
			json:              `{"active":false}`,
			expectedSingleUse: true,
			expectedActive:    false,
			expectedUID:       "current-prod-uid",
		},
		{
			name:              "explicit single-use false is honored",
			json:              `{"singleUse":false}`,
			expectedSingleUse: false,
			expectedActive:    true,
			expectedUID:       "current-prod-uid",
		},
		{
			name:              "product id replaces current product uid",
			json:              `{"productId":42}`,
			expectedSingleUse: true,
			expectedActive:    true,
			expectedUID:       "",
			expectedID:        42,
		},
		{
			name:          "product uid and product id both set",
			json:          `{"productUid":"new-prod-uid","productId":42}`,
			expectedError: "productUid and productId cannot both be set",
		},
		{
			name:          "invalid JSON syntax",
			json:          `{invalid}`,
			expectedError: "failed to parse JSON",
		},
		{
			name:              "raw validFor key is ignored",
			json:              `{"active":false,"validFor":{"start":111,"end":222}}`,
			expectedSingleUse: true,
			expectedActive:    false,
			expectedUID:       "current-prod-uid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockServiceKeyService{
				GetServiceKeyResult: currentKey,
			}
			cleanup := testutil.SetupLogin(func(c *megaport.Client) {
				c.ServiceKeyService = mockService
			})
			defer cleanup()

			cmd := newUpdateServiceKeyCmd()
			testutil.SetFlags(t, cmd, map[string]string{"json": tt.json})

			var err error
			output.CaptureOutput(func() {
				err = cmd.RunE(cmd, []string{"test-key-123"})
			})

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			assert.NoError(t, err)
			req := mockService.CapturedUpdateServiceKeyRequest
			if assert.NotNil(t, req) {
				assert.Equal(t, "test-key-123", req.Key)
				assert.Equal(t, tt.expectedSingleUse, req.SingleUse)
				assert.Equal(t, tt.expectedActive, req.Active)
				assert.Equal(t, tt.expectedUID, req.ProductUID)
				assert.Equal(t, tt.expectedID, req.ProductID)
				assert.Nil(t, req.ValidFor)
				assert.Nil(t, req.OrderValidFor)
			}
		})
	}
}

func TestUpdateServiceKey_JSONModeIgnoresProductFlagConflict(t *testing.T) {
	currentKey := &megaport.ServiceKey{
		Key:        "test-key-123",
		ProductUID: "current-prod-uid",
		SingleUse:  true,
		Active:     true,
	}
	mockService := &MockServiceKeyService{GetServiceKeyResult: currentKey}
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {
		c.ServiceKeyService = mockService
	})
	defer cleanup()

	cmd := newUpdateServiceKeyCmd()
	testutil.SetFlags(t, cmd, map[string]string{
		"json":        `{"productUid":"json-prod-uid"}`,
		"product-uid": "flag-prod-uid",
		"product-id":  "42",
	})

	var err error
	output.CaptureOutput(func() {
		err = cmd.RunE(cmd, []string{"test-key-123"})
	})

	assert.NoError(t, err)
	req := mockService.CapturedUpdateServiceKeyRequest
	if assert.NotNil(t, req) {
		assert.Equal(t, "json-prod-uid", req.ProductUID)
	}
}

func TestUpdateServiceKey_InteractiveMode(t *testing.T) {
	currentKey := &megaport.ServiceKey{
		Key:        "test-key-123",
		ProductUID: "current-prod-uid",
		SingleUse:  true,
		Active:     true,
	}

	tests := []struct {
		name              string
		resourcePrompts   []string
		confirmResponses  []bool
		expectedSingleUse bool
		expectedActive    bool
		expectedUID       string
		expectedID        int
	}{
		{
			name:              "skipping every prompt preserves current values",
			resourcePrompts:   []string{"", "", "no", "no"},
			expectedSingleUse: true,
			expectedActive:    true,
			expectedUID:       "current-prod-uid",
		},
		{
			name:              "new product uid preserves bools",
			resourcePrompts:   []string{"new-prod-uid", "no", "no"},
			expectedSingleUse: true,
			expectedActive:    true,
			expectedUID:       "new-prod-uid",
		},
		{
			name:              "new product id replaces current product uid",
			resourcePrompts:   []string{"", "42", "no", "no"},
			expectedSingleUse: true,
			expectedActive:    true,
			expectedUID:       "",
			expectedID:        42,
		},
		{
			name:              "explicit active false is honored",
			resourcePrompts:   []string{"", "", "no", "yes"},
			confirmResponses:  []bool{false},
			expectedSingleUse: true,
			expectedActive:    false,
			expectedUID:       "current-prod-uid",
		},
		{
			name:              "explicit single-use false is honored",
			resourcePrompts:   []string{"", "", "yes", "no"},
			confirmResponses:  []bool{false},
			expectedSingleUse: false,
			expectedActive:    true,
			expectedUID:       "current-prod-uid",
		},
	}

	originalResourcePrompt := utils.GetResourcePrompt()
	originalConfirmPrompt := utils.GetConfirmPrompt()
	defer func() {
		utils.SetResourcePrompt(originalResourcePrompt)
		utils.SetConfirmPrompt(originalConfirmPrompt)
	}()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockServiceKeyService{
				GetServiceKeyResult: currentKey,
			}
			cleanup := testutil.SetupLogin(func(c *megaport.Client) {
				c.ServiceKeyService = mockService
			})
			defer cleanup()

			promptIndex := 0
			utils.SetResourcePrompt(func(_, _ string, _ bool) (string, error) {
				resp := tt.resourcePrompts[promptIndex]
				promptIndex++
				return resp, nil
			})

			confirmIndex := 0
			utils.SetConfirmPrompt(func(_ string, _ bool) bool {
				resp := tt.confirmResponses[confirmIndex]
				confirmIndex++
				return resp
			})

			cmd := newUpdateServiceKeyCmd()
			testutil.SetFlags(t, cmd, map[string]string{"interactive": "true"})

			var err error
			output.CaptureOutput(func() {
				err = cmd.RunE(cmd, []string{"test-key-123"})
			})

			assert.NoError(t, err)
			req := mockService.CapturedUpdateServiceKeyRequest
			if assert.NotNil(t, req) {
				assert.Equal(t, "test-key-123", req.Key)
				assert.Equal(t, tt.expectedSingleUse, req.SingleUse)
				assert.Equal(t, tt.expectedActive, req.Active)
				assert.Equal(t, tt.expectedUID, req.ProductUID)
				assert.Equal(t, tt.expectedID, req.ProductID)
			}
		})
	}
}

func TestUpdateServiceKey_InteractiveConflict(t *testing.T) {
	mockService := &MockServiceKeyService{}
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {
		c.ServiceKeyService = mockService
	})
	defer cleanup()

	cmd := newUpdateServiceKeyCmd()
	testutil.SetFlags(t, cmd, map[string]string{
		"interactive": "true",
		"json":        `{"active":false}`,
	})

	var err error
	output.CaptureOutput(func() {
		err = cmd.RunE(cmd, []string{"test-key-123"})
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be combined with")
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
			name:           "json format returns empty array without message",
			outputFormat:   "json",
			expectedOutput: "[]",
			notExpected:    "No service keys found.",
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

			var capturedErr string
			capturedOutput := output.CaptureOutput(func() {
				capturedErr = captureStderr(t, func() {
					_ = cmd.RunE(cmd, []string{})
				})
			})

			if tt.expectedOutput != "" {
				assert.Contains(t, capturedOutput+capturedErr, tt.expectedOutput)
			}
			if tt.notExpected != "" {
				assert.NotContains(t, capturedOutput+capturedErr, tt.notExpected)
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

func captureStderr(t *testing.T, fn func()) (result string) {
	t.Helper()
	old := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = w
	defer func() { os.Stderr = old }()
	var buf bytes.Buffer
	done := make(chan struct{})
	go func() { defer close(done); _, _ = io.Copy(&buf, r) }()
	defer func() { _ = w.Close(); <-done; _ = r.Close(); result = buf.String() }()
	fn()
	return
}

func TestListServiceKeys_NilResponse(t *testing.T) {
	mockService := &MockServiceKeyService{ListServiceKeysReturnNil: true}
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {
		c.ServiceKeyService = mockService
	})
	defer cleanup()

	cmd := testutil.NewCommand("list", testutil.OutputAdapter(ListServiceKeys))
	cmd.Flags().String("product-uid", "", "")
	cmd.Flags().Int("limit", 0, "")
	defer output.SetOutputFormat("table")

	var err error
	output.CaptureOutput(func() {
		err = cmd.RunE(cmd, []string{})
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty response from API")
}
