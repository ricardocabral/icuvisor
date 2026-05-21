# TP-008 — Status

**Issue:** v0.2 — read path
**Status:** ✅ Complete
**Review Level:** 2
**Iteration:** 3
**Current Step:** Step 4: Tests
**Last Updated:** 2026-05-11
**State:** Complete

_Task scaffolded from PROMPT.md; hydrated with executable checkboxes._

## Execution Log

| Time             | Event                   | Notes                                                                                                    |
| ---------------- | ----------------------- | -------------------------------------------------------------------------------------------------------- |
| 2026-05-11 14:33 | Task started            | Runtime V2 lane-runner execution                                                                         |
| 2026-05-11 14:33 | Step 1 started          | Build the typed `Unit` enum                                                                              |
| 2026-05-11 14:35 | Hydrated                | Added task steps/checklists from PROMPT.md                                                               |
| 2026-05-11 14:45 | Step 1 complete         | Code review APPROVE; typed unit enum implemented                                                         |
| 2026-05-11 15:05 | Step 2 complete         | Code review APPROVE; response-boundary unit conversion implemented                                       |
| 2026-05-11 15:18 | Step 3 complete         | Code review APPROVE after R001 fixes                                                                     |
| 2026-05-11 15:23 | Step 4 complete         | gofmt, go vet, make test, make lint, and make build passed                                               |
| 2026-05-11 15:17 | Exit intercept reprompt | Supervisor provided instructions (615 chars) — reprompting worker                                        |
| 2026-05-11 15:24 | Completion pass         | STATUS.md complete and verification recorded; `.DONE` absent and left to runtime per worker instructions |
| 2026-05-11 15:18 | Exit intercept close    | Supervisor directed session close: "close"                                                               |
| 2026-05-11 15:18 | Worker iter 1           | done in 2692s, tools: 157                                                                                |
| 2026-05-11 15:18 | No progress             | Iteration 1: 0 new checkboxes (1/3 stall limit)                                                          |
| 2026-05-11 15:18 | Step 1 started          | Build the typed `Unit` enum                                                                              |
| 2026-05-11 15:20 | Exit intercept close | Supervisor directed session close: "close" |
| 2026-05-11 15:20 | Worker iter 2 | done in 87s, tools: 9 |
| 2026-05-11 15:20 | No progress | Iteration 2: 0 new checkboxes (2/3 stall limit) |
| 2026-05-11 15:20 | Step 1 started | Build the typed `Unit` enum |

## Step 1: Build the typed `Unit` enum

**Status:** ✅ Complete

- [x] Create isolated `internal/units/` package for the intervals.icu enum; keep response-layer `UnitSystem` / preferred-unit shaping out of this package
- [x] Cover every member listed in PRD §7.4 #17 with stable exported constants and exact upstream string values: distance (`M`, `KM`, `MI`, `YD`); pace (`MINS_KM`, `MINS_MILE`, `SECS_100M`, `SECS_500M`); speed (`KMH`, `MPH`, `MS`); time (`SECS`, `MINS`, `HOURS`); power (`WATTS`, `WKG`, `PERCENT_FTP`); HR (`BPM`, `PERCENT_HR`, `PERCENT_LTHR`, `PERCENT_MAX_HR`); misc (`RPE`, `Z1`…`Z7`, `PERCENT`, `KCAL`, `KJ`)
- [x] Implement `ParseUnit(string) (Unit, raw string)` as never-error parsing: trim whitespace, match known uppercase values case-sensitively, return empty raw for known units, and return `UnitUnknown` plus the exact trimmed upstream token for unknown/empty values
- [x] Preserve unknown-unit data by documenting that callers must keep the second return value whenever `UnitUnknown` is returned
- [x] Log unknown values at `WARN` via `slog.Default().Warn("unknown intervals.icu unit", "unit", raw)` so we can grow the enum from telemetry; never log full response bodies
- [x] Add table-driven `internal/units` tests in Step 1 for every enum member, unknown/mixed-case/future tokens, raw preservation, and no-error parsing

## Step 2: Convert at the response boundary, not at decode

**Status:** ✅ Complete

