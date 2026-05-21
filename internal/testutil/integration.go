//go:build integration

package testutil

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/megaport/megaport-cli/internal/commands/config"
	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/require"
)

// SetupIntegrationClient reads staging credentials from environment variables,
// authorises against the staging API, and returns a ready-to-use *megaport.Client.
// Skips the test if MEGAPORT_ACCESS_KEY or MEGAPORT_SECRET_KEY are not set.
//
// Suitable for serial read-only tests. For tests that use t.Parallel(), prefer
// RequireSharedIntegrationClient which installs the login function exactly
// once per process via sync.Once and avoids the save/restore race in
// LoginWithClient.
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
//
// Not safe under t.Parallel(): concurrent tests can capture each other's
// installed function as their "original" and leave the global in an
// unexpected state on cleanup. Use RequireSharedIntegrationClient instead.
func LoginWithClient(t *testing.T, client *megaport.Client) func() {
	t.Helper()
	original := config.GetLoginFunc()
	config.SetLoginFunc(func(ctx context.Context) (*megaport.Client, error) {
		return client, nil
	})
	return func() { config.SetLoginFunc(original) }
}

var (
	sharedIntegrationClient     *megaport.Client
	sharedIntegrationClientOnce sync.Once
	sharedIntegrationClientErr  error
)

// RequireSharedIntegrationClient installs a process-wide staging client
// suitable for parallel integration tests. Unlike SetupIntegrationClient +
// LoginWithClient, it authorises against staging and installs
// config.SetLoginFunc exactly once via sync.Once and never restores. This
// avoids a race in which two t.Parallel() tests concurrently capture and
// restore config.LoginFunc, leaving the global pointing at a stale closure.
//
// Sharing one authorised client across parallel tests is safe because they
// all target the same staging environment. Subsequent callers reuse the
// cached client. Skips when MEGAPORT_ACCESS_KEY or MEGAPORT_SECRET_KEY are
// not set.
func RequireSharedIntegrationClient(t *testing.T) {
	t.Helper()
	sharedIntegrationClientOnce.Do(func() {
		accessKey := os.Getenv("MEGAPORT_ACCESS_KEY")
		secretKey := os.Getenv("MEGAPORT_SECRET_KEY")
		if accessKey == "" || secretKey == "" {
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()

		client, err := megaport.New(nil,
			megaport.WithCredentials(accessKey, secretKey),
			megaport.WithEnvironment(megaport.EnvironmentStaging),
		)
		if err != nil {
			sharedIntegrationClientErr = fmt.Errorf("failed to create megaport client: %w", err)
			return
		}
		if _, err := client.Authorize(ctx); err != nil {
			sharedIntegrationClientErr = fmt.Errorf("failed to authorize against staging API: %w", err)
			return
		}
		sharedIntegrationClient = client
		config.SetLoginFunc(func(context.Context) (*megaport.Client, error) {
			return client, nil
		})
	})
	if sharedIntegrationClientErr != nil {
		t.Fatalf("integration setup failed: %v", sharedIntegrationClientErr)
	}
	if sharedIntegrationClient == nil {
		t.Skip("MEGAPORT_ACCESS_KEY and MEGAPORT_SECRET_KEY required for integration tests")
	}
}
