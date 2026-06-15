package ix

import (
	"testing"

	"github.com/megaport/megaport-cli/internal/base/output"
	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
)

func TestPrintIXs_XML(t *testing.T) {
	ixs := []*megaport.IX{
		{
			ProductUID:         "ix-xml-1",
			ProductName:        "XML Test IX",
			NetworkServiceType: "Tokyo IX",
			ASN:                65100,
			RateLimit:          1000,
			VLAN:               10,
			MACAddress:         "AA:BB:CC:DD:EE:01",
			ProvisioningStatus: "LIVE",
		},
	}

	out := output.CaptureOutput(func() {
		err := printIXs(ixs, "xml", true)
		assert.NoError(t, err)
	})

	assert.Contains(t, out, "<items>")
	assert.Contains(t, out, "<item>")
	assert.Contains(t, out, "ix-xml-1")
	assert.Contains(t, out, "XML Test IX")
}

func TestPrintIXs_XML_List(t *testing.T) {
	ixs := []*megaport.IX{
		{
			ProductUID:         "ix-a",
			ProductName:        "First IX",
			NetworkServiceType: "London IX",
			ProvisioningStatus: "LIVE",
		},
		{
			ProductUID:         "ix-b",
			ProductName:        "Second IX",
			NetworkServiceType: "Paris IX",
			ProvisioningStatus: "CONFIGURED",
		},
	}

	out := output.CaptureOutput(func() {
		err := printIXs(ixs, "xml", true)
		assert.NoError(t, err)
	})

	assert.Contains(t, out, "<items>")
	assert.Contains(t, out, "ix-a")
	assert.Contains(t, out, "ix-b")
}

func TestPrintIXs_XML_Empty(t *testing.T) {
	out := output.CaptureOutput(func() {
		err := printIXs([]*megaport.IX{}, "xml", true)
		assert.NoError(t, err)
	})

	assert.Contains(t, out, "<items>")
}

func TestToIXOutput_FieldMapping(t *testing.T) {
	ix := &megaport.IX{
		ProductUID:         "uid-map",
		ProductName:        "Mapping IX",
		NetworkServiceType: "Miami IX",
		ASN:                65200,
		RateLimit:          2000,
		VLAN:               300,
		MACAddress:         "11:22:33:44:55:66",
		ProvisioningStatus: "CONFIGURED",
	}

	out, err := toIXOutput(ix)
	assert.NoError(t, err)
	assert.Equal(t, "uid-map", out.UID)
	assert.Equal(t, "Mapping IX", out.Name)
	assert.Equal(t, "Miami IX", out.NetworkServiceType)
	assert.Equal(t, 65200, out.ASN)
	assert.Equal(t, 2000, out.RateLimit)
	assert.Equal(t, 300, out.VLAN)
	assert.Equal(t, "11:22:33:44:55:66", out.MACAddress)
	assert.Equal(t, "CONFIGURED", out.Status)
}

func TestDisplayIXChanges_NilSafe(t *testing.T) {
	// Should not panic with nil inputs
	assert.NotPanics(t, func() { displayIXChanges(nil, nil, true) })
	assert.NotPanics(t, func() { displayIXChanges(&megaport.IX{}, nil, true) })
	assert.NotPanics(t, func() { displayIXChanges(nil, &megaport.IX{}, true) })
}

func TestDisplayIXChanges_ShowsDiff(t *testing.T) {
	original := &megaport.IX{ProductName: "Old IX", RateLimit: 1000, VLAN: 10, MACAddress: "00:00:00:00:00:01", ASN: 65000}
	updated := &megaport.IX{ProductName: "New IX", RateLimit: 2000, VLAN: 20, MACAddress: "00:00:00:00:00:02", ASN: 65001}

	out := output.CaptureOutput(func() {
		displayIXChanges(original, updated, true)
	})

	assert.Contains(t, out, "Name")
	assert.Contains(t, out, "Old IX")
	assert.Contains(t, out, "New IX")
}
