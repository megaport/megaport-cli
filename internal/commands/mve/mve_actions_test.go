package mve

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/commands/config"
	"github.com/megaport/megaport-cli/internal/utils"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

var noColor = true // Disable color for testing

var testMVEImages = []*megaport.MVEImage{
	{
		ID:                1,
		Version:           "1.0",
		Product:           "Product1",
		Vendor:            "Cisco",
		VendorDescription: "Cisco Description",
		ReleaseImage:      true,
		ProductCode:       "CISCO123",
	},
	{
		ID:                2,
		Version:           "2.0",
		Product:           "Product2",
		Vendor:            "Fortinet",
		VendorDescription: "Fortinet Description",
		ReleaseImage:      false,
		ProductCode:       "FORTINET456",
	},
}

var testMVESizes = []*megaport.MVESize{
	{
		Size:         "small",
		Label:        "Small",
		CPUCoreCount: 2,
		RamGB:        8,
	},
	{
		Size:         "large",
		Label:        "Large",
		CPUCoreCount: 8,
		RamGB:        32,
	},
}

func TestFilterMVEImages(t *testing.T) {
	tests := []struct {
		name          string
		vendor        string
		productCode   string
		id            int
		version       string
		releaseImage  bool
		expectedCount int
	}{
		{
			name:          "filter by vendor",
			vendor:        "Cisco",
			expectedCount: 1,
		},
		{
			name:          "filter by product code",
			productCode:   "FORTINET456",
			expectedCount: 1,
		},
		{
			name:          "filter by ID",
			id:            1,
			expectedCount: 1,
		},
		{
			name:          "filter by version",
			version:       "2.0",
			expectedCount: 1,
		},
		{
			name:          "filter by release image",
			releaseImage:  true,
			expectedCount: 1,
		},
		{
			name:          "no filters",
			expectedCount: 2,
		},
		{
			name:          "no matches",
			vendor:        "NonExistentVendor",
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filteredImages := filterMVEImages(testMVEImages, tt.vendor, tt.productCode, tt.id, tt.version, tt.releaseImage)
			assert.Equal(t, tt.expectedCount, len(filteredImages))
		})
	}
}
func TestListMVEImages(t *testing.T) {
	originalLoginFunc := config.LoginFunc
	defer func() {
		config.LoginFunc = originalLoginFunc
	}()

	mockService := &MockMVEService{
		ListMVEImagesResult: testMVEImages,
	}

	config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
		client := &megaport.Client{
			MVEService: mockService,
		}
		return client, nil
	}

	tests := []struct {
		name          string
		args          []string
		expectedCount int
		checkFor      string
	}{
		{
			name:          "no filters",
			args:          []string{},
			expectedCount: 2,
			checkFor:      "Product1",
		},
		{
			name:          "filter by vendor",
			args:          []string{"--vendor", "Cisco"},
			expectedCount: 1,
			checkFor:      "Product1",
		},
		{
			name:          "filter by product code",
			args:          []string{"--product-code", "FORTINET456"},
			expectedCount: 1,
			checkFor:      "Product2",
		},
		{
			name:          "filter by ID",
			args:          []string{"--id", "1"},
			expectedCount: 1,
			checkFor:      "Product1",
		},
		{
			name:          "filter by version",
			args:          []string{"--version", "2.0"},
			expectedCount: 1,
			checkFor:      "Product2",
		},
		{
			name:          "filter by release image",
			args:          []string{"--release-image"},
			expectedCount: 1,
			checkFor:      "Product1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test-specific command structure
			rootCmd := &cobra.Command{Use: "megaport"}
			mveCmd := &cobra.Command{Use: "mve"}

			listImagesCmd := &cobra.Command{
				Use: "list-images",
				RunE: func(cmd *cobra.Command, args []string) error {
					ctx := context.Background()
					client, err := config.LoginFunc(ctx)
					if err != nil {
						return err
					}

					// Get parameters from flags
					vendor, _ := cmd.Flags().GetString("vendor")
					productCode, _ := cmd.Flags().GetString("product-code")
					id, _ := cmd.Flags().GetInt("id")
					version, _ := cmd.Flags().GetString("version")
					releaseImage, _ := cmd.Flags().GetBool("release-image")

					// Fetch images from mock service (provided by loginFunc)
					images, err := client.MVEService.ListMVEImages(ctx)
					if err != nil {
						return err
					}

					// Filter images based on flags
					filteredImages := filterMVEImages(images, vendor, productCode, id, version, releaseImage)

					// Print the images as a table
					for _, img := range filteredImages {
						fmt.Printf("%d    %s       %s    %s     %s      %t           %s\n",
							img.ID, img.Version, img.Product, img.Vendor, img.VendorDescription, img.ReleaseImage, img.ProductCode)
					}

					return nil
				},
			}

			// Add flags to the command
			listImagesCmd.Flags().String("vendor", "", "Filter by vendor")
			listImagesCmd.Flags().String("product-code", "", "Filter by product code")
			listImagesCmd.Flags().Int("id", 0, "Filter by ID")
			listImagesCmd.Flags().String("version", "", "Filter by version")
			listImagesCmd.Flags().Bool("release-image", false, "Filter by release image")

			// Build command hierarchy
			mveCmd.AddCommand(listImagesCmd)
			rootCmd.AddCommand(mveCmd)

			// Set arguments for this test case
			rootCmd.SetArgs(append([]string{"mve", "list-images"}, tt.args...))

			// Capture and check the output
			output, err := output.CaptureOutputErr(func() error {
				return rootCmd.Execute()
			})

			assert.NoError(t, err)
			assert.Contains(t, output, tt.checkFor)
			productCount := strings.Count(output, "Product")
			assert.Equal(t, tt.expectedCount, productCount, "Expected %d products, but got %d", tt.expectedCount, productCount)
		})
	}
}

