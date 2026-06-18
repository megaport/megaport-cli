# Apply Integration Tests Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add staging integration tests for `megaport-cli apply` covering dry-run validation, full multi-resource provisioning lifecycle, and `--rollback-on-failure`.

**Architecture:** A single `//go:build integration` test file in package `apply` that drives `ApplyConfig` against the staging API via `testutil.RequireSharedIntegrationClient`. Because `ApplyConfig` returns only `error` (not created UIDs), tests use unique name prefixes and discover/clean up resources by name through the SDK (`ListPorts` + `GetPort` → `AssociatedVXCs`). Tests run serially (no `t.Parallel()`) to avoid the documented `CaptureOutput`/stdout-swap race.

**Tech Stack:** Go, `//go:build integration` tag, `testify`, cobra, `megaportgo` SDK v1.13.0.

**Spec:** `docs/superpowers/specs/2026-06-09-apply-integration-tests-design.md`

**Ticket:** ESD-1380

---

## File Structure

- **Create** `internal/commands/apply/apply_integration_test.go` — all three tests plus helpers (command builder, config writer, SDK discovery, cleanup sweep). One file, mirrors the per-resource suites (`ports_integration_test.go`, `vxc_integration_test.go`).
- **Modify** `.github/workflows/integration-test.yml` — add `./internal/commands/apply/...` to the discovery-guard list and the provisioning-run command.

### Notes on verification model

These are integration tests gated on staging credentials. The local "red/green" loop is:
- **Compiles + is discovered:** `go test -tags integration -list '^TestIntegration_' ./internal/commands/apply/...` (works without credentials — without `MEGAPORT_ACCESS_KEY`/`MEGAPORT_SECRET_KEY`, each test calls `t.Skip`).
- **Actually passes against staging:** requires credentials; run in the final task.

Per-task verification uses the `-list` form. Each task ends with a commit.

---

### Task 1: Scaffold the integration test file (build tag, imports, helpers)

**Files:**
- Create: `internal/commands/apply/apply_integration_test.go`

- [ ] **Step 1: Create the file with the build tag, imports, constants, and helpers**

```go
//go:build integration

package apply

import (
	"context"
	crypto_rand "crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/testutil"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// integrationLocationID is the staging data center used for all apply lifecycle
// tests. Location 67 matches the canonical example used across the CLI test
// suite and supports the 1G port speed these tests exercise.
const integrationLocationID = 67

const (
	cleanupStatusTimeout  = 90 * time.Second
	cleanupStatusInterval = 2 * time.Second
)

func init() {
	// Integration runs poll real staging status; 2s keeps lifecycle tests
	// responsive without the 10s production default. Only compiled under the
	// integration build tag, so unit tests keep the default.
	provisionPollInterval = 2 * time.Second
}

func generateUniqueID(t *testing.T) string {
	t.Helper()
	buf := make([]byte, 4)
	_, err := crypto_rand.Read(buf)
	require.NoError(t, err, "failed to read crypto/rand entropy")
	return hex.EncodeToString(buf)
}

// applyIntegrationCmd builds a cobra.Command carrying the flags ApplyConfig reads.
func applyIntegrationCmd(file string, dryRun, yes, rollback bool) *cobra.Command {
	cmd := &cobra.Command{Use: "apply"}
	cmd.Flags().StringP("file", "f", "", "")
	cmd.Flags().Bool("dry-run", false, "")
	cmd.Flags().BoolP("yes", "y", false, "")
	cmd.Flags().Bool("rollback-on-failure", false, "")
	_ = cmd.Flags().Set("file", file)
	if dryRun {
		_ = cmd.Flags().Set("dry-run", "true")
	}
	if yes {
		_ = cmd.Flags().Set("yes", "true")
	}
	if rollback {
		_ = cmd.Flags().Set("rollback-on-failure", "true")
	}
	return cmd
}

// writeApplyConfig writes a YAML config to a temp file and returns its path.
func writeApplyConfig(t *testing.T, yaml string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "apply.yaml")
	require.NoError(t, os.WriteFile(path, []byte(yaml), 0o600))
	return path
}
```

