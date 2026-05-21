# Code Review — TP-077 Step 2

Verdict: **request changes**

## Findings

1. **`TestCreateLibraryWorkoutRequiresWritableBasics` does not actually test the new required-folder validation and can perform real network I/O.**  
   In `internal/intervals/workout_library_test.go:114-125`, the new missing/blank `FolderID` assertions use `newTestClient(t, "https://example.invalid", http.DefaultClient, ...)`. The current production code still accepts `Name+Sport` without `FolderID`, builds a POST, and only returns a non-nil error because the request to `example.invalid` fails. That means the new intervals coverage passes before the fix and does not prove that the client rejects missing `folder_id` before I/O. It also violates the project test rule to never hit the network from tests. Replace this with a no-network client/transport (or an `httptest.Server`) that fails the test if any request is made, and assert the local validation error for omitted and blank `FolderID`.

2. **The schema test should also guard examples against the new required `folder_id` contract.**  
   `internal/tools/create_workout_test.go:173-179` now expects `folder_id` to be required and documented, but the production examples still include a first example without `folder_id` (`internal/tools/create_workout.go:138-144`). If Step 3 only changes the required list/description, the schema can still present an invalid example to MCP clients/LLMs. Please extend the registration metadata test to assert every `examples` / `input_examples` entry includes a non-blank `folder_id`, so the eventual schema fix is complete and internally consistent.

## Validation run

- `go test ./internal/tools ./internal/intervals -run 'CreateWorkout|CreateLibraryWorkout'` currently fails in the intended tool tests, while `internal/intervals` passes for the wrong reason described above.