func TestListAvailableMVESizes(t *testing.T) {
	originalLoginFunc := config.LoginFunc
	defer func() {
		config.LoginFunc = originalLoginFunc
	}()

	mockService := &MockMVEService{
		ListAvailableMVESizesResult: testMVESizes,
	}

	config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
		client := &megaport.Client{
			MVEService: mockService,
		}
		return client, nil
	}

	// Create test-specific command structure
	rootCmd := &cobra.Command{Use: "megaport"}
	mveCmd := &cobra.Command{Use: "mve"}

	listSizesCmd := &cobra.Command{
		Use: "list-sizes",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, err := config.LoginFunc(ctx)
			if err != nil {
				return err
			}

			// Fetch sizes from mock service
			sizes, err := client.MVEService.ListAvailableMVESizes(ctx)
			if err != nil {
				return err
			}

			// Print size information
			for _, size := range sizes {
				fmt.Printf("%s    %s    %d    %d\n", size.Size, size.Label, size.CPUCoreCount, size.RamGB)
			}

			return nil
		},
	}

	// Build command hierarchy
	mveCmd.AddCommand(listSizesCmd)
	rootCmd.AddCommand(mveCmd)

	// Set arguments
	rootCmd.SetArgs([]string{"mve", "list-sizes"})

	// Capture and check the output
	output, err := output.CaptureOutputErr(func() error {
		return rootCmd.Execute()
	})

	assert.NoError(t, err)
	assert.Contains(t, output, "small")
	assert.Contains(t, output, "large")
}

