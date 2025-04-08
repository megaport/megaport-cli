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

// TestMVELifecycle is an integration test that tests the full lifecycle of an MVE
// It requires actual API credentials to be set as environment variables
// MEGAPORT_ACCESS_KEY, MEGAPORT_SECRET_KEY, and MEGAPORT_ENVIRONMENT=staging
func TestMVELifecycle(t *testing.T) {
	// Skip if integration tests are not enabled or missing env vars
	if !shouldRunIntegrationTests(t) {
		return
	}

	// Enable parallel test execution
	t.Parallel()

	// Use unique identifier to prevent resource name conflicts
	uniqueID := generateUniqueID()
	mveName := fmt.Sprintf("CLI-Test-MVE-%s-%d", uniqueID, time.Now().Unix())
	var mveUID string
	var output string

	// STEP 1: Create an MVE with flag-based input
	t.Run("Create MVE", func(t *testing.T) {
		vendorConfig := `{
            "vendor": "aruba",
            "productSize": "MEDIUM", 
            "imageId": 23, 
            "accountName": "test", 
            "accountKey": "test", 
            "systemTag": "test"
        }`

		vnicsConfig := `[{"description": "MVE VNIC 1", "vlan": 55}]`

		cmd := exec.Command("megaport-cli", "mve", "buy",
			"--name", mveName,
			"--term", "1",
			"--location-id", "65",
			"--vendor-config", vendorConfig,
			"--vnics", vnicsConfig)

		// Capture both stdout and stderr
		outputBytes, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Command failed: %v", err)
		}
		output = string(outputBytes)
		t.Logf("Output: %s", output)

		// Check for successful execution
		assert.NoError(t, err, "Failed to create MVE: %s", output)
		assert.Contains(t, output, "MVE created")

		// Extract MVE UID from the output
		mveUID = extractMVEUID(output)
		require.NotEmpty(t, mveUID, "Failed to extract MVE UID from output")
		t.Logf("Created MVE with UID: %s", mveUID)
	})

	// Wait for MVE to be ready
	t.Run("Verify MVE Creation", func(t *testing.T) {
		// Skip if previous step failed
		if mveUID == "" {
			t.Skip("Skipping because MVE creation failed")
		}

		maxAttempts := 20
		for attempt := 1; attempt <= maxAttempts; attempt++ {
			t.Logf("Checking MVE status, attempt %d of %d", attempt, maxAttempts)

			cmd := exec.Command("megaport-cli", "mve", "get", mveUID)
			outputBytes, err := cmd.CombinedOutput()
			output = string(outputBytes)

			// Check for either CONFIGURED or LIVE state
			if err == nil && (strings.Contains(output, "CONFIGURED") || strings.Contains(output, "LIVE")) {
				t.Logf("MVE is ready")
				break
			}

			if attempt == maxAttempts {
				t.Fatalf("MVE did not reach CONFIGURED or LIVE state within the expected time: %s", output)
			}

			// MVE provisioning can take longer than other resources
			time.Sleep(30 * time.Second)
		}
	})

	// STEP 2: Update the MVE
	t.Run("Update MVE", func(t *testing.T) {
		// Skip if previous step failed
		if mveUID == "" {
			t.Skip("Skipping because MVE creation failed")
		}

		newName := fmt.Sprintf("%s-Updated", mveName)
		cmd := exec.Command("megaport-cli", "mve", "update", mveUID,
			"--name", newName)

		outputBytes, err := cmd.CombinedOutput()
		output = string(outputBytes)
		t.Logf("Update output: %s", output)

		assert.NoError(t, err, "Failed to update MVE: %s", output)
		assert.Contains(t, output, "MVE updated")

		// Verify the update
		cmd = exec.Command("megaport-cli", "mve", "get", mveUID)
		outputBytes, err = cmd.CombinedOutput()
		output = string(outputBytes)

		assert.NoError(t, err, "Failed to get MVE info: %s", output)
		assert.Contains(t, output, newName)
	})

	// STEP 3: Delete the MVE
	t.Run("Delete MVE", func(t *testing.T) {
		// Skip if previous step failed
		if mveUID == "" {
			t.Skip("Skipping because MVE creation failed")
		}

		cmd := exec.Command("megaport-cli", "mve", "delete", mveUID, "--now", "--force")
		outputBytes, err := cmd.CombinedOutput()
		output = string(outputBytes)
		t.Logf("Delete output: %s", output)

		assert.NoError(t, err, "Failed to delete MVE: %s", output)
		assert.Contains(t, output, "Deleting MVE")

		// Wait for deletion to start
		time.Sleep(5 * time.Second)

		// Verify the deletion status
		cmd = exec.Command("megaport-cli", "mve", "get", mveUID)
		outputBytes, err = cmd.CombinedOutput()
		output = string(outputBytes)

		// The MVE should still exist but be in DECOMMISSIONING or DECOMMISSIONED state
		assert.NoError(t, err)
		assert.True(t, strings.Contains(output, "DECOMMISSIONING") || strings.Contains(output, "DECOMMISSIONED"),
			"Expected DECOMMISSIONING or DECOMMISSIONED status in output: %s", output)
	})
}

