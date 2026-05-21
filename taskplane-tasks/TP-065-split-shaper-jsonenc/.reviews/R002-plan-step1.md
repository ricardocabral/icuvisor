# Plan Review: TP-065 Step 1 — Re-baseline after TP-043 + TP-047

## Verdict: approve

The Step 1 plan in `STATUS.md` now has the right level of specificity for a safe mechanical split. It addresses the prior review's gaps by requiring concrete dependency evidence, a public-surface/caller baseline, a declaration-by-declaration inventory format, and explicit handling of the ambiguous response/meta/walker helpers before any files are moved.

I spot-checked the current baseline:

- `TP-043` and `TP-047` status files are both marked complete.
- `internal/response/shaper.go` still contains the expected post-TP-047 declarations, including `Options` with `DeleteMode`/`Toolset`, `Shape`, `marshalToJSONValue`/`toJSONValue`, `walkJSON`, shaping helpers, and meta helpers.
- `internal/response/scales.go` already owns `defaultScaleLabels`, so the plan's scale/meta classification callout is necessary.
- The old whole-response marshal/unmarshal path is gone from the visible baseline; the remaining converter is reflection-based with narrow marshal fallbacks.
- The working tree is clean apart from Taskplane review state/files, which is acceptable for this plan review.

## Why this is sufficient

Step 1 is not supposed to move code yet; it is supposed to produce the map that makes the later split reviewable. The current checklist now requires exactly that artifact in `STATUS.md`: every top-level declaration in `shaper.go` with kind, target file, public/internal/test-only status, and ambiguity/deferred notes. That should prevent accidental API changes or semantic reshuffling in Step 3.

## Non-blocking execution notes

When implementing Step 1, keep the inventory purely observational:

- Do not decide the `jsonenc/` subpackage in this step beyond marking marshal declarations as pending Step 2.
- Record the exact commands or grep results used for dependency confirmation so future reviewers do not need to reconstruct the baseline.
- Include all top-level declaration groups, including the second `const` block for `jsonWalkContainer` values and the anonymous `catalogRuntime` var.

No further plan changes are required before carrying out Step 1.
