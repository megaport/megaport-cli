//nolint:staticcheck // V2 methods use deprecated megaport.Location to satisfy the interface
package locations

import (
	"context"
	"fmt"

	megaport "github.com/megaport/megaportgo"
)

type MockLocationsService struct {
	// V3 methods
	ListLocationsV3Result []*megaport.LocationV3
	ListLocationsV3Err    error

	GetLocationByIDV3Result *megaport.LocationV3
	GetLocationByIDV3Err    error

	GetLocationByNameV3Result *megaport.LocationV3
	GetLocationByNameV3Err    error

	GetLocationByNameFuzzyV3Result []*megaport.LocationV3
	GetLocationByNameFuzzyV3Err    error

	FilterLocationsByMarketCodeV3Result []*megaport.LocationV3
	FilterLocationsByMarketCodeV3Err    error

	FilterLocationsByMcrAvailabilityV3Result []*megaport.LocationV3

	FilterLocationsByMetroV3Result []*megaport.LocationV3

	// V2 methods
	ListLocationsResult []*megaport.Location
	ListLocationsErr    error

	GetLocationByIDResult *megaport.Location
	GetLocationByIDErr    error

	GetLocationByNameResult *megaport.Location
	GetLocationByNameErr    error

	GetLocationByNameFuzzyResult []*megaport.Location
	GetLocationByNameFuzzyErr    error

	FilterLocationsByMarketCodeResult []*megaport.Location
	FilterLocationsByMarketCodeErr    error

	FilterLocationsByMcrAvailabilityResult []*megaport.Location

	// Shared methods
	ListCountriesResult []*megaport.Country
	ListCountriesErr    error

	ListMarketCodesResult []string
	ListMarketCodesErr    error

	IsValidMarketCodeResult bool
	IsValidMarketCodeErr    error
}

// V2 methods

func (m *MockLocationsService) ListLocations(ctx context.Context) ([]*megaport.Location, error) {
	return m.ListLocationsResult, m.ListLocationsErr
}

func (m *MockLocationsService) GetLocationByID(ctx context.Context, locationID int) (*megaport.Location, error) {
	return m.GetLocationByIDResult, m.GetLocationByIDErr
}

func (m *MockLocationsService) GetLocationByName(ctx context.Context, locationName string) (*megaport.Location, error) {
	return m.GetLocationByNameResult, m.GetLocationByNameErr
}

func (m *MockLocationsService) GetLocationByNameFuzzy(ctx context.Context, search string) ([]*megaport.Location, error) {
	return m.GetLocationByNameFuzzyResult, m.GetLocationByNameFuzzyErr
}

func (m *MockLocationsService) FilterLocationsByMarketCode(ctx context.Context, marketCode string, locations []*megaport.Location) ([]*megaport.Location, error) {
	return m.FilterLocationsByMarketCodeResult, m.FilterLocationsByMarketCodeErr
}

func (m *MockLocationsService) FilterLocationsByMcrAvailability(ctx context.Context, mcrAvailable bool, locations []*megaport.Location) []*megaport.Location {
	return m.FilterLocationsByMcrAvailabilityResult
}

// V3 methods

func (m *MockLocationsService) ListLocationsV3(ctx context.Context) ([]*megaport.LocationV3, error) {
	return m.ListLocationsV3Result, m.ListLocationsV3Err
}

func (m *MockLocationsService) GetLocationByIDV3(ctx context.Context, locationID int) (*megaport.LocationV3, error) {
	return m.GetLocationByIDV3Result, m.GetLocationByIDV3Err
}

func (m *MockLocationsService) GetLocationByNameV3(ctx context.Context, locationName string) (*megaport.LocationV3, error) {
	return m.GetLocationByNameV3Result, m.GetLocationByNameV3Err
}

func (m *MockLocationsService) GetLocationByNameFuzzyV3(ctx context.Context, search string) ([]*megaport.LocationV3, error) {
	return m.GetLocationByNameFuzzyV3Result, m.GetLocationByNameFuzzyV3Err
}

func (m *MockLocationsService) FilterLocationsByMarketCodeV3(ctx context.Context, marketCode string, locations []*megaport.LocationV3) ([]*megaport.LocationV3, error) {
	return m.FilterLocationsByMarketCodeV3Result, m.FilterLocationsByMarketCodeV3Err
}

func (m *MockLocationsService) FilterLocationsByMcrAvailabilityV3(ctx context.Context, mcrAvailable bool, locations []*megaport.LocationV3) []*megaport.LocationV3 {
	return m.FilterLocationsByMcrAvailabilityV3Result
}

func (m *MockLocationsService) FilterLocationsByMetroV3(ctx context.Context, metro string, locations []*megaport.LocationV3) []*megaport.LocationV3 {
	return m.FilterLocationsByMetroV3Result
}

// Shared methods

func (m *MockLocationsService) ListCountries(ctx context.Context) ([]*megaport.Country, error) {
	return m.ListCountriesResult, m.ListCountriesErr
}

func (m *MockLocationsService) ListMarketCodes(ctx context.Context) ([]string, error) {
	return m.ListMarketCodesResult, m.ListMarketCodesErr
}

func (m *MockLocationsService) IsValidMarketCode(ctx context.Context, marketCode string) (bool, error) {
	return m.IsValidMarketCodeResult, m.IsValidMarketCodeErr
}

func (m *MockLocationsService) GetRoundTripTimes(_ context.Context, _, _, _ int) ([]*megaport.RoundTripTime, error) {
	return nil, fmt.Errorf("mock: GetRoundTripTimes not configured")
}
