# Code Review — Step 4: Adjacent P2 cleanups

**Decision: APPROVE.**

## Findings

No blocking or non-blocking code findings for this step.

## Notes

- `defaultScaleLabels` was moved to `internal/response/scales.go` without changing package visibility or the defensive-copy behavior of `RegisteredScaleLabels`.
- `finalizeShapedRow` keeps the helper boundary narrow and preserves the previous `shapeRow` / `shapeWrapperRow` semantics:
  - wrapper rows still get wrapper traversal and top-level nil handling before finalization;
  - normal rows still handle `include_full` vs terse shaping before finalization;
  - nested row collections still avoid added common/debug metadata when `includeCommonMeta` is false;
  - debug stripping still runs when `opts.DebugMetadata` is false regardless of helper flags;
  - strip metadata remains gated by `!opts.IncludeFull && len(missing) > 0`.

## Verification

- Ran `go test ./internal/response` — passed.
