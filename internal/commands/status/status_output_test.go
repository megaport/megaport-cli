package status

import (
	"io"
	"os"
	"testing"

	op "github.com/megaport/megaport-cli/internal/base/output"
	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
)

// printDashboard's helpers (PrintPlain/PrintWarning/PrintNewline) route to
// stdout or stderr based on output's *global* format, which only StatusDashboard
// sets. Calling printDashboard directly leaves that global state to whatever a
// prior test set, so pin it here to keep CaptureOutput (stdout-only) reliable.
func withOutputFormat(t *testing.T, format string) {
	t.Helper()
	prev := op.GetOutputFormat()
	op.SetOutputFormat(format)
	t.Cleanup(func() { op.SetOutputFormat(prev) })
}

func statusTestDashboard(t *testing.T) dashboardOutput {
	t.Helper()
	port := &megaport.Port{UID: "port-1", Name: "Port One", ProvisioningStatus: "LIVE", PortSpeed: 1000, LocationID: 1}
	mcr := &megaport.MCR{UID: "mcr-1", Name: "MCR One", ProvisioningStatus: "LIVE", PortSpeed: 5000}
	mcr.Resources.VirtualRouter.ASN = 65000
	mve := &megaport.MVE{UID: "mve-1", Name: "MVE One", ProvisioningStatus: "LIVE", Vendor: "cisco", Size: "MEDIUM"}
	vxc := &megaport.VXC{UID: "vxc-1", Name: "VXC One", ProvisioningStatus: "LIVE", RateLimit: 500}
	vxc.AEndConfiguration.UID = "a-end"
	vxc.BEndConfiguration.UID = "b-end"
	ix := &megaport.IX{ProductUID: "ix-1", ProductName: "IX One", ProvisioningStatus: "LIVE", ASN: 64512, RateLimit: 100}

	dashboard, err := buildDashboard(
		[]*megaport.Port{port},
		[]*megaport.MCR{mcr},
		[]*megaport.MVE{mve},
		[]*megaport.VXC{vxc},
		[]*megaport.IX{ix},
	)
	assert.NoError(t, err)
	return dashboard
}

func TestBuildDashboard_FieldMapping(t *testing.T) {
	dashboard := statusTestDashboard(t)

	assert.Len(t, dashboard.Ports, 1)
	assert.Equal(t, "port-1", dashboard.Ports[0].UID)
	assert.Equal(t, 1000, dashboard.Ports[0].Speed)

	assert.Len(t, dashboard.MCRs, 1)
	assert.Equal(t, 65000, dashboard.MCRs[0].ASN)

	assert.Len(t, dashboard.MVEs, 1)
	assert.Equal(t, "cisco", dashboard.MVEs[0].Vendor)
	assert.Equal(t, "MEDIUM", dashboard.MVEs[0].Size)

	assert.Len(t, dashboard.VXCs, 1)
	assert.Equal(t, "a-end", dashboard.VXCs[0].AEndUID)
	assert.Equal(t, "b-end", dashboard.VXCs[0].BEndUID)

	assert.Len(t, dashboard.IXs, 1)
	assert.Equal(t, "ix-1", dashboard.IXs[0].UID)
	assert.Equal(t, 64512, dashboard.IXs[0].ASN)

	assert.Equal(t, dashboardSummary{Ports: 1, MCRs: 1, MVEs: 1, VXCs: 1, IXs: 1}, dashboard.Summary)
}

func TestPrintDashboard_Table(t *testing.T) {
	withOutputFormat(t, "table")
	dashboard := statusTestDashboard(t)
	out := op.CaptureOutput(func() {
		err := printDashboard(os.Stdout, dashboard, "table", true)
		assert.NoError(t, err)
	})

	assert.Contains(t, out, "PORTS (1)")
	assert.Contains(t, out, "port-1")
	assert.Contains(t, out, "MCRS (1)")
	assert.Contains(t, out, "mcr-1")
	assert.Contains(t, out, "ix-1")
	assert.Contains(t, out, "Total: 1 port(s), 1 MCR(s), 1 MVE(s), 1 VXC(s), 1 IX(s)")
}