- [ ] **Step 2: Verify it compiles and is parsed under the integration tag**

Run: `go vet -tags integration ./internal/commands/apply/...`
Expected: PASS (no output). The file references `provisionPollInterval` (defined in `apply_actions.go`) and the helpers; some are unused until later tasks, but `go vet` does not fail on unused package-level funcs. If `go vet` complains about unused imports, that is expected to be resolved by Task 2 which uses them — to keep this task green, the imports `context`, `strings`, `time`, `output`, `testutil`, `megaport`, `assert` are used by later tasks; if the compiler errors on unused imports now, proceed to Step 3 which adds the discovery helpers in the same commit.

> Note: Go fails compilation on unused imports. To keep Task 1 self-contained and green, **combine Task 1 and Task 2 into a single commit** — write the file scaffold and the discovery/cleanup helpers (Task 2) together, then verify. The split below is for reading clarity; commit them together.

- [ ] **Step 3: Commit (together with Task 2 — see note)**

Deferred to Task 2's commit.

---

### Task 2: SDK discovery + cleanup sweep helpers

**Files:**
- Modify: `internal/commands/apply/apply_integration_test.go`

- [ ] **Step 1: Append the discovery and cleanup helpers**

```go
// portsByPrefix lists staging ports whose name begins with prefix.
func portsByPrefix(t *testing.T, prefix string) []*megaport.Port {
	t.Helper()
	client := testutil.SharedIntegrationClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	all, err := client.PortService.ListPorts(ctx)
	require.NoError(t, err, "SDK ListPorts failed")
	var out []*megaport.Port
	for _, p := range all {
		if p != nil && strings.HasPrefix(p.Name, prefix) {
			out = append(out, p)
		}
	}
	return out
}

// portByExactName returns the port with the given exact name from the set
// matching prefix, or nil if not present.
func portByExactName(t *testing.T, prefix, name string) *megaport.Port {
	t.Helper()
	for _, p := range portsByPrefix(t, prefix) {
		if p.Name == name {
			return p
		}
	}
	return nil
}

// portFromSDK reads a single port via GetPort (which populates AssociatedVXCs).
func portFromSDK(t *testing.T, uid string) *megaport.Port {
	t.Helper()
	client := testutil.SharedIntegrationClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	p, err := client.PortService.GetPort(ctx, uid)
	require.NoErrorf(t, err, "SDK GetPort failed for %s", uid)
	require.NotNilf(t, p, "SDK GetPort returned nil for %s", uid)
	return p
}

// waitForDecommission polls status until it contains "DECOMMISSION" (covering
// DECOMMISSIONING and DECOMMISSIONED), the resource is gone, or the timeout
// elapses. Best-effort: logs and returns rather than failing, so a slow
// teardown does not mask the test result.
func waitForDecommission(t *testing.T, kind, uid string, status func(context.Context) (string, error)) {
	t.Helper()
	deadline := time.Now().Add(cleanupStatusTimeout)
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		s, err := status(ctx)
		cancel()
		if err != nil {
			t.Logf("cleanup: get %s %s after delete: %v (may already be gone)", kind, uid, err)
			return
		}
		if strings.Contains(s, "DECOMMISSION") {
			return
		}
		if time.Now().After(deadline) {
			t.Logf("cleanup: %s %s did not reach DECOMMISSION within %s (last status %q)", kind, uid, cleanupStatusTimeout, s)
			return
		}
		time.Sleep(cleanupStatusInterval)
	}
}

// registerSweepCleanup schedules a best-effort teardown of every port whose name
// begins with prefix and any VXC attached to those ports. VXCs are deleted
// first and polled to DECOMMISSIONING before ports are deleted, because staging
// rejects deleting a port with a live VXC attached. Runs even when the test
// fails, preventing orphaned billing resources.
func registerSweepCleanup(t *testing.T, prefix string) {
	t.Helper()
	t.Cleanup(func() {
		client := testutil.SharedIntegrationClient(t)

		matched := portsByPrefix(t, prefix)

		// 1. Delete attached VXCs first (dedup across both ports of a VXC).
		seenVXC := map[string]bool{}
		for _, p := range matched {
			full := portFromSDK(t, p.UID)
			for _, vxc := range full.AssociatedVXCs {
				if vxc == nil || vxc.UID == "" || seenVXC[vxc.UID] {
					continue
				}
				seenVXC[vxc.UID] = true
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				err := client.VXCService.DeleteVXC(ctx, vxc.UID, &megaport.DeleteVXCRequest{DeleteNow: true})
				cancel()
				if err != nil {
					t.Logf("cleanup: delete VXC %s: %v", vxc.UID, err)
				}
			}
		}
		for uid := range seenVXC {
			waitForDecommission(t, "VXC", uid, func(ctx context.Context) (string, error) {
				vxc, err := client.VXCService.GetVXC(ctx, uid)
				if err != nil {
					return "", err
				}
				if vxc == nil {
					return "DECOMMISSIONED", nil
				}
				return vxc.ProvisioningStatus, nil
			})
		}

		// 2. Delete ports.
		for _, p := range matched {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			_, err := client.PortService.DeletePort(ctx, &megaport.DeletePortRequest{PortID: p.UID, DeleteNow: true})
			cancel()
			if err != nil {
				t.Logf("cleanup: delete port %s: %v", p.UID, err)
			}
		}
	})
}
```

