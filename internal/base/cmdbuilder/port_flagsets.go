package cmdbuilder

import (
	"fmt"

	"github.com/megaport/megaport-cli/internal/validation"
)

// WithPortCommonFlags adds common flags for port operations
func (b *CommandBuilder) WithPortCommonFlags() *CommandBuilder {
	b.WithFlag("name", "", "Port name")
	b.WithIntFlag("term", 0, fmt.Sprintf("Contract term in months (%s)", validation.FormatIntSlice(validation.ValidContractTerms)))
	b.WithBoolFlag("marketplace-visibility", false, "Whether the port is visible in marketplace")
	b.WithFlag("diversity-zone", "", "Diversity zone for the port")
	b.WithFlag("cost-centre", "", "Cost centre for billing")
	return b
}

// WithPortCreationFlags adds flags needed for port creation without marking them required
func (b *CommandBuilder) WithPortCreationFlags() *CommandBuilder {
	// Add all flags but don't mark them as required - we'll use conditional validation
	b.WithPortCommonFlags()
	b.WithIntFlag("port-speed", 0, fmt.Sprintf("Port speed in Mbps (%s)", validation.FormatIntSlice(validation.ValidPortSpeeds)))
	b.WithIntFlag("location-id", 0, "Location ID where the port will be provisioned")
	b.WithFlag("promo-code", "", "Promotional code for discounts")
	b.WithResourceTagFlags()
	return b
}

// WithPortLAGFlags adds flags specific to LAG port operations
func (b *CommandBuilder) WithPortLAGFlags() *CommandBuilder {
	b.WithPortCreationFlags()
	b.WithIntFlag("lag-count", 0, fmt.Sprintf("Number of LAGs (%d-%d)", validation.MinLAGCount, validation.MaxLAGCount))
	return b
}

// WithPortUpdateFlags adds flags needed for port updates
func (b *CommandBuilder) WithPortUpdateFlags() *CommandBuilder {
	b.WithFlag("name", "", "New port name")
	b.WithBoolFlag("marketplace-visibility", false, "Whether the port is visible in marketplace")
	b.WithFlag("cost-centre", "", "Cost centre for billing")
	b.WithIntFlag("term", 0, fmt.Sprintf("New contract term in months (%s)", validation.FormatIntSlice(validation.ValidContractTerms)))
	return b
}

// WithPortFilterFlags adds flags for filtering port lists
func (b *CommandBuilder) WithPortFilterFlags() *CommandBuilder {
	b.WithIntFlag("location-id", 0, "Filter ports by location ID")
	b.WithIntFlag("port-speed", 0, "Filter ports by port speed")
	b.WithFlag("port-name", "", "Filter ports by port name")
	b.WithBoolFlag("include-inactive", false, "Include inactive ports in the list")
	return b
}
