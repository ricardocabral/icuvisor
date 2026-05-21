# TP-008 — Pace/unit enum coverage and snake_case stream-key canonicalization

## Mission

Make activity reads safe across cycling / running / swimming / rowing / walking by typing the intervals.icu unit enum end-to-end and canonicalizing every stream key to snake_case at the response boundary. Downstream tool tasks (TP-009 / TP-010 / TP-011) depend on this.

Roadmap items (ROADMAP.md v0.2):

- Canonical snake_case stream keys across all activities/devices (upstream public discussion).
- Exhaustive pace-unit enum coverage (`MINS_KM`, `MINS_MILE`, `SECS_100M`, `SECS_500M`, …); unknown units degrade to `_meta.unknown_unit: true` rather than failing the call.

PRD anchors: §7.2.D ("Stream-key canonicalization"), §7.4 #17 (pace-unit enum, treat as load-bearing).

Complexity: Blast radius 2, Pattern novelty 2, Security 1, Reversibility 1 = 6 → Review Level 2. Size: M.

## Dependencies

- **TP-002** — intervals.icu client (decode path)

## Context to Read First

- `CLAUDE.md`
- `docs/prd/PRD-icuvisor.md` §7.2.D, §7.4 #17
- `ROADMAP.md` v0.2
- `internal/intervals/` decoder code from TP-002

## File Scope

Expected files:

- `internal/units/` — typed `Unit` enum, parser, conversion helpers, `UnitUnknown` fallback carrying the raw upstream string
- `internal/streams/` — canonical stream-key map and canonicalizer (snake_case)
- `internal/units/*_test.go`, `internal/streams/*_test.go`
- `internal/streams/testdata/` — fixtures capturing inconsistent upstream casings (`groundContactTime` vs `ground_contact_time`, etc.)
- `CHANGELOG.md`
- `taskplane-tasks/TP-008-units-and-stream-canonicalization/STATUS.md`

## Steps

### Step 1: Build the typed `Unit` enum

- [ ] Cover every member listed in PRD §7.4 #17: distance (`M`, `KM`, `MI`, `YD`); pace (`MINS_KM`, `MINS_MILE`, `SECS_100M`, `SECS_500M`); speed (`KMH`, `MPH`, `MS`); time (`SECS`, `MINS`, `HOURS`); power (`WATTS`, `WKG`, `PERCENT_FTP`); HR (`BPM`, `PERCENT_HR`, `PERCENT_LTHR`, `PERCENT_MAX_HR`); misc (`RPE`, `Z1`…`Z7`, `PERCENT`, `KCAL`, `KJ`)
- [ ] Implement `ParseUnit(string) (Unit, raw string)` that returns `UnitUnknown` + the raw upstream string for unrecognized values; never error
- [ ] Log unknown values at `WARN` via `slog` so we can grow the enum from telemetry; never log full response bodies

### Step 2: Convert at the response boundary, not at decode

- [ ] Decode keeps the upstream `Unit` verbatim
- [ ] Add converters `ToPreferred(value, fromUnit, sys UnitSystem)` that produce the athlete-preferred view (miles vs km, etc.) without losing the original
- [ ] Unknown units pass through with their raw label intact; the caller emits `_meta.unknown_unit: <value>` for that field
- [ ] Coordinate with the `_meta.units` field defined in TP-007 so the response is self-describing

### Step 3: Build the stream-key canonicalizer

- [ ] Map every known upstream stream key (camelCase, PascalCase, snake_case variants) to a single canonical snake_case key
- [ ] Apply the map at the response boundary for `get_activity_streams` (and any tool that surfaces stream keys)
- [ ] Unknown keys pass through as snake_case-converted best-effort and emit `_meta.unknown_stream_keys: [...]` for the row; never drop data

### Step 4: Tests

- [ ] Table-driven tests for `ParseUnit` over every enum member plus several unknowns
- [ ] Tests for `ToPreferred` covering each conversion family (distance, pace, speed)
- [ ] Tests for stream-key canonicalization across the fixtures captured under `testdata/`
- [ ] Run `gofmt`, `go vet`, `make test`, `make lint`, `make build`

## Reference Implementation Policy

- `hhopke/intervals-icu-mcp` (MIT) may be consulted for unit/stream-key behavior; do not add as a dependency.
- GPL/copyleft implementation code is off limits.

## Acceptance Criteria

- A typed `Unit` enum covers every member in §7.4 #17 with `UnitUnknown` fallback.
- A stream-key canonicalizer turns every known upstream key into snake_case at the response boundary.
- Unknown units and unknown stream keys never crash a tool call; they surface via `_meta`.
- Tests for both packages pass.
- `make test`, `make build`, `make lint` pass.

## Do NOT

- Do not convert units at decode time; preserve the raw upstream value for downstream use.
- Do not silently drop unknown stream keys.
- Do not bake `preferred_units` into the unit package; it belongs in the response layer (TP-007).

## Documentation

Must update:

- `STATUS.md`
- `CHANGELOG.md`

## Git Commit Convention

Commit at step boundaries with messages prefixed by `TP-008`, for example: `TP-008 add typed unit enum`.

---

## Amendments

_Add amendments below this line only._
