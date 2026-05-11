# Changelog

All notable changes to icuvisor are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

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
