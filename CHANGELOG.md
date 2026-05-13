# Changelog

All notable changes to icuvisor are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `add_or_update_event` MCP tool for non-destructive calendar event creates/updates in write-enabled modes, with verbatim free-text descriptions, structured `workout_doc` serialization to the upstream description DSL, tag preservation, and planned target fields.
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

[Unreleased]: https://github.com/ricardocabral/icuvisor/compare/HEAD...HEAD
