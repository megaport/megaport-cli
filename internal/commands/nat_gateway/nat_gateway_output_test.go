package nat_gateway

import (
	"testing"

	"github.com/megaport/megaport-cli/internal/base/output"
	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
)

var testGateways = []*megaport.NATGateway{
	{
		ProductUID:         "ng-1",
		ProductName:        "Gateway One",
		LocationID:         100,
		Speed:              1000,
		Term:               12,
		ProvisioningStatus: "LIVE",
		Config: megaport.NATGatewayNetworkConfig{
			SessionCount:  500,
			DiversityZone: "blue",
			ASN:           65000,
		},
	},
	{
		ProductUID:         "ng-2",
		ProductName:        "Gateway Two",
		LocationID:         200,
		Speed:              2000,
		Term:               24,
		ProvisioningStatus: "CONFIGURED",
	},
}

// ---- printNATGateways ----

func TestPrintNATGateways_Table(t *testing.T) {
	out := output.CaptureOutput(func() {
		err := printNATGateways(testGateways, "table", true)
		assert.NoError(t, err)
	})
	assert.Contains(t, out, "UID")
	assert.Contains(t, out, "ng-1")
	assert.Contains(t, out, "Gateway One")
}

func TestPrintNATGateways_CSV(t *testing.T) {
	out := output.CaptureOutput(func() {
		err := printNATGateways(testGateways, "csv", true)
		assert.NoError(t, err)
	})
	assert.Contains(t, out, "uid")
	assert.Contains(t, out, "ng-1")
	assert.Contains(t, out, "ng-2")
}

func TestPrintNATGateways_XML(t *testing.T) {
	out := output.CaptureOutput(func() {
		err := printNATGateways(testGateways, "xml", true)
		assert.NoError(t, err)
	})
	assert.Contains(t, out, "<items>")
	assert.Contains(t, out, "ng-1")
}

func TestPrintNATGateways_Empty(t *testing.T) {
	for _, format := range []string{"table", "json", "csv", "xml"} {
		err := printNATGateways([]*megaport.NATGateway{}, format, true)
		assert.NoError(t, err)
	}
}

// ---- printNATGatewaySessions ----

var testSessions = []*megaport.NATGatewaySession{
	{SpeedMbps: 1000, SessionCount: []int{100, 200, 300}},
	{SpeedMbps: 2000, SessionCount: []int{50}},
}

func TestPrintNATGatewaySessions_Table(t *testing.T) {
	out := output.CaptureOutput(func() {
		err := printNATGatewaySessions(testSessions, "table", true)
		assert.NoError(t, err)
	})
	assert.Contains(t, out, "SPEED")
}

func TestPrintNATGatewaySessions_CSV(t *testing.T) {
	out := output.CaptureOutput(func() {
		err := printNATGatewaySessions(testSessions, "csv", true)
		assert.NoError(t, err)
	})
	assert.Contains(t, out, "speed_mbps")
	assert.Contains(t, out, "1000")
}

func TestPrintNATGatewaySessions_XML(t *testing.T) {
	out := output.CaptureOutput(func() {
		err := printNATGatewaySessions(testSessions, "xml", true)
		assert.NoError(t, err)
	})
	assert.Contains(t, out, "<items>")
}

func TestPrintNATGatewaySessions_Empty(t *testing.T) {
	for _, format := range []string{"table", "json", "csv", "xml"} {
		err := printNATGatewaySessions([]*megaport.NATGatewaySession{}, format, true)
		assert.NoError(t, err)
	}
}

func TestPrintNATGatewaySessions_NilEntries(t *testing.T) {
	sessions := []*megaport.NATGatewaySession{nil, {SpeedMbps: 500, SessionCount: []int{10}}, nil}
	out := output.CaptureOutput(func() {
		err := printNATGatewaySessions(sessions, "table", true)
		assert.NoError(t, err)
	})
	assert.Contains(t, out, "500")
}

// ---- printNATGatewayTelemetry ----

var testTelemetryResp = &megaport.ServiceTelemetryResponse{
	Data: []*megaport.TelemetryMetricData{
		{
			Type:    "BITS",
			Subtype: "ingress",
			Unit:    megaport.TelemetryUnit{Name: "bps"},
			Samples: []megaport.TelemetrySample{
				{Timestamp: 1700000000000, Value: 1234.5},
			},
		},
	},
}

