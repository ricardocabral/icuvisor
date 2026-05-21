# TP-065-split-shaper-jsonenc — Status

**Current Step:** Step 4: Verify
**Status:** ✅ Complete
**Last Updated:** 2026-05-17
**Review Level:** 2
**Review Counter:** 11
**Iteration:** 5
**Size:** M

---

### Step 1: Re-baseline after TP-043 + TP-047

**Status:** ✅ Complete

- [x] Confirm both have landed; if not, block this task. Record evidence from TP-043/TP-047 `STATUS.md`, code-level post-merge markers (`Options` has `DeleteMode`/`Toolset`, no old setters, `scales.go` ownership, `walkJSON` survival, no old whole-response marshal/unmarshal round trip), and working-tree cleanliness.
- [x] Baseline exported `internal/response` surface and callers for `Shape`, `Options`, `SetRuntimeCatalogMetadata`, and `RegisteredScaleLabels` before any move.
- [x] Inventory every top-level decl in the post-merge `shaper.go` and assign each to a target file. Record in `STATUS.md` with kind, target file, public/internal/test-only status, and ambiguity/deferred notes.
- [x] Explicitly classify ambiguous declarations without behavior changes: catalog runtime/test reset helpers as meta, scale label surface versus `scales.go`, and path/debug/provenance/meta predicates across `walk.go`, `shape.go`, and `meta.go`.

## Notes

### Step 1 dependency evidence

- TP-043 `STATUS.md`: task-level status is `✅ Complete`; Step 4 verify is complete, including removal of response globals/setters and no remaining `init()` in `internal/response/`.
- TP-047 `STATUS.md`: task-level status is `✅ Complete`; Step 6 verify is complete, and decisions confirm the whole-response marshal/unmarshal round trip was replaced by reflection with only narrow per-value fallbacks.
- Code markers checked on 2026-05-17: `internal/response/shaper.go` is 771 LOC; `Options` includes `DeleteMode safety.Mode` and `Toolset safety.Toolset`; `walkJSON` exists; `marshalToJSONValue` calls `toJSONValue` instead of whole-response marshal/unmarshal; `defaultScaleLabels` lives in `internal/response/scales.go`; `grep -R "SetDeleteMode\\|SetToolset"` finds no live source references outside historical task files.
- Pre-split working tree check: no source files are modified; only TP-065 `STATUS.md` and plan review artifacts are present.

### Step 1 public surface baseline

- Exported declarations in `internal/response`: `type Options`, `func SetRuntimeCatalogMetadata`, `func RegisteredScaleLabels`, and `func Shape`; `internal/response/scales.go` has no exported declarations.
- `response.Shape` callers: `internal/tools/get_activity_details.go`, `internal/tools/get_activities.go`, `internal/tools/fitness_shared.go`, `internal/tools/get_activity_streams.go`, `internal/tools/update_wellness.go`, `internal/tools/get_activity_messages.go`, and `internal/athleteprofile/profile.go`.
- `response.Options` external construction/use: `internal/tools/registry.go` (`responseShaping.options`), `internal/resources/athlete_profile.go`, and `internal/athleteprofile/profile.go`.
- `response.SetRuntimeCatalogMetadata` caller: `internal/mcp/server.go` only; `resetRuntimeCatalogMetadataForTest` remains package-internal for response tests.
- `response.RegisteredScaleLabels` caller: `internal/tools/update_wellness.go` only, plus package-local tests in `internal/response/shaper_test.go`.

### Step 1 declaration inventory

