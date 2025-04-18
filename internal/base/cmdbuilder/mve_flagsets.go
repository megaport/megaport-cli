package cmdbuilder

// WithMVECommonFlags adds common flags for MVE operations
func (b *CommandBuilder) WithMVECommonFlags() *CommandBuilder {
	b.WithFlag("name", "", "MVE name")
	b.WithFlag("cost-centre", "", "Cost centre for billing")
	return b
}

// WithMVECreateFlags adds flags needed for MVE creation
func (b *CommandBuilder) WithMVECreateFlags() *CommandBuilder {
	b.WithFlag("name", "", "The name of the MVE")
	b.WithIntFlag("term", 0, "The term of the MVE (1, 12, 24, or 36 months)")
	b.WithIntFlag("location-id", 0, "The ID of the location where the MVE will be provisioned")
	b.WithFlag("vendor-config", "", "JSON string with vendor-specific configuration (for flag mode)")
	b.WithFlag("vnics", "", "JSON array of network interfaces (for flag mode)")

	b.WithFlag("diversity-zone", "", "The diversity zone for the MVE")
	b.WithFlag("promo-code", "", "Promotional code for discounts")
	b.WithFlag("cost-centre", "", "Cost centre for billing")
	b.WithResourceTagFlags()
	return b
}

// WithMVEUpdateFlags adds flags for updating an MVE
func (b *CommandBuilder) WithMVEUpdateFlags() *CommandBuilder {
	// All update flags are optional
	b.WithFlag("name", "", "The new name of the MVE (1-64 characters)")
	b.WithFlag("cost-centre", "", "The new cost centre for billing purposes")
	b.WithIntFlag("term", 0, "New contract term in months (1, 12, 24, or 36)")
	return b
}

// WithMVEImageFilterFlags adds flags for filtering MVE images
func (b *CommandBuilder) WithMVEImageFilterFlags() *CommandBuilder {
	b.WithFlag("vendor", "", "Filter images by vendor")
	b.WithFlag("product-code", "", "Filter images by product code")
	b.WithIntFlag("id", 0, "Filter images by ID")
	b.WithFlag("version", "", "Filter images by version")
	b.WithBoolFlag("release-image", false, "Filter images by release image (only show release images)")
	return b
}

// WithMVEFilterFlags adds flags for filtering MVE lists
func (b *CommandBuilder) WithMVEFilterFlags() *CommandBuilder {
	b.WithIntFlag("location-id", 0, "Filter MVEs by location ID")
	b.WithFlag("vendor", "", "Filter MVEs by vendor")
	b.WithFlag("name", "", "Filter MVEs by name")
	b.WithBoolFlag("include-inactive", false, "Include inactive MVEs in the list")
	return b
}
