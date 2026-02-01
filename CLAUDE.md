# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Test Commands

```bash
# Build native CLI
go build -o megaport-cli .

# Build WASM version
GOOS=js GOARCH=wasm go build -tags js,wasm -o web/megaport.wasm .

# Run all tests
go test ./...

# Run tests for a specific package
go test -v ./internal/commands/ports

# Run a single test
go test -run TestFunctionName ./path/to/package

# Lint
golangci-lint run

# Generate CLI documentation
./megaport-cli generate-docs ./docs
```

## Architecture Overview

This is a Go CLI tool for the Megaport API with dual build targets: native CLI and WebAssembly (browser).

### Key Directories

- `cmd/megaport/` - Entry points. Native (`megaport.go`) and WASM (`megaport_wasm.go`) versions
- `internal/base/cmdbuilder/` - Command builder pattern for declarative command creation
- `internal/base/output/` - Output formatters (table, JSON, CSV, XML)
- `internal/commands/` - Command modules (ports, vxc, mcr, mve, locations, partners, servicekeys, config)
- `internal/validation/` - Input validation for each resource type
- `internal/utils/` - Shared utilities including interactive prompts
- `web/` - JavaScript/WASM runtime for browser version
- `frontend-integration/` - Vue 3 component library for WASM integration

### Build Tag Separation

- Native CLI: `!js && !wasm` tags
- WASM version: `js && wasm` tags
- WASM excludes: config, completion, generate-docs, version commands

### Module System

Each command group is a `Module` implementing `RegisterCommands()`. Modules are registered in `cmd/megaport/modules.go` (native) or `modules_wasm.go` (WASM).

### Command Builder Pattern

Commands use a fluent builder API in `internal/base/cmdbuilder/`:

```go
cmdbuilder.NewCommand("action", "Description").
    WithLongDesc("...").
    WithExample("...").
    WithColorAwareRunFunc(ActionFunc).
    WithPortCreationFlags().  // Reusable flag sets
    WithRootCmd(rootCmd).
    Build()
```

### Configuration

- Native: Profile-based config in `~/.megaport/config.json` or env vars (`MEGAPORT_ACCESS_KEY`, `MEGAPORT_SECRET_KEY`, `MEGAPORT_ENVIRONMENT`)
- WASM: Session-based authentication via web UI

## Adding New Commands

1. Create module in `internal/commands/<resource>/`
2. Implement `<resource>_module.go` with `Module` interface
3. Create `<resource>_actions.go` for business logic
4. Add validation in `internal/validation/`
5. Register in `cmd/megaport/modules.go` (and `modules_wasm.go` if applicable)
6. Add tests in `<resource>_test.go`

## Key Dependencies

- `github.com/spf13/cobra` - CLI framework
- `github.com/megaport/megaportgo` - Megaport API client
- `github.com/jedib0t/go-pretty/v6` - Table output rendering
- `github.com/stretchr/testify` - Testing assertions
