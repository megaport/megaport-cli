# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [v0.5.5] — 2026-04-02

### Added
- Docker image for the native CLI binary, published to GitHub Container Registry
- Aggregate status dashboard (`megaport status`) showing all resources in one view
- Improved API error messages for 401, 403, 404, and 429 HTTP responses

### Fixed
- GoReleaser configuration updated for v1.26.2 compatibility (separate Docker build target, `builds` field syntax)

## [v0.5.4] — 2026-04-01

### Added
- `users` module: list, get, create, update, and delete users
- `--profile` global flag for per-command credential profile selection
- `validate` / dry-run subcommands for all resource types (port, lag, vxc, mcr, mve, ix)
- Confirmation prompt before buy operations
- `--no-wait` flag on all buy commands for non-blocking provisioning
- `lock` and `unlock` commands for MCR and MVE; `restore` for MVE
- VXC list command with `--status` and `--name-contains` filter flags
- Auto-disable colors and spinners when output is piped
- Test coverage reporting added to CI workflow

### Fixed
- VXC mock field names standardized to PascalCase for consistency

## [v0.5.3] — 2026-03-31

### Added
- `--quiet` / `-q` and `--verbose` / `-v` global flags

### Fixed
- Locations commands migrated to native LocationV3 structs
- Duplicate login and API calls removed from MCR prefix filter list updates
- Misleading AWS connect type validation error message corrected
- Duplicate flag definitions removed from partners list command
- Dead code and redundant branches removed

## [v0.5.2] — 2026-03-11

### Added
- XML output format via `--output xml`
- `list-countries`, `list-market-codes`, and `search` subcommands for `locations`
- Distinct exit codes: 2 = usage/validation, 3 = authentication, 4 = API error

### Fixed
- Date range validation for service key creation
- Output header tag casing standardized to Title Case across all resources
- `vxc update` now returns an error when called without flags or `--interactive`
- MCR delete now uses a confirm prompt for consistency

## [v0.5.1] — 2026-03-10

### Added
- Billing markets commands (`list-billing-markets`, `set-billing-market`)
- Token authentication support (`--auth-token`) for portal session tokens
- Full SDK flag parity: `--safe-delete`, `--is-approved`, `--vnic-index`, service key and location filters
- Thread-safe output format handling via atomic value

### Fixed
- Context timeout handling improvements across VXC, MCR, MVE, and port commands
- JSON output mode: spinner and progress messages routed to stderr for clean `jq` piping
- Extensive correctness fixes following code review

## [v0.5.0] — 2026-03-06

### Added
- `managed-account` commands: list, get, create, update, delete
- `ix` (Internet Exchange) commands: list, get, buy, update, delete with filtering
- VXC list command with client-side filtering

### Fixed
- VXC VLAN validation: correct upper bound (4094), support for untagged VLANs
- Azure partner port lookup in `getPartnerPortUID`
- Data race on output format global replaced with atomic value

## [v0.4.9] — 2026-02-26

### Added
- WebAssembly (WASM) browser version with session-based authentication
- Vue 3 + xterm.js frontend for the browser terminal
- Auto-environment detection for WASM builds based on hostname
- `setAuthToken` command for portal session authentication
- WASM support for MCR, MVE, and VXC commands

### Fixed
- Config file permissions hardened to 0600; recovery on file corruption
- Access keys masked in `config view` and `list-profiles` output
- Empty buy UID guard and nil status return handling

## [v0.4.8] — 2025-08-04

### Fixed
- Locations commands migrated from deprecated v2 API to v3
- Progress messages in JSON output mode now go to stderr (enables clean `jq` piping)
- GoReleaser `ldflags` updated to inject version into the binary at build time

## [v0.4.7] — 2025-07-31

### Fixed
- Environment variable credentials now take priority when `--env` is explicitly set

## [v0.4.6] — 2025-07-30

### Fixed
- MCR list command flag name inconsistency resolved

## [v0.4.5] — 2025-07-28

### Fixed
- GoReleaser `ldflags` version injection
- IBM name validation logic
- Updated megaportgo from v1.3.5 to v1.3.9

## [v0.4.4] — 2025-04-30

### Added
- `get-status` subcommand for port, MVE, MCR, and VXC resources

### Fixed
- Untagged VLAN now accepted in VXC flag validation
- JSON parsing in VXC input handling optimized

