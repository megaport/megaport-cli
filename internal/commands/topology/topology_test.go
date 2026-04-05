package topology

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/commands/config"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// ── buildTopologyNodes ────────────────────────────────────────────────────────

func TestBuildTopologyNodes_VXCDeduplication(t *testing.T) {
	// Port owns vxc-1 (A-End) and vxc-2 (B-End — should be excluded)
	port := &megaport.Port{
		UID:                "port-aaa",
		Name:               "Sydney-Primary",
		ProvisioningStatus: "LIVE",
		PortSpeed:          10000,
		AssociatedVXCs: []*megaport.VXC{
			{
				UID:                "vxc-1",
				Name:               "AWS-SYD",
				ProvisioningStatus: "LIVE",
				RateLimit:          500,
				AEndConfiguration:  megaport.VXCEndConfiguration{UID: "port-aaa"},
				BEndConfiguration:  megaport.VXCEndConfiguration{UID: "cloud-bbb", Name: "AWS Gateway", Location: "ap-southeast-2"},
			},
			{
				UID:                "vxc-2",
				Name:               "Remote-VXC",
				ProvisioningStatus: "LIVE",
				RateLimit:          200,
				AEndConfiguration:  megaport.VXCEndConfiguration{UID: "port-other"},
				BEndConfiguration:  megaport.VXCEndConfiguration{UID: "port-aaa"},
			},
		},
	}

	nodes := buildTopologyNodes([]*megaport.Port{port}, nil, nil, "", false)

	assert.Len(t, nodes, 1)
	assert.Equal(t, "port-aaa", nodes[0].UID)
	assert.Len(t, nodes[0].Connections, 1, "only A-End VXC should be included")
	assert.Equal(t, "vxc-1", nodes[0].Connections[0].UID)
}

func TestBuildTopologyNodes_NoConnections(t *testing.T) {
	mcr := &megaport.MCR{
		UID:                "mcr-111",
		Name:               "Cloud-Router",
		ProvisioningStatus: "LIVE",
		PortSpeed:          5000,
		AssociatedVXCs:     nil,
	}

	nodes := buildTopologyNodes(nil, []*megaport.MCR{mcr}, nil, "", false)

	assert.Len(t, nodes, 1)
	assert.Empty(t, nodes[0].Connections)
}

func TestBuildTopologyNodes_TypeFilter(t *testing.T) {
	port := &megaport.Port{UID: "port-1", Name: "P1", ProvisioningStatus: "LIVE"}
	mcr := &megaport.MCR{UID: "mcr-1", Name: "M1", ProvisioningStatus: "LIVE"}
	mve := &megaport.MVE{UID: "mve-1", Name: "V1", ProvisioningStatus: "LIVE"}

	// filter to mcr only
	nodes := buildTopologyNodes([]*megaport.Port{port}, []*megaport.MCR{mcr}, []*megaport.MVE{mve}, "mcr", false)
	assert.Len(t, nodes, 1)
	assert.Equal(t, "MCR", nodes[0].Type)

	// filter to port only
	nodes = buildTopologyNodes([]*megaport.Port{port}, []*megaport.MCR{mcr}, []*megaport.MVE{mve}, "port", false)
	assert.Len(t, nodes, 1)
	assert.Equal(t, "Port", nodes[0].Type)

	// no filter → all three
	nodes = buildTopologyNodes([]*megaport.Port{port}, []*megaport.MCR{mcr}, []*megaport.MVE{mve}, "", false)
	assert.Len(t, nodes, 3)
}

func TestBuildTopologyNodes_InactiveFiltering(t *testing.T) {
	active := &megaport.Port{UID: "p-live", Name: "Live", ProvisioningStatus: "LIVE"}
	inactive := &megaport.Port{UID: "p-dead", Name: "Dead", ProvisioningStatus: "DECOMMISSIONED"}

	nodes := buildTopologyNodes([]*megaport.Port{active, inactive}, nil, nil, "", false)
	assert.Len(t, nodes, 1)
	assert.Equal(t, "p-live", nodes[0].UID)

	nodes = buildTopologyNodes([]*megaport.Port{active, inactive}, nil, nil, "", true)
	assert.Len(t, nodes, 2)
}

