# Plan Review — Step 1

Decision: **approved**

The revised Step 1 plan now addresses the blocking issue from R001: it explicitly updates the terse-default fixture with a non-empty `description` before asserting that `description` is absent from the default response. Given the current implementation in `workoutInFolderToRow()` unconditionally sets `Description: stringValue(workout.Description)`, this should produce the intended red test before the production fix.

## Validation notes

- Keep the change scoped to `TestGetWorkoutsInFolderHidesWorkoutDocByDefault` for this step.
- Do not modify `workoutInFolderToRow()` yet; Step 1 should remain a reproduction-only red test.
- A targeted red-test command is sufficient for this step, for example:

```sh
go test ./internal/tools -run TestGetWorkoutsInFolderHidesWorkoutDocByDefault -count=1
```

## Expected outcome

After adding a non-empty `description` to the fixture and asserting that the default row omits it, the test should fail on the current code because `description` is still included in the terse response.
