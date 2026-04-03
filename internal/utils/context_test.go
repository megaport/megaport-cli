package utils

import (
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContextFromCmd(t *testing.T) {
	t.Run("uses default timeout when command is nil", func(t *testing.T) {
		ctx, cancel := ContextFromCmd(nil)
		defer cancel()

		deadline, ok := ctx.Deadline()
		require.True(t, ok)

		remaining := time.Until(deadline)
		assert.GreaterOrEqual(t, remaining, 89*time.Second)
		assert.LessOrEqual(t, remaining, 90*time.Second)
	})

	t.Run("uses default timeout when timeout flag is missing", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}

		ctx, cancel := ContextFromCmd(cmd)
		defer cancel()

		deadline, ok := ctx.Deadline()
		require.True(t, ok)

		remaining := time.Until(deadline)
		assert.GreaterOrEqual(t, remaining, 89*time.Second)
		assert.LessOrEqual(t, remaining, 90*time.Second)
	})

	t.Run("uses default timeout when timeout flag is zero", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().Duration("timeout", 0, "")
		require.NoError(t, cmd.Flags().Set("timeout", "0s"))

		ctx, cancel := ContextFromCmd(cmd)
		defer cancel()

		deadline, ok := ctx.Deadline()
		require.True(t, ok)

		remaining := time.Until(deadline)
		assert.GreaterOrEqual(t, remaining, 89*time.Second)
		assert.LessOrEqual(t, remaining, 90*time.Second)
	})

	t.Run("uses configured timeout from timeout flag", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().Duration("timeout", 0, "")
		require.NoError(t, cmd.Flags().Set("timeout", "2m"))

		ctx, cancel := ContextFromCmd(cmd)
		defer cancel()

		deadline, ok := ctx.Deadline()
		require.True(t, ok)

		remaining := time.Until(deadline)
		assert.GreaterOrEqual(t, remaining, 119*time.Second)
		assert.LessOrEqual(t, remaining, 120*time.Second)
	})

	t.Run("uses default timeout when timeout flag exists with wrong type", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("timeout", "", "")
		require.NoError(t, cmd.Flags().Set("timeout", "2m"))

		ctx, cancel := ContextFromCmd(cmd)
		defer cancel()

		deadline, ok := ctx.Deadline()
		require.True(t, ok)

		remaining := time.Until(deadline)
		assert.GreaterOrEqual(t, remaining, 89*time.Second)
		assert.LessOrEqual(t, remaining, 90*time.Second)
	})
}
