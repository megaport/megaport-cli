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
test-integration:
	go test -tags integration -run '^TestIntegration_' -v -timeout 30m ./internal/commands/...

# Run only read-only integration tests — fast, no resources provisioned.
# The locations package has only read-only tests; core packages also hold
# provisioning lifecycle tests, so they are scoped to the ReadOnly-suffixed names.
test-integration-readonly:
	go test -tags integration -run '^TestIntegration_' -v -timeout 5m \
		./internal/commands/billing_market/... \
		./internal/commands/locations/... \
		./internal/commands/product/... \
		./internal/commands/status/... \
		./internal/commands/topology/...
	go test -tags integration -run 'TestIntegration_.*ReadOnly$$' -v -timeout 5m \
		./internal/commands/ix/... \
		./internal/commands/mcr/... \
		./internal/commands/mve/... \
		./internal/commands/ports/... \
		./internal/commands/vxc/...

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
