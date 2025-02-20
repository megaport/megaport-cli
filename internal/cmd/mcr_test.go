package cmd

import (
	"testing"

	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
)

var testMCRs = []*megaport.MCR{
	{
		UID:        "mcr-1",
		Name:       "MyMCROne",
		LocationID: 1,
	},
	{
		UID:        "mcr-2",
		Name:       "AnotherMCR",
		LocationID: 2,
	},
}

func TestPrintMCRs_Table(t *testing.T) {
	output := captureOutput(func() {
		printMCRs(testMCRs, "table")
	})

	// Table output should contain headers and both MCR UIDs
	assert.Contains(t, output, "UID")
	assert.Contains(t, output, "MyMCROne")
	assert.Contains(t, output, "AnotherMCR")
}

func TestPrintMCRs_JSON(t *testing.T) {
	output := captureOutput(func() {
		printMCRs(testMCRs, "json")
	})

	// JSON output should contain an array of objects
	assert.Contains(t, output, `"uid":"mcr-1"`)
	assert.Contains(t, output, `"uid":"mcr-2"`)
}

func TestPrintMCRs_Invalid(t *testing.T) {
	output := captureOutput(func() {
		printMCRs(testMCRs, "invalid")
	})

	assert.Contains(t, output, "Invalid output format")
}
