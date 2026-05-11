# TP-007 — Status

**Issue:** v0.2 — response shaping primitives
**Status:** ✅ Complete
**Review Level:** 2
**Iteration:** 2
**Current Step:** Step 7: Lock conventions with tests
**Last Updated:** 2026-05-11
**State:** Complete

_Task scaffolded from PROMPT.md; STATUS hydrated for execution recovery._

| 2026-05-11 12:15 | Task started | Runtime V2 lane-runner execution |
| 2026-05-11 12:15 | Step 1 started | Design the shaping pipeline |
| 2026-05-11 12:20 | Hydration | Expanded STATUS.md with all prompt steps and review level |
| 2026-05-11 12:24 | Step 2 started | Implement null-stripping with `_meta` callouts |
| 2026-05-11 12:38 | Step 3 started | Implement `_meta.server_version` and debug-metadata gate |
| 2026-05-11 13:02 | Step 4 started | Scale-label registry and in-response labels |
| 2026-05-11 13:18 | Step 5 started | Timezone, athlete-ID, and unit-system plumbing |
| 2026-05-11 14:20 | Step 6 started | Refactor `get_athlete_profile` onto the new helpers |
| 2026-05-11 14:30 | Step 7 started | Lock conventions with tests |

### Step 1: Design the shaping pipeline

**Status:** ✅ Complete

- [x] Define the order of operations at the response boundary: typed-struct → marshal-to-map → strip nulls → add `_meta` → marshal-to-JSON for MCP
- [x] Decide where canonical field renames (`distance` → `distance_km` / `distance_mi`) live: per-tool struct tags vs central rename map. Prefer struct tags; document the choice in `STATUS.md`
- [x] Record decisions and tradeoffs in `STATUS.md` before coding

#### Step 1 decisions

**Pipeline order:** tools construct typed response structs with stable JSON tags, the response package converts those structs to a `map[string]any` through JSON marshal/unmarshal semantics so tags and `omitempty` stay authoritative, the shaper strips JSON-null values unless `include_full` is true, row metadata is assembled after stripping so `_meta.fields_present` reflects the final terse row, common metadata (`server_version`, units, debug metadata, scales) is merged through one response-package chokepoint, and the shaped map is finally marshaled for MCP text content while the same shaped value is used as structured content.

**Canonical field names:** field renames live in per-tool response structs via JSON tags (for example, downstream tools expose `distance_km` or `distance_mi` directly rather than emitting `distance` and asking a central map to rewrite it). A central rename map would hide schema changes away from the typed contract, make output schemas harder to audit, and risk unexpected rewrites across tools. Shared response helpers may provide unit-name/key builders, but each tool still chooses the final JSON field through its struct or explicit row map.

**Tradeoffs:** JSON round-tripping adds a small CPU allocation cost, but keeps shaping behavior aligned with the serialized API contract and avoids reflection code that would need to duplicate `encoding/json` tag rules. Metadata is added after null stripping so meta reports the payload the LLM actually sees; `include_full: true` bypasses stripping to preserve raw typed nulls for debugging/inspection while still going through common metadata injection. Debug-only metadata is controlled by startup configuration rather than per-call environment reads so tests and long-running MCP sessions are deterministic. Central helpers own cross-cutting conventions (`_meta`, scale labels, timezone rendering, units, debug gate), while tool structs own product-specific field selection and terse/full semantics.

### Step 2: Implement null-stripping with `_meta` callouts

**Status:** ✅ Complete

- [x] Recursive null-strip over nested objects and arrays-of-objects; **do not** strip `0`, `""`, or `false`
- [x] Apply per top-level row independently so multi-row shapes stay consistent
- [x] Emit `_meta.fields_present: [...]` (keys actually present after stripping) and `_meta.missing_fields: [...]` (keys that were stripped) per row; only emit either when at least one strip happened
- [x] `include_full: true` skips stripping and surfaces raw nulls
- [x] R001: Preserve null array elements while still stripping null object keys inside arrays-of-objects
- [x] R001: Add an explicit API for independently shaping named row collections inside wrapper responses
- [x] R001: Capture and enforce the `include_full` null-preservation convention for typed DTOs that use `omitempty`

### Step 3: Implement `_meta.server_version` and debug-metadata gate

**Status:** ✅ Complete

- [x] Inject `_meta.server_version` into every response from a single chokepoint
- [x] `fetched_at` / `query_type` only present when `ICUVISOR_DEBUG_METADATA=true`; read once at startup, not per call
- [x] Invalid `ICUVISOR_DEBUG_METADATA` values default to `false` quietly (this is a debug toggle, not a safety gate)
- [x] R001: Prevent reserved debug keys from leaking through `_meta.missing_fields`, including row collections
- [x] R001: Enforce response-wrapper objects as the common response shape so `_meta.server_version` is response-level
- [x] R001: Capture `ICUVISOR_DEBUG_METADATA` once during startup and pass the boolean through server/tool options
- [x] R001: Route `get_athlete_profile` through the shared response shaping chokepoint instead of ad-hoc marshal/meta
- [x] R001: Add common debug metadata assembly (`query_type`, deterministic `fetched_at`) to `internal/response`

### Step 4: Scale-label registry and in-response labels

