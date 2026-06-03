# Plan Review — Step 2

Result: Approved.

The Step 2 checklist is acceptable because it builds on the detailed Step 1 discoveries: compact `workout_doc_summary.target_previews`, profile reuse without extra API calls, sport/indoor threshold matching, explicit power/HR/pace formulas, and fail-closed omission rules.

Carry these expectations into implementation/tests:

1. Add positive tests for supported `% FTP`, `% LTHR`/`% HR`, and threshold-pace previews; do not treat the already-documented omissions as a substitute for tests on supported families.
2. Freeze the exact compact row shape in assertions, including `target`, `preview`, `basis`, `path`, and omission of `target_previews` when nothing resolves.
3. Cover both event and workout row helpers enough to catch call-site drift, plus `include_full:false`/`true` regressions showing raw workout docs are still controlled by existing flags.
4. Verify profile data is reused from the existing handler fetch and that missing/unmatched/zero thresholds produce no misleading preview.

No blocker to Step 2 implementation.