## [v0.4.3] — 2025-04-17

### Added
- Enhanced validation for VXC and MVE: VLAN ranges, IP/CIDR, partner config, untagged VLAN

### Fixed
- VXC type checking and validation corrections
- MCR and MVE validation consolidated into the `validation` package

## [v0.4.2] — 2025-04-15

### Added
- Resource tag support for VXC, port, and MCR (list and update subcommands)
- Support for higher-speed MCR configurations

## [v0.4.1] — 2025-04-14

### Added
- Colorized table output with Megaport branding for improved readability

## [v0.4.0] — 2025-04-11

### Changed
- Enhanced CLI prompting and feedback experience across interactive flows

## [v0.3.9] — 2025-04-10

### Added
- `mve list` command
- `mcr list` command
- `--show-inactive` flag for `ports list`

## [v0.3.8] — 2025-04-10

### Added
- Configuration management: named profiles, `config set`, `config get`, `list-profiles`, `import`, `export`

### Fixed
- golangci-lint CI configuration stabilized
- Table formatting updated for go-pretty library compatibility

## [v0.3.7] — 2025-04-09

### Added
- Table output now uses go-pretty with Megaport brand colors; `--no-color` renders plain ASCII tables

### Fixed
- Partners listing bug when filter flags were combined
- `golang.org/x/net` security vulnerability patched

## [v0.3.6] — 2025-04-09

### Added
- `docs` subcommand renders Markdown CLI reference in the terminal using Glamour

## [v0.3.1] – [v0.3.5] — 2025-04-05 to 2025-04-08

### Added
- Command builder pattern: all commands refactored to use `cmdbuilder.NewCommand()` fluent API
- Port, MCR, and MVE integration tests
- Improved output for update change display

### Fixed
- VXC flag sets for create and update
- MCR prefix filter list commands and buy interactive mode

## [v0.3.0] — 2025-04-04

### Added
- `docs generate` command for auto-generating CLI reference documentation
- Interactive mode for partner port selection in VXC buy flow
- VXC: enhanced buy and update with Azure, GCP, Oracle partner port lookup
- `--version` output based on git tag
- Shell `completion` command

### Changed
- CLI reorganized into modular command packages (`internal/commands/<resource>/`)

## [v0.2.0] – [v0.2.9] — 2025-03-05 to 2025-04-03

### Added
- Color output across all resource types with `--no-color` flag
- VXC full support: buy, update, list with interactive, flag, and JSON input modes
- Partner port selection for AWS, Azure, GCP, Oracle, and IBM VXC types
- MVE images and sizes listing with filtering
- MCR prefix filter list management commands
- `--env` flag for environment selection (`production`, `staging`, `development`)

### Fixed
- VXC partner port and VLAN prompt handling
- Documentation improvements across all command groups

## [v0.1.0] – [v0.1.9] — 2025-02-21 to 2025-03-04

### Added
- Initial release with ports, VXCs, MCRs, MVEs, partner ports, locations, and service keys
- Port lifecycle management: buy, update, delete, list, get, and LAG port support
- MCR lifecycle management: buy, update, delete, and prefix filter list creation
- Interactive, flag-based, and JSON input modes for buy and update commands
- CSV, JSON, and table output formats via `--output` flag
- Shell completion support
- Credential support via environment variables (`MEGAPORT_ACCESS_KEY`, `MEGAPORT_SECRET_KEY`, `MEGAPORT_ENVIRONMENT`)
- Automated CI with golangci-lint and Go test suite

