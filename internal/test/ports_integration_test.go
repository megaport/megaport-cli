package integration

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPortLifecycle is an integration test that tests the full lifecycle of a port
// It requires actual API credentials to be set as environment variables
// MEGAPORT_ACCESS_KEY, MEGAPORT_SECRET_KEY, and MEGAPORT_ENVIRONMENT=staging
func TestPortLifecycle(t *testing.T) {
	// Skip if integration tests are not enabled or missing env vars
	if !shouldRunIntegrationTests(t) {
		return
	}

	// Enable parallel test execution
	t.Parallel()

	// Use unique identifier to prevent resource name conflicts
	uniqueID := generateUniqueID()
	portName := fmt.Sprintf("CLI-Test-Port-%s-%d", uniqueID, time.Now().Unix())
	var portUID string
	var output string

	// STEP 1: Create a port
	t.Run("Create Port", func(t *testing.T) {
		cmd := exec.Command("megaport-cli", "ports", "buy",
			"--name", portName,
			"--term", "1",
			"--port-speed", "1000",
			"--location-id", "67",
			"--marketplace-visibility", "false")

		// Capture both stdout and stderr
		outputBytes, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Command failed: %v", err)
		}
		output = string(outputBytes)
		t.Logf("Output: %s", output)

		// Check for successful execution
		assert.NoError(t, err, "Failed to create port: %s", output)
		assert.Contains(t, output, "Port created")

		// Extract port UID from the output
		portUID = extractPortUID(output)
		require.NotEmpty(t, portUID, "Failed to extract port UID from output")
		t.Logf("Created port with UID: %s", portUID)
	})

	// Wait for port to be ready
	t.Run("Verify Port Creation", func(t *testing.T) {
		// Skip if previous step failed
		if portUID == "" {
			t.Skip("Skipping because port creation failed")
		}

		maxAttempts := 20
		for attempt := 1; attempt <= maxAttempts; attempt++ {
			t.Logf("Checking port status, attempt %d of %d", attempt, maxAttempts)

			cmd := exec.Command("megaport-cli", "ports", "get", portUID)
			outputBytes, err := cmd.CombinedOutput()
			output = string(outputBytes)

			if err == nil && strings.Contains(output, "CONFIGURED") || strings.Contains(output, "LIVE") {
				t.Logf("Port is ready")
				break
			}

			if attempt == maxAttempts {
				t.Fatalf("Port did not reach CONFIGURED state within the expected time: %s", output)
			}

			time.Sleep(30 * time.Second)
		}
	})

	// STEP 2: Update the port
	t.Run("Update Port", func(t *testing.T) {
		// Skip if previous step failed
		if portUID == "" {
			t.Skip("Skipping because port creation failed")
		}

		newName := fmt.Sprintf("%s-Updated", portName)
		cmd := exec.Command("megaport-cli", "ports", "update", portUID,
			"--name", newName)

		outputBytes, err := cmd.CombinedOutput()
		output = string(outputBytes)
		t.Logf("Update output: %s", output)

		assert.NoError(t, err, "Failed to update port: %s", output)
		assert.Contains(t, output, "Port updated")

		// Verify the update
		cmd = exec.Command("megaport-cli", "ports", "get", portUID)
		outputBytes, err = cmd.CombinedOutput()
		output = string(outputBytes)

		assert.NoError(t, err, "Failed to get updated port info: %s", output)
		assert.Contains(t, output, newName)
	})

	// STEP 3: Delete the port
	t.Run("Delete Port", func(t *testing.T) {
		// Skip if previous step failed
		if portUID == "" {
			t.Skip("Skipping because port creation failed")
		}

		cmd := exec.Command("megaport-cli", "ports", "delete", portUID, "--now", "--force")

		outputBytes, err := cmd.CombinedOutput()
		output = string(outputBytes)
		t.Logf("Delete output: %s", output)

		assert.NoError(t, err, "Failed to delete port: %s", output)
		assert.Contains(t, output, "Deleting port")

		// Wait for deletion to start
		time.Sleep(5 * time.Second)

		// Verify the deletion status
		cmd = exec.Command("megaport-cli", "ports", "get", portUID)
		outputBytes, err = cmd.CombinedOutput()
		output = string(outputBytes)

		// The port should still exist but be in DECOMMISSIONED state
		assert.NoError(t, err)
		// Accept either DECOMMISSIONING or DECOMMISSIONED status
		assert.True(t, strings.Contains(output, "DECOMMISSIONING") || strings.Contains(output, "DECOMMISSIONED"),
			"Expected DECOMMISSIONING or DECOMMISSIONED in output: %s", output)
	})
}

