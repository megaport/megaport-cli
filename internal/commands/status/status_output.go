package status

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"

	"github.com/megaport/megaport-cli/internal/base/output"
	megaport "github.com/megaport/megaportgo"
)

// statusPortOutput represents a port in the status dashboard.
type statusPortOutput struct {
	output.Output `json:"-" header:"-"`
	UID           string `json:"uid" header:"UID"`
	Name          string `json:"name" header:"Name"`
	Status        string `json:"status" header:"Status"`
	Speed         int    `json:"speed" header:"Speed"`
	LocationID    int    `json:"location_id" header:"Location ID"`
}

// statusMCROutput represents an MCR in the status dashboard.
type statusMCROutput struct {
	output.Output `json:"-" header:"-"`
	UID           string `json:"uid" header:"UID"`
	Name          string `json:"name" header:"Name"`
	Status        string `json:"status" header:"Status"`
	Speed         int    `json:"speed" header:"Speed"`
	ASN           int    `json:"asn" header:"ASN"`
}

// statusMVEOutput represents an MVE in the status dashboard.
type statusMVEOutput struct {
	output.Output `json:"-" header:"-"`
	UID           string `json:"uid" header:"UID"`
	Name          string `json:"name" header:"Name"`
	Status        string `json:"status" header:"Status"`
	Vendor        string `json:"vendor" header:"Vendor"`
	Size          string `json:"size" header:"Size"`
}

// statusVXCOutput represents a VXC in the status dashboard.
type statusVXCOutput struct {
	output.Output `json:"-" header:"-"`
	UID           string `json:"uid" header:"UID"`
	Name          string `json:"name" header:"Name"`
	Status        string `json:"status" header:"Status"`
	RateLimit     int    `json:"rate_limit" header:"Rate Limit"`
	AEndUID       string `json:"a_end_uid" header:"A-End UID"`
	BEndUID       string `json:"b_end_uid" header:"B-End UID"`
}

// statusIXOutput represents an IX in the status dashboard.
type statusIXOutput struct {
	output.Output `json:"-" header:"-"`
	UID           string `json:"uid" header:"UID"`
	Name          string `json:"name" header:"Name"`
	Status        string `json:"status" header:"Status"`
	ASN           int    `json:"asn" header:"ASN"`
	RateLimit     int    `json:"rate_limit" header:"Rate Limit"`
}

// dashboardSummary holds resource counts.
type dashboardSummary struct {
	Ports int `json:"ports" xml:"ports"`
	MCRs  int `json:"mcrs" xml:"mcrs"`
	MVEs  int `json:"mves" xml:"mves"`
	VXCs  int `json:"vxcs" xml:"vxcs"`
	IXs   int `json:"ixs" xml:"ixs"`
}

// dashboardOutput is the combined output for non-table formats.
type dashboardOutput struct {
	Ports   []statusPortOutput `json:"ports" xml:"ports>port"`
	MCRs    []statusMCROutput  `json:"mcrs" xml:"mcrs>mcr"`
	MVEs    []statusMVEOutput  `json:"mves" xml:"mves>mve"`
	VXCs    []statusVXCOutput  `json:"vxcs" xml:"vxcs>vxc"`
	IXs     []statusIXOutput   `json:"ixs" xml:"ixs>ix"`
	Summary dashboardSummary   `json:"summary" xml:"summary"`
}

// toStatusPortOutput converts a megaport.Port to statusPortOutput.
func toStatusPortOutput(p *megaport.Port) (statusPortOutput, error) {
	if p == nil {
		return statusPortOutput{}, fmt.Errorf("invalid port: nil value")
	}
	return statusPortOutput{
		UID:        p.UID,
		Name:       p.Name,
		Status:     p.ProvisioningStatus,
		Speed:      p.PortSpeed,
		LocationID: p.LocationID,
	}, nil
}

// toStatusMCROutput converts a megaport.MCR to statusMCROutput.
func toStatusMCROutput(m *megaport.MCR) (statusMCROutput, error) {
	if m == nil {
		return statusMCROutput{}, fmt.Errorf("invalid MCR: nil value")
	}
	return statusMCROutput{
		UID:    m.UID,
		Name:   m.Name,
		Status: m.ProvisioningStatus,
		Speed:  m.PortSpeed,
		ASN:    m.Resources.VirtualRouter.ASN,
	}, nil
}

// toStatusMVEOutput converts a megaport.MVE to statusMVEOutput.
func toStatusMVEOutput(m *megaport.MVE) (statusMVEOutput, error) {
	if m == nil {
		return statusMVEOutput{}, fmt.Errorf("invalid MVE: nil value")
	}
	return statusMVEOutput{
		UID:    m.UID,
		Name:   m.Name,
		Status: m.ProvisioningStatus,
		Vendor: m.Vendor,
		Size:   m.Size,
	}, nil
}

// toStatusVXCOutput converts a megaport.VXC to statusVXCOutput.
func toStatusVXCOutput(v *megaport.VXC) (statusVXCOutput, error) {
	if v == nil {
		return statusVXCOutput{}, fmt.Errorf("invalid VXC: nil value")
	}
	return statusVXCOutput{
		UID:       v.UID,
		Name:      v.Name,
		Status:    v.ProvisioningStatus,
		RateLimit: v.RateLimit,
		AEndUID:   v.AEndConfiguration.UID,
		BEndUID:   v.BEndConfiguration.UID,
	}, nil
}

