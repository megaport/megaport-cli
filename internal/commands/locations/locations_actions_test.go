package locations

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/commands/config"
	"github.com/megaport/megaport-cli/internal/testutil"
	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/require"
)

func setupTestEnvironment() *MockLocationsService {
	return &MockLocationsService{}
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

	mockSvc.ListLocationsV3Result = testLocationsV3

	testClient := &megaport.Client{}
	testClient.LocationService = mockSvc

	originalListLocationsFunc := listLocationsFunc
	defer func() {
		listLocationsFunc = originalListLocationsFunc
	}()

	listLocationsFunc = func(ctx context.Context, client *megaport.Client, statuses []string) ([]*megaport.LocationV3, error) {
		return client.LocationService.ListLocationsV3(ctx)
	}

	locations, err := listLocationsFunc(context.Background(), testClient, nil)

	assert.NoError(t, err)
	assert.Equal(t, 3, len(locations))
	assert.Equal(t, "Sydney Data Center", locations[0].Name)
	assert.Equal(t, "London Data Center", locations[1].Name)
	assert.Equal(t, "New York Data Center", locations[2].Name)
}

func TestListLocationsSendsActiveByDefault(t *testing.T) {
	mockSvc := &MockLocationsService{}
	originalFunc := config.GetNewUnauthenticatedClientFunc()
	defer config.SetNewUnauthenticatedClientFunc(originalFunc)
	config.SetNewUnauthenticatedClientFunc(func() (*megaport.Client, error) {
		client := &megaport.Client{}
		client.LocationService = mockSvc
		return client, nil
	})
	defer output.SetOutputFormat("table")

	cmd := &cobra.Command{Use: "list"}
	cmd.Flags().StringArray("status", nil, "")
	cmd.Flags().Int("limit", 0, "")
	output.CaptureOutput(func() {
		require.NoError(t, ListLocations(cmd, nil, true, "json"))
	})

	require.NotNil(t, mockSvc.CapturedListLocationsV3Options)
	assert.Equal(t, []string{megaport.LocationStatusActive}, mockSvc.CapturedListLocationsV3Options.Statuses)
}

