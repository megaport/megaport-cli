package cmd

import (
	"testing"

	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
)

var testVXCs = []*megaport.VXC{
	{
		UID:  "vxc-1",
		Name: "MyVXCOne",
		AEndConfiguration: megaport.VXCEndConfiguration{
			UID: "a-end-1",
		},
		BEndConfiguration: megaport.VXCEndConfiguration{
			UID: "b-end-1",
		},
	},
	{
		UID:  "vxc-2",
		Name: "AnotherVXC",
		AEndConfiguration: megaport.VXCEndConfiguration{
			UID: "a-end-2",
		},
		BEndConfiguration: megaport.VXCEndConfiguration{
			UID: "b-end-2",
		},
	},
}

func TestPrintVXCs_Table(t *testing.T) {
	output := captureOutput(func() {
		printVXCs(testVXCs, "table")
	})

	// Table output should contain headers and both VXC UIDs
	assert.Contains(t, output, "UID")
	assert.Contains(t, output, "MyVXCOne")
	assert.Contains(t, output, "AnotherVXC")
	assert.Contains(t, output, "a-end-1")
	assert.Contains(t, output, "b-end-1")
	assert.Contains(t, output, "a-end-2")
	assert.Contains(t, output, "b-end-2")
}

func TestPrintVXCs_JSON(t *testing.T) {
	output := captureOutput(func() {
		printVXCs(testVXCs, "json")
	})

	// JSON output should contain an array of objects
	assert.Contains(t, output, `"uid":"vxc-1"`)
	assert.Contains(t, output, `"uid":"vxc-2"`)
	assert.Contains(t, output, `"a_end_uid":"a-end-1"`)
	assert.Contains(t, output, `"b_end_uid":"b-end-1"`)
	assert.Contains(t, output, `"a_end_uid":"a-end-2"`)
	assert.Contains(t, output, `"b_end_uid":"b-end-2"`)
}

func TestPrintVXCs_Invalid(t *testing.T) {
	output := captureOutput(func() {
		printVXCs(testVXCs, "invalid")
	})

	assert.Contains(t, output, "Invalid output format")
}
