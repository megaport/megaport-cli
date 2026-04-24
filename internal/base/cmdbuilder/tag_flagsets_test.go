package cmdbuilder

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithTagFilterFlags(t *testing.T) {
	cmd := NewCommand("list", "List resources").
		WithTagFilterFlags().
		Build()

	f := cmd.Flags().Lookup("tag")
	require.NotNil(t, f, "tag flag should be registered by WithTagFilterFlags")
	assert.Equal(t, "stringArray", f.Value.Type())
	assert.Contains(t, f.Usage, "key=value")
}