[Unreleased]: https://github.com/megaport/megaport-cli/compare/v0.5.5...HEAD
[v0.5.5]: https://github.com/megaport/megaport-cli/compare/v0.5.4...v0.5.5
[v0.5.4]: https://github.com/megaport/megaport-cli/compare/v0.5.3...v0.5.4
[v0.5.3]: https://github.com/megaport/megaport-cli/compare/v0.5.2...v0.5.3
[v0.5.2]: https://github.com/megaport/megaport-cli/compare/v0.5.1...v0.5.2
[v0.5.1]: https://github.com/megaport/megaport-cli/compare/v0.5.0...v0.5.1
[v0.5.0]: https://github.com/megaport/megaport-cli/compare/v0.4.9...v0.5.0
[v0.4.9]: https://github.com/megaport/megaport-cli/compare/v0.4.8...v0.4.9
[v0.4.8]: https://github.com/megaport/megaport-cli/compare/v0.4.7...v0.4.8
[v0.4.7]: https://github.com/megaport/megaport-cli/compare/v0.4.6...v0.4.7
[v0.4.6]: https://github.com/megaport/megaport-cli/compare/v0.4.5...v0.4.6
[v0.4.5]: https://github.com/megaport/megaport-cli/compare/v0.4.4...v0.4.5
[v0.4.4]: https://github.com/megaport/megaport-cli/compare/v0.4.3...v0.4.4
[v0.4.3]: https://github.com/megaport/megaport-cli/compare/v0.4.2...v0.4.3
[v0.4.2]: https://github.com/megaport/megaport-cli/compare/v0.4.1...v0.4.2
[v0.4.1]: https://github.com/megaport/megaport-cli/compare/v0.4.0...v0.4.1
[v0.4.0]: https://github.com/megaport/megaport-cli/compare/v0.3.9...v0.4.0
[v0.3.9]: https://github.com/megaport/megaport-cli/compare/v0.3.8...v0.3.9
[v0.3.8]: https://github.com/megaport/megaport-cli/compare/v0.3.7...v0.3.8
[v0.3.7]: https://github.com/megaport/megaport-cli/compare/v0.3.6...v0.3.7
[v0.3.6]: https://github.com/megaport/megaport-cli/compare/v0.3.5...v0.3.6
[v0.3.5]: https://github.com/megaport/megaport-cli/compare/v0.3.4...v0.3.5
[v0.3.4]: https://github.com/megaport/megaport-cli/compare/v0.3.3...v0.3.4
[v0.3.3]: https://github.com/megaport/megaport-cli/compare/v0.3.2...v0.3.3
[v0.3.2]: https://github.com/megaport/megaport-cli/compare/v0.3.1...v0.3.2
[v0.3.1]: https://github.com/megaport/megaport-cli/compare/v0.3.0...v0.3.1
[v0.3.0]: https://github.com/megaport/megaport-cli/compare/v0.2.9...v0.3.0
[v0.2.9]: https://github.com/megaport/megaport-cli/compare/v0.2.8...v0.2.9
[v0.2.8]: https://github.com/megaport/megaport-cli/compare/v0.2.7...v0.2.8
[v0.2.7]: https://github.com/megaport/megaport-cli/compare/v0.2.6...v0.2.7
[v0.2.6]: https://github.com/megaport/megaport-cli/compare/v0.2.5...v0.2.6
[v0.2.5]: https://github.com/megaport/megaport-cli/compare/v0.2.4...v0.2.5
[v0.2.4]: https://github.com/megaport/megaport-cli/compare/v0.2.3...v0.2.4
[v0.2.3]: https://github.com/megaport/megaport-cli/compare/v0.2.2...v0.2.3
[v0.2.2]: https://github.com/megaport/megaport-cli/compare/v0.2.1...v0.2.2
[v0.2.1]: https://github.com/megaport/megaport-cli/compare/v0.2.0...v0.2.1
[v0.2.0]: https://github.com/megaport/megaport-cli/compare/v0.1.9...v0.2.0
[v0.1.9]: https://github.com/megaport/megaport-cli/compare/v0.1.8...v0.1.9
[v0.1.8]: https://github.com/megaport/megaport-cli/compare/v0.1.7...v0.1.8
[v0.1.7]: https://github.com/megaport/megaport-cli/compare/v0.1.6...v0.1.7
[v0.1.6]: https://github.com/megaport/megaport-cli/compare/v0.1.5...v0.1.6
[v0.1.5]: https://github.com/megaport/megaport-cli/compare/v0.1.4...v0.1.5
[v0.1.4]: https://github.com/megaport/megaport-cli/compare/v0.1.3...v0.1.4
[v0.1.3]: https://github.com/megaport/megaport-cli/compare/v0.1.2...v0.1.3
[v0.1.2]: https://github.com/megaport/megaport-cli/compare/v0.1.1...v0.1.2
[v0.1.1]: https://github.com/megaport/megaport-cli/compare/v0.1.0...v0.1.1
[v0.1.0]: https://github.com/megaport/megaport-cli/releases/tag/v0.1.0
