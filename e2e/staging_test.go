//go:build e2e

package e2e

import (
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This file holds the live-staging tier: black-box tests that drive the compiled
// binary against the real staging API and parse its stdout. They build on the
// hermetic harness in this package but, unlike the hermetic tests, forward the
// MEGAPORT_* credentials into the subprocess. Every case is read-only (only list
// commands) and runs in its own process, so the suite is safe under t.Parallel()
// and never creates, updates, or deletes anything on staging.
//
// Cases skip when credentials are absent, mirroring the SDK-side gate in
// testutil.SetupIntegrationClient, so the whole tier shares one "needs staging
// access" gate. locations list reaches a public endpoint and does not strictly
// require credentials, but it is gated with the rest for a uniform contract;
// partners list is the case that actually authenticates. The tests reach staging
// only through the binary and never import command internals or the SDK.

// requireStagingCreds skips the test unless both API credentials are present in
// the host environment, mirroring testutil.SetupIntegrationClient's gate.
func requireStagingCreds(t *testing.T) {
	t.Helper()
	if os.Getenv("MEGAPORT_ACCESS_KEY") == "" || os.Getenv("MEGAPORT_SECRET_KEY") == "" {
		t.Skip("MEGAPORT_ACCESS_KEY and MEGAPORT_SECRET_KEY required for live staging e2e tests")
	}
}

// stagingEnv is the environment spec forwarded into each CLI subprocess. The
// credentials are forwarded from the host by name, while MEGAPORT_ENVIRONMENT is
// pinned to an explicit, normalized value so the binary targets the intended
// environment instead of falling back to its production default.
func stagingEnv() []string {
	return []string{
		"MEGAPORT_ACCESS_KEY",
		"MEGAPORT_SECRET_KEY",
		"MEGAPORT_ENVIRONMENT=" + stagingEnvironment(),
	}
}

// stagingEnvironment resolves the target environment from the host
// MEGAPORT_ENVIRONMENT, failing safe to staging on an empty or unrecognized
// value. This mirrors testutil.IntegrationEnvironment so a typo cannot silently
// retarget the tests at production, which is where the CLI itself maps any
// value it does not recognize.
func stagingEnvironment() string {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("MEGAPORT_ENVIRONMENT"))) {
	case "production", "prod":
		return "production"
	case "development", "dev":
		return "development"
	default:
		return "staging"
	}
}

// TestE2E_Staging_LocationsListJSON drives `locations list --output json` against
// staging and asserts the binary renders a real, non-empty JSON array. The
// locations endpoint is public, so this case exercises reachability and rendering
// rather than authentication; TestE2E_Staging_PartnersListJSON covers the
// authenticated path.
func TestE2E_Staging_LocationsListJSON(t *testing.T) {
	requireStagingCreds(t)
	t.Parallel()

	res := RunWithEnv(t, stagingEnv(), "locations", "list", "--output", "json")

	require.Equalf(t, 0, res.Exit, "locations list should exit 0\nstdout: %s\nstderr: %s", res.Stdout, res.Stderr)

	var locations []json.RawMessage
	require.NoErrorf(t, json.Unmarshal([]byte(res.Stdout), &locations),
		"stdout should be a JSON array\nstdout: %s", res.Stdout)
	assert.NotEmpty(t, locations, "staging should always return at least one location")
}

// TestE2E_Staging_PartnersListJSON drives `partners list --output json`, which
// authenticates against staging before listing, and checks the decoded structure
// has the partner-port shape. This is the canonical authentication check for the
// tier.
func TestE2E_Staging_PartnersListJSON(t *testing.T) {
	requireStagingCreds(t)
	t.Parallel()

	res := RunWithEnv(t, stagingEnv(), "partners", "list", "--output", "json")

	require.Equalf(t, 0, res.Exit, "partners list should exit 0\nstdout: %s\nstderr: %s", res.Stdout, res.Stderr)

	// A populated UID proves at least one element carries real partner data rather
	// than the decode silently yielding a slice of empty objects.
	var partners []struct {
		ProductName string `json:"product_name"`
		UID         string `json:"uid"`
		CompanyName string `json:"company_name"`
		ConnectType string `json:"connect_type"`
	}
	require.NoErrorf(t, json.Unmarshal([]byte(res.Stdout), &partners),
		"stdout should decode into the partner shape\nstdout: %s", res.Stdout)
	require.NotEmpty(t, partners, "staging should always return at least one partner port")

	withUID := 0
	for _, p := range partners {
		if p.UID != "" {
			withUID++
		}
	}
	assert.Positivef(t, withUID, "at least one partner should carry a UID\nstdout: %s", res.Stdout)
}

// TestE2E_Staging_LocationsListFormats runs the same read command across every
// output format and asserts each exits 0 and produces output the format's parser
// accepts. Each format runs in its own subtest, hence its own process.
func TestE2E_Staging_LocationsListFormats(t *testing.T) {
	requireStagingCreds(t)
	t.Parallel()

	cases := []struct {
		format string
		check  func(t *testing.T, stdout string)
	}{
		{"json", checkParsesAsJSONArray},
		{"csv", checkParsesAsCSV},
		{"xml", checkParsesAsXML},
		{"table", checkNonEmptyTable},
	}

	for _, tc := range cases {
		t.Run(tc.format, func(t *testing.T) {
			t.Parallel()

			res := RunWithEnv(t, stagingEnv(), "locations", "list", "--output", tc.format)

			require.Equalf(t, 0, res.Exit, "locations list --output %s should exit 0\nstdout: %s\nstderr: %s",
				tc.format, res.Stdout, res.Stderr)
			tc.check(t, res.Stdout)
		})
	}
}

func checkParsesAsJSONArray(t *testing.T, stdout string) {
	t.Helper()
	var arr []json.RawMessage
	require.NoErrorf(t, json.Unmarshal([]byte(stdout), &arr), "stdout should be a JSON array\nstdout: %s", stdout)
	assert.NotEmpty(t, arr, "JSON array should be non-empty")
}

func checkParsesAsCSV(t *testing.T, stdout string) {
	t.Helper()
	records, err := csv.NewReader(strings.NewReader(stdout)).ReadAll()
	require.NoErrorf(t, err, "stdout should be valid CSV\nstdout: %s", stdout)
	// Require a header plus at least one data row. An empty result set still emits
	// a header-only CSV, so checking the header alone would pass on no data and be
	// weaker than the json/xml/table checks, which all require real content.
	require.Greaterf(t, len(records), 1, "CSV should have a header and at least one data row\nstdout: %s", stdout)
	assert.NotEmpty(t, records[0], "CSV header row should have columns")
}

func checkParsesAsXML(t *testing.T, stdout string) {
	t.Helper()
	var doc struct {
		XMLName xml.Name   `xml:"items"`
		Items   []struct{} `xml:"item"`
	}
	require.NoErrorf(t, xml.Unmarshal([]byte(stdout), &doc), "stdout should be valid XML\nstdout: %s", stdout)
	assert.NotEmpty(t, doc.Items, "XML should contain at least one <item>")
}

func checkNonEmptyTable(t *testing.T, stdout string) {
	t.Helper()
	require.NotEmpty(t, strings.TrimSpace(stdout), "table output should be non-empty")
	// A rendered table is a header plus at least one data row, so it always spans
	// more than one line; a single line would signal an empty or malformed table.
	assert.Greater(t, strings.Count(strings.TrimSpace(stdout), "\n"), 0, "table should span multiple lines")
}
