package nat_gateway

import (
	"fmt"
	"strings"
	"time"

	"github.com/megaport/megaport-cli/internal/base/output"
	megaport "github.com/megaport/megaportgo"
)

// natGatewayOutput is the display/serialisation struct for a NAT Gateway.
type natGatewayOutput struct {
	output.Output      `json:"-" header:"-"`
	UID                string `json:"uid"                header:"UID"`
	Name               string `json:"name"               header:"Name"`
	LocationID         int    `json:"location_id"        header:"Location ID"`
	Speed              int    `json:"speed"              header:"Speed (Mbps)"`
	Term               int    `json:"term"               header:"Term"`
	ProvisioningStatus string `json:"provisioning_status" header:"Status"`
	SessionCount       int    `json:"session_count"      header:"Sessions"`
	DiversityZone      string `json:"diversity_zone,omitempty" header:"Diversity Zone"`
	ASN                int    `json:"asn,omitempty"      header:"ASN"`
}

// natGatewaySessionOutput is the display struct for a NAT Gateway session entry.
type natGatewaySessionOutput struct {
	output.Output `json:"-" header:"-"`
	SpeedMbps     int    `json:"speed_mbps"     header:"Speed (Mbps)"`
	SessionCounts string `json:"session_counts" header:"Session Counts"`
}

// natGatewayValidateOutput is the display struct for a NAT Gateway validate result.
type natGatewayValidateOutput struct {
	output.Output `json:"-" header:"-"`
	UID           string  `json:"uid"            header:"UID"`
	ProductType   string  `json:"product_type"   header:"Product Type"`
	Metro         string  `json:"metro"          header:"Metro"`
	Currency      string  `json:"currency"       header:"Currency"`
	MonthlyRate   float64 `json:"monthly_rate"   header:"Monthly Rate"`
	HourlyRate    float64 `json:"hourly_rate"    header:"Hourly Rate"`
	MonthlySetup  float64 `json:"monthly_setup"  header:"Monthly Setup"`
}

// natGatewayBuyOutput is the display struct for a NAT Gateway buy result.
type natGatewayBuyOutput struct {
	output.Output      `json:"-" header:"-"`
	UID                string `json:"uid"                header:"UID"`
	Name               string `json:"name"               header:"Name"`
	ServiceName        string `json:"service_name"       header:"Service Name"`
	ProductType        string `json:"product_type"       header:"Product Type"`
	ProvisioningStatus string `json:"provisioning_status" header:"Status"`
	RateLimit          int    `json:"rate_limit"         header:"Rate Limit (Mbps)"`
	LocationID         int    `json:"location_id"        header:"Location ID"`
	ContractTermMonths int    `json:"contract_term_months" header:"Term"`
}

// natGatewayTelemetrySampleOutput is the display struct for a single telemetry sample.
type natGatewayTelemetrySampleOutput struct {
	output.Output `json:"-" header:"-"`
	Type          string  `json:"type"      header:"Type"`
	Subtype       string  `json:"subtype"   header:"Subtype"`
	Unit          string  `json:"unit"      header:"Unit"`
	Timestamp     string  `json:"timestamp" header:"Timestamp"`
	Value         float64 `json:"value"     header:"Value"`
}

func toNATGatewayOutput(gw *megaport.NATGateway) (natGatewayOutput, error) {
	if gw == nil {
		return natGatewayOutput{}, fmt.Errorf("invalid NAT Gateway: nil value")
	}
	return natGatewayOutput{
		UID:                gw.ProductUID,
		Name:               gw.ProductName,
		LocationID:         gw.LocationID,
		Speed:              gw.Speed,
		Term:               gw.Term,
		ProvisioningStatus: gw.ProvisioningStatus,
		SessionCount:       gw.Config.SessionCount,
		DiversityZone:      gw.Config.DiversityZone,
		ASN:                gw.Config.ASN,
	}, nil
}

func printNATGateways(gateways []*megaport.NATGateway, format string, noColor bool) error {
	outputs := make([]natGatewayOutput, 0, len(gateways))
	for _, gw := range gateways {
		o, err := toNATGatewayOutput(gw)
		if err != nil {
			return err
		}
		outputs = append(outputs, o)
	}
	return output.PrintOutput(outputs, format, noColor)
}

