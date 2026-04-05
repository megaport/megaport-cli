package product

import (
	"context"
	"fmt"
	"testing"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/commands/config"
	"github.com/megaport/megaport-cli/internal/testutil"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestListProducts(t *testing.T) {
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {})
	defer cleanup()

	tests := []struct {
		name            string
		setupMock       func(*MockProductService)
		includeInactive bool
		limit           int
		outputFormat    string
		expectedError   string
		expectedOutput  string
	}{
		{
			name: "successful list - table format",
			setupMock: func(m *MockProductService) {
				m.ListProductsResult = []megaport.Product{
					&megaport.Port{
						UID:                "port-uid-1",
						Name:               "Test Port",
						Type:               "MEGAPORT",
						ProvisioningStatus: "LIVE",
						PortSpeed:          10000,
						LocationID:         1,
					},
					&megaport.MCR{
						UID:                "mcr-uid-1",
						Name:               "Test MCR",
						Type:               "MCR2",
						ProvisioningStatus: "CONFIGURED",
						PortSpeed:          5000,
						LocationID:         2,
					},
				}
			},
			outputFormat:   "table",
			expectedOutput: "port-uid-1",
		},
		{
			name: "successful list - json format",
			setupMock: func(m *MockProductService) {
				m.ListProductsResult = []megaport.Product{
					&megaport.Port{
						UID:                "port-uid-1",
						Name:               "Test Port",
						Type:               "MEGAPORT",
						ProvisioningStatus: "LIVE",
						PortSpeed:          10000,
						LocationID:         1,
					},
				}
			},
			outputFormat:   "json",
			expectedOutput: "port-uid-1",
		},
		{
			name: "filters cancelled products by default",
			setupMock: func(m *MockProductService) {
				m.ListProductsResult = []megaport.Product{
					&megaport.Port{
						UID:                "port-active",
						Name:               "Active Port",
						Type:               "MEGAPORT",
						ProvisioningStatus: "LIVE",
						PortSpeed:          10000,
						LocationID:         1,
					},
					&megaport.Port{
						UID:                "port-cancelled",
						Name:               "Cancelled Port",
						Type:               "MEGAPORT",
						ProvisioningStatus: "CANCELLED",
						PortSpeed:          10000,
						LocationID:         1,
					},
				}
			},
			outputFormat:   "table",
			expectedOutput: "port-active",
		},
		{
			name: "includes inactive when flag set",
			setupMock: func(m *MockProductService) {
				m.ListProductsResult = []megaport.Product{
					&megaport.Port{
						UID:                "port-active",
						Name:               "Active Port",
						Type:               "MEGAPORT",
						ProvisioningStatus: "LIVE",
						PortSpeed:          10000,
						LocationID:         1,
					},
					&megaport.Port{
						UID:                "port-cancelled",
						Name:               "Cancelled Port",
						Type:               "MEGAPORT",
						ProvisioningStatus: "CANCELLED",
						PortSpeed:          10000,
						LocationID:         1,
					},
				}
			},
			includeInactive: true,
			outputFormat:    "table",
			expectedOutput:  "port-cancelled",
		},
		{
			name: "API error",
			setupMock: func(m *MockProductService) {
				m.ListProductsErr = fmt.Errorf("API error")
			},
			outputFormat:  "table",
			expectedError: "error listing products",
		},
		{
			name: "empty list",
			setupMock: func(m *MockProductService) {
				m.ListProductsResult = []megaport.Product{}
			},
			outputFormat: "table",
		},
		{
			name: "limit results",
			setupMock: func(m *MockProductService) {
				m.ListProductsResult = []megaport.Product{
					&megaport.Port{UID: "port-1", Name: "Port 1", Type: "MEGAPORT", ProvisioningStatus: "LIVE", PortSpeed: 10000, LocationID: 1},
					&megaport.Port{UID: "port-2", Name: "Port 2", Type: "MEGAPORT", ProvisioningStatus: "LIVE", PortSpeed: 10000, LocationID: 2},
					&megaport.Port{UID: "port-3", Name: "Port 3", Type: "MEGAPORT", ProvisioningStatus: "LIVE", PortSpeed: 10000, LocationID: 3},
				}
			},
			limit:          2,
			outputFormat:   "table",
			expectedOutput: "port-2",
		},
		{
			name:          "negative limit returns error",
			setupMock:     func(m *MockProductService) { m.ListProductsResult = []megaport.Product{} },
			limit:         -1,
			outputFormat:  "table",
			expectedError: "--limit must be a non-negative integer",
		},
		{
			name: "MVE product included",
			setupMock: func(m *MockProductService) {
				m.ListProductsResult = []megaport.Product{
					&megaport.MVE{
						UID:                "mve-uid-1",
						Name:               "Test MVE",
						Type:               "MVE",
						ProvisioningStatus: "LIVE",
						LocationID:         3,
					},
				}
			},
			outputFormat:   "table",
			expectedOutput: "mve-uid-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockProductService{}
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			config.SetLoginFunc(func(ctx context.Context) (*megaport.Client, error) {
				client := &megaport.Client{}
				client.ProductService = mockService
				return client, nil
			})

			cmd := &cobra.Command{Use: "list"}
			cmd.Flags().Bool("include-inactive", false, "")
			cmd.Flags().Int("limit", 0, "")
			cmd.Flags().StringP("output", "o", "table", "")

			if tt.includeInactive {
				err := cmd.Flags().Set("include-inactive", "true")
				assert.NoError(t, err)
			}
			if tt.limit != 0 {
				err := cmd.Flags().Set("limit", fmt.Sprintf("%d", tt.limit))
				assert.NoError(t, err)
			}

			var err error
			capturedOutput := output.CaptureOutput(func() {
				err = ListProducts(cmd, []string{}, true, tt.outputFormat)
			})

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				if tt.expectedOutput != "" {
					assert.Contains(t, capturedOutput, tt.expectedOutput)
				}
			}
		})
	}
}

