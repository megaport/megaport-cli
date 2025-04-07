package cmdbuilder

// WithMCRCreateFlags adds flags for MCR creation
func (b *CommandBuilder) WithMCRCreateFlags() *CommandBuilder {
	b.WithFlag("name", "", "The name of the MCR (1-64 characters)").
		WithRequiredFlag("name", "The name of the MCR (1-64 characters)").
		WithIntFlag("term", 0, "The contract term for the MCR (1, 12, 24, or 36 months)").
		WithRequiredFlag("term", "The contract term for the MCR (1, 12, 24, or 36 months)").
		WithIntFlag("port-speed", 0, "The speed of the MCR (1000, 2500, 5000, or 10000 Mbps)").
		WithRequiredFlag("port-speed", "The speed of the MCR (1000, 2500, 5000, or 10000 Mbps)").
		WithIntFlag("location-id", 0, "The ID of the location where the MCR will be provisioned").
		WithRequiredFlag("location-id", "The ID of the location where the MCR will be provisioned").
		WithIntFlag("mcr-asn", 0, "The ASN for the MCR (64512-65534 for private ASN, or a public ASN)").
		WithOptionalFlag("mcr-asn", "The ASN for the MCR (64512-65534 for private ASN, or a public ASN)").
		WithFlag("diversity-zone", "", "The diversity zone for the MCR").
		WithOptionalFlag("diversity-zone", "The diversity zone for the MCR").
		WithFlag("cost-centre", "", "The cost centre for billing purposes").
		WithOptionalFlag("cost-centre", "The cost centre for billing purposes").
		WithFlag("marketplace-visibility", "", "Whether the MCR is visible in the marketplace (true/false)").
		WithRequiredFlag("marketplace-visibility", "Whether the MCR is visible in the marketplace (true/false)").
		WithFlag("promo-code", "", "A promotional code for discounts").
		WithOptionalFlag("promo-code", "A promotional code for discounts")
	return b
}

// WithMCRUpdateFlags adds flags for MCR updates
func (b *CommandBuilder) WithMCRUpdateFlags() *CommandBuilder {

	b.WithFlag("name", "", "The new name of the MCR (1-64 characters)").
		WithFlag("cost-centre", "", "The new cost centre for the MCR").
		WithBoolFlag("marketplace-visibility", false, "Whether the MCR is visible in the marketplace (true/false)").
		WithOptionalFlag("name", "The new name of the MCR (1-64 characters)").
		WithOptionalFlag("cost-centre", "The new cost centre for the MCR").
		WithOptionalFlag("marketplace-visibility", "Whether the MCR is visible in the marketplace (true/false)")
	return b
}

// WithCreateMCRPrefixFilterListFlags adds flags for creating MCR prefix filter lists
func (b *CommandBuilder) WithCreateMCRPrefixFilterListFlags() *CommandBuilder {
	b.WithRequiredFlag("description", "The description of the prefix filter list (1-255 characters)").
		WithRequiredFlag("address-family", "The address family (IPv4 or IPv6)").
		WithRequiredFlag("entries", "JSON array of prefix filter entries")
	return b
}

// WithPrefixFilterFlags adds flags for prefix filter operations
func (b *CommandBuilder) WithPrefixFilterFlags() *CommandBuilder {
	b.WithFlag("description", "", "Description of the prefix filter list")
	b.WithFlag("address-family", "", "Address family (IPv4 or IPv6)")
	b.WithFlag("entries", "", "JSON array of prefix filter entries")
	return b
}