func TestBuildTopologyNodes_MVESpeedIsZero(t *testing.T) {
	mve := &megaport.MVE{UID: "mve-1", Name: "Edge", ProvisioningStatus: "LIVE"}
	nodes := buildTopologyNodes(nil, nil, []*megaport.MVE{mve}, "", false)
	assert.Len(t, nodes, 1)
	assert.Equal(t, 0, nodes[0].SpeedMbps, "MVE has no numeric speed")
}

// ── renderTree ────────────────────────────────────────────────────────────────

func TestRenderTree_Empty(t *testing.T) {
	out := renderTree(nil, true)
	assert.Contains(t, out, "no resources found")
}

func TestRenderTree_NoConnections(t *testing.T) {
	nodes := []TopologyNode{
		{UID: "p-1", Name: "My Port", Type: "Port", Status: "LIVE", SpeedMbps: 10000},
	}
	out := renderTree(nodes, true)
	assert.Contains(t, out, "My Port")
	assert.Contains(t, out, "Port")
	assert.Contains(t, out, "10 Gbps")
	assert.Contains(t, out, "(no connections)")
}

func TestRenderTree_SingleVXC_UsesLastPrefix(t *testing.T) {
	nodes := []TopologyNode{
		{
			UID: "p-1", Name: "My Port", Type: "Port", Status: "LIVE", SpeedMbps: 1000,
			Connections: []TopologyVXC{
				{UID: "v-1", Name: "AWS Link", Status: "LIVE", RateMbps: 500, BEndName: "AWS", BEndLocation: "us-east-1"},
			},
		},
	}
	out := renderTree(nodes, true)
	assert.Contains(t, out, "└── ")
	assert.NotContains(t, out, "├── ")
	assert.Contains(t, out, "AWS Link")
	assert.Contains(t, out, "AWS (us-east-1)")
	assert.Contains(t, out, "500 Mbps")
}

func TestRenderTree_MultipleVXCs_Prefixes(t *testing.T) {
	nodes := []TopologyNode{
		{
			UID: "p-1", Name: "My Port", Type: "Port", Status: "LIVE", SpeedMbps: 10000,
			Connections: []TopologyVXC{
				{UID: "v-1", Name: "First", Status: "LIVE", RateMbps: 500},
				{UID: "v-2", Name: "Second", Status: "LIVE", RateMbps: 200},
			},
		},
	}
	out := renderTree(nodes, true)
	assert.Contains(t, out, "├── ")
	assert.Contains(t, out, "└── ")
	lines := strings.Split(out, "\n")
	firstConnLine := ""
	lastConnLine := ""
	for _, l := range lines {
		if strings.Contains(l, "First") {
			firstConnLine = l
		}
		if strings.Contains(l, "Second") {
			lastConnLine = l
		}
	}
	assert.Contains(t, firstConnLine, "├── ")
	assert.Contains(t, lastConnLine, "└── ")
}

func TestRenderTree_BEndNoLocation(t *testing.T) {
	nodes := []TopologyNode{
		{
			UID: "p-1", Name: "Port", Type: "Port", Status: "LIVE", SpeedMbps: 1000,
			Connections: []TopologyVXC{
				{UID: "v-1", Name: "Link", Status: "LIVE", RateMbps: 100, BEndName: "Remote Port", BEndLocation: ""},
			},
		},
	}
	out := renderTree(nodes, true)
	// Should not have " ()" when location is empty
	assert.Contains(t, out, "Remote Port")
	assert.NotContains(t, out, "Remote Port ()")
}

// ── formatSpeed ──────────────────────────────────────────────────────────────

func TestFormatSpeed(t *testing.T) {
	assert.Equal(t, "-", formatSpeed(0))
	assert.Equal(t, "100 Mbps", formatSpeed(100))
	assert.Equal(t, "500 Mbps", formatSpeed(500))
	assert.Equal(t, "1 Gbps", formatSpeed(1000))
	assert.Equal(t, "10 Gbps", formatSpeed(10000))
	assert.Equal(t, "100 Gbps", formatSpeed(100000))
}

// ── ShowTopology integration ──────────────────────────────────────────────────

func setupTopologyMocks(portSvc *MockPortService, mcrSvc *MockMCRService, mveSvc *MockMVEService) func() {
	original := config.GetLoginFunc()
	config.SetLoginFunc(func(ctx context.Context) (*megaport.Client, error) {
		client := &megaport.Client{}
		client.PortService = portSvc
		client.MCRService = mcrSvc
		client.MVEService = mveSvc
		return client, nil
	})
	return func() { config.SetLoginFunc(original) }
}

