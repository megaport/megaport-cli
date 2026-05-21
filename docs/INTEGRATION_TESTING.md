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

## Adding a new integration test

1. Create `internal/commands/<resource>/<resource>_integration_test.go`
2. Set `//go:build integration` at the top and use `package <resource>` (not `package <resource>_test`)
3. Authenticate using one of the helpers below (see "Authentication helpers")
4. Call action functions directly and capture output with `output.CaptureOutput`
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