- [ ] **Step 2: Verify the file compiles and lists (no tests yet, so no tests discovered)**

Run: `go vet -tags integration ./internal/commands/apply/...`
Expected: PASS. Some helpers (`applyIntegrationCmd`, `writeApplyConfig`, `portByExactName`, `output`, `assert`, `fmt`) are still unused until Tasks 3-5. Go does NOT error on unused functions, but DOES error on unused imports.

If the compiler reports unused imports `fmt`, `output`, or `assert` at this point: that is expected. To keep this commit green, add a temporary `var _ = fmt.Sprintf` style reference is NOT needed — instead, **proceed directly into Task 3 within the same commit** (Task 3 uses `fmt`, `output`, and `assert`). Treat Tasks 2 and 3 as the same commit if the unused-import error appears.

- [ ] **Step 3: Commit (Tasks 1-2, optionally through 3 if needed for imports)**

```bash
git add internal/commands/apply/apply_integration_test.go
git commit -m "test(ESD-1380): scaffold apply integration test helpers"
```

---

### Task 3: `TestIntegration_ApplyDryRun`

**Files:**
- Modify: `internal/commands/apply/apply_integration_test.go`

- [ ] **Step 1: Append the dry-run test**

```go
// twoPortVXCConfig returns a YAML apply config with two ports and one
// port-to-port VXC whose endpoints reference the ports via {{.port.<name>}}
// templates. Names are derived from prefix so a single sweep cleans them up.
func twoPortVXCConfig(prefix string) (yaml, portAName, portBName, vxcName string) {
	portAName = prefix + "-PortA"
	portBName = prefix + "-PortB"
	vxcName = prefix + "-VXC"
	yaml = fmt.Sprintf(`ports:
  - name: %s
    location_id: %d
    speed: 1000
    term: 1
    marketplace_visibility: false
  - name: %s
    location_id: %d
    speed: 1000
    term: 1
    marketplace_visibility: false
vxcs:
  - name: %s
    rate_limit: 100
    term: 1
    a_end:
      product_uid: "{{.port.%s}}"
    b_end:
      product_uid: "{{.port.%s}}"