// TestLAGPortLifecycle is an integration test that tests the full lifecycle of a LAG port
// It requires actual API credentials to be set as environment variables
func TestLAGPortLifecycle(t *testing.T) {
	// Skip if integration tests are not enabled or missing env vars
	if !shouldRunIntegrationTests(t) {
		return
	}

	// Enable parallel test execution
	t.Parallel()

	// Use unique identifier to prevent resource name conflicts
	uniqueID := generateUniqueID()
	portName := fmt.Sprintf("CLI-Test-LAG-%s-%d", uniqueID, time.Now().Unix())
	var portUID string
	var output string

	// STEP 1: Create a LAG port
	t.Run("Create LAG Port", func(t *testing.T) {
		cmd := exec.Command("megaport-cli", "ports", "buy-lag",
			"--name", portName,
			"--term", "1",
			"--port-speed", "10000",
			"--location-id", "67",
			"--lag-count", "1",
			"--marketplace-visibility", "false")

		// Capture both stdout and stderr
		outputBytes, err := cmd.CombinedOutput()
		output = string(outputBytes)
		t.Logf("Output: %s", output)

		// Check for successful execution
		assert.NoError(t, err, "Failed to create LAG port: %s", output)
		assert.Contains(t, output, "LAG Port created")

		// Extract port UID from the output
		portUID = extractPortUID(output)
		require.NotEmpty(t, portUID, "Failed to extract LAG port UID from output")
		t.Logf("Created LAG port with UID: %s", portUID)
	})

	// Wait for LAG port to be ready
	t.Run("Verify LAG Port Creation", func(t *testing.T) {
		// Skip if previous step failed
		if portUID == "" {
			t.Skip("Skipping because LAG port creation failed")
		}

		maxAttempts := 20
		for attempt := 1; attempt <= maxAttempts; attempt++ {
			t.Logf("Checking LAG port status, attempt %d of %d", attempt, maxAttempts)

			cmd := exec.Command("megaport-cli", "ports", "get", portUID)
			outputBytes, err := cmd.CombinedOutput()
			output = string(outputBytes)

			if err == nil && strings.Contains(output, "CONFIGURED") || strings.Contains(output, "LIVE") {
				t.Logf("LAG Port is ready")
				break
			}

			if attempt == maxAttempts {
				t.Fatalf("LAG Port did not reach CONFIGURED state within the expected time: %s", output)
			}

			time.Sleep(30 * time.Second)
		}
	})

	// STEP 2: Update the LAG port
	t.Run("Update LAG Port", func(t *testing.T) {
		// Skip if previous step failed
		if portUID == "" {
			t.Skip("Skipping because LAG port creation failed")
		}

		newName := fmt.Sprintf("%s-Updated", portName)
		cmd := exec.Command("megaport-cli", "ports", "update", portUID,
			"--name", newName)

		outputBytes, err := cmd.CombinedOutput()
		output = string(outputBytes)
		t.Logf("Update output: %s", output)

		assert.NoError(t, err, "Failed to update LAG port: %s", output)
		assert.Contains(t, output, "Port updated")

		// Verify the update
		cmd = exec.Command("megaport-cli", "ports", "get", portUID)
		outputBytes, err = cmd.CombinedOutput()
		output = string(outputBytes)

		assert.NoError(t, err, "Failed to get updated LAG port info: %s", output)
		assert.Contains(t, output, newName)
	})

	// STEP 3: Delete the LAG port
	t.Run("Delete LAG Port", func(t *testing.T) {
		// Skip if previous step failed
		if portUID == "" {
			t.Skip("Skipping because LAG port creation failed")
		}

		cmd := exec.Command("megaport-cli", "ports", "delete", portUID, "--now", "--force")

		outputBytes, err := cmd.CombinedOutput()
		output = string(outputBytes)
		t.Logf("Delete output: %s", output)

		assert.NoError(t, err, "Failed to delete LAG port: %s", output)
		assert.Contains(t, output, "Deleting port")

		// Wait for deletion to start
		time.Sleep(5 * time.Second)

		// Verify the deletion status
		cmd = exec.Command("megaport-cli", "ports", "get", portUID)
		outputBytes, err = cmd.CombinedOutput()
		output = string(outputBytes)

		// The port should still exist but be in DECOMMISSIONING or DECOMMISSIONED state
		assert.NoError(t, err)
		assert.True(t, strings.Contains(output, "DECOMMISSIONING") || strings.Contains(output, "DECOMMISSIONED"),
			"Expected DECOMMISSIONING or DECOMMISSIONED in output: %s", output)
	})
}

