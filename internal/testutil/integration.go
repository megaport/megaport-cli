//go:build integration

package testutil

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/megaport/megaport-cli/internal/commands/config"
	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/require"
)

// IntegrationEnvironment resolves the target API environment from
// MEGAPORT_ENVIRONMENT, returning the SDK environment and its display name. It
// defaults to staging when the variable is empty or unrecognized, so a typo
// can never silently point the suite at production.
func IntegrationEnvironment() (megaport.Environment, string) {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("MEGAPORT_ENVIRONMENT"))) {
	case "production", "prod":
		return megaport.EnvironmentProduction, "production"
	case "development", "dev":
		return megaport.EnvironmentDevelopment, "development"
	default:
		return megaport.EnvironmentStaging, "staging"
	}
}

// RequireStagingForProvisioning skips the test unless the resolved environment
// is staging. Provisioning lifecycle tests use hardcoded staging location IDs
// and must never create real resources in production or development, so they
// opt out of the configurable environment that read-only tests support.
func RequireStagingForProvisioning(t *testing.T) {
	t.Helper()
	if _, name := IntegrationEnvironment(); name != "staging" {
		t.Skipf("provisioning lifecycle tests are staging-only (hardcoded location IDs); MEGAPORT_ENVIRONMENT=%q resolved to %s", os.Getenv("MEGAPORT_ENVIRONMENT"), name)
	}
}

// SetupIntegrationClient reads credentials from environment variables,
// authorises against the environment named by MEGAPORT_ENVIRONMENT (staging by
// default), and returns a ready-to-use *megaport.Client. Skips the test if
// MEGAPORT_ACCESS_KEY or MEGAPORT_SECRET_KEY are not set.
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

	env, envName := IntegrationEnvironment()

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	client, err := megaport.New(nil,
		megaport.WithCredentials(accessKey, secretKey),
		megaport.WithEnvironment(env),
	)
	require.NoError(t, err, "failed to create megaport client")

	_, err = client.Authorize(ctx)
	require.NoError(t, err, "failed to authorize against %s API", envName)

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