func TestListLocationsFuncError(t *testing.T) {
	mockSvc := setupTestEnvironment()

	expectedError := errors.New("api connection failed")

	mockSvc.ListLocationsV3Err = expectedError

	testClient := &megaport.Client{}
	testClient.LocationService = mockSvc

	originalListLocationsFunc := listLocationsFunc
	defer func() {
		listLocationsFunc = originalListLocationsFunc
	}()

	listLocationsFunc = func(ctx context.Context, client *megaport.Client, statuses []string) ([]*megaport.LocationV3, error) {
		return client.LocationService.ListLocationsV3(ctx)
	}

	locations, err := listLocationsFunc(context.Background(), testClient, nil)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Empty(t, locations)
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

	mockSvc.ListLocationsV3Result = testLocationsV3

	originalFunc := config.GetNewUnauthenticatedClientFunc()
	defer func() { config.SetNewUnauthenticatedClientFunc(originalFunc) }()

	testClient := &megaport.Client{}
	testClient.LocationService = mockSvc
	config.SetNewUnauthenticatedClientFunc(func() (*megaport.Client, error) {
		return testClient, nil
	})

	originalListLocationsFunc := listLocationsFunc
	defer func() {
		listLocationsFunc = originalListLocationsFunc
	}()

	listLocationsFunc = func(ctx context.Context, client *megaport.Client, statuses []string) ([]*megaport.LocationV3, error) {
		return client.LocationService.ListLocationsV3(ctx)
	}

	newListCmd := func() *cobra.Command {
		cmd := &cobra.Command{Use: "list"}
		cmd.Flags().String("metro", "", "Filter by metro")
		cmd.Flags().String("country", "", "Filter by country")
		cmd.Flags().String("name", "", "Filter by name")
		cmd.Flags().Int("limit", 0, "Maximum number of results to display")
		return cmd
	}

	t.Run("NoFilters", func(t *testing.T) {
		output := output.CaptureOutput(func() {
			cmd := newListCmd()
			err := ListLocations(cmd, []string{}, true, "json")
			assert.NoError(t, err)
		})

		assert.Contains(t, output, "Sydney Data Center")
		assert.Contains(t, output, "London Data Center")
		assert.Contains(t, output, "New York Data Center")
	})

	t.Run("FilterByMetro", func(t *testing.T) {
		output := output.CaptureOutput(func() {
			cmd := newListCmd()
			testutil.SetFlags(t, cmd, map[string]string{"metro": "New York"})
			err := ListLocations(cmd, []string{}, true, "json")
			assert.NoError(t, err)
		})

		assert.NotContains(t, output, "Sydney Data Center")
		assert.NotContains(t, output, "London Data Center")
		assert.Contains(t, output, "New York Data Center")
	})

	t.Run("FilterByCountry", func(t *testing.T) {
		output := output.CaptureOutput(func() {
			cmd := newListCmd()
			testutil.SetFlags(t, cmd, map[string]string{"country": "United Kingdom"})
			err := ListLocations(cmd, []string{}, true, "json")
			assert.NoError(t, err)
		})

		assert.NotContains(t, output, "Sydney Data Center")
		assert.Contains(t, output, "London Data Center")
		assert.NotContains(t, output, "New York Data Center")
	})

	t.Run("FilterByName", func(t *testing.T) {
		output := output.CaptureOutput(func() {
			cmd := newListCmd()
			testutil.SetFlags(t, cmd, map[string]string{"name": "Sydney Data Center"})
			err := ListLocations(cmd, []string{}, true, "json")
			assert.NoError(t, err)
		})

		assert.Contains(t, output, "Sydney Data Center")
		assert.NotContains(t, output, "London Data Center")
		assert.NotContains(t, output, "New York Data Center")
	})

	t.Run("NoMatchingLocations", func(t *testing.T) {
		stderr := captureStderr(t, func() {
			cmd := newListCmd()
			testutil.SetFlags(t, cmd, map[string]string{"name": "Non-existent Location"})
			err := ListLocations(cmd, []string{}, true, "table")
			assert.NoError(t, err)
		})

		assert.Contains(t, stderr, "No locations found matching your filters.")
	})

	t.Run("LimitResults", func(t *testing.T) {
		out := output.CaptureOutput(func() {
			cmd := newListCmd()
			testutil.SetFlags(t, cmd, map[string]string{"limit": "2"})
			err := ListLocations(cmd, []string{}, true, "json")
			assert.NoError(t, err)
		})

		assert.Contains(t, out, "Sydney Data Center")
		assert.Contains(t, out, "London Data Center")
		assert.NotContains(t, out, "New York Data Center")
	})

	t.Run("NegativeLimitReturnsError", func(t *testing.T) {
		cmd := newListCmd()
		testutil.SetFlags(t, cmd, map[string]string{"limit": "-1"})
		err := ListLocations(cmd, []string{}, true, "table")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "--limit must be a non-negative integer")
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
		clientErr      error
		expectedErr    string
		expectedOutput string
	}{
		{
			name: "success",
			args: []string{"1"},
			setupMock: func(m *MockLocationsService) {
				m.ListLocationsV3Result = testLocationsV3
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
				m.ListLocationsV3Result = testLocationsV3
			},
			expectedErr: "no location found with ID: 999",
		},
		{
			name: "API error",
			args: []string{"1"},
			setupMock: func(m *MockLocationsService) {
				m.ListLocationsV3Err = fmt.Errorf("API failure")
			},
			expectedErr: "failed to list locations",
		},
		{
			name:        "client creation error",
			args:        []string{"1"},
			setupMock:   func(m *MockLocationsService) {},
			clientErr:   fmt.Errorf("config error"),
			expectedErr: "failed to create API client",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := setupTestEnvironment()
			tt.setupMock(mockSvc)

			originalFunc := config.GetNewUnauthenticatedClientFunc()
			defer func() { config.SetNewUnauthenticatedClientFunc(originalFunc) }()

			originalListLocationsFunc := listLocationsFunc
			defer func() {
				listLocationsFunc = originalListLocationsFunc
			}()

			config.SetNewUnauthenticatedClientFunc(func() (*megaport.Client, error) {
				if tt.clientErr != nil {
					return nil, tt.clientErr
				}
				testClient := &megaport.Client{}
				testClient.LocationService = mockSvc
				return testClient, nil
			})

			listLocationsFunc = func(ctx context.Context, client *megaport.Client, statuses []string) ([]*megaport.LocationV3, error) {
				return client.LocationService.ListLocationsV3(ctx)
			}

			cmd := testutil.NewCommand("get", nil)

			defer output.SetOutputFormat("table")

			var err error
			capturedOutput := output.CaptureStdout(func() {
				err = GetLocation(cmd, tt.args, true, "json")
			})

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
				// Parse failures must not leak strconv internals to the user.
				assert.NotContains(t, err.Error(), "strconv")
				assert.NotContains(t, capturedOutput, "strconv")
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
		clientErr      error
		expectedErr    string
		expectedOutput string
	}{
		{
			name: "success",
			setupMock: func(m *MockLocationsService) {
				m.ListCountriesResult = testCountries
			},
			expectedOutput: "Australia",
		},
		{
			name: "API error",
			setupMock: func(m *MockLocationsService) {
				m.ListCountriesErr = fmt.Errorf("API failure")
			},
			expectedErr: "failed to list countries",
		},
		{
			name: "empty result",
			setupMock: func(m *MockLocationsService) {
				m.ListCountriesResult = []*megaport.Country{}
			},
			expectedOutput: "[]",
		},
		{
			name:        "client creation error",
			setupMock:   func(m *MockLocationsService) {},
			clientErr:   fmt.Errorf("config error"),
			expectedErr: "failed to create API client",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := setupTestEnvironment()
			tt.setupMock(mockSvc)

			originalFunc := config.GetNewUnauthenticatedClientFunc()
			defer func() { config.SetNewUnauthenticatedClientFunc(originalFunc) }()

			originalListCountriesFunc := listCountriesFunc
			defer func() {
				listCountriesFunc = originalListCountriesFunc
			}()

			config.SetNewUnauthenticatedClientFunc(func() (*megaport.Client, error) {
				if tt.clientErr != nil {
					return nil, tt.clientErr
				}
				testClient := &megaport.Client{}
				testClient.LocationService = mockSvc
				return testClient, nil
			})

			listCountriesFunc = func(ctx context.Context, client *megaport.Client) ([]*megaport.Country, error) {
				return client.LocationService.ListCountries(ctx)
			}

			cmd := testutil.NewCommand("list-countries", nil)
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
		clientErr      error
		expectedErr    string
		expectedOutput string
	}{
		{
			name: "success",
			setupMock: func(m *MockLocationsService) {
				m.ListMarketCodesResult = testMarketCodes
			},
			expectedOutput: "AU",
		},
		{
			name: "API error",
			setupMock: func(m *MockLocationsService) {
				m.ListMarketCodesErr = fmt.Errorf("API failure")
			},
			expectedErr: "failed to list market codes",
		},
		{
			name: "empty result",
			setupMock: func(m *MockLocationsService) {
				m.ListMarketCodesResult = []string{}
			},
			expectedOutput: "[]",
		},
		{
			name: "filters blank market codes",
			setupMock: func(m *MockLocationsService) {
				m.ListMarketCodesResult = []string{"AU", "", "  ", "US"}
			},
			expectedOutput: `[
  {
    "market_code": "AU"
  },
  {
    "market_code": "US"
  }
]`,
		},
		{
			name:        "client creation error",
			setupMock:   func(m *MockLocationsService) {},
			clientErr:   fmt.Errorf("config error"),
			expectedErr: "failed to create API client",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := setupTestEnvironment()
			tt.setupMock(mockSvc)

			originalFunc := config.GetNewUnauthenticatedClientFunc()
			defer func() { config.SetNewUnauthenticatedClientFunc(originalFunc) }()

			originalListMarketCodesFunc := listMarketCodesFunc
			defer func() {
				listMarketCodesFunc = originalListMarketCodesFunc
			}()

			config.SetNewUnauthenticatedClientFunc(func() (*megaport.Client, error) {
				if tt.clientErr != nil {
					return nil, tt.clientErr
				}
				testClient := &megaport.Client{}
				testClient.LocationService = mockSvc
				return testClient, nil
			})

			listMarketCodesFunc = func(ctx context.Context, client *megaport.Client) ([]string, error) {
				return client.LocationService.ListMarketCodes(ctx)
			}

			cmd := testutil.NewCommand("list-market-codes", nil)
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
		clientErr      error
		expectedErr    string
		expectedOutput string
	}{
		{
			name: "success",
			args: []string{"Equinix"},
			setupMock: func(m *MockLocationsService) {
				m.GetLocationByNameFuzzyV3Result = testLocationsV3
			},
			expectedOutput: "Equinix SY1",
		},
		{
			name: "no matches",
			args: []string{"NonExistent"},
			setupMock: func(m *MockLocationsService) {
				m.GetLocationByNameFuzzyV3Result = []*megaport.LocationV3{}
			},
			expectedOutput: "[]",
		},
		{
			name: "API error",
			args: []string{"Equinix"},
			setupMock: func(m *MockLocationsService) {
				m.GetLocationByNameFuzzyV3Err = fmt.Errorf("API failure")
			},
			expectedErr: "failed to search locations",
		},
		{
			name:        "client creation error",
			args:        []string{"Equinix"},
			setupMock:   func(m *MockLocationsService) {},
			clientErr:   fmt.Errorf("config error"),
			expectedErr: "failed to create API client",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := setupTestEnvironment()
			tt.setupMock(mockSvc)

			originalFunc := config.GetNewUnauthenticatedClientFunc()
			defer func() { config.SetNewUnauthenticatedClientFunc(originalFunc) }()

			originalSearchLocationsFunc := searchLocationsFunc
			defer func() {
				searchLocationsFunc = originalSearchLocationsFunc
			}()

			config.SetNewUnauthenticatedClientFunc(func() (*megaport.Client, error) {
				if tt.clientErr != nil {
					return nil, tt.clientErr
				}
				testClient := &megaport.Client{}
				testClient.LocationService = mockSvc
				return testClient, nil
			})

			searchLocationsFunc = func(ctx context.Context, client *megaport.Client, search string) ([]*megaport.LocationV3, error) {
				return client.LocationService.GetLocationByNameFuzzyV3(ctx, search)
			}

			cmd := testutil.NewCommand("search", nil)
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

func TestListLocationsStatusFilter(t *testing.T) {
	allLocations := []*megaport.LocationV3{
		{ID: 1, Name: "Active DC", Status: megaport.LocationStatusActive},
		{ID: 2, Name: "Restricted DC", Status: megaport.LocationStatusRestricted},
		{ID: 3, Name: "Deployment DC", Status: megaport.LocationStatusDeployment},
	}

	originalFunc := config.GetNewUnauthenticatedClientFunc()
	defer func() { config.SetNewUnauthenticatedClientFunc(originalFunc) }()
	config.SetNewUnauthenticatedClientFunc(func() (*megaport.Client, error) {
		return &megaport.Client{}, nil
	})

	originalListLocationsFunc := listLocationsFunc
	defer func() { listLocationsFunc = originalListLocationsFunc }()

	// The stub filters the way the API does: no statuses means its default set.
	var sentStatuses []string
	listLocationsFunc = func(ctx context.Context, client *megaport.Client, statuses []string) ([]*megaport.LocationV3, error) {
		sentStatuses = statuses
		if len(statuses) == 0 {
			return allLocations, nil
		}
		var matched []*megaport.LocationV3
		for _, loc := range allLocations {
			for _, s := range statuses {
				if loc.Status == s {
					matched = append(matched, loc)
				}
			}
		}
		return matched, nil
	}

	defer output.SetOutputFormat("table")

	tests := []struct {
		name         string
		statusFlags  []string
		wantStatuses []string
		wantNames    []string
	}{
		{
			name:         "default is Active",
			wantStatuses: []string{"Active"},
			wantNames:    []string{"Active DC"},
		},
		{
			name:         "repeated status flag",
			statusFlags:  []string{"Restricted", "Deployment"},
			wantStatuses: []string{"Restricted", "Deployment"},
			wantNames:    []string{"Restricted DC", "Deployment DC"},
		},
		{
			name:         "lowercase status is normalized",
			statusFlags:  []string{"restricted"},
			wantStatuses: []string{"Restricted"},
			wantNames:    []string{"Restricted DC"},
		},
		{
			name:         "empty status falls back to Active",
			statusFlags:  []string{""},
			wantStatuses: []string{"Active"},
			wantNames:    []string{"Active DC"},
		},
		{
			name:         "blank statuses are skipped",
			statusFlags:  []string{"", " ", "Restricted"},
			wantStatuses: []string{"Restricted"},
			wantNames:    []string{"Restricted DC"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "list"}
			cmd.Flags().StringArray("status", nil, "")
			cmd.Flags().Int("limit", 0, "")
			for _, s := range tt.statusFlags {
				require.NoError(t, cmd.Flags().Set("status", s))
			}

			var err error
			out := output.CaptureStdout(func() {
				err = ListLocations(cmd, nil, true, "json")
			})
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatuses, sentStatuses)

			var parsed []map[string]interface{}
			require.NoError(t, json.Unmarshal([]byte(out), &parsed))
			var names []string
			for _, p := range parsed {
				names = append(names, fmt.Sprint(p["name"]))
			}
			assert.ElementsMatch(t, tt.wantNames, names)
		})
	}

	t.Run("status info line prints only when --status is set", func(t *testing.T) {
		for _, set := range []bool{false, true} {
			cmd := &cobra.Command{Use: "list"}
			cmd.Flags().StringArray("status", nil, "")
			cmd.Flags().Int("limit", 0, "")
			if set {
				require.NoError(t, cmd.Flags().Set("status", "Active"))
			}
			out := output.CaptureOutput(func() {
				require.NoError(t, ListLocations(cmd, nil, true, "table"))
			})
			assert.Equal(t, set, strings.Contains(out, "Filtering by status"))
		}
	})

	t.Run("get resolves a non-Active location", func(t *testing.T) {
		var err error
		out := output.CaptureStdout(func() {
			err = GetLocation(testutil.NewCommand("get", nil), []string{"2"}, true, "json")
		})
		require.NoError(t, err)
		assert.Nil(t, sentStatuses)
		assert.Contains(t, out, "Restricted DC")
	})
}
