# Apply Command Integration Tests — Design

**Ticket:** ESD-1380 — Megaport CLI: Integration Tests for Apply Command

## Goal

Add staging integration tests for `megaport-cli apply`, covering dry-run
validation, full multi-resource provisioning lifecycle, and
`--rollback-on-failure`. Follow the existing integration-test framework rather
than inventing a new one.

## Context

`apply` (`internal/commands/apply/`) provisions resources sequentially in
dependency order (Ports → MCRs → MVEs → VXCs) from a YAML/JSON config, resolves
`{{.type.name}}` templates so later resources can reference earlier ones, and
supports `--dry-run` (validate only) and `--rollback-on-failure` (delete
created resources if a later one fails).

The repo already has a mature integration suite:

- `//go:build integration` tag.
- `testutil.RequireSharedIntegrationClient(t)` — process-wide staging client,
  parallel-safe via `sync.Once`.
- Per-resource lifecycle tests (ports, mcr, mve, vxc, ix, locations) that buy →
  assert via the SDK → clean up via `t.Cleanup`.
- The suite prefers reading state directly through the SDK over asserting on
  captured CLI output. `output.CaptureOutput` swaps global `os.Stdout`, so it
  races with spinner goroutines when tests run in parallel within the same
  package process; serial tests may still use it to validate user-facing
  messaging (the apply dry-run test does exactly this).
- `Makefile`: `test-integration` runs
  `go test -tags integration -run '^TestIntegration_' ./internal/commands/...`
  (apply is picked up by the glob automatically).
- `.github/workflows/integration-test.yml`: a provisioning job with an
  **explicit package list** (`ports`, `vxc`, `mcr`, `mve`) and a discovery
  guard that fails if no `TestIntegration_*` tests exist in the targeted
  packages.

## Design problem unique to apply

Two properties of apply drive the design:

1. **`ApplyConfig` returns only `error`, not the created UIDs.** The
   per-resource suites recover UIDs by hooking package-level `buyXFunc` seams;
   apply calls `client.XService.BuyX` directly, so no equivalent seam exists.
2. **Cleanup must survive partial failure.** The rollback test deliberately
   fails mid-run, and the happy path can orphan resources on a provisioning
   timeout. Cleanup cannot depend on apply succeeding.

Both point to the same solution: **name-based SDK discovery**. Each test uses a
unique name prefix; a `t.Cleanup` sweep lists ports/MCRs/MVEs/VXCs matching that
prefix and deletes them in reverse dependency order (VXCs → MVEs → MCRs →
Ports). Assertions also query by name. No production change, robust to partial
failure, consistent with the suite's "read state via SDK" approach.

Alternatives rejected:

- **Refactor `ApplyConfig` to return `[]ApplyResult`** — production change for
  testability, and still wouldn't help cleanup when apply fails before
  returning.
- **Add `buyXFunc` seams to apply** mirroring ports/vxc — four new production
  indirections, and still needs the name sweep as a partial-failure safety net.

## Test set

New file: `internal/commands/apply/apply_integration_test.go` (`//go:build
integration`, package `apply`).

All three tests are **serial** (no `t.Parallel()` — apply writes spinner output
to stdout; serial avoids the documented `CaptureOutput`/stdout-swap race). All
use `RequireSharedIntegrationClient`, a unique `CLI-Apply-Test-<id>` name
prefix, location ID 67 (the suite-wide staging location), and term 1.

### 1. `TestIntegration_ApplyDryRun`

Config: 2 ports + 1 port-to-port VXC using `{{.port.X}}` templates. Run with
`--dry-run`.

- Asserts `ApplyConfig` returns nil.
- Asserts **zero** resources created (SDK list by prefix is empty).
- Confirms template references resolve to valid / `skipped: requires
  provisioning` results.

Cheap and safe — no provisioning.

### 2. `TestIntegration_ApplyLifecycle`

Config: same 2 ports + 1 port-to-port VXC. Run with `--yes`.

- Real provisioning end to end.
- Asserts via SDK that both ports and the VXC exist by name.
- Asserts the VXC's two endpoints point at the two provisioned port UIDs
  (proves end-to-end template resolution).
- Prefix-sweep cleanup.

### 3. `TestIntegration_ApplyRollbackOnFailure`

Config: 1 valid port + 1 VXC whose b-end is a deliberately unresolved template
(`{{.port.DoesNotExist}}`). Run with `--yes --rollback-on-failure`.

- The port provisions for real; the VXC stage fails at template resolution →
  `handleFailure` → `doRollback` → real `DeletePort`.
- Asserts `ApplyConfig` returns an error.
- Asserts the port is gone / decommissioning (SDK list by prefix).
- Prefix-sweep as a safety net.

The unresolved-template trigger is deterministic and apply-specific, and does
not couple the test to server-side validation rules that could change. (A
rejected alternative — an out-of-range VLAN failing client-side
`ValidateVXCRequest` — also routes through `handleFailure` but couples the test
to validation bounds.)

## Shared helpers (same file)

- `applyIntegrationCmd(file string, dryRun, yes, rollback bool) *cobra.Command`
  — mirrors the unit test's `applyCmd` plus the rollback flag.
- `writeApplyConfig(t, yaml string) string` — writes the config to `t.TempDir()`
  and returns the path.
- `sweepByPrefix(t, prefix string)` — registered via `t.Cleanup`; lists and
  deletes all matching resources in reverse dependency order.
- Override `provisionPollInterval = 2 * time.Second` for the integration run to
  avoid 10s poll sleeps.

## CI wiring

In `.github/workflows/integration-test.yml`, add `./internal/commands/apply/...`
to **both** the discovery-guard package list and the provisioning-run package
list. The `Makefile` needs no change (its `./internal/commands/...` glob already
covers apply).

## Out of scope

- MCR/MVE resources in apply configs (ports + VXC is the cheapest config that
  exercises template resolution and the full provisioning path).
- JSON-format config files (YAML is sufficient; parsing is unit-tested).
- Exhaustive validation-error matrices (covered by unit tests).
