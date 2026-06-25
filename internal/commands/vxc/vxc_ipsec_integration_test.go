//go:build integration

package vxc

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/megaport/megaport-cli/internal/commands/mcr"
	"github.com/megaport/megaport-cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// findTransitPartnerUID returns the product UID of a VXC-permitted TRANSIT
// (internet) partner megaport in the same region as the MCR. The IPsec tunnel
// egresses over the internet, so the VXC B-End is a TRANSIT port, mirroring the
// terraform provider's IPsec acceptance test. A Transit VXC must stay within one
// region, so the partner must share the MCR's metro: it prefers one co-located
// with the MCR, then any in the same metro. Skips when none is available.
func findTransitPartnerUID(t *testing.T, mcrLocationID int) string {
	t.Helper()
	sdkClient := testutil.SharedIntegrationClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	mcrLoc, err := sdkClient.LocationService.GetLocationByIDV3(ctx, mcrLocationID)
	require.NoErrorf(t, err, "GetLocationByIDV3(%d) failed", mcrLocationID)

	partners, err := sdkClient.PartnerService.ListPartnerMegaports(ctx)
	require.NoError(t, err, "ListPartnerMegaports failed")

	transit, err := sdkClient.PartnerService.FilterPartnerMegaportByConnectType(ctx, partners, "TRANSIT", true)
	if err != nil || len(transit) == 0 {
		t.Skipf("no TRANSIT partner megaport available on staging for IPsec tunnel B-End: %v", err)
	}

	// Co-located is cheapest (no per-partner location lookups) and satisfies the
	// same-region rule outright.
	for _, p := range transit {
		if p.LocationId == mcrLocationID {
			return p.ProductUID
		}
	}
	for _, p := range transit {
		loc, err := sdkClient.LocationService.GetLocationByIDV3(ctx, p.LocationId)
		if err == nil && loc.Metro == mcrLoc.Metro {
			return p.ProductUID
		}
	}
	t.Skipf("no TRANSIT partner in metro %q (location %d) for a same-region IPsec VXC", mcrLoc.Metro, mcrLocationID)
	return ""
}

// buyMCRWithIPsecAndGetUID buys an MCR with an IPsec add-on enabled at purchase
// (via --ipsec-tunnel-count) and returns its UID. BuyMCR blocks until the MCR is
// LIVE, so the add-on is provisioned before a tunnel VXC is attached. Cleanup is
// registered before the UID is asserted so a billable MCR is never leaked.
func buyMCRWithIPsecAndGetUID(t *testing.T, mcrName string, tunnelCount int) string {
	t.Helper()
	cmd := buildMCRCmd(t, mcrName, integrationLocationID)
	require.NoError(t, cmd.Flags().Set("ipsec-tunnel-count", fmt.Sprintf("%d", tunnelCount)))
	require.NoErrorf(t, mcr.BuyMCR(cmd, nil, true), "BuyMCR with IPsec add-on failed for %q", mcrName)

	uid, ok := mcr.IntegrationBuyMCRUID(mcrName)
	if ok {
		registerMCRCleanup(t, uid)
	}
	require.Truef(t, ok, "no MCR buy response captured for %q", mcrName)
	return uid
}

// TestIntegration_VXCMCRIPsecTunnelLifecycle provisions an MCR with an IPsec
// add-on and a VXC whose A-End vrouter config declares a subInterface carrying
// the tunnel source IP plus an ipSecTunnel interface with a single
// ipSecTunnelOptions object (one tunnel per ipSecTunnel interface). It exercises
// the ESD-1538 path (parsePartnerConfigFromJSON -> parseVRouterConfig ->
// parseIPsecTunnelOptions) and proves the live staging API accepts the
// single-object tunnel shape introduced in megaportgo v1.14.1.
//
// The B-End is a TRANSIT internet port so the tunnel egresses over the internet.
// Peer addresses are illustrative (RFC 5737 / link-local); the tunnel is never
// expected to come up, the test only asserts the order is accepted and the VXC
// provisions.
func TestIntegration_VXCMCRIPsecTunnelLifecycle(t *testing.T) {
	t.Parallel()
	testutil.RequireStagingForProvisioning(t)
	testutil.RequireSharedIntegrationClient(t)

	id := generateUniqueID(t)
	mcrName := fmt.Sprintf("CLI-Test-IPsec-MCR-%s", id)
	vxcName := fmt.Sprintf("CLI-Test-IPsec-VXC-%s", id)

	transitUID := findTransitPartnerUID(t, integrationLocationID)
	t.Logf("Using TRANSIT partner (B-End): %s", transitUID)

	mcrUID := buyMCRWithIPsecAndGetUID(t, mcrName, 10)
	t.Logf("Created MCR (A-End) with IPsec add-on: %s", mcrUID)

	vrouterConfig := `{
		"connectType": "VROUTER",
		"interfaces": [
			{
				"interfaceType": "subInterface",
				"ipAddresses": ["169.254.100.1/30"]
			},
			{
				"interfaceType": "ipSecTunnel",
				"ipSecTunnelOptions": {
					"sourceIpAddress": "169.254.100.1",
					"destinationIpAddress": "203.0.113.10",
					"preSharedKey": "cli-integration-psk",
					"phase1Lifetime": 28800,
					"phase2Lifetime": 3600
				}
			}
		]
	}`

	vxcCmd := buildVXCCmd(t, vxcName, mcrUID, transitUID, 100)
	require.NoError(t, vxcCmd.Flags().Set("a-end-partner-config", vrouterConfig))

	vxcUID := runBuyVXC(t, vxcCmd, vxcName)
	t.Logf("Created VXC (MCR IPsec tunnel A-End): %s", vxcUID)

	vxc := vxcFromSDK(t, vxcUID)
	assert.Equal(t, vxcUID, vxc.UID)
	assert.Equal(t, vxcName, vxc.Name)
	assert.Equal(t, 100, vxc.RateLimit)
	assert.Equal(t, mcrUID, vxc.AEndConfiguration.UID)
	assert.Equal(t, transitUID, vxc.BEndConfiguration.UID)
}
