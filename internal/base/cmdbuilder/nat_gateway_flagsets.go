package cmdbuilder

// WithNATGatewayCreateFlags adds flags for NAT Gateway creation.
func (b *CommandBuilder) WithNATGatewayCreateFlags() *CommandBuilder {
	b.WithFlag("name", "", "The name of the NAT Gateway (required)").
		WithIntFlag("term", 0, "The contract term in months (1, 12, 24, or 36)").
		WithIntFlag("speed", 0, "The speed of the NAT Gateway in Mbps").
		WithIntFlag("location-id", 0, "The ID of the location where the NAT Gateway will be provisioned").
		WithIntFlag("session-count", 0, "The number of NAT sessions (optional)").
		WithOptionalFlag("session-count", "The number of NAT sessions").
		WithFlag("diversity-zone", "", "The diversity zone for the NAT Gateway (optional)").
		WithOptionalFlag("diversity-zone", "The diversity zone for the NAT Gateway").
		WithFlag("promo-code", "", "A promotional code for discounts (optional)").
		WithOptionalFlag("promo-code", "A promotional code for discounts").
		WithFlag("service-level-reference", "", "A service level reference for the NAT Gateway (optional)").
		WithOptionalFlag("service-level-reference", "A service level reference for the NAT Gateway").
		WithBoolFlag("auto-renew", false, "Whether to automatically renew the contract term").
		WithOptionalFlag("auto-renew", "Whether to automatically renew the contract term").
		WithResourceTagFlags()
	return b
}

// WithNATGatewayUpdateFlags adds flags for NAT Gateway updates.
func (b *CommandBuilder) WithNATGatewayUpdateFlags() *CommandBuilder {
	b.WithFlag("name", "", "The new name of the NAT Gateway").
		WithOptionalFlag("name", "The new name of the NAT Gateway").
		WithIntFlag("term", 0, "The new contract term in months (1, 12, 24, or 36)").
		WithOptionalFlag("term", "The new contract term in months").
		WithIntFlag("speed", 0, "The new speed of the NAT Gateway in Mbps").
		WithOptionalFlag("speed", "The new speed of the NAT Gateway in Mbps").
		WithIntFlag("location-id", 0, "The new location ID").
		WithOptionalFlag("location-id", "The new location ID").
		WithIntFlag("session-count", 0, "The new session count").
		WithOptionalFlag("session-count", "The new session count").
		WithFlag("diversity-zone", "", "The new diversity zone").
		WithOptionalFlag("diversity-zone", "The new diversity zone").
		WithFlag("promo-code", "", "A promotional code").
		WithOptionalFlag("promo-code", "A promotional code").
		WithFlag("service-level-reference", "", "A service level reference").
		WithOptionalFlag("service-level-reference", "A service level reference").
		WithBoolFlag("auto-renew", false, "Whether to automatically renew the contract term").
		WithOptionalFlag("auto-renew", "Whether to automatically renew the contract term").
		WithResourceTagFlags()
	return b
}

// WithNATGatewayFilterFlags adds flags for filtering NAT Gateway lists.
func (b *CommandBuilder) WithNATGatewayFilterFlags() *CommandBuilder {
	b.WithIntFlag("location-id", 0, "Filter NAT Gateways by location ID").
		WithFlag("name", "", "Filter NAT Gateways by name (substring match)").
		WithBoolFlag("include-inactive", false, "Include inactive NAT Gateways in the list")
	return b
}

// WithNATGatewayTelemetryFlags adds flags for the NAT Gateway telemetry command.
func (b *CommandBuilder) WithNATGatewayTelemetryFlags() *CommandBuilder {
	b.WithFlag("types", "", "Comma-separated telemetry types to retrieve (e.g. BITS,PACKETS,SPEED)").
		WithIntFlag("days", 0, "Number of days of telemetry to retrieve (1-180); mutually exclusive with --from/--to").
		WithOptionalFlag("days", "Number of days of telemetry to retrieve (1-180)").
		WithFlag("from", "", "Start time for telemetry in RFC3339 format (e.g. 2024-01-01T00:00:00Z); use with --to").
		WithOptionalFlag("from", "Start time for telemetry (RFC3339); requires --to").
		WithFlag("to", "", "End time for telemetry in RFC3339 format; use with --from").
		WithOptionalFlag("to", "End time for telemetry (RFC3339); requires --from")
	return b
}
