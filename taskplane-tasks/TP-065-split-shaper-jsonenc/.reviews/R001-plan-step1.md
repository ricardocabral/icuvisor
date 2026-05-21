# Plan Review: TP-065 Step 1 — Re-baseline after TP-043 + TP-047

## Verdict: needs refinement before implementation

The Step 1 direction is right, but the current plan in `STATUS.md` only repeats the prompt checklist. For this task, the re-baseline is the safety rail for a purely mechanical split, so it should be more explicit about how dependencies are proven landed and what artifact the inventory will produce.

I checked the current worktree: TP-043 and TP-047 status files are marked complete, `internal/response/shaper.go` is still 771 LOC, `internal/response/scales.go` already exists, `toJSONValue` and `walkJSON` both survived, and the old delete-mode/toolset setters are not present. That is good evidence that the task can proceed, but Step 1 should record this evidence rather than just checking a box.

## Required plan adjustments

1. **Define the dependency confirmation evidence.**
   Record concrete checks in `STATUS.md`, for example:
   - TP-043 and TP-047 `STATUS.md` show complete.
   - Code-level markers match the expected post-merge state: `Options` carries `DeleteMode` / `Toolset`, no `SetDeleteMode` / `SetToolset`, `defaultScaleLabels` is already in `scales.go`, `walkJSON` exists, and `marshalToJSONValue` no longer does the old whole-response marshal/unmarshal round trip.
   - Working tree is otherwise clean before the split, except task/status/review files.

2. **Specify the inventory format before collecting it.**
   The plan should say that every top-level `const`, `var`, `type`, and `func` in `internal/response/shaper.go` will be listed in `STATUS.md` with:
   - current name and kind,
   - proposed target file,
   - whether it is exported/public API, test-only, or internal helper,
   - any ambiguity/deferred decision for Step 2.

3. **Call out the already-known ambiguous declarations.**
   The current file contains concerns that do not map one-to-one without a decision. The Step 1 inventory should explicitly avoid silently changing behavior for these:
   - `catalogRuntime`, `SetRuntimeCatalogMetadata`, and test reset helpers are pre-existing catalog metadata state/API; move/classify them as meta concerns, do not treat TP-065 as permission to delete or redesign them.
   - `RegisteredScaleLabels` and `defaultScaleLabels` straddle scale registry vs `_meta` enrichment; note whether the public function stays with `scales.go` or moves to `meta.go`.
   - Path helpers/predicates (`joinPath`, `indexPath`, debug/provenance/meta predicates) should be assigned deliberately between `walk.go`, `shape.go`, and `meta.go` instead of being swept into whichever file currently uses them first.

4. **Baseline the public surface and callers.**
   Because acceptance requires no public API change, Step 1 should include a quick exported-symbol/caller inventory for `response.Shape`, `Options`, `SetRuntimeCatalogMetadata`, and `RegisteredScaleLabels` before any move. This is especially useful if imports need to change in tests or generated catalog code later.

## Suggested Step 1 outcome

Before moving to Step 2, `STATUS.md` should contain a table similar to:

| Decl | Kind | Target | Notes |
| --- | --- | --- | --- |
| `Options` | type | `shape.go` | exported public shaping option surface |
| `Shape` | func | `shape.go` | exported public entry point |
| `marshalToJSONValue` / `toJSONValue` / helpers | funcs/types | `marshal.go` or `jsonenc/` pending Step 2 | include LOC evidence for subpackage decision |
| `jsonWalkContainer` / `walkJSON` / walker types | types/func | `walk.go` | tree-walker primitive |
| `addCommonMeta` / catalog/schema/unit/scale helpers | funcs/vars | `meta.go` or `scales.go` as documented | no `_meta` semantic changes |

With those details added, the Step 1 plan will be strong enough to keep the later split mechanical and reviewable.