func TestPrintNATGatewayTelemetry_Table(t *testing.T) {
	out := output.CaptureOutput(func() {
		err := printNATGatewayTelemetry(testTelemetryResp, "table", true)
		assert.NoError(t, err)
	})
	assert.Contains(t, out, "BITS")
}

func TestPrintNATGatewayTelemetry_CSV(t *testing.T) {
	out := output.CaptureOutput(func() {
		err := printNATGatewayTelemetry(testTelemetryResp, "csv", true)
		assert.NoError(t, err)
	})
	assert.Contains(t, out, "type")
	assert.Contains(t, out, "BITS")
}

func TestPrintNATGatewayTelemetry_XML(t *testing.T) {
	out := output.CaptureOutput(func() {
		err := printNATGatewayTelemetry(testTelemetryResp, "xml", true)
		assert.NoError(t, err)
	})
	assert.Contains(t, out, "<items>")
}

// ---- printNATGatewayValidateResult ----

var testValidateResult = &megaport.NATGatewayValidateResult{
	ProductUID:  "uid-val-1",
	ProductType: "NAT_GATEWAY",
	Metro:       "Sydney",
	Price: megaport.NATGatewayOrderPrice{
		Currency:     "AUD",
		MonthlyRate:  99.99,
		HourlyRate:   0.15,
		MonthlySetup: 0,
	},
}

func TestPrintNATGatewayValidateResult_Nil(t *testing.T) {
	err := printNATGatewayValidateResult(nil, "table", true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nil value")
}

func TestPrintNATGatewayValidateResult_Table(t *testing.T) {
	out := output.CaptureOutput(func() {
		err := printNATGatewayValidateResult(testValidateResult, "table", true)
		assert.NoError(t, err)
	})
	assert.Contains(t, out, "uid-val-1")
}

func TestPrintNATGatewayValidateResult_JSON(t *testing.T) {
	out := output.CaptureOutput(func() {
		err := printNATGatewayValidateResult(testValidateResult, "json", true)
		assert.NoError(t, err)
	})
	assert.Contains(t, out, "uid-val-1")
	assert.Contains(t, out, "NAT_GATEWAY")
}

func TestPrintNATGatewayValidateResult_CSV(t *testing.T) {
	out := output.CaptureOutput(func() {
		err := printNATGatewayValidateResult(testValidateResult, "csv", true)
		assert.NoError(t, err)
	})
	assert.Contains(t, out, "uid")
}

func TestPrintNATGatewayValidateResult_XML(t *testing.T) {
	out := output.CaptureOutput(func() {
		err := printNATGatewayValidateResult(testValidateResult, "xml", true)
		assert.NoError(t, err)
	})
	assert.Contains(t, out, "<items>")
	assert.Contains(t, out, "uid-val-1")
}

// ---- printNATGatewayBuyResult ----

var testBuyResult = &megaport.NATGatewayBuyResult{
	ProductUID:         "uid-buy-1",
	ProductName:        "Purchased GW",
	ServiceName:        "NAT Gateway Service",
	ProductType:        "NAT_GATEWAY",
	ProvisioningStatus: "LIVE",
	RateLimit:          1000,
	LocationID:         1,
	ContractTermMonths: 12,
}

func TestPrintNATGatewayBuyResult_Nil(t *testing.T) {
	err := printNATGatewayBuyResult(nil, "table", true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nil value")
}

func TestPrintNATGatewayBuyResult_Table(t *testing.T) {
	out := output.CaptureOutput(func() {
		err := printNATGatewayBuyResult(testBuyResult, "table", true)
		assert.NoError(t, err)
	})
	assert.Contains(t, out, "uid-buy-1")
}

func TestPrintNATGatewayBuyResult_JSON(t *testing.T) {
	out := output.CaptureOutput(func() {
		err := printNATGatewayBuyResult(testBuyResult, "json", true)
		assert.NoError(t, err)
	})
	assert.Contains(t, out, "uid-buy-1")
	assert.Contains(t, out, "Purchased GW")
}

func TestPrintNATGatewayBuyResult_CSV(t *testing.T) {
	out := output.CaptureOutput(func() {
		err := printNATGatewayBuyResult(testBuyResult, "csv", true)
		assert.NoError(t, err)
	})
	assert.Contains(t, out, "uid")
}

func TestPrintNATGatewayBuyResult_XML(t *testing.T) {
	out := output.CaptureOutput(func() {
		err := printNATGatewayBuyResult(testBuyResult, "xml", true)
		assert.NoError(t, err)
	})
	assert.Contains(t, out, "<items>")
	assert.Contains(t, out, "uid-buy-1")
}
