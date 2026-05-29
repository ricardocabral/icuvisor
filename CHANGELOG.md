# Changelog

All notable changes to icuvisor are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Documented the strength-training upstream API gap and current best-effort gym support via simple calendar notes or supported events.
- Cookbook eval scenarios now cover athlete-local date lookup before activity detail/interval/split analysis for prompts like race retrospectives and run split comparisons.

### Changed

- Weekly planning and workout cookbook guidance now allow simple gym time blocks while warning that detailed structured strength sets remain future scope until upstream API support is documented.
- Recovery and weekly-review prompts now tell assistants to state missing/null Intervals readiness before using HRV, resting HR, sleep, subjective wellness scales, and provider `_native` fields as cautious fallback evidence.
- The readiness-check cookbook now includes Garmin/null-readiness fallback guidance while preserving explicit sleepQuality, sleepScore, feel, and provider-native scale labels.
- Activity tool descriptions and the activity-retrospective cookbook now make the `get_activities` → `activity_id` → detail/interval/splits routing path explicit for described or relative-date activities.

## [0.1.5] - 2026-05-27

### Added

- Documentation now explains stale conversations and cached MCP tool catalogs, including when to start a new chat, reconnect tools, verify `icuvisor version`, run `icuvisor diagnostics`, and avoid pasting API keys into assistant conversations.
- New curated MCP prompt `weekly_review` guides assistants through a structured previous-week training review, planned-vs-completed comparison, wellness caveats, and optional next-week preview using existing read/analyzer tools.
- Time-relative reads now include athlete-local as-of anchors: `get_today` always returns `_meta.as_of`, `_meta.as_of_date`, `_meta.as_of_weekday`, and `_meta.timezone`, and `get_activities`, `get_events`, and `get_wellness_data` return the same metadata when the requested range includes the athlete-local current day.
- Wellness rows now include row-level `_meta.field_semantics` for `hydration` and `hydrationVolume` when those fields are present, preserving the upstream field names without inventing units.
- Terse event and activity read responses now include upstream `tags` arrays when intervals.icu returns them, preserving order and explicit empty lists without requiring `include_full:true`.

### Changed

- `add_or_update_event`, `create_workout`, and `update_workout` now merge free-text `description` prose with structured `workout_doc` steps instead of forcing callers to choose one source; the `<!-- icuvisor:steps -->` sentinel controls insertion point when present.
- Clarified write-tool and planning guidance that `description` writes replace the upstream description/DSL field rather than appending notes; preserving structured workout steps on updates requires supplying the desired `workout_doc` explicitly.

### Fixed

- `add_or_update_event` and `update_workout` now surface `_meta.description_only_workout_warning` when an existing workout-shaped item is updated with `description` but no `workout_doc`, so assistants can avoid accidentally replacing structured steps when they only intended to add prose.
- Structured WorkoutDoc serialization now rejects step descriptions containing duration or distance tokens (for example `2h15m`, `45m`, `400mtr`, or `5km`) so planned workout duration/load cannot be doubled by submitting both a structured duration field and an inline DSL time token. `validate_workout` reports this as `STRUCTURAL_TOKEN_IN_STEP_DESCRIPTION`.

## [0.1.4] - 2026-05-24

### Added