**Status:** ✅ Complete

- [x] Central registry mapping field name → scale label string (e.g. `"feel": "1-5 (athlete-reported)"`)
- [x] Helper that, given a response row, emits `_meta.scales: { <field>: <label> }` for every registered field present in the row
- [x] No labels for fields not in the registry (silence is fine; over-labelling is noise)
- [x] R001: Use PRD field keys/labels for `sleepQuality` and `sleepScore`
- [x] R001: Make `_meta.scales` response-owned by replacing/removing stale caller-supplied scale metadata
- [x] R001: Use a truly unknown custom field in negative scale-label tests
- [x] R001: Match the PRD `sleepQuality` scale label exactly

### Step 5: Timezone, athlete-ID, and unit-system plumbing

**Status:** ✅ Complete

- [x] Centralize athlete-ID normalization (`i12345` / `12345` → emit `i12345`) in `internal/config`; remove any duplicate logic
- [x] Render date/time fields in the athlete's configured TZ at the presentation boundary; document the convention in package doc
- [x] Add a `UnitSystem` type sourced from `preferred_units` on the athlete profile; expose a helper that downstream tools call to choose `distance_km` vs `distance_mi` field names and to convert values
- [x] Surface the active unit system in `_meta.units` per response so the LLM cannot misread the choice
- [x] R001: Source `_meta.units` from athlete profile `preferred_units`/fallbacks in `get_athlete_profile`
- [x] R001: Distinguish absent/unknown `preferred_units` from known metric/imperial values with explicit profile fallback behavior
- [x] R001: Add primitive tests for unit-system derivation/conversion, timezone rendering, and athlete-ID display normalization
- [x] R001: Keep visible profile units consistent with `_meta.units` when `preferred_units` is present
- [x] R001: Make `_meta.units` response-owned by replacing/removing stale caller-supplied unit metadata
- [x] R001: Strip response-owned `_meta.units` from row-collection rows as well as root responses
- [x] R001: Use one source of truth for visible profile units and `_meta.units` when profile unit fields are empty/unknown

### Step 6: Refactor `get_athlete_profile` onto the new helpers

**Status:** ✅ Complete

- [x] Replace ad-hoc shaping with the new pipeline
- [x] Add an `include_full: bool` argument; default keeps the existing terse shape
- [x] Verify response includes `_meta.server_version` and `_meta.units` (when known)
- [x] Update tests to assert the new contract

### Step 7: Lock conventions with tests

**Status:** ✅ Complete

- [x] Table-driven tests for null-stripping (incl. `0` / `""` / `false` preservation)
- [x] Tests for `_meta.fields_present` / `_meta.missing_fields` correctness
- [x] Tests for the debug-metadata gate (env on / off / invalid)
- [x] Tests for the scale-label registry
- [x] Tests for athlete-ID normalization (both input forms)
- [x] Tests for TZ rendering at the boundary
- [x] Run `gofmt`, `go vet`, `make test`, `make lint`, `make build`

## Discoveries

| Date             | Finding      | Impact                           |
| ---------------- | ------------ | -------------------------------- |
| 2026-05-11 12:15 | Task started | Runtime V2 lane-runner execution |

## Blockers

None.
| 2026-05-11 12:18 | Review R001 | plan Step 1: APPROVE |
| 2026-05-11 12:21 | Review R001 | code Step 1: APPROVE |
| 2026-05-11 12:23 | Review R001 | plan Step 2: APPROVE |
| 2026-05-11 12:27 | Review R001 | code Step 2: REVISE |
| 2026-05-11 12:32 | Review R001 | code Step 2: APPROVE |
| 2026-05-11 12:35 | Review R001 | plan Step 3: APPROVE |
| 2026-05-11 12:40 | Review R001 | code Step 3: REVISE |
| 2026-05-11 12:47 | Review R001 | code Step 3: REVISE |
| 2026-05-11 12:58 | Review R001 | code Step 3: APPROVE |
| 2026-05-11 13:00 | Review R001 | plan Step 4: APPROVE |
| 2026-05-11 13:05 | Review R001 | code Step 4: REVISE |
| 2026-05-11 13:09 | Review R001 | code Step 4: REVISE |
| 2026-05-11 13:13 | Review R001 | code Step 4: APPROVE |
| 2026-05-11 13:16 | Review R001 | plan Step 5: APPROVE |
| 2026-05-11 13:22 | Review R001 | code Step 5: REVISE |
| 2026-05-11 13:58 | Review R001 | code Step 5: REVISE |
| 2026-05-11 14:08 | Review R001 | code Step 5: REVISE |
| 2026-05-11 14:13 | Review R001 | code Step 5: REVISE |

| 2026-05-11 14:15 | Worker iter 1 | killed (wall-clock timeout) in 7200s, tools: 250 |
| 2026-05-11 14:19 | Review R001 | code Step 5: APPROVE |
| 2026-05-11 14:22 | Review R001 | plan Step 6: APPROVE |
| 2026-05-11 14:26 | Review R001 | code Step 6: APPROVE |

| 2026-05-11 14:33 | Worker iter 2 | done in 1133s, tools: 106 |
| 2026-05-11 14:33 | Task complete | .DONE created |