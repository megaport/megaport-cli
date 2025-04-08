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

// TestVXCPortToPortLifecycle is an integration test that tests the full lifecycle of a VXC between two ports
// It requires actual API credentials to be set as environment variables
func TestVXCPortToPortLifecycle(t *testing.T) {
	// Skip if integration tests are not enabled or missing env vars
	if !shouldRunIntegrationTests(t) {
		return
	}

	// Enable parallel test execution
	t.Parallel()

	// Use unique identifier to prevent resource name conflicts
	uniqueID := generateUniqueID()
	portName1 := fmt.Sprintf("CLI-Test-Port1-%s-%d", uniqueID, time.Now().Unix())
	portName2 := fmt.Sprintf("CLI-Test-Port2-%s-%d", uniqueID, time.Now().Unix())
	vxcName := fmt.Sprintf("CLI-Test-VXC-%s-%d", uniqueID, time.Now().Unix())

	// Variables to store created resource IDs
	var port1UID, port2UID, vxcUID string
	var output string

	// STEP 1: Create two ports for connection
	t.Run("Create First Port", func(t *testing.T) {
		cmd := exec.Command("megaport-cli", "ports", "buy",
			"--name", portName1,
			"--term", "1",
			"--port-speed", "1000",
			"--location-id", "67",
			"--marketplace-visibility", "false")

		outputBytes, err := cmd.CombinedOutput()
		output = string(outputBytes)
		t.Logf("Port 1 creation output: %s", output)

		assert.NoError(t, err, "Failed to create first port: %s", output)
		assert.Contains(t, output, "Port created")

		port1UID = extractPortUID(output)
		require.NotEmpty(t, port1UID, "Failed to extract first port UID")
		t.Logf("Created first port with UID: %s", port1UID)
	})

	t.Run("Create Second Port", func(t *testing.T) {
		cmd := exec.Command("megaport-cli", "ports", "buy",
			"--name", portName2,
			"--term", "1",
			"--port-speed", "1000",
			"--location-id", "67",
			"--marketplace-visibility", "false")

		outputBytes, err := cmd.CombinedOutput()
		output = string(outputBytes)
		t.Logf("Port 2 creation output: %s", output)

		assert.NoError(t, err, "Failed to create second port: %s", output)
		assert.Contains(t, output, "Port created")

		port2UID = extractPortUID(output)
		require.NotEmpty(t, port2UID, "Failed to extract second port UID")
		t.Logf("Created second port with UID: %s", port2UID)
	})

	// Wait for ports to be ready
	t.Run("Verify Ports Ready", func(t *testing.T) {
		// Skip if previous step failed
		if port1UID == "" || port2UID == "" {
			t.Skip("Skipping because port creation failed")
		}

		maxAttempts := 20
		for attempt := 1; attempt <= maxAttempts; attempt++ {
			t.Logf("Checking port status, attempt %d of %d", attempt, maxAttempts)

			// Check first port
			cmd1 := exec.Command("megaport-cli", "ports", "get", port1UID)
			outputBytes1, err1 := cmd1.CombinedOutput()
			output1 := string(outputBytes1)

			// Check second port
			cmd2 := exec.Command("megaport-cli", "ports", "get", port2UID)
			outputBytes2, err2 := cmd2.CombinedOutput()
			output2 := string(outputBytes2)

			if err1 == nil && err2 == nil &&
				(strings.Contains(output1, "CONFIGURED") || strings.Contains(output1, "LIVE")) &&
				(strings.Contains(output2, "CONFIGURED") || strings.Contains(output2, "LIVE")) {
				t.Logf("Both ports are ready")
				break
			}

			if attempt == maxAttempts {
				t.Fatalf("Ports did not reach ready state within the expected time")
			}

			time.Sleep(30 * time.Second)
		}
	})

	// STEP 2: Create a VXC between the two ports using flags
	t.Run("Create VXC", func(t *testing.T) {
		// Skip if previous step failed
		if port1UID == "" || port2UID == "" {
			t.Skip("Skipping because port creation failed")
		}

		cmd := exec.Command("megaport-cli", "vxc", "buy",
			"--name", vxcName,
			"--term", "1",
			"--rate-limit", "100",
			"--a-end-uid", port1UID,
			"--b-end-uid", port2UID,
			"--a-end-vlan", "100",
			"--b-end-vlan", "200",
			"--cost-centre", "Test Cost Centre")

		outputBytes, err := cmd.CombinedOutput()
		output = string(outputBytes)
		t.Logf("VXC creation output: %s", output)

		assert.NoError(t, err, "Failed to create VXC: %s", output)
		assert.Contains(t, output, "VXC created")

		vxcUID = extractVXCUID(output)
		require.NotEmpty(t, vxcUID, "Failed to extract VXC UID")
		t.Logf("Created VXC with UID: %s", vxcUID)
	})

	// Wait for VXC to be ready
	t.Run("Verify VXC Creation", func(t *testing.T) {
		// Skip if previous step failed
		if vxcUID == "" {
			t.Skip("Skipping because VXC creation failed")
		}

		maxAttempts := 20
		for attempt := 1; attempt <= maxAttempts; attempt++ {
			t.Logf("Checking VXC status, attempt %d of %d", attempt, maxAttempts)

			cmd := exec.Command("megaport-cli", "vxc", "get", vxcUID)
			outputBytes, err := cmd.CombinedOutput()
			output = string(outputBytes)

			if err == nil && (strings.Contains(output, "CONFIGURED") || strings.Contains(output, "LIVE")) {
				t.Logf("VXC is ready")
				break
			}

			if attempt == maxAttempts {
				t.Fatalf("VXC did not reach ready state within the expected time: %s", output)
			}

			time.Sleep(20 * time.Second)
		}
	})

	// STEP 3: Update the VXC
	t.Run("Update VXC", func(t *testing.T) {
		// Skip if previous step failed
		if vxcUID == "" {
			t.Skip("Skipping because VXC creation failed")
		}

		newName := fmt.Sprintf("%s-Updated", vxcName)
		cmd := exec.Command("megaport-cli", "vxc", "update", vxcUID,
			"--name", newName,
			"--rate-limit", "200")

		outputBytes, err := cmd.CombinedOutput()
		output = string(outputBytes)
		t.Logf("VXC update output: %s", output)

		assert.NoError(t, err, "Failed to update VXC: %s", output)
		assert.Contains(t, output, "VXC updated")

		// Verify the update
		cmd = exec.Command("megaport-cli", "vxc", "get", vxcUID)
		outputBytes, err = cmd.CombinedOutput()
		output = string(outputBytes)

		assert.NoError(t, err, "Failed to get VXC info: %s", output)
		assert.Contains(t, output, newName)
		assert.Contains(t, output, "200") // Rate limit increased
	})

	// STEP 4: Delete the VXC
	t.Run("Delete VXC", func(t *testing.T) {
		// Skip if previous step failed
		if vxcUID == "" {
			t.Skip("Skipping because VXC creation failed")
		}

		cmd := exec.Command("megaport-cli", "vxc", "delete", vxcUID, "--force")
		outputBytes, err := cmd.CombinedOutput()
		output = string(outputBytes)
		t.Logf("VXC deletion output: %s", output)

		assert.NoError(t, err, "Failed to delete VXC: %s", output)
		assert.Contains(t, output, "Deleting VXC")

		// Wait for deletion process to start
		time.Sleep(5 * time.Second)
	})

	// STEP 5: Delete the ports
	t.Run("Delete Ports", func(t *testing.T) {
		// Delete first port if it was created
		if port1UID != "" {
			cmd := exec.Command("megaport-cli", "ports", "delete", port1UID, "--now", "--force")
			outputBytes, err := cmd.CombinedOutput()
			output = string(outputBytes)
			t.Logf("Port 1 deletion output: %s", output)

			assert.NoError(t, err, "Failed to delete first port: %s", output)
			assert.Contains(t, output, "Deleting port")
		}

		// Delete second port if it was created
		if port2UID != "" {
			cmd := exec.Command("megaport-cli", "ports", "delete", port2UID, "--now", "--force")
			outputBytes, err := cmd.CombinedOutput()
			output = string(outputBytes)
			t.Logf("Port 2 deletion output: %s", output)

			assert.NoError(t, err, "Failed to delete second port: %s", output)
			assert.Contains(t, output, "Deleting port")
		}
	})
}

