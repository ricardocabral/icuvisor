# Code Review — TP-079 Step 5

**Verdict:** REVISE

Step 5 is not ready to approve. The implementation marks verification complete, but the checked-in review/status record is contradictory and the known activity pagination-token regression called out in prior reviews is still present in `HEAD`.

## Findings

### 1. `STATUS.md` records `REVISE` reviews as approved, including the new Step 5 plan review

**Severity:** P1  
**Location:** `taskplane-tasks/TP-079-gear-read-name-resolution/STATUS.md:111-114`, `taskplane-tasks/TP-079-gear-read-name-resolution/STATUS.md:159-162`, `taskplane-tasks/TP-079-gear-read-name-resolution/.reviews/R014-plan-step5.md:1-4`

The status table and execution log say R011/R012/R013/R014 are `APPROVE`, but the checked-in review files have `**Verdict:** REVISE` for all four. The new R014 file explicitly rejects the Step 5 plan, yet this diff adds it to `STATUS.md` as approved and then checks off the Step 5 verification boxes.

This makes the task state non-auditable and bypasses unresolved review blockers. Please make `STATUS.md` match the actual review files, resolve or explicitly supersede each `REVISE` finding, and only then mark Step 5 as complete.

### 2. The old-token `gear_id` pagination regression remains unfixed

**Severity:** P1  
**Location:** `internal/tools/get_activities_cursor.go:234-247`

R011/R013/R014 all called out that accepted terse `next_page_token`s issued before `gear_id` existed can overwrite the current terse field list. That code path is unchanged:

```go
if !args.IncludeFull {
    cursor.token.Fields = append([]string(nil), terseActivityFields...)
}
if token != nil {
    ...
    cursor.token.Fields = append([]string(nil), token.Fields...)
}
```

For an old valid token whose `fields` array lacks `gear_id`, a continued terse request still omits `gear_id` from `ListActivities`. With upstream field filtering, the returned rows then have no `Activity.GearID`, `resolveActivityGear` skips lookup, and the page omits `gear_id`/`gear_name`/`gear_resolution` even when upstream gear data exists. This violates the TP-079 requirement that activity reads request and surface gear IDs/names when upstream permits.

Please either merge mandatory current terse fields into token fields when `include_full=false` (at least `gear_id`) or reject/bump incompatible tokens, and add the requested regression test using an accepted old token without `gear_id`.

### 3. Verification is checked off without an auditable command log

**Severity:** P2  
**Location:** `taskplane-tasks/TP-079-gear-read-name-resolution/STATUS.md:76-83`, `taskplane-tasks/TP-079-gear-read-name-resolution/STATUS.md:126-162`

Step 5 now checks off targeted tests, `make test`, `make build`, and `make lint`, but `STATUS.md` does not record the exact commands, timestamps, outcomes, failure summaries, or generated-artifact/working-tree checks requested by the Step 5 plan review. Given the unresolved blocker above, a green package test alone would not prove the task requirements are satisfied.

Please record the final verification commands and outcomes in `STATUS.md` after fixing the blocker, including the targeted gear/activity pagination regression coverage and final `make test`, `make build`, and `make lint` results.

## Tests run during review

- `git diff c1d9bd31e862cecd1cf2be3c7b4c3bd3c3bcf9fb..HEAD --name-only`
- `git diff c1d9bd31e862cecd1cf2be3c7b4c3bd3c3bcf9fb..HEAD`
- `go test ./internal/tools`
