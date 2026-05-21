# TP-007 — Response shaping primitives (terse, null-strip, _meta, scale labels, TZ, athlete-ID)

## Mission

Land the response-shaping foundation that every v0.2 read tool will share, so individual tool tasks downstream stay small and consistent. Build the helpers, document the conventions, refactor `get_athlete_profile` to use them, and lock the conventions with tests.

Roadmap items (ROADMAP.md v0.2):

- Terse-by-default + `include_full` opt-in.
- Null-value keys stripped from responses before serialization.
- `ICUVISOR_DEBUG_METADATA=true` re-enables `fetched_at` / `query_type`.
- `_meta.missing_fields` on every read tool that strips nulls.
- `_meta.server_version` in every response.
- In-response scale labels on subjective fields.
- Disambiguating field names (`calories_burned`, `distance_km` / `distance_mi`).
- Timezone normalization to the athlete's configured TZ.
- Athlete-ID normalization (`i12345` / `12345`) — confirm and centralize.
- Per-athlete unit normalization from `preferred_units` (framework; downstream tools apply it).

PRD anchors: §7.2.D Response shaping (every bullet), §7.4 #11 (`preferred_units` round-trip).

Complexity: Blast radius 3 (every read tool depends on this), Pattern novelty 2, Security 1, Reversibility 1 = 7 → Review Level 2. Size: M.

## Dependencies

- **TP-002** — intervals.icu client (athlete profile field shape)
- **TP-004** — existing `get_athlete_profile` tool to refactor onto the new primitives

## Context to Read First

- `CLAUDE.md` (esp. "Tools must be terse-by-default")
- `docs/prd/PRD-icuvisor.md` §7.2.D, §7.4 #11
- `ROADMAP.md` v0.2
- `internal/tools/` current tool implementations
- `internal/config/` for athlete-ID / TZ helpers from TP-001
- `internal/intervals/` for the athlete profile struct

## File Scope

Expected files:

- `internal/response/` — new package with shaping helpers (null-strip, meta-assembly, scale-label registry, TZ rendering, field-name disambiguation helpers)
- `internal/response/*_test.go`
- `internal/tools/get_athlete_profile.go` — refactor to consume the new helpers
- `internal/config/` — small additions if athlete-ID / TZ normalization needs centralizing
- `CHANGELOG.md`
- `taskplane-tasks/TP-007-response-shaping-primitives/STATUS.md`

Do **not** add new tools, write paths, or wellness-specific logic here. This is plumbing only.

## Steps

### Step 1: Design the shaping pipeline

- [ ] Define the order of operations at the response boundary: typed-struct → marshal-to-map → strip nulls → add `_meta` → marshal-to-JSON for MCP
- [ ] Decide where canonical field renames (`distance` → `distance_km` / `distance_mi`) live: per-tool struct tags vs central rename map. Prefer struct tags; document the choice in `STATUS.md`
- [ ] Record decisions and tradeoffs in `STATUS.md` before coding

### Step 2: Implement null-stripping with `_meta` callouts

- [ ] Recursive null-strip over nested objects and arrays-of-objects; **do not** strip `0`, `""`, or `false`
- [ ] Apply per top-level row independently so multi-row shapes stay consistent
- [ ] Emit `_meta.fields_present: [...]` (keys actually present after stripping) and `_meta.missing_fields: [...]` (keys that were stripped) per row; only emit either when at least one strip happened
- [ ] `include_full: true` skips stripping and surfaces raw nulls

### Step 3: Implement `_meta.server_version` and debug-metadata gate

- [ ] Inject `_meta.server_version` into every response from a single chokepoint
- [ ] `fetched_at` / `query_type` only present when `ICUVISOR_DEBUG_METADATA=true`; read once at startup, not per call
- [ ] Invalid `ICUVISOR_DEBUG_METADATA` values default to `false` quietly (this is a debug toggle, not a safety gate)

### Step 4: Scale-label registry and in-response labels

- [ ] Central registry mapping field name → scale label string (e.g. `"feel": "1-5 (athlete-reported)"`)
- [ ] Helper that, given a response row, emits `_meta.scales: { <field>: <label> }` for every registered field present in the row
- [ ] No labels for fields not in the registry (silence is fine; over-labelling is noise)

### Step 5: Timezone, athlete-ID, and unit-system plumbing

- [ ] Centralize athlete-ID normalization (`i12345` / `12345` → emit `i12345`) in `internal/config`; remove any duplicate logic
- [ ] Render date/time fields in the athlete's configured TZ at the presentation boundary; document the convention in package doc
- [ ] Add a `UnitSystem` type sourced from `preferred_units` on the athlete profile; expose a helper that downstream tools call to choose `distance_km` vs `distance_mi` field names and to convert values
- [ ] Surface the active unit system in `_meta.units` per response so the LLM cannot misread the choice

### Step 6: Refactor `get_athlete_profile` onto the new helpers

- [ ] Replace ad-hoc shaping with the new pipeline
- [ ] Add an `include_full: bool` argument; default keeps the existing terse shape
- [ ] Verify response includes `_meta.server_version` and `_meta.units` (when known)
- [ ] Update tests to assert the new contract

### Step 7: Lock conventions with tests

- [ ] Table-driven tests for null-stripping (incl. `0` / `""` / `false` preservation)
- [ ] Tests for `_meta.fields_present` / `_meta.missing_fields` correctness
- [ ] Tests for the debug-metadata gate (env on / off / invalid)
- [ ] Tests for the scale-label registry
- [ ] Tests for athlete-ID normalization (both input forms)
- [ ] Tests for TZ rendering at the boundary
- [ ] Run `gofmt`, `go vet`, `make test`, `make lint`, `make build`

## Reference Implementation Policy

- `hhopke/intervals-icu-mcp` (MIT) may be consulted for shaping semantics. Do not depend on it.
- GPL/copyleft implementation code is off limits — do not read, copy, paraphrase, or transliterate.

## Acceptance Criteria

- A single `internal/response` package owns null-stripping, `_meta` assembly, scale labels, TZ rendering, and unit-system plumbing.
- `get_athlete_profile` consumes it and exposes `include_full`.
- Every response from `get_athlete_profile` includes `_meta.server_version`.
- Tests cover the cases in Step 7.
- `make test`, `make build`, and `make lint` pass.

## Do NOT

- Do not add new tools.
- Do not implement wellness-specific provenance / sleep dual-scale (that is TP-011).
- Do not implement stream-key canonicalization or the unit enum (that is TP-008).
- Do not introduce a global logger or change logging conventions.
- Do not break the v0.1 `get_athlete_profile` contract for existing manual smokes without updating the smoke docs.

## Documentation

Must update:

- `STATUS.md`
- `CHANGELOG.md`

Check if affected:

- `README.md` only if a user-visible env var changed
- `docs/clients/*.md` only if response shape examples need refreshing

## Git Commit Convention

Commit at step boundaries with messages prefixed by `TP-007`, for example: `TP-007 add response null-strip helper`.

---

## Amendments

_Add amendments below this line only._
