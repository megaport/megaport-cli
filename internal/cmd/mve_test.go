package cmd

import (
	"context"
	"fmt"
	"strings"
	"testing"

	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

var testMVEs = []*megaport.MVE{
	{
		UID:        "mve-1",
		Name:       "MyMVEOne",
		LocationID: 1,
	},
	{
		UID:        "mve-2",
		Name:       "AnotherMVE",
		LocationID: 2,
	},
}

func TestPrintMVEs_Table(t *testing.T) {
	output := captureOutput(func() {
		err := printMVEs(testMVEs, "table")
		assert.NoError(t, err)
	})

	expected := `uid     name         location_id
mve-1   MyMVEOne     1
mve-2   AnotherMVE   2
`
	assert.Equal(t, expected, output)
}

func TestPrintMVEs_JSON(t *testing.T) {
	output := captureOutput(func() {
		err := printMVEs(testMVEs, "json")
		assert.NoError(t, err)
	})

	expected := `[
  {
    "uid": "mve-1",
    "name": "MyMVEOne",
    "location_id": 1
  },
  {
    "uid": "mve-2",
    "name": "AnotherMVE",
    "location_id": 2
  }
]`
	assert.JSONEq(t, expected, output)
}

func TestPrintMVEs_CSV(t *testing.T) {
	output := captureOutput(func() {
		err := printMVEs(testMVEs, "csv")
		assert.NoError(t, err)
	})

	expected := `uid,name,location_id
mve-1,MyMVEOne,1
mve-2,AnotherMVE,2
`
	assert.Equal(t, expected, output)
}

func TestPrintMVEs_Invalid(t *testing.T) {
	var err error
	output := captureOutput(func() {
		err = printMVEs(testMVEs, "invalid")
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid output format")
	assert.Empty(t, output)
}

func TestPrintMVEs_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		mves        []*megaport.MVE
		format      string
		shouldError bool
		expected    string
	}{
		{
			name:        "nil slice",
			mves:        nil,
			format:      "table",
			shouldError: false,
			expected:    "uid   name   location_id\n",
		},
		{
			name:        "empty slice",
			mves:        []*megaport.MVE{},
			format:      "json",
			shouldError: false,
			expected:    "[]",
		},
		{
			name: "nil mve in slice",
			mves: []*megaport.MVE{
				nil,
				{
					UID:        "mve-1",
					Name:       "TestMVE",
					LocationID: 1,
				},
			},
			format:      "table",
			shouldError: true,
			expected:    "invalid MVE: nil value",
		},
		{
			name: "zero values",
			mves: []*megaport.MVE{
				{
					UID:        "",
					Name:       "",
					LocationID: 0,
				},
			},
			format:      "csv",
			shouldError: false,
			expected:    "uid,name,location_id\n,,0\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output string
			var err error

			output = captureOutput(func() {
				err = printMVEs(tt.mves, tt.format)
			})

			if tt.shouldError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expected)
				assert.Empty(t, output)
			} else {
				assert.NoError(t, err)
				switch tt.format {
				case "json":
					assert.JSONEq(t, tt.expected, output)
				case "table", "csv":
					assert.Equal(t, tt.expected, output)
				}
			}
		})
	}
}

func TestToMVEOutput_EdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		mve           *megaport.MVE
		shouldError   bool
		errorContains string
		validateFunc  func(*testing.T, MVEOutput)
	}{
		{
			name:          "nil mve",
			mve:           nil,
			shouldError:   true,
			errorContains: "invalid MVE: nil value",
		},
		{
			name: "zero values",
			mve:  &megaport.MVE{},
			validateFunc: func(t *testing.T, output MVEOutput) {
				assert.Empty(t, output.UID)
				assert.Empty(t, output.Name)
				assert.Zero(t, output.LocationID)
			},
		},
		{
			name: "whitespace values",
			mve: &megaport.MVE{
				UID:        "   ",
				Name:       "   ",
				LocationID: 0,
			},
			validateFunc: func(t *testing.T, output MVEOutput) {
				assert.Equal(t, "   ", output.UID)
				assert.Equal(t, "   ", output.Name)
				assert.Zero(t, output.LocationID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := ToMVEOutput(tt.mve)

			if tt.shouldError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				if tt.validateFunc != nil {
					tt.validateFunc(t, output)
				}
			}
		})
	}
}

func TestBuyMVE(t *testing.T) {
	originalPrompt := prompt
	originalLoginFunc := loginFunc
	originalBuyMVEFunc := buyMVEFunc
	defer func() {
		prompt = originalPrompt
		loginFunc = originalLoginFunc
		buyMVEFunc = originalBuyMVEFunc
	}()

	tests := []struct {
		name            string
		prompts         []string
		mockSetup       func(*MockMVEService)
		expectedError   string
		expectedOutput  string
		validateRequest func(*testing.T, *megaport.BuyMVERequest)
	}{
		{
			name: "successful MVE purchase",
			prompts: []string{
				"Test MVE",   // name
				"12",         // term
				"123",        // location ID
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
			},
			mockSetup: func(m *MockMVEService) {
				m.ValidateMVEOrderError = nil
				m.BuyMVEError = nil
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
			},
		},
		{
			name: "validation error",
			prompts: []string{
				"Test MVE",   // name
				"12",         // term
				"123",        // location ID
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
			},
			mockSetup: func(m *MockMVEService) {
				m.ValidateMVEOrderError = fmt.Errorf("validation failed")
			},
			expectedError: "validation failed",
		},
		{
			name: "purchase error",
			prompts: []string{
				"Test MVE",   // name
				"12",         // term
				"123",        // location ID
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
			},
			mockSetup: func(m *MockMVEService) {
				m.ValidateMVEOrderError = nil
				m.BuyMVEError = fmt.Errorf("purchase failed")
			},
			expectedError: "purchase failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			promptIndex := 0
			prompt = func(msg string) (string, error) {
				if promptIndex < len(tt.prompts) {
					response := tt.prompts[promptIndex]
					promptIndex++
					return response, nil
				}
				return "", fmt.Errorf("unexpected prompt call")
			}

			mockService := &MockMVEService{}
			tt.mockSetup(mockService)

			loginFunc = func(ctx context.Context) (*megaport.Client, error) {
				client := &megaport.Client{
					MVEService: mockService,
				}
				return client, nil
			}

			// Use the actual buyMVEFunc to make sure we call the mock service methods
			buyMVEFunc = func(ctx context.Context, client *megaport.Client, req *megaport.BuyMVERequest) (*megaport.BuyMVEResponse, error) {
				return client.MVEService.BuyMVE(ctx, req)
			}

			cmd := buyMVECmd
			var err error
			output := captureOutput(func() {
				err = cmd.RunE(cmd, []string{})
			})

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
