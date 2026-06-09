# Integration Testing

Integration tests run against the Megaport staging API and verify that CLI commands work correctly end-to-end. They are separate from unit tests, which use mocks and run on every PR.

## Prerequisites

Set the following environment variables before running integration tests:

```bash
export MEGAPORT_ACCESS_KEY=<your-staging-access-key>
export MEGAPORT_SECRET_KEY=<your-staging-secret-key>
export MEGAPORT_ENVIRONMENT=staging
```

Staging credentials can be obtained from the Megaport staging portal. The staging environment is hardcoded in the test helper (`testutil.SetupIntegrationClient`) — it is not possible to accidentally target production through the test suite. `MEGAPORT_ENVIRONMENT=staging` is passed to the CI runner for consistency with local usage, but is not read by the test helper itself.

## Running tests

```bash
# Read-only tests only — fast (< 5 min), no resources provisioned
make test-integration-readonly

# Full suite including any provisioning lifecycle tests (~20–30 min)
# Currently only read-only tests exist; provisioning tests will be added incrementally.
make test-integration

# A single package
go test -tags integration -run '^TestIntegration_' -v -timeout 30m ./internal/commands/ports/...
```

## What gets created on staging

Once provisioning lifecycle tests are written (ports, VXC, MCR, MVE, IX, NAT Gateway), they will create real resources on the staging account. All test resources will be named with the prefix `CLI-Test-` for easy identification.

Resources will be cleaned up automatically via `t.Cleanup()` at the end of each test, even when the test fails. However, if a test run is interrupted (e.g. `Ctrl+C`), cleanup may not run. In that case, log in to the staging portal and delete any resources prefixed with `CLI-Test-`.

Currently, only read-only integration tests exist (`locations`). No resources are provisioned.

## Build tag

All integration test files use the `//go:build integration` build tag:

```go
//go:build integration

package ports
```

Running `go test ./...` (without `-tags integration`) excludes these files entirely. They only compile and run when `-tags integration` is passed explicitly.

## CI

Integration tests run in CI via `.github/workflows/integration-test.yml`:

- **Read-only job**: runs nightly on `main` and on manual trigger, tests `locations` (and additional packages as read-only integration tests are written). Fast, no resource cost.
- **Provisioning job**: manual trigger only (`workflow_dispatch`). Runs lifecycle tests for ports, VXC, MCR, MVE, and additional resources as they are added.

### MCR Looking Glass tests

`internal/commands/mcr/mcr_looking_glass_integration_test.go` covers the four read-only Looking Glass diagnostics (`ip-routes`, `bgp-routes`, `bgp-sessions`, `bgp-neighbor-routes`). They are read-only but live in the `mcr` package alongside the provisioning lifecycle tests, so they run under the manual provisioning job (`mcr` is already listed there) and never run nightly. They are deliberately kept out of the nightly read-only job: that job runs `-run '^TestIntegration_'` across whole packages, and adding `mcr` there would also run the MCR provisioning lifecycle tests every night.

These tests need an MCR whose Looking Glass reports at least one BGP session in the UP state. They discover one automatically by listing MCRs and probing each one's BGP-sessions endpoint; the first MCR with an active session is used and the result is cached per run. If none qualifies, all four tests `t.Skip` with a clear message rather than failing.

Staging prerequisite: at the time of writing, the `/lookingGlass/*` endpoints return 404 for every MCR on staging, so these tests skip by design there. The test helper targets staging, so to exercise the assertions an MCR with an active BGP peering session must exist on staging; the discovery step then finds it automatically.

## Adding a new integration test

1. Create `internal/commands/<resource>/<resource>_integration_test.go`
2. Set `//go:build integration` at the top and use `package <resource>` (not `package <resource>_test`)
3. Authenticate using one of the helpers below (see "Authentication helpers")
4. Call action functions directly. For parallel tests, read state via `testutil.SharedIntegrationClient(t)` rather than `output.CaptureOutput` (see "Output capture and parallelism")
5. Use `t.Cleanup()` for resource deletion (e.g. deleting staging ports/VXCs) so cleanup runs even on test failure; `defer` is fine for non-resource cleanup like restoring login state
6. Add the package to the provisioning job in `.github/workflows/integration-test.yml`

See `internal/commands/locations/locations_integration_test.go` for a serial read-only example and `internal/commands/ports/ports_integration_test.go` for a parallel provisioning-lifecycle example.

### Authentication helpers

Two helpers in `internal/testutil` handle staging authentication. Pick based on whether your tests use `t.Parallel()`.

**Serial tests** (no `t.Parallel()`): use `testutil.SetupIntegrationClient` plus `testutil.LoginWithClient`. The login override is saved on entry and restored on cleanup:

```go
func TestIntegration_Foo(t *testing.T) {
    client := testutil.SetupIntegrationClient(t)
    defer testutil.LoginWithClient(t, client)()
    // ...
}
```

**Parallel tests** (`t.Parallel()`): use `testutil.RequireSharedIntegrationClient`. It authorises once per process via `sync.Once` and installs the login override exactly once; it never restores. All callers share a single authorised `*megaport.Client`, which is safe because they all target the same staging environment.

```go
func TestIntegration_Foo(t *testing.T) {
    t.Parallel()
    testutil.RequireSharedIntegrationClient(t)
    // ...
}
```

Do not combine `LoginWithClient` with `t.Parallel()`: its save/restore pattern races when concurrent tests each capture and restore `config.LoginFunc`, leaving the global pointing at a stale closure.

### Output capture and parallelism

`output.CaptureOutput` swaps the process-wide `os.Stdout` for a tmpfile and holds a mutex for the duration of the wrapped function. Combined with `t.Parallel()`, that has two failure modes:

1. **Serialisation**: the lock serialises every captured call, so parallel goroutines block whenever any one of them is inside `CaptureOutput`. Long blocking actions (e.g. `BuyPort` with `WaitForProvision=true`) erase the wall-time benefit of `t.Parallel()`.
2. **Stdout races**: action code uses spinner goroutines that write to `os.Stdout` asynchronously. While goroutine A holds the capture mutex with `os.Stdout = tmpA`, goroutine B's spinner can write into `tmpA` or into a just-closed file descriptor after A's defer runs. Captures come back polluted or empty.

For parallel lifecycle tests, prefer this pattern:

- Drive the side-effecting action (`BuyPort`, `UpdatePort`, `DeletePort`, …) without wrapping it in `CaptureOutput`. Its real stdout output interleaves with other tests' output on the terminal, which is harmless.
- If the action returns data via the SDK response (e.g. `BuyPortResponse.TechnicalServiceUIDs`), hook the underlying package-level function variable (`buyPortFunc` etc.) in an `init()` and store the response in a `sync.Map` keyed by something the test controls (request name, UID). See `internal/commands/ports/ports_integration_test.go` for an example.
- Read state for assertions via `testutil.SharedIntegrationClient(t)` rather than scraping `GetPort`/`ListPorts` stdout. This bypasses `CaptureOutput` entirely and is parallel-safe.

Serial read-only tests (e.g. `locations`) can continue to use `output.CaptureOutput` directly because there is no concurrent goroutine to race with.
