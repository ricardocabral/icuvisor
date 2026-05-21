# Plan Review — Step 1

Decision: **changes requested**

## Blocking finding

- The Step 1 plan says to add an assertion that `description` is absent in `TestGetWorkoutsInFolderHidesWorkoutDocByDefault`, then confirm the test fails on `main`. As written, that will not reproduce the bug because the existing fixture for this test does **not** include a `description` field:
  - `internal/tools/get_workout_library_test.go` currently decodes only `id`, `name`, `type`, `folder_id`, and `workout_doc` for the terse-default test.
  - Since `workoutInFolderRow.Description` has `omitempty`, an absent upstream description produces no `description` key even with the current buggy implementation.

  Result: simply adding `if _, ok := row["description"]; ok { ... }` would pass on `main`, violating the step's purpose as a red test.

## Required plan adjustment

Update the Step 1 test fixture to include a non-empty workout description, for example:

```json
{"id":2,"name":"Sweet Spot","description":"multi-paragraph coach notes","type":"Ride","folder_id":20,"workout_doc":{"steps":[{"duration":600}]}}
```

Then assert that the default response omits `description`. With the current implementation (`Description: stringValue(workout.Description)` unconditionally set in `workoutInFolderToRow`), that test should fail on `main`.

## Non-blocking suggestion

Keep the assertion in the existing terse-default test rather than creating a new test; that matches the task prompt and keeps coverage focused. The Step 3 full-mode test can then use a similar non-empty `description` fixture to assert that `include_full: true` preserves it.
