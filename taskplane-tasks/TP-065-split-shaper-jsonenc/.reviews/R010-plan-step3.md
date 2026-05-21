# Plan Review: TP-065 Step 3 — Mechanical split

## Verdict: approved

The revised Step 3 plan now addresses the gaps from R009. It keeps the task focused on a behavior-preserving split, makes the `jsonenc` package boundary explicit enough for implementation, preserves the exported `internal/response` API, separates walker / shape / meta extraction into independently testable moves, and tracks the required final `shaper.go` cleanup plus `CHANGELOG.md` update.

I re-read the task prompt and current status, and spot-checked the baseline: TP-043 and TP-047 are complete, `internal/response/shaper.go` still contains the post-TP-047 encoder/walker/shaper/meta declarations, `defaultScaleLabels` remains in `scales.go`, and `internal/response/jsonenc/doc.go` already exists as the target subpackage. The Step 2 evidence for using a subpackage is consistent with the surviving 273 LOC encoder block.

## Execution notes

- Treat the first two Step 3 bullets as one atomic subpackage extraction for testing purposes: move the encoder declarations, switch them to `package jsonenc`, add the single narrow exported conversion function, update `response.Shape` to call it, run gofmt/goimports, then run the full test suite. Testing between the move and the call-site update would only exercise a knowingly uncompilable intermediate state.
- Keep the new `jsonenc` exported surface to exactly the one conversion entry point needed by `response.Shape`; all reflection/cycle/tag/fallback helpers should stay unexported in the subpackage. Do not add any exported names to `internal/response` itself.
- Preserve the existing error wrapping/text for the encoder move where tests cover it, especially the `marshaling response value:` / `unmarshaling response value:` paths from the TP-047 regressions.
- Keep `RegisteredScaleLabels` in `scales.go`; Step 3 should not reclassify scale registry ownership while moving scale `_meta` enrichment into `meta.go`.
- The final `shaper.go` cleanup item is important: after the split, it should either be gone or contain only the accepted package-doc/public-entry-point residue, with no duplicate legacy declarations.

With those execution details followed, the plan is safe to implement.