func TestUpdateMVE(t *testing.T) {
	originalPrompt := utils.ResourcePrompt
	originalLoginFunc := config.LoginFunc
	defer func() {
		utils.ResourcePrompt = originalPrompt
		config.LoginFunc = originalLoginFunc
	}()

	mockService := &MockMVEService{
		GetMVEResult: &megaport.MVE{
			Name: "Mock MVE",
		},
	}
	config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
		client := &megaport.Client{
			MVEService: mockService,
		}
		return client, nil
	}

	tests := []struct {
		name            string
		args            []string
		interactive     bool
		prompts         []string
		flags           map[string]string
		jsonInput       string
		mockSetup       func(*MockMVEService)
		expectedError   string
		expectedOutput  string
		validateRequest func(*testing.T, *megaport.ModifyMVERequest)
	}{
		{
			name:        "interactive mode success",
			args:        []string{"mve-123"},
			interactive: true,
			prompts: []string{
				"Updated MVE",     // name
				"New Cost Centre", // cost centre
				"24",              // contract term months
			},
			mockSetup: func(m *MockMVEService) {
				m.ModifyMVEResult = &megaport.ModifyMVEResponse{
					MVEUpdated: true,
				}
			},
			expectedOutput: "MVE updated mve-123",
			validateRequest: func(t *testing.T, req *megaport.ModifyMVERequest) {
				assert.Equal(t, "mve-123", req.MVEID)
				assert.Equal(t, "Updated MVE", req.Name)
				assert.Equal(t, "New Cost Centre", req.CostCentre)
				assert.NotNil(t, req.ContractTermMonths)
				assert.Equal(t, 24, *req.ContractTermMonths)
				assert.True(t, req.WaitForUpdate)
			},
		},
		{
			name: "flag mode success",
			args: []string{"mve-123"},
			flags: map[string]string{
				"name":          "Flag Updated MVE",
				"cost-centre":   "Flag Cost Centre",
				"contract-term": "36",
			},
			mockSetup: func(m *MockMVEService) {
				m.ModifyMVEResult = &megaport.ModifyMVEResponse{
					MVEUpdated: true,
				}
			},
			expectedOutput: "MVE updated mve-123",
			validateRequest: func(t *testing.T, req *megaport.ModifyMVERequest) {
				assert.Equal(t, "mve-123", req.MVEID)
				assert.Equal(t, "Flag Updated MVE", req.Name)
				assert.Equal(t, "Flag Cost Centre", req.CostCentre)
				assert.NotNil(t, req.ContractTermMonths)
				assert.Equal(t, 36, *req.ContractTermMonths)
				assert.True(t, req.WaitForUpdate)
			},
		},
		{
			name: "json mode success",
			args: []string{"mve-123"},
			flags: map[string]string{
				"json": `{
                    "name": "JSON Updated MVE",
                    "costCentre": "JSON Cost Centre",
                    "contractTermMonths": 12
                }`,
			},
			mockSetup: func(m *MockMVEService) {
				m.ModifyMVEResult = &megaport.ModifyMVEResponse{
					MVEUpdated: true,
				}
			},
			expectedOutput: "MVE updated mve-123",
			validateRequest: func(t *testing.T, req *megaport.ModifyMVERequest) {
				assert.Equal(t, "mve-123", req.MVEID)
				assert.Equal(t, "JSON Updated MVE", req.Name)
				assert.Equal(t, "JSON Cost Centre", req.CostCentre)
				assert.NotNil(t, req.ContractTermMonths)
				assert.Equal(t, 12, *req.ContractTermMonths)
				assert.True(t, req.WaitForUpdate)
			},
		},
		{
			name:          "no input provided",
			args:          []string{"mve-123"},
			expectedError: "no input provided",
		},
		{
			name: "update error",
			args: []string{"mve-123"},
			flags: map[string]string{
				"name": "Updated MVE",
			},
			mockSetup: func(m *MockMVEService) {
				m.ModifyMVEError = fmt.Errorf("update failed")
			},
			expectedError: "update failed",
		},
		{
			name: "invalid contract term",
			args: []string{"mve-123"},
			flags: map[string]string{
				"contract-term": "13", // Not 1, 12, 24, or 36
			},
			expectedError: "invalid contract term",
		},
		{
			name: "update not successful",
			args: []string{"mve-123"},
			flags: map[string]string{
				"name": "Updated MVE",
			},
			mockSetup: func(m *MockMVEService) {
				m.ModifyMVEResult = &megaport.ModifyMVEResponse{
					MVEUpdated: false,
				}
				m.GetMVEResult = &megaport.MVE{
					Name: "Mock MVE",
				}
			},
			expectedError: "MVE update request was not successful",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.Reset()
			if tt.mockSetup != nil {
				tt.mockSetup(mockService)
			}

			// Set up prompts
			promptIndex := 0
			utils.ResourcePrompt = func(_, msg string, _ bool) (string, error) {
				if promptIndex < len(tt.prompts) {
					response := tt.prompts[promptIndex]
					promptIndex++
					return response, nil
				}
				return "", fmt.Errorf("unexpected prompt call")
			}

			// Create a fresh command for each test
			cmd := &cobra.Command{Use: "update"}
			cmd.Flags().Bool("interactive", tt.interactive, "")
			cmd.Flags().String("json", "", "")
			cmd.Flags().String("json-file", "", "")
			cmd.Flags().String("name", "", "")
			cmd.Flags().String("cost-centre", "", "")
			cmd.Flags().Int("contract-term", 0, "")

			// Set flag values
			for flag, value := range tt.flags {
				err := cmd.Flags().Set(flag, value)
				assert.NoError(t, err)
			}

			// Run the command
			var err error
			output := output.CaptureOutput(func() {
				err = UpdateMVE(cmd, tt.args, noColor)
			})

			// Check results
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, output, tt.expectedOutput)

				if tt.validateRequest != nil {
					tt.validateRequest(t, mockService.CapturedModifyMVERequest)
				}
			}
		})
	}
}

