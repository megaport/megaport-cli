package locations

import (
	"testing"

	op "github.com/megaport/megaport-cli/internal/base/output"
	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
)

func TestToLocationOutput_Valid(t *testing.T) {
	msg := "scheduled maintenance"
	l := &megaport.LocationV3{
		ID:        3,
		Name:      "Tokyo",
		Metro:     "Tokyo",
		Market:    "APAC",
		Status:    "ACTIVE",
		Latitude:  35.6,
		Longitude: 139.7,
		Address:   megaport.LocationV3Address{Country: "Japan"},
		DataCentre: megaport.LocationV3DataCentre{
			ID:   300,
			Name: "TYO1",
		},
		OrderingMessage: &msg,
	}

	out := toLocationOutput(l)
	assert.Equal(t, 3, out.ID)
	assert.Equal(t, "Tokyo", out.Name)
	assert.Equal(t, "Japan", out.Country)
	assert.Equal(t, "APAC", out.Market)
	assert.Equal(t, 35.6, out.Latitude)
	assert.Equal(t, "TYO1", out.DataCentreName)
	assert.Equal(t, 300, out.DataCentreID)
	assert.Equal(t, "scheduled maintenance", out.OrderingMessage)
}

func TestToLocationOutput_Nil(t *testing.T) {
	assert.NotPanics(t, func() {
		out := toLocationOutput(nil)
		assert.Equal(t, locationOutput{}, out)
	})
}

func TestToLocationTableOutput_Valid(t *testing.T) {
	l := testLocations[0]
	out := toLocationTableOutput(l)
	assert.Equal(t, 1, out.ID)
	assert.Equal(t, "Sydney", out.Name)
	assert.Equal(t, "Australia", out.Country)
	assert.Equal(t, "ACTIVE", out.Status)
	assert.True(t, out.MVEAvailable)
}

func TestToLocationTableOutput_Nil(t *testing.T) {
	assert.NotPanics(t, func() {
		out := toLocationTableOutput(nil)
		assert.Equal(t, locationTableOutput{}, out)
	})
}

func TestPrintLocations_XML(t *testing.T) {
	out := op.CaptureOutput(func() {
		err := printLocations(testLocations, "xml", noColor)
		assert.NoError(t, err)
	})

	assert.Contains(t, out, "<id>1</id>")
	assert.Contains(t, out, "<name>Sydney</name>")
	assert.Contains(t, out, "<country>Australia</country>")
	assert.Contains(t, out, "<market>APAC</market>")
	assert.Contains(t, out, "London")
}

func TestPrintLocations_NilEntriesSkipped(t *testing.T) {
	locs := []*megaport.LocationV3{nil, testLocations[0], nil}
	for _, format := range []string{"csv", "table"} {
		assert.NotPanics(t, func() {
			out := op.CaptureOutput(func() {
				err := printLocations(locs, format, noColor)
				assert.NoError(t, err)
			})
			assert.Contains(t, out, "Sydney")
		})
	}
}

func TestPrintCountries_AllFormats(t *testing.T) {
	countries := []*megaport.Country{
		{Code: "AU", Name: "Australia", Prefix: "AUS", SiteCount: 12},
		nil,
		{Code: "GB", Name: "United Kingdom", Prefix: "GBR", SiteCount: 8},
	}

	xmlOut := op.CaptureOutput(func() {
		err := printCountries(countries, "xml", noColor)
		assert.NoError(t, err)
	})
	assert.Contains(t, xmlOut, "<code>AU</code>")
	assert.Contains(t, xmlOut, "<name>Australia</name>")
	assert.Contains(t, xmlOut, "<site_count>12</site_count>")

	csvOut := op.CaptureOutput(func() {
		err := printCountries(countries, "csv", noColor)
		assert.NoError(t, err)
	})
	assert.Contains(t, csvOut, "code,name,prefix,site_count")
	assert.Contains(t, csvOut, "AU,Australia,AUS,12")

	tableOut := op.CaptureOutput(func() {
		err := printCountries(countries, "table", noColor)
		assert.NoError(t, err)
	})
	assert.Contains(t, tableOut, "AU")

	jsonOut := op.CaptureOutput(func() {
		err := printCountries(countries, "json", noColor)
		assert.NoError(t, err)
	})
	assert.Contains(t, jsonOut, `"code": "AU"`)
}

func TestPrintMarketCodes_AllFormats(t *testing.T) {
	codes := []string{"AU", "US", "UK"}

	xmlOut := op.CaptureOutput(func() {
		err := printMarketCodes(codes, "xml", noColor)
		assert.NoError(t, err)
	})
	assert.Contains(t, xmlOut, "<market_code>AU</market_code>")

	csvOut := op.CaptureOutput(func() {
		err := printMarketCodes(codes, "csv", noColor)
		assert.NoError(t, err)
	})
	assert.Contains(t, csvOut, "market_code")
	assert.Contains(t, csvOut, "AU")

	tableOut := op.CaptureOutput(func() {
		err := printMarketCodes(codes, "table", noColor)
		assert.NoError(t, err)
	})
	assert.Contains(t, tableOut, "US")

	emptyOut := op.CaptureOutput(func() {
		err := printMarketCodes(nil, "json", noColor)
		assert.NoError(t, err)
	})
	assert.Equal(t, "[]\n", emptyOut)
}

func TestPrintRoundTripTimes_XML(t *testing.T) {
	rtts := []*megaport.RoundTripTime{
		{SrcLocation: 1, DstLocation: 2, MedianRTT: 12.5},
	}

	out := op.CaptureOutput(func() {
		err := printRoundTripTimes(rtts, "xml", noColor)
		assert.NoError(t, err)
	})

	assert.Contains(t, out, "<src_location_id>1</src_location_id>")
	assert.Contains(t, out, "<dst_location_id>2</dst_location_id>")
	assert.Contains(t, out, "<median_rtt_ms>12.5</median_rtt_ms>")
}
