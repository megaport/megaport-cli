package cmdbuilder

// WithMVECommonFlags adds common flags for MVE operations
func (b *CommandBuilder) WithMVECommonFlags() *CommandBuilder {
	b.WithFlag("name", "", "MVE name")
	b.WithFlag("cost-centre", "", "Cost centre for billing")
	return b
}

// WithMVECreateFlags adds flags needed for MVE creation
func (b *CommandBuilder) WithMVECreateFlags() *CommandBuilder {
	b.WithRequiredFlag("name", "The name of the MVE").
		WithRequiredFlag("term", "The term of the MVE (1, 12, 24, or 36 months)").
		WithRequiredFlag("location-id", "The ID of the location where the MVE will be provisioned").
		WithRequiredFlag("vendor-config", "JSON string with vendor-specific configuration (for flag mode)").
		WithRequiredFlag("vnics", "JSON array of network interfaces (for flag mode)").
		WithOptionalFlag("diversity-zone", "The diversity zone for the MVE").
		WithOptionalFlag("promo-code", "Promotional code for discounts").
		WithOptionalFlag("cost-centre", "Cost centre for billing")
	return b
}

// WithMVEUpdateFlags adds flags for updating an MVE
func (b *CommandBuilder) WithMVEUpdateFlags() *CommandBuilder {
	b.WithMVECommonFlags()
	b.WithFlag("contract-term", "0", "New contract term in months (1, 12, 24, or 36)")
	return b
}

// WithMVEImageFilterFlags adds flags for filtering MVE images
func (b *CommandBuilder) WithMVEImageFilterFlags() *CommandBuilder {
	b.WithFlag("vendor", "", "Filter images by vendor")
	b.WithFlag("product-code", "", "Filter images by product code")
	b.WithIntFlag("id", 0, "Filter images by ID")
	b.WithFlag("version", "", "Filter images by version")
	b.WithFlag("release-image", "false", "Filter images by release image")
	return b
}