`, portAName, integrationLocationID, portBName, integrationLocationID, vxcName, portAName, portBName)
	return yaml, portAName, portBName, vxcName
}

func TestIntegration_ApplyDryRun(t *testing.T) {
	testutil.RequireSharedIntegrationClient(t)

	prefix := fmt.Sprintf("CLI-Apply-Test-%s", generateUniqueID(t))
	registerSweepCleanup(t, prefix) // safety net; dry-run should create nothing

	yaml, _, _, _ := twoPortVXCConfig(prefix)
	cfgPath := writeApplyConfig(t, yaml)
	cmd := applyIntegrationCmd(cfgPath, true /*dryRun*/, true /*yes*/, false /*rollback*/)

	out := output.CaptureOutput(func() {
		require.NoError(t, ApplyConfig(cmd, nil, true, "table"), "dry-run apply should succeed")
	})

	// Validation results are reported; the VXC's template endpoints cannot be
	// validated server-side without provisioning, so it is reported as skipped.
	assert.Contains(t, out, "valid", "dry-run should report port validation results")
	assert.Contains(t, out, "skipped: requires provisioning", "templated VXC should be skipped in dry-run")

	// Crucially, nothing was provisioned.
	created := portsByPrefix(t, prefix)
	assert.Emptyf(t, created, "dry-run must not provision any ports; found %d", len(created))
}
```

- [ ] **Step 2: Verify it compiles and is discovered**

Run: `go test -tags integration -list '^TestIntegration_' ./internal/commands/apply/...`
Expected: output includes `TestIntegration_ApplyDryRun`.

- [ ] **Step 3: (If staging credentials available) run it**

Run: `MEGAPORT_ENVIRONMENT=staging go test -tags integration -run '^TestIntegration_ApplyDryRun$' -v ./internal/commands/apply/...`
Expected: PASS (or SKIP if `MEGAPORT_ACCESS_KEY`/`MEGAPORT_SECRET_KEY` unset).

- [ ] **Step 4: Commit**

```bash
git add internal/commands/apply/apply_integration_test.go
git commit -m "test(ESD-1380): add apply dry-run integration test"
```

---

### Task 4: `TestIntegration_ApplyLifecycle`

**Files:**
- Modify: `internal/commands/apply/apply_integration_test.go`

- [ ] **Step 1: Append the lifecycle test**

```go
// vxcOnPortByName polls GetPort(portUID).AssociatedVXCs for a VXC with the given
// name and returns its UID. The VXC is attached immediately after provisioning,
// but the association can lag a poll or two, so this retries briefly.
func vxcOnPortByName(t *testing.T, portUID, vxcName string) string {
	t.Helper()
	deadline := time.Now().Add(60 * time.Second)
	for {
		p := portFromSDK(t, portUID)
		for _, vxc := range p.AssociatedVXCs {
			if vxc != nil && vxc.Name == vxcName {
				return vxc.UID
			}
		}
		if time.Now().After(deadline) {
			t.Fatalf("VXC %q not found on port %s within timeout", vxcName, portUID)
		}
		time.Sleep(2 * time.Second)
	}
}

func TestIntegration_ApplyLifecycle(t *testing.T) {
	testutil.RequireSharedIntegrationClient(t)

	prefix := fmt.Sprintf("CLI-Apply-Test-%s", generateUniqueID(t))
	registerSweepCleanup(t, prefix)

	yaml, portAName, portBName, vxcName := twoPortVXCConfig(prefix)
	cfgPath := writeApplyConfig(t, yaml)
	cmd := applyIntegrationCmd(cfgPath, false /*dryRun*/, true /*yes*/, false /*rollback*/)

	require.NoError(t, ApplyConfig(cmd, nil, true, "table"), "apply should provision all resources")

	// Both ports exist.
	portA := portByExactName(t, prefix, portAName)
	require.NotNilf(t, portA, "port %q not found after apply", portAName)
	portB := portByExactName(t, prefix, portBName)
	require.NotNilf(t, portB, "port %q not found after apply", portBName)
	assert.NotEmpty(t, portA.UID)
	assert.NotEmpty(t, portB.UID)
	assert.NotEmpty(t, portA.ProvisioningStatus)

	// The VXC exists, attached to port A, with endpoints wired to the two
	// provisioned port UIDs — proving end-to-end {{.port.X}} template resolution.
	vxcUID := vxcOnPortByName(t, portA.UID, vxcName)
	client := testutil.SharedIntegrationClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	vxc, err := client.VXCService.GetVXC(ctx, vxcUID)
	require.NoError(t, err, "SDK GetVXC failed")
	require.NotNil(t, vxc)
	assert.Equal(t, vxcName, vxc.Name)
	assert.Equal(t, portA.UID, vxc.AEndConfiguration.UID, "VXC a_end should resolve to port A")
	assert.Equal(t, portB.UID, vxc.BEndConfiguration.UID, "VXC b_end should resolve to port B")
}
```

