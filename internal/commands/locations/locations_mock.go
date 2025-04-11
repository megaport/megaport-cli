package locations

import (
	"context"

	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/mock"
)

type MockLocationsService struct {
	mock.Mock
}

// ListLocations mocks the ListLocations method.
func (m *MockLocationsService) ListLocations(ctx context.Context) ([]*megaport.Location, error) {
	args := m.Called(ctx)
	val := args.Get(0)
	if val == nil {
		return nil, args.Error(1)
	}
	locations, ok := val.([]*megaport.Location)
	if !ok {
		return nil, args.Error(1)
	}
	return locations, args.Error(1)
}

// GetLocationByID mocks the GetLocationByID method.
func (m *MockLocationsService) GetLocationByID(ctx context.Context, locationID int) (*megaport.Location, error) {
	args := m.Called(ctx, locationID)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	location, ok := args.Get(0).(*megaport.Location)
	if !ok {
		panic("mock returned wrong type for GetLocationByID")
	}
	return location, args.Error(1)
}

// GetLocationByName mocks the GetLocationByName method.
func (m *MockLocationsService) GetLocationByName(ctx context.Context, locationName string) (*megaport.Location, error) {
	args := m.Called(ctx, locationName)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	location, ok := args.Get(0).(*megaport.Location)
	if !ok {
		panic("mock returned wrong type for GetLocationByName")
	}
	return location, args.Error(1)
}

// GetLocationByNameFuzzy mocks the GetLocationByNameFuzzy method.
func (m *MockLocationsService) GetLocationByNameFuzzy(ctx context.Context, search string) ([]*megaport.Location, error) {
	args := m.Called(ctx, search)
	val := args.Get(0)
	if val == nil {
		return nil, args.Error(1)
	}
	locations, ok := val.([]*megaport.Location)
	if !ok {
		panic("mock returned wrong type for GetLocationByNameFuzzy")
	}
	return locations, args.Error(1)
}

// ListCountries mocks the ListCountries method.
func (m *MockLocationsService) ListCountries(ctx context.Context) ([]*megaport.Country, error) {
	args := m.Called(ctx)
	val := args.Get(0)
	if val == nil {
		return nil, args.Error(1)
	}
	countries, ok := val.([]*megaport.Country)
	if !ok {
		panic("mock returned wrong type for ListCountries")
	}
	return countries, args.Error(1)
}

// ListMarketCodes mocks the ListMarketCodes method.
func (m *MockLocationsService) ListMarketCodes(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	val := args.Get(0)
	if val == nil {
		return nil, args.Error(1)
	}
	marketCodes, ok := val.([]string)
	if !ok {
		panic("mock returned wrong type for ListMarketCodes")
	}
	return marketCodes, args.Error(1)
}

// IsValidMarketCode mocks the IsValidMarketCode method.
func (m *MockLocationsService) IsValidMarketCode(ctx context.Context, marketCode string) (bool, error) {
	args := m.Called(ctx, marketCode)
	return args.Bool(0), args.Error(1)
}

// FilterLocationsByMarketCode mocks the FilterLocationsByMarketCode method.
func (m *MockLocationsService) FilterLocationsByMarketCode(ctx context.Context, marketCode string, locations []*megaport.Location) ([]*megaport.Location, error) {
	args := m.Called(ctx, marketCode, locations)
	val := args.Get(0)
	if val == nil {
		return nil, args.Error(1)
	}
	filteredLocations, ok := val.([]*megaport.Location)
	if !ok {
		panic("mock returned wrong type for FilterLocationsByMarketCode")
	}
	return filteredLocations, args.Error(1)
}

// FilterLocationsByMcrAvailability mocks the FilterLocationsByMcrAvailability method.
func (m *MockLocationsService) FilterLocationsByMcrAvailability(ctx context.Context, mcrAvailable bool, locations []*megaport.Location) []*megaport.Location {
	args := m.Called(ctx, mcrAvailable, locations)
	val := args.Get(0)
	if val == nil {
		return nil
	}
	filteredLocations, ok := val.([]*megaport.Location)
	if !ok {
		panic("mock returned wrong type for FilterLocationsByMcrAvailability")
	}
	return filteredLocations
}
