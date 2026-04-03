package utils

import (
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertDeadlineWithin(t *testing.T, ctxDeadline time.Time, start time.Time, expected time.Duration) {
	t.Helper()
	assert.WithinDuration(t, start.Add(expected), ctxDeadline, 500*time.Millisecond)
}

func TestContextFromCmd(t *testing.T) {
	t.Run("uses default timeout when command is nil", func(t *testing.T) {
		start := time.Now()
		ctx, cancel := ContextFromCmd(nil)
		defer cancel()

		deadline, ok := ctx.Deadline()
		require.True(t, ok)
		assertDeadlineWithin(t, deadline, start, 90*time.Second)
	})

	t.Run("uses default timeout when timeout flag is missing", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}

		start := time.Now()
		ctx, cancel := ContextFromCmd(cmd)
		defer cancel()

		deadline, ok := ctx.Deadline()
		require.True(t, ok)
		assertDeadlineWithin(t, deadline, start, 90*time.Second)
	})

	t.Run("uses default timeout when timeout flag is zero", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().Duration("timeout", 0, "")
		require.NoError(t, cmd.Flags().Set("timeout", "0s"))

		start := time.Now()
		ctx, cancel := ContextFromCmd(cmd)
		defer cancel()

		deadline, ok := ctx.Deadline()
		require.True(t, ok)
		assertDeadlineWithin(t, deadline, start, 90*time.Second)
	})

	t.Run("uses configured timeout from timeout flag", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().Duration("timeout", 0, "")
		require.NoError(t, cmd.Flags().Set("timeout", "2m"))

		start := time.Now()
		ctx, cancel := ContextFromCmd(cmd)
		defer cancel()

		deadline, ok := ctx.Deadline()
		require.True(t, ok)
		assertDeadlineWithin(t, deadline, start, 2*time.Minute)
	})

	t.Run("uses configured timeout from inherited persistent timeout flag", func(t *testing.T) {
		root := &cobra.Command{Use: "root"}
		child := &cobra.Command{Use: "child"}
		root.PersistentFlags().Duration("timeout", 0, "")
		root.AddCommand(child)
		require.NoError(t, child.ParseFlags([]string{"--timeout", "2m"}))

		start := time.Now()
		ctx, cancel := ContextFromCmd(child)
		defer cancel()

		deadline, ok := ctx.Deadline()
		require.True(t, ok)
		assertDeadlineWithin(t, deadline, start, 2*time.Minute)
	})

	t.Run("uses default timeout when timeout flag is negative", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().Duration("timeout", 0, "")
		require.NoError(t, cmd.Flags().Set("timeout", "-5s"))

		start := time.Now()
		ctx, cancel := ContextFromCmd(cmd)
		defer cancel()

		deadline, ok := ctx.Deadline()
		require.True(t, ok)
		assertDeadlineWithin(t, deadline, start, 90*time.Second)
	})

	t.Run("uses default timeout when timeout flag exists with wrong type", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("timeout", "", "")
		require.NoError(t, cmd.Flags().Set("timeout", "2m"))

		start := time.Now()
		ctx, cancel := ContextFromCmd(cmd)
		defer cancel()

		deadline, ok := ctx.Deadline()
		require.True(t, ok)
		assertDeadlineWithin(t, deadline, start, 90*time.Second)
	})
}
