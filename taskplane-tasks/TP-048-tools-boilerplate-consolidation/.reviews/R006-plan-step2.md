# Plan Review — Step 2: Mechanical replacement across tool files

## Verdict

APPROVE

The revised Step 2 plan addresses the remaining blocker from R005. It now explicitly calls out the old package-local `decodeStrict(raw, &args)` callers, preserves their current empty/whitespace `arguments must be a JSON object` behavior via minimal prechecks where needed, preserves bespoke raw-field validation ordering, removes the old helper when unused, and keeps `TextResult` replacements constrained to exact/safe text-result constructions.

## Notes to carry into implementation

- For each current `decodeStrict(raw, &args)` caller, decide before editing whether empty/whitespace was previously rejected or explicitly allowed. Add only the minimal precheck needed to preserve that behavior before calling `DecodeStrict[T]`.
- Keep special decoders such as `decodeGetActivitiesRequest`, `decodeActivityReadRequest`, and raw-field validation paths behaviorally identical except for replacing the strict decode core.
- Remove the old unexported `decodeStrict` function once the package no longer calls it; do not leave it as a wrapper.
- Use `TextResult` only for exact text-content result construction where marshal failure is already ignored or impossible by construction. Checked response-encoding paths should only be converted where the plan's JSON-marshalable-by-construction assumption truly holds.
- Run the proposed targeted verification before moving to Step 3: `go test ./internal/tools` plus greps for `DisallowUnknownFields`, `decodeStrict(`, and `ContentTypeText`.

No further plan revisions are required for Step 2.