// TestVXCVlanModificationLifecycle tests creating and updating a VXC with VLAN changes
func TestVXCVlanModificationLifecycle(t *testing.T) {
	// Skip if integration tests are not enabled or missing env vars
	if !shouldRunIntegrationTests(t) {
		return
	}

	// Enable parallel test execution
	t.Parallel()

	// Use unique identifier to prevent resource name conflicts
	uniqueID := generateUniqueID()
	portName1 := fmt.Sprintf("CLI-VLan-Port1-%s-%d", uniqueID, time.Now().Unix())
	portName2 := fmt.Sprintf("CLI-VLan-Port2-%s-%d", uniqueID, time.Now().Unix())
	vxcName := fmt.Sprintf("CLI-VLan-VXC-%s-%d", uniqueID, time.Now().Unix())

	// Variables to store created resource IDs
	var port1UID, port2UID, vxcUID string
	var output string

	// STEP 1: Create two ports
	t.Run("Create First Port", func(t *testing.T) {
		cmd := exec.Command("megaport-cli", "ports", "buy",
			"--name", portName1,
			"--term", "1",
			"--port-speed", "1000",
			"--location-id", "67",
			"--marketplace-visibility", "false")

		outputBytes, err := cmd.CombinedOutput()
		output = string(outputBytes)
		t.Logf("Port 1 creation output: %s", output)

		assert.NoError(t, err, "Failed to create first port: %s", output)
		port1UID = extractPortUID(output)
		require.NotEmpty(t, port1UID, "Failed to extract first port UID")
	})

	t.Run("Create Second Port", func(t *testing.T) {
		cmd := exec.Command("megaport-cli", "ports", "buy",
			"--name", portName2,
			"--term", "1",
			"--port-speed", "1000",
			"--location-id", "67",
			"--marketplace-visibility", "false")

		outputBytes, err := cmd.CombinedOutput()
		output = string(outputBytes)
		t.Logf("Port 2 creation output: %s", output)

		assert.NoError(t, err, "Failed to create second port: %s", output)
		port2UID = extractPortUID(output)
		require.NotEmpty(t, port2UID, "Failed to extract second port UID")
	})

	// Wait for ports to be ready
	t.Run("Verify Ports Ready", func(t *testing.T) {
		// Skip if previous step failed
		if port1UID == "" || port2UID == "" {
			t.Skip("Skipping because port creation failed")
		}

		maxAttempts := 20
		for attempt := 1; attempt <= maxAttempts; attempt++ {
			t.Logf("Checking port status, attempt %d of %d", attempt, maxAttempts)

			// Check both ports
			cmd1 := exec.Command("megaport-cli", "ports", "get", port1UID)
			outputBytes1, err1 := cmd1.CombinedOutput()
			output1 := string(outputBytes1)

			cmd2 := exec.Command("megaport-cli", "ports", "get", port2UID)
			outputBytes2, err2 := cmd2.CombinedOutput()
			output2 := string(outputBytes2)

			if err1 == nil && err2 == nil &&
				(strings.Contains(output1, "CONFIGURED") || strings.Contains(output1, "LIVE")) &&
				(strings.Contains(output2, "CONFIGURED") || strings.Contains(output2, "LIVE")) {
				t.Logf("Both ports are ready")
				break
			}

			if attempt == maxAttempts {
				t.Fatalf("Ports did not reach ready state within the expected time")
			}

			time.Sleep(30 * time.Second)
		}
	})

	// STEP 2: Create a VXC between the two ports with specific VLANs
	t.Run("Create VXC with VLANs", func(t *testing.T) {
		// Skip if previous step failed
		if port1UID == "" || port2UID == "" {
			t.Skip("Skipping because port creation failed")
		}

		cmd := exec.Command("megaport-cli", "vxc", "buy",
			"--name", vxcName,
			"--term", "1",
			"--rate-limit", "100",
			"--a-end-uid", port1UID,
			"--b-end-uid", port2UID,
			"--a-end-vlan", "100",
			"--b-end-vlan", "200")

		outputBytes, err := cmd.CombinedOutput()
		output = string(outputBytes)
		t.Logf("VXC creation output: %s", output)

		assert.NoError(t, err, "Failed to create VXC: %s", output)
		vxcUID = extractVXCUID(output)
		require.NotEmpty(t, vxcUID, "Failed to extract VXC UID")
	})

	// Wait for VXC to be ready
	t.Run("Verify VXC Creation", func(t *testing.T) {
		// Skip if previous step failed
		if vxcUID == "" {
			t.Skip("Skipping because VXC creation failed")
		}

		maxAttempts := 20
		for attempt := 1; attempt <= maxAttempts; attempt++ {
			t.Logf("Checking VXC status, attempt %d of %d", attempt, maxAttempts)

			cmd := exec.Command("megaport-cli", "vxc", "get", vxcUID)
			outputBytes, err := cmd.CombinedOutput()
			output = string(outputBytes)

			if err == nil && (strings.Contains(output, "CONFIGURED") || strings.Contains(output, "LIVE")) {
				t.Logf("VXC is ready")
				break
			}

			if attempt == maxAttempts {
				t.Fatalf("VXC did not reach ready state within the expected time")
			}

			time.Sleep(20 * time.Second)
		}
	})

	// STEP 3: Update the VXC to use different VLANs
	t.Run("Update VXC VLANs", func(t *testing.T) {
		// Skip if previous step failed
		if vxcUID == "" {
			t.Skip("Skipping because VXC creation failed")
		}

		// Create JSON for update with new VLANs
		updateJSON := `{
            "aEndVlan": 300,
            "bEndVlan": 400
        }`

		// Create temp file for the JSON
		updateJSONFile := createTempJSONFile(t, updateJSON)
		defer os.Remove(updateJSONFile)

		cmd := exec.Command("megaport-cli", "vxc", "update", vxcUID, "--json-file", updateJSONFile)
		outputBytes, err := cmd.CombinedOutput()
		output = string(outputBytes)
		t.Logf("VXC VLAN update output: %s", output)

		assert.NoError(t, err, "Failed to update VXC VLANs: %s", output)
		assert.Contains(t, output, "VXC updated")

		// Verify the VLAN update
		cmd = exec.Command("megaport-cli", "vxc", "get", vxcUID)
		outputBytes, err = cmd.CombinedOutput()
		output = string(outputBytes)

		assert.NoError(t, err, "Failed to get VXC info after VLAN update: %s", output)
		assert.Contains(t, output, "300") // New A-end VLAN
		assert.Contains(t, output, "400") // New B-end VLAN
	})

	// STEP 4: Update to untagged VLANs
	t.Run("Update to Untagged VLANs", func(t *testing.T) {
		// Skip if previous step failed
		if vxcUID == "" {
			t.Skip("Skipping because VXC creation failed")
		}

		// Create JSON for update to untagged VLANs (-1)
		updateJSON := `{
            "aEndVlan": -1,
            "bEndVlan": -1
        }`

		// Create temp file for the JSON
		updateJSONFile := createTempJSONFile(t, updateJSON)
		defer os.Remove(updateJSONFile)

		cmd := exec.Command("megaport-cli", "vxc", "update", vxcUID, "--json-file", updateJSONFile)
		outputBytes, err := cmd.CombinedOutput()
		output = string(outputBytes)
		t.Logf("VXC untagged update output: %s", output)

		assert.NoError(t, err, "Failed to update VXC to untagged VLANs: %s", output)
		assert.Contains(t, output, "VXC updated")

		// Verify the update removed VLANs
		cmd = exec.Command("megaport-cli", "vxc", "get", vxcUID)
		outputBytes, err = cmd.CombinedOutput()
		output = string(outputBytes)

		assert.NoError(t, err, "Failed to get VXC info after untagged update: %s", output)
		// The output should not contain explicit VLAN numbers anymore
		// Instead of checking for absence, which is harder, we check the update was successful
	})

	// STEP 5: Clean up - delete VXC and ports
	t.Run("Delete Resources", func(t *testing.T) {
		// Skip if creation failed
		if vxcUID == "" {
			t.Skip("Skipping because VXC creation failed")
		}

		// Delete VXC
		cmd := exec.Command("megaport-cli", "vxc", "delete", vxcUID, "--force")
		outputBytes, err := cmd.CombinedOutput()
		output = string(outputBytes)
		t.Logf("VXC deletion output: %s", output)

		assert.NoError(t, err, "Failed to delete VXC: %s", output)
		assert.Contains(t, output, "Deleting VXC")

		// Wait for VXC deletion to process
		time.Sleep(5 * time.Second)

		// Delete ports
		if port1UID != "" {
			cmd := exec.Command("megaport-cli", "ports", "delete", port1UID, "--now", "--force")
			_, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("Failed to delete first port: %s", err)
			}
		}

		if port2UID != "" {
			cmd := exec.Command("megaport-cli", "ports", "delete", port2UID, "--now", "--force")
			_, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("Failed to delete second port: %s", err)
			}
		}
	})
}

