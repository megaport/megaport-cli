package topology

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/commands/config"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

// TopologyNode represents a parent resource (Port, MCR, or MVE) with its VXC connections.
type TopologyNode struct {
	UID         string        `json:"uid"`
	Name        string        `json:"name"`
	Type        string        `json:"type"`
	Status      string        `json:"status"`
	SpeedMbps   int           `json:"speedMbps"`
	Location    string        `json:"location"`
	Connections []TopologyVXC `json:"connections"`
}

// TopologyVXC represents a VXC connection hanging off a parent node.
type TopologyVXC struct {
	UID          string `json:"uid"`
	Name         string `json:"name"`
	Status       string `json:"status"`
	RateMbps     int    `json:"rateMbps"`
	BEndUID      string `json:"bEndUid"`
	BEndName     string `json:"bEndName"`
	BEndLocation string `json:"bEndLocation"`
}

// ShowTopology is the cobra run function for the topology command.
func ShowTopology(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	output.SetOutputFormat(outputFormat)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	client, err := config.Login(ctx)
	if err != nil {
		output.PrintError("Failed to log in: %v", noColor, err)
		return fmt.Errorf("error logging in: %v", err)
	}

	includeInactive, _ := cmd.Flags().GetBool("include-inactive")
	typeFilter, _ := cmd.Flags().GetString("type")
	typeFilter = strings.ToLower(strings.TrimSpace(typeFilter))
	switch typeFilter {
	case "", "port", "mcr", "mve":
	default:
		return fmt.Errorf("invalid value for --type: %q (must be one of: port, mcr, mve)", typeFilter)
	}

	// Fetch ports, MCRs, and MVEs in parallel.
	var (
		ports    []*megaport.Port
		mcrs     []*megaport.MCR
		mves     []*megaport.MVE
		portsErr error
		mcrsErr  error
		mvesErr  error
		wg       sync.WaitGroup
	)

	wg.Add(3)
	go func() {
		defer wg.Done()
		ports, portsErr = client.PortService.ListPorts(ctx)
	}()
	go func() {
		defer wg.Done()
		mcrs, mcrsErr = client.MCRService.ListMCRs(ctx, &megaport.ListMCRsRequest{IncludeInactive: includeInactive})
	}()
	go func() {
		defer wg.Done()
		mves, mvesErr = client.MVEService.ListMVEs(ctx, &megaport.ListMVEsRequest{IncludeInactive: includeInactive})
	}()
	wg.Wait()

	if portsErr != nil {
		output.PrintError("Failed to list ports: %v", noColor, portsErr)
		return fmt.Errorf("error listing ports: %v", portsErr)
	}
	if mcrsErr != nil {
		output.PrintError("Failed to list MCRs: %v", noColor, mcrsErr)
		return fmt.Errorf("error listing MCRs: %v", mcrsErr)
	}
	if mvesErr != nil {
		output.PrintError("Failed to list MVEs: %v", noColor, mvesErr)
		return fmt.Errorf("error listing MVEs: %v", mvesErr)
	}

	nodes := buildTopologyNodes(ports, mcrs, mves, typeFilter, includeInactive)

	switch outputFormat {
	case "json":
		jsonBytes, err := json.MarshalIndent(nodes, "", "  ")
		if err != nil {
			return fmt.Errorf("error marshaling topology: %v", err)
		}
		fmt.Println(string(jsonBytes))
	case "csv", "xml":
		return fmt.Errorf("output format %q is not supported for topology — use table (default) or json", outputFormat)
	default:
		fmt.Print(renderTree(nodes, noColor))
	}

	return nil
}

// buildTopologyNodes constructs the topology from the fetched resources.
// Only VXCs where AEndConfiguration.UID == node.UID are included (avoids duplicates).
func buildTopologyNodes(
	ports []*megaport.Port,
	mcrs []*megaport.MCR,
	mves []*megaport.MVE,
	typeFilter string,
	includeInactive bool,
) []TopologyNode {
	var nodes []TopologyNode

	tf := strings.ToLower(typeFilter)

	if tf == "" || tf == "port" {
		for _, p := range ports {
			if p == nil {
				continue
			}
			if !includeInactive && isInactive(p.ProvisioningStatus) {
				continue
			}
			node := TopologyNode{
				UID:       p.UID,
				Name:      p.Name,
				Type:      "Port",
				Status:    p.ProvisioningStatus,
				SpeedMbps: p.PortSpeed,
				Location:  locationName(p.LocationDetails),
			}
			for _, vxc := range p.AssociatedVXCs {
				if vxc == nil || vxc.AEndConfiguration.UID != p.UID {
					continue
				}
				node.Connections = append(node.Connections, vxcToTopology(vxc))
			}
			nodes = append(nodes, node)
		}
	}

	if tf == "" || tf == "mcr" {
		for _, m := range mcrs {
			if m == nil {
				continue
			}
			if !includeInactive && isInactive(m.ProvisioningStatus) {
				continue
			}
			node := TopologyNode{
				UID:       m.UID,
				Name:      m.Name,
				Type:      "MCR",
				Status:    m.ProvisioningStatus,
				SpeedMbps: m.PortSpeed,
				Location:  locationName(m.LocationDetails),
			}
			for _, vxc := range m.AssociatedVXCs {
				if vxc == nil || vxc.AEndConfiguration.UID != m.UID {
					continue
				}
				node.Connections = append(node.Connections, vxcToTopology(vxc))
			}
			nodes = append(nodes, node)
		}
	}

	if tf == "" || tf == "mve" {
		for _, mv := range mves {
			if mv == nil {
				continue
			}
			if !includeInactive && isInactive(mv.ProvisioningStatus) {
				continue
			}
			node := TopologyNode{
				UID:      mv.UID,
				Name:     mv.Name,
				Type:     "MVE",
				Status:   mv.ProvisioningStatus,
				Location: locationName(mv.LocationDetails),
			}
			for _, vxc := range mv.AssociatedVXCs {
				if vxc == nil || vxc.AEndConfiguration.UID != mv.UID {
					continue
				}
				node.Connections = append(node.Connections, vxcToTopology(vxc))
			}
			nodes = append(nodes, node)
		}
	}

	return nodes
}

