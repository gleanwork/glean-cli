# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.18.0](https://github.com/gleanwork/glean-cli/compare/v0.17.1...v0.18.0) (2026-07-27)


### Features

* **agents:** serve agents commands from the platform API by default ([#136](https://github.com/gleanwork/glean-cli/issues/136)) ([21737e2](https://github.com/gleanwork/glean-cli/commit/21737e22c7fc7abfc309d0a583030ba9d5bde836))
* **api:** route /api/* platform paths and send experimental header ([#140](https://github.com/gleanwork/glean-cli/issues/140)) ([b8a88cc](https://github.com/gleanwork/glean-cli/commit/b8a88ccae971c4ab88416a2b6dc97ccc9f4be737))
* **cmdutil:** teach Build the platform-first fallback protocol ([#135](https://github.com/gleanwork/glean-cli/issues/135)) ([0b8da87](https://github.com/gleanwork/glean-cli/commit/0b8da8790ea62bf8799351dca7fe87eba015d3c8))
* **platform:** platform-first routing with legacy fallback ([#132](https://github.com/gleanwork/glean-cli/issues/132)) ([4b9131a](https://github.com/gleanwork/glean-cli/commit/4b9131aff4ccaedacfd85b7580d067ed4125cdc8))
* **search:** platform search request builder and runner ([#133](https://github.com/gleanwork/glean-cli/issues/133)) ([046a998](https://github.com/gleanwork/glean-cli/commit/046a9987d34d294ed7198d3fe31c119435887b38))
* **search:** serve glean search from the platform API by default ([#134](https://github.com/gleanwork/glean-cli/issues/134)) ([bf4c747](https://github.com/gleanwork/glean-cli/commit/bf4c7472e5fc84db0cea9702316e5da6a71ca371))


### Dependencies

* bump github.com/gleanwork/api-client-go to v0.13.3 ([#131](https://github.com/gleanwork/glean-cli/issues/131)) ([aefae40](https://github.com/gleanwork/glean-cli/commit/aefae40ada7db81557a672bd06638aa1aa80d45a))
* **deps:** bump github.com/coreos/go-oidc/v3 from 3.19.0 to 3.20.0 ([#130](https://github.com/gleanwork/glean-cli/issues/130)) ([0951fdd](https://github.com/gleanwork/glean-cli/commit/0951fdd1e766b5821351e0f38f95a2d701b51a53))
* **deps:** bump github.com/gkampitakis/go-snaps from 0.5.22 to 0.5.23 ([#129](https://github.com/gleanwork/glean-cli/issues/129)) ([ab4b9eb](https://github.com/gleanwork/glean-cli/commit/ab4b9ebedfd1c2758227ad8d06195bb124b10f5e))
* **deps:** bump github.com/gleanwork/api-client-go ([#142](https://github.com/gleanwork/glean-cli/issues/142)) ([76feda1](https://github.com/gleanwork/glean-cli/commit/76feda1b33367d3085c411e174c46738fcd2e041))
* **deps:** bump golang.org/x/mod from 0.37.0 to 0.38.0 ([#128](https://github.com/gleanwork/glean-cli/issues/128)) ([c3bb409](https://github.com/gleanwork/glean-cli/commit/c3bb409b2fa4d64ddb70c71b4c3900c2963c1e0b))
* **deps:** bump golang.org/x/net ([#124](https://github.com/gleanwork/glean-cli/issues/124)) ([562f078](https://github.com/gleanwork/glean-cli/commit/562f07869f3d7d25ce6ba43c33f537230eed1d93))
* **deps:** bump golang.org/x/term from 0.44.0 to 0.45.0 ([#126](https://github.com/gleanwork/glean-cli/issues/126)) ([a85d09a](https://github.com/gleanwork/glean-cli/commit/a85d09a2787f784895b61081710f02f94394861e))

## [0.17.1](https://github.com/gleanwork/glean-cli/compare/v0.17.0...v0.17.1) (2026-06-29)


### Dependencies

* **deps:** bump github.com/coreos/go-oidc/v3 from 3.18.0 to 3.19.0 ([#121](https://github.com/gleanwork/glean-cli/issues/121)) ([914fcb9](https://github.com/gleanwork/glean-cli/commit/914fcb9256dbaaa192cf614c30ad743d837ee297))
* **deps:** bump github.com/gkampitakis/go-snaps from 0.5.21 to 0.5.22 ([#116](https://github.com/gleanwork/glean-cli/issues/116)) ([21841c1](https://github.com/gleanwork/glean-cli/commit/21841c1dd053f06906dbd397f935b6b333a17875))
* **deps:** bump github.com/gleanwork/api-client-go ([#117](https://github.com/gleanwork/glean-cli/issues/117)) ([5ddbc57](https://github.com/gleanwork/glean-cli/commit/5ddbc57567f20e4813740c3c7665e2496542fefb))
* **deps:** bump golang.org/x/mod from 0.35.0 to 0.37.0 ([#119](https://github.com/gleanwork/glean-cli/issues/119)) ([f961e9d](https://github.com/gleanwork/glean-cli/commit/f961e9d4bec5ee91efdf5953b691df208cf22c07))
* **deps:** bump golang.org/x/term from 0.42.0 to 0.44.0 ([#118](https://github.com/gleanwork/glean-cli/issues/118)) ([7e405b9](https://github.com/gleanwork/glean-cli/commit/7e405b9e499a3a6b95748d3bb9844b5a1ec20b56))


### Continuous Integration

* **deps:** bump actions/checkout from 6 to 7 ([#120](https://github.com/gleanwork/glean-cli/issues/120)) ([70b0735](https://github.com/gleanwork/glean-cli/commit/70b0735da0465643d83ecde55c86896c3b12e1e0))

## [0.17.0](https://github.com/gleanwork/glean-cli/compare/v0.16.1...v0.17.0) (2026-05-05)


### Features

* consolidate 17 Agent Skills into one skill with reference files ([#110](https://github.com/gleanwork/glean-cli/issues/110)) ([adf1913](https://github.com/gleanwork/glean-cli/commit/adf1913e22d302498bd2e8c1d2570f3b237e6c96))


### Continuous Integration

* **deps:** bump actions/create-github-app-token from 1 to 3 ([#111](https://github.com/gleanwork/glean-cli/issues/111)) ([8d4f10c](https://github.com/gleanwork/glean-cli/commit/8d4f10c3b9a2c30369d62913f4572f68a9d7e2d9))
* **deps:** bump googleapis/release-please-action from 4 to 5 ([#112](https://github.com/gleanwork/glean-cli/issues/112)) ([04358a2](https://github.com/gleanwork/glean-cli/commit/04358a2c920e897ec01e31b0106be669af7bca78))

## [0.16.1](https://github.com/gleanwork/glean-cli/compare/v0.16.0...v0.16.1) (2026-05-02)


### Bug Fixes

* **ci:** mint App token for release-please so tags trigger release.yml ([#108](https://github.com/gleanwork/glean-cli/issues/108)) ([d9184f8](https://github.com/gleanwork/glean-cli/commit/d9184f8fe2f9ef19b055843da077e6b78645497c))

## [0.16.0](https://github.com/gleanwork/glean-cli/compare/v0.15.0...v0.16.0) (2026-05-02)


### Features

* **ci:** automate releases via release-please ([#106](https://github.com/gleanwork/glean-cli/issues/106)) ([fd4b3c0](https://github.com/gleanwork/glean-cli/commit/fd4b3c05344f51da195534e5d2868fef85e56ef4))


### Continuous Integration

* mint homebrew-tap token via GitHub App instead of PAT ([#105](https://github.com/gleanwork/glean-cli/issues/105)) ([adbcb9c](https://github.com/gleanwork/glean-cli/commit/adbcb9c1a02475da3a4022bbddb5dd581c220c8b))

## [0.5.5] - 2026-03-17

### Fixed
- CI: skip `TestStateDir_FilePermissions` on Windows — Unix permission bits are not enforced on Windows
- Remove `glean version` subcommand in favour of the conventional `--version` flag

## [0.5.4] - 2026-03-17

### Added
- Update notifications: after each command, a background goroutine checks the GitHub releases API (cached for 24h in `~/.glean/update-check.json`) and prints a notice to stderr when a newer version is available

## [0.5.3] - 2026-03-17

### Added
- `--version` flag on the root command (via Cobra built-in)

## [0.5.2] - 2026-03-17

### Fixed
- Add `--dry-run` flag to `documents get-permissions`, `answers get`, and `shortcuts get` — these were inconsistently missing the flag vs their sibling subcommands

## [0.5.1] - 2026-03-17

### Added
- `User-Agent: glean-cli/<version>` header on all outbound HTTP requests (both SDK-routed and streaming chat), allowing Glean's backend to identify and attribute CLI traffic by version

## [0.5.0] - 2026-03-17

### Added
- Full release pipeline: GoReleaser with cosign signing, CycloneDX SBOM, Homebrew tap publishing
- Checksum verification in `install.sh`
- `SECURITY.md` with vulnerability disclosure process
- `--version` / `--help` flag improvements

### Fixed
- Release workflow now gates GoReleaser on tests and lint passing
- `glean chat --json --dry-run` now correctly includes `stream: true`
- All delete/remove subcommands now support `--dry-run`
- `documents get-by-facets`, `entities read-people`, `messages get` now support `--dry-run`
- README Quick Start uses correct full hostname format and includes auth as step 0
- `glean chat --timeout` help text corrected to reflect 60s default
- Error messages across namespace commands now include `--help` guidance

## [0.4.0] - 2026-03-14

### Added
- Full-screen interactive TUI as the default `glean` invocation (no args)
  - Streaming chat with live stage indicators (Searching / Reading / Writing)
  - Slash commands: `/mode auto|fast|advanced`, `/clear`, `/help`
  - `@filename` file attachment support
  - Session persistence with `--continue` flag
  - `ctrl+y` to copy last response
- `glean mcp` stdio MCP server exposing `glean_search`, `glean_chat`, `glean_schema`, `glean_people`
- `--fields` dot-path projection for search and namespace commands
- Agent skill files in `skills/`

## [0.3.0] - 2026-03-13

### Added
- 18 SDK namespace command groups: `activity`, `agents`, `announcements`, `answers`, `collections`, `documents`, `entities`, `insights`, `messages`, `pins`, `shortcuts`, `tools`, `verification`, plus core `search`, `chat`, `api`, `auth`, `schema`
- `--json` raw payload flag on all namespace commands
- `--output json|ndjson|text` on all commands
- `--dry-run` on all mutating commands
- `glean schema [command]` for machine-readable flag documentation

## [0.2.x] - 2025-2026

### Added
- OAuth PKCE + Dynamic Client Registration (`glean auth login`)
- Official Glean Go SDK (`github.com/gleanwork/api-client-go`) replacing hand-rolled HTTP client
- Shell completions: `glean completion bash|zsh|fish`
- Cross-platform builds (macOS, Linux, Windows — amd64 and arm64)

## [0.1.0] - 2025

### Added
- Initial release: `glean search` and `glean chat` commands
- API token authentication via environment variables
