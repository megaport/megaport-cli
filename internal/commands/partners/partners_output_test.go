package partners

import (
	"strings"
	"testing"

	op "github.com/megaport/megaport-cli/internal/base/output"
	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
)

func TestToPartnerOutput_Valid(t *testing.T) {
	p := &megaport.PartnerMegaport{
		ProductUID:    "uid-1",
		ProductName:   "Partner One",
		ConnectType:   "AWS",
		CompanyName:   "Acme",
		LocationId:    42,
		DiversityZone: "red",
		VXCPermitted:  true,
	}

	out := toPartnerOutput(p)
	assert.Equal(t, "Partner One", out.ProductName)
	assert.Equal(t, "uid-1", out.UID)
	assert.Equal(t, "AWS", out.ConnectType)
	assert.Equal(t, "Acme", out.CompanyName)
	assert.Equal(t, 42, out.LocationId)
	assert.Equal(t, "red", out.DiversityZone)
	assert.True(t, out.VXCPermitted)
}

func TestToPartnerOutput_Nil(t *testing.T) {
	assert.NotPanics(t, func() {
		out := toPartnerOutput(nil)
		assert.Equal(t, partnerOutput{}, out)
	})
}

func TestPrintPartners_XML(t *testing.T) {
	out := op.CaptureOutput(func() {
		err := printPartnersFunc(testPartners, "xml", noColor)
		assert.NoError(t, err)
	})

	assert.Contains(t, out, "<product_name>ProductOne</product_name>")
	assert.Contains(t, out, "<uid>uid1</uid>")
	assert.Contains(t, out, "<connect_type>TypeA</connect_type>")
	assert.Contains(t, out, "<company_name>CompanyA</company_name>")
	assert.Contains(t, out, "<location_id>1</location_id>")
	assert.Contains(t, out, "<diversity_zone>ZoneA</diversity_zone>")
	assert.Contains(t, out, "<vxc_permitted>true</vxc_permitted>")
	assert.Contains(t, out, "ProductTwo")
}

func TestPrintPartners_NilEntriesSkipped(t *testing.T) {
	partners := []*megaport.PartnerMegaport{
		nil,
		{ProductUID: "uid-keep", ProductName: "Keep"},
		nil,
	}

	out := op.CaptureOutput(func() {
		err := printPartnersFunc(partners, "csv", noColor)
		assert.NoError(t, err)
	})

	assert.Contains(t, out, "uid-keep")
	// Only one data row beyond the header.
	assert.Equal(t, 2, strings.Count(strings.TrimSpace(out), "\n")+1)
}
