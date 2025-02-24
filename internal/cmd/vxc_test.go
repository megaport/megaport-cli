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
		err := printVXCs(testVXCs, "table")
		assert.NoError(t, err)
	})

	expected := `uid     name         a_end_uid   b_end_uid
vxc-1   MyVXCOne     a-end-1     b-end-1
vxc-2   AnotherVXC   a-end-2     b-end-2
`
	assert.Equal(t, expected, output)
}
func TestPrintVXCs_JSON(t *testing.T) {
	output := captureOutput(func() {
		err := printVXCs(testVXCs, "json")
		assert.NoError(t, err)
	})

	expected := `[
  {
    "uid": "vxc-1",
    "name": "MyVXCOne",
    "a_end_uid": "a-end-1",
    "b_end_uid": "b-end-1"
  },
  {
    "uid": "vxc-2",
    "name": "AnotherVXC",
    "a_end_uid": "a-end-2",
    "b_end_uid": "b-end-2"
  }
]`
	assert.JSONEq(t, expected, output)
}

func TestPrintVXCs_Invalid(t *testing.T) {
	var err error
	output := captureOutput(func() {
		err = printVXCs(testVXCs, "invalid")
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid output format")
	assert.Empty(t, output)
}
