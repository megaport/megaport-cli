//go:build integration && provisioning

package servicekeys

import (
	"context"
	crypto_rand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/testutil"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// integrationLocationID is the staging data center used to provision the
// throwaway port a service key needs as its product. ID 67 is the canonical
// staging example location reused across the CLI's other integration tests.
const integrationLocationID = 67

func uniqueSuffix(t *testing.T) string {
	t.Helper()
	buf := make([]byte, 4)
	_, err := crypto_rand.Read(buf)
	require.NoError(t, err, "failed to read crypto/rand entropy")
	return hex.EncodeToString(buf)
}

// findPortUIDByName looks up the port with the given name and returns its UID,
// with ok=false if no live (non-decommissioned) port matches. It returns the
// list error rather than aborting, so it is safe to call from a t.Cleanup
// callback, where a FailNow would mask the original failure.
func findPortUIDByName(client *megaport.Client, name string) (uid string, ok bool, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	ports, err := client.PortService.ListPorts(ctx)
	if err != nil {
		return "", false, err
	}
	for _, p := range ports {
		if p == nil || p.Name != name || p.ProvisioningStatus == "DECOMMISSIONED" || p.ProvisioningStatus == "CANCELLED" {
			continue
		}
		return p.UID, true, nil
	}
	return "", false, nil
}

// provisionPortForServiceKey buys a 1G port on staging via the SDK and registers
// its teardown. A service key must reference an existing product (port), so the
// lifecycle test provisions a real one and deletes it in t.Cleanup even on failure.
func provisionPortForServiceKey(t *testing.T, client *megaport.Client) string {
	t.Helper()
	portName := fmt.Sprintf("CLI-Test-SvcKey-Port-%s", uniqueSuffix(t))

	// Safety net registered before BuyPort: if the buy partially succeeds (e.g.
	// the port is created but provisioning times out and BuyPort returns an
	// error), this still removes the port. It resolves the UID by the unique
	// name at cleanup time and is best-effort, so it is harmless when the port
	// was never created or is already gone.
	t.Cleanup(func() {
		uid, ok, err := findPortUIDByName(client, portName)
		if err != nil {
			t.Logf("cleanup: could not list ports to find %s: %v", portName, err)
			return
		}
		if !ok {
			return
		}
		delCtx, delCancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer delCancel()
		if _, err := client.PortService.DeletePort(delCtx, &megaport.DeletePortRequest{
			PortID:    uid,
			DeleteNow: true,
		}); err != nil {
			t.Errorf("cleanup: failed to delete port %s (%s): %v", uid, portName, err)
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Minute)
	defer cancel()

	resp, err := client.PortService.BuyPort(ctx, &megaport.BuyPortRequest{
		Name:                  portName,
		Term:                  1,
		PortSpeed:             1000,
		LocationId:            integrationLocationID,
		MarketPlaceVisibility: false,
		WaitForProvision:      true,
		WaitForTime:           10 * time.Minute,
	})
	require.NoError(t, err, "failed to provision port for service key test")
	require.NotEmpty(t, resp.TechnicalServiceUIDs, "buy port returned no technical service UIDs")
	portUID := resp.TechnicalServiceUIDs[0]

	return portUID
}

// serviceKeyUIDForProduct finds the key just created against productUID by its
// unique description. CreateServiceKey prints the UID but does not return it, so
// the test recovers it from the SDK list filtered by product.
func serviceKeyUIDForProduct(t *testing.T, client *megaport.Client, productUID, description string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	resp, err := client.ServiceKeyService.ListServiceKeys(ctx, &megaport.ListServiceKeysRequest{ProductUID: &productUID})
	require.NoError(t, err, "SDK ListServiceKeys failed")
	require.NotNil(t, resp)
	for _, k := range resp.ServiceKeys {
		if k != nil && k.Description == description {
			require.NotEmpty(t, k.Key, "service key UID should not be empty")
			return k.Key
		}
	}
	t.Fatalf("created service key with description %q not found for product %s", description, productUID)
	return ""
}

func serviceKeyFromSDK(t *testing.T, client *megaport.Client, keyUID string) *megaport.ServiceKey {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	sk, err := client.ServiceKeyService.GetServiceKey(ctx, keyUID)
	require.NoErrorf(t, err, "SDK GetServiceKey failed for %s", keyUID)
	require.NotNil(t, sk)
	return sk
}

func newCreateServiceKeyCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "create"}
	cmd.Flags().String("product-uid", "", "")
	cmd.Flags().Int("product-id", 0, "")
	cmd.Flags().Bool("single-use", false, "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("start-date", "", "")
	cmd.Flags().String("end-date", "", "")
	cmd.Flags().Int("max-speed", 0, "")
	cmd.Flags().Bool("active", false, "")
	cmd.Flags().Bool("pre-approved", false, "")
	cmd.Flags().Int("vlan", 0, "")
	return cmd
}

func newUpdateServiceKeyCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "update"}
	cmd.Flags().String("product-uid", "", "")
	cmd.Flags().Int("product-id", 0, "")
	cmd.Flags().Bool("single-use", false, "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().Bool("active", false, "")
	return cmd
}

// TestIntegration_ServiceKeyLifecycle provisions a port, then exercises the full
// create/get/update path of the service key CLI actions against it. Only the port
// is torn down in t.Cleanup; service keys have no delete API and are left attached
// to the (deleted) port. This test carries the extra `provisioning` build tag so
// the nightly read-only job (which builds only `-tags integration`) never runs it;
// it runs in the manual provisioning job.
func TestIntegration_ServiceKeyLifecycle(t *testing.T) {
	client := testutil.SetupIntegrationClient(t)
	defer testutil.LoginWithClient(t, client)()

	portUID := provisionPortForServiceKey(t, client)
	t.Logf("provisioned port %s for service key lifecycle", portUID)

	description := fmt.Sprintf("CLI-Test-Key-%s", uniqueSuffix(t))

	createCmd := newCreateServiceKeyCmd()
	require.NoError(t, createCmd.Flags().Set("product-uid", portUID))
	require.NoError(t, createCmd.Flags().Set("description", description))
	require.NoError(t, createCmd.Flags().Set("max-speed", "1000"))

	require.NoError(t, CreateServiceKey(createCmd, nil, true), "CreateServiceKey failed")

	keyUID := serviceKeyUIDForProduct(t, client, portUID, description)
	t.Logf("created service key %s", keyUID)

	getCmd := &cobra.Command{Use: "get"}
	getOut := output.CaptureOutput(func() {
		require.NoError(t, GetServiceKey(getCmd, []string{keyUID}, true, "json"))
	})

	var got []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(getOut), &got), "get output should be valid JSON")
	require.NotEmpty(t, got)
	assert.Equal(t, keyUID, got[0]["key_uid"])
	assert.Equal(t, portUID, got[0]["product_uid"])
	assert.Equal(t, description, got[0]["description"])

	updateCmd := newUpdateServiceKeyCmd()
	require.NoError(t, updateCmd.Flags().Set("active", "true"))
	require.NoError(t, UpdateServiceKey(updateCmd, []string{keyUID}, true), "UpdateServiceKey failed")

	updated := serviceKeyFromSDK(t, client, keyUID)
	assert.True(t, updated.Active, "service key should be active after update")
}
