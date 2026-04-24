# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test Commands

```bash
# Build
go build -v

# Run all tests
go test -v ./...

# Run tests for a single package
go test -v ./internal/commands/ports

# Run a single test
go test -v ./internal/commands/ports -run TestFilterPorts

# Lint (uses golangci-lint with config in .golangci.yml)
golangci-lint run

# Format
gofmt -w .

# WASM build
GOOS=js GOARCH=wasm go build -tags js,wasm -o web/megaport.wasm .
```

## Architecture

This is a Go CLI application for managing Megaport network infrastructure resources (ports, VXCs, MCRs, MVEs). It has dual build targets: native CLI and WebAssembly (browser).

### Entry Points
- `main.go` — Native CLI entry point, calls `megaport.Execute()`
- `main_wasm.go` — WASM entry point for browser version
- `cmd/megaport/` — Root command setup, module registration, and platform-specific init

### Key Design Patterns

**Module Registry Pattern**: All command groups implement the `registry.Module` interface (`Name()` + `RegisterCommands()`). Modules are registered in `cmd/megaport/megaport_common.go:registerModules()` and connected to the root cobra command via `moduleRegistry.RegisterAll()`.

**Command Builder Pattern**: Commands are built using a fluent builder API (`internal/base/cmdbuilder/builder.go`). Every command should use `cmdbuilder.NewCommand()` with chained methods like `.WithFlag()`, `.WithRunFunc()`, `.WithExample()`, `.Build()`. Flag sets for each resource type are in separate files (e.g., `port_flagsets.go`, `vxc_flagsets.go`).

**Three Input Modes**: All "buy" and "update" commands support three input methods:
1. Interactive prompts (`--interactive`)
2. CLI flags (`--name "..." --term 12 ...`)
3. JSON input (`--json '{...}'` or `--json-file ./config.json`)

Use `WithConditionalRequirements()` to enforce required flags only when not using interactive/JSON mode.

**Login via `config.LoginFunc`**: API authentication is done through a package-level `config.LoginFunc` variable. Tests override this variable to inject mock clients.

### Package Structure

- `internal/commands/<resource>/` — Each resource type (ports, vxc, mcr, mve, etc.) is a package containing:
  - `<resource>.go` — Command definitions using CommandBuilder
  - `<resource>_actions.go` — Business logic (API calls, input processing)
  - `<resource>_module.go` — Module interface implementation
  - `<resource>_mock.go` — Mock service structs for testing
  - `<resource>_test.go` — Tests
  - `<resource>_prompts.go` — Interactive prompt logic
- `internal/base/cmdbuilder/` — Command builder and reusable flag sets
- `internal/base/output/` — Output formatters (table, JSON, CSV, XML) and terminal UI (spinners, colors, messages)
- `internal/base/registry/` — Module registration system
- `internal/utils/` — Shared utilities (output format helpers, prompt helpers)
- `internal/validation/` — Input validation functions

### Build Tags

Platform-specific code uses build tags:
- Native: `//go:build !js || !wasm`
- WASM: `//go:build js && wasm`

Files with `_wasm` suffix contain browser-specific implementations. When adding platform-specific behavior, create paired files with appropriate build tags.

### Testing Conventions

- Tests use `testify/assert` for assertions
- Mock services are struct-based (not interface-generated) — each command package defines its own mock (e.g., `MockPortService` in `ports_mock.go`)
- Tests override `config.LoginFunc` to inject mock clients, restoring the original in `defer`
- Table-driven tests are the standard pattern

### Output

All commands that display data should use the `output` package. Support `--output` flag with table (default), json, csv, xml formats. Use `output.PrintInfo`, `output.PrintError`, `output.PrintSuccess` for user messages.

### Dependencies

- `github.com/megaport/megaportgo` — Megaport API SDK
- `github.com/spf13/cobra` — CLI framework
- `github.com/jedib0t/go-pretty/v6` — Table rendering
- `github.com/stretchr/testify` — Test assertions
