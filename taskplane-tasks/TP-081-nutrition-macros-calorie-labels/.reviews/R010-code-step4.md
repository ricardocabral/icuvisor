# R010 Code Review — Step 4: Docs and full verification

**Verdict:** REVISE

## Findings

### P1 — Step 4 closes with unresolved REVISE reviews recorded as APPROVE

`STATUS.md:98-99` still records R007/R008 as `APPROVE`, and `STATUS.md:139-141` appends approval outcomes, but the checked-in review files say otherwise: `.reviews/R007-plan-step3.md:3` is `REVISE`, `.reviews/R008-code-step3.md:3` is `REVISE`, and the newly added `.reviews/R009-plan-step4.md:3` is also `REVISE`. Despite that, Step 4 marks all work complete at `STATUS.md:57-62` and logs R009 as approved at `STATUS.md:141`.

This is not just bookkeeping: R008 called out missing/ineffective regression coverage that Step 4 was supposed to resolve before final docs/full verification. Please either move the task back to the appropriate step and fix those blockers, or add the follow-up review files that actually approve them, then update the Reviews table/execution log to match the files. Do not mark Step 4 complete while the only checked-in Step 4 plan review is a revision request.

### P2 — `get_activity_details` output schema still omits the documented calorie key

The docs update covers the `get_activity_details` summary, but the MCP output schema description still says only gear fields are included: `internal/tools/get_activity_details.go:293-294`. The Step 4 plan review explicitly asked for the affected registered tool/output-schema descriptions to name activity `calories_burned` as active/exercise calories. `get_activities` and `get_wellness_data` output schemas were updated, but the detail schema exposed to MCP clients remains stale.

Please update `activityReadOutputSchema()` so activity-detail responses document `calories_burned` semantics. If the shared schema wording needs to stay generic because it is reused by interval/stream/message tools, make the description clear that the calorie key applies to activity-detail rows.

### P2 — Full verification results are checked off but not recorded in STATUS

`STATUS.md:60-62` marks tool docs, changelog, and the full quality gate complete, but `STATUS.md:109-116` has no entries for `go test -count=1 ./internal/tools`, `make docs-tools`, `make test`, `make build`, or `make lint`, and Step 5 remains entirely unchecked at `STATUS.md:67-74`. R009 specifically requested concrete command outcomes in STATUS before checking the gate complete.

Please record the commands and outcomes in the execution log (or Step 5 checklist if that is the intended source of truth). This keeps the task audit trail consistent with the completion criteria.

## Verification

Ran:

```sh
git diff cce260df0f23353e08ec29698904889c7323ea17..HEAD --name-only
git diff cce260df0f23353e08ec29698904889c7323ea17..HEAD
make docs-tools
go test -count=1 ./internal/tools
make test
make build
make lint
```

Results: all commands passed. `make docs-tools` produced no additional diff.