| Decl                                         | Kind  | Target                     | Surface              | Notes                                                                    |
| -------------------------------------------- | ----- | -------------------------- | -------------------- | ------------------------------------------------------------------------ |
| `defaultCatalogHash`                         | const | `meta.go`                  | internal             | catalog metadata default                                                 |
| `catalogRuntime`                             | var   | `meta.go`                  | internal state       | pre-existing synchronized catalog runtime; keep behavior                 |
| `catalogSnapshot`                            | type  | `meta.go`                  | internal             | catalog metadata snapshot                                                |
| `responseOwnedMetaKeys`                      | var   | `meta.go`                  | internal             | `_meta` overwrite allowlist                                              |
| `Options`                                    | type  | `shape.go`                 | exported public API  | response shaping option surface                                          |
| `SetRuntimeCatalogMetadata`                  | func  | `meta.go`                  | exported public API  | caller remains `internal/mcp/server.go`                                  |
| `resetRuntimeCatalogMetadataForTest`         | func  | `meta.go`                  | internal test helper | package-private helper used by tests                                     |
| `setRuntimeCatalogMetadataForTest`           | func  | `meta.go`                  | internal test helper | package-private helper used by tests                                     |
| `RegisteredScaleLabels`                      | func  | `scales.go`                | exported public API  | registry accessor belongs with `defaultScaleLabels`; see ambiguity notes |
| `Shape`                                      | func  | `shape.go`                 | exported public API  | public entry point unchanged                                             |
| `marshalToJSONValue`                         | func  | `marshal.go` or `jsonenc/` | internal             | Step 2 decides package by LOC                                            |
| `jsonVisit`                                  | type  | `marshal.go` or `jsonenc/` | internal             | cycle tracking for encoder                                               |
| `toJSONValue`                                | func  | `marshal.go` or `jsonenc/` | internal             | Step 2 decides package by LOC                                            |
| `reflectJSONValue`                           | func  | `marshal.go` or `jsonenc/` | internal             | encoder helper                                                           |
| `unsupportedFloatError`                      | func  | `marshal.go` or `jsonenc/` | internal             | encoder error helper                                                     |
| `mapToJSONValue`                             | func  | `marshal.go` or `jsonenc/` | internal             | encoder helper                                                           |
| `sliceToJSONValue`                           | func  | `marshal.go` or `jsonenc/` | internal             | encoder helper                                                           |
| `structToJSONValue`                          | func  | `marshal.go` or `jsonenc/` | internal             | encoder helper                                                           |
| `enterJSONVisit`                             | func  | `marshal.go` or `jsonenc/` | internal             | encoder cycle guard                                                      |
| `jsonField`                                  | func  | `marshal.go` or `jsonenc/` | internal             | JSON tag helper                                                          |
| `isEmptyJSONValue`                           | func  | `marshal.go` or `jsonenc/` | internal             | omitempty helper                                                         |
| `marshalSpecialValue`                        | func  | `marshal.go` or `jsonenc/` | internal             | narrow marshal fallback                                                  |
| `marshalJSONValue`                           | func  | `marshal.go` or `jsonenc/` | internal             | narrow marshal fallback                                                  |
| `canInterface`                               | func  | `marshal.go` or `jsonenc/` | internal             | reflection helper                                                        |
| `jsonWalkContainer`                          | type  | `walk.go`                  | internal             | walker primitive                                                         |
| `walkRoot`, `walkMapValue`, `walkSliceValue` | const | `walk.go`                  | internal             | walker container enum                                                    |
| `jsonWalkDecision`                           | type  | `walk.go`                  | internal             | walker primitive                                                         |
| `jsonWalkVisitor`                            | type  | `walk.go`                  | internal             | walker primitive                                                         |
| `walkJSON`                                   | func  | `walk.go`                  | internal             | single recursive tree walker                                             |
| `shapeRoot`                                  | func  | `shape.go`                 | internal             | selects wrapper vs row shaping                                           |
| `shapeRows`                                  | func  | `shape.go`                 | internal             | row collection shaping                                                   |
| `shapeWrapperRow`                            | func  | `shape.go`                 | internal             | wrapper shaping                                                          |
| `shapeRow`                                   | func  | `shape.go`                 | internal             | row shaping                                                              |
| `finalizeShapedRow`                          | func  | `shape.go`                 | internal             | shape pipeline orchestrator that calls meta helpers                      |
| `addDebugMetadata`                           | func  | `meta.go`                  | internal             | debug `_meta`/row metadata policy                                        |
| `addStripMeta`                               | func  | `meta.go`                  | internal             | missing-field `_meta` enrichment                                         |
| `addScaleMeta`                               | func  | `meta.go`                  | internal             | scale `_meta` enrichment                                                 |
| `scalesForRow`                               | func  | `meta.go`                  | internal             | scale enrichment helper                                                  |
| `collectScaleLabels`                         | func  | `meta.go`                  | internal             | scale traversal policy using walker                                      |
| `addCommonMeta`                              | func  | `meta.go`                  | internal             | common `_meta` enrichment                                                |
| `schemaCatalogMeta`                          | func  | `meta.go`                  | internal             | catalog/version enrichment                                               |
| `schemaChangeMessage`                        | func  | `meta.go`                  | internal             | catalog/version message                                                  |
| `stripNulls`                                 | func  | `shape.go`                 | internal             | terse shaping policy using walker                                        |
| `stripNullVisitor`                           | func  | `shape.go`                 | internal             | terse shaping visitor                                                    |
| `filterDebugMissing`                         | func  | `meta.go`                  | internal             | debug metadata policy                                                    |
| `isDebugPath`                                | func  | `meta.go`                  | internal             | debug path predicate                                                     |
| `dropDebugMetadata`                          | func  | `meta.go`                  | internal             | debug metadata policy using walker                                       |
| `dropDebugVisitor`                           | func  | `meta.go`                  | internal             | debug visitor                                                            |
| `isProvenancePath`                           | func  | `meta.go`                  | internal             | provenance debug-preservation predicate                                  |
| `isProvenanceFetchedAtPath`                  | func  | `meta.go`                  | internal             | provenance debug-preservation predicate                                  |
| `cloneMap`                                   | func  | `shape.go`                 | internal             | shape helper                                                             |
| `normalizeVersion`                           | func  | `meta.go`                  | internal             | catalog/common metadata normalization                                    |
| `normalizeCatalogHash`                       | func  | `meta.go`                  | internal             | catalog metadata normalization                                           |
| `joinPath`                                   | func  | `walk.go`                  | internal             | walker path construction                                                 |
| `indexPath`                                  | func  | `walk.go`                  | internal             | walker path construction                                                 |
| `isMetaPath`                                 | func  | `meta.go`                  | internal             | metadata traversal boundary predicate                                    |
| `rowCollectionSet`                           | func  | `shape.go`                 | internal             | shape helper                                                             |
| `presentFields`                              | func  | `meta.go`                  | internal             | strip metadata helper                                                    |
| `sortedStrings`                              | func  | `meta.go`                  | internal             | strip metadata helper                                                    |

