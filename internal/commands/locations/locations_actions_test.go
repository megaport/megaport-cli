package locations

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/commands/config"
	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/require"
)

func TestMockSetup(t *testing.T) {
	mockSvc := new(MockLocationsService)
	testLocs := []*megaport.LocationV3{
		{
			ID:    1,
			Name:  "Test Location 1",
			Metro: "Sydney",
			Address: megaport.LocationV3Address{
				Country: "Australia",
			},
		},
	}

	mockSvc.On("ListLocationsV3", mock.Anything).Return(testLocs, nil)

	locations, err := mockSvc.ListLocationsV3(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, 1, len(locations))
	assert.Equal(t, "Test Location 1", locations[0].Name)
	mockSvc.AssertExpectations(t)
}

func setupTestEnvironment() *MockLocationsService {
	mockSvc := new(MockLocationsService)

	return mockSvc
}

func TestListLocationsFunc(t *testing.T) {
	mockSvc := setupTestEnvironment()

	testLocationsV3 := []*megaport.LocationV3{
		{
			ID:     1,
			Name:   "Sydney Data Center",
			Metro:  "Sydney",
			Market: "AU",
			Status: "Active",
			Address: megaport.LocationV3Address{
				Country: "Australia",
			},
			DiversityZones: &megaport.LocationV3DiversityZones{
				Red: &megaport.LocationV3DiversityZone{
					McrSpeedMbps:      []int{1000, 10000},
					MegaportSpeedMbps: []int{1, 10},
				},
			},
		},
		{
			ID:     2,
			Name:   "London Data Center",
			Metro:  "London",
			Market: "UK",
			Status: "Active",
			Address: megaport.LocationV3Address{
				Country: "United Kingdom",
			},
			DiversityZones: &megaport.LocationV3DiversityZones{
				Red: &megaport.LocationV3DiversityZone{
					MegaportSpeedMbps: []int{1},
				},
			},
		},
		{
			ID:     3,
			Name:   "New York Data Center",
			Metro:  "New York",
			Market: "US",
			Status: "Active",
			Address: megaport.LocationV3Address{
				Country: "USA",
			},
			DiversityZones: &megaport.LocationV3DiversityZones{
				Red: &megaport.LocationV3DiversityZone{
					McrSpeedMbps:      []int{10000},
					MegaportSpeedMbps: []int{10},
				},
			},
		},
	}

	mockSvc.On("ListLocationsV3", mock.Anything).Return(testLocationsV3, nil)

	originalLoginFunc := config.LoginFunc
	originalListLocationsFunc := listLocationsFunc
	defer func() {
		config.LoginFunc = originalLoginFunc
		listLocationsFunc = originalListLocationsFunc
	}()

	config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
		testClient := &megaport.Client{}
		testClient.LocationService = mockSvc
		return testClient, nil
	}

	testClient, err := config.LoginFunc(context.Background())
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}

	listLocationsFunc = func(ctx context.Context, client *megaport.Client) ([]*megaport.LocationV3, error) {
		return client.LocationService.ListLocationsV3(ctx)
	}

	locations, err := listLocationsFunc(context.Background(), testClient)

	assert.NoError(t, err)
	assert.Equal(t, 3, len(locations))
	assert.Equal(t, "Sydney Data Center", locations[0].Name)
	assert.Equal(t, "London Data Center", locations[1].Name)
	assert.Equal(t, "New York Data Center", locations[2].Name)

	mockSvc.AssertExpectations(t)
}

func TestListLocationsFuncError(t *testing.T) {
	mockSvc := setupTestEnvironment()

	expectedError := errors.New("api connection failed")

	mockSvc.On("ListLocationsV3", mock.Anything).Return([]*megaport.LocationV3{}, expectedError)

	originalListLocationsFunc := listLocationsFunc
	originalLoginFunc := config.LoginFunc

	defer func() {
		config.LoginFunc = originalLoginFunc
		listLocationsFunc = originalListLocationsFunc
	}()

	config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
		testClient := &megaport.Client{}
		testClient.LocationService = mockSvc
		return testClient, nil
	}

	testClient, err := config.LoginFunc(context.Background())
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}

	listLocationsFunc = func(ctx context.Context, client *megaport.Client) ([]*megaport.LocationV3, error) {
		return client.LocationService.ListLocationsV3(ctx)
	}

	locations, err := listLocationsFunc(context.Background(), testClient)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Empty(t, locations)

	mockSvc.AssertExpectations(t)
}

