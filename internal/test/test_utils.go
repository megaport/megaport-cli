package integration

import (
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// Define a flag to enable integration tests
var runIntegration = flag.Bool("integration", false, "Run integration tests that make real API calls")

// TestMain handles the setup and flag parsing for all tests in this package
func TestMain(m *testing.M) {
	// Parse flags before running tests
	flag.Parse()
	os.Exit(m.Run())
}

// Generate a unique identifier for test resources
func generateUniqueID() string {
	buf := make([]byte, 4)
	if _, err := rand.Read(buf); err != nil {
		panic(fmt.Sprintf("Failed to generate random bytes: %v", err))
	}
	return hex.EncodeToString(buf)
}

// shouldRunIntegrationTests determines if integration tests should run
func shouldRunIntegrationTests(t *testing.T) bool {
	// Check if the integration flag was provided
	if !*runIntegration {
		t.Skip("Skipping integration tests. Use --integration flag to run")
		return false
	}

	// Check for required API credentials
	required := []string{"MEGAPORT_ACCESS_KEY", "MEGAPORT_SECRET_KEY"}
	for _, env := range required {
		if os.Getenv(env) == "" {
			t.Skipf("Skipping integration tests. Required environment variable %s not set", env)
			return false
		}
	}

	// Ensure we're using staging environment
	if os.Getenv("MEGAPORT_ENVIRONMENT") != "staging" {
		t.Skip("Integration tests require MEGAPORT_ENVIRONMENT=staging to prevent accidental production usage")
		return false
	}

	return true
}

// createTempJSONFile creates a temporary JSON file for testing
func createTempJSONFile(t *testing.T, content string) string {
	tmpfile, err := os.CreateTemp("", "megaport-test-*.json")
	require.NoError(t, err)

	_, err = tmpfile.WriteString(content)
	require.NoError(t, err)

	err = tmpfile.Close()
	require.NoError(t, err)

	return tmpfile.Name()
}
