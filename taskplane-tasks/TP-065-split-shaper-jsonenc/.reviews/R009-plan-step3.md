# Plan Review: TP-065 Step 3 — Mechanical split

## Verdict: needs refinement before implementation

The Step 3 plan is directionally correct and uses the Step 1 inventory well, but it is still too terse for the one non-mechanical part of this step: moving the JSON encoder into a real Go subpackage. Splitting declarations within `package response` is mostly copy/move work; extracting `internal/response/jsonenc/` requires an explicit package boundary, because `response.Shape` cannot call unexported declarations from that subpackage.

I spot-checked the current baseline: `internal/response/shaper.go` is still 771 LOC, `toJSONValue`/`marshalToJSONValue` and `walkJSON` remain in the file, `defaultScaleLabels` already lives in `scales.go`, and `internal/response/jsonenc/doc.go` exists. The Step 2 decision to use `jsonenc/` is consistent with the recorded 273 LOC marshalling block, but Step 3 should say exactly how that boundary will preserve the task's “no public API change / no new public response surface” requirement.

## Required plan adjustments

1. **Define the `jsonenc` package boundary before moving code.**
   The plan should specify the single entry point that `response.Shape` will call after extraction, and how it is scoped. For example: keep all encoder helpers unexported in `internal/response/jsonenc`, expose at most one narrowly documented package function for converting a value to a JSON-shaped tree, and keep a package-local `response` wrapper only if needed for naming compatibility. Also note that no new exported names are added to `internal/response` itself.

2. **Call out the unavoidable non-mechanical edit caused by the subpackage.**
   The JSON encoder move will require changing `package response` to `package jsonenc`, updating imports, and changing `Shape` to call the subpackage entry point. That is acceptable, but the plan should identify it explicitly so reviewers can distinguish this boundary edit from behavior changes. Error strings and JSON conversion behavior should remain byte-for-byte equivalent where tests cover them.

3. **Tighten the test cadence to match the prompt.**
   The prompt says to run the full test suite after each file move. The current third bullet combines `shape.go` and `meta.go` and says to test after “these file moves.” Either split that into separate moves/test runs, or state why they will be moved as one atomic chunk due to mutual references and still run the full suite immediately after that atomic move.

4. **State the final `shaper.go` outcome.**
   Add a Step 3 checklist item that `shaper.go` will be deleted or left only with the package doc/public shaping entry point, as required by the acceptance criteria. This prevents a successful extraction from leaving an empty or duplicate legacy file behind.

## Non-blocking execution notes

- Keep `RegisteredScaleLabels` with `scales.go` as recorded in the Step 1 inventory; Step 3 should not reclassify scale registry ownership.
- After each move, use `gofmt`/`goimports` so import-only diffs do not accumulate across files.
- If tests are split, keep that secondary to the source split; preserving the existing golden behavior is more important than reorganizing tests in this task.

Once these details are added to `STATUS.md`, the mechanical split plan should be safe to execute.
