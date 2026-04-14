package cmdbuilder

// WithTagFilterFlags adds the --tag repeatable flag for filtering list results by resource tag.
// Format: --tag key=value (exact match) or --tag key (key-exists match).
// Multiple --tag flags are AND-ed together.
func (b *CommandBuilder) WithTagFilterFlags() *CommandBuilder {
	return b.WithStringArrayFlag("tag", "Filter by resource tag (format: key=value or key; repeatable, AND logic)")
}
