package product

import (
	"testing"

	op "github.com/megaport/megaport-cli/internal/base/output"
	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
)

func TestToProductOutput_Port(t *testing.T) {
	port := &megaport.Port{
		UID:                "port-1",
		Name:               "Test Port",
		Type:               "MEGAPORT",
		ProvisioningStatus: "LIVE",
		PortSpeed:          10000,
		LocationID:         5,
	}

	o, err := toProductOutput(port)
	assert.NoError(t, err)
	assert.Equal(t, "port-1", o.UID)
	assert.Equal(t, "Test Port", o.Name)
	assert.Equal(t, "LIVE", o.ProvisioningStatus)
	assert.Equal(t, 10000, o.Speed)
	assert.Equal(t, 5, o.LocationID)
}

func TestToProductOutput_MVE(t *testing.T) {
	mve := &megaport.MVE{
		UID:                "mve-1",
		Name:               "Test MVE",
		ProvisioningStatus: "LIVE",
		LocationID:         9,
	}

	o, err := toProductOutput(mve)
	assert.NoError(t, err)
	assert.Equal(t, "mve-1", o.UID)
	assert.Equal(t, "Test MVE", o.Name)
	assert.Equal(t, 9, o.LocationID)
	assert.Equal(t, 0, o.Speed)
}

var productTestProducts = []megaport.Product{
	&megaport.Port{
		UID:                "port-1",
		Name:               "Port One",
		ProvisioningStatus: "LIVE",
		PortSpeed:          1000,
		LocationID:         1,
	},
	&megaport.MCR{
		UID:                "mcr-1",
		Name:               "MCR One",
		ProvisioningStatus: "CONFIGURED",
		PortSpeed:          5000,
		LocationID:         2,
	},
}

func TestPrintProducts_Table(t *testing.T) {
	out := op.CaptureOutput(func() {
		err := printProducts(productTestProducts, "table", true)
		assert.NoError(t, err)
	})

	assert.Contains(t, out, "UID")
	assert.Contains(t, out, "STATUS")
	assert.Contains(t, out, "port-1")
	assert.Contains(t, out, "Port One")
	assert.Contains(t, out, "mcr-1")
	assert.Contains(t, out, "MCR One")
}

func TestPrintProducts_JSON(t *testing.T) {
	out := op.CaptureOutput(func() {
		err := printProducts(productTestProducts, "json", true)
		assert.NoError(t, err)
	})

	assert.Contains(t, out, `"uid": "port-1"`)
	assert.Contains(t, out, `"name": "Port One"`)
	assert.Contains(t, out, `"speed": 1000`)
	assert.Contains(t, out, `"uid": "mcr-1"`)
}

func TestPrintProducts_CSV(t *testing.T) {
	out := op.CaptureOutput(func() {
		err := printProducts(productTestProducts, "csv", true)
		assert.NoError(t, err)
	})

	assert.Contains(t, out, "uid,name,type,provisioning_status,speed,location_id")
	assert.Contains(t, out, "port-1,Port One")
	assert.Contains(t, out, "mcr-1,MCR One")
}

func TestPrintProducts_XML(t *testing.T) {
	out := op.CaptureOutput(func() {
		err := printProducts(productTestProducts, "xml", true)
		assert.NoError(t, err)
	})

	assert.Contains(t, out, "<uid>port-1</uid>")
	assert.Contains(t, out, "<name>Port One</name>")
	assert.Contains(t, out, "<speed>1000</speed>")
	assert.Contains(t, out, "<uid>mcr-1</uid>")
}

func TestPrintProducts_Empty(t *testing.T) {
	out := op.CaptureOutput(func() {
		err := printProducts([]megaport.Product{}, "json", true)
		assert.NoError(t, err)
	})
	assert.Equal(t, "[]\n", out)
}

func TestPrintProducts_NilProductReturnsError(t *testing.T) {
	var err error
	op.CaptureOutput(func() {
		err = printProducts([]megaport.Product{nil}, "table", true)
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nil value")
}