// TestMVEJSONLifecycle tests MVE lifecycle with JSON input
func TestMVEJSONLifecycle(t *testing.T) {
	// Skip if integration tests are not enabled or missing env vars
	if !shouldRunIntegrationTests(t) {
		return
	}

	// Enable parallel test execution
	t.Parallel()

	// Use unique identifier to prevent resource name conflicts
	uniqueID := generateUniqueID()
	mveName := fmt.Sprintf("CLI-JSON-MVE-%s-%d", uniqueID, time.Now().Unix())
	var mveUID string

	// STEP 1: Create MVE using JSON file
	t.Run("Create MVE with JSON", func(t *testing.T) {
		// Create JSON content for MVE
		mveJSON := fmt.Sprintf(`{
            "name": "%s",
            "term": 1,
            "locationId": 65,
            "vendorConfig": {
                "vendor": "aruba",
                "productSize": "MEDIUM",
                "imageId": 23,
                "accountName": "test",
                "accountKey": "test",
                "systemTag": "test"
            },
            "vnics": [
                {
                    "description": "MVE VNIC 1",
                    "vlan": 55
                }
            ]
        }`, mveName)

		mveJSONFile := createTempJSONFile(t, mveJSON)
		defer os.Remove(mveJSONFile)

		cmd := exec.Command("megaport-cli", "mve", "buy", "--json-file", mveJSONFile)
		outputBytes, err := cmd.CombinedOutput()
		output := string(outputBytes)
		t.Logf("Create MVE output: %s", output)

		assert.NoError(t, err, "Failed to create MVE with JSON: %s", output)
		assert.Contains(t, output, "MVE created")

		mveUID = extractMVEUID(output)
		require.NotEmpty(t, mveUID, "Failed to extract MVE UID")
	})

	// Wait for MVE to be ready
	t.Run("Verify JSON MVE Creation", func(t *testing.T) {
		// Skip if previous step failed
		if mveUID == "" {
			t.Skip("Skipping because MVE creation failed")
		}

		maxAttempts := 20
		for attempt := 1; attempt <= maxAttempts; attempt++ {
			t.Logf("Checking MVE status, attempt %d of %d", attempt, maxAttempts)

			cmd := exec.Command("megaport-cli", "mve", "get", mveUID)
			outputBytes, err := cmd.CombinedOutput()
			output := string(outputBytes)

			if err == nil && (strings.Contains(output, "CONFIGURED") || strings.Contains(output, "LIVE")) {
				t.Logf("MVE is ready")
				break
			}

			if attempt == maxAttempts {
				t.Fatalf("MVE did not reach CONFIGURED or LIVE state within the expected time")
			}

			time.Sleep(30 * time.Second)
		}
	})

	// STEP 2: Update MVE using JSON
	t.Run("Update MVE with JSON", func(t *testing.T) {
		// Skip if previous step failed
		if mveUID == "" {
			t.Skip("Skipping because MVE creation failed")
		}

		newName := fmt.Sprintf("%s-JSON-Updated", mveName)
		updateJSON := fmt.Sprintf(`{
            "name": "%s",
            "costCentre": "Testing Department"
        }`, newName)

		updateJSONFile := createTempJSONFile(t, updateJSON)
		defer os.Remove(updateJSONFile)

		cmd := exec.Command("megaport-cli", "mve", "update", mveUID, "--json-file", updateJSONFile)
		outputBytes, err := cmd.CombinedOutput()
		output := string(outputBytes)
		t.Logf("Update MVE output: %s", output)

		assert.NoError(t, err, "Failed to update MVE with JSON: %s", output)
		assert.Contains(t, output, "MVE updated")

		// Verify the update
		cmd = exec.Command("megaport-cli", "mve", "get", mveUID)
		outputBytes, err = cmd.CombinedOutput()
		output = string(outputBytes)

		assert.NoError(t, err, "Failed to get updated MVE info: %s", output)
		assert.Contains(t, output, newName)
		assert.Contains(t, output, "Testing Department")
	})

	// STEP 3: Delete the MVE
	t.Run("Delete JSON Created MVE", func(t *testing.T) {
		// Skip if previous steps failed
		if mveUID == "" {
			t.Skip("Skipping because MVE creation failed")
		}

		cmd := exec.Command("megaport-cli", "mve", "delete", mveUID, "--now", "--force")
		outputBytes, err := cmd.CombinedOutput()
		output := string(outputBytes)
		t.Logf("Delete output: %s", output)

		assert.NoError(t, err, "Failed to delete MVE: %s", output)
		assert.Contains(t, output, "Deleting MVE")
	})
}

// TestMVEInteractiveOutput tests the interactive mode output format (without actually running interactive mode)
func TestMVEInteractiveOutput(t *testing.T) {
	// This test just verifies that the command execution structure is valid
	// It doesn't actually run interactive mode (which would require user input)

	cmd := exec.Command("megaport-cli", "mve", "buy", "--help")
	outputBytes, err := cmd.CombinedOutput()
	output := string(outputBytes)

	assert.NoError(t, err, "Failed to get MVE buy help: %s", output)
	assert.Contains(t, output, "--interactive")
}

// extractMVEUID extracts the MVE UID from command output
func extractMVEUID(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "MVE created") {
			parts := strings.Split(line, "MVE created")
			if len(parts) > 1 {
				return strings.TrimSpace(parts[1])
			}
		}
		// For our enhanced output format
		if strings.Contains(line, "MVE created successfully - ID:") {
			parts := strings.Split(line, "ID:")
			if len(parts) > 1 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return ""
}