func TestPrintDashboard_JSON(t *testing.T) {
	withOutputFormat(t, "json")
	dashboard := statusTestDashboard(t)
	out := op.CaptureOutput(func() {
		err := printDashboard(os.Stdout, dashboard, "json", true)
		assert.NoError(t, err)
	})

	assert.Contains(t, out, `"uid": "port-1"`)
	assert.Contains(t, out, `"asn": 65000`)
	assert.Contains(t, out, `"a_end_uid": "a-end"`)
	assert.Contains(t, out, `"summary"`)
}

func TestPrintDashboard_CSV(t *testing.T) {
	withOutputFormat(t, "csv")
	dashboard := statusTestDashboard(t)
	out := op.CaptureOutput(func() {
		err := printDashboard(os.Stdout, dashboard, "csv", true)
		assert.NoError(t, err)
	})

	assert.Contains(t, out, "# PORTS")
	assert.Contains(t, out, "port-1")
	assert.Contains(t, out, "# IXS")
	assert.Contains(t, out, "ix-1")
}

func TestPrintDashboard_XML(t *testing.T) {
	withOutputFormat(t, "xml")
	dashboard := statusTestDashboard(t)
	out := op.CaptureOutput(func() {
		err := printDashboard(os.Stdout, dashboard, "xml", true)
		assert.NoError(t, err)
	})

	assert.Contains(t, out, "<dashboard>")
	assert.Contains(t, out, "<uid>port-1</uid>")
	assert.Contains(t, out, "<asn>65000</asn>")
	assert.Contains(t, out, "<a_end_uid>a-end</a_end_uid>")
	assert.Contains(t, out, "<summary>")
}

// TestPrintDashboardTable_SectionError verifies that a table-render failure in
// any section is propagated. Setting an unknown --fields makes
// PrintTableToWriter fail, and each subtest populates exactly one section so the
// error originates from that section's render call.
func TestPrintDashboardTable_SectionError(t *testing.T) {
	withOutputFormat(t, "table")

	port := &megaport.Port{UID: "port-1", Name: "Port One", ProvisioningStatus: "LIVE"}
	mcr := &megaport.MCR{UID: "mcr-1", Name: "MCR One", ProvisioningStatus: "LIVE"}
	mve := &megaport.MVE{UID: "mve-1", Name: "MVE One", ProvisioningStatus: "LIVE"}
	vxc := &megaport.VXC{UID: "vxc-1", Name: "VXC One", ProvisioningStatus: "LIVE"}
	ix := &megaport.IX{ProductUID: "ix-1", ProductName: "IX One", ProvisioningStatus: "LIVE"}

	cases := []struct {
		name  string
		ports []*megaport.Port
		mcrs  []*megaport.MCR
		mves  []*megaport.MVE
		vxcs  []*megaport.VXC
		ixs   []*megaport.IX
	}{
		{name: "ports", ports: []*megaport.Port{port}},
		{name: "mcrs", mcrs: []*megaport.MCR{mcr}},
		{name: "mves", mves: []*megaport.MVE{mve}},
		{name: "vxcs", vxcs: []*megaport.VXC{vxc}},
		{name: "ixs", ixs: []*megaport.IX{ix}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dashboard, err := buildDashboard(tc.ports, tc.mcrs, tc.mves, tc.vxcs, tc.ixs)
			assert.NoError(t, err)

			op.SetOutputFields([]string{"definitely_not_a_field"})
			t.Cleanup(func() { op.SetOutputFields(nil) })

			err = printDashboard(io.Discard, dashboard, "table", true)
			assert.Error(t, err)
		})
	}
}

func TestPrintDashboard_Empty(t *testing.T) {
	dashboard, err := buildDashboard(nil, nil, nil, nil, nil)
	assert.NoError(t, err)

	for _, format := range []string{"table", "json", "csv", "xml"} {
		t.Run(format, func(t *testing.T) {
			withOutputFormat(t, format)
			out := op.CaptureOutput(func() {
				err := printDashboard(os.Stdout, dashboard, format, true)
				assert.NoError(t, err)
			})
			assert.NotEmpty(t, out)
		})
	}
}
