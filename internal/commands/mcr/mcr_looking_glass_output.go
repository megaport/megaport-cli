package mcr

import (
	"fmt"
	"strings"

	"github.com/megaport/megaport-cli/internal/base/output"
	megaport "github.com/megaport/megaportgo"
)

// IPRouteOutput represents the output format for IP routes
type IPRouteOutput struct {
	output.Output `json:"-" header:"-"`
	Prefix        string `json:"prefix" header:"Prefix"`
	NextHop       string `json:"next_hop" header:"Next Hop"`
	Protocol      string `json:"protocol" header:"Protocol"`
	Metric        string `json:"metric,omitempty" header:"Metric"`
	LocalPref     string `json:"local_pref,omitempty" header:"Local Pref"`
	ASPath        string `json:"as_path,omitempty" header:"AS Path"`
	Age           string `json:"age,omitempty" header:"Age"`
	Interface     string `json:"interface,omitempty" header:"Interface"`
	VXCName       string `json:"vxc_name,omitempty" header:"VXC Name"`
	Best          string `json:"best,omitempty" header:"Best"`
}

// ToIPRouteOutput converts a megaport.LookingGlassIPRoute to IPRouteOutput
func ToIPRouteOutput(route *megaport.LookingGlassIPRoute) (IPRouteOutput, error) {
	if route == nil {
		return IPRouteOutput{}, fmt.Errorf("invalid route: nil value")
	}

	out := IPRouteOutput{
		Prefix:    route.Prefix,
		NextHop:   route.NextHop,
		Protocol:  string(route.Protocol),
		Interface: route.Interface,
		VXCName:   route.VXCName,
	}

	if route.Metric != nil {
		out.Metric = fmt.Sprintf("%d", *route.Metric)
	}
	if route.LocalPref != nil {
		out.LocalPref = fmt.Sprintf("%d", *route.LocalPref)
	}
	if len(route.ASPath) > 0 {
		asPathStrs := make([]string, len(route.ASPath))
		for i, as := range route.ASPath {
			asPathStrs[i] = fmt.Sprintf("%d", as)
		}
		out.ASPath = strings.Join(asPathStrs, " ")
	}
	if route.Age != nil {
		out.Age = formatDuration(*route.Age)
	}
	if route.Best != nil {
		if *route.Best {
			out.Best = "Yes"
		} else {
			out.Best = "No"
		}
	}

	return out, nil
}

// BGPRouteOutput represents the output format for BGP routes
type BGPRouteOutput struct {
	output.Output `json:"-" header:"-"`
	Prefix        string `json:"prefix" header:"Prefix"`
	NextHop       string `json:"next_hop" header:"Next Hop"`
	ASPath        string `json:"as_path" header:"AS Path"`
	LocalPref     string `json:"local_pref,omitempty" header:"Local Pref"`
	MED           string `json:"med,omitempty" header:"MED"`
	Origin        string `json:"origin,omitempty" header:"Origin"`
	Communities   string `json:"communities,omitempty" header:"Communities"`
	NeighborIP    string `json:"neighbor_ip,omitempty" header:"Neighbor IP"`
	NeighborASN   string `json:"neighbor_asn,omitempty" header:"Neighbor ASN"`
	Valid         string `json:"valid" header:"Valid"`
	Best          string `json:"best" header:"Best"`
	VXCName       string `json:"vxc_name,omitempty" header:"VXC Name"`
	Age           string `json:"age,omitempty" header:"Age"`
}

// ToBGPRouteOutput converts a megaport.LookingGlassBGPRoute to BGPRouteOutput
func ToBGPRouteOutput(route *megaport.LookingGlassBGPRoute) (BGPRouteOutput, error) {
	if route == nil {
		return BGPRouteOutput{}, fmt.Errorf("invalid BGP route: nil value")
	}

	out := BGPRouteOutput{
		Prefix:     route.Prefix,
		NextHop:    route.NextHop,
		Origin:     route.Origin,
		NeighborIP: route.NeighborIP,
		VXCName:    route.VXCName,
		Valid:      boolToYesNo(route.Valid),
		Best:       boolToYesNo(route.Best),
	}

	if len(route.ASPath) > 0 {
		asPathStrs := make([]string, len(route.ASPath))
		for i, as := range route.ASPath {
			asPathStrs[i] = fmt.Sprintf("%d", as)
		}
		out.ASPath = strings.Join(asPathStrs, " ")
	}
	if route.LocalPref != nil {
		out.LocalPref = fmt.Sprintf("%d", *route.LocalPref)
	}
	if route.MED != nil {
		out.MED = fmt.Sprintf("%d", *route.MED)
	}
	if len(route.Communities) > 0 {
		out.Communities = strings.Join(route.Communities, ", ")
	}
	if route.NeighborASN != nil {
		out.NeighborASN = fmt.Sprintf("%d", *route.NeighborASN)
	}
	if route.Age != nil {
		out.Age = formatDuration(*route.Age)
	}

	return out, nil
}

// BGPSessionOutput represents the output format for BGP sessions
type BGPSessionOutput struct {
	output.Output   `json:"-" header:"-"`
	SessionID       string `json:"session_id" header:"Session ID"`
	NeighborAddress string `json:"neighbor_address" header:"Neighbor Address"`
	NeighborASN     int    `json:"neighbor_asn" header:"Neighbor ASN"`
	LocalASN        int    `json:"local_asn" header:"Local ASN"`
	Status          string `json:"status" header:"Status"`
	Uptime          string `json:"uptime,omitempty" header:"Uptime"`
	PrefixesIn      string `json:"prefixes_in,omitempty" header:"Prefixes In"`
	PrefixesOut     string `json:"prefixes_out,omitempty" header:"Prefixes Out"`
	VXCName         string `json:"vxc_name,omitempty" header:"VXC Name"`
	Description     string `json:"description,omitempty" header:"Description"`
}

