# Plan Review — Step 1

Result: Approved.

The revised Step 1 discoveries address the prior review: the preview is placed under `workout_doc_summary`, unsupported/omission behavior is fail-closed, profile reuse avoids extra API calls, sport/indoor threshold matching is documented, and conversion semantics for power/HR/pace are explicit. Targeted baseline tests also pass:

```text
go test ./internal/tools ./internal/workoutdoc
ok   github.com/ricardocabral/icuvisor/internal/tools
ok   github.com/ricardocabral/icuvisor/internal/workoutdoc
```

Carry these minor clarifications into Step 2 implementation/tests:

1. Freeze concrete field types/format for `target_previews` fields (`step`, `path`, `basis`, `repeat_reps`) in tests so the shape does not drift.
2. Include at least one pace preview test that locks the rendered label/unit, not just the formula.
3. Since the plan intentionally updates shared row helpers, add regression coverage for the additional affected paths or explicitly document any helper call site that remains preview-free.

No blocker to implementation.
