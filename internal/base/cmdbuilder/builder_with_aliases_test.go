package cmdbuilder

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestWithAliases(t *testing.T) {
	tests := []struct {
		name     string
		aliases  []string
		expected []string
	}{
		{
			name:     "single alias",
			aliases:  []string{"ls"},
			expected: []string{"ls"},
		},
		{
			name:     "multiple aliases",
			aliases:  []string{"ls", "list-items"},
			expected: []string{"ls", "list-items"},
		},
		{
			name:     "empty aliases",
			aliases:  []string{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewCommand("test", "Test command")
			result := builder.WithAliases(tt.aliases)

			// Verify the builder returns itself for chaining
			assert.NotNil(t, result)
			assert.Equal(t, builder, result)

			// Verify aliases are set on the cobra command
			assert.Equal(t, tt.expected, builder.cmd.Aliases)
		})
	}
}

func TestWithAliasesChaining(t *testing.T) {
	// Test that WithAliases works properly in a chain
	cmd := NewCommand("list", "List resources").
		WithAliases([]string{"ls"}).
		WithLongDesc("List all resources").
		Build()

	assert.Equal(t, []string{"ls"}, cmd.Aliases)
	assert.Equal(t, "list", cmd.Use)
	assert.Equal(t, "List resources", cmd.Short)
}

func TestWithAliasesPreservesOtherProperties(t *testing.T) {
	// Test that adding aliases doesn't affect other command properties
	builder := NewCommand("delete", "Delete a resource").
		WithLongDesc("Delete a resource from the system").
		WithExample("delete resource-123").
		WithAliases([]string{"rm"})

	cmd := builder.Build()

	assert.Equal(t, []string{"rm"}, cmd.Aliases)
	assert.Equal(t, "delete", cmd.Use)
	assert.Equal(t, "Delete a resource", cmd.Short)
	assert.Contains(t, cmd.Long, "Delete a resource from the system")
}

func TestMultipleAliasesForDifferentCommands(t *testing.T) {
	// Test that different commands can have different aliases
	listCmd := NewCommand("list", "List items").
		WithAliases([]string{"ls"}).
		Build()

	getCmd := NewCommand("get", "Get item details").
		WithAliases([]string{"show"}).
		Build()

	deleteCmd := NewCommand("delete", "Delete item").
		WithAliases([]string{"rm"}).
		Build()

	assert.Equal(t, []string{"ls"}, listCmd.Aliases)
	assert.Equal(t, []string{"show"}, getCmd.Aliases)
	assert.Equal(t, []string{"rm"}, deleteCmd.Aliases)
}

func TestAliasesIntegrationWithCommandTree(t *testing.T) {
	// Test that aliases work correctly when commands are added to parent
	rootCmd := &cobra.Command{
		Use:   "test",
		Short: "Test root",
	}

	listCmd := NewCommand("list", "List items").
		WithAliases([]string{"ls"}).
		WithRootCmd(rootCmd).
		Build()

	rootCmd.AddCommand(listCmd)

	// Verify the command can be found by its name
	foundCmd, _, err := rootCmd.Find([]string{"list"})
	assert.NoError(t, err)
	assert.Equal(t, "list", foundCmd.Use)

	// Verify the command can be found by its alias
	foundCmd, _, err = rootCmd.Find([]string{"ls"})
	assert.NoError(t, err)
	assert.Equal(t, "list", foundCmd.Use)
}

func TestBuilderFluentInterface(t *testing.T) {
	// Test the complete fluent interface including WithAliases
	cmd := NewCommand("status", "Show status").
		WithLongDesc("Display the status of the system").
		WithExample("status").
		WithExample("status --verbose").
		WithAliases([]string{"st"}).
		WithBoolFlag("verbose", false, "Verbose output").
		Build()

	assert.Equal(t, "status", cmd.Use)
	assert.Equal(t, "Show status", cmd.Short)
	assert.Equal(t, []string{"st"}, cmd.Aliases)
	assert.NotNil(t, cmd.Flags().Lookup("verbose"))
}