// RequireSharedIntegrationClient installs a process-wide client suitable for
// parallel integration tests, targeting the environment named by
// MEGAPORT_ENVIRONMENT (staging by default). Unlike SetupIntegrationClient +
// LoginWithClient, it authorises and installs config.SetLoginFunc exactly once
// via sync.Once and never restores. This avoids a race in which two
// t.Parallel() tests concurrently capture and restore config.LoginFunc,
// leaving the global pointing at a stale closure.
//
// Sharing one authorised client across parallel tests is safe because they
// all target the same environment. Subsequent callers reuse the cached
// client. Skips when MEGAPORT_ACCESS_KEY or MEGAPORT_SECRET_KEY are not set.
func RequireSharedIntegrationClient(t *testing.T) {
	t.Helper()
	sharedIntegrationClientOnce.Do(func() {
		accessKey := os.Getenv("MEGAPORT_ACCESS_KEY")
		secretKey := os.Getenv("MEGAPORT_SECRET_KEY")
		if accessKey == "" || secretKey == "" {
			return
		}

		env, envName := IntegrationEnvironment()

		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()

		client, err := megaport.New(nil,
			megaport.WithCredentials(accessKey, secretKey),
			megaport.WithEnvironment(env),
		)
		if err != nil {
			sharedIntegrationClientErr = fmt.Errorf("failed to create megaport client: %w", err)
			return
		}
		if _, err := client.Authorize(ctx); err != nil {
			sharedIntegrationClientErr = fmt.Errorf("failed to authorize against %s API: %w", envName, err)
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

// SharedIntegrationClient returns the process-wide client installed by
// RequireSharedIntegrationClient. Tests use it to read state directly from the
// SDK in assertions, avoiding output.CaptureOutput on hot paths where parallel
// goroutines would race on the global os.Stdout swap. Callers must invoke
// RequireSharedIntegrationClient(t) first; this fails the test if not.
func SharedIntegrationClient(t *testing.T) *megaport.Client {
	t.Helper()
	if sharedIntegrationClient == nil {
		t.Fatal("SharedIntegrationClient called before RequireSharedIntegrationClient")
	}
	return sharedIntegrationClient
}

// locationHasPortSpeed reports whether loc advertises port (Megaport) capacity
// at speedMbps in either diversity zone.
func locationHasPortSpeed(loc *megaport.LocationV3, speedMbps int) bool {
	for _, s := range loc.GetMegaportSpeeds() {
		if s == speedMbps {
			return true
		}
	}
	return false
}

// locationHasMCRSpeed reports whether loc advertises MCR capacity at speedMbps
// in either diversity zone.
func locationHasMCRSpeed(loc *megaport.LocationV3, speedMbps int) bool {
	for _, s := range loc.GetMCRSpeeds() {
		if s == speedMbps {
			return true
		}
	}
	return false
}

// findOrderableLocation returns the ID of an active staging location satisfying
// qualifies, preferring preferredID so a healthy canonical location keeps the
// suite pinned where it has always run, and otherwise falling back to the first
// other qualifying location. It skips the test when none qualifies: a location
// losing a capability should route the suite elsewhere, not fail the build.
// capability names the requirement for the log and skip messages.
//
// Unlike the MVE host-capacity probe, this only checks the speeds the location
// advertises (the same signal the terraform provider's location helpers use);
// it never places a validate order, so it is cheap and side-effect free.
func findOrderableLocation(t *testing.T, client *megaport.Client, preferredID int, capability string, qualifies func(*megaport.LocationV3) bool) int {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	locations, err := client.LocationService.ListLocationsV3(ctx)
	require.NoErrorf(t, err, "list locations to find %s", capability)

	var fallback *megaport.LocationV3
	for _, loc := range locations {
		if loc == nil || !loc.IsStatusOrderable() || !qualifies(loc) {
			continue
		}
		if loc.ID == preferredID {
			return loc.ID
		}
		if fallback == nil {
			fallback = loc
		}
	}
	if fallback != nil {
		t.Logf("preferred location %d unavailable for %s; using location %d (%s)", preferredID, capability, fallback.ID, fallback.Name)
		return fallback.ID
	}
	t.Skipf("no active staging location advertises %s", capability)
	return 0
}

// FindPortTestLocation returns a staging location that advertises port capacity
// at speedMbps, preferring preferredID. See findOrderableLocation for the
// fallback and skip behavior.
func FindPortTestLocation(t *testing.T, client *megaport.Client, speedMbps, preferredID int) int {
	t.Helper()
	return findOrderableLocation(t, client, preferredID, fmt.Sprintf("%d Mbps port capacity", speedMbps), func(loc *megaport.LocationV3) bool {
		return locationHasPortSpeed(loc, speedMbps)
	})
}

// FindMCRTestLocation returns a staging location that advertises MCR capacity at
// speedMbps, preferring preferredID.
func FindMCRTestLocation(t *testing.T, client *megaport.Client, speedMbps, preferredID int) int {
	t.Helper()
	return findOrderableLocation(t, client, preferredID, fmt.Sprintf("%d Mbps MCR capacity", speedMbps), func(loc *megaport.LocationV3) bool {
		return locationHasMCRSpeed(loc, speedMbps)
	})
}

// FindPortAndMCRTestLocation returns a single staging location that advertises
// both port capacity at portSpeedMbps and MCR capacity at mcrSpeedMbps, for VXC
// tests that co-locate a port and an MCR. Prefers preferredID.
func FindPortAndMCRTestLocation(t *testing.T, client *megaport.Client, portSpeedMbps, mcrSpeedMbps, preferredID int) int {
	t.Helper()
	return findOrderableLocation(t, client, preferredID, fmt.Sprintf("%d Mbps port + %d Mbps MCR capacity", portSpeedMbps, mcrSpeedMbps), func(loc *megaport.LocationV3) bool {
		return locationHasPortSpeed(loc, portSpeedMbps) && locationHasMCRSpeed(loc, mcrSpeedMbps)
	})
}