// TestJSONInputLifecycle tests port lifecycle with JSON input
func TestJSONInputLifecycle(t *testing.T) {
	// Skip if integration tests are not enabled or missing env vars
	if !shouldRunIntegrationTests(t) {
		return
	}

	// Enable parallel test execution
	t.Parallel()

	// Use unique identifier to prevent resource name conflicts
	uniqueID := generateUniqueID()
	portName := fmt.Sprintf("CLI-JSON-Test-%s-%d", uniqueID, time.Now().Unix())
	var portUID string

	// Create JSON file for port creation
	jsonContent := fmt.Sprintf(`{
        "name": "%s",
        "term": 1,
        "portSpeed": 1000,
        "locationId": 67,
        "marketPlaceVisibility": false
    }`, portName)

	jsonFile := createTempJSONFile(t, jsonContent)
	defer os.Remove(jsonFile)

	// STEP 1: Create a port using JSON file
	t.Run("Create Port with JSON", func(t *testing.T) {
		cmd := exec.Command("megaport-cli", "ports", "buy", "--json-file", jsonFile)
		outputBytes, err := cmd.CombinedOutput()
		output := string(outputBytes)

		assert.NoError(t, err, "Failed to create port with JSON: %s", output)
		assert.Contains(t, output, "Port created")

		portUID = extractPortUID(output)
		require.NotEmpty(t, portUID)
	})

	// STEP 2: Update port using JSON
	t.Run("Update Port with JSON", func(t *testing.T) {
		// Skip if previous step failed
		if portUID == "" {
			t.Skip("Skipping because port creation failed")
		}

		newName := fmt.Sprintf("%s-Updated-JSON", portName)
		updateJSON := fmt.Sprintf(`{"name": "%s"}`, newName)

		cmd := exec.Command("megaport-cli", "ports", "update", portUID,
			"--json", updateJSON)

		outputBytes, err := cmd.CombinedOutput()
		output := string(outputBytes)

		assert.NoError(t, err, "Failed to update port with JSON: %s", output)
		assert.Contains(t, output, "Port updated")

		// Verify the update
		cmd = exec.Command("megaport-cli", "ports", "get", portUID)
		outputBytes, err = cmd.CombinedOutput()
		output = string(outputBytes)

		assert.NoError(t, err)
		assert.Contains(t, output, newName)
	})

	// STEP 3: Delete the port
	t.Run("Delete JSON-created Port", func(t *testing.T) {
		// Skip if previous step failed
		if portUID == "" {
			t.Skip("Skipping because port creation failed")
		}

		cmd := exec.Command("megaport-cli", "ports", "delete", portUID, "--now", "--force")
		outputBytes, err := cmd.CombinedOutput()
		output := string(outputBytes)

		assert.NoError(t, err, "Failed to delete port: %s", output)
		assert.Contains(t, output, "Deleting port")
	})
}

