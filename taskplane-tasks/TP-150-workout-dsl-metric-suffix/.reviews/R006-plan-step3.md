# Plan Review: Step 3 — Refresh schemas and user guidance

**Verdict:** Revise.

## Findings

1. **Step 2 is not cleanly closed in the review artifacts.** `STATUS.md` records R005 as approved and Step 2 complete, but `.reviews/R005-code-step2.md` says “Request changes” and asks for missing direct write-path regression tests. Resolve that discrepancy before starting Step 3, either by fixing the tests and getting/recording the follow-up approval, or by correcting the review/status artifacts if a newer approval exists elsewhere.

2. **The Step 3 plan is too conditional for known-required schema guidance.** Step 1 explicitly decided that `update_workout` without `sport` falls back to bare zone DSL, and that schema/docs should tell callers to include `sport` with `workout_doc` when sport-aware metric suffixes matter. The plan should name the concrete schema/description changes: at minimum `update_workout` `sport`/`workout_doc` wording (and likely its workout-doc example), plus `create_workout.sport` and `add_or_update_event.type` wording that those fields let icuvisor apply athlete sport metric priority and emit explicit `Power`/`HR`/`Pace` zone suffixes.

3. **User-visible docs/changelog need explicit handling, not “if needed”.** This behavior changes planned-workout DSL output for users, so a `[Unreleased]` changelog entry is required. The cookbook should mention automatic sport-aware zone suffixing and the `update_workout` caveat. `web/content/reference/resources-prompts.md` and `docs/prd/PRD-icuvisor.md` can be marked unaffected if no product contract/resource summary changes are needed, but that review should be recorded in `STATUS.md`.

## Expected revised plan

- Update the relevant tool descriptions/input schemas, then regenerate schema snapshots with `go run ./scripts/snapshot_tool_schemas.go`.
- Update `web/content/cookbook/build-workouts.md` and `CHANGELOG.md`.
- Explicitly review/record unaffected docs, including PRD/resource-reference docs if left unchanged.