// ToBGPSessionOutput converts a megaport.LookingGlassBGPSession to BGPSessionOutput
func ToBGPSessionOutput(session *megaport.LookingGlassBGPSession) (BGPSessionOutput, error) {
	if session == nil {
		return BGPSessionOutput{}, fmt.Errorf("invalid BGP session: nil value")
	}

	out := BGPSessionOutput{
		SessionID:       session.SessionID,
		NeighborAddress: session.NeighborAddress,
		NeighborASN:     session.NeighborASN,
		LocalASN:        session.LocalASN,
		Status:          string(session.Status),
		VXCName:         session.VXCName,
		Description:     session.Description,
	}

	if session.Uptime != nil {
		out.Uptime = formatDuration(*session.Uptime)
	}
	if session.PrefixesIn != nil {
		out.PrefixesIn = fmt.Sprintf("%d", *session.PrefixesIn)
	}
	if session.PrefixesOut != nil {
		out.PrefixesOut = fmt.Sprintf("%d", *session.PrefixesOut)
	}

	return out, nil
}

// BGPNeighborRouteOutput represents the output format for BGP neighbor routes
type BGPNeighborRouteOutput struct {
	output.Output `json:"-" header:"-"`
	Prefix        string `json:"prefix" header:"Prefix"`
	NextHop       string `json:"next_hop" header:"Next Hop"`
	ASPath        string `json:"as_path" header:"AS Path"`
	LocalPref     string `json:"local_pref,omitempty" header:"Local Pref"`
	MED           string `json:"med,omitempty" header:"MED"`
	Origin        string `json:"origin,omitempty" header:"Origin"`
	Communities   string `json:"communities,omitempty" header:"Communities"`
	Valid         string `json:"valid" header:"Valid"`
	Best          string `json:"best" header:"Best"`
}

// ToBGPNeighborRouteOutput converts a megaport.LookingGlassBGPNeighborRoute to BGPNeighborRouteOutput
func ToBGPNeighborRouteOutput(route *megaport.LookingGlassBGPNeighborRoute) (BGPNeighborRouteOutput, error) {
	if route == nil {
		return BGPNeighborRouteOutput{}, fmt.Errorf("invalid BGP neighbor route: nil value")
	}

	out := BGPNeighborRouteOutput{
		Prefix:  route.Prefix,
		NextHop: route.NextHop,
		Origin:  route.Origin,
		Valid:   boolToYesNo(route.Valid),
		Best:    boolToYesNo(route.Best),
	}

	if len(route.ASPath) > 0 {
		asPathStrs := make([]string, len(route.ASPath))
		for i, as := range route.ASPath {
			asPathStrs[i] = fmt.Sprintf("%d", as)
		}
		out.ASPath = strings.Join(asPathStrs, " ")
	}
	if route.LocalPref != nil {
		out.LocalPref = fmt.Sprintf("%d", *route.LocalPref)
	}
	if route.MED != nil {
		out.MED = fmt.Sprintf("%d", *route.MED)
	}
	if len(route.Communities) > 0 {
		out.Communities = strings.Join(route.Communities, ", ")
	}

	return out, nil
}

// Helper functions

func boolToYesNo(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

func formatDuration(seconds int) string {
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	if seconds < 3600 {
		return fmt.Sprintf("%dm%ds", seconds/60, seconds%60)
	}
	if seconds < 86400 {
		hours := seconds / 3600
		minutes := (seconds % 3600) / 60
		return fmt.Sprintf("%dh%dm", hours, minutes)
	}
	days := seconds / 86400
	hours := (seconds % 86400) / 3600
	return fmt.Sprintf("%dd%dh", days, hours)
}

// Print functions for each output type

func printIPRoutes(routes []*megaport.LookingGlassIPRoute, format string, noColor bool) error {
	outputs := make([]IPRouteOutput, 0, len(routes))
	for _, route := range routes {
		out, err := ToIPRouteOutput(route)
		if err != nil {
			return err
		}
		outputs = append(outputs, out)
	}
	return output.PrintOutput(outputs, format, noColor)
}

func printBGPRoutes(routes []*megaport.LookingGlassBGPRoute, format string, noColor bool) error {
	outputs := make([]BGPRouteOutput, 0, len(routes))
	for _, route := range routes {
		out, err := ToBGPRouteOutput(route)
		if err != nil {
			return err
		}
		outputs = append(outputs, out)
	}
	return output.PrintOutput(outputs, format, noColor)
}

func printBGPSessions(sessions []*megaport.LookingGlassBGPSession, format string, noColor bool) error {
	outputs := make([]BGPSessionOutput, 0, len(sessions))
	for _, session := range sessions {
		out, err := ToBGPSessionOutput(session)
		if err != nil {
			return err
		}
		outputs = append(outputs, out)
	}
	return output.PrintOutput(outputs, format, noColor)
}

func printBGPNeighborRoutes(routes []*megaport.LookingGlassBGPNeighborRoute, format string, noColor bool) error {
	outputs := make([]BGPNeighborRouteOutput, 0, len(routes))
	for _, route := range routes {
		out, err := ToBGPNeighborRouteOutput(route)
		if err != nil {
			return err
		}
		outputs = append(outputs, out)
	}
	return output.PrintOutput(outputs, format, noColor)
}
