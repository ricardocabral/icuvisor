# TP-047-shaper-tree-walker-consolidation — Status

**Current Step:** Step 3: Implement
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-15
**Review Level:** 2
**Review Counter:** 8
**Iteration:** 1
**Size:** M

---

### Step 1: Snapshot pre-refactor output

**Status:** ✅ Complete

- [x] Lock deterministic fixture plan: `get_activities_terse.golden.json`, `get_activities_full.golden.json`, `get_fitness.golden.json`, `get_events_wrapper.golden.json`, `wellness_provenance.golden.json`; each uses synthetic DTO input, stable `response.Options`, reset catalog metadata, and canonical indented JSON bytes
- [x] Add automated golden snapshot test/generator that maps each case to its input, exact shaping options, fixture path, and comparison command without hitting the network
- [x] Capture `_meta`-shaped output as golden fixtures under `internal/response/testdata/`
- [x] Commit fixtures before touching shaper code
- [x] Fix R003 catalog runtime isolation so golden snapshots restore default metadata under shuffle/update/early returns
- [x] Convert representative golden inputs (`get_activities` terse/full and `get_fitness`) from maps to typed local DTO structs with JSON tags/omitempty matching tool shapes, then regenerate fixtures

### Step 2: Pick the approach

**Status:** ✅ Complete

- [x] Decide typed-shape vs single visitor walker
- [x] Justify in Decisions below (diff size, blast radius, `include_full` fit)
- [x] Sketch struct set or visitor signature
- [x] Address R005 plan details: specify marshal-round-trip replacement, package-boundary strategy, `include_full` handling, and provenance/debug predicate preservation before Step 3

### Step 3: Implement

**Status:** 🟨 In Progress

- [x] Add R008 guardrails: fresh-copy ownership for all JSON containers (including `include_full` maps), explicit converter fallback scope (`json.Marshaler`, `encoding.TextMarshaler`, `json.RawMessage`, unsupported values, numeric behavior), and fallback accounting in Decisions
- [ ] Add focused Step 3 tests for converter tag/omitempty/deep-copy/fallback behavior and walker provenance/debug/scale semantics
- [ ] Remove marshal round-trip from `marshalToJSONValue` on happy path
- [ ] Collapse five near-duplicate recursive walkers
- [ ] Preserve every existing path predicate's semantics

### Step 4: Adjacent P2 cleanups

**Status:** ⏳ Not started

- [ ] Move `defaultScaleLabels` to `internal/response/scales.go`
- [ ] Extract common helper shared by `shapeRow` / `shapeWrapperRow`

### Step 5: Verify byte-identical output

**Status:** ⏳ Not started

- [ ] Re-run snapshot fixtures; diff must be empty
- [ ] If diff non-empty, stop and resolve

### Step 6: Build / test / lint

**Status:** ⏳ Not started

- [ ] `make build` / `test` / `test-race` / `lint`
- [ ] Eyeball-benchmark large `include_full` response; must not regress

---

## Decisions

**Step 2 decision:** Use the fallback **single visitor walker** plus a reflection-based JSON-value builder for typed DTOs. Full typed shaping is not selected because `internal/response` cannot import tool DTOs without cycles and mirroring every response envelope/row would balloon the diff and duplicate `internal/tools` / `internal/intervals` contracts.

**Rationale:** This keeps the diff M-sized and the blast radius inside `internal/response`: public `response.Shape` call sites stay stable, `include_full` payloads remain ordinary map/slice values passed through the same shaper, and JSON tag / `omitempty` behavior is preserved centrally by reflecting typed DTOs into JSON-shaped maps before visitor passes. A narrow marshal/unmarshal fallback may remain only for custom `json.Marshaler` or unsupported reflection cases; if used, it is outside the normal tool DTO happy path and will be documented after Step 3.

