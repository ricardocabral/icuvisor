# Plan Review: TP-151 Step 3 — Make apply_training_plan retry-safer

**Verdict: REVISE**

The Step 3 checklist is directionally aligned with the task, but it is not specific enough to protect the riskiest path in `apply_training_plan`: `replace_existing` currently treats any same-day WORKOUT conflict as replaceable unless `eventMatchesWriteParams` is exact. After adding deterministic `external_id`s, a retry with the same key but a drifted body must be treated as an idempotent/protected duplicate, not deleted and recreated.

## Required plan updates before implementation

1. **Specify the deterministic ID helper contract.** Add a plan item for a small helper that takes the accepted tuple `(plan_id, start_date, workout_id, relative_day, event_date)`, uses canonical serialization, and pins the `icuvisor-plan-v1-` digest length in tests. The exact output should not include raw plan/workout IDs.
2. **Thread the tuple through event creation.** `eventParamsFromPlanWorkout` currently only receives `(date, workout)`, so the plan should say how it will receive `plan_id`, `start_date`, and `relative_day` and assign `WriteEventParams.ExternalID` for every proposed/applied workout.
3. **Update conflict/preflight semantics for external IDs.** Add an explicit item to make `applyTrainingPlanConflictsForParams` (or a shared preflight helper) check `eventMatchesExternalID` before exact writable-field matching, report a clear reason such as `matching_external_id`, and mark it protected so `replace_existing` does not delete retry-owned events whose writable fields drifted.
4. **Make dry-run output shape explicit.** Step 1 says proposed plan rows will show the hashed `external_id`; Step 3 should state that `applyTrainingPlanProposedEvent` gains `external_id` and that tests assert it is present and non-leaking.
5. **Strengthen the tests named in the plan.** Besides stability of repeated write payloads, include coverage that:
   - the same plan/workout/start/date produces the exact same `external_id` across dry-run and apply;
   - changing at least the start date or event date changes the ID;
   - a same-day existing event with matching `external_id` but drifted fields is skipped/protected and is not deleted under `replace_existing`.

Once these details are captured in `STATUS.md`, the implementation should be straightforward and remain within the existing file scope.
