.PHONY: build test test-cover test-cover-html test-integration test-integration-readonly e2e lint fmt vet check clean wasm

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

# Generate an HTML coverage report for local inspection (coverage.html kept on disk)
test-cover-html:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@rm -f coverage.out
	@echo "Coverage report written to coverage.html"

# Run integration tests against staging API (requires credentials — see docs/INTEGRATION_TESTING.md)
# The `provisioning` tag also pulls in lifecycle tests that create and tear down
# real staging resources (e.g. the service key test provisions a port).
test-integration:
	go test -tags 'integration provisioning' -run '^TestIntegration_' -v -timeout 30m ./internal/commands/...

# Run only read-only integration tests — fast, no resources provisioned.
# Package list mirrors the integration-readonly CI job.
test-integration-readonly:
	go test -tags integration -run '^TestIntegration_' -v -timeout 5m \
		./internal/commands/billing_market/... \
		./internal/commands/locations/... \
		./internal/commands/managed_account/... \
		./internal/commands/partners/... \
		./internal/commands/product/... \
		./internal/commands/servicekeys/... \
		./internal/commands/status/... \
		./internal/commands/topology/... \
		./internal/commands/users/...

# Run native-binary black-box e2e tests (built behind the `e2e` build tag).
# Placeholder: the harness and specs land in a later PR. Until then this compiles
# cleanly and reports no matching tests, so it is safe to wire into CI now.
e2e:
	go test -tags e2e -run '^TestE2E_' -v ./...

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
	rm -f megaport-cli cover*.out coverage*.out coverage.html web/megaport.wasm