func TestDeleteMVE(t *testing.T) {
	originalLoginFunc := config.LoginFunc
	originalConfirmPrompt := utils.ConfirmPrompt
	defer func() {
		config.LoginFunc = originalLoginFunc
		utils.ConfirmPrompt = originalConfirmPrompt
	}()

	tests := []struct {
		name           string
		mockSetup      func(*MockMVEService)
		confirmDelete  bool
		forceFlag      bool
		nowFlag        bool
		expectedError  string
		expectedOutput string
	}{
		{
			name: "successful MVE deletion",
			mockSetup: func(m *MockMVEService) {
				m.DeleteMVEError = nil
			},
			confirmDelete:  true,
			expectedOutput: "MVE deleted mve-uid",
		},
		{
			name: "deletion error",
			mockSetup: func(m *MockMVEService) {
				m.DeleteMVEError = fmt.Errorf("deletion failed")
				m.DeleteMVEResult = &megaport.DeleteMVEResponse{
					IsDeleted: false,
				}
			},
			confirmDelete: true,
			expectedError: "deletion failed",
		},
		{
			name: "deletion cancelled",
			mockSetup: func(m *MockMVEService) {
				// No setup needed as deletion won't be called
			},
			confirmDelete:  false,
			expectedOutput: "Deletion cancelled",
		},
		{
			name: "force deletion",
			mockSetup: func(m *MockMVEService) {
				m.DeleteMVEError = nil
			},
			forceFlag:      true,
			expectedOutput: "MVE deleted mve-uid",
		},
		{
			name: "immediate deletion",
			mockSetup: func(m *MockMVEService) {
				m.DeleteMVEError = nil
			},
			confirmDelete:  true,
			nowFlag:        true,
			expectedOutput: "MVE deleted mve-uid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockMVEService{}
			tt.mockSetup(mockService)

			config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
				client := &megaport.Client{
					MVEService: mockService,
				}
				return client, nil
			}

			// Mock the confirmation prompt
			utils.ConfirmPrompt = func(question string, _ bool) bool {
				return tt.confirmDelete
			}

			// Create a new command for testing
			cmd := &cobra.Command{
				Use: "delete",
				RunE: func(cmd *cobra.Command, args []string) error {
					return DeleteMVE(cmd, []string{"mve-uid"}, noColor)
				},
			}

			cmd.Flags().Bool("force", tt.forceFlag, "")
			cmd.Flags().Bool("now", tt.nowFlag, "")

			var err error
			output := output.CaptureOutput(func() {
				err = cmd.Execute()
			})

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, output, tt.expectedOutput)
			}
		})
	}
}