func TestListLocationsCommand(t *testing.T) {
	mockSvc := setupTestEnvironment()

	testLocationsV3 := []*megaport.LocationV3{
		{
			ID:     1,
			Name:   "Sydney Data Center",
			Metro:  "Sydney",
			Market: "AU",
			Status: "Active",
			Address: megaport.LocationV3Address{
				Country: "Australia",
			},
			DiversityZones: &megaport.LocationV3DiversityZones{
				Red: &megaport.LocationV3DiversityZone{
					McrSpeedMbps:      []int{1000, 10000},
					MegaportSpeedMbps: []int{1, 10},
				},
			},
		},
		{
			ID:     2,
			Name:   "London Data Center",
			Metro:  "London",
			Market: "UK",
			Status: "Active",
			Address: megaport.LocationV3Address{
				Country: "United Kingdom",
			},
			DiversityZones: &megaport.LocationV3DiversityZones{
				Red: &megaport.LocationV3DiversityZone{
					MegaportSpeedMbps: []int{1},
				},
			},
		},
		{
			ID:     3,
			Name:   "New York Data Center",
			Metro:  "New York",
			Market: "US",
			Status: "Active",
			Address: megaport.LocationV3Address{
				Country: "USA",
			},
			DiversityZones: &megaport.LocationV3DiversityZones{
				Red: &megaport.LocationV3DiversityZone{
					McrSpeedMbps:      []int{10000},
					MegaportSpeedMbps: []int{10},
				},
			},
		},
	}

	mockSvc.On("ListLocationsV3", mock.Anything).Return(testLocationsV3, nil)

	originalListLocationsFunc := listLocationsFunc
	originalLoginFunc := config.LoginFunc

	defer func() {
		config.LoginFunc = originalLoginFunc
		listLocationsFunc = originalListLocationsFunc
	}()

	config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
		testClient := &megaport.Client{}
		testClient.LocationService = mockSvc
		return testClient, nil
	}

	listLocationsFunc = func(ctx context.Context, client *megaport.Client) ([]*megaport.LocationV3, error) {
		return client.LocationService.ListLocationsV3(ctx)
	}

	t.Run("NoFilters", func(t *testing.T) {
		output := output.CaptureOutput(func() {
			cmd := &cobra.Command{
				Use: "list",
			}
			cmd.Flags().String("metro", "", "Filter by metro")
			cmd.Flags().String("country", "", "Filter by country")
			cmd.Flags().String("name", "", "Filter by name")

			err := ListLocations(cmd, []string{}, true, "json")
			assert.NoError(t, err)
		})

		assert.Contains(t, output, "Sydney Data Center")
		assert.Contains(t, output, "London Data Center")
		assert.Contains(t, output, "New York Data Center")
	})

	t.Run("FilterByMetro", func(t *testing.T) {
		output := output.CaptureOutput(func() {
			cmd := &cobra.Command{
				Use: "list",
			}
			cmd.Flags().String("metro", "", "Filter by metro")
			cmd.Flags().String("country", "", "Filter by country")
			cmd.Flags().String("name", "", "Filter by name")

			if err := cmd.Flags().Set("metro", "New York"); err != nil {
				t.Fatalf("Failed to set flag: %v", err)
			}

			err := ListLocations(cmd, []string{}, true, "json")
			assert.NoError(t, err)
		})

		assert.NotContains(t, output, "Sydney Data Center")
		assert.NotContains(t, output, "London Data Center")
		assert.Contains(t, output, "New York Data Center")
	})

	t.Run("FilterByCountry", func(t *testing.T) {
		output := output.CaptureOutput(func() {
			cmd := &cobra.Command{
				Use: "list",
			}
			cmd.Flags().String("metro", "", "Filter by metro")
			cmd.Flags().String("country", "", "Filter by country")
			cmd.Flags().String("name", "", "Filter by name")

			if err := cmd.Flags().Set("country", "United Kingdom"); err != nil {
				t.Fatalf("Failed to set flag: %v", err)
			}

			err := ListLocations(cmd, []string{}, true, "json")
			assert.NoError(t, err)
		})

		assert.NotContains(t, output, "Sydney Data Center")
		assert.Contains(t, output, "London Data Center")
		assert.NotContains(t, output, "New York Data Center")
	})

	t.Run("FilterByName", func(t *testing.T) {
		output := output.CaptureOutput(func() {
			cmd := &cobra.Command{
				Use: "list",
			}
			cmd.Flags().String("metro", "", "Filter by metro")
			cmd.Flags().String("country", "", "Filter by country")
			cmd.Flags().String("name", "", "Filter by name")

			if err := cmd.Flags().Set("name", "Sydney Data Center"); err != nil {
				t.Fatalf("Failed to set flag: %v", err)
			}

			err := ListLocations(cmd, []string{}, true, "json")
			assert.NoError(t, err)
		})

		assert.Contains(t, output, "Sydney Data Center")
		assert.NotContains(t, output, "London Data Center")
		assert.NotContains(t, output, "New York Data Center")
	})

	t.Run("NoMatchingLocations", func(t *testing.T) {
		output := output.CaptureOutput(func() {
			cmd := &cobra.Command{
				Use: "list",
			}
			cmd.Flags().String("metro", "", "Filter by metro")
			cmd.Flags().String("country", "", "Filter by country")
			cmd.Flags().String("name", "", "Filter by name")

			if err := cmd.Flags().Set("name", "Non-existent Location"); err != nil {
				t.Fatalf("Failed to set flag: %v", err)
			}

			err := ListLocations(cmd, []string{}, true, "table")
			assert.NoError(t, err)
		})

		assert.Contains(t, output, "No locations found matching the specified filters")
	})
}

