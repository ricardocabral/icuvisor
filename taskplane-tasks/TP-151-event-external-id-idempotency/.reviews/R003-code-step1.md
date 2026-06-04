# Code Review: TP-151 Step 1 — Design external_id contract

**Verdict: APPROVE**

No product code or public schemas changed in this step; the diff is limited to task status and prior review artifacts. The Step 1 contract recorded in `STATUS.md` is concrete enough to guide the upcoming implementation, including create/update/omit/clear semantics, deterministic `apply_training_plan` IDs, read exposure, conservative preflight behavior, and upstream uncertainty.

## Findings

None.

## Notes for implementation

Carry forward the R002 notes when coding Step 2/3, especially pinning the hash serialization/digest length in tests and surfacing existing event IDs/warnings when an `external_id` match causes an idempotent skip.
