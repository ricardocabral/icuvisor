# Code Review: TP-082 Step 2 — Add failing golden tests

## Verdict: APPROVE

No blocking findings for the Step 2 test-only change. The added coverage matches the approved matrix: default write responses are checked with strict key-presence assertions, meaningful zero/false/empty values are preserved, and raw-null preservation is asserted only on tools that already expose `include_full`.

## What I checked

- Ran `git diff 9f19375..HEAD --name-only` and reviewed the full diff.
- Read `PROMPT.md`, `STATUS.md`, and the changed `internal/tools/*_test.go` files.
- Verified the audited write-tool set is covered by the new/strengthened tests:
  - `add_or_update_event`
  - `link_activity_to_event`
  - `add_activity_message`
  - `update_wellness`
  - `update_sport_settings`
  - `apply_training_plan`
  - `create_workout`
  - `update_workout`
  - `create_custom_item`
  - `update_custom_item`
- Ran:

```sh
go test ./internal/tools -run 'Test(AddOrUpdateEvent|LinkActivityToEvent|AddActivityMessage|UpdateWellness|UpdateSportSettings|ApplyTrainingPlan|CreateWorkout|UpdateWorkout|CreateCustomItem|UpdateCustomItem)'
```

The command failed only on the expected red tests:

- `TestCreateCustomItemDefaultStripsSparseNullsAndPreservesMapValues`
- `TestUpdateCustomItemDefaultStripsSparseNullsAndPreservesMapValues`

Both failures are the planned pre-Step-3 behavior: custom-item write responses still hard-code full shaping and preserve `image: null` by default.

## Notes

- The new `assertKeyAbsent` / `assertKeyPresentNil` helpers are scoped to tests and correctly distinguish absent keys from present JSON nulls.
- Non-blocking status hygiene: `STATUS.md` has `Review Counter: 6` and adds the R006 plan review artifact, but the Reviews table still lists only R001-R005. Consider adding R006 to that table during the next status update.