func TestShowTopology_TreeOutput(t *testing.T) {
	cleanup := setupTopologyMocks(
		&MockPortService{
			ListPortsResult: []*megaport.Port{
				{
					UID:                "port-aaa",
					Name:               "Sydney-Primary",
					ProvisioningStatus: "LIVE",
					PortSpeed:          10000,
					AssociatedVXCs: []*megaport.VXC{
						{
							UID: "vxc-1", Name: "AWS-SYD",
							ProvisioningStatus: "LIVE", RateLimit: 500,
							AEndConfiguration: megaport.VXCEndConfiguration{UID: "port-aaa"},
							BEndConfiguration: megaport.VXCEndConfiguration{Name: "AWS Gateway", Location: "ap-southeast-2"},
						},
					},
				},
			},
		},
		&MockMCRService{ListMCRsResult: nil},
		&MockMVEService{ListMVEsResult: nil},
	)
	defer cleanup()

	cmd := &cobra.Command{Use: "topology"}
	cmd.Flags().Bool("include-inactive", false, "")
	cmd.Flags().String("type", "", "")

	captured := output.CaptureOutput(func() {
		err := ShowTopology(cmd, nil, true, "table")
		assert.NoError(t, err)
	})

	assert.Contains(t, captured, "Sydney-Primary")
	assert.Contains(t, captured, "AWS-SYD")
	assert.Contains(t, captured, "10 Gbps")
	assert.Contains(t, captured, "└── ")
	assert.Contains(t, captured, "ap-southeast-2")
}

func TestShowTopology_JSONOutput(t *testing.T) {
	cleanup := setupTopologyMocks(
		&MockPortService{
			ListPortsResult: []*megaport.Port{
				{
					UID:                "port-bbb",
					Name:               "Melbourne-DR",
					ProvisioningStatus: "LIVE",
					PortSpeed:          1000,
				},
			},
		},
		&MockMCRService{ListMCRsResult: nil},
		&MockMVEService{ListMVEsResult: nil},
	)
	defer cleanup()

	cmd := &cobra.Command{Use: "topology"}
	cmd.Flags().Bool("include-inactive", false, "")
	cmd.Flags().String("type", "", "")

	captured := output.CaptureOutput(func() {
		err := ShowTopology(cmd, nil, true, "json")
		assert.NoError(t, err)
	})

	var parsed []TopologyNode
	assert.NoError(t, json.Unmarshal([]byte(captured), &parsed))
	assert.Len(t, parsed, 1)
	assert.Equal(t, "Melbourne-DR", parsed[0].Name)
	assert.Equal(t, "Port", parsed[0].Type)
	assert.Empty(t, parsed[0].Connections)
}

func TestShowTopology_UnsupportedFormat(t *testing.T) {
	cleanup := setupTopologyMocks(
		&MockPortService{},
		&MockMCRService{},
		&MockMVEService{},
	)
	defer cleanup()

	cmd := &cobra.Command{Use: "topology"}
	cmd.Flags().Bool("include-inactive", false, "")
	cmd.Flags().String("type", "", "")

	err := ShowTopology(cmd, nil, true, "csv")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not supported")
}

func TestShowTopology_LoginError(t *testing.T) {
	original := config.GetLoginFunc()
	config.SetLoginFunc(func(ctx context.Context) (*megaport.Client, error) {
		return nil, fmt.Errorf("auth failed")
	})
	defer func() { config.SetLoginFunc(original) }()

	cmd := &cobra.Command{Use: "topology"}
	cmd.Flags().Bool("include-inactive", false, "")
	cmd.Flags().String("type", "", "")

	err := ShowTopology(cmd, nil, true, "table")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "auth failed")
}

func TestShowTopology_ListPortsError(t *testing.T) {
	cleanup := setupTopologyMocks(
		&MockPortService{ListPortsErr: fmt.Errorf("port service unavailable")},
		&MockMCRService{},
		&MockMVEService{},
	)
	defer cleanup()

	cmd := &cobra.Command{Use: "topology"}
	cmd.Flags().Bool("include-inactive", false, "")
	cmd.Flags().String("type", "", "")

	err := ShowTopology(cmd, nil, true, "table")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "port service unavailable")
}
