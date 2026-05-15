# Changelog

All notable changes to icuvisor are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Post-update schema-change notification metadata: every tool response now includes `_meta.catalog_hash`, and simulated catalog divergence emits `_meta.schema_changed` guidance to open a new conversation.
- OS keychain credential storage for the intervals.icu API key via macOS Keychain, Windows Credential Manager, and Linux libsecret/Secret Service, with a no-CGO wrapper and injectable test store.
- CLI help via `icuvisor --help`, `-h`, `help`, and `version --help`, documenting commands, flags, environment variables, examples, and exit codes.
- `--env-file` CLI flag and `ICUVISOR_ENV_FILE` environment variable for reading a custom local env file instead of the default `.env`; explicitly requested env-file paths must exist (the default `.env` remains silently skipped when absent).
- KR5 benchmark harness, shared prompt set, redacted fixtures, and methodology/results documentation comparing icuvisor core/full against the hhopke and mvilanova Python reference servers.
- MCP Prompts for curated training analysis, recovery check, weekly planning, race-week taper, and coach roster triage workflows.
- MCP Resources for long-form reference content: `icuvisor://workout-syntax`, `icuvisor://event-categories`, `icuvisor://custom-item-schemas`, and a dynamic cached `icuvisor://athlete-profile`, with inline tool descriptions trimmed to point at resource URIs.
- `ICUVISOR_TOOLSET` with default `core` and opt-in `full` tool catalog tiers, including `icuvisor_list_advanced_capabilities` for discovering hidden tools and `_meta.toolset` response metadata.
- Streamable HTTP MCP transport alongside stdio, with `stdio` remaining the default, loopback-only `127.0.0.1:8765` HTTP binding by default, explicit `ICUVISOR_TRANSPORT`/`ICUVISOR_HTTP_BIND` and CLI overrides, and warning logs for LAN binds.
- Startup logs now include the server version and a structured "started listening" entry once the MCP transport is ready.
- Adversarial safe-mode delete prompt corpus with redacted Codex local-binary outcomes, plus static catalog safety matrix coverage for v0.3 write/delete gating.
- `input_examples` and JSON Schema `examples` for complex v0.3 write-tool input schemas, with catalog-wide regression tests ensuring new complex write tools carry examples.
- Destructive delete MCP tools gated by `ICUVISOR_DELETE_MODE=full`: `delete_event`, `delete_events_by_date_range` with a 31-day inclusive range cap, `delete_activity`, `delete_custom_item`, `delete_sport_settings`, and `delete_gear`, each returning deleted IDs and terse before-shape deletion metadata.
- `create_custom_item` and `update_custom_item` MCP tools for write-enabled modes to create custom chart/field/stream/panel/zones items, validate schema-driven `content` before upload using readable custom-item schema samples, and echo the full read shape.
- `create_workout`, `update_workout`, and gated `delete_workout` MCP tools for workout-library template CRUD, including TP-019 structured `workout_doc` serialization to the upstream description DSL and delete registration only in `ICUVISOR_DELETE_MODE=full`.
- `update_sport_settings` MCP tool for write-enabled modes to update sport-scoped FTP, threshold heart rate, threshold pace, and gated zone-definition overwrites with recompute/delete-mode/unit metadata.
- `update_wellness` MCP tool for write-enabled modes to sparsely update manual wellness fields, enforce subjective scales, convert preferred-unit weight to upstream kilograms, reject device-owned `sleepScore`/`_native` fields, and echo the updated read shape.
- `add_or_update_event` MCP tool for non-destructive calendar event creates/updates in write-enabled modes, with verbatim free-text descriptions, structured `workout_doc` serialization to the upstream description DSL, tag preservation, and planned target fields.
- `apply_training_plan` MCP tool for write-enabled modes to fetch workout-library plan content server-side, default to dry-run previews with conflict markers, partially apply conflict-free days, and gate replacement deletes to `ICUVISOR_DELETE_MODE=full`.
- `link_activity_to_event` MCP tool for write-enabled modes to manually pair completed activities with planned events when auto-pairing misses, including date-mismatch warnings.
- `add_activity_message` MCP tool for write-enabled modes to append free-text activity comments/messages without overwriting prior messages, including terse append confirmations and normalized athlete-ID metadata.
- `internal/workoutdoc` public package API with `WorkoutDoc`, `Parse`, and `Serialize` for deterministic Intervals.icu workout-description DSL round-trips, including golden fixtures for repeats, ramps, freeride, cadence, power, heart-rate, pace, and RPE targets.
- `ICUVISOR_DELETE_MODE` safety gate with `safe`/`full`/`none` modes, registration-time write/delete tool filtering, and `_meta.delete_mode` response metadata.
- CI guards for tool-schema snapshot stability and confusable tool-name first sentences, plus canonical per-tool argument schema snapshots for the v0.2 read catalog.
- Disambiguated first-sentence descriptions for activity, event/calendar, workout-library, and custom-item read-tool clusters.
- `get_custom_items` and `get_custom_item_by_id` MCP tools for custom chart/field/stream/panel/zones reads, with terse list rows and per-`item_type` `content` preservation on detail reads.
- `get_workout_library` and `get_workouts_in_folder` MCP tools for workout-library folder/plan and template reads, with terse defaults and raw `workout_doc` preservation behind `include_full`.
- `get_events`, `get_event_by_id`, and `get_training_plan` MCP tools for bounded calendar event reads, detail reads with one bounded list-scan recovery for upstream 404 inconsistencies, and active training-plan assignment reads, with terse defaults, upstream category preservation, structured unavailable handling, and `include_full` raw payload opt-in.
- `get_wellness_data` MCP tool with distinct `sleepQuality`/`sleepScore`/`sleepSecs`, custom-field preservation, provenance/staleness metadata, native Polar/Garmin/Oura sub-fields, scale labels, null stripping, and `include_full` raw payload opt-in.
- Fitness and metrics read cluster: `get_fitness`, `get_best_efforts`, `get_power_curves`, `get_training_summary`, and `get_extended_metrics`, including upstream availability evidence for extended metrics and no zero-filling of unavailable fields.
- `get_activity_messages` MCP tool with terse activity comments/notes and athlete-timezone timestamp rendering.
- `get_activity_streams` and `get_activity_splits` MCP tools with canonical stream keys, heavy stream sample opt-in, and preferred-unit virtual splits.
- `get_activity_details` and `get_activity_intervals` MCP tools with terse/default responses, `include_full` raw payload opt-in, canonical interval units, and Strava-unavailable handling for blocked activity details.
- `get_activities` MCP tool with date-range listing, bounded pagination tokens, `include_unnamed` filtering, terse unit-disambiguated rows, `include_full` raw payload opt-in, and structured Strava-unavailable markers.
- Response-boundary intervals.icu unit enum parsing and athlete-preferred distance, pace, and speed conversion with unknown-unit metadata preservation.
- Shared response-shaping primitives for terse/null-stripped MCP responses, response-owned `_meta.server_version`/`_meta.units`, scale labels, debug metadata gating, timezone rendering, athlete-ID normalization, and `get_athlete_profile` integration.
- Codex CLI local MCP validation guide with ephemeral stdio configuration, non-interactive tool-call settings, cleanup guidance, and README pointer.
- Manual macOS Claude Desktop v0.1 setup guide and repeatable local smoke checklist for the binary → MCP stdio → intervals.icu → `get_athlete_profile` path.
- `get_athlete_profile` MCP tool with typed intervals.icu client wiring, terse/default and `include_full` responses, normalized athlete IDs, unit/timezone metadata, sanitized errors, and protocol/tool tests.
- MCP stdio server skeleton using the official Go MCP SDK, with SDK-free tool registry scaffolding and protocol tests for initialize, tool listing, tool calls, malformed requests, and sanitized handler errors.
- intervals.icu HTTP client core with Basic Auth, `User-Agent` propagation, retry/backoff handling, structured errors, and athlete profile retrieval.
- v0.1 foundation CLI with thin `main`, internal app startup wiring, build-version propagation, and `icuvisor version` support.
- Manual v0.1 config loader for JSON/env/`.env` inputs with centralized athlete-ID normalization and secret redaction.
- Foundation tests covering CLI delegation, config precedence/defaults, validation errors, redaction, and athlete-ID normalization.
- Actionable, redacted config-file startup errors for missing or invalid v0.1 JSON config.
- Initial repository scaffolding: Go module, Makefile, GoReleaser config, GitHub Actions CI/release pipelines, golangci-lint config, issue/PR templates, CODEOWNERS.
- Project documentation: README, CONTRIBUTING, CODE_OF_CONDUCT, SECURITY, ROADMAP, CHANGELOG.
- PRD for v1.0 (`docs/prd/PRD-icuvisor.md`).

### Changed

- Config loading now resolves API keys in the order `INTERVALS_ICU_API_KEY` process env, OS keychain, plaintext `.env`/JSON legacy files, then error; plaintext file-sourced keys remain supported but emit a migration warning.

[Unreleased]: https://github.com/ricardocabral/icuvisor/compare/HEAD...HEAD
