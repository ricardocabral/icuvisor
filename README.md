<p align="center">
  <img src="docs/brand/logo-wordmark.png" alt="icuvisor" width="540" />
</p>

[![Go Reference](https://pkg.go.dev/badge/github.com/ricardocabral/icuvisor.svg)](https://pkg.go.dev/github.com/ricardocabral/icuvisor)
[![CI](https://github.com/ricardocabral/icuvisor/actions/workflows/ci.yml/badge.svg)](https://github.com/ricardocabral/icuvisor/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/ricardocabral/icuvisor?sort=semver)](https://github.com/ricardocabral/icuvisor/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/ricardocabral/icuvisor)](go.mod)
[![Conventional Commits](https://img.shields.io/badge/Conventional%20Commits-1.0.0-blue.svg)](https://www.conventionalcommits.org)
<!-- [![Go Report Card](https://goreportcard.com/badge/github.com/ricardocabral/icuvisor)](https://goreportcard.com/report/github.com/ricardocabral/icuvisor) -->
<!-- [![codecov](https://codecov.io/gh/ricardocabral/icuvisor/branch/main/graph/badge.svg)](https://codecov.io/gh/ricardocabral/icuvisor) -->
<!-- [![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/ricardocabral/icuvisor/badge)](https://securityscorecards.dev/viewer/?uri=github.com/ricardocabral/icuvisor) -->


> icuvisor is an open-source, local-first [Model Context Protocol](https://modelcontextprotocol.io) server for [intervals.icu](https://intervals.icu), shipped as a single signed Go binary so athletes and coaches can talk to their training data from Claude, ChatGPT, Pi, Cursor, and other MCP-compatible clients. Your intervals.icu API key stays in the OS keychain, not in an icuvisor cloud service or MCP tool argument. There is no icuvisor-hosted account, onboarding credit, SaaS quota, or subscription gate. Usage limits from your AI client or model provider, GitHub/package downloads, and intervals.icu account are separate. End-user docs live at <https://icuvisor.app>.

## For users

The fastest path on Linux, macOS (without Homebrew), WSL, and CI is the shell installer:

```bash
curl -fsSL https://icuvisor.app/install.sh | sh
```

On native Windows / PowerShell:

```powershell
iwr -useb https://icuvisor.app/install.ps1 | iex
```

See [SECURITY.md](SECURITY.md#installer-integrity) for installer signature verification and in-place update behaviour.

Prefer a package manager? `brew install ricardocabral/icuvisor/icuvisor`, `scoop install icuvisor`, or download the `.dmg` / `.msi` from the [releases page](https://github.com/ricardocabral/icuvisor/releases).

Learn more on how to connect your AI assistant, read the tool catalog, and troubleshoot stale conversations or cached tool catalogs at <https://icuvisor.app>.

## For developers

### Build from source

```bash
git clone https://github.com/ricardocabral/icuvisor.git
cd icuvisor
make build
./bin/icuvisor version
```

### Project layout

```
cmd/icuvisor/             Binary entrypoint
internal/app/             CLI dispatch, startup wiring, `setup` / `diagnostics` commands
internal/cli/prompt/      Terminal prompting (masked input) for first-run setup
internal/config/          Config load/validate/write, athlete-ID/timezone normalization, HTTP bind, dotenv, redaction
internal/credstore/       OS keychain wrapper (macOS Keychain, Windows Credential Manager, libsecret)
internal/diagnostics/     Redacted runtime/config snapshot for `icuvisor diagnostics`
internal/intervals/       intervals.icu API client (Basic Auth, retries, structured errors)
internal/clients/         Shared typed client interfaces (athlete profile, etc.)
internal/mcp/             MCP server + stdio/Streamable HTTP transports, schema, recovery
internal/tools/           Tool implementations (registered via `tools.Catalog()`)
internal/toolcatalog/     Catalog hashing and stale-catalog CI guard surface
internal/toolchecks/      Cross-tool invariants (delete-mode gating, examples, schema snapshots)
internal/coach/           Coach-mode roster and per-athlete tool ACLs
internal/safety/          Delete-mode resolution and registration-time gating
internal/response/        Terse/full response shaping and `_meta` plumbing
internal/analysis/        Deterministic analyzer math + interval-source / auto-lap classifier
internal/prompts/         Curated MCP prompt registry
internal/resources/       MCP Resources (workout syntax, event categories, schemas, analysis formulas, athlete profile)
internal/athleteprofile/  Athlete profile read shaping shared by tool + resource
internal/workoutdoc/      WorkoutDoc Parse/Serialize for the upstream description DSL
internal/customitemschemas/ Custom-item content schema samples used by write validation
internal/units/           Unit enum parsing + preferred-unit conversion
internal/streams/         Canonical stream key normalization
docs/                     PRD, roadmap, design notes
```

### Development

Requires Go 1.25+ and (optionally) [`golangci-lint`](https://golangci-lint.run) and [`goreleaser`](https://goreleaser.com).

```bash
make build       # build ./bin/icuvisor
make test        # unit tests
make test-race   # tests with the race detector
make lint        # golangci-lint
make check       # fmt-check + vet + lint + test-race (run before pushing)
make snapshot    # local goreleaser snapshot
make docs-tools  # regenerate website tool catalog data
make help        # list all targets
```

See [CONTRIBUTING.md](CONTRIBUTING.md), [SECURITY.md](SECURITY.md), and the [PRD](docs/prd/PRD-icuvisor.md).

## License

[MIT](LICENSE).
