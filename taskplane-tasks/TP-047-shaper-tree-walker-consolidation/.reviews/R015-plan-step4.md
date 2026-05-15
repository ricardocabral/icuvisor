# Plan Review — Step 4: Adjacent P2 cleanups

**Decision: APPROVE.**

The updated Step 4 plan addresses the R014 gap. It now defines a narrow helper boundary for the `shapeRow` / `shapeWrapperRow` extraction instead of trying to merge wrapper traversal, row-collection handling, and null/missing-field setup into one larger routine.

## Why this is safe to proceed

- `defaultScaleLabels` is planned as a pure move to `internal/response/scales.go` in the same package, preserving the existing unexported registry and the defensive-copy behavior of `RegisteredScaleLabels`.
- The helper is scoped to shared row finalization only: debug handling, missing-field strip metadata, scale metadata, and optional common metadata.
- Wrapper-specific behavior remains outside the helper: `RowCollections` traversal, top-level nil handling, and wrapper-level common metadata.
- `shapeRow` keeps the `include_full` vs terse setup outside the helper, which avoids conflating payload shaping with final metadata decoration.
- Nested row collections are explicitly called out as continuing to avoid common/debug metadata additions.

## Implementation reminders

- Preserve current debug semantics exactly: the new debug/helper flag should control whether debug metadata is *added*, but existing debug fields still need to be dropped when `opts.DebugMetadata` is false, including nested rows.
- Keep `addStripMeta` gated by `!opts.IncludeFull && len(missing) > 0` exactly as today.
- Keep `addScaleMeta` behavior unchanged for both wrapper and normal rows.
- After the refactor, run `go test ./internal/response` before Step 5; Step 5 should then prove the golden output diff is empty.
