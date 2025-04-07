package cmdbuilder

import (
	"fmt"
	"os"
)

// WithMVECommonFlags adds common flags for MVE operations
func (b *CommandBuilder) WithMVECommonFlags() *CommandBuilder {
	b.WithFlag("name", "", "MVE name")
	b.WithFlag("cost-centre", "", "Cost centre for billing")
	return b
}

// WithMVECreateFlags adds flags needed for MVE creation
func (b *CommandBuilder) WithMVECreateFlags() *CommandBuilder {
	// Required flags
	b.WithFlag("name", "", "The name of the MVE")
	b.WithFlag("term", "0", "The term of the MVE (1, 12, 24, or 36 months)")
	b.WithFlag("location-id", "0", "The ID of the location where the MVE will be provisioned")
	b.WithFlag("vendor-config", "", "JSON string with vendor-specific configuration (for flag mode)")
	b.WithFlag("vnics", "", "JSON array of network interfaces (for flag mode)")

	// Mark these flags as required and handle potential errors
	requiredFlags := []string{"name", "term", "location-id", "vendor-config", "vnics"}
	for _, flag := range requiredFlags {
		if err := b.cmd.MarkFlagRequired(flag); err != nil {
			// Log the error but continue - this is a development-time error
			fmt.Fprintf(os.Stderr, "Warning: Failed to mark flag '%s' as required: %v\n", flag, err)
		}
	}

	// Optional flags
	b.WithFlag("diversity-zone", "", "The diversity zone for the MVE")
	b.WithFlag("promo-code", "", "Promotional code for discounts")
	b.WithFlag("cost-centre", "", "Cost centre for billing")
	return b
}

// WithMVEUpdateFlags adds flags for updating an MVE
func (b *CommandBuilder) WithMVEUpdateFlags() *CommandBuilder {
	// All update flags are optional
	b.WithFlag("name", "", "The new name of the MVE (1-64 characters)")
	b.WithFlag("cost-centre", "", "The new cost centre for billing purposes")
	b.WithFlag("contract-term", "0", "New contract term in months (1, 12, 24, or 36)")
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
