package megaport

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestResetAllFlags_ResetsChangedState verifies that flags marked as "Changed"
// by a previous Cobra execution are properly cleared before the next run.
func TestResetAllFlags_ResetsChangedState(t *testing.T) {
	root := &cobra.Command{Use: "root"}
	child := &cobra.Command{Use: "child", RunE: func(cmd *cobra.Command, args []string) error { return nil }}
	child.Flags().String("country", "", "filter by country")
	child.Flags().String("metro", "", "filter by metro")
	root.AddCommand(child)

	// Simulate first execution with --country flag
	root.SetArgs([]string{"child", "--country", "Australia"})
	err := root.Execute()
	require.NoError(t, err)

	// Verify flags are marked as changed after execution
	countryFlag := child.Flags().Lookup("country")
	require.NotNil(t, countryFlag)
	assert.True(t, countryFlag.Changed, "country flag should be changed after first execution")
	assert.Equal(t, "Australia", countryFlag.Value.String())

	metroFlag := child.Flags().Lookup("metro")
	require.NotNil(t, metroFlag)
	assert.False(t, metroFlag.Changed, "metro flag should not be changed")

	// Reset all flags
	resetAllFlags(root)

	// Verify flags are reset
	assert.False(t, countryFlag.Changed, "country flag should not be changed after reset")
	assert.Equal(t, "", countryFlag.Value.String(), "country flag value should be reset to default")
	assert.False(t, metroFlag.Changed, "metro flag should remain unchanged")
}

// TestResetAllFlags_ResetsNestedCommands verifies that flag reset works
// recursively through a multi-level command tree.
func TestResetAllFlags_ResetsNestedCommands(t *testing.T) {
	root := &cobra.Command{Use: "root"}
	root.PersistentFlags().String("output", "table", "output format")

	level1 := &cobra.Command{Use: "locations"}
	level2 := &cobra.Command{Use: "list", RunE: func(cmd *cobra.Command, args []string) error { return nil }}
	level2.Flags().String("country", "", "filter by country")
	level2.Flags().Bool("verbose", false, "verbose output")

	level1.AddCommand(level2)
	root.AddCommand(level1)

	// Execute with flags at multiple levels
	root.SetArgs([]string{"locations", "list", "--country", "Australia", "--output", "json", "--verbose"})
	err := root.Execute()
	require.NoError(t, err)

	// Verify all flags are set
	assert.True(t, root.PersistentFlags().Lookup("output").Changed)
	assert.Equal(t, "json", root.PersistentFlags().Lookup("output").Value.String())
	assert.True(t, level2.Flags().Lookup("country").Changed)
	assert.Equal(t, "Australia", level2.Flags().Lookup("country").Value.String())
	assert.True(t, level2.Flags().Lookup("verbose").Changed)

	// Reset
	resetAllFlags(root)

	// Verify all flags at all levels are reset
	outputFlag := root.PersistentFlags().Lookup("output")
	assert.False(t, outputFlag.Changed, "persistent output flag should be reset")
	assert.Equal(t, "table", outputFlag.Value.String(), "output should revert to default")

	countryFlag := level2.Flags().Lookup("country")
	assert.False(t, countryFlag.Changed, "nested country flag should be reset")
	assert.Equal(t, "", countryFlag.Value.String())

	verboseFlag := level2.Flags().Lookup("verbose")
	assert.False(t, verboseFlag.Changed, "nested verbose flag should be reset")
	assert.Equal(t, "false", verboseFlag.Value.String())
}

// TestResetAllFlags_SimulatesConsecutiveCommands reproduces the exact user bug:
// running "locations list --country Australia" then "locations list --metro Ashburn"
// should NOT carry over the country filter.
func TestResetAllFlags_SimulatesConsecutiveCommands(t *testing.T) {
	var capturedCountry, capturedMetro string
	var countryChanged, metroChanged bool

	root := &cobra.Command{Use: "megaport-cli"}
	locations := &cobra.Command{Use: "locations"}
	list := &cobra.Command{
		Use: "list",
		RunE: func(cmd *cobra.Command, args []string) error {
			capturedCountry, _ = cmd.Flags().GetString("country")
			capturedMetro, _ = cmd.Flags().GetString("metro")
			countryChanged = cmd.Flags().Lookup("country").Changed
			metroChanged = cmd.Flags().Lookup("metro").Changed
			return nil
		},
	}
	list.Flags().String("country", "", "filter by country")
	list.Flags().String("metro", "", "filter by metro")
	locations.AddCommand(list)
	root.AddCommand(locations)

	// First command: locations list --country "Australia"
	root.SetArgs([]string{"locations", "list", "--country", "Australia"})
	err := root.Execute()
	require.NoError(t, err)
	assert.Equal(t, "Australia", capturedCountry)
	assert.True(t, countryChanged)

	// Reset flags (this is what our fix does)
	resetAllFlags(root)

	// Second command: locations list --metro "Ashburn" (no --country!)
	root.SetArgs([]string{"locations", "list", "--metro", "Ashburn"})
	err = root.Execute()
	require.NoError(t, err)

	assert.Equal(t, "Ashburn", capturedMetro, "metro should be set")
	assert.True(t, metroChanged, "metro should be marked as changed")
	assert.Equal(t, "", capturedCountry, "country should be empty (not carried over)")
	assert.False(t, countryChanged, "country should NOT be marked as changed")
}

// TestResetAllFlags_WithoutReset_BugReproduction demonstrates the bug without
// the fix — proving that Cobra does persist flag state across executions.
func TestResetAllFlags_WithoutReset_BugReproduction(t *testing.T) {
	var capturedCountry string
	var countryChanged bool

	root := &cobra.Command{Use: "root"}
	list := &cobra.Command{
		Use: "list",
		RunE: func(cmd *cobra.Command, args []string) error {
			capturedCountry, _ = cmd.Flags().GetString("country")
			countryChanged = cmd.Flags().Lookup("country").Changed
			return nil
		},
	}
	list.Flags().String("country", "", "filter by country")
	root.AddCommand(list)

	// First execution with --country
	root.SetArgs([]string{"list", "--country", "Australia"})
	err := root.Execute()
	require.NoError(t, err)
	assert.Equal(t, "Australia", capturedCountry)
	assert.True(t, countryChanged)

	// Second execution WITHOUT --country and WITHOUT resetAllFlags
	root.SetArgs([]string{"list"})
	err = root.Execute()
	require.NoError(t, err)

	// BUG: Without reset, the flag value and Changed state persist!
	assert.Equal(t, "Australia", capturedCountry, "without reset, country value persists (this is the bug)")
	assert.True(t, countryChanged, "without reset, Changed state persists (this is the bug)")
}

// TestEnableTraversalForAllCommands verifies that traversal is enabled on
// the entire command tree recursively.
func TestEnableTraversalForAllCommands(t *testing.T) {
	root := &cobra.Command{Use: "root"}
	level1 := &cobra.Command{Use: "child1"}
	level2 := &cobra.Command{Use: "grandchild"}
	level1.AddCommand(level2)
	root.AddCommand(level1)

	// Initially traversal is false
	assert.False(t, root.TraverseChildren)
	assert.False(t, level1.TraverseChildren)
	assert.False(t, level2.TraverseChildren)

	enableTraversalForAllCommands(root)

	assert.True(t, root.TraverseChildren, "root should have traversal enabled")
	assert.True(t, level1.TraverseChildren, "child should have traversal enabled")
	assert.True(t, level2.TraverseChildren, "grandchild should have traversal enabled")
}
