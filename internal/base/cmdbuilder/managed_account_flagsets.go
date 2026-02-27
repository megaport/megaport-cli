package cmdbuilder

// WithManagedAccountCreateFlags adds flags for managed account creation
func (b *CommandBuilder) WithManagedAccountCreateFlags() *CommandBuilder {
	b.WithFlag("account-name", "", "The name of the managed account").
		WithFlag("account-ref", "", "The reference ID for the managed account")
	return b
}

// WithManagedAccountUpdateFlags adds flags for managed account updates
func (b *CommandBuilder) WithManagedAccountUpdateFlags() *CommandBuilder {
	b.WithFlag("account-name", "", "The new name of the managed account").
		WithFlag("account-ref", "", "The new reference ID for the managed account").
		WithOptionalFlag("account-name", "The new name of the managed account").
		WithOptionalFlag("account-ref", "The new reference ID for the managed account")
	return b
}

// WithManagedAccountFilterFlags adds flags for filtering managed account lists
func (b *CommandBuilder) WithManagedAccountFilterFlags() *CommandBuilder {
	b.WithFlag("account-name", "", "Filter managed accounts by name (partial match)").
		WithFlag("account-ref", "", "Filter managed accounts by reference (partial match)")
	return b
}
