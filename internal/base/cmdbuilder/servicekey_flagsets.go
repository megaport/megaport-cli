package cmdbuilder

// WithServiceKeyCommonFlags adds common flags for service key operations
func (b *CommandBuilder) WithServiceKeyCommonFlags() *CommandBuilder {
	b.WithFlag("product-uid", "", "Product UID for the service key")
	b.WithIntFlag("product-id", 0, "Product ID for the service key")
	b.WithBoolFlag("single-use", false, "Single-use service key")
	b.WithFlag("description", "", "Description for the service key")
	return b
}

// WithServiceKeyCreateFlags adds all flags needed for service key creation
func (b *CommandBuilder) WithServiceKeyCreateFlags() *CommandBuilder {
	b.WithServiceKeyCommonFlags()
	b.WithDateRangeFlags()
	b.WithIntFlag("max-speed", 0, "Maximum speed for the service key")
	return b
}

// WithServiceKeyUpdateFlags adds flags for updating a service key
func (b *CommandBuilder) WithServiceKeyUpdateFlags() *CommandBuilder {
	b.WithServiceKeyCommonFlags()
	b.WithBoolFlag("active", false, "Activate the service key")
	return b
}

// WithResourceIdentificationFlags adds UID and name flags for resource identification
func (b *CommandBuilder) WithResourceIdentificationFlags() *CommandBuilder {
	b.WithFlag("uid", "", "Unique identifier of the resource")
	b.WithFlag("name", "", "Name of the resource")
	return b
}

// WithTimeRangeFilterFlags adds flags for filtering resources by time range
func (b *CommandBuilder) WithTimeRangeFilterFlags() *CommandBuilder {
	b.WithFlag("from", "", "Start time for filtering (YYYY-MM-DD or YYYY-MM-DDTHH:MM:SS)")
	b.WithFlag("to", "", "End time for filtering (YYYY-MM-DD or YYYY-MM-DDTHH:MM:SS)")
	return b
}

// WithStatusFilterFlags adds flags for filtering resources by status
func (b *CommandBuilder) WithStatusFilterFlags() *CommandBuilder {
	b.WithFlag("status", "", "Filter by resource status")
	b.WithBoolFlag("include-deleted", false, "Include deleted resources")
	return b
}
