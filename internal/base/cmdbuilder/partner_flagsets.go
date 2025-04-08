package cmdbuilder

// WithPartnerFilterFlags adds common partner filter flags
func (b *CommandBuilder) WithPartnerFilterFlags() *CommandBuilder {
	b.WithFlag("product-name", "", "Filter partner ports by product name")
	b.WithFlag("connect-type", "", "Filter partner ports by connect type")
	b.WithFlag("company-name", "", "Filter partner ports by company name")
	return b
}
