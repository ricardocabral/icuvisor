# Plan Review — Step 2

Verdict: Approve with required tightenings.

The Step 2 plan is aligned with the task: it strengthens assistant-facing preview guidance, keeps `validate_workout` as a read-only preflight aid, and explicitly avoids model-controlled confirmation or safety-gate bypasses.

Tighten the plan before/during implementation:

1. If tool descriptions/schema examples change, regenerate and/or update the generated catalog artifacts as part of this task path, not only at full-suite time: `web/data/tools.json` and `cmd/gendocs/testdata/tools.golden.json` may need updates, and `go test ./cmd/gendocs` should be included in the targeted verification.
2. Add regression coverage for the new guidance where practical. Existing prompt tests check generic approval wording, but Step 2 should assert the specific preview content: total duration, key steps, target intensities, load/distance/time deltas, preserved fields, and `validate_workout` for uncertain DSL/structured changes.
3. Keep JSON `input_examples` schema-valid. Do not add pseudo-fields like `preview`, `approval`, or `confirm`; put human-readable preview/approval instructions in prompt text and tool/schema descriptions.
4. Preserve current write semantics: no mandatory server-side validation precondition, no runtime confirmation argument, and no changes to delete/write registration gates.
5. Update `STATUS.md` bookkeeping before handoff; the header still reports Step 1 even though Step 2 is in progress.

No other blocking plan issues found.