func TestGetProductType(t *testing.T) {
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {})
	defer cleanup()

	tests := []struct {
		name           string
		productUID     string
		setupMock      func(*MockProductService)
		outputFormat   string
		expectedError  string
		expectedOutput string
	}{
		{
			name:       "successful get type - table format",
			productUID: "port-uid-1",
			setupMock: func(m *MockProductService) {
				m.GetProductTypeResult = "MEGAPORT"
			},
			outputFormat:   "table",
			expectedOutput: "MEGAPORT",
		},
		{
			name:       "successful get type - json format",
			productUID: "mcr-uid-1",
			setupMock: func(m *MockProductService) {
				m.GetProductTypeResult = "MCR2"
			},
			outputFormat:   "json",
			expectedOutput: "MCR2",
		},
		{
			name:       "product not found",
			productUID: "unknown-uid",
			setupMock: func(m *MockProductService) {
				m.GetProductTypeErr = fmt.Errorf("product not found")
			},
			outputFormat:  "table",
			expectedError: "error getting product type",
		},
		{
			name:       "API error",
			productUID: "error-uid",
			setupMock: func(m *MockProductService) {
				m.GetProductTypeErr = fmt.Errorf("API error")
			},
			outputFormat:  "table",
			expectedError: "API error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockProductService{}
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			config.SetLoginFunc(func(ctx context.Context) (*megaport.Client, error) {
				client := &megaport.Client{}
				client.ProductService = mockService
				return client, nil
			})

			cmd := &cobra.Command{Use: "get-type"}
			cmd.Flags().StringP("output", "o", "table", "")

			var err error
			capturedOutput := output.CaptureOutput(func() {
				err = GetProductType(cmd, []string{tt.productUID}, true, tt.outputFormat)
			})

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				if tt.expectedOutput != "" {
					assert.Contains(t, capturedOutput, tt.expectedOutput)
				}
			}
		})
	}
}

func TestAddCommandsTo(t *testing.T) {
	rootCmd := &cobra.Command{Use: "test-cli"}
	AddCommandsTo(rootCmd)

	productCmd, _, err := rootCmd.Find([]string{"product"})
	assert.NoError(t, err)
	assert.NotNil(t, productCmd)
	assert.Equal(t, "product", productCmd.Use)

	listCmd, _, err := rootCmd.Find([]string{"product", "list"})
	assert.NoError(t, err)
	assert.NotNil(t, listCmd)
	assert.Equal(t, "list", listCmd.Use)

	getTypeCmd, _, err := rootCmd.Find([]string{"product", "get-type"})
	assert.NoError(t, err)
	assert.NotNil(t, getTypeCmd)
	assert.Equal(t, "get-type", getTypeCmd.Use)
}

