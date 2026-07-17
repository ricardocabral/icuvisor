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


> icuvisor is an open-source [Model Context Protocol](https://modelcontextprotocol.io) server for [intervals.icu](https://intervals.icu), shipped as a single signed Go binary so athletes and coaches can talk to their training data from Claude, ChatGPT, Cursor, and other MCP-compatible clients. Local mode runs on your machine and keeps your intervals.icu API key in the OS keychain. Hosted mode is optional and uses `https://connect.icuvisor.app/mcp` for clients that need a public HTTPS MCP endpoint. Usage limits from your AI client or model provider, GitHub/package downloads, and intervals.icu account are separate. End-user docs live at <https://icuvisor.app>.

## For users

### Why icuvisor

icuvisor is designed to keep training-data access simple, local, and easy for an AI assistant to use correctly:

- **Local-first by default:** your intervals.icu API key is read by the local `icuvisor` process from the OS keychain or an explicit headless fallback, not passed as an MCP tool argument.
- **Hosted when required:** if your AI client cannot run a local MCP server and needs a public HTTPS connector, use `https://connect.icuvisor.app/mcp` and Intervals.icu OAuth instead of an API key.
- **One binary to install:** the server ships as a Go binary with shell installers and package-manager paths, so setup does not depend on a language runtime in the user's AI chat.
- **Structured, terse responses:** read tools return compact JSON-shaped results by default, with fuller payloads behind explicit `include_full` options for cases such as raw streams.
- **Model-compatible tool profiles:** instead of exposing a large generated API surface by default, `ICUVISOR_TOOLSET=core` is the curated daily training catalog, `compact` further reduces the catalog for smaller/local models, and `full` opts in to advanced or heavier workflows. If a capability is hidden, ask the assistant to call `icuvisor_list_advanced_capabilities` before switching profiles.
- **Gear names when they are knowable:** activity summaries and details include `gear_id` plus `gear_name` for bikes, shoes, and other gear when intervals.icu exposes a resolvable gear item; unresolved IDs keep an explicit `gear_resolution` status instead of inventing a name.
- **Units and scales are explicit:** distances, paces, work, hydration, and related fields use unit-labelled names or `_meta` legends so the model does not have to infer whether a value is kilometres, miles, kilojoules, or a 1-5 rating. Activity fuel fields distinguish active `calories_burned`, athlete-logged `carbs_ingested_g`, upstream `carbs_used_g`, and wellness `kcal_consumed` intake.
- **Profile readiness warnings:** athlete-profile reads flag missing sport thresholds or zones in `_meta.warnings` so the assistant can ask you to update settings before producing threshold- or zone-based analysis and plans.
- **Calendar writes are safety-shaped:** assistants can add workouts, notes, races, and unavailable date ranges such as Sick, Injured, or Holiday blocks; range writes are per-day, retry-safe, and report same-day conflicts without deleting or overwriting existing workouts.
- **Delete safety is outside the model's reach:** destructive tools are registered only when the process-level `ICUVISOR_DELETE_MODE` allows them; there is no per-call `confirm` argument for the assistant to invent.

The fastest path on Linux, macOS (without Homebrew), WSL, and CI is the shell installer:

```bash
curl -fsSL https://icuvisor.app/install.sh | sh
```

On native Windows / PowerShell:

```powershell
iwr -useb https://icuvisor.app/install.ps1 | iex
```

See [SECURITY.md](SECURITY.md#installer-integrity) for installer signature verification and in-place update behaviour.

Prefer a package manager? Use `brew install ricardocabral/tap/icuvisor` on macOS, or install from Winget on Windows:

```powershell
winget install --id RicardoCabral.icuvisor --exact
```

Open a new PowerShell or Command Prompt window after installation, then run `icuvisor version`. Windows users can also use the PowerShell installer above or download the `.msi` from the [releases page](https://github.com/ricardocabral/icuvisor/releases).

### Connect from Cursor

Run `icuvisor setup` first so the local server can read your intervals.icu API key from the OS keychain or your explicit headless config. Do not paste API keys into Cursor or into an MCP tool call.

Use the Cursor install link for a local stdio server with no secrets embedded: [Install icuvisor in Cursor](https://cursor.com/install-mcp?name=icuvisor&config=%7B%22command%22%3A%22icuvisor%22%2C%22args%22%3A%5B%5D%2C%22env%22%3A%7B%22ICUVISOR_CONFIG%22%3A%22%22%7D%7D).

If you prefer manual configuration, add this to `~/.cursor/mcp.json` and leave `ICUVISOR_CONFIG` empty unless you intentionally use a non-default config path:

```json
{
  "mcpServers": {
    "icuvisor": {
      "command": "icuvisor",
      "args": [],
      "env": {
        "ICUVISOR_CONFIG": ""
      }
    }
  }
}
```

### MCP discovery

`server.json` in this repository describes icuvisor for MCP Registry-style discovery and points to the signed MCPB release artifact. No Glama badge is shown until Glama exposes a stable public icuvisor page; speculative discovery links are intentionally omitted.

Learn how to connect your AI assistant, try beginner prompt examples, read the tool catalog, and troubleshoot stale conversations or cached tool catalogs at <https://icuvisor.app>.

### Downloadable prompt packs

Copyable client prompt packs are available in [`docs/prompts/`](docs/prompts/README.md) for Weekly review, Coaching conversation handoff, Race-week taper, Ride analysis, Fueling review, Masters plan review, and Coach roster triage modes. Use them as custom mode/profile instructions in clients that support saved modes, or paste the prompt block at the start of a fresh chat in clients that do not. The packs are designed to route assistants through Icuvisor's deterministic tools, `_meta.method`/unit metadata, and explicit caveats instead of generic prompt-only coaching or invented baselines.

### Fitness projection with ATP/periodization targets

`get_annual_training_plan` summarizes existing PLAN, TARGET, and NOTE calendar events into season phases, weekly load/TSS targets, recovery/context notes, and `_meta.projection_bridge.weekly_plan_targets`. `propose_annual_training_plan` creates a read-only deterministic season proposal from caller-provided goals and constraints when no calendar writes should occur. `apply_annual_training_plan` is the separate full-toolset write path for those proposals: it defaults to dry-run, returns a preview token, and only writes deterministic Icuvisor-owned ATP phase notes after an explicit commit call with that token. Copy existing or proposed bridge rows into `get_fitness_projection.weekly_plan_targets` to model future CTL/ATL/TSB scenarios without asking the assistant to invent daily loads. `get_fitness_projection` distributes each ISO-Monday weekly `training_load` evenly as `training_load/7` across projected future dates in that week. Explicit `planned_daily_loads` win for matching dates and are not redistributed.

```json
{
  "start_date": "2026-06-03",
  "horizon_days": 14,
  "weekly_plan_targets": [
    { "week_start_date": "2026-06-01", "training_load": 700 },
    { "week_start_date": "2026-06-08", "training_load": 840 }
  ],
  "planned_daily_loads": [
    { "date": "2026-06-10", "training_load": 60 }
  ]
}
```

The ATP `_meta.projection_bridge` reports which weekly TARGET rows are safe to copy and which partial or missing-load weeks were excluded. The projection `_meta.assumptions` reports target counts, filled days, override counts, the ISO-Monday anchor convention, and `source_tools` adds planning sources when weekly targets are supplied.

Example calendar write prompt: "Mark me sick from 2026-08-10 through 2026-08-12" maps to `add_unavailable_date_range` with `category: "SICK"`, `start_date`, and `end_date`; use `include_full: true` only when raw upstream event payloads are needed.

## For developers

### Build from source

```bash
git clone https://github.com/ricardocabral/icuvisor.git
cd icuvisor
make build
./bin/icuvisor version
```

### Direct CLI

The release also includes `icuvisor-cli`, a standalone view over the same registered tool catalog used by MCP. It is useful for scripts or local agents that load a concise skill/command contract instead of an MCP tool schema. Run `icuvisor setup` once to provision credentials; never pass an API key as a tool argument.

```bash
icuvisor-cli list
icuvisor-cli describe get_today
icuvisor-cli call get_today --args '{}'
icuvisor-cli call get_activities --args-file request.json
```

`list`, `describe`, and successful `call` results are JSON on stdout. Diagnostics and errors are written to stderr. `call` accepts exactly one JSON object through `--args <json>` or `--args-file <path>` (and defaults to `{}`). The standalone view uses the configured local athlete only; coach workflows remain MCP-only.

### Project layout

```
cmd/icuvisor/             MCP server binary entrypoint
cmd/icuvisor-cli/         Standalone direct-tool CLI binary entrypoint
internal/app/             MCP process dispatch, startup wiring, `setup` / `diagnostics` commands
internal/cli/             Standalone CLI view and setup prompting
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
make docs-tools  # regenerate website tool catalog/schema data
make help        # list all targets
```

See [CONTRIBUTING.md](CONTRIBUTING.md), [SECURITY.md](SECURITY.md), and the [PRD](docs/prd/PRD-icuvisor.md).

## License

[MIT](LICENSE).
