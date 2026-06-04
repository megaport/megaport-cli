.PHONY: build test test-cover test-integration test-integration-readonly lint fmt vet check clean wasm

# Build the CLI binary
build:
	go build -v -o megaport-cli .

# Run all tests
test:
	go test -v ./...

# Run tests with coverage report
test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	@rm -f coverage.out

# Run integration tests against staging API (requires credentials — see docs/INTEGRATION_TESTING.md)
# The `provisioning` tag also pulls in lifecycle tests that create and tear down
# real staging resources (e.g. the service key test provisions a port).
test-integration:
	go test -tags 'integration provisioning' -run '^TestIntegration_' -v -timeout 30m ./internal/commands/...

# Run only read-only integration tests — fast, no resources provisioned.
# Package list mirrors the integration-readonly CI job.
test-integration-readonly:
	go test -tags integration -run '^TestIntegration_' -v -timeout 5m ./internal/commands/locations/... ./internal/commands/partners/... ./internal/commands/servicekeys/... ./internal/commands/users/... ./internal/commands/managed_account/...

# Run linter
lint:
	golangci-lint run

# Format code
fmt:
	gofmt -w .

# Run go vet
vet:
	go vet ./...

# Run all checks (lint + test) — use this before committing
check: lint test

# Build WASM binary
wasm:
	GOOS=js GOARCH=wasm go build -trimpath -tags js,wasm -ldflags="-s -w" -o web/megaport.wasm .

# Clean build artifacts
clean:
	rm -f megaport-cli cover*.out coverage*.out web/megaport.wasm