func TestGetLocation(t *testing.T) {
	testLocationsV3 := []*megaport.LocationV3{
		{
			ID:     1,
			Name:   "Sydney Data Center",
			Metro:  "Sydney",
			Market: "AU",
			Status: "Active",
			Address: megaport.LocationV3Address{
				Country: "Australia",
			},
			DiversityZones: &megaport.LocationV3DiversityZones{
				Red: &megaport.LocationV3DiversityZone{
					McrSpeedMbps:      []int{1000, 10000},
					MegaportSpeedMbps: []int{1, 10},
				},
			},
		},
		{
			ID:     2,
			Name:   "London Data Center",
			Metro:  "London",
			Market: "UK",
			Status: "Active",
			Address: megaport.LocationV3Address{
				Country: "United Kingdom",
			},
			DiversityZones: &megaport.LocationV3DiversityZones{
				Red: &megaport.LocationV3DiversityZone{
					MegaportSpeedMbps: []int{1},
				},
			},
		},
	}

	tests := []struct {
		name           string
		args           []string
		setupMock      func(*MockLocationsService)
		loginErr       error
		expectedErr    string
		expectedOutput string
	}{
		{
			name: "success",
			args: []string{"1"},
			setupMock: func(m *MockLocationsService) {
				m.On("ListLocationsV3", mock.Anything).Return(testLocationsV3, nil)
			},
			expectedOutput: "Sydney Data Center",
		},
		{
			name:        "invalid ID arg",
			args:        []string{"abc"},
			setupMock:   func(m *MockLocationsService) {},
			expectedErr: "invalid location ID",
		},
		{
			name: "not found",
			args: []string{"999"},
			setupMock: func(m *MockLocationsService) {
				m.On("ListLocationsV3", mock.Anything).Return(testLocationsV3, nil)
			},
			expectedErr: "no location found with ID: 999",
		},
		{
			name: "API error",
			args: []string{"1"},
			setupMock: func(m *MockLocationsService) {
				m.On("ListLocationsV3", mock.Anything).Return([]*megaport.LocationV3{}, fmt.Errorf("API failure"))
			},
			expectedErr: "error listing locations",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := setupTestEnvironment()
			tt.setupMock(mockSvc)

			originalLoginFunc := config.LoginFunc
			originalListLocationsFunc := listLocationsFunc
			defer func() {
				config.LoginFunc = originalLoginFunc
				listLocationsFunc = originalListLocationsFunc
			}()

			config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
				if tt.loginErr != nil {
					return nil, tt.loginErr
				}
				testClient := &megaport.Client{}
				testClient.LocationService = mockSvc
				return testClient, nil
			}

			listLocationsFunc = func(ctx context.Context, client *megaport.Client) ([]*megaport.LocationV3, error) {
				return client.LocationService.ListLocationsV3(ctx)
			}

			cmd := &cobra.Command{
				Use: "get",
			}

			defer output.SetOutputFormat("table")

			var err error
			capturedOutput := output.CaptureOutput(func() {
				err = GetLocation(cmd, tt.args, true, "json")
			})

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)
				var parsed []map[string]interface{}
				assert.NoError(t, json.Unmarshal([]byte(capturedOutput), &parsed), "JSON output should be valid JSON")
				if assert.NotEmpty(t, parsed) {
					assert.Contains(t, capturedOutput, tt.expectedOutput)
				}
			}
		})
	}
}

