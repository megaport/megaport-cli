package partners

import (
	"context"

	megaport "github.com/megaport/megaportgo"
)

// mockPartnerService implements the PartnerService interface for testing
type mockPartnerService struct {
	listPartnersResponse []*megaport.PartnerMegaport
	listPartnersErr      error
	filterResponse       []*megaport.PartnerMegaport
	filterErr            error
}

func (m *mockPartnerService) ListPartnerMegaports(ctx context.Context) ([]*megaport.PartnerMegaport, error) {
	return m.listPartnersResponse, m.listPartnersErr
}

func (m *mockPartnerService) FilterPartnerMegaportByProductName(ctx context.Context, partners []*megaport.PartnerMegaport, productName string, exactMatch bool) ([]*megaport.PartnerMegaport, error) {
	return m.filterResponse, m.filterErr
}

func (m *mockPartnerService) FilterPartnerMegaportByConnectType(ctx context.Context, partners []*megaport.PartnerMegaport, connectType string, exactMatch bool) ([]*megaport.PartnerMegaport, error) {
	return m.filterResponse, m.filterErr
}

func (m *mockPartnerService) FilterPartnerMegaportByCompanyName(ctx context.Context, partners []*megaport.PartnerMegaport, companyName string, exactMatch bool) ([]*megaport.PartnerMegaport, error) {
	return m.filterResponse, m.filterErr
}

func (m *mockPartnerService) FilterPartnerMegaportByLocationId(ctx context.Context, partners []*megaport.PartnerMegaport, locationId int) ([]*megaport.PartnerMegaport, error) {
	return m.filterResponse, m.filterErr
}

func (m *mockPartnerService) FilterPartnerMegaportByDiversityZone(ctx context.Context, partners []*megaport.PartnerMegaport, diversityZone string, exactMatch bool) ([]*megaport.PartnerMegaport, error) {
	return m.filterResponse, m.filterErr
}
