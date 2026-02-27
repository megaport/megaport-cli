package cmdbuilder

// WithIXCreateFlags adds flags for IX creation
func (b *CommandBuilder) WithIXCreateFlags() *CommandBuilder {
	b.WithFlag("product-uid", "", "The UID of the port to attach the IX to").
		WithFlag("name", "", "The name of the IX").
		WithFlag("network-service-type", "", "The IX type/network service to connect to (e.g. \"Los Angeles IX\")").
		WithIntFlag("asn", 0, "ASN (Autonomous System Number) for BGP peering").
		WithFlag("mac-address", "", "MAC address for the IX interface").
		WithIntFlag("rate-limit", 0, "Rate limit in Mbps").
		WithIntFlag("vlan", 0, "VLAN ID for the IX connection").
		WithBoolFlag("shutdown", false, "Whether the IX is initially shut down").
		WithFlag("promo-code", "", "Optional promotion code for discounts").
		WithOptionalFlag("shutdown", "Whether the IX is initially shut down").
		WithOptionalFlag("promo-code", "Optional promotion code for discounts")
	return b
}

// WithIXUpdateFlags adds flags for IX updates
func (b *CommandBuilder) WithIXUpdateFlags() *CommandBuilder {
	b.WithFlag("name", "", "The new name of the IX").
		WithIntFlag("rate-limit", 0, "Rate limit in Mbps").
		WithFlag("cost-centre", "", "Cost centre for invoicing purposes").
		WithIntFlag("vlan", 0, "VLAN ID for the IX connection").
		WithFlag("mac-address", "", "MAC address for the IX interface").
		WithIntFlag("asn", 0, "ASN (Autonomous System Number) for BGP peering").
		WithFlag("password", "", "BGP password").
		WithBoolFlag("public-graph", false, "Whether the IX usage statistics are publicly viewable").
		WithFlag("reverse-dns", "", "DNS lookup of a domain name from an IP address").
		WithFlag("a-end-product-uid", "", "Move the IX by changing the A-End of the IX").
		WithBoolFlag("shutdown", false, "Shut down or re-enable the IX").
		WithOptionalFlag("name", "The new name of the IX").
		WithOptionalFlag("rate-limit", "Rate limit in Mbps").
		WithOptionalFlag("cost-centre", "Cost centre for invoicing purposes").
		WithOptionalFlag("vlan", "VLAN ID for the IX connection").
		WithOptionalFlag("mac-address", "MAC address for the IX interface").
		WithOptionalFlag("asn", "ASN (Autonomous System Number) for BGP peering").
		WithOptionalFlag("password", "BGP password").
		WithOptionalFlag("public-graph", "Whether the IX usage statistics are publicly viewable").
		WithOptionalFlag("reverse-dns", "DNS lookup of a domain name from an IP address").
		WithOptionalFlag("a-end-product-uid", "Move the IX by changing the A-End of the IX").
		WithOptionalFlag("shutdown", "Shut down or re-enable the IX")
	return b
}

// WithIXFilterFlags adds flags for filtering IX lists
func (b *CommandBuilder) WithIXFilterFlags() *CommandBuilder {
	b.WithFlag("name", "", "Filter IXs by name (partial match)").
		WithIntFlag("asn", 0, "Filter IXs by ASN").
		WithIntFlag("vlan", 0, "Filter IXs by VLAN").
		WithFlag("network-service-type", "", "Filter IXs by network service type").
		WithIntFlag("location-id", 0, "Filter IXs by location ID").
		WithIntFlag("rate-limit", 0, "Filter IXs by rate limit in Mbps").
		WithBoolFlag("include-inactive", false, "Include inactive IXs in the list")
	return b
}