func TestListCountries(t *testing.T) {
	testCountries := []*megaport.Country{
		{Code: "AU", Name: "Australia", Prefix: "61", SiteCount: 15},
		{Code: "US", Name: "United States", Prefix: "1", SiteCount: 30},
		{Code: "GB", Name: "United Kingdom", Prefix: "44", SiteCount: 10},
	}

	tests := []struct {
		name           string
		setupMock      func(*MockLocationsService)
		loginErr       error
		expectedErr    string
		expectedOutput string
	}{
		{
			name: "success",
			setupMock: func(m *MockLocationsService) {
				m.On("ListCountries", mock.Anything).Return(testCountries, nil)
			},
			expectedOutput: "Australia",
		},
		{
			name: "API error",
			setupMock: func(m *MockLocationsService) {
				m.On("ListCountries", mock.Anything).Return(([]*megaport.Country)(nil), fmt.Errorf("API failure"))
			},
			expectedErr: "error listing countries",
		},
		{
			name: "empty result",
			setupMock: func(m *MockLocationsService) {
				m.On("ListCountries", mock.Anything).Return([]*megaport.Country{}, nil)
			},
			expectedOutput: "[]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := setupTestEnvironment()
			tt.setupMock(mockSvc)

			originalLoginFunc := config.LoginFunc
			originalListCountriesFunc := listCountriesFunc
			defer func() {
				config.LoginFunc = originalLoginFunc
				listCountriesFunc = originalListCountriesFunc
			}()

			config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
				if tt.loginErr != nil {
					return nil, tt.loginErr
				}
				testClient := &megaport.Client{}
				testClient.LocationService = mockSvc
				return testClient, nil
			}

			listCountriesFunc = func(ctx context.Context, client *megaport.Client) ([]*megaport.Country, error) {
				return client.LocationService.ListCountries(ctx)
			}

			cmd := &cobra.Command{Use: "list-countries"}
			defer output.SetOutputFormat("table")

			var err error
			capturedOutput := output.CaptureOutput(func() {
				err = ListCountries(cmd, []string{}, true, "json")
			})

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, capturedOutput, tt.expectedOutput)
			}
		})
	}
}

