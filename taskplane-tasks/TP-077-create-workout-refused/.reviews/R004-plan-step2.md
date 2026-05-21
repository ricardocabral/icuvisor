# Plan Review — TP-077 Step 2

Verdict: **request changes**

## Findings

1. **The planned payload assertion may not produce a failing test.**  
   Step 1 found that upstream accepts `type: "Ride"` with an existing `folder_id`, and rejects missing/`null` `folder_id`. The current client already serializes create `Sport` as JSON `"type"` and preserves a non-empty `folder_id`; `internal/intervals/workout_library_test.go` already checks that shape. Updating only the expected outbound body to the accepted `type` + `folder_id` payload is likely to pass on current code, which would not satisfy Step 2's purpose. Revise the plan to add at least one assertion that fails today, focused on the discovered defect: create requires a non-empty existing folder ID.

2. **Make the failing contract explicit at the tool boundary.**  
   Step 3 calls for a clear public validation error if `folder_id` is required, so Step 2 should plan a failing `create_workout` tool test for omitted/blank `folder_id` instead of relying solely on the intervals client body test. Good targets:
   - `TestCreateWorkoutRejectsBadArguments` includes `{"name":"Tempo","sport":"Ride"}` / blank `folder_id` as validation errors.
   - `TestCreateWorkoutRegistrationMetadata` expects `folder_id` in the JSON Schema `required` list and a description that says it must be an existing folder owned by the athlete.
   These fail on current code and directly protect the user-visible contract.

3. **Keep existing non-folder tests coherent with the new contract.**  
   `TestCreateWorkoutWithFreeTextOnlyPreservesDescription` and `TestCreateWorkoutGoldenFixtureRoundTripFromWorkoutDocSerializer` currently omit `folder_id`. The Step 2 plan should update those otherwise-unrelated happy-path tests to include a synthetic existing folder ID, so the only intentional failing assertions are the new required-folder validation/schema checks. Otherwise Step 3 may inherit noisy, confusing failures.

## Recommended amended Step 2 plan

- Add/update an intervals client test only if it asserts the missing-folder create contract, e.g. `CreateLibraryWorkout(... Name+Sport but no FolderID ...)` returns an error before making a request. Reuse `create_request.json` / `create_response.json` for the accepted body/response fixture if touching the outbound-body test.
- Add the user-facing tool validation/schema tests described above.
- Update current happy-path create-workout tool tests to pass a sanitized folder ID (`f-test-folder` or similar).
- Run the narrow tests and record that the new required-folder assertions fail before production code changes, for example `go test ./internal/tools ./internal/intervals -run 'CreateWorkout|CreateLibraryWorkout'`.
- Do not change production validation/serialization code in Step 2; save that for Step 3.