- Shell installers at `https://icuvisor.app/install.sh` (POSIX sh; Linux, macOS, WSL, Git-Bash, CI) and `https://icuvisor.app/install.ps1` (native Windows PowerShell). Both detect OS/arch, pick the right release archive, verify the SHA-256 checksum, and verify the keyless Sigstore signature on `SHA256SUMS.txt` when `cosign` is available. The release workflow now signs `SHA256SUMS.txt` with cosign keyless OIDC and publishes the `.pem` + `.sig` alongside it. Re-running the installer updates an existing install in place: it reuses the directory of any `icuvisor` already on `PATH`, refuses to overwrite Homebrew- or Scoop-managed binaries (directing users to `brew upgrade` / `scoop update` instead), short-circuits when already at the target version (`--force` / `-Force` to override), uses an atomic rename so a running binary keeps executing off its old inode, and exposes `--check` / `-Check` for "is an update available?" in scripts (exit 0 = up to date, 1 = update available).
- `get_today` read-only MCP tool returns one terse daily digest for "how's today looking?": today's CTL/ATL/TSB, wellness, completed activities, planned events, and NOTE/race annotations, with `include_full` widening each source-shaped section.
- New `update_activity` write MCP tool renames and/or edits the free-text description of one completed activity by `activity_id` with sparse fields (omit to leave unchanged; explicit empty string clears the description). Non-destructive metadata edit that does not alter recorded streams or interval analysis; gated by `ICUVISOR_DELETE_MODE` like other write tools (registered in `safe` and `full`).
- New `set_activity_intervals` delete-mode MCP tool writes a structured `workout_doc` as one completed activity's description so intervals.icu re-parses the DSL into rendered intervals; optional `prose` argument is interleaved verbatim around the serialized steps. The response `_meta.interval_source_intent: "structured_workout"` records the write's intent so downstream readers can distinguish a manually written interval set from device auto-laps (see `internal/analysis/interval_source.go`); `_meta.workout_doc_warning` is set when upstream stored the description but did not parse the DSL. Treated as destructive: registered only when `ICUVISOR_DELETE_MODE=full`.
- New read-only `validate_workout` MCP tool validates a free-text `description`, a structured `workout_doc`, or both, and returns the canonical merged DSL icuvisor would submit on a write. Prose passes through verbatim; only malformed structured-step lines surface as PARSE_ERROR. The reusable merge logic lives in `internal/workoutdoc.MergeDescription` for future write-path use.
- MCP tool calls now emit structured `slog` entries for call start/completion, including tool name, status, duration, redacted argument/response byte counts, and approximate MCP token counts (`bytes/4`) without logging raw arguments or response payloads.
- `get_activities` and `get_activity_details` now surface athlete-defined activity custom fields under each row's `custom_fields` map in terse mode. The field codes are discovered from `ACTIVITY_FIELD` custom-item definitions (fetched once per athlete and cached) and `get_activities` requests them alongside its terse field set so field-limited list responses still include them. Previously these values were only reachable with `include_full:true`. A custom-item lookup failure degrades gracefully: the activity read still succeeds, just without `custom_fields`.
- `get_activities` and `get_activity_details` responses now report the athlete IANA timezone in `_meta.timezone` — the zone each row's `start_date_local` is expressed in — so assistants derive calendar dates from the athlete's timezone instead of reporting activities on the wrong day.
- `add_or_update_event` now accepts a top-level `indoor` boolean and event reads echo `indoor`, allowing assistants to set intervals.icu's planned-event Indoor toggle for trainer rides instead of relying on tags or `VirtualRide` alone.

### Changed

- Homebrew formula and Scoop manifest homepages now point at `https://icuvisor.app` (the canonical project site) instead of the unused `icuvisor.dev`.
- Expand `icuvisor://workout-syntax` resource with a cheat sheet and common-mistakes section; nudge `add_or_update_event`, `create_workout`, `update_workout`, and the `weekly_planning` prompt toward structured `workout_doc` + coexisting prose and recommend a `validate_workout` pre-flight when uncertain about the DSL.

### Fixed

- `add_or_update_event` now treats an explicit `tags: []` as a tag-clearing update and sends `"tags": []` upstream instead of omitting the field as "no change".

## [0.1.3] - 2026-05-22

### Fixed