**Visitor sketch:** Introduce one recursive helper over JSON-shaped values, e.g. `walkJSON(value any, path string, visitor jsonVisitor) (any, []string)` with `jsonVisitor` returning a keep/drop decision and optional missing paths. Predicate/action helpers remain small: `debugPathPredicate`, `provenancePathPredicate`, `provenanceFetchedAtPredicate`, `stripNullVisitor`, `dropDebugVisitor`, and `scaleCollectVisitor`. Path construction stays dotted/indexed via the existing `joinPath` and array-index formatting so golden missing-field paths remain byte-identical.

**Marshal replacement / package-boundary plan:** Replace `marshalToJSONValue` with `toJSONValue` implemented in `internal/response` using reflection: maps/slices/arrays recurse directly; structs honor exported fields, embedded fields, `json:"-"`, renamed fields, and `omitempty`; pointers/interfaces unwrap or become nil; primitives remain primitives. This covers typed tool DTOs (`get_activities`, `get_fitness`, athlete profile, stream/detail envelopes) without importing `internal/tools`. `include_full` maps such as activity `Full`, curve raw payloads, and training summary raw rows remain `map[string]any` / `[]any` and are not decoded/re-encoded. `dropDebugVisitor` must drop ordinary `fetched_at` / `query_type` only outside provenance; `provenanceFetchedAtPredicate` preserves `_meta.provenance.<field>.fetched_at` and keeps it out of debug filtering; scale collection must still skip nested `_meta` content.

**Potential fallback:** If Step 3 encounters a custom `json.Marshaler` value that cannot be represented without calling its marshaler, retain a tiny fallback for that value class only and record it here. No normal tool response fixture should take that fallback.

**Step 3 guardrails:** `toJSONValue` must allocate fresh maps and slices for every container it returns, including nested `Full` / `include_full` raw maps, so shaper mutations never touch caller-owned inputs. Fast path scope is plain structs, maps with string keys, slices/arrays, pointers/interfaces, and JSON primitives used by tool DTOs; unsupported values must fail with a wrapped error comparable to the current `json.Marshal` failure path. Values implementing `json.Marshaler` / `encoding.TextMarshaler` may use a per-value marshal fallback; `json.RawMessage` must be decoded as JSON (or nil for nil raw messages), not reflected as `[]byte`; numeric values may retain their concrete Go numeric type internally as long as canonical JSON bytes match the Step 1 goldens. Step 3 tests must cover tags/renames, `json:"-"`, `omitempty`, deep-copy/no-mutation, fallback/raw-message behavior, null stripping, debug removal, provenance `fetched_at`, and scale collection skipping nested `_meta`.

## Notes

_Add notes as work progresses._

| 2026-05-15 18:01 | Plan review R008 | Added Step 3 guardrails: deep-copy ownership, converter fallback scope, fallback accounting, and focused converter/walker tests. |
| 2026-05-15 17:55 | Plan review R005 | Added Step 2 planning item to specify the exact marshal replacement, package-boundary strategy, include_full fit, and predicate preservation. |
| 2026-05-15 17:49 | Code review R003 | Added blocking revision items: isolate catalog runtime state and use typed DTO inputs for activity/fitness snapshots so JSON tags and omitempty are locked. |
| 2026-05-15 17:44 | Plan review R001 | Added blocking Step 1 plan items: deterministic named fixtures, automated regeneration/comparison, stable metadata setup, canonical JSON; using synthetic fixtures to avoid network/tool import cycles. |
| 2026-05-15 17:43 | Task started | Runtime V2 lane-runner execution |
| 2026-05-15 17:43 | Step 1 started | Snapshot pre-refactor output |
| 2026-05-15 17:47 | Review R001 | plan Step 1: REVISE |
| 2026-05-15 17:49 | Review R002 | plan Step 1: APPROVE |
| 2026-05-15 17:56 | Review R003 | code Step 1: REVISE |
| 2026-05-15 18:00 | Review R004 | code Step 1: APPROVE |
| 2026-05-15 18:03 | Review R005 | plan Step 2: REVISE |
| 2026-05-15 18:08 | Review R006 | plan Step 2: APPROVE |
| 2026-05-15 18:11 | Review R007 | code Step 2: APPROVE |
| 2026-05-15 18:14 | Review R008 | plan Step 3: REVISE |
