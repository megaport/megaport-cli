package billing_market

import (
	"testing"

	"github.com/megaport/megaport-cli/internal/base/output"
	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
)

func TestToBillingMarketOutput_Nil(t *testing.T) {
	assert.NotPanics(t, func() {
		out := toBillingMarketOutput(nil)
		assert.Equal(t, billingMarketOutput{}, out)
	})
}

func TestToBillingMarketOutput_AllFields(t *testing.T) {
	m := &megaport.BillingMarket{
		ID:                  7,
		SupplierName:        "Megaport EU",
		CurrencyEnum:        "EUR",
		Country:             "DE",
		Region:              "Europe",
		BillingContactPhone: "+49000",
		TaxRate:             19.5,
		SecondPartyID:       321,
		PaymentTermInDays:   30,
		VATExempt:           true,
		Active:              false,
	}

	out := toBillingMarketOutput(m)
	assert.Equal(t, 7, out.ID)
	assert.Equal(t, "EUR", out.CurrencyEnum)
	assert.Equal(t, "+49000", out.BillingContactPhone)
	assert.Equal(t, 19.5, out.TaxRate)
	assert.Equal(t, 321, out.SecondPartyID)
	assert.Equal(t, 30, out.PaymentTermInDays)
	assert.True(t, out.VATExempt)
	assert.False(t, out.Active)
}

func TestBillingMarketOutput_XML(t *testing.T) {
	outputs := make([]billingMarketOutput, 0, len(mockBillingMarkets))
	for _, bm := range mockBillingMarkets {
		outputs = append(outputs, toBillingMarketOutput(bm))
	}

	out := output.CaptureOutput(func() {
		err := output.PrintOutput(outputs, "xml", true)
		assert.NoError(t, err)
	})

	assert.Contains(t, out, "<id>1</id>")
	assert.Contains(t, out, "<supplier_name>Megaport US</supplier_name>")
	assert.Contains(t, out, "<currency>USD</currency>")
	assert.Contains(t, out, "<billing_contact_email>john@example.com</billing_contact_email>")
	assert.Contains(t, out, "<active>true</active>")
	assert.Contains(t, out, "Megaport AU")
}

func TestBillingMarketOutput_EmptySlice(t *testing.T) {
	cases := map[string]func(string){
		"table": func(out string) { assert.Contains(t, out, "SUPPLIER NAME") },
		"json":  func(out string) { assert.Equal(t, "[]\n", out) },
		"csv":   func(out string) { assert.Contains(t, out, "id,supplier_name,currency") },
		"xml":   func(out string) { assert.Contains(t, out, "<items></items>") },
	}
	for format, check := range cases {
		t.Run(format, func(t *testing.T) {
			out := output.CaptureOutput(func() {
				err := output.PrintOutput([]billingMarketOutput{}, format, true)
				assert.NoError(t, err)
			})
			check(out)
		})
	}
}
