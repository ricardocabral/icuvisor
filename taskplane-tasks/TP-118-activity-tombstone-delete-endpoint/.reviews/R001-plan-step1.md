# R001 Plan Review — Step 1

**Verdict:** Changes requested

## Findings

1. **Targeted test command does not exercise the current intervals delete-path coverage.**  
   I ran `go test ./internal/intervals ./internal/tools -run 'DeleteActivity|delete_activity|DeleteTools'`; `internal/intervals` reported `[no tests to run]`. The existing path assertion is `TestDeleteMethodsSendDeletePaths`, and target-athlete safety is covered under `TestActivityIDEndpointsRequireResolvedTargetOwnership`, neither of which matches the planned regex. Step 1 should add a package-level intervals run or adjust the regex, e.g. include `DeleteMethods|ActivityIDEndpointsRequireResolvedTargetOwnership`, so the endpoint/safety contract is actually validated before the decision is recorded.

2. **Make the `/api/v1` base-path handling explicit in the plan/discovery.**  
   The observed OpenAPI path includes `/api/v1`, while this client’s default `APIBaseURL` already includes `/api/v1` and code passes relative path parts to `JoinPath`. If tombstone is selected, the implementation/test expectation should be `DELETE /activity/{id}/tombstone` relative to the configured base URL, not a duplicated `/api/v1/api/v1/...` path. Capture that in Step 1 Discoveries or Step 2 test notes.

## Recommendation

Update the Step 1 checklist/execution notes to include the corrected intervals test coverage and the base-path decision. The clean-room/source-evidence guardrails and the “record endpoint decision in Discoveries” requirement are otherwise appropriate.
