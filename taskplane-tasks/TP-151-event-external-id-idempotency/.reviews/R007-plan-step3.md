# Plan Review: TP-151 Step 3 — Make apply_training_plan retry-safer

**Verdict: APPROVE**

The revised Step 3 plan now captures the main idempotency risks called out in R006. In particular, it explicitly plans to:

- add a deterministic helper with canonical tuple serialization and pinned digest length;
- thread `plan_id`, `start_date`, `workout_id`, `relative_day`, and `event_date` into event creation;
- set stable `WriteEventParams.ExternalID` values for plan-created events;
- protect same-day matching-`external_id` conflicts before `replace_existing` can delete anything;
- expose the hashed `external_id` in dry-run proposed rows; and
- add stability/protection tests plus run targeted `ApplyTrainingPlan` tests.

This is sufficient to proceed with implementation and stays within the intended file scope.

## Non-blocking implementation note

When pinning the helper contract, include at least one negative/stability test that changing `start_date` or `event_date` changes the generated ID, and one assertion that dry-run and actual apply use the same ID for the same tuple. This is implied by the checklist, but worth making explicit in the test cases.