func TestProductModule(t *testing.T) {
	module := NewModule()
	assert.Equal(t, "product", module.Name())

	rootCmd := &cobra.Command{Use: "test-cli"}
	module.RegisterCommands(rootCmd)

	productCmd, _, err := rootCmd.Find([]string{"product"})
	assert.NoError(t, err)
	assert.NotNil(t, productCmd)
}

func TestListProductsLoginError(t *testing.T) {
	cleanup := testutil.SetupLoginError(fmt.Errorf("login failed"))
	defer cleanup()

	cmd := &cobra.Command{Use: "list"}
	cmd.Flags().Bool("include-inactive", false, "")
	cmd.Flags().Int("limit", 0, "")
	cmd.Flags().StringP("output", "o", "table", "")

	var err error
	output.CaptureOutput(func() {
		err = ListProducts(cmd, []string{}, true, "table")
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error logging in")
}

func TestGetProductTypeLoginError(t *testing.T) {
	cleanup := testutil.SetupLoginError(fmt.Errorf("login failed"))
	defer cleanup()

	cmd := &cobra.Command{Use: "get-type"}
	cmd.Flags().StringP("output", "o", "table", "")

	var err error
	output.CaptureOutput(func() {
		err = GetProductType(cmd, []string{"some-uid"}, true, "table")
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error logging in")
}

func TestListProductsEmptyListJsonFormat(t *testing.T) {
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {})
	defer cleanup()

	mockService := &MockProductService{
		ListProductsResult: []megaport.Product{},
	}
	config.SetLoginFunc(func(ctx context.Context) (*megaport.Client, error) {
		client := &megaport.Client{}
		client.ProductService = mockService
		return client, nil
	})

	cmd := &cobra.Command{Use: "list"}
	cmd.Flags().Bool("include-inactive", false, "")
	cmd.Flags().Int("limit", 0, "")
	cmd.Flags().StringP("output", "o", "json", "")

	var err error
	output.CaptureOutput(func() {
		err = ListProducts(cmd, []string{}, true, "json")
	})
	assert.NoError(t, err)
}

func TestFilterProductsEdgeCases(t *testing.T) {
	tests := []struct {
		name            string
		products        []megaport.Product
		includeInactive bool
		expectedCount   int
	}{
		{
			name:            "nil products in list",
			products:        []megaport.Product{nil, &megaport.Port{UID: "port-1", ProvisioningStatus: "LIVE"}},
			includeInactive: false,
			expectedCount:   1,
		},
		{
			name: "filters DECOMMISSIONED",
			products: []megaport.Product{
				&megaport.Port{UID: "port-1", ProvisioningStatus: "DECOMMISSIONED"},
				&megaport.Port{UID: "port-2", ProvisioningStatus: "LIVE"},
			},
			includeInactive: false,
			expectedCount:   1,
		},
		{
			name: "filters DECOMMISSIONING",
			products: []megaport.Product{
				&megaport.Port{UID: "port-1", ProvisioningStatus: "DECOMMISSIONING"},
				&megaport.Port{UID: "port-2", ProvisioningStatus: "LIVE"},
			},
			includeInactive: false,
			expectedCount:   1,
		},
		{
			name:            "nil input list",
			products:        nil,
			includeInactive: false,
			expectedCount:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterProducts(tt.products, tt.includeInactive)
			assert.Equal(t, tt.expectedCount, len(result))
		})
	}
}

func TestToProductOutput(t *testing.T) {
	t.Run("nil product", func(t *testing.T) {
		_, err := toProductOutput(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "nil value")
	})

	t.Run("MCR product", func(t *testing.T) {
		mcr := &megaport.MCR{
			UID:                "mcr-uid-1",
			Name:               "Test MCR",
			Type:               "MCR2",
			ProvisioningStatus: "LIVE",
			PortSpeed:          5000,
			LocationID:         2,
		}
		o, err := toProductOutput(mcr)
		assert.NoError(t, err)
		assert.Equal(t, "mcr-uid-1", o.UID)
		assert.Equal(t, "Test MCR", o.Name)
		assert.Equal(t, 5000, o.Speed)
		assert.Equal(t, 2, o.LocationID)
	})
}