// toStatusIXOutput converts a megaport.IX to statusIXOutput.
func toStatusIXOutput(i *megaport.IX) (statusIXOutput, error) {
	if i == nil {
		return statusIXOutput{}, fmt.Errorf("invalid IX: nil value")
	}
	return statusIXOutput{
		UID:       i.ProductUID,
		Name:      i.ProductName,
		Status:    i.ProvisioningStatus,
		ASN:       i.ASN,
		RateLimit: i.RateLimit,
	}, nil
}

// buildDashboard converts raw resource slices into a dashboardOutput.
func buildDashboard(
	ports []*megaport.Port,
	mcrs []*megaport.MCR,
	mves []*megaport.MVE,
	vxcs []*megaport.VXC,
	ixs []*megaport.IX,
) (dashboardOutput, error) {
	dashboard := dashboardOutput{
		Ports: make([]statusPortOutput, 0, len(ports)),
		MCRs:  make([]statusMCROutput, 0, len(mcrs)),
		MVEs:  make([]statusMVEOutput, 0, len(mves)),
		VXCs:  make([]statusVXCOutput, 0, len(vxcs)),
		IXs:   make([]statusIXOutput, 0, len(ixs)),
	}

	for _, p := range ports {
		o, err := toStatusPortOutput(p)
		if err != nil {
			return dashboardOutput{}, err
		}
		dashboard.Ports = append(dashboard.Ports, o)
	}
	for _, m := range mcrs {
		o, err := toStatusMCROutput(m)
		if err != nil {
			return dashboardOutput{}, err
		}
		dashboard.MCRs = append(dashboard.MCRs, o)
	}
	for _, m := range mves {
		o, err := toStatusMVEOutput(m)
		if err != nil {
			return dashboardOutput{}, err
		}
		dashboard.MVEs = append(dashboard.MVEs, o)
	}
	for _, v := range vxcs {
		o, err := toStatusVXCOutput(v)
		if err != nil {
			return dashboardOutput{}, err
		}
		dashboard.VXCs = append(dashboard.VXCs, o)
	}
	for _, i := range ixs {
		o, err := toStatusIXOutput(i)
		if err != nil {
			return dashboardOutput{}, err
		}
		dashboard.IXs = append(dashboard.IXs, o)
	}

	dashboard.Summary = dashboardSummary{
		Ports: len(dashboard.Ports),
		MCRs:  len(dashboard.MCRs),
		MVEs:  len(dashboard.MVEs),
		VXCs:  len(dashboard.VXCs),
		IXs:   len(dashboard.IXs),
	}

	return dashboard, nil
}

// printDashboard dispatches to the appropriate format printer.
func printDashboard(dashboard dashboardOutput, format string, noColor bool) error {
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

func printDashboardTable(dashboard dashboardOutput, noColor bool) error {
	// PORTS
	output.PrintInfo("\nPORTS (%d)", noColor, len(dashboard.Ports))
	if len(dashboard.Ports) == 0 {
		output.PrintWarning("No ports found.", noColor)
	} else {
		if err := output.PrintOutput(dashboard.Ports, "table", noColor); err != nil {
			return err
		}
	}

	// MCRS
	output.PrintInfo("\nMCRS (%d)", noColor, len(dashboard.MCRs))
	if len(dashboard.MCRs) == 0 {
		output.PrintWarning("No MCRs found.", noColor)
	} else {
		if err := output.PrintOutput(dashboard.MCRs, "table", noColor); err != nil {
			return err
		}
	}

	// MVES
	output.PrintInfo("\nMVES (%d)", noColor, len(dashboard.MVEs))
	if len(dashboard.MVEs) == 0 {
		output.PrintWarning("No MVEs found.", noColor)
	} else {
		if err := output.PrintOutput(dashboard.MVEs, "table", noColor); err != nil {
			return err
		}
	}

	// VXCS
	output.PrintInfo("\nVXCS (%d)", noColor, len(dashboard.VXCs))
	if len(dashboard.VXCs) == 0 {
		output.PrintWarning("No VXCs found.", noColor)
	} else {
		if err := output.PrintOutput(dashboard.VXCs, "table", noColor); err != nil {
			return err
		}
	}

	// IXS
	output.PrintInfo("\nIXS (%d)", noColor, len(dashboard.IXs))
	if len(dashboard.IXs) == 0 {
		output.PrintWarning("No IXs found.", noColor)
	} else {
		if err := output.PrintOutput(dashboard.IXs, "table", noColor); err != nil {
			return err
		}
	}

	s := dashboard.Summary
	output.PrintInfo("\nTotal: %d port(s), %d MCR(s), %d MVE(s), %d VXC(s), %d IX(s)", noColor,
		s.Ports, s.MCRs, s.MVEs, s.VXCs, s.IXs)

	return nil
}

func printDashboardJSON(dashboard dashboardOutput) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(dashboard)
}

func printDashboardXML(dashboard dashboardOutput) error {
	type xmlDashboard struct {
		XMLName xml.Name           `xml:"dashboard"`
		Ports   []statusPortOutput `xml:"ports>port"`
		MCRs    []statusMCROutput  `xml:"mcrs>mcr"`
		MVEs    []statusMVEOutput  `xml:"mves>mve"`
		VXCs    []statusVXCOutput  `xml:"vxcs>vxc"`
		IXs     []statusIXOutput   `xml:"ixs>ix"`
		Summary dashboardSummary   `xml:"summary"`
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

func printDashboardCSV(dashboard dashboardOutput, noColor bool) error {
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