- Setup and athlete-ID validation now accept the bare-numeric athlete IDs that intervals.icu issues to Strava-linked accounts (e.g. `612345`), not just the `i12345` form. The previous validation rejected any ID without a leading `i`, blocking setup for those athletes. Bare-numeric IDs are kept as-is — the `i` prefix is part of the ID and is never added or stripped. (#24)

## [0.1.2] - 2026-05-22

### Added

- `get_activity_intervals` now exposes scalar upstream custom interval fields, such as manually-entered lactate values, under each interval's `custom_fields` map in terse mode without requiring `include_full:true`.
- `create_workout`, `update_workout`, and `add_or_update_event` now set `_meta.workout_doc_warning` when a structured `workout_doc` was uploaded but intervals.icu did not parse it into a rendered structured workout, so callers learn the workout will display as plain text without graphical interval segments instead of silently assuming success.

### Changed

- Strava-blocked activity reads now report a single stable `unavailable.reason` of `strava_blocked` across every tool. `get_activities` and `get_activity_messages` previously emitted `strava_tos` for the same condition that `get_activity_details`, `get_extended_metrics`, and `get_activity_streams` reported as `strava_blocked`.

## [0.1.1] - 2026-05-21

### Added

- Cookbook section on the documentation website (`web/content/cookbook/`): a quick prompt library plus eight reusable multi-step recipes (weekly review, readiness check, activity retrospective, FTP/zones review, season and block plan, build and schedule workouts, race-week taper, coach roster triage). Each recipe is written for reliable tool activation and grounded, hallucination-resistant answers, derived from common athlete and coach tasks on the intervals.icu forum.
- Cookbook activation and answer-quality eval harness (`scripts/eval/`): self-contained scenario specs, an LLM-judge rubric scoring tool activation, grounding, subjective-scale correctness, coverage, actionability, and conciseness, and a `run_eval.py` runner with a CI-safe `--validate` mode (`make eval-validate`) that checks scenarios against the tool catalog, plus a live mode that drives icuvisor over MCP stdio and scores each run.

### Fixed

- Interval `start_time` and `end_time` are now decoded correctly when intervals.icu returns them as numeric second offsets instead of strings. The previous string-only typing failed to decode those activities, breaking `get_extended_metrics`, `get_activity_intervals`, and `compute_activity_segment_stats` for them.

## [0.1.0] - 2026-05-21

### Fixed

- HTTP 422 responses from intervals.icu are now categorized as a stable `ErrValidation` sentinel (matchable via `errors.Is`/`errors.As`) instead of falling through to the generic `ErrUpstream` path. Write tools (`update_wellness`, `create_custom_item`, `update_custom_item`, `update_sport_settings`) parse the upstream rejection body to extract the offending field name and return a short, actionable message — for example `intervals.icu rejected field "WhoopStrain": create it under Settings > Custom Fields in intervals.icu, or omit it from this call`. The full upstream body is logged via `slog` and never surfaced to the LLM.

### Added

- `get_extended_metrics` now surfaces the strain-score power-duration model parameters when intervals.icu has fitted them: `strain_score_cp_watts` (critical power), `strain_score_w_prime_kj` (W', converted from upstream joules), and `strain_score_p_max_watts` (maximal power). Terse mode still drops them when upstream returns null.
- KR5 benchmark harness/report now compares analyzer-enabled versus analyzer-disabled fixture modes with v2 response-token metrics, source-tool usage validation, raw-stream pull counts, and TP-098 core-promotion evidence without promoting tools in this change.
- Documented the upstream coverage gap for precomputed per-activity zone-time fields after auditing the v0.2 fixture corpus with a reproducible local script.
- `analyze_trend`, `analyze_distribution`, `analyze_correlation`, and `analyze_efforts_delta` are registered in the full toolset to compute deterministic analyzer reductions with mandatory `_meta`, closed `analysis_metric` validation, terse-by-default responses, and unit-explicit best-efforts deltas.
- `compute_zone_time`, `compute_load_balance`, `compute_baseline`, and `compute_compliance_rate` are registered in the full toolset as deterministic analyzer-family compute tools using existing read outputs, mandatory analyzer `_meta`, explicit missing/insufficient signals, and activation hints that tell assistants not to roll their own row/stream reductions.
- `get_activity_histogram` summarizes one activity's power, heart-rate, or pace distribution into terse bucketed seconds/percentages using configured zones when available and fixed-width stream buckets otherwise.
- `compute_activity_segment_stats` is registered in the full toolset as the analyzer-family raw-stream exception, computing segment mean/median/p90, drift, Pw:HR decoupling, NP, and IF from canonical activity streams with mandatory analyzer `_meta`.
- `get_fitness_projection` simulates deterministic CTL/ATL/TSB scenarios from current fitness with horizon, ramp, recovery-week, and optional planned-load assumptions documented in analyzer `_meta`.
- `get_activity_intervals` now emits additive `_meta.interval_source` and `_meta.auto_lap_suspected` signals so generic device auto-laps are not confused with structured workout segments.
- Added shared analyzer response scaffolding so planned analyzer tools consistently emit mandatory `_meta.method`, `_meta.source_tools`, `_meta.n`, `_meta.missing_days`, `_meta.missing_action`, and `_meta.insufficient_sample` fields.
- Claude Desktop Extension (`.mcpb`) packaging for the macOS universal release artifact, including a binary-server manifest, secure Desktop-managed `api_key` configuration, local pack/validate tooling, and extension-first install docs.
- `icuvisor://analysis-formulas` MCP Resource documents canonical analyzer formula refs for HR drift, Pw:HR decoupling, polarization index, EF, VI, and z-score.
- `get_hr_curves` and `get_pace_curves` expose upstream-computed heart-rate and pace curves as full-toolset siblings to `get_power_curves`, with terse/default buckets, raw `include_full` payloads, and athlete-preferred pace units.
- Added NOTE-category examples for `add_or_update_event` so assistants and users can discover nutrition plans, travel logistics, daily reminders, and coach annotations without a separate note tool.
- Added reusable closed `analysis_metric` enum helpers for planned analyzer tools, including canonical schema values, conservative aliases, source metadata, and concise unknown-metric hints.
- `get_gear_list` lists intervals.icu gear IDs/names in the full toolset with a manual refresh path for the per-athlete gear cache.
- Activity reads now surface `gear_id`, resolved `gear_name`, and explicit `gear_resolution` statuses when upstream exposes gear IDs.
- Activity and wellness reads now disambiguate nutrition fields: activity `calories` is exposed as `calories_burned`, while wellness intake/macros use `calories_intake`, `carbs_g`, `protein_g`, and `fat_g` with `_meta.field_semantics` labels.
- Homebrew install path: releases now produce a `darwin_universal` tarball and publish a Homebrew formula to the [`ricardocabral/homebrew-icuvisor`](https://github.com/ricardocabral/homebrew-icuvisor) tap. Install with `brew install ricardocabral/icuvisor/icuvisor`. The macOS tarball binary is not yet codesigned independently of the DMG, so first run may require dismissing Gatekeeper; signing the tarball binary is tracked as a follow-up.

### Changed

- Promoted `analyze_trend`, `compute_zone_time`, and `compute_baseline` into the default core toolset after TP-098/KR5 benchmark evidence showed positive net token savings and eliminated or avoided LLM-visible raw-stream pulls.
- Analyzer-family tool descriptions now lead with concrete activation prompts and explicitly tell assistants not to fetch raw `get_*` rows or streams and reduce them in chat.
- Strava-blocked activity responses now point users to the intervals.icu Connections page and Download old data action for the native device provider, naming Garmin/Wahoo/etc. when explicit payload evidence is available.
- `get_wellness_data` provenance now reports provider-native sleep/readiness `native_scale` labels for Garmin, WHOOP, Oura, and Polar, while unresolved sources report `unknown` instead of a guessed device scale.
- Write-tool echo responses now strip upstream null keys by default, including custom-item create/update responses, while preserving meaningful zero, false, and empty-string values.
- Documented the TP-102 remote-connector decision: icuvisor will not add legacy SSE or generic public-tunnel recipes for ChatGPT-style remote custom connector UIs before the hosted relay; local stdio and loopback Streamable HTTP remain the supported ChatGPT paths.
- `icuvisor setup` now writes a non-secret `credential_ref` to generated config files so users and docs can see the OS keychain service/account while the API key remains outside JSON. Setup stores and verifies the keychain credential before writing the config file, so keychain failures do not leave a fresh onboarding config behind.

## [0.0.2] - 2026-05-19

### Changed

- `icuvisor setup` always prompts for the athlete ID and requires the `i` prefix (for example `i12345`). intervals.icu does not expose the ID through the API, so the previous `/athlete/0/profile` autodetect could silently succeed with an empty ID; the API key and athlete ID are now verified together against `/athlete/{id}`.
- The server loads the platform default config (`os.UserConfigDir()/icuvisor/config.json`) when neither `--config` nor `ICUVISOR_CONFIG` is set, so a fresh `icuvisor setup` works without needing an env file or explicit `--config` path.
- Startup logs drop the noisy "env file not found" line on healthy runs, log when the API key was loaded from the OS keychain, and include the active athlete ID and key source on the "server starting" line.

## [0.0.1] - 2026-05-18

### Added

- Initial public release.

[Unreleased]: https://github.com/ricardocabral/icuvisor/compare/v0.1.5...HEAD
[0.1.5]: https://github.com/ricardocabral/icuvisor/compare/v0.1.4...v0.1.5
[0.1.4]: https://github.com/ricardocabral/icuvisor/compare/v0.1.3...v0.1.4
[0.1.3]: https://github.com/ricardocabral/icuvisor/compare/v0.1.2...v0.1.3
[0.1.2]: https://github.com/ricardocabral/icuvisor/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/ricardocabral/icuvisor/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/ricardocabral/icuvisor/compare/v0.0.2...v0.1.0
[0.0.2]: https://github.com/ricardocabral/icuvisor/compare/v0.0.1...v0.0.2
[0.0.1]: https://github.com/ricardocabral/icuvisor/releases/tag/v0.0.1
