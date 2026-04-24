package testutil

import (
	"context"
	"fmt"
	"testing"

	"github.com/megaport/megaport-cli/internal/commands/config"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestSetupLogin(t *testing.T) {
	original := config.GetLoginFunc()
	defer config.SetLoginFunc(original)

	cleanup := SetupLogin(func(c *megaport.Client) {
		c.PortService = nil // just verify setupFn is called
	})

	client, err := config.GetLoginFunc()(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, client)

	cleanup()

	// After cleanup, the original function should be restored.
	restored := config.GetLoginFunc()
	assert.NotNil(t, restored)
}

func TestSetupLoginError(t *testing.T) {
	original := config.GetLoginFunc()
	defer config.SetLoginFunc(original)

	expectedErr := fmt.Errorf("mock login error")
	cleanup := SetupLoginError(expectedErr)

	client, err := config.GetLoginFunc()(context.Background())
	assert.Nil(t, client)
	assert.Equal(t, expectedErr, err)

	cleanup()
}

func TestNewCommand(t *testing.T) {
	cmd := NewCommand("test", func(cmd *cobra.Command, args []string) error {
		return nil
	})
	assert.Equal(t, "test", cmd.Use)

	val, err := cmd.Flags().GetString("output")
	assert.NoError(t, err)
	assert.Equal(t, "table", val)
}

func TestSetFlags(t *testing.T) {
	cmd := NewCommand("test", nil)
	cmd.Flags().String("name", "", "test flag")

	SetFlags(t, cmd, map[string]string{"name": "hello", "output": "json"})

	name, _ := cmd.Flags().GetString("name")
	assert.Equal(t, "hello", name)
	output, _ := cmd.Flags().GetString("output")
	assert.Equal(t, "json", output)
}
