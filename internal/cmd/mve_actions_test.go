package cmd

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

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
	originalLoginFunc := loginFunc
	defer func() {
		loginFunc = originalLoginFunc
	}()

	mockService := &MockMVEService{
		ListMVEImagesResult: testMVEImages,
	}

	loginFunc = func(ctx context.Context) (*megaport.Client, error) {
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
					client, err := loginFunc(ctx)
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
			output, err := captureOutputErr(func() error {
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
	originalLoginFunc := loginFunc
	defer func() {
		loginFunc = originalLoginFunc
	}()

	mockService := &MockMVEService{
		ListAvailableMVESizesResult: testMVESizes,
	}

	loginFunc = func(ctx context.Context) (*megaport.Client, error) {
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
			client, err := loginFunc(ctx)
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
	output, err := captureOutputErr(func() error {
		return rootCmd.Execute()
	})

	assert.NoError(t, err)
	assert.Contains(t, output, "small")
	assert.Contains(t, output, "large")
}

func TestUpdateMVE(t *testing.T) {
	originalPrompt := prompt
	originalLoginFunc := loginFunc
	defer func() {
		prompt = originalPrompt
		loginFunc = originalLoginFunc
	}()

	mockService := &MockMVEService{}
	loginFunc = func(ctx context.Context) (*megaport.Client, error) {
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
			expectedOutput: "MVE updated successfully",
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
			expectedOutput: "MVE updated successfully",
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
			expectedOutput: "MVE updated successfully",
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
			expectedError: "error updating MVE: update failed",
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
			},
			expectedOutput: "MVE update request was not successful",
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
			prompt = func(msg string) (string, error) {
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
			output := captureOutput(func() {
				err = UpdateMVE(cmd, tt.args)
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
	originalLoginFunc := loginFunc
	defer func() {
		loginFunc = originalLoginFunc
	}()

	tests := []struct {
		name           string
		mockSetup      func(*MockMVEService)
		expectedError  string
		expectedOutput string
	}{
		{
			name: "successful MVE deletion",
			mockSetup: func(m *MockMVEService) {
				m.DeleteMVEError = nil
			},
			expectedOutput: "MVE deleted successfully",
		},
		{
			name: "deletion error",
			mockSetup: func(m *MockMVEService) {
				m.DeleteMVEError = fmt.Errorf("deletion failed")
			},
			expectedError: "deletion failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockMVEService{}
			tt.mockSetup(mockService)

			loginFunc = func(ctx context.Context) (*megaport.Client, error) {
				client := &megaport.Client{
					MVEService: mockService,
				}
				return client, nil
			}

			cmd := deleteMVECmd
			cmd.SetArgs([]string{"mve-uid"})
			var err error
			output := captureOutput(func() {
				err = cmd.RunE(cmd, []string{"mve-uid"})
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
	originalPrompt := prompt
	originalLoginFunc := loginFunc
	defer func() {
		prompt = originalPrompt
		loginFunc = originalLoginFunc
	}()

	mockService := &MockMVEService{}
	loginFunc = func(ctx context.Context) (*megaport.Client, error) {
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
			expectedOutput: "MVE purchased successfully - UID: mock-mve-uid",
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
			expectedOutput: "MVE purchased successfully - UID: mock-mve-uid",
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
			prompt = func(msg string) (string, error) {
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
			output := captureOutput(func() {
				err = BuyMVE(cmd, tt.args)
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

// TestListMVEsCmd_WithMockClient tests the ListMVEs command with various filter combinations
func TestListMVEsCmd_WithMockClient(t *testing.T) {
	originalLoginFunc := loginFunc
	originalListMVEsFunc := listMVEsFunc
	originalOutputFormat := outputFormat
	defer func() {
		loginFunc = originalLoginFunc
		listMVEsFunc = originalListMVEsFunc
		outputFormat = originalOutputFormat
	}()

	// Set up test MVEs
	mves := []*megaport.MVE{
		{
			UID:                "mve-123",
			Name:               "Production MVE",
			LocationID:         123,
			Vendor:             "Cisco",
			Size:               "MEDIUM",
			ProvisioningStatus: "LIVE",
			CreateDate:         &megaport.Time{Time: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)},
			ContractTermMonths: 12,
			LocationDetails: &megaport.ProductLocationDetails{
				Name: "Sydney Data Center",
			},
		},
		{
			UID:                "mve-456",
			Name:               "Dev MVE",
			LocationID:         456,
			Vendor:             "Palo Alto",
			Size:               "LARGE",
			ProvisioningStatus: "LIVE",
			CreateDate:         &megaport.Time{Time: time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC)},
			ContractTermMonths: 24,
			LocationDetails: &megaport.ProductLocationDetails{
				Name: "Melbourne Data Center",
			},
		},
		{
			UID:                "mve-789",
			Name:               "Test MVE",
			LocationID:         123,
			Vendor:             "Cisco",
			Size:               "SMALL",
			ProvisioningStatus: "DECOMMISSIONED",
			CreateDate:         &megaport.Time{Time: time.Date(2023, 3, 1, 0, 0, 0, 0, time.UTC)},
			ContractTermMonths: 12,
			LocationDetails: &megaport.ProductLocationDetails{
				Name: "Sydney Data Center",
			},
		},
	}

	tests := []struct {
		name           string
		flags          map[string]string
		format         string
		expectedOutput []string
		expectedMVEs   int
		excludedMVEs   []string
	}{
		{
			name:           "list active MVEs only",
			flags:          map[string]string{},
			format:         "table",
			expectedOutput: []string{"Production MVE", "Dev MVE"},
			expectedMVEs:   2,
			excludedMVEs:   []string{"Test MVE"},
		},
		{
			name:           "list all MVEs including inactive",
			flags:          map[string]string{"inactive": "true"},
			format:         "table",
			expectedOutput: []string{"Production MVE", "Dev MVE", "Test MVE"},
			expectedMVEs:   3,
		},
		{
			name:           "filter by name",
			flags:          map[string]string{"name": "Production"},
			format:         "table",
			expectedOutput: []string{"Production MVE"},
			expectedMVEs:   1,
			excludedMVEs:   []string{"Dev MVE", "Test MVE"},
		},
		{
			name:           "filter by location ID",
			flags:          map[string]string{"location-id": "123"},
			format:         "table",
			expectedOutput: []string{"Production MVE"},
			expectedMVEs:   1,
			excludedMVEs:   []string{"Dev MVE", "Test MVE"},
		},
		{
			name:           "filter by vendor",
			flags:          map[string]string{"vendor": "Palo Alto"},
			format:         "table",
			expectedOutput: []string{"Dev MVE"},
			expectedMVEs:   1,
			excludedMVEs:   []string{"Production MVE", "Test MVE"},
		},
		{
			name: "combine filters - name and vendor",
			flags: map[string]string{
				"name":   "Production",
				"vendor": "Cisco",
			},
			format:         "table",
			expectedOutput: []string{"Production MVE"},
			expectedMVEs:   1,
			excludedMVEs:   []string{"Dev MVE", "Test MVE"},
		},
		{
			name: "combine filters - inactive and location ID",
			flags: map[string]string{
				"inactive":    "true",
				"location-id": "123",
			},
			format:         "table",
			expectedOutput: []string{"Production MVE", "Test MVE"},
			expectedMVEs:   2,
			excludedMVEs:   []string{"Dev MVE"},
		},
		{
			name:           "no match with filters",
			flags:          map[string]string{"name": "NonExistent"},
			format:         "table",
			expectedOutput: []string{},
			expectedMVEs:   0,
			excludedMVEs:   []string{"Production MVE", "Dev MVE", "Test MVE"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up mock function to return our test MVEs
			listMVEsFunc = func(ctx context.Context, client *megaport.Client, req *megaport.ListMVEsRequest) ([]*megaport.MVE, error) {
				if req.IncludeInactive {
					return mves, nil
				}
				// Filter out decommissioned MVEs
				activeMVEs := make([]*megaport.MVE, 0)
				for _, mve := range mves {
					if mve.ProvisioningStatus != "DECOMMISSIONED" {
						activeMVEs = append(activeMVEs, mve)
					}
				}
				return activeMVEs, nil
			}

			// Set up mock login
			loginFunc = func(ctx context.Context) (*megaport.Client, error) {
				return &megaport.Client{}, nil
			}

			// Set output format
			outputFormat = tt.format

			// Create a fresh command for each test to avoid flag conflicts
			cmd := &cobra.Command{
				Use:  "list",
				RunE: ListMVEs,
			}

			// Add all the necessary flags
			cmd.Flags().Bool("inactive", false, "Include inactive MVEs")
			cmd.Flags().String("name", "", "Filter by name")
			cmd.Flags().Int("location-id", 0, "Filter by location ID")
			cmd.Flags().String("vendor", "", "Filter by vendor")

			// Set flag values for this test
			for flagName, flagValue := range tt.flags {
				err := cmd.Flags().Set(flagName, flagValue)
				if err != nil {
					t.Fatalf("Failed to set %s flag: %v", flagName, err)
				}
			}

			// Execute command and capture output
			var err error
			output := captureOutput(func() {
				err = cmd.RunE(cmd, []string{})
			})

			// Check for errors
			assert.NoError(t, err)

			// Verify expected content in output
			for _, expected := range tt.expectedOutput {
				assert.Contains(t, output, expected)
			}

			// Verify excluded content not in output
			for _, excluded := range tt.excludedMVEs {
				assert.NotContains(t, output, excluded)
			}
		})
	}
}