### Step 1 ambiguous declaration decisions

- Catalog runtime state/API (`defaultCatalogHash`, `catalogRuntime`, `catalogSnapshot`, `SetRuntimeCatalogMetadata`, `resetRuntimeCatalogMetadataForTest`, `setRuntimeCatalogMetadataForTest`, `schemaCatalogMeta`, `schemaChangeMessage`, normalizers) is metadata concern and moves to `meta.go`; TP-065 will not redesign or delete the pre-existing process-global catalog metadata state.
- Scale registry accessor `RegisteredScaleLabels` stays with `defaultScaleLabels` in `scales.go` because it is the exported registry surface, while scale `_meta` enrichment (`addScaleMeta`, `scalesForRow`, `collectScaleLabels`) moves to `meta.go`.
- Walker primitive declarations and path constructors (`jsonWalk*`, `walkJSON`, `joinPath`, `indexPath`) move to `walk.go`; visitor policies remain with their owning concerns.
- Null stripping (`stripNulls`, `stripNullVisitor`) is terse response shaping policy and moves to `shape.go`, even though it uses the walker.
- Debug/provenance predicates and visitors (`filterDebugMissing`, `isDebugPath`, `dropDebugMetadata`, `dropDebugVisitor`, `isProvenancePath`, `isProvenanceFetchedAtPath`) are metadata policy and move to `meta.go` to preserve provenance behavior without logic changes.
- General row/wrapper helpers (`shapeRoot`, `shapeRows`, `shapeWrapperRow`, `shapeRow`, `finalizeShapedRow`, `cloneMap`, `rowCollectionSet`) move to `shape.go`; `finalizeShapedRow` stays there as pipeline orchestration and calls metadata helpers rather than owning `_meta` details.

### Step 2: Decide on subpackage

**Status:** ✅ Complete

- [x] If `toJSONValue` survives TP-047 and > 200 LOC: create `internal/response/jsonenc/`. Otherwise: keep in `internal/response/marshal.go`.
- [x] Document the decision in `STATUS.md` with the line-count evidence.

### Step 2 decision evidence

- `toJSONValue` survived TP-047 in `internal/response/shaper.go`; the current marshalling block from `marshalToJSONValue` through `canInterface` is 273 LOC (`awk` over lines 98-370 on 2026-05-17), exceeding the >200 LOC threshold.
- Decision: use `internal/response/jsonenc/` for the hand-rolled JSON encoding helpers rather than keeping them in `internal/response/marshal.go`.

