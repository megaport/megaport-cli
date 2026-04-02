package status

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"

	"github.com/megaport/megaport-cli/internal/base/output"
	megaport "github.com/megaport/megaportgo"
)

// StatusPortOutput represents a port in the status dashboard.
type StatusPortOutput struct {
	output.Output `json:"-" header:"-"`
	UID           string `json:"uid" header:"UID"`
	Name          string `json:"name" header:"Name"`
	Status        string `json:"status" header:"Status"`
	Speed         int    `json:"speed" header:"Speed"`
	LocationID    int    `json:"location_id" header:"Location ID"`
}

// StatusMCROutput represents an MCR in the status dashboard.
type StatusMCROutput struct {
	output.Output `json:"-" header:"-"`
	UID           string `json:"uid" header:"UID"`
	Name          string `json:"name" header:"Name"`
	Status        string `json:"status" header:"Status"`
	Speed         int    `json:"speed" header:"Speed"`
	ASN           int    `json:"asn" header:"ASN"`
}

// StatusMVEOutput represents an MVE in the status dashboard.
type StatusMVEOutput struct {
	output.Output `json:"-" header:"-"`
	UID           string `json:"uid" header:"UID"`
	Name          string `json:"name" header:"Name"`
	Status        string `json:"status" header:"Status"`
	Vendor        string `json:"vendor" header:"Vendor"`
	Size          string `json:"size" header:"Size"`
}

// StatusVXCOutput represents a VXC in the status dashboard.
type StatusVXCOutput struct {
	output.Output `json:"-" header:"-"`
	UID           string `json:"uid" header:"UID"`
	Name          string `json:"name" header:"Name"`
	Status        string `json:"status" header:"Status"`
	RateLimit     int    `json:"rate_limit" header:"Rate Limit"`
	AEndUID       string `json:"a_end_uid" header:"A-End UID"`
	BEndUID       string `json:"b_end_uid" header:"B-End UID"`
}

// StatusIXOutput represents an IX in the status dashboard.
type StatusIXOutput struct {
	output.Output `json:"-" header:"-"`
	UID           string `json:"uid" header:"UID"`
	Name          string `json:"name" header:"Name"`
	Status        string `json:"status" header:"Status"`
	ASN           int    `json:"asn" header:"ASN"`
	RateLimit     int    `json:"rate_limit" header:"Rate Limit"`
}

// DashboardSummary holds resource counts.
type DashboardSummary struct {
	Ports int `json:"ports" xml:"ports"`
	MCRs  int `json:"mcrs" xml:"mcrs"`
	MVEs  int `json:"mves" xml:"mves"`
	VXCs  int `json:"vxcs" xml:"vxcs"`
	IXs   int `json:"ixs" xml:"ixs"`
}

// DashboardOutput is the combined output for non-table formats.
type DashboardOutput struct {
	Ports   []StatusPortOutput `json:"ports" xml:"ports>port"`
	MCRs    []StatusMCROutput  `json:"mcrs" xml:"mcrs>mcr"`
	MVEs    []StatusMVEOutput  `json:"mves" xml:"mves>mve"`
	VXCs    []StatusVXCOutput  `json:"vxcs" xml:"vxcs>vxc"`
	IXs     []StatusIXOutput   `json:"ixs" xml:"ixs>ix"`
	Summary DashboardSummary   `json:"summary" xml:"summary"`
}

// toStatusPortOutput converts a megaport.Port to StatusPortOutput.
func toStatusPortOutput(p *megaport.Port) (StatusPortOutput, error) {
	if p == nil {
		return StatusPortOutput{}, fmt.Errorf("invalid port: nil value")
	}
	return StatusPortOutput{
		UID:        p.UID,
		Name:       p.Name,
		Status:     p.ProvisioningStatus,
		Speed:      p.PortSpeed,
		LocationID: p.LocationID,
	}, nil
}

// toStatusMCROutput converts a megaport.MCR to StatusMCROutput.
func toStatusMCROutput(m *megaport.MCR) (StatusMCROutput, error) {
	if m == nil {
		return StatusMCROutput{}, fmt.Errorf("invalid MCR: nil value")
	}
	return StatusMCROutput{
		UID:    m.UID,
		Name:   m.Name,
		Status: m.ProvisioningStatus,
		Speed:  m.PortSpeed,
		ASN:    m.Resources.VirtualRouter.ASN,
	}, nil
}

