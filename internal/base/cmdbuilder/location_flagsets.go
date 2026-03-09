package cmdbuilder

// WithLocationsFilterFlags adds common location filter flags
func (b *CommandBuilder) WithLocationsFilterFlags() *CommandBuilder {
	b.WithFlag("metro", "", "Filter locations by metro area")
	b.WithOptionalFlag("metro", "Filter locations by metro area")
	b.WithFlag("country", "", "Filter locations by country")
	b.WithOptionalFlag("country", "Filter locations by country")
	b.WithFlag("name", "", "Filter locations by name")
	b.WithOptionalFlag("name", "Filter locations by name")
	b.WithFlag("market-code", "", "Filter locations by market code")
	b.WithOptionalFlag("market-code", "Filter locations by market code")
	b.WithBoolFlag("mcr-available", false, "Filter locations that support MCR")
	b.WithOptionalFlag("mcr-available", "Filter locations that support MCR")
	return b
}
