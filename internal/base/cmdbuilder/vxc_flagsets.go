package cmdbuilder

import (
	"fmt"

	"github.com/megaport/megaport-cli/internal/validation"
)

// WithVXCCommonFlags adds common flags for VXC operations
func (b *CommandBuilder) WithVXCCommonFlags() *CommandBuilder {
	b.WithFlag("name", "", "Name of the VXC")
	b.WithIntFlag("rate-limit", 0, "Bandwidth in Mbps")
	b.WithIntFlag("term", 0, fmt.Sprintf("Contract term in months (%s)", validation.FormatIntSlice(validation.ValidContractTerms)))
	b.WithFlag("cost-centre", "", "Cost centre for billing")
	return b
}

// WithVXCEndpointFlags adds endpoint-related flags for VXC operations
func (b *CommandBuilder) WithVXCEndpointFlags() *CommandBuilder {
	b.WithFlag("a-end-uid", "", "UID of the A-End product")
	b.WithFlag("b-end-uid", "", "UID of the B-End product")
	b.WithIntFlag("a-end-vlan", 0, "VLAN for A-End ("+validation.VLANHelpText()+")")
	b.WithIntFlag("b-end-vlan", 0, "VLAN for B-End ("+validation.VLANHelpText()+")")
	b.WithIntFlag("a-end-inner-vlan", 0, "Inner VLAN for A-End ("+validation.InnerVLANHelpText()+")")
	b.WithIntFlag("b-end-inner-vlan", 0, "Inner VLAN for B-End ("+validation.InnerVLANHelpText()+")")
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
	b.WithResourceTagFlags()
	return b
}

// WithVXCUpdateFlags adds all flags needed for VXC updates
func (b *CommandBuilder) WithVXCUpdateFlags() *CommandBuilder {
	b.WithVXCCommonFlags()
	b.WithVXCEndpointFlags()
	b.WithVXCPartnerConfigFlags()
	b.WithBoolFlag("shutdown", false, "Whether to shut down the VXC")
	b.WithBoolFlag("is-approved", false, "Approve or reject a VXC via the Megaport Marketplace")
	b.WithIntFlag("a-vnic-index", -1, "New A-End vNIC index when moving a VXC on an MVE")
	b.WithIntFlag("b-vnic-index", -1, "New B-End vNIC index when moving a VXC on an MVE")
	return b
}

// WithVXCFilterFlags adds flags for filtering VXC lists
func (b *CommandBuilder) WithVXCFilterFlags() *CommandBuilder {
	b.WithFlag("name", "", "Filter VXCs by name (case-sensitive partial match)")
	b.WithFlag("name-contains", "", "Filter VXCs by name (case-sensitive partial match; takes precedence over --name)")
	b.WithIntFlag("rate-limit", 0, "Filter VXCs by rate limit in Mbps")
	b.WithFlag("a-end-uid", "", "Filter VXCs by A-End product UID")
	b.WithFlag("b-end-uid", "", "Filter VXCs by B-End product UID")
	b.WithFlag("status", "", "Filter VXCs by status (comma-separated, e.g. LIVE,CONFIGURED)")
	b.WithBoolFlag("include-inactive", false, "Include inactive VXCs in the list")
	return b
}