// toStatusMVEOutput converts a megaport.MVE to StatusMVEOutput.
func toStatusMVEOutput(m *megaport.MVE) (StatusMVEOutput, error) {
	if m == nil {
		return StatusMVEOutput{}, fmt.Errorf("invalid MVE: nil value")
	}
	return StatusMVEOutput{
		UID:    m.UID,
		Name:   m.Name,
		Status: m.ProvisioningStatus,
		Vendor: m.Vendor,
		Size:   m.Size,
	}, nil
}

// toStatusVXCOutput converts a megaport.VXC to StatusVXCOutput.
func toStatusVXCOutput(v *megaport.VXC) (StatusVXCOutput, error) {
	if v == nil {
		return StatusVXCOutput{}, fmt.Errorf("invalid VXC: nil value")
	}
	return StatusVXCOutput{
		UID:       v.UID,
		Name:      v.Name,
		Status:    v.ProvisioningStatus,
		RateLimit: v.RateLimit,
		AEndUID:   v.AEndConfiguration.UID,
		BEndUID:   v.BEndConfiguration.UID,
	}, nil
}

// toStatusIXOutput converts a megaport.IX to StatusIXOutput.
func toStatusIXOutput(i *megaport.IX) (StatusIXOutput, error) {
	if i == nil {
		return StatusIXOutput{}, fmt.Errorf("invalid IX: nil value")
	}
	return StatusIXOutput{
		UID:       i.ProductUID,
		Name:      i.ProductName,
		Status:    i.ProvisioningStatus,
		ASN:       i.ASN,
		RateLimit: i.RateLimit,
	}, nil
}

// buildDashboard converts raw resource slices into a DashboardOutput.
func buildDashboard(
	ports []*megaport.Port,
	mcrs []*megaport.MCR,
	mves []*megaport.MVE,
	vxcs []*megaport.VXC,
	ixs []*megaport.IX,
) (DashboardOutput, error) {
	dashboard := DashboardOutput{
		Ports: make([]StatusPortOutput, 0, len(ports)),
		MCRs:  make([]StatusMCROutput, 0, len(mcrs)),
		MVEs:  make([]StatusMVEOutput, 0, len(mves)),
		VXCs:  make([]StatusVXCOutput, 0, len(vxcs)),
		IXs:   make([]StatusIXOutput, 0, len(ixs)),
	}

	for _, p := range ports {
		o, err := toStatusPortOutput(p)
		if err != nil {
			return DashboardOutput{}, err
		}
		dashboard.Ports = append(dashboard.Ports, o)
	}
	for _, m := range mcrs {
		o, err := toStatusMCROutput(m)
		if err != nil {
			return DashboardOutput{}, err
		}
		dashboard.MCRs = append(dashboard.MCRs, o)
	}
	for _, m := range mves {
		o, err := toStatusMVEOutput(m)
		if err != nil {
			return DashboardOutput{}, err
		}
		dashboard.MVEs = append(dashboard.MVEs, o)
	}
	for _, v := range vxcs {
		o, err := toStatusVXCOutput(v)
		if err != nil {
			return DashboardOutput{}, err
		}
		dashboard.VXCs = append(dashboard.VXCs, o)
	}
	for _, i := range ixs {
		o, err := toStatusIXOutput(i)
		if err != nil {
			return DashboardOutput{}, err
		}
		dashboard.IXs = append(dashboard.IXs, o)
	}

	dashboard.Summary = DashboardSummary{
		Ports: len(dashboard.Ports),
		MCRs:  len(dashboard.MCRs),
		MVEs:  len(dashboard.MVEs),
		VXCs:  len(dashboard.VXCs),
		IXs:   len(dashboard.IXs),
	}

	return dashboard, nil
}

// printDashboard dispatches to the appropriate format printer.
func printDashboard(dashboard DashboardOutput, format string, noColor bool) error {
	switch format {
	case "json":
		return printDashboardJSON(dashboard)
	case "xml":
		return printDashboardXML(dashboard)
	case "csv":
		return printDashboardCSV(dashboard, noColor)
	default:
		return printDashboardTable(dashboard, noColor)
	}
}

