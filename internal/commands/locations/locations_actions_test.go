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

func TestMockSetup(t *testing.T) {
	mockSvc := new(MockLocationsService)
	testLocs := []*megaport.Location{
		{
			ID:      1,
			Name:    "Test Location 1",
			Country: "Australia",
			Metro:   "Sydney",
		},
	}

	mockSvc.On("ListLocations", mock.Anything).Return(testLocs, nil)

	locations, err := mockSvc.ListLocations(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, 1, len(locations))
	assert.Equal(t, "Test Location 1", locations[0].Name)
	mockSvc.AssertExpectations(t)
}

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

func TestListLocationsFunc(t *testing.T) {
	mockSvc, testLocations := setupTestEnvironment()

	mockSvc.On("ListLocations", mock.Anything).Return(testLocations, nil)

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

	listLocationsFunc = func(ctx context.Context, client *megaport.Client) ([]*megaport.Location, error) {
		return client.LocationService.ListLocations(ctx)
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
	mockSvc, _ := setupTestEnvironment()

	expectedError := errors.New("api connection failed")

	mockSvc.On("ListLocations", mock.Anything).Return([]*megaport.Location{}, expectedError)

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

	listLocationsFunc = func(ctx context.Context, client *megaport.Client) ([]*megaport.Location, error) {
		return client.LocationService.ListLocations(ctx)
	}

	locations, err := listLocationsFunc(context.Background(), testClient)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Empty(t, locations)

	mockSvc.AssertExpectations(t)
}

func TestListLocationsCommand(t *testing.T) {
	mockSvc, testLocations := setupTestEnvironment()

	mockSvc.On("ListLocations", mock.Anything).Return(testLocations, nil)

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

	listLocationsFunc = func(ctx context.Context, client *megaport.Client) ([]*megaport.Location, error) {
		return client.LocationService.ListLocations(ctx)
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
