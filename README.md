# icuvisor

[![Go Reference](https://pkg.go.dev/badge/github.com/ricardocabral/icuvisor.svg)](https://pkg.go.dev/github.com/ricardocabral/icuvisor)
[![Go Report Card](https://goreportcard.com/badge/github.com/ricardocabral/icuvisor)](https://goreportcard.com/report/github.com/ricardocabral/icuvisor)
[![CI](https://github.com/ricardocabral/icuvisor/actions/workflows/ci.yml/badge.svg)](https://github.com/ricardocabral/icuvisor/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/ricardocabral/icuvisor?sort=semver)](https://github.com/ricardocabral/icuvisor/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/ricardocabral/icuvisor)](go.mod)
[![codecov](https://codecov.io/gh/ricardocabral/icuvisor/branch/main/graph/badge.svg)](https://codecov.io/gh/ricardocabral/icuvisor)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/ricardocabral/icuvisor/badge)](https://securityscorecards.dev/viewer/?uri=github.com/ricardocabral/icuvisor)
[![Conventional Commits](https://img.shields.io/badge/Conventional%20Commits-1.0.0-blue.svg)](https://www.conventionalcommits.org)

> An open-source, locally-installed [Model Context Protocol](https://modelcontextprotocol.io) server for [intervals.icu](https://intervals.icu), distributed as a single Go binary.

icuvisor lets amateur athletes talk to their training data through Claude, ChatGPT, Pi, Cursor, and other MCP-compatible AI clients — without running Python, without trusting a third-party SaaS, and without paying anything.

## Status

> Pre-alpha. The walking-skeleton (v0.1) is in progress. See [ROADMAP.md](ROADMAP.md).

## Features (planned for v1.0)

- Single signed binary for macOS, Windows, and Linux.
- ~25 MCP tools covering activities, wellness, fitness, events, and custom items.
- Terse-by-default responses tuned for LLM context windows.
- Coach mode for multi-athlete rosters.
- stdio and Streamable HTTP transports.
- One-click client configs for Claude Desktop, Claude Code, ChatGPT, Pi, Cursor, Continue, Zed.
- API key stored in the OS keychain, never on disk in plain text.

See the full [PRD](docs/prd/PRD-icuvisor.md).

## MCP tool catalog

Currently implemented tools:

- `get_athlete_profile` — retrieves athlete identity, units, timezone, FTP/thresholds, zones, and sport settings.
- `get_activities` — lists activities for a date range with `include_unnamed`, server-side pagination via opaque `next_page_token`, terse unit-disambiguated rows by default, `include_full` raw payload opt-in, and structured Strava-unavailable markers.
- `get_activity_details` — retrieves one activity's terse metadata/metrics with `include_full` raw payload opt-in and structured Strava-unavailable markers.
- `get_activity_intervals` — retrieves analyzed intervals/groups for one activity with canonical intervals.icu unit enum values and raw payload opt-in.
- `get_activity_streams` — retrieves canonical snake_case stream channels; sample arrays require `include_full:true` or explicit `keys`.
- `get_activity_splits` — returns manual or virtual per-km/per-mile splits from intervals/streams while honoring preferred units.
- `get_activity_messages` — lists comments/notes for one activity with athlete-timezone timestamp rendering.
- `add_activity_message` — appends a non-destructive free-text comment/message to one activity in write-enabled modes without overwriting prior messages.
- `get_fitness` — returns CTL/ATL/TSB trends over a local date range.
- `get_best_efforts` — returns upstream best efforts grouped by sport and power/heart-rate/pace buckets.
- `get_power_curves` — returns upstream mean-maximal power curve buckets with raw arrays behind `include_full`.
- `get_training_summary` — aggregates volume, neutral training load, sRPE, and upstream zone-order totals over a date range.
- `get_extended_metrics` — returns only upstream-exposed extended activity metrics, dropping unavailable fields instead of zero-filling.
- `get_wellness_data` — returns daily wellness rows with custom fields, distinct `sleepQuality`/`sleepScore`/`sleepSecs`, provenance/staleness metadata, `_native` provider sub-fields, scale labels, and `include_full` raw payload opt-in.
- `update_wellness` — sparsely updates writable manual wellness fields in write-enabled modes (`safe`/`full`), including subjective scales, preferred-unit weight conversion, measurements, injury text, and the `locked` flag while rejecting device-owned `sleepScore`/`_native` fields.
- `update_sport_settings` — updates sport-scoped FTP, threshold heart rate, and threshold pace in write-enabled modes; optional `zones` overwrites prior zone definitions and is rejected unless `ICUVISOR_DELETE_MODE=full`.
- `get_events` — lists bounded athlete-local date-range calendar events with upstream category enum values, terse rows by default, truncation metadata, and `include_full` raw payload opt-in.
- `get_event_by_id` — fetches one calendar event by ID, with one bounded list-scan recovery for upstream detail 404 inconsistencies and structured non-error `upstream_inconsistency` misses.
- `add_or_update_event` — creates or updates a non-destructive calendar event when write tools are enabled (`safe`/`full`), preserving free-text descriptions verbatim or serializing structured `workout_doc` steps to the upstream description DSL.
- `apply_training_plan` — applies a workout-library training plan from a `start_date` anchor with `dry_run:true` by default, per-day conflict markers, skip-existing partial writes, and full-mode-only replacement of conflicting events.
- `delete_event` — deletes one calendar event only when `ICUVISOR_DELETE_MODE=full`, returning the deleted ID and `_meta.deleted` terse before-shape echo.
- `delete_events_by_date_range` — deletes calendar events only when `ICUVISOR_DELETE_MODE=full` across a required athlete-local `start_date`/`end_date` range capped at 31 inclusive days, optionally filtered by category.
- `link_activity_to_event` — manually pairs a completed activity with a planned event when auto-pairing misses, surfacing date-mismatch warnings without requiring a destructive confirmation flag.
- `get_training_plan` — fetches the active upstream training-plan assignment with lightweight plan summary by default, structured no-active-plan responses, and raw nested plan/workout payloads behind `include_full`.
- `get_workout_library` — lists workout-library folders/plans with terse counts and optional top-level workout templates.
- `get_workouts_in_folder` — lists workout-library templates in one folder with structured-step summaries by default and raw `workout_doc` only with `include_full`.
- `create_workout` — creates a reusable workout-library template in write-enabled modes (`safe`/`full`), preserving free-text descriptions or serializing structured `workout_doc` steps to the upstream description DSL.
- `update_workout` — sparsely updates one workout-library template in write-enabled modes (`safe`/`full`), leaving omitted fields untouched and serializing replacement `workout_doc` steps to the upstream description DSL.
- `delete_workout` — deletes one workout-library template only when `ICUVISOR_DELETE_MODE=full`; it has no model-controlled `confirm` argument.
- `delete_activity` — deletes one activity only when `ICUVISOR_DELETE_MODE=full`, returning `_meta.deleted` with a terse before-shape echo.
- `delete_sport_settings` — deletes one sport-settings definition only when `ICUVISOR_DELETE_MODE=full`, echoing the removed sport/threshold summary in `_meta.deleted`.
- `delete_gear` — deletes one gear item only when `ICUVISOR_DELETE_MODE=full`, echoing the removed gear summary in `_meta.deleted`.
- `get_custom_items` — lists custom charts, fields, streams, panels, histograms, maps, and zones with terse `id`/`name`/`item_type` rows.
- `get_custom_item_by_id` — fetches one custom item with its full per-`item_type` `content` payload preserved.
- `create_custom_item` — creates custom charts, fields, streams, panels, histograms, maps, or zones in write-enabled modes (`safe`/`full`), validating `content` against readable per-`item_type` schema samples before upload.
- `update_custom_item` — sparsely updates one custom item in write-enabled modes (`safe`/`full`), validating content patches against the existing item's readable schema and leaving omitted fields untouched.
- `delete_custom_item` — deletes one custom item only when `ICUVISOR_DELETE_MODE=full`, returning `_meta.deleted` with a terse before-shape echo.
- `icuvisor_list_advanced_capabilities` — lists tools hidden from the default core catalog and explains how to enable the full toolset.

## Install

> Installers will land with v1.0. For now, build from source:

```bash
git clone https://github.com/ricardocabral/icuvisor.git
cd icuvisor
make build
./bin/icuvisor version
```

## Quickstart

```bash
# 1. Get an intervals.icu API key from https://intervals.icu/settings
# 2. Provide v0.1 manual config via env or JSON
export INTERVALS_ICU_API_KEY="YOUR_INTERVALS_ICU_API_KEY"
export INTERVALS_ICU_ATHLETE_ID="i12345"
./bin/icuvisor version
```

For local development, `icuvisor` can read a local untracked `.env` file containing `INTERVALS_ICU_API_KEY` and `INTERVALS_ICU_ATHLETE_ID`. Do not commit real API keys. For MCP client config, use process env vars or pass a JSON file with `--config /path/to/icuvisor.json` using fields `api_key`, `athlete_id`, `timezone`, `api_base_url`, and `http_timeout`.

### Delete/write safety mode

`ICUVISOR_DELETE_MODE` is read once at startup and controls which write-capable tools are registered with the MCP server:

- `safe` (default): write tools are registered, delete tools are omitted from the catalog.
- `full`: write and delete tools are registered.
- `none`: write and delete tools are omitted, leaving read-only tools only.

Unknown or empty values resolve to `safe`. The active mode is reported in response metadata as `_meta.delete_mode`.

### Toolset tiers

`ICUVISOR_TOOLSET` is read once at startup and controls how much of the MCP tool catalog is registered:

- `core` (default): registers the daily-use catalog for activities, fitness, wellness, events, non-destructive writes, and `icuvisor_list_advanced_capabilities`.
- `full`: registers the core catalog plus advanced/heavier tools such as raw streams, workout-library and custom-item management, sport settings, training-plan application, and delete-capable tools when delete mode also allows them.

Unknown or empty values resolve to `core`. Change the environment variable in your MCP client/server configuration and restart icuvisor for the catalog to change. The active tier is reported in response metadata as `_meta.toolset`.

`icuvisor_list_advanced_capabilities` remains available in `core` so an AI client can discover hidden full-only tools and tell the user to set `ICUVISOR_TOOLSET=full` when a prompt needs them. Toolset tiering is orthogonal to delete/write safety: destructive tools still require `ICUVISOR_DELETE_MODE=full` even when `ICUVISOR_TOOLSET=full` is set.

For the v0.1 macOS Claude Desktop manual JSON setup and smoke checklist, see [`docs/clients/claude-desktop.md`](docs/clients/claude-desktop.md). For Codex CLI local MCP validation, see [`docs/clients/codex-local.md`](docs/clients/codex-local.md).

## Project layout

```
cmd/icuvisor/       Binary entrypoint
internal/app/       CLI/default startup wiring
internal/config/    Manual v0.1 config loading and athlete-ID normalization
internal/intervals/ intervals.icu API client
internal/mcp/       MCP server + transports
internal/tools/     Tool implementations
docs/               PRD, roadmap, design notes
```

## Development

Requires Go 1.23+ and (optionally) [`golangci-lint`](https://golangci-lint.run) and [`goreleaser`](https://goreleaser.com).

```bash
make build       # build ./bin/icuvisor
make test        # unit tests
make test-race   # tests with the race detector
make lint        # golangci-lint
make snapshot    # local goreleaser snapshot
make help        # list all targets
```

See [CONTRIBUTING.md](CONTRIBUTING.md) before opening a PR.

## Security

Found a vulnerability? Please read [SECURITY.md](SECURITY.md) — do **not** open a public issue.

## Acknowledgements

icuvisor is a clean-room Go implementation that draws inspiration from the wider intervals.icu MCP community, with particular thanks to [Marc Vilanova](https://github.com/mvilanova)'s [`intervals-mcp-server`](https://github.com/mvilanova/intervals-mcp-server) (Python, GPLv3) as the reference that proved the use case. No GPL source is read or copied; the implementation is built from intervals.icu's public API documentation.

## License

[MIT](LICENSE).