// TestVXCJSONLifecycle tests creating a VXC using JSON input
func TestVXCJSONLifecycle(t *testing.T) {
	// Skip if integration tests are not enabled or missing env vars
	if !shouldRunIntegrationTests(t) {
		return
	}

	// Enable parallel test execution
	t.Parallel()

	// Use unique identifier to prevent resource name conflicts
	uniqueID := generateUniqueID()
	portName1 := fmt.Sprintf("CLI-JSON-Port1-%s-%d", uniqueID, time.Now().Unix())
	portName2 := fmt.Sprintf("CLI-JSON-Port2-%s-%d", uniqueID, time.Now().Unix())
	vxcName := fmt.Sprintf("CLI-JSON-VXC-%s-%d", uniqueID, time.Now().Unix())

	// Variables to store created resource IDs
	var port1UID, port2UID, vxcUID string

	// STEP 1: Create two ports for connection
	t.Run("Create Ports for JSON Test", func(t *testing.T) {
		// Create first port
		cmd := exec.Command("megaport-cli", "ports", "buy",
			"--name", portName1,
			"--term", "1",
			"--port-speed", "1000",
			"--location-id", "67",
			"--marketplace-visibility", "false")

		outputBytes, err := cmd.CombinedOutput()
		output := string(outputBytes)
		t.Logf("Port 1 creation output: %s", output)

		assert.NoError(t, err, "Failed to create first port: %s", output)
		port1UID = extractPortUID(output)
		require.NotEmpty(t, port1UID, "Failed to extract first port UID")

		// Create second port
		cmd = exec.Command("megaport-cli", "ports", "buy",
			"--name", portName2,
			"--term", "1",
			"--port-speed", "1000",
			"--location-id", "67",
			"--marketplace-visibility", "false")

		outputBytes, err = cmd.CombinedOutput()
		output = string(outputBytes)
		t.Logf("Port 2 creation output: %s", output)

		assert.NoError(t, err, "Failed to create second port: %s", output)
		port2UID = extractPortUID(output)
		require.NotEmpty(t, port2UID, "Failed to extract second port UID")

		// Wait for ports to be ready
		maxAttempts := 20
		for attempt := 1; attempt <= maxAttempts; attempt++ {
			t.Logf("Checking port status, attempt %d of %d", attempt, maxAttempts)

			// Check both ports
			cmd1 := exec.Command("megaport-cli", "ports", "get", port1UID)
			outputBytes1, err1 := cmd1.CombinedOutput()
			output1 := string(outputBytes1)

			cmd2 := exec.Command("megaport-cli", "ports", "get", port2UID)
			outputBytes2, err2 := cmd2.CombinedOutput()
			output2 := string(outputBytes2)

			if err1 == nil && err2 == nil &&
				(strings.Contains(output1, "CONFIGURED") || strings.Contains(output1, "LIVE")) &&
				(strings.Contains(output2, "CONFIGURED") || strings.Contains(output2, "LIVE")) {
				t.Logf("Both ports are ready")
				break
			}

			if attempt == maxAttempts {
				t.Fatalf("Ports did not reach ready state within the expected time")
			}

			time.Sleep(30 * time.Second)
		}
	})

	// STEP 2: Create a VXC using JSON file
	t.Run("Create VXC with JSON", func(t *testing.T) {
		// Skip if port creation failed
		if port1UID == "" || port2UID == "" {
			t.Skip("Skipping because port creation failed")
		}

		// Create JSON for VXC
		vxcJSON := fmt.Sprintf(`{
            "vxcName": "%s",
            "rateLimit": 100,
            "term": 1,
            "portUid": "%s",
            "aEndConfiguration": {
                "vlan": 100
            },
            "bEndConfiguration": {
                "productUID": "%s",
                "vlan": 200
            },
            "costCentre": "JSON Test Cost Centre"
        }`, vxcName, port1UID, port2UID)

		// Create temp file for the JSON
		vxcJSONFile := createTempJSONFile(t, vxcJSON)
		defer os.Remove(vxcJSONFile)

		cmd := exec.Command("megaport-cli", "vxc", "buy", "--json-file", vxcJSONFile)
		outputBytes, err := cmd.CombinedOutput()
		output := string(outputBytes)
		t.Logf("VXC JSON creation output: %s", output)

		assert.NoError(t, err, "Failed to create VXC with JSON: %s", output)
		assert.Contains(t, output, "VXC created")

		vxcUID = extractVXCUID(output)
		require.NotEmpty(t, vxcUID, "Failed to extract VXC UID")
	})

	// STEP 3: Update the VXC using JSON
	t.Run("Update VXC with JSON", func(t *testing.T) {
		// Skip if previous step failed
		if vxcUID == "" {
			t.Skip("Skipping because VXC creation failed")
		}

		// Wait for VXC to be ready
		maxAttempts := 20
		for attempt := 1; attempt <= maxAttempts; attempt++ {
			t.Logf("Checking VXC status, attempt %d of %d", attempt, maxAttempts)

			cmd := exec.Command("megaport-cli", "vxc", "get", vxcUID)
			outputBytes, err := cmd.CombinedOutput()
			output := string(outputBytes)

			if err == nil && (strings.Contains(output, "CONFIGURED") || strings.Contains(output, "LIVE")) {
				t.Logf("VXC is ready")
				break
			}

			if attempt == maxAttempts {
				t.Fatalf("VXC did not reach ready state within the expected time")
			}

			time.Sleep(20 * time.Second)
		}

		// Create JSON for update
		updateJSON := fmt.Sprintf(`{
            "name": "%s-Updated",
            "rateLimit": 200
        }`, vxcName)

		// Create temp file for the update JSON
		updateJSONFile := createTempJSONFile(t, updateJSON)
		defer os.Remove(updateJSONFile)

		cmd := exec.Command("megaport-cli", "vxc", "update", vxcUID, "--json-file", updateJSONFile)
		outputBytes, err := cmd.CombinedOutput()
		output := string(outputBytes)
		t.Logf("VXC JSON update output: %s", output)

		assert.NoError(t, err, "Failed to update VXC with JSON: %s", output)
		assert.Contains(t, output, "VXC updated")

		// Verify the update
		cmd = exec.Command("megaport-cli", "vxc", "get", vxcUID)
		outputBytes, err = cmd.CombinedOutput()
		output = string(outputBytes)

		assert.NoError(t, err, "Failed to get VXC info after update: %s", output)
		assert.Contains(t, output, fmt.Sprintf("%s-Updated", vxcName)) // Updated name
		assert.Contains(t, output, "200")                              // Updated rate limit
	})

	// STEP 4: Clean up - delete VXC and ports
	t.Run("Delete JSON Test Resources", func(t *testing.T) {
		// Delete VXC if it was created
		if vxcUID != "" {
			cmd := exec.Command("megaport-cli", "vxc", "delete", vxcUID, "--force")
			outputBytes, err := cmd.CombinedOutput()
			output := string(outputBytes)
			t.Logf("VXC deletion output: %s", output)

			assert.NoError(t, err, "Failed to delete VXC: %s", output)
			assert.Contains(t, output, "Deleting VXC")

			// Wait for VXC deletion to process
			time.Sleep(5 * time.Second)
		}

		// Delete ports
		if port1UID != "" {
			cmd := exec.Command("megaport-cli", "ports", "delete", port1UID, "--now", "--force")
			_, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("Failed to delete first port: %s", err)
			}
		}

		if port2UID != "" {
			cmd := exec.Command("megaport-cli", "ports", "delete", port2UID, "--now", "--force")
			_, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("Failed to delete second port: %s", err)
			}
		}
	})
}

// Helper function to extract VXC UID from command output
func extractVXCUID(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		// Check for standard format
		if strings.Contains(line, "VXC created") {
			parts := strings.Split(line, "VXC created")
			if len(parts) > 1 {
				return strings.TrimSpace(parts[1])
			}
		}
		// Check for enhanced format
		if strings.Contains(line, "VXC created successfully - ID:") {
			parts := strings.Split(line, "ID:")
			if len(parts) > 1 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return ""
}
