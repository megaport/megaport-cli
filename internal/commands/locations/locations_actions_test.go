package locations

import (
	"context"
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/commands/config"
	megaport "github.com/megaport/megaportgo"
)

// TestMockSetup is a simple test to verify that our mock is properly implemented
func TestMockSetup(t *testing.T) {
	mockSvc := new(MockLocationsService)

	// Set up expected return values
	testLocs := []*megaport.Location{
		{
			ID:      1,
			Name:    "Test Location 1",
			Country: "Australia",
			Metro:   "Sydney",
		},
	}

	mockSvc.On("ListLocations", mock.Anything).Return(testLocs, nil)

	// Call the mock method
	locations, err := mockSvc.ListLocations(context.Background())

	// Assert expectations
	assert.NoError(t, err)
	assert.Equal(t, 1, len(locations))
	assert.Equal(t, "Test Location 1", locations[0].Name)
	mockSvc.AssertExpectations(t)
}

// setupTestEnvironment prepares test data and mock service
func setupTestEnvironment() (*MockLocationsService, []*megaport.Location) {
	mockSvc := new(MockLocationsService)

	testLocations := []*megaport.Location{
		{
			ID:       1,
			Name:     "Sydney Data Center",
			Country:  "Australia",
			Metro:    "Sydney",
			SiteCode: "SYD",
			Status:   "Active",
			Products: &megaport.LocationProducts{
				MCR:      true,
				Megaport: []int{1, 10},
			},
		},
		{
			ID:       2,
			Name:     "London Data Center",
			Country:  "United Kingdom",
			Metro:    "London",
			SiteCode: "LON",
			Status:   "Active",
			Products: &megaport.LocationProducts{
				MCR:      false,
				Megaport: []int{1},
			},
		},
		{
			ID:       3,
			Name:     "New York Data Center",
			Country:  "USA",
			Metro:    "New York",
			SiteCode: "NYC",
			Status:   "Active",
			Products: &megaport.LocationProducts{
				MCR:      true,
				Megaport: []int{10},
			},
		},
	}

	return mockSvc, testLocations
}

// TestListLocationsFunc tests the listLocationsFunc with MockLocationsService
func TestListLocationsFunc(t *testing.T) {
	mockSvc, testLocations := setupTestEnvironment()

	// Setup the mock to return the test locations
	mockSvc.On("ListLocations", mock.Anything).Return(testLocations, nil)

	originalLoginFunc := config.LoginFunc

	// Store the original function to restore it later
	originalListLocationsFunc := listLocationsFunc
	defer func() {
		// Restore the original function
		config.LoginFunc = originalLoginFunc
		listLocationsFunc = originalListLocationsFunc
	}()

	// Setup login to return our mock client
	config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
		testClient := &megaport.Client{}
		testClient.LocationService = mockSvc
		return testClient, nil
	}

	testClient, err := config.LoginFunc(context.Background())
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}

	// Replace the function with our test version that uses the mock
	listLocationsFunc = func(ctx context.Context, client *megaport.Client) ([]*megaport.Location, error) {
		return client.LocationService.ListLocations(ctx)
	}

	// Call the function
	locations, err := listLocationsFunc(context.Background(), testClient)

	// Verify the results
	assert.NoError(t, err)
	assert.Equal(t, 3, len(locations))
	assert.Equal(t, "Sydney Data Center", locations[0].Name)
	assert.Equal(t, "London Data Center", locations[1].Name)
	assert.Equal(t, "New York Data Center", locations[2].Name)

	// Verify that the mock was called as expected
	mockSvc.AssertExpectations(t)
}

// TestListLocationsFuncError tests error handling in listLocationsFunc
func TestListLocationsFuncError(t *testing.T) {
	mockSvc, _ := setupTestEnvironment()

	expectedError := errors.New("api connection failed")

	// Setup the mock to return an error
	mockSvc.On("ListLocations", mock.Anything).Return([]*megaport.Location{}, expectedError)

	// Store the original function to restore it later
	originalListLocationsFunc := listLocationsFunc
	originalLoginFunc := config.LoginFunc

	defer func() {
		// Restore the original function
		config.LoginFunc = originalLoginFunc
		listLocationsFunc = originalListLocationsFunc
	}()

	// Setup login to return our mock client
	config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
		testClient := &megaport.Client{}
		testClient.LocationService = mockSvc
		return testClient, nil
	}

	testClient, err := config.LoginFunc(context.Background())
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}

	// Replace the function with our test version that uses the mock
	listLocationsFunc = func(ctx context.Context, client *megaport.Client) ([]*megaport.Location, error) {
		return client.LocationService.ListLocations(ctx)
	}

	// Call the function
	locations, err := listLocationsFunc(context.Background(), testClient)

	// Verify the error was returned
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Empty(t, locations)

	// Verify that the mock was called as expected
	mockSvc.AssertExpectations(t)
}

