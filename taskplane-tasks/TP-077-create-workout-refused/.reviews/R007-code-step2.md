# Code Review — TP-077 Step 2

Verdict: **request changes**

## Findings

1. **The existing mutually-exclusive description/workout_doc validation coverage is being masked by the new missing-folder contract.**  
   In `internal/tools/create_workout_test.go:146`, the bad-arguments row for `description` plus `workout_doc` still omits `folder_id`. Once Step 3 adds required-folder validation, that case can pass by failing on the missing folder before it ever exercises the mutually-exclusive payload rule. This weakens existing coverage while adding the new contract. Please add a valid sanitized `folder_id` to bad-argument cases whose purpose is not “missing folder” (at least the description+workout_doc row; doing the same for the unknown-field row would keep that case similarly isolated).

## Validation run

- `go test ./internal/tools ./internal/intervals -run 'CreateWorkout|CreateLibraryWorkout'` fails as expected on the new required-folder assertions in `TestCreateWorkoutRejectsBadArguments`, `TestCreateWorkoutRegistrationMetadata`, and `TestCreateLibraryWorkoutRequiresWritableBasics`.
