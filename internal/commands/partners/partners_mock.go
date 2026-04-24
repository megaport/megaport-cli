package partners

import (
	"context"

	megaport "github.com/megaport/megaportgo"
)

type MockPartnerService struct {
	listPartnersResponse []*megaport.PartnerMegaport
	listPartnersErr      error
	filterResponse       []*megaport.PartnerMegaport
	filterErr            error
}

func (m *MockPartnerService) ListPartnerMegaports(ctx context.Context) ([]*megaport.PartnerMegaport, error) {
	return m.listPartnersResponse, m.listPartnersErr
}

func (m *MockPartnerService) FilterPartnerMegaportByProductName(ctx context.Context, partners []*megaport.PartnerMegaport, productName string, exactMatch bool) ([]*megaport.PartnerMegaport, error) {
	return m.filterResponse, m.filterErr
}

func (m *MockPartnerService) FilterPartnerMegaportByConnectType(ctx context.Context, partners []*megaport.PartnerMegaport, connectType string, exactMatch bool) ([]*megaport.PartnerMegaport, error) {
	return m.filterResponse, m.filterErr
}

func (m *MockPartnerService) FilterPartnerMegaportByCompanyName(ctx context.Context, partners []*megaport.PartnerMegaport, companyName string, exactMatch bool) ([]*megaport.PartnerMegaport, error) {
	return m.filterResponse, m.filterErr
}

func (m *MockPartnerService) FilterPartnerMegaportByLocationId(ctx context.Context, partners []*megaport.PartnerMegaport, locationId int) ([]*megaport.PartnerMegaport, error) {
	return m.filterResponse, m.filterErr
}

func (m *MockPartnerService) FilterPartnerMegaportByDiversityZone(ctx context.Context, partners []*megaport.PartnerMegaport, diversityZone string) ([]*megaport.PartnerMegaport, error) {
	return m.filterResponse, m.filterErr
}

func (m *MockPartnerService) FilterPartnerMegaportByMetro(ctx context.Context, partners []*megaport.PartnerMegaport, locationService megaport.LocationService, metro string) ([]*megaport.PartnerMegaport, error) {
	return m.filterResponse, m.filterErr
}
