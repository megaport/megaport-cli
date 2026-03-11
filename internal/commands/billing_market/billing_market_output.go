package billing_market

import (
	"github.com/megaport/megaport-cli/internal/base/output"
	megaport "github.com/megaport/megaportgo"
)

type BillingMarketOutput struct {
	output.Output       `json:"-" header:"-"`
	ID                  int     `json:"id" header:"ID"`
	SupplierName        string  `json:"supplier_name" header:"Supplier Name"`
	CurrencyEnum        string  `json:"currency" header:"Currency"`
	Country             string  `json:"country" header:"Country"`
	Region              string  `json:"region" header:"Region"`
	BillingContactName  string  `json:"billing_contact_name" header:"Contact Name"`
	BillingContactEmail string  `json:"billing_contact_email" header:"Contact Email"`
	BillingContactPhone string  `json:"billing_contact_phone" header:"-"`
	Address1            string  `json:"address1" header:"-"`
	City                string  `json:"city" header:"City"`
	State               string  `json:"state" header:"-"`
	Postcode            string  `json:"postcode" header:"-"`
	Language            string  `json:"language" header:"-"`
	TaxRate             float64 `json:"tax_rate" header:"-"`
	FirstPartyID        int     `json:"first_party_id" header:"-"`
	SecondPartyID       int     `json:"second_party_id" header:"-"`
	PaymentTermInDays   int     `json:"payment_term_in_days" header:"-"`
	Active              bool    `json:"active" header:"Active"`
	VATExempt           bool    `json:"vat_exempt" header:"-"`
}

func ToBillingMarketOutput(m *megaport.BillingMarket) BillingMarketOutput {
	return BillingMarketOutput{
		ID:                  m.ID,
		SupplierName:        m.SupplierName,
		CurrencyEnum:        m.CurrencyEnum,
		Country:             m.Country,
		Region:              m.Region,
		BillingContactName:  m.BillingContactName,
		BillingContactEmail: m.BillingContactEmail,
		BillingContactPhone: m.BillingContactPhone,
		Address1:            m.Address1,
		City:                m.City,
		State:               m.State,
		Postcode:            m.Postcode,
		Language:            m.Language,
		TaxRate:             m.TaxRate,
		FirstPartyID:        m.FirstPartyID,
		SecondPartyID:       m.SecondPartyID,
		PaymentTermInDays:   m.PaymentTermInDays,
		Active:              m.Active,
		VATExempt:           m.VATExempt,
	}
}