### Step 3: Mechanical split

**Status:** ✅ Complete

- [x] Define and implement the `jsonenc` package boundary: move encoder declarations into `internal/response/jsonenc/`, expose only one narrow package entry point for JSON-tree conversion, keep helpers unexported, add no new exported names to `internal/response`, and run the full test suite after this file move. Evidence: `go test ./...` passed after extracting `jsonenc.Encode` on 2026-05-17.
- [x] Apply the unavoidable subpackage boundary edit without behavior changes: update package/imports and `Shape`'s call site while preserving covered error strings and JSON conversion semantics.
- [x] Extract walker declarations into `internal/response/walk.go` with no logic changes; run the full test suite after this file move. Evidence: `go test ./...` passed after extracting `walk.go` on 2026-05-17.
- [x] Extract shaping declarations into `internal/response/shape.go` with no logic changes; run the full test suite after this file move. Evidence: `go test ./...` passed after extracting `shape.go` on 2026-05-17.
- [x] Extract metadata declarations into `internal/response/meta.go` with no logic changes; run the full test suite after this file move. Evidence: `go test ./...` passed after extracting `meta.go` on 2026-05-17.
- [x] Delete `internal/response/shaper.go` or leave it only with package doc/public shaping entry point, with no duplicate legacy declarations. Evidence: moved `RegisteredScaleLabels` to `scales.go`, removed `shaper.go`, and `go test ./...` passed on 2026-05-17.
- [x] Update `CHANGELOG.md` `[Unreleased]` with the internal response split.

### Step 3 plan review adjustments

- R009 required the `jsonenc` subpackage boundary to be explicit: `response.Shape` will call a single narrow exported `jsonenc` conversion function, while all encoder helpers stay unexported inside `internal/response/jsonenc/` and `internal/response` gains no new exported API.
- The only intended non-mechanical behavior-neutral edit is the package-boundary call-site/import change needed for `response` to use `jsonenc`.
- Shape and metadata extraction are now separate file moves with a full test-suite run after each move, matching the task cadence.
- The final `shaper.go` outcome is tracked explicitly to avoid leaving duplicate legacy declarations behind.

### Step 4: Verify

**Status:** ✅ Complete

- [x] Run `make build`, `make test`, `make test-race`, and `make lint` and record pass evidence. Evidence: all four commands passed in sequence on 2026-05-17; `make lint` reported `0 issues`.
- [x] Run `go run scripts/snapshot_tool_schemas.go` and confirm the schema snapshot diff is empty. Evidence: command passed on 2026-05-17; `git diff --name-only` showed only `STATUS.md`.
- [x] Run `wc -l internal/response/*.go internal/response/jsonenc/*.go` and confirm each focused source file is ≤ ~300 LOC. Evidence: focused non-test files are `meta.go` 257, `scales.go` 23, `shape.go` 163, `time.go` 37, `units.go` 255, `walk.go` 73, `jsonenc/jsonenc.go` 285, and doc files 13/2 LOC; test files are excluded from the focused source-file limit.

| 2026-05-17 04:16 | Task started | Runtime V2 lane-runner execution |
| 2026-05-17 04:16 | Step 1 started | Re-baseline after TP-043 + TP-047 |
| 2026-05-17 04:19 | Review R001 | plan Step 1: UNKNOWN |
| 2026-05-17 04:21 | Review R002 | plan Step 1: APPROVE |

| 2026-05-17 06:19 | Worker iter 1 | killed (wall-clock timeout) in 7383s, tools: 43 |

| 2026-05-17 06:52 | Worker iter 2 | done in 2016s, tools: 8 |
| 2026-05-17 07:59 | Review R009 | plan Step 3: UNKNOWN |
| 2026-05-17 08:02 | Review R010 | plan Step 3: APPROVE |

| 2026-05-17 08:36 | Worker iter 3 | done in 6213s, tools: 36 |

| 2026-05-17 09:33 | Worker iter 4 | done in 3432s, tools: 19 |
| 2026-05-17 09:54 | Review R011 | code Step 3: APPROVE |

| 2026-05-17 09:57 | Worker iter 5 | done in 1399s, tools: 50 |
| 2026-05-17 09:57 | Task complete | .DONE created |