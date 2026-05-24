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


> icuvisor is an open-source, locally installed [Model Context Protocol](https://modelcontextprotocol.io) server for [intervals.icu](https://intervals.icu), shipped as a single Go binary so athletes and coaches can talk to their training data from Claude, ChatGPT, Pi, Cursor, and other MCP-compatible clients. End-user docs live at <https://icuvisor.app>.

## For users

The fastest path on Linux, macOS (without Homebrew), WSL, and CI is the shell installer:

```bash
curl -fsSL https://icuvisor.app/install.sh | sh
```

On native Windows / PowerShell:

```powershell
iwr -useb https://icuvisor.app/install.ps1 | iex
```

Both scripts verify the SHA-256 checksum and, when [`cosign`](https://docs.sigstore.dev/cosign/installation/) is installed, the keyless Sigstore signature on `SHA256SUMS.txt`. **Re-running either one-liner updates an existing install in place** — it overwrites the binary at the existing `icuvisor` on your `PATH`, no-ops when already at the latest version, and uses `--check` / `-Check` to report whether an update is available without changing anything. To inspect the script before running:

```bash
curl -fsSL https://icuvisor.app/install.sh -o install.sh
less install.sh
sh install.sh
```

Prefer a package manager? `brew install ricardocabral/icuvisor/icuvisor`, `scoop install icuvisor`, or download the `.dmg` / `.msi` from the [releases page](https://github.com/ricardocabral/icuvisor/releases).

Learn more on how to connect your AI assistant, read the tool catalog and more at <https://icuvisor.app>.

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
cmd/icuvisor/        Binary entrypoint
internal/app/        CLI dispatch, startup wiring, `setup` / `diagnostics` commands
internal/config/     Config load/validate/write, athlete-ID/timezone normalization, HTTP bind, dotenv, redaction
internal/credstore/  OS keychain wrapper (macOS Keychain, Windows Credential Manager, libsecret)
internal/intervals/  intervals.icu API client (Basic Auth, retries, structured errors)
internal/mcp/        MCP server + stdio/Streamable HTTP transports, schema, recovery
internal/tools/      Tool implementations (registered via `tools.Catalog()`)
internal/toolcatalog/ Catalog hashing and stale-catalog CI guard
internal/coach/      Coach-mode roster and per-athlete tool ACLs
internal/safety/     Delete-mode resolution and registration-time gating
internal/response/   Terse/full response shaping and `_meta` plumbing
internal/prompts/    Curated MCP prompt registry
internal/resources/  MCP Resources (workout syntax, event categories, schemas, analysis formulas, athlete profile)
internal/workoutdoc/ WorkoutDoc Parse/Serialize for the upstream description DSL
docs/                PRD, roadmap, design notes
```

### Development

Requires Go 1.23+ and (optionally) [`golangci-lint`](https://golangci-lint.run) and [`goreleaser`](https://goreleaser.com).

```bash
make build       # build ./bin/icuvisor
make test        # unit tests
make test-race   # tests with the race detector
make lint        # golangci-lint
make snapshot    # local goreleaser snapshot
make docs-tools  # regenerate website tool catalog data
make help        # list all targets
```

See [CONTRIBUTING.md](CONTRIBUTING.md), [SECURITY.md](SECURITY.md), and the [PRD](docs/prd/PRD-icuvisor.md).

## License

[MIT](LICENSE).