- [ ] **Step 2: Verify it compiles and is discovered**

Run: `go test -tags integration -list '^TestIntegration_' ./internal/commands/apply/...`
Expected: output includes `TestIntegration_ApplyLifecycle`.

- [ ] **Step 3: (If staging credentials available) run it**

Run: `MEGAPORT_ENVIRONMENT=staging go test -tags integration -run '^TestIntegration_ApplyLifecycle$' -v -timeout 20m ./internal/commands/apply/...`
Expected: PASS (provisions 2 ports + 1 VXC, asserts, then cleans up). Allow several minutes.

- [ ] **Step 4: Commit**

```bash
git add internal/commands/apply/apply_integration_test.go
git commit -m "test(ESD-1380): add apply provisioning lifecycle integration test"
```

---

### Task 5: `TestIntegration_ApplyRollbackOnFailure`

**Files:**
- Modify: `internal/commands/apply/apply_integration_test.go`

- [ ] **Step 1: Append the rollback test**

```go
func TestIntegration_ApplyRollbackOnFailure(t *testing.T) {
	testutil.RequireSharedIntegrationClient(t)

	prefix := fmt.Sprintf("CLI-Apply-Test-%s", generateUniqueID(t))
	registerSweepCleanup(t, prefix) // safety net if rollback itself fails

	portName := prefix + "-Port"
	vxcName := prefix + "-VXC"
	// One valid port, then a VXC whose b_end references a non-existent resource.
	// The port provisions for real; the VXC stage fails at template resolution,
	// triggering rollback that must delete the just-created port.
	yaml := fmt.Sprintf(`ports:
  - name: %s
    location_id: %d
    speed: 1000
    term: 1
    marketplace_visibility: false
vxcs:
  - name: %s
    rate_limit: 100
    term: 1
    a_end:
      product_uid: "{{.port.%s}}"
    b_end:
      product_uid: "{{.port.DoesNotExist}}"
`, portName, integrationLocationID, vxcName, portName)

	cfgPath := writeApplyConfig(t, yaml)
	cmd := applyIntegrationCmd(cfgPath, false /*dryRun*/, true /*yes*/, true /*rollback*/)

	err := ApplyConfig(cmd, nil, true, "table")
	require.Error(t, err, "apply should fail on the unresolved VXC template")

	// The port created before the failure must be rolled back: gone or
	// decommissioning.
	deadline := time.Now().Add(cleanupStatusTimeout)
	for {
		p := portByExactName(t, prefix, portName)
		if p == nil {
			break // rolled back and already gone
		}
		if strings.Contains(p.ProvisioningStatus, "DECOMMISSION") {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("port %q was not rolled back within %s (status %q)", portName, cleanupStatusTimeout, p.ProvisioningStatus)
		}
		time.Sleep(cleanupStatusInterval)
	}
}
```

- [ ] **Step 2: Verify it compiles and is discovered**

Run: `go test -tags integration -list '^TestIntegration_' ./internal/commands/apply/...`
Expected: output includes all three: `TestIntegration_ApplyDryRun`, `TestIntegration_ApplyLifecycle`, `TestIntegration_ApplyRollbackOnFailure`.

- [ ] **Step 3: (If staging credentials available) run it**

