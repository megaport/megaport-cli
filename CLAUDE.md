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


## Code Exploration Policy

Always use jCodemunch-MCP tools for code navigation. Never fall back to Read, Grep, Glob, or Bash for code exploration.
**Exception:** Use `Read` when you need to edit a file — the agent harness requires a `Read` before `Edit`/`Write` will succeed. Use jCodemunch tools to *find and understand* code, then `Read` only the specific file you're about to modify.

**Start any session:**
1. `resolve_repo { "path": "." }` — confirm the project is indexed. If not: `index_folder { "path": "." }`
2. `suggest_queries` — when the repo is unfamiliar

**Finding code:**
- symbol by name → `search_symbols` (add `kind=`, `language=`, `file_pattern=`, `decorator=` to narrow)
- decorator-aware queries → `search_symbols(decorator="X")` to find symbols with a specific decorator (e.g. `@property`, `@route`); combine with set-difference to find symbols *lacking* a decorator (e.g. "which endpoints lack CSRF protection?")
- string, comment, config value → `search_text` (supports regex, `context_lines`)
- database columns (dbt/SQLMesh) → `search_columns`

**Reading code:**
- before opening any file → `get_file_outline` first
- one or more symbols → `get_symbol_source` (single ID → flat object; array → batch)
- symbol + its imports → `get_context_bundle`
- specific line range only → `get_file_content` (last resort)

**Repo structure:**
- `get_repo_outline` → dirs, languages, symbol counts
- `get_file_tree` → file layout, filter with `path_prefix`

**Relationships & impact:**
- what imports this file → `find_importers`
- where is this name used → `find_references`
- is this identifier used anywhere → `check_references`
- file dependency graph → `get_dependency_graph`
- what breaks if I change X → `get_blast_radius`
- what symbols actually changed since last commit → `get_changed_symbols`
- find unreachable/dead code → `find_dead_code`
- class hierarchy → `get_class_hierarchy`

## Session-Aware Routing

**Opening move for any task:**
1. `plan_turn { "repo": "...", "query": "your task description", "model": "<your-model-id>" }` — get confidence + recommended files; the `model` parameter narrows the exposed tool list to match your capabilities at zero extra requests.
2. Obey the confidence level:
   - `high` → go directly to recommended symbols, max 2 supplementary reads
   - `medium` → explore recommended files, max 5 supplementary reads
   - `low` → the feature likely doesn't exist. Report the gap to the user. Do NOT search further hoping to find it.

**Interpreting search results:**
- If `search_symbols` returns `negative_evidence` with `verdict: "no_implementation_found"`:
  - Do NOT re-search with different terms hoping to find it
  - Do NOT assume a related file (e.g. auth middleware) implements the missing feature (e.g. CSRF)
  - DO report: "No existing implementation found for X. This would need to be created."
  - DO check `related_existing` files — they show what's nearby, not what exists
- If `verdict: "low_confidence_matches"`: examine the matches critically before assuming they implement the feature

**After editing files:**
- If PostToolUse hooks are installed (Claude Code only), edited files are auto-reindexed
- Otherwise, call `register_edit` with edited file paths to invalidate caches and keep the index fresh
- For bulk edits (5+ files), always use `register_edit` with all paths to batch-invalidate

**Token efficiency:**
- If `_meta` contains `budget_warning`: stop exploring and work with what you have
- If `auto_compacted: true` appears: results were automatically compressed due to turn budget
- Use `get_session_context` to check what you've already read — avoid re-reading the same files

## Model-Driven Tool Tiering

Your jcodemunch-mcp server narrows the exposed tool list based on the model you are running as. To avoid wasting requests on primitives when a composite would do, always include `model="<your-model-id>"` in your opening `plan_turn` call.

Replace `<your-model-id>` with your active model:
- Claude Opus variants → `claude-opus-4-7` (or any `claude-opus-*`)
- Claude Sonnet variants → `claude-sonnet-4-6`
- Claude Haiku variants → `claude-haiku-4-5`
- GPT-4o / GPT-5 / o1 / Llama → use the model id as printed by your runner

The `model=` parameter rides on the existing `plan_turn` call — it does **not** add a separate tool invocation. If `plan_turn` is not appropriate for a non-code task, call `announce_model(model="...")` once instead.