func printDashboardTable(dashboard DashboardOutput, noColor bool) error {
	// PORTS
	fmt.Printf("\nPORTS (%d)\n", len(dashboard.Ports))
	if len(dashboard.Ports) == 0 {
		output.PrintWarning("No ports found.", noColor)
	} else {
		if err := output.PrintOutput(dashboard.Ports, "table", noColor); err != nil {
			return err
		}
	}

	// MCRS
	fmt.Printf("\nMCRS (%d)\n", len(dashboard.MCRs))
	if len(dashboard.MCRs) == 0 {
		output.PrintWarning("No MCRs found.", noColor)
	} else {
		if err := output.PrintOutput(dashboard.MCRs, "table", noColor); err != nil {
			return err
		}
	}

	// MVES
	fmt.Printf("\nMVES (%d)\n", len(dashboard.MVEs))
	if len(dashboard.MVEs) == 0 {
		output.PrintWarning("No MVEs found.", noColor)
	} else {
		if err := output.PrintOutput(dashboard.MVEs, "table", noColor); err != nil {
			return err
		}
	}

	// VXCS
	fmt.Printf("\nVXCS (%d)\n", len(dashboard.VXCs))
	if len(dashboard.VXCs) == 0 {
		output.PrintWarning("No VXCs found.", noColor)
	} else {
		if err := output.PrintOutput(dashboard.VXCs, "table", noColor); err != nil {
			return err
		}
	}

	// IXS
	fmt.Printf("\nIXS (%d)\n", len(dashboard.IXs))
	if len(dashboard.IXs) == 0 {
		output.PrintWarning("No IXs found.", noColor)
	} else {
		if err := output.PrintOutput(dashboard.IXs, "table", noColor); err != nil {
			return err
		}
	}

	s := dashboard.Summary
	fmt.Printf("\nTotal: %d port(s), %d MCR(s), %d MVE(s), %d VXC(s), %d IX(s)\n",
		s.Ports, s.MCRs, s.MVEs, s.VXCs, s.IXs)

	return nil
}

func printDashboardJSON(dashboard DashboardOutput) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(dashboard)
}

func printDashboardXML(dashboard DashboardOutput) error {
	type xmlDashboard struct {
		XMLName xml.Name           `xml:"dashboard"`
		Ports   []StatusPortOutput `xml:"ports>port"`
		MCRs    []StatusMCROutput  `xml:"mcrs>mcr"`
		MVEs    []StatusMVEOutput  `xml:"mves>mve"`
		VXCs    []StatusVXCOutput  `xml:"vxcs>vxc"`
		IXs     []StatusIXOutput   `xml:"ixs>ix"`
		Summary DashboardSummary   `xml:"summary"`
	}
	out := xmlDashboard{
		Ports:   dashboard.Ports,
		MCRs:    dashboard.MCRs,
		MVEs:    dashboard.MVEs,
		VXCs:    dashboard.VXCs,
		IXs:     dashboard.IXs,
		Summary: dashboard.Summary,
	}
	encoder := xml.NewEncoder(os.Stdout)
	encoder.Indent("", "  ")
	if err := encoder.Encode(out); err != nil {
		return err
	}
	fmt.Fprintln(os.Stdout)
	return nil
}

func printDashboardCSV(dashboard DashboardOutput, noColor bool) error {
	sections := []struct {
		name string
		fn   func() error
	}{
		{"PORTS", func() error { return output.PrintOutput(dashboard.Ports, "csv", noColor) }},
		{"MCRS", func() error { return output.PrintOutput(dashboard.MCRs, "csv", noColor) }},
		{"MVES", func() error { return output.PrintOutput(dashboard.MVEs, "csv", noColor) }},
		{"VXCS", func() error { return output.PrintOutput(dashboard.VXCs, "csv", noColor) }},
		{"IXS", func() error { return output.PrintOutput(dashboard.IXs, "csv", noColor) }},
	}
	for _, s := range sections {
		fmt.Printf("# %s\n", s.name)
		if err := s.fn(); err != nil {
			return err
		}
		fmt.Println()
	}
	return nil
}