// locationName safely dereferences a *ProductLocationDetails Name.
func locationName(d *megaport.ProductLocationDetails) string {
	if d == nil {
		return ""
	}
	return d.Name
}

func vxcToTopology(vxc *megaport.VXC) TopologyVXC {
	return TopologyVXC{
		UID:          vxc.UID,
		Name:         vxc.Name,
		Status:       vxc.ProvisioningStatus,
		RateMbps:     vxc.RateLimit,
		BEndUID:      vxc.BEndConfiguration.UID,
		BEndName:     vxc.BEndConfiguration.Name,
		BEndLocation: vxc.BEndConfiguration.Location,
	}
}

// isInactive returns true for statuses that represent decommissioned or deleted resources.
func isInactive(status string) bool {
	s := strings.ToUpper(status)
	return s == "DECOMMISSIONED" || s == "CANCELLED" || s == "DELETED"
}

// formatSpeed converts Mbps to a human-readable string.
func formatSpeed(mbps int) string {
	if mbps == 0 {
		return "-"
	}
	if mbps%1000 == 0 {
		return fmt.Sprintf("%d Gbps", mbps/1000)
	}
	return fmt.Sprintf("%d Mbps", mbps)
}

// statusBadge returns a colour-coded status string for tree output.
func statusBadge(status string, noColor bool) string {
	if noColor {
		return status
	}
	s := strings.ToUpper(status)
	switch {
	case strings.Contains(s, "LIVE") || strings.Contains(s, "ACTIVE"):
		return color.New(color.FgHiWhite, color.BgGreen, color.Bold).Sprintf(" %s ", s)
	case strings.Contains(s, "CONFIGURED") || strings.Contains(s, "PENDING") ||
		strings.Contains(s, "PROVISIONING"):
		return color.New(color.FgBlack, color.BgYellow, color.Bold).Sprintf(" %s ", s)
	case strings.Contains(s, "ERROR") || strings.Contains(s, "FAILED"):
		return color.New(color.FgHiWhite, color.BgRed, color.Bold).Sprintf(" %s ", s)
	default:
		return s
	}
}

// boldText returns a bold-formatted string, or plain if noColor.
func boldText(s string, noColor bool) string {
	if noColor {
		return s
	}
	return color.New(color.Bold).Sprint(s)
}

// renderTree produces an ASCII tree representation of the topology.
func renderTree(nodes []TopologyNode, noColor bool) string {
	if len(nodes) == 0 {
		return "(no resources found)\n"
	}

	var sb strings.Builder

	for i, node := range nodes {
		speed := formatSpeed(node.SpeedMbps)
		fmt.Fprintf(&sb, "%s (%s, %s, %s)\n",
			boldText(node.Name, noColor),
			node.Type,
			statusBadge(node.Status, noColor),
			speed,
		)

		if len(node.Connections) == 0 {
			fmt.Fprintf(&sb, "  (no connections)\n")
		} else {
			for j, conn := range node.Connections {
				isLast := j == len(node.Connections)-1
				prefix := "├── "
				if isLast {
					prefix = "└── "
				}
				bEnd := conn.BEndName
				if conn.BEndLocation != "" {
					bEnd = fmt.Sprintf("%s (%s)", conn.BEndName, conn.BEndLocation)
				}
				fmt.Fprintf(&sb, "%s%s (%s, %s) → %s\n",
					prefix,
					conn.Name,
					statusBadge(conn.Status, noColor),
					formatSpeed(conn.RateMbps),
					bEnd,
				)
			}
		}

		if i < len(nodes)-1 {
			fmt.Fprintln(&sb)
		}
	}

	return sb.String()
}