Run: `MEGAPORT_ENVIRONMENT=staging go test -tags integration -run '^TestIntegration_ApplyRollbackOnFailure$' -v -timeout 20m ./internal/commands/apply/...`
Expected: PASS (provisions a port, fails the VXC, rolls back the port).

- [ ] **Step 4: Commit**

```bash
git add internal/commands/apply/apply_integration_test.go
git commit -m "test(ESD-1380): add apply rollback-on-failure integration test"
```

---

### Task 6: Wire apply into the CI provisioning job

**Files:**
- Modify: `.github/workflows/integration-test.yml:44-48` (discovery `-list`) and `:62` (run command)

- [ ] **Step 1: Add the apply package to the discovery guard**

In the `Verify provisioning integration tests are discovered` step, add the apply package to the `go test -list` invocation so the list reads:

```yaml
          if ! list_output=$(go test -tags integration -list '^TestIntegration_' \
            ./internal/commands/ports/... \
            ./internal/commands/vxc/... \
            ./internal/commands/mcr/... \
            ./internal/commands/mve/... \
            ./internal/commands/apply/... 2>&1); then
```

- [ ] **Step 2: Add the apply package to the provisioning run command**

Change the `Run provisioning integration tests` step's command to:

```yaml
        run: go test -tags integration -run '^TestIntegration_' -v -timeout 30m ./internal/commands/ports/... ./internal/commands/vxc/... ./internal/commands/mcr/... ./internal/commands/mve/... ./internal/commands/apply/...
```

- [ ] **Step 3: Verify the workflow YAML is well-formed**

Run: `python3 -c "import yaml,sys; yaml.safe_load(open('.github/workflows/integration-test.yml')); print('ok')"`
Expected: `ok`

- [ ] **Step 4: Commit**

```bash
git add .github/workflows/integration-test.yml
git commit -m "ci(ESD-1380): run apply integration tests in provisioning job"
```

---

### Task 7: Full verification and lint

**Files:** none (verification only)

- [ ] **Step 1: Lint the new test file**

Run: `golangci-lint run --build-tags integration ./internal/commands/apply/...`
Expected: no findings. Fix any reported issues inline (e.g. unused helper — every helper added here is referenced by a test).

- [ ] **Step 2: Confirm the non-integration build is unaffected**

Run: `go build -v ./... && go test ./internal/commands/apply/...`
Expected: PASS — the integration file is excluded without the tag, existing unit tests still pass.

- [ ] **Step 3: (If staging credentials available) run the full apply integration suite**

Run: `MEGAPORT_ENVIRONMENT=staging go test -tags integration -run '^TestIntegration_' -v -timeout 30m ./internal/commands/apply/...`
Expected: all three tests PASS, no orphaned resources left on staging (cleanup sweeps run).

- [ ] **Step 4: Final commit if lint required changes**

```bash
git add -A
git commit -m "test(ESD-1380): lint fixes for apply integration tests"
```

---

## Self-Review Notes

- **Spec coverage:** dry-run (Task 3), lifecycle (Task 4), rollback-on-failure (Task 5), name-based discovery + sweep cleanup (Task 2), `provisionPollInterval` override (Task 1), CI wiring (Task 6). All spec sections mapped.
- **Type consistency:** `applyIntegrationCmd`, `writeApplyConfig`, `twoPortVXCConfig`, `portsByPrefix`, `portByExactName`, `portFromSDK`, `vxcOnPortByName`, `waitForDecommission`, `registerSweepCleanup`, `generateUniqueID` used with consistent signatures across tasks. SDK calls match v1.13.0: `DeletePortRequest{PortID, DeleteNow}`, `DeleteVXCRequest{DeleteNow}`, `VXCService.DeleteVXC(ctx, uid, req)`, `Port.AssociatedVXCs`, `VXC.AEndConfiguration.UID`.
- **Unused-import caveat:** flagged in Tasks 1-2 — commit the scaffold and helpers together (and fold Task 3 in if `fmt`/`output`/`assert` are still unused) so no commit leaves an uncompilable file.