// TestListLocationsCommand tests the cobra command structure for the list command
func TestListLocationsCommand(t *testing.T) {
	mockSvc, testLocations := setupTestEnvironment()

	// Setup the mock
	mockSvc.On("ListLocations", mock.Anything).Return(testLocations, nil)

	// Store the original function to restore it later
	originalListLocationsFunc := listLocationsFunc
	originalLoginFunc := config.LoginFunc

	defer func() {
		// Restore the original functions
		config.LoginFunc = originalLoginFunc
		listLocationsFunc = originalListLocationsFunc
	}()

	// Setup login to return our mock client
	config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
		testClient := &megaport.Client{}
		testClient.LocationService = mockSvc
		return testClient, nil
	}

	// Replace the function with our test version that uses the mock
	listLocationsFunc = func(ctx context.Context, client *megaport.Client) ([]*megaport.Location, error) {
		return client.LocationService.ListLocations(ctx)
	}

	// Test case 1: No filters - should return all locations
	t.Run("NoFilters", func(t *testing.T) {
		// Capture the output of the command
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

		// JSON output should contain all 3 locations
		assert.Contains(t, output, "Sydney Data Center")
		assert.Contains(t, output, "London Data Center")
		assert.Contains(t, output, "New York Data Center")
	})

	// Test case 2: Filter by metro
	t.Run("FilterByMetro", func(t *testing.T) {
		// Capture the output of the command
		output := output.CaptureOutput(func() {
			cmd := &cobra.Command{
				Use: "list",
			}
			cmd.Flags().String("metro", "", "Filter by metro")
			cmd.Flags().String("country", "", "Filter by country")
			cmd.Flags().String("name", "", "Filter by name")

			// Set the metro flag
			if err := cmd.Flags().Set("metro", "New York"); err != nil {
				t.Fatalf("Failed to set flag: %v", err)
			}

			err := ListLocations(cmd, []string{}, true, "json")
			assert.NoError(t, err)
		})

		// Should contain only New York
		assert.NotContains(t, output, "Sydney Data Center")
		assert.NotContains(t, output, "London Data Center")
		assert.Contains(t, output, "New York Data Center")
	})

	// Test case 3: Filter by country
	t.Run("FilterByCountry", func(t *testing.T) {
		// Capture the output of the command
		output := output.CaptureOutput(func() {
			cmd := &cobra.Command{
				Use: "list",
			}
			cmd.Flags().String("metro", "", "Filter by metro")
			cmd.Flags().String("country", "", "Filter by country")
			cmd.Flags().String("name", "", "Filter by name")

			// Set the country flag
			if err := cmd.Flags().Set("country", "United Kingdom"); err != nil {
				t.Fatalf("Failed to set flag: %v", err)
			}

			err := ListLocations(cmd, []string{}, true, "json")
			assert.NoError(t, err)
		})

		// Should contain only London
		assert.NotContains(t, output, "Sydney Data Center")
		assert.Contains(t, output, "London Data Center")
		assert.NotContains(t, output, "New York Data Center")
	})

	// Test case 4: Filter by name
	t.Run("FilterByName", func(t *testing.T) {
		// Capture the output of the command
		output := output.CaptureOutput(func() {
			cmd := &cobra.Command{
				Use: "list",
			}
			cmd.Flags().String("metro", "", "Filter by metro")
			cmd.Flags().String("country", "", "Filter by country")
			cmd.Flags().String("name", "", "Filter by name")

			// Set the name flag
			if err := cmd.Flags().Set("name", "Sydney Data Center"); err != nil {
				t.Fatalf("Failed to set flag: %v", err)
			}

			err := ListLocations(cmd, []string{}, true, "json")
			assert.NoError(t, err)
		})

		// Should contain only Sydney
		assert.Contains(t, output, "Sydney Data Center")
		assert.NotContains(t, output, "London Data Center")
		assert.NotContains(t, output, "New York Data Center")
	})

	// Test case 5: No matching locations
	t.Run("NoMatchingLocations", func(t *testing.T) {
		// Capture the output of the command
		output := output.CaptureOutput(func() {
			cmd := &cobra.Command{
				Use: "list",
			}
			cmd.Flags().String("metro", "", "Filter by metro")
			cmd.Flags().String("country", "", "Filter by country")
			cmd.Flags().String("name", "", "Filter by name")

			// Set a filter that won't match any locations
			if err := cmd.Flags().Set("name", "Non-existent Location"); err != nil {
				t.Fatalf("Failed to set flag: %v", err)
			}

			err := ListLocations(cmd, []string{}, true, "json")
			assert.NoError(t, err)
		})

		// Should show warning about no matching locations
		assert.Contains(t, output, "No locations found matching the specified filters")
	})
}
