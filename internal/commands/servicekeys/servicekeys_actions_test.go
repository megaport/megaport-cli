package servicekeys

import (
	"context"
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
