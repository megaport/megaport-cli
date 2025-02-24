package cmd

import (
	"testing"

	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
)

var testMCRs = []*megaport.MCR{
	{
		UID:                "mcr-1",
		Name:               "MyMCROne",
		LocationID:         1,
		ProvisioningStatus: "ACTIVE",
	},
	{
		UID:                "mcr-2",
		Name:               "AnotherMCR",
		LocationID:         2,
		ProvisioningStatus: "INACTIVE",
	},
}

func TestPrintMCRs_Table(t *testing.T) {
	output := captureOutput(func() {
		err := printMCRs(testMCRs, "table")
		assert.NoError(t, err)
	})

	expected := `uid     name         location_id   provisioning_status
mcr-1   MyMCROne     1             ACTIVE
mcr-2   AnotherMCR   2             INACTIVE
`
	assert.Equal(t, expected, output)
}

func TestPrintMCRs_JSON(t *testing.T) {
	output := captureOutput(func() {
		err := printMCRs(testMCRs, "json")
		assert.NoError(t, err)
	})

	expected := `[
  {
    "uid": "mcr-1",
    "name": "MyMCROne",
    "location_id": 1,
    "provisioning_status": "ACTIVE"
  },
  {
    "uid": "mcr-2",
    "name": "AnotherMCR",
    "location_id": 2,
    "provisioning_status": "INACTIVE"
  }
]`
	assert.JSONEq(t, expected, output)
}

func TestPrintMCRs_Invalid(t *testing.T) {
	var err error
	output := captureOutput(func() {
		err = printMCRs(testMCRs, "invalid")
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid output format")
	assert.Empty(t, output)
}
