package cmd

import (
	"testing"

	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
)

var testMVEs = []*megaport.MVE{
	{
		UID:        "mve-1",
		Name:       "MyMVEOne",
		LocationID: 1,
	},
	{
		UID:        "mve-2",
		Name:       "AnotherMVE",
		LocationID: 2,
	},
}

func TestPrintMVEs_Table(t *testing.T) {
	output := captureOutput(func() {
		printMVEs(testMVEs, "table")
	})

	// Table output should contain headers and both MVE UIDs
	assert.Contains(t, output, "UID")
	assert.Contains(t, output, "MyMVEOne")
	assert.Contains(t, output, "AnotherMVE")
}

func TestPrintMVEs_JSON(t *testing.T) {
	output := captureOutput(func() {
		printMVEs(testMVEs, "json")
	})

	// JSON output should contain an array of objects
	assert.Contains(t, output, `"uid":"mve-1"`)
	assert.Contains(t, output, `"uid":"mve-2"`)
}

func TestPrintMVEs_Invalid(t *testing.T) {
	output := captureOutput(func() {
		printMVEs(testMVEs, "invalid")
	})

	assert.Contains(t, output, "Invalid output format")
}
