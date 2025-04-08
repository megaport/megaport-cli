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

// TestMCRLifecycle is an integration test that tests the full lifecycle of an MCR
// including prefix filter list operations
func TestMCRLifecycle(t *testing.T) {
	// Skip if integration tests are not enabled or missing env vars
	if !shouldRunIntegrationTests(t) {
		return
	}

	// Enable parallel test execution
	t.Parallel()

	// Use unique identifier to prevent resource name conflicts
	uniqueID := generateUniqueID()
	mcrName := fmt.Sprintf("CLI-Test-MCR-%s-%d", uniqueID, time.Now().Unix())
	var mcrUID string
	var prefixFilterListID string
	var output string

	// STEP 1: Create an MCR
	t.Run("Create MCR", func(t *testing.T) {
		cmd := exec.Command("megaport-cli", "mcr", "buy",
			"--name", mcrName,
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
		assert.NoError(t, err, "Failed to create MCR: %s", output)
		assert.Contains(t, output, "MCR created")

		// Extract MCR UID from the output
		mcrUID = extractMCRUID(output)
		require.NotEmpty(t, mcrUID, "Failed to extract MCR UID from output")
		t.Logf("Created MCR with UID: %s", mcrUID)
	})

	// Wait for MCR to be ready
	t.Run("Verify MCR Creation", func(t *testing.T) {
		// Skip if previous step failed
		if mcrUID == "" {
			t.Skip("Skipping because MCR creation failed")
		}

		maxAttempts := 20
		for attempt := 1; attempt <= maxAttempts; attempt++ {
			t.Logf("Checking MCR status, attempt %d of %d", attempt, maxAttempts)

			cmd := exec.Command("megaport-cli", "mcr", "get", mcrUID)
			outputBytes, err := cmd.CombinedOutput()
			output = string(outputBytes)

			if err == nil && strings.Contains(output, "CONFIGURED") {
				t.Logf("MCR is ready")
				break
			}

			if attempt == maxAttempts {
				t.Fatalf("MCR did not reach CONFIGURED state within the expected time: %s", output)
			}

			time.Sleep(30 * time.Second)
		}
	})

	// STEP 2: Create a prefix filter list on the MCR
	t.Run("Create Prefix Filter List", func(t *testing.T) {
		// Skip if previous step failed
		if mcrUID == "" {
			t.Skip("Skipping because MCR creation failed")
		}

		// Create JSON content for the prefix filter list
		prefixListJSON := `{
            "description": "Test Prefix Filter List",
            "addressFamily": "IPv4",
            "entries": [
                {
                    "action": "permit",
                    "prefix": "10.0.1.0/24",
                    "ge": 25,
                    "le": 32
                },
                {
                    "action": "deny",
                    "prefix": "10.0.2.0/24",
                    "ge": 0,
                    "le": 25
                }
            ]
        }`

		// Create temp file for the JSON
		prefixListFile := createTempJSONFile(t, prefixListJSON)
		defer os.Remove(prefixListFile)

		cmd := exec.Command("megaport-cli", "mcr", "create-prefix-filter-list", mcrUID,
			"--json-file", prefixListFile)

		outputBytes, err := cmd.CombinedOutput()
		output = string(outputBytes)
		t.Logf("Create prefix filter list output: %s", output)

		assert.NoError(t, err, "Failed to create prefix filter list: %s", output)
		assert.Contains(t, output, "Prefix filter list created")

		// Extract prefix filter list ID from output
		prefixFilterListID = extractPrefixFilterListID(output)
		require.NotEmpty(t, prefixFilterListID, "Failed to extract prefix filter list ID")
		t.Logf("Created prefix filter list with ID: %s", prefixFilterListID)
	})

	// STEP 3: Verify the prefix filter list was created
	t.Run("Get Prefix Filter List", func(t *testing.T) {
		// Skip if previous steps failed
		if mcrUID == "" || prefixFilterListID == "" {
			t.Skip("Skipping because previous step failed")
		}

		cmd := exec.Command("megaport-cli", "mcr", "get-prefix-filter-list", mcrUID, prefixFilterListID)
		outputBytes, err := cmd.CombinedOutput()
		output = string(outputBytes)
		t.Logf("Get prefix filter list output: %s", output)

		assert.NoError(t, err, "Failed to get prefix filter list details: %s", output)
		assert.Contains(t, output, "Test Prefix Filter List")
		assert.Contains(t, output, "10.0.1.0/24")
		assert.Contains(t, output, "10.0.2.0/24")
	})

	// STEP 4: Update the prefix filter list
	t.Run("Update Prefix Filter List", func(t *testing.T) {
		// Skip if previous steps failed
		if mcrUID == "" || prefixFilterListID == "" {
			t.Skip("Skipping because previous step failed")
		}

		// Create JSON for update
		updateJSON := `{
            "description": "Test Prefix Filter List JSON File Updated New 2",
            "entries": [
                {
                    "action": "permit",
                    "prefix": "10.0.1.0/24",
                    "ge": 25,
                    "le": 32
                },
                {
                    "action": "deny",
                    "prefix": "10.0.2.0/24",
                    "ge": 0,
                    "le": 25
                },
                {
                    "action": "permit",
                    "prefix": "192.168.0.0/16",
                    "ge": 24,
                    "le": 32
                }
            ]
        }`

		updateFile := createTempJSONFile(t, updateJSON)
		defer os.Remove(updateFile)

		cmd := exec.Command("megaport-cli", "mcr", "update-prefix-filter-list", mcrUID, prefixFilterListID,
			"--json-file", updateFile)

		outputBytes, err := cmd.CombinedOutput()
		output = string(outputBytes)
		t.Logf("Update prefix filter list output: %s", output)

		assert.NoError(t, err, "Failed to update prefix filter list: %s", output)
		assert.Contains(t, output, "Prefix filter list updated")

		// Verify the update
		cmd = exec.Command("megaport-cli", "mcr", "get-prefix-filter-list", mcrUID, prefixFilterListID)
		outputBytes, err = cmd.CombinedOutput()
		output = string(outputBytes)

		assert.NoError(t, err)
		assert.Contains(t, output, "Test Prefix Filter List JSON File Updated New 2")
		assert.Contains(t, output, "192.168.0.0/16") // New entry
	})

	// STEP 5: Delete the prefix filter list
	t.Run("Delete Prefix Filter List", func(t *testing.T) {
		// Skip if previous steps failed
		if mcrUID == "" || prefixFilterListID == "" {
			t.Skip("Skipping because previous step failed")
		}

		cmd := exec.Command("megaport-cli", "mcr", "delete-prefix-filter-list", mcrUID, prefixFilterListID, "--force")
		outputBytes, err := cmd.CombinedOutput()
		output = string(outputBytes)
		t.Logf("Delete prefix filter list output: %s", output)

		assert.NoError(t, err, "Failed to delete prefix filter list: %s", output)
		assert.Contains(t, output, "Prefix filter list deleted")

		// Verify it's deleted - should get an error or empty list
		cmd = exec.Command("megaport-cli", "mcr", "get-prefix-filter-list", mcrUID, prefixFilterListID)
		outputBytes, err = cmd.CombinedOutput()
		output = string(outputBytes)
		t.Logf("Get deleted prefix filter list output: %s", output)

		// It's either an error or "not found" message
		assert.True(t, err != nil || strings.Contains(output, "not found") ||
			strings.Contains(output, "does not exist"),
			"Prefix filter list should not exist anymore")
	})

	// STEP 6: Update the MCR
	t.Run("Update MCR", func(t *testing.T) {
		// Skip if previous steps failed
		if mcrUID == "" {
			t.Skip("Skipping because MCR creation failed")
		}

		newName := fmt.Sprintf("%s-Updated", mcrName)
		cmd := exec.Command("megaport-cli", "mcr", "update", mcrUID,
			"--name", newName)

		outputBytes, err := cmd.CombinedOutput()
		output = string(outputBytes)
		t.Logf("Update MCR output: %s", output)

		assert.NoError(t, err, "Failed to update MCR: %s", output)
		assert.Contains(t, output, "MCR updated")

		// Verify the update
		cmd = exec.Command("megaport-cli", "mcr", "get", mcrUID)
		outputBytes, err = cmd.CombinedOutput()
		output = string(outputBytes)

		assert.NoError(t, err, "Failed to get MCR info: %s", output)
		assert.Contains(t, output, newName)
	})

	// STEP 7: Delete the MCR
	t.Run("Delete MCR", func(t *testing.T) {
		// Skip if previous step failed
		if mcrUID == "" {
			t.Skip("Skipping because MCR creation failed")
		}

		cmd := exec.Command("megaport-cli", "mcr", "delete", mcrUID, "--now", "--force")
		outputBytes, err := cmd.CombinedOutput()
		output = string(outputBytes)
		t.Logf("Delete MCR output: %s", output)

		assert.NoError(t, err, "Failed to delete MCR: %s", output)
		assert.Contains(t, output, "Deleting MCR")

		// Wait for deletion to start
		time.Sleep(5 * time.Second)

		// Verify the deletion status
		cmd = exec.Command("megaport-cli", "mcr", "get", mcrUID)
		outputBytes, err = cmd.CombinedOutput()
		output = string(outputBytes)

		// The MCR should still exist but be in DECOMMISSIONING state
		assert.NoError(t, err)
		assert.True(t, strings.Contains(output, "DECOMMISSIONING") || strings.Contains(output, "DECOMMISSIONED"),
			"Expected DECOMMISSIONING or DECOMMISSIONED status in output: %s", output)
	})
}

// extractMCRUID extracts the MCR UID from the command output
func extractMCRUID(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "MCR created") {
			parts := strings.Split(line, "MCR created")
			if len(parts) > 1 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return ""
}

// extractPrefixFilterListID extracts the prefix filter list ID from command output
func extractPrefixFilterListID(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Prefix filter list created") {
			parts := strings.Split(line, "ID:")
			if len(parts) > 1 {
				return strings.TrimSpace(parts[1])
			}
		}
	}

	// Alternative format parsing if the above doesn't work
	for _, line := range lines {
		if strings.Contains(line, "successfully") && strings.Contains(line, "ID:") {
			parts := strings.Split(line, "ID:")
			if len(parts) > 1 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return ""
}

// TestMCRJSONLifecycle tests MCR lifecycle with JSON input
func TestMCRJSONLifecycle(t *testing.T) {
	// Skip if integration tests are not enabled or missing env vars
	if !shouldRunIntegrationTests(t) {
		return
	}

	// Enable parallel test execution
	t.Parallel()

	// Use unique identifier to prevent resource name conflicts
	uniqueID := generateUniqueID()
	mcrName := fmt.Sprintf("CLI-JSON-MCR-%s-%d", uniqueID, time.Now().Unix())
	var mcrUID string
	var prefixFilterListID string

	// STEP 1: Create MCR using JSON
	t.Run("Create MCR with JSON", func(t *testing.T) {
		// Create JSON content for MCR
		mcrJSON := fmt.Sprintf(`{
            "name": "%s",
            "term": 1,
            "portSpeed": 1000,
            "locationId": 67,
            "marketPlaceVisibility": false
        }`, mcrName)

		mcrJSONFile := createTempJSONFile(t, mcrJSON)
		defer os.Remove(mcrJSONFile)

		cmd := exec.Command("megaport-cli", "mcr", "buy", "--json-file", mcrJSONFile)
		outputBytes, err := cmd.CombinedOutput()
		output := string(outputBytes)
		t.Logf("Create MCR output: %s", output)

		assert.NoError(t, err, "Failed to create MCR with JSON: %s", output)
		assert.Contains(t, output, "MCR created")

		mcrUID = extractMCRUID(output)
		require.NotEmpty(t, mcrUID, "Failed to extract MCR UID")
	})

	// Wait for MCR to be ready
	t.Run("Verify JSON MCR Creation", func(t *testing.T) {
		// Skip if previous step failed
		if mcrUID == "" {
			t.Skip("Skipping because MCR creation failed")
		}

		maxAttempts := 20
		for attempt := 1; attempt <= maxAttempts; attempt++ {
			t.Logf("Checking MCR status, attempt %d of %d", attempt, maxAttempts)

			cmd := exec.Command("megaport-cli", "mcr", "get", mcrUID)
			outputBytes, err := cmd.CombinedOutput()
			output := string(outputBytes)

			if err == nil && strings.Contains(output, "CONFIGURED") || strings.Contains(output, "LIVE") {
				t.Logf("MCR is ready")
				break
			}

			if attempt == maxAttempts {
				t.Fatalf("MCR did not reach CONFIGURED or LIVE state within the expected time")
			}

			time.Sleep(30 * time.Second)
		}
	})

	// STEP 2: Create prefix filter list using direct JSON (not file)
	t.Run("Create Prefix Filter List with JSON", func(t *testing.T) {
		// Skip if previous steps failed
		if mcrUID == "" {
			t.Skip("Skipping because MCR creation failed")
		}

		jsonStr := `{"description":"Test JSON Direct Prefix List","addressFamily":"IPv4","entries":[{"action":"permit","prefix":"172.16.0.0/12","ge":16,"le":24},{"action":"deny","prefix":"192.168.0.0/16"}]}`

		cmd := exec.Command("megaport-cli", "mcr", "create-prefix-filter-list", mcrUID, "--json", jsonStr)
		outputBytes, err := cmd.CombinedOutput()
		output := string(outputBytes)
		t.Logf("Create prefix list output: %s", output)

		assert.NoError(t, err, "Failed to create prefix filter list with JSON: %s", output)
		assert.Contains(t, output, "Prefix filter list created")

		prefixFilterListID = extractPrefixFilterListID(output)
		require.NotEmpty(t, prefixFilterListID, "Failed to extract prefix filter list ID")
	})

	// STEP 3: Delete the MCR
	t.Run("Delete JSON Created MCR", func(t *testing.T) {
		// Skip if previous steps failed
		if mcrUID == "" {
			t.Skip("Skipping because MCR creation failed")
		}

		cmd := exec.Command("megaport-cli", "mcr", "delete", mcrUID, "--now", "--force")
		outputBytes, err := cmd.CombinedOutput()
		output := string(outputBytes)
		t.Logf("Delete output: %s", output)

		assert.NoError(t, err, "Failed to delete MCR: %s", output)
		assert.Contains(t, output, "Deleting MCR")
	})
}