func printNATGatewaySessions(sessions []*megaport.NATGatewaySession, format string, noColor bool) error {
	outputs := make([]natGatewaySessionOutput, 0, len(sessions))
	for _, s := range sessions {
		if s == nil {
			continue
		}
		counts := make([]string, len(s.SessionCount))
		for i, c := range s.SessionCount {
			counts[i] = fmt.Sprintf("%d", c)
		}
		outputs = append(outputs, natGatewaySessionOutput{
			SpeedMbps:     s.SpeedMbps,
			SessionCounts: strings.Join(counts, ", "),
		})
	}
	return output.PrintOutput(outputs, format, noColor)
}

func printNATGatewayTelemetry(resp *megaport.ServiceTelemetryResponse, format string, noColor bool) error {
	var outputs []natGatewayTelemetrySampleOutput
	for _, metric := range resp.Data {
		if metric == nil {
			continue
		}
		for _, sample := range metric.Samples {
			ts := time.UnixMilli(sample.Timestamp).UTC().Format(time.RFC3339)
			outputs = append(outputs, natGatewayTelemetrySampleOutput{
				Type:      metric.Type,
				Subtype:   metric.Subtype,
				Unit:      metric.Unit.Name,
				Timestamp: ts,
				Value:     sample.Value,
			})
		}
	}
	return output.PrintOutput(outputs, format, noColor)
}

func printNATGatewayValidateResult(res *megaport.NATGatewayValidateResult, format string, noColor bool) error {
	if res == nil {
		return fmt.Errorf("invalid NAT Gateway validate result: nil value")
	}
	out := natGatewayValidateOutput{
		UID:          res.ProductUID,
		ProductType:  res.ProductType,
		Metro:        res.Metro,
		Currency:     res.Price.Currency,
		MonthlyRate:  res.Price.MonthlyRate,
		HourlyRate:   res.Price.HourlyRate,
		MonthlySetup: res.Price.MonthlySetup,
	}
	return output.PrintOutput([]natGatewayValidateOutput{out}, format, noColor)
}

func printNATGatewayBuyResult(res *megaport.NATGatewayBuyResult, format string, noColor bool) error {
	if res == nil {
		return fmt.Errorf("invalid NAT Gateway buy result: nil value")
	}
	out := natGatewayBuyOutput{
		UID:                res.ProductUID,
		Name:               res.ProductName,
		ServiceName:        res.ServiceName,
		ProductType:        res.ProductType,
		ProvisioningStatus: res.ProvisioningStatus,
		RateLimit:          res.RateLimit,
		LocationID:         res.LocationID,
		ContractTermMonths: res.ContractTermMonths,
	}
	return output.PrintOutput([]natGatewayBuyOutput{out}, format, noColor)
}

// exportNATGatewayConfig returns a map suitable for use as JSON input to create.
func exportNATGatewayConfig(gw *megaport.NATGateway) map[string]interface{} {
	m := map[string]interface{}{
		"name":       gw.ProductName,
		"term":       gw.Term,
		"speed":      gw.Speed,
		"locationId": gw.LocationID,
	}
	if gw.Config.SessionCount > 0 {
		m["sessionCount"] = gw.Config.SessionCount
	}
	if gw.Config.DiversityZone != "" {
		m["diversityZone"] = gw.Config.DiversityZone
	}
	if gw.Config.ASN != 0 {
		m["asn"] = gw.Config.ASN
	}
	return m
}

// displayNATGatewayChanges prints a before/after diff for an update.
func displayNATGatewayChanges(original, updated *megaport.NATGateway, noColor bool) {
	if original == nil || updated == nil {
		return
	}
	changes := []output.FieldChange{
		{Label: "Name", OldValue: original.ProductName, NewValue: updated.ProductName},
		{Label: "Speed", OldValue: fmt.Sprintf("%d Mbps", original.Speed), NewValue: fmt.Sprintf("%d Mbps", updated.Speed)},
		{Label: "Term", OldValue: fmt.Sprintf("%d months", original.Term), NewValue: fmt.Sprintf("%d months", updated.Term)},
		{Label: "Session Count", OldValue: fmt.Sprintf("%d", original.Config.SessionCount), NewValue: fmt.Sprintf("%d", updated.Config.SessionCount)},
		{Label: "Diversity Zone", OldValue: original.Config.DiversityZone, NewValue: updated.Config.DiversityZone},
	}
	output.DisplayChanges(changes, noColor)
}
