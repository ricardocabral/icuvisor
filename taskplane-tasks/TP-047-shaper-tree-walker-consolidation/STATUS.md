# TP-047-shaper-tree-walker-consolidation — Status

**Current Step:** Step 4: Adjacent P2 cleanups
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-15
**Review Level:** 2
**Review Counter:** 16
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

**Status:** ✅ Complete

- [x] Add R008 guardrails: fresh-copy ownership for all JSON containers (including `include_full` maps), explicit converter fallback scope (`json.Marshaler`, `encoding.TextMarshaler`, `json.RawMessage`, unsupported values, numeric behavior), and fallback accounting in Decisions
- [x] Add focused Step 3 tests for converter tag/omitempty/deep-copy/fallback behavior and walker provenance/debug/scale semantics
- [x] Remove marshal round-trip from `marshalToJSONValue` on happy path
- [x] Collapse five near-duplicate recursive walkers
- [x] Preserve every existing path predicate's semantics
- [x] Fix R010 float conversion semantics: reject NaN/Inf early with wrapped errors and preserve float32 JSON byte behavior via narrow fallback, with regression tests
- [x] Fix R010 provenance debug semantics so `_meta.provenance.<field>.query_type` is preserved and make `_meta` path matching segment-exact, with regression tests
- [x] Fix R011 `json.Number` semantics by routing valid/invalid numbers through the narrow JSON fallback with regression tests
- [x] Fix R012 struct field semantics by falling back for duplicate JSON field names and unsupported tag options such as `,string`, with regression tests
- [x] Fix R012 cycle handling so self-referential maps/slices/pointers return wrapped JSON errors instead of recursing unbounded, with regression tests

### Step 4: Adjacent P2 cleanups

**Status:** ✅ Complete

- [x] Address R014 plan note: helper only finalizes row metadata/debug/scale/common meta; wrapper traversal, RowCollections, and wrapper-specific null handling stay outside; nested row collections call helper with common/debug disabled
- [x] Move `defaultScaleLabels` to `internal/response/scales.go`
- [x] Extract common helper shared by `shapeRow` / `shapeWrapperRow`

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

**Step 4 helper plan:** Extract a narrow `finalizeShapedRow(row, missing, opts, includeDebugMetadata, includeCommonMeta) map[string]any` helper for the shared post-processing only: debug drop/add, missing-field strip metadata, scale metadata, and optional common metadata. `shapeWrapperRow` keeps RowCollections traversal and top-level nil/missing handling outside the helper, then calls it with debug/common enabled. `shapeRow` keeps include_full vs terse setup outside the helper and passes `includeCommonMeta` through; nested row collections still call `shapeRow(..., includeCommonMeta=false)` so they do not gain common/debug metadata.

**Step 3 fallback accounting:** The old whole-response marshal/unmarshal round-trip is removed from the happy path. Narrow per-value marshal fallbacks remain for values that depend on `encoding/json` contracts: `json.Marshaler` (including `json.RawMessage` and `time.Time`), `encoding.TextMarshaler`, `[]byte`, non-string map keys, and anonymous embedded structs where field-promotion semantics would otherwise be reimplemented. The Step 1 typed activity/fitness/wrapper/provenance golden fixtures do not rely on a whole-response fallback; only their special leaf values would use per-value fallback if present.

**Step 3 guardrails:** `toJSONValue` must allocate fresh maps and slices for every container it returns, including nested `Full` / `include_full` raw maps, so shaper mutations never touch caller-owned inputs. Fast path scope is plain structs, maps with string keys, slices/arrays, pointers/interfaces, and JSON primitives used by tool DTOs; unsupported values must fail with a wrapped error comparable to the current `json.Marshal` failure path. Values implementing `json.Marshaler` / `encoding.TextMarshaler` may use a per-value marshal fallback; `json.RawMessage` must be decoded as JSON (or nil for nil raw messages), not reflected as `[]byte`; numeric values may retain their concrete Go numeric type internally as long as canonical JSON bytes match the Step 1 goldens. Step 3 tests must cover tags/renames, `json:"-"`, `omitempty`, deep-copy/no-mutation, fallback/raw-message behavior, null stripping, debug removal, provenance `fetched_at`, and scale collection skipping nested `_meta`.

## Notes

_Add notes as work progresses._

| 2026-05-15 18:31 | Plan review R014 | Added Step 4 helper-boundary planning item for shapeRow/shapeWrapperRow extraction. |
| 2026-05-15 18:23 | Code review R012 | Added blocking revision items: preserve duplicate-field/string-tag struct semantics and detect cycles with wrapped JSON errors. |
| 2026-05-15 18:18 | Code review R011 | Added blocking revision item: preserve `json.Number` as a JSON number and reject invalid numbers. |
| 2026-05-15 18:13 | Code review R010 | Added blocking revision items: restore JSON float error/float32 behavior and preserve provenance query_type while tightening _meta segment matching. |
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
| 2026-05-15 18:17 | Review R009 | plan Step 3: APPROVE |
| 2026-05-15 18:29 | Review R010 | code Step 3: REVISE |
| 2026-05-15 18:37 | Review R011 | code Step 3: REVISE |
| 2026-05-15 18:42 | Review R012 | code Step 3: REVISE |
| 2026-05-15 18:59 | Review R014 | plan Step 4: REVISE |
| 2026-05-15 19:01 | Review R015 | plan Step 4: APPROVE |
| 2026-05-15 19:05 | Review R016 | code Step 4: APPROVE |
