package utils

import (
	"context"
	"fmt"
	"testing"
	"time"

	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newLoginTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Duration("timeout", 0, "timeout")
	return cmd
}

func TestLoginClient_Success(t *testing.T) {
	mockClient := &megaport.Client{}
	login := func(ctx context.Context) (*megaport.Client, error) {
		return mockClient, nil
	}

	cmd := newLoginTestCmd()
	ctx, cancel, client, err := LoginClient(cmd, 90*time.Second, login)
	require.NoError(t, err)
	defer cancel()

	assert.NotNil(t, ctx)
	assert.NotNil(t, cancel)
	assert.Equal(t, mockClient, client)
}

func TestLoginClient_LoginError(t *testing.T) {
	login := func(ctx context.Context) (*megaport.Client, error) {
		return nil, fmt.Errorf("auth failed")
	}

	cmd := newLoginTestCmd()
	ctx, cancel, client, err := LoginClient(cmd, 90*time.Second, login)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error logging in")
	assert.Contains(t, err.Error(), "auth failed")
	assert.Nil(t, ctx)
	assert.Nil(t, cancel)
	assert.Nil(t, client)
}

func TestLoginClient_UsesDefaultTimeout(t *testing.T) {
	var capturedCtx context.Context
	login := func(ctx context.Context) (*megaport.Client, error) {
		capturedCtx = ctx
		return &megaport.Client{}, nil
	}

	cmd := newLoginTestCmd()
	_, cancel, _, err := LoginClient(cmd, 5*time.Second, login)
	require.NoError(t, err)
	defer cancel()

	deadline, ok := capturedCtx.Deadline()
	assert.True(t, ok)
	assert.WithinDuration(t, time.Now().Add(5*time.Second), deadline, 1*time.Second)
}