// TestTempJSONFileLifecycle tests port lifecycle with JSON file for both creation and update
func TestTempJSONFileLifecycle(t *testing.T) {
	// Skip if integration tests are not enabled or missing env vars
	if !shouldRunIntegrationTests(t) {
		return
	}

	// Enable parallel test execution
	t.Parallel()

	// Use unique identifier to prevent resource name conflicts
	uniqueID := generateUniqueID()
	portName := fmt.Sprintf("CLI-TempJSON-%s-%d", uniqueID, time.Now().Unix())
	var portUID string

	// STEP 1: Create a port using JSON in temp file
	t.Run("Create Port with Temp JSON File", func(t *testing.T) {
		// Create JSON file for port creation
		createJSON := fmt.Sprintf(`{
            "name": "%s",
            "term": 1,
            "portSpeed": 1000,
            "locationId": 67,
            "marketPlaceVisibility": false
        }`, portName)

		createJSONFile := createTempJSONFile(t, createJSON)
		defer os.Remove(createJSONFile)

		cmd := exec.Command("megaport-cli", "ports", "buy", "--json-file", createJSONFile)
		outputBytes, err := cmd.CombinedOutput()
		output := string(outputBytes)
		t.Logf("Create output: %s", output)

		assert.NoError(t, err, "Failed to create port with JSON file: %s", output)
		assert.Contains(t, output, "Port created")

		portUID = extractPortUID(output)
		require.NotEmpty(t, portUID, "Failed to extract port UID from output")
	})

	// Wait for port to be ready
	t.Run("Verify Port Creation", func(t *testing.T) {
		// Skip if previous step failed
		if portUID == "" {
			t.Skip("Skipping because port creation failed")
		}

		maxAttempts := 20
		for attempt := 1; attempt <= maxAttempts; attempt++ {
			t.Logf("Checking port status, attempt %d of %d", attempt, maxAttempts)

			cmd := exec.Command("megaport-cli", "ports", "get", portUID)
			outputBytes, err := cmd.CombinedOutput()
			output := string(outputBytes)

			if err == nil && strings.Contains(output, "CONFIGURED") || strings.Contains(output, "LIVE") {
				t.Logf("Port is ready")
				break
			}

			if attempt == maxAttempts {
				t.Fatalf("Port did not reach CONFIGURED state within the expected time: %s", output)
			}

			time.Sleep(30 * time.Second)
		}
	})

	// STEP 2: Update port using a different JSON temp file
	t.Run("Update Port with Temp JSON File", func(t *testing.T) {
		// Skip if previous step failed
		if portUID == "" {
			t.Skip("Skipping because port creation failed")
		}

		// Create a new JSON file for updating
		newName := fmt.Sprintf("%s-Updated-TempJSON", portName)
		updateJSON := fmt.Sprintf(`{
            "name": "%s",
            "marketPlaceVisibility": true
        }`, newName)

		updateJSONFile := createTempJSONFile(t, updateJSON)
		defer os.Remove(updateJSONFile)

		cmd := exec.Command("megaport-cli", "ports", "update", portUID, "--json-file", updateJSONFile)
		outputBytes, err := cmd.CombinedOutput()
		output := string(outputBytes)
		t.Logf("Update output: %s", output)

		assert.NoError(t, err, "Failed to update port with JSON file: %s", output)
		assert.Contains(t, output, "Port updated")

		// Verify the update
		cmd = exec.Command("megaport-cli", "ports", "get", portUID)
		outputBytes, err = cmd.CombinedOutput()
		output = string(outputBytes)

		assert.NoError(t, err)
		assert.Contains(t, output, newName)
	})

	// STEP 3: Delete the port
	t.Run("Delete JSON File-created Port", func(t *testing.T) {
		// Skip if previous step failed
		if portUID == "" {
			t.Skip("Skipping because port creation failed")
		}

		cmd := exec.Command("megaport-cli", "ports", "delete", portUID, "--now", "--force")
		outputBytes, err := cmd.CombinedOutput()
		output := string(outputBytes)
		t.Logf("Delete output: %s", output)

		assert.NoError(t, err, "Failed to delete port: %s", output)
		assert.Contains(t, output, "Deleting port")

		// Wait for deletion to start
		time.Sleep(5 * time.Second)

		// Verify the deletion status
		cmd = exec.Command("megaport-cli", "ports", "get", portUID)
		outputBytes, err = cmd.CombinedOutput()
		output = string(outputBytes)

		// The port should still exist but be in DECOMMISSIONING or DECOMMISSIONED state
		assert.NoError(t, err)
		assert.True(t, strings.Contains(output, "DECOMMISSIONING") || strings.Contains(output, "DECOMMISSIONED"),
			"Expected DECOMMISSIONING or DECOMMISSIONED in output: %s", output)
	})
}

// extractPortUID extracts the port UID from command output
func extractPortUID(output string) string {
	// This function needs to be adapted based on actual CLI output format
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Port created") {
			parts := strings.Split(line, "Port created")
			if len(parts) > 1 {
				return strings.TrimSpace(parts[1])
			}
		}
		if strings.Contains(line, "LAG Port created") {
			parts := strings.Split(line, "LAG Port created")
			if len(parts) > 1 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return ""
}