func TestBuyMVE(t *testing.T) {
	originalPrompt := utils.ResourcePrompt
	originalLoginFunc := config.LoginFunc
	defer func() {
		utils.ResourcePrompt = originalPrompt
		config.LoginFunc = originalLoginFunc
	}()

	mockService := &MockMVEService{}
	config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
		client := &megaport.Client{
			MVEService: mockService,
		}
		return client, nil
	}

	tests := []struct {
		name            string
		args            []string
		interactive     bool
		prompts         []string
		flags           map[string]string
		jsonInput       string
		jsonFilePath    string
		mockSetup       func(*MockMVEService)
		expectedError   string
		expectedOutput  string
		validateRequest func(*testing.T, *megaport.BuyMVERequest)
	}{
		{
			name:        "interactive mode success",
			args:        []string{},
			interactive: true,
			prompts: []string{
				"Test MVE",   // name
				"12",         // term
				"123",        // location ID
				"",           // diversity zone
				"",           // promo code
				"CC-123",     // cost centre
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
				"VNIC 1",     // VNIC description
				"100",        // VNIC VLAN
				"",           // End VNIC input
				"",           // No resource tags
			},
			mockSetup: func(m *MockMVEService) {
				m.ValidateMVEOrderError = nil
				m.BuyMVEResult = &megaport.BuyMVEResponse{
					TechnicalServiceUID: "mock-mve-uid",
				}
			},
			expectedOutput: "MVE created mock-mve-uid",
			validateRequest: func(t *testing.T, req *megaport.BuyMVERequest) {
				assert.Equal(t, "Test MVE", req.Name)
				assert.Equal(t, 12, req.Term)
				assert.Equal(t, 123, req.LocationID)
				assert.Equal(t, "CC-123", req.CostCentre)

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

				assert.Len(t, req.Vnics, 1)
				assert.Equal(t, "VNIC 1", req.Vnics[0].Description)
				assert.Equal(t, 100, req.Vnics[0].VLAN)
			},
		},
		{
			name: "flag mode success",
			args: []string{},
			flags: map[string]string{
				"name":          "Test MVE",
				"term":          "12",
				"location-id":   "123",
				"vendor-config": `{"vendor":"cisco","imageId":1,"productSize":"large","mveLabel":"label-1","manageLocally":true,"adminSshPublicKey":"admin-ssh","sshPublicKey":"ssh-key","cloudInit":"cloud-init","fmcIpAddress":"fmc-ip","fmcRegistrationKey":"fmc-key","fmcNatId":"fmc-nat"}`,
				"vnics":         `[{"description":"VNIC 1","vlan":100}]`,
			},
			mockSetup: func(m *MockMVEService) {
				m.ValidateMVEOrderError = nil
				m.BuyMVEResult = &megaport.BuyMVEResponse{
					TechnicalServiceUID: "mock-mve-uid",
				}
			},
			expectedOutput: "MVE created mock-mve-uid",
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

				assert.Len(t, req.Vnics, 1)
				assert.Equal(t, "VNIC 1", req.Vnics[0].Description)
				assert.Equal(t, 100, req.Vnics[0].VLAN)
			},
		},
		{
			name: "json mode success",
			args: []string{},
			flags: map[string]string{
				"json": `{
                    "name": "JSON MVE",
                    "term": 12,
                    "locationId": 123,
                    "vendorConfig": {
                        "vendor": "cisco",
                        "imageId": 1,
                        "productSize": "large",
                        "mveLabel": "json-label",
                        "manageLocally": true,
                        "adminSshPublicKey": "admin-ssh",
                        "sshPublicKey": "ssh-key",
                        "cloudInit": "cloud-init",
                        "fmcIpAddress": "fmc-ip",
                        "fmcRegistrationKey": "fmc-key",
                        "fmcNatId": "fmc-nat"
                    },
                    "vnics": [
                        {"description": "JSON VNIC", "vlan": 200}
                    ]
                }`,
			},
			mockSetup: func(m *MockMVEService) {
				m.ValidateMVEOrderError = nil
				m.BuyMVEResult = &megaport.BuyMVEResponse{
					TechnicalServiceUID: "mock-mve-uid",
				}
			},
			expectedOutput: "MVE created mock-mve-uid",
			validateRequest: func(t *testing.T, req *megaport.BuyMVERequest) {
				assert.Equal(t, "JSON MVE", req.Name)
				assert.Equal(t, 12, req.Term)
				assert.Equal(t, 123, req.LocationID)

				ciscoConfig, ok := req.VendorConfig.(*megaport.CiscoConfig)
				assert.True(t, ok, "Expected a CiscoConfig")
				assert.Equal(t, 1, ciscoConfig.ImageID)
				assert.Equal(t, "large", ciscoConfig.ProductSize)
				assert.Equal(t, "json-label", ciscoConfig.MVELabel)
				assert.True(t, ciscoConfig.ManageLocally)
				assert.Equal(t, "admin-ssh", ciscoConfig.AdminSSHPublicKey)
				assert.Equal(t, "ssh-key", ciscoConfig.SSHPublicKey)
				assert.Equal(t, "cloud-init", ciscoConfig.CloudInit)
				assert.Equal(t, "fmc-ip", ciscoConfig.FMCIPAddress)
				assert.Equal(t, "fmc-key", ciscoConfig.FMCRegistrationKey)
				assert.Equal(t, "fmc-nat", ciscoConfig.FMCNatID)

				assert.Len(t, req.Vnics, 1)
				assert.Equal(t, "JSON VNIC", req.Vnics[0].Description)
				assert.Equal(t, 200, req.Vnics[0].VLAN)
			},
		},
		{
			name:          "no input provided",
			args:          []string{},
			expectedError: "no input provided",
		}, {
			name: "validation error",
			flags: map[string]string{
				"name":          "Test MVE",
				"term":          "12",
				"location-id":   "123",
				"vendor-config": `{"vendor":"cisco","imageId":1,"productSize":"large","mveLabel":"label-1","manageLocally":true,"adminSshPublicKey":"admin-ssh","sshPublicKey":"ssh-key","cloudInit":"cloud-init","fmcIpAddress":"fmc-ip","fmcRegistrationKey":"fmc-key","fmcNatId":"fmc-nat"}`,
				"vnics":         `[{"description":"VNIC 1","vlan":100}]`,
			},
			mockSetup: func(m *MockMVEService) {
				m.ValidateMVEOrderError = fmt.Errorf("validation failed")
			},
			expectedError: "validation failed",
		},
		{
			name: "purchase error",
			flags: map[string]string{
				"name":          "Test MVE",
				"term":          "12",
				"location-id":   "123",
				"vendor-config": `{"vendor":"cisco","imageId":1,"productSize":"large","mveLabel":"label-1","manageLocally":true,"adminSshPublicKey":"admin-ssh","sshPublicKey":"ssh-key","cloudInit":"cloud-init","fmcIpAddress":"fmc-ip","fmcRegistrationKey":"fmc-key","fmcNatId":"fmc-nat"}`,
				"vnics":         `[{"description":"VNIC 1","vlan":100}]`,
			},
			mockSetup: func(m *MockMVEService) {
				// Make sure validation passes but purchase fails
				m.ValidateMVEOrderError = nil
				m.BuyMVEError = fmt.Errorf("purchase failed")
			},
			expectedError: "purchase failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.Reset()
			if tt.mockSetup != nil {
				tt.mockSetup(mockService)
			}

			// Set up prompts
			promptIndex := 0
			utils.ResourcePrompt = func(_, msg string, _ bool) (string, error) {
				if promptIndex < len(tt.prompts) {
					response := tt.prompts[promptIndex]
					promptIndex++
					return response, nil
				}
				return "", fmt.Errorf("unexpected prompt call")
			}

			// Create a fresh command for each test
			cmd := &cobra.Command{Use: "buy"}
			cmd.Flags().Bool("interactive", tt.interactive, "")
			cmd.Flags().String("json", "", "")
			cmd.Flags().String("json-file", "", "")
			cmd.Flags().String("name", "", "")
			cmd.Flags().Int("term", 0, "")
			cmd.Flags().Int("location-id", 0, "")
			cmd.Flags().String("vendor-config", "", "")
			cmd.Flags().String("vnics", "", "")

			// Set flag values
			for flag, value := range tt.flags {
				err := cmd.Flags().Set(flag, value)
				assert.NoError(t, err)
			}

			// Run the command
			var err error
			output := output.CaptureOutput(func() {
				err = BuyMVE(cmd, tt.args, noColor)
			})

			// Check results
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

// Add this test function to the existing file
func TestListMVEsCmd_WithMockClient(t *testing.T) {
	// Save original login function and restore after test
	originalLoginFunc := config.LoginFunc
	defer func() {
		config.LoginFunc = originalLoginFunc
	}()

	// Test MVEs for our mock response
	mves := []*megaport.MVE{
		{
			UID:                "mve-1",
			Name:               "TestMVE-1",
			LocationID:         123,
			ProvisioningStatus: "LIVE",
			Vendor:             "cisco",
		},
		{
			UID:                "mve-2",
			Name:               "TestMVE-2",
			LocationID:         456,
			ProvisioningStatus: "CONFIGURED",
			Vendor:             "fortinet",
		},
		{
			UID:                "mve-3",
			Name:               "MVE-Decommissioned",
			LocationID:         789,
			ProvisioningStatus: "DECOMMISSIONED",
			Vendor:             "versa",
		},
	}

	tests := []struct {
		name             string
		flags            map[string]string
		outputFormat     string
		setupMock        func(*MockMVEService)
		expectedError    string
		expectedOutput   []string
		unexpectedOutput []string
	}{
		{
			name:         "list all active mves",
			outputFormat: "table",
			setupMock: func(m *MockMVEService) {
				m.ListMVEsResult = mves
			},
			expectedOutput:   []string{"mve-1", "TestMVE-1", "mve-2", "TestMVE-2"},
			unexpectedOutput: []string{"mve-3", "MVE-Decommissioned", "DECOMMISSIONED"}, // Shouldn't include inactive MVEs
		},
		{
			name:         "list all mves including inactive",
			flags:        map[string]string{"include-inactive": "true"},
			outputFormat: "table",
			setupMock: func(m *MockMVEService) {
				m.ListMVEsResult = mves
			},
			expectedOutput: []string{"mve-1", "TestMVE-1", "mve-2", "TestMVE-2", "mve-3", "MVE-Decommissioned", "DECOMMISSIONED"},
		},
		{
			name:         "filter by location ID",
			flags:        map[string]string{"location-id": "123"},
			outputFormat: "table",
			setupMock: func(m *MockMVEService) {
				m.ListMVEsResult = mves
			},
			expectedOutput:   []string{"mve-1", "TestMVE-1"},
			unexpectedOutput: []string{"mve-2", "TestMVE-2", "mve-3"},
		},
		{
			name:         "filter by vendor",
			flags:        map[string]string{"vendor": "cisco"},
			outputFormat: "table",
			setupMock: func(m *MockMVEService) {
				m.ListMVEsResult = mves
			},
			expectedOutput:   []string{"mve-1", "TestMVE-1"},
			unexpectedOutput: []string{"mve-2", "fortinet", "mve-3"},
		},
		{
			name:         "filter by name",
			flags:        map[string]string{"name": "TestMVE"},
			outputFormat: "table",
			setupMock: func(m *MockMVEService) {
				m.ListMVEsResult = mves
			},
			expectedOutput:   []string{"mve-1", "TestMVE-1", "mve-2", "TestMVE-2"},
			unexpectedOutput: []string{"mve-3", "MVE-Decommissioned"},
		},
		{
			name: "combined filters",
			flags: map[string]string{
				"location-id": "123",
				"vendor":      "cisco",
				"name":        "TestMVE",
			},
			outputFormat: "table",
			setupMock: func(m *MockMVEService) {
				m.ListMVEsResult = mves
			},
			expectedOutput:   []string{"mve-1", "TestMVE-1"},
			unexpectedOutput: []string{"mve-2", "TestMVE-2", "mve-3"},
		},
		{
			name:         "json format",
			outputFormat: "json",
			setupMock: func(m *MockMVEService) {
				m.ListMVEsResult = mves
			},
			expectedOutput:   []string{`"uid": "mve-1"`, `"name": "TestMVE-1"`, `"uid": "mve-2"`, `"name": "TestMVE-2"`},
			unexpectedOutput: []string{`"uid": "mve-3"`},
		},
		{
			name:         "no matching mves",
			flags:        map[string]string{"location-id": "999"},
			outputFormat: "table",
			setupMock: func(m *MockMVEService) {
				m.ListMVEsResult = mves
			},
			// Fix case to match actual output (capital N in "No")
			expectedOutput:   []string{"No MVEs found matching the specified filters"},
			unexpectedOutput: []string{"mve-1", "mve-2", "mve-3"},
		},
		{
			name:         "API error",
			outputFormat: "table",
			setupMock: func(m *MockMVEService) {
				m.ListMVEsError = fmt.Errorf("API error: service unavailable")
			},
			expectedError: "error listing MVEs",
		},
		{
			name:         "empty result",
			outputFormat: "table",
			setupMock: func(m *MockMVEService) {
				m.ListMVEsResult = []*megaport.MVE{}
			},
			// Fix case to match actual output (capital N in "No")
			expectedOutput: []string{"No MVEs found matching the specified filters"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new mock mve service for each test
			mockMVEService := &MockMVEService{}
			if tt.setupMock != nil {
				tt.setupMock(mockMVEService)
			}

			// Mock the login function to return a client with our mock service
			config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
				return &megaport.Client{
					MVEService: mockMVEService,
				}, nil
			}

			// Create command with flags
			cmd := &cobra.Command{}
			cmd.Flags().Bool("include-inactive", false, "")
			cmd.Flags().Int("location-id", 0, "")
			cmd.Flags().String("vendor", "", "")
			cmd.Flags().String("name", "", "")
			cmd.Flags().String("output", tt.outputFormat, "")

			// Set flag values from test case
			for flag, value := range tt.flags {
				if flag == "include-inactive" {
					boolVal, _ := strconv.ParseBool(value)
					err := cmd.Flags().Set(flag, strconv.FormatBool(boolVal))
					assert.NoError(t, err)
				} else if flag == "location-id" {
					err := cmd.Flags().Set(flag, value)
					assert.NoError(t, err)
				} else {
					err := cmd.Flags().Set(flag, value)
					assert.NoError(t, err)
				}
			}
			err := cmd.Flags().Set("output", tt.outputFormat)
			assert.NoError(t, err)

			// Capture output and run the command
			out, err := output.CaptureOutputErr(func() error {
				return ListMVEs(cmd, []string{}, true, tt.outputFormat)
			})

			// Check error if expected
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)

				// Verify expected output
				for _, expected := range tt.expectedOutput {
					assert.Contains(t, out, expected)
				}

				// Verify unexpected output is not present
				for _, unexpected := range tt.unexpectedOutput {
					assert.NotContains(t, out, unexpected)
				}

				// Verify that the right request was made with include-inactive
				includeInactive, _ := cmd.Flags().GetBool("include-inactive")
				assert.Equal(t, includeInactive, mockMVEService.CapturedListMVEsRequest.IncludeInactive)
			}
		})
	}
}