- [x] Verify and test that intervals decode keeps upstream unit tokens verbatim as strings (for example `SportSettings.PaceUnits` remains `MINS_KM`) rather than converting at JSON decode time
- [x] Add response-boundary converters in `internal/response` (not `internal/units`) with a typed result from `ToPreferred(value, fromUnit, sys UnitSystem)` containing preferred value/unit/field suffix, original value/upstream unit label, conversion flag, and unknown-unit metadata when applicable
- [x] Define conversion policy for distance (`M`, `KM`, `MI`, `YD`), speed (`KMH`, `MPH`, `MS`), and run/walk pace (`MINS_KM`, `MINS_MILE`) while preserving sport-specific `SECS_100M` and `SECS_500M` labels with no generic pace conversion
- [x] Add raw-aware unknown handling so callers can pass the `ParseUnit` raw value through unchanged and surface `_meta.unknown_unit: <value>` instead of losing labels or failing conversion
- [x] Migrate existing `get_athlete_profile` response-boundary pace shaping to use the new converter for `SportSettings.PaceUnits`, preserving original upstream pace unit metadata and existing terse/full behavior
- [x] Coordinate converted values with TP-007 `_meta.units` by extending response unit metadata only at the response boundary, keeping caller-supplied `_meta.units` stripped
- [x] Add Step 2 tests for distance/speed/run-pace conversions, swim/row pace pass-through, unknown raw metadata, profile shaping integration, and response-owned `_meta.units`
- [x] Update `CHANGELOG.md` under `[Unreleased]` for user-visible response-boundary unit conversion

## Step 3: Build the stream-key canonicalizer

**Status:** ✅ Complete

- [x] Create `internal/streams` canonicalizer package with a greppable map from known upstream camelCase/PascalCase/snake_case stream keys to one canonical snake_case key
- [x] Include known activity stream/dynamics keys used by intervals.icu and PRD §7.2.C (`watts`, `heartrate`, `cadence`, `distance`, `altitude`, `velocity_smooth`, `latlng`, `ground_contact_time`, `vertical_oscillation`, `stride_length`, `ground_contact_balance`, `left_right_balance`, `core_temperature`, `w_prime_balance`, etc.)
- [x] Add response-boundary helpers for future `get_activity_streams` responses that canonicalize map rows/series without mutating inputs and merge collisions deterministically without dropping values
- [x] Unknown keys pass through as snake_case-converted best-effort and emit `_meta.unknown_stream_keys: [...]` for the row; never drop data
- [x] Add stream canonicalization fixtures under `internal/streams/testdata/` covering inconsistent upstream casings
- [x] Fix collision merging so JSON-decoded stream sample arrays are preserved as separate collision values without mutating input slices
- [x] Preserve original upstream unknown key spellings in `_meta.unknown_stream_keys` while keeping best-effort snake_case output fields

## Step 4: Tests

**Status:** ✅ Complete

- [x] Table-driven tests for `ParseUnit` over every enum member plus several unknowns
- [x] Tests for `ToPreferred` covering each conversion family (distance, pace, speed)
- [x] Tests for stream-key canonicalization across the fixtures captured under `testdata/`
- [x] Run `gofmt`, `go vet`, `make test`, `make lint`, `make build`

## Discoveries

| Date | Discovery | Impact |
| ---- | --------- | ------ |

## Blockers

| Date             | Blocker     | Attempts             | Status |
| ---------------- | ----------- | -------------------- | ------ |
| 2026-05-11 14:36 | Review R001 | plan Step 1: REVISE  |
| 2026-05-11 14:39 | Review R001 | plan Step 1: APPROVE |
| 2026-05-11 14:44 | Review R001 | code Step 1: APPROVE |
| 2026-05-11 14:49 | Review R001 | plan Step 2: REVISE  |
| 2026-05-11 14:51 | Review R001 | plan Step 2: APPROVE |
| 2026-05-11 15:01 | Review R001 | code Step 2: APPROVE |
| 2026-05-11 15:04 | Review R001 | plan Step 3: APPROVE |
| 2026-05-11 15:09 | Review R001 | code Step 3: REVISE  |
| 2026-05-11 15:14 | Review R001 | code Step 3: APPROVE |
