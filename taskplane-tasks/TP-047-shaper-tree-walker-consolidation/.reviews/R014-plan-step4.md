# Plan Review — Step 4: Adjacent P2 cleanups

**Decision: REVISE.**

Step 4 is low risk, but the current `STATUS.md` update only marks the step in progress and repeats the two task checkboxes. It does not yet define the helper boundary for the `shapeRow` / `shapeWrapperRow` extraction, which is the only part of this step that can accidentally drift public `_meta`, `include_full`, or debug-metadata behavior.

## Blocking plan gap

1. **Specify the helper extraction shape before implementing it.**  
   `shapeWrapperRow` and `shapeRow` share the final metadata/post-processing block, but their setup semantics differ (`internal/response/shaper.go:464-541`):
   - wrapper rows handle `RowCollections` and preserve top-level missing paths like `key`;
   - nested row collections call `shapeRows(..., includeCommonMeta=false)`, so they must continue to omit common/debug metadata;
   - `shapeWrapperRow` always adds wrapper-level common metadata, while `shapeRow` gates common/debug metadata on `includeCommonMeta`.

   Please add a short Step 4 plan note in `STATUS.md` describing a narrow helper such as `finalizeShapedRow(row, missing, opts, includeDebugMetadata, includeCommonMeta) map[string]any` (name is flexible). The plan should make clear that row-collection traversal and wrapper-specific null handling stay outside the helper, and that the existing call semantics remain equivalent, e.g. wrapper calls with debug/common enabled and nested rows call with both disabled.

## Non-blocking implementation reminders

- Move `defaultScaleLabels` to `internal/response/scales.go` in the same `package response`; keep it unexported and keep `RegisteredScaleLabels` returning a defensive copy.
- Do not regenerate golden fixtures for Step 4 unless a test harness requires formatting changes. This should be a pure code move/refactor with Step 5 proving an empty output diff.
- After the helper extraction, run at least `go test ./internal/response` before moving to Step 5, because this area is covered by the golden snapshot tests.

Once the helper boundary and preservation notes are added to `STATUS.md`, the Step 4 plan should be safe to proceed.
