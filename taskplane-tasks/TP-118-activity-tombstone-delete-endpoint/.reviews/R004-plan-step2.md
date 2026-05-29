# R004 Plan Review — Step 2

**Verdict:** REVISE

## Findings

1. **The planned targeted test selector still misses the intervals endpoint tests.**  
   The Step 2 prompt/test scope uses `go test ./internal/intervals ./internal/tools -run 'DeleteActivity|delete_activity|DeleteTools|Schema'`. I ran the intervals half and it reports `[no tests to run]`, so it would not verify the exact DELETE path nor target-athlete ownership safety. Step 2 should explicitly include the tests identified in Step 1, e.g. `DeleteMethods|ActivityIDEndpointsRequireResolvedTargetOwnership`, along with the relevant tools/schema checks.

2. **Lock the affected public source-endpoint metadata, not just the HTTP client path.**  
   `delete_activity` returns `_meta.source_endpoint` via `deleteActivityEndpoint`, currently `"/activity/{activityId}"`, and there is no committed `internal/tools/schema_snapshot/delete_activity.json` in this tree. If the client moves to tombstone, the plan should explicitly update this constant and add/update a tool test assertion for the returned `source_endpoint` (or record why it remains unchanged), rather than relying on a non-existent snapshot file.

## Recommendation

Revise the Step 2 plan/checklist before implementation to include the corrected test command, for example:

```sh
go test ./internal/intervals ./internal/tools -run 'DeleteMethods|ActivityIDEndpointsRequireResolvedTargetOwnership|DeleteTools|delete_activity|Schema'
```

Also call out the `delete_activity` `_meta.source_endpoint` expectation (`/activity/{activityId}/tombstone` if tombstone is selected) as an explicit artifact/test outcome.
