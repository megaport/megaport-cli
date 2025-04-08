package cmdbuilder

// WithVXCCommonFlags adds common flags for VXC operations
func (b *CommandBuilder) WithVXCCommonFlags() *CommandBuilder {
	b.WithFlag("name", "", "Name of the VXC")
	b.WithIntFlag("rate-limit", 0, "Bandwidth in Mbps")
	b.WithIntFlag("term", 0, "Contract term in months (1, 12, 24, or 36)")
	b.WithFlag("cost-centre", "", "Cost centre for billing")
	return b
}

// WithVXCEndpointFlags adds endpoint-related flags for VXC operations
func (b *CommandBuilder) WithVXCEndpointFlags() *CommandBuilder {
	b.WithFlag("a-end-uid", "", "UID of the A-End product")
	b.WithFlag("b-end-uid", "", "UID of the B-End product")
	b.WithIntFlag("a-end-vlan", 0, "VLAN for A-End (0-4093, except 1)")
	b.WithIntFlag("b-end-vlan", 0, "VLAN for B-End (0-4093, except 1)")
	b.WithIntFlag("a-end-inner-vlan", 0, "Inner VLAN for A-End (-1 or higher)")
	b.WithIntFlag("b-end-inner-vlan", 0, "Inner VLAN for B-End (-1 or higher)")
	return b
}

// WithVXCPartnerConfigFlags adds partner configuration flags for VXCs
func (b *CommandBuilder) WithVXCPartnerConfigFlags() *CommandBuilder {
	b.WithFlag("a-end-partner-config", "", "JSON string with A-End partner configuration")
	b.WithFlag("b-end-partner-config", "", "JSON string with B-End partner configuration")
	return b
}

// WithVXCCreateFlags adds all flags needed for VXC creation
func (b *CommandBuilder) WithVXCCreateFlags() *CommandBuilder {
	b.WithVXCCommonFlags()
	b.WithVXCEndpointFlags()
	b.WithVXCPartnerConfigFlags()
	b.WithIntFlag("a-end-vnic-index", 0, "vNIC index for A-End MVE")
	b.WithIntFlag("b-end-vnic-index", 0, "vNIC index for B-End MVE")
	b.WithFlag("promo-code", "", "Promotional code")
	b.WithFlag("service-key", "", "Service key")
	return b
}

// WithVXCUpdateFlags adds all flags needed for VXC updates
func (b *CommandBuilder) WithVXCUpdateFlags() *CommandBuilder {
	b.WithVXCCommonFlags()
	b.WithVXCEndpointFlags()
	b.WithVXCPartnerConfigFlags()
	b.WithBoolFlag("shutdown", false, "Whether to shut down the VXC")
	return b
}
