package cmdbuilder

// AddStandardInputFlags adds interactive, json and json-file flags
func (b *CommandBuilder) WithStandardInputFlags() *CommandBuilder {
	b.WithBoolFlagP("interactive", "i", false, "Use interactive mode with prompts")
	b.WithFlag("json", "", "JSON string containing configuration")
	b.WithFlag("json-file", "", "Path to JSON file containing configuration")
	return b
}

// WithDateRangeFlags adds start and end date flags for time-bound resources
func (b *CommandBuilder) WithDateRangeFlags() *CommandBuilder {
	b.WithFlag("start-date", "", "Start date (YYYY-MM-DD)")
	b.WithFlag("end-date", "", "End date (YYYY-MM-DD)")
	return b
}

// WithJSONConfigFlags adds flags for JSON configuration input
func (b *CommandBuilder) WithJSONConfigFlags() *CommandBuilder {
	b.WithFlag("json", "", "JSON string containing configuration")
	b.WithFlag("json-file", "", "Path to JSON file containing configuration")
	return b
}

// WithMCRDeleteFlags adds flags for deletion
func (b *CommandBuilder) WithDeleteFlags() *CommandBuilder {
	b.WithBoolFlagP("force", "f", false, "Skip confirmation prompt")
	b.WithBoolFlag("now", false, "Delete resource immediately instead of at end of billing cycle")
	return b
}

// WithInteractiveFlag adds just the interactive flag
func (b *CommandBuilder) WithInteractiveFlag() *CommandBuilder {
	b.WithBoolFlagP("interactive", "i", false, "Use interactive mode with prompts")
	return b
}
