package vxc

import (
	"testing"

	op "github.com/megaport/megaport-cli/internal/base/output"
	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
)

func TestToVXCOutput_Valid(t *testing.T) {
	v := &megaport.VXC{
		UID:       "vxc-abc",
		Name:      "Test VXC",
		RateLimit: 500,
		AEndConfiguration: megaport.VXCEndConfiguration{
			UID:  "port-a",
			VLAN: 100,
		},
		BEndConfiguration: megaport.VXCEndConfiguration{
			UID:  "port-b",
			VLAN: 200,
		},
		ProvisioningStatus: "LIVE",
	}

	out, err := toVXCOutput(v)
	assert.NoError(t, err)
	assert.Equal(t, "vxc-abc", out.UID)
	assert.Equal(t, "Test VXC", out.Name)
	assert.Equal(t, 500, out.RateLimit)
	assert.Equal(t, "port-a", out.AEndUID)
	assert.Equal(t, "port-b", out.BEndUID)
	assert.Equal(t, 100, out.AEndVLAN)
	assert.Equal(t, 200, out.BEndVLAN)
	assert.Equal(t, "LIVE", out.Status)
}

func TestPrintVXCs_XML(t *testing.T) {
	vxcs := []*megaport.VXC{
		{
			UID:                "vxc-xml-1",
			Name:               "XMLTestVXC",
			RateLimit:          1000,
			ProvisioningStatus: "LIVE",
		},
	}

	out := op.CaptureOutput(func() {
		err := printVXCs(vxcs, "xml", true)
		assert.NoError(t, err)
	})

	assert.NotEmpty(t, out)
	assert.Contains(t, out, "vxc-xml-1")
	assert.Contains(t, out, "XMLTestVXC")
}
