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
		err := printMVEs(testMVEs, "table")
		assert.NoError(t, err)
	})

	expected := `uid     name         location_id
mve-1   MyMVEOne     1
mve-2   AnotherMVE   2
`
	assert.Equal(t, expected, output)
}

func TestPrintMVEs_JSON(t *testing.T) {
	output := captureOutput(func() {
		err := printMVEs(testMVEs, "json")
		assert.NoError(t, err)
	})

	expected := `[
  {
    "uid": "mve-1",
    "name": "MyMVEOne",
    "location_id": 1
  },
  {
    "uid": "mve-2",
    "name": "AnotherMVE",
    "location_id": 2
  }
]`
	assert.JSONEq(t, expected, output)
}

func TestPrintMVEs_CSV(t *testing.T) {
	output := captureOutput(func() {
		err := printMVEs(testMVEs, "csv")
		assert.NoError(t, err)
	})

	expected := `uid,name,location_id
mve-1,MyMVEOne,1
mve-2,AnotherMVE,2
`
	assert.Equal(t, expected, output)
}

func TestPrintMVEs_Invalid(t *testing.T) {
	var err error
	output := captureOutput(func() {
		err = printMVEs(testMVEs, "invalid")
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid output format")
	assert.Empty(t, output)
}
