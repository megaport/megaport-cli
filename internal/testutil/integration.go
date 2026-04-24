//go:build integration
// +build integration

package testutil

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/megaport/megaport-cli/internal/commands/config"
	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/require"
)

// SetupIntegrationClient reads staging credentials from environment variables,
// authorises against the staging API, and returns a ready-to-use *megaport.Client.
// Skips the test if MEGAPORT_ACCESS_KEY or MEGAPORT_SECRET_KEY are not set.
func SetupIntegrationClient(t *testing.T) *megaport.Client {
	t.Helper()
	accessKey := os.Getenv("MEGAPORT_ACCESS_KEY")
	secretKey := os.Getenv("MEGAPORT_SECRET_KEY")
	if accessKey == "" || secretKey == "" {
		t.Skip("MEGAPORT_ACCESS_KEY and MEGAPORT_SECRET_KEY required for integration tests")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	client, err := megaport.New(nil,
		megaport.WithCredentials(accessKey, secretKey),
		megaport.WithEnvironment(megaport.EnvironmentStaging),
	)
	require.NoError(t, err, "failed to create megaport client")

	_, err = client.Authorize(ctx)
	require.NoError(t, err, "failed to authorize against staging API")

	return client
}

// LoginWithClient overrides the login function to return the given client for the
// duration of the test. Returns a cleanup function that restores the original.
func LoginWithClient(t *testing.T, client *megaport.Client) func() {
	t.Helper()
	original := config.GetLoginFunc()
	config.SetLoginFunc(func(ctx context.Context) (*megaport.Client, error) {
		return client, nil
	})
	return func() { config.SetLoginFunc(original) }
}
