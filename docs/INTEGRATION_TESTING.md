# Integration Testing

Integration tests run against the Megaport staging API and verify that CLI commands work correctly end-to-end. They are separate from unit tests, which use mocks and run on every PR.

## Prerequisites

Set the following environment variables before running integration tests:

```bash
export MEGAPORT_ACCESS_KEY=<your-staging-access-key>
export MEGAPORT_SECRET_KEY=<your-staging-secret-key>
export MEGAPORT_ENVIRONMENT=staging
```

Staging credentials can be obtained from the Megaport staging portal. The `MEGAPORT_ENVIRONMENT=staging` value is enforced by the test helpers — tests will skip automatically if it is not set, preventing accidental runs against production.

## Running tests

```bash
# Read-only tests only — fast (< 5 min), no resources provisioned
make test-integration-readonly

# Full suite including provisioning lifecycle tests (~20–30 min)
make test-integration

# A single package
go test -tags integration -v -timeout 30m ./internal/commands/ports/...
```

## What gets created on staging

Provisioning lifecycle tests (ports, VXC, MCR, MVE, IX, NAT Gateway) create real resources on the staging account. All test resources are named with the prefix `CLI-Test-` for easy identification.

Resources are cleaned up automatically via `t.Cleanup()` at the end of each test, even when the test fails. However, if a test run is interrupted (e.g. `Ctrl+C`), cleanup may not run. In that case, log in to the staging portal and delete any resources prefixed with `CLI-Test-`.

## Build tag

All integration test files use the `//go:build integration` build tag:

```go
//go:build integration
// +build integration

package ports
```

Running `go test ./...` (without `-tags integration`) excludes these files entirely. They only compile and run when `-tags integration` is passed explicitly.

## CI

Integration tests run in CI via `.github/workflows/integration-test.yml`:

- **Read-only job**: runs nightly on `main`, tests `locations` (and `partners` once its integration test is written). Fast, no resource cost.
- **Provisioning job**: manual trigger only (`workflow_dispatch`). Runs lifecycle tests for ports, VXC, MCR, MVE, and additional resources as they are added.

## Adding a new integration test

1. Create `internal/commands/<resource>/<resource>_integration_test.go`
2. Set `//go:build integration` at the top and use `package <resource>` (not `package <resource>_test`)
3. Use `testutil.SetupIntegrationClient(t)` and `defer testutil.LoginWithClient(t, client)()` to authenticate
4. Call action functions directly and capture output with `output.CaptureOutput`
5. Use `t.Cleanup()` (not `defer`) for resource deletion so cleanup runs even on test failure
6. Add the package to the provisioning job in `.github/workflows/integration-test.yml`

See `internal/commands/locations/locations_integration_test.go` for a complete read-only example and `internal/commands/ports/ports_integration_test.go` for a full lifecycle example.