func TestListMarketCodes(t *testing.T) {
	testMarketCodes := []string{"AU", "US", "UK", "SG", "HK"}

	tests := []struct {
		name           string
		setupMock      func(*MockLocationsService)
		loginErr       error
		expectedErr    string
		expectedOutput string
	}{
		{
			name: "success",
			setupMock: func(m *MockLocationsService) {
				m.On("ListMarketCodes", mock.Anything).Return(testMarketCodes, nil)
			},
			expectedOutput: "AU",
		},
		{
			name: "API error",
			setupMock: func(m *MockLocationsService) {
				m.On("ListMarketCodes", mock.Anything).Return(([]string)(nil), fmt.Errorf("API failure"))
			},
			expectedErr: "error listing market codes",
		},
		{
			name: "empty result",
			setupMock: func(m *MockLocationsService) {
				m.On("ListMarketCodes", mock.Anything).Return([]string{}, nil)
			},
			expectedOutput: "[]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := setupTestEnvironment()
			tt.setupMock(mockSvc)

			originalLoginFunc := config.LoginFunc
			originalListMarketCodesFunc := listMarketCodesFunc
			defer func() {
				config.LoginFunc = originalLoginFunc
				listMarketCodesFunc = originalListMarketCodesFunc
			}()

			config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
				if tt.loginErr != nil {
					return nil, tt.loginErr
				}
				testClient := &megaport.Client{}
				testClient.LocationService = mockSvc
				return testClient, nil
			}

			listMarketCodesFunc = func(ctx context.Context, client *megaport.Client) ([]string, error) {
				return client.LocationService.ListMarketCodes(ctx)
			}

			cmd := &cobra.Command{Use: "list-market-codes"}
			defer output.SetOutputFormat("table")

			var err error
			capturedOutput := output.CaptureOutput(func() {
				err = ListMarketCodes(cmd, []string{}, true, "json")
			})

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, capturedOutput, tt.expectedOutput)
			}
		})
	}
}

func TestSearchLocations(t *testing.T) {
	testLocationsV3 := []*megaport.LocationV3{
		{
			ID:     1,
			Name:   "Equinix SY1",
			Metro:  "Sydney",
			Market: "AU",
			Status: "Active",
			Address: megaport.LocationV3Address{
				Country: "Australia",
			},
			DiversityZones: &megaport.LocationV3DiversityZones{
				Red: &megaport.LocationV3DiversityZone{
					MegaportSpeedMbps: []int{1},
				},
			},
		},
		{
			ID:     2,
			Name:   "Equinix SY3",
			Metro:  "Sydney",
			Market: "AU",
			Status: "Active",
			Address: megaport.LocationV3Address{
				Country: "Australia",
			},
			DiversityZones: &megaport.LocationV3DiversityZones{
				Red: &megaport.LocationV3DiversityZone{
					MegaportSpeedMbps: []int{1},
				},
			},
		},
	}

	tests := []struct {
		name           string
		args           []string
		setupMock      func(*MockLocationsService)
		loginErr       error
		expectedErr    string
		expectedOutput string
	}{
		{
			name: "success",
			args: []string{"Equinix"},
			setupMock: func(m *MockLocationsService) {
				m.On("GetLocationByNameFuzzyV3", mock.Anything, "Equinix").Return(testLocationsV3, nil)
			},
			expectedOutput: "Equinix SY1",
		},
		{
			name: "no matches",
			args: []string{"NonExistent"},
			setupMock: func(m *MockLocationsService) {
				m.On("GetLocationByNameFuzzyV3", mock.Anything, "NonExistent").Return([]*megaport.LocationV3{}, nil)
			},
			expectedOutput: "[]",
		},
		{
			name: "API error",
			args: []string{"Equinix"},
			setupMock: func(m *MockLocationsService) {
				m.On("GetLocationByNameFuzzyV3", mock.Anything, "Equinix").Return(([]*megaport.LocationV3)(nil), fmt.Errorf("API failure"))
			},
			expectedErr: "error searching locations",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := setupTestEnvironment()
			tt.setupMock(mockSvc)

			originalLoginFunc := config.LoginFunc
			originalSearchLocationsFunc := searchLocationsFunc
			defer func() {
				config.LoginFunc = originalLoginFunc
				searchLocationsFunc = originalSearchLocationsFunc
			}()

			config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
				if tt.loginErr != nil {
					return nil, tt.loginErr
				}
				testClient := &megaport.Client{}
				testClient.LocationService = mockSvc
				return testClient, nil
			}

			searchLocationsFunc = func(ctx context.Context, client *megaport.Client, search string) ([]*megaport.LocationV3, error) {
				return client.LocationService.GetLocationByNameFuzzyV3(ctx, search)
			}

			cmd := &cobra.Command{Use: "search"}
			require.Len(t, tt.args, 1)
			defer output.SetOutputFormat("table")

			var err error
			capturedOutput := output.CaptureOutput(func() {
				err = SearchLocations(cmd, tt.args, true, "json")
			})

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, capturedOutput, tt.expectedOutput)
			}
		})
	}
}
