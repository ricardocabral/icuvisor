# Code Review — TP-079 Step 4

**Verdict:** REVISE

The changelog/PRD/tool-data documentation updates are directionally correct, and the generated docs artifact is reproducible. However, Step 4 was completed on top of an unresolved prior `REVISE` review, and the task status now records review outcomes that contradict the checked-in review files. The known activity pagination regression from R011 is still present in `HEAD`, so the verify/document step should not be marked complete yet.

## Findings

### 1. Step 4 records an approved plan review even though the checked-in review says `REVISE`

**Severity:** P1  
**Location:** `taskplane-tasks/TP-079-gear-read-name-resolution/STATUS.md:109-112`, `taskplane-tasks/TP-079-gear-read-name-resolution/.reviews/R012-plan-step4.md:1-45`

`STATUS.md` now records `R012 | plan | 4 | APPROVE`, but `.reviews/R012-plan-step4.md` has `**Verdict:** REVISE` and lists required plan revisions. The same table still records `R011` as approved even though `.reviews/R011-code-step3.md` is also `REVISE`.

This is not just bookkeeping: R012 explicitly required reconciling/fixing the R011 pagination issue before doing Step 4 documentation. The Step 4 commit instead marks the original Step 4 checklist complete without updating the plan or correcting the review record. Please make the status table match the actual review files, apply the required Step 4 plan revisions, and only mark Step 4 complete after the revised plan and the known Step 3 blocker are addressed.

### 2. The R011 pagination-token regression remains unfixed in current code

**Severity:** P1  
**Location:** `internal/tools/get_activities_cursor.go:234-247`

R011 identified that accepted terse `next_page_token`s issued before `gear_id` was added can overwrite the current terse field set and drop `gear_id` from subsequent page fetches. That code path is unchanged:

```go
if !args.IncludeFull {
    cursor.token.Fields = append([]string(nil), terseActivityFields...)
}
if token != nil {
    ...
    cursor.token.Fields = append([]string(nil), token.Fields...)
}
```

So an old/accepted token without `gear_id` still suppresses the field request, `resolveActivityGear` never sees the upstream gear ID, and the follow-up page can omit the new `gear_id`/`gear_name`/`gear_resolution` behavior that Step 4 now documents. Please fix this by merging mandatory current terse fields into token fields when `include_full=false` (or rejecting/bumping incompatible tokens) and add the regression test requested in R011 before declaring verification complete.

### 3. Step 4 verification is marked done without recording the commands/results in STATUS

**Severity:** P2  
**Location:** `taskplane-tasks/TP-079-gear-read-name-resolution/STATUS.md:66-72`, `taskplane-tasks/TP-079-gear-read-name-resolution/STATUS.md:126-156`

The Step 4 checklist now says targeted tests and the full suite were run, but the execution log has no Step 4 verification entries and there are no notes recording the generated-docs command, reviewed docs, or results. R012 specifically asked for concrete verification/documentation outcomes in the plan/status. Please record the actual commands and outcomes (for example targeted package tests, `make docs-tools`, `git diff` review of `web/data/tools.json`, and whether `go test ./...`/`make test` was run now or deferred to Step 5) so the task state is auditable.

## Tests run during review

- `git diff 18e7a7251f15d0058852359eb8010c5e02b13fb7..HEAD --name-only`
- `git diff 18e7a7251f15d0058852359eb8010c5e02b13fb7..HEAD`
- `go test ./internal/intervals ./internal/tools ./internal/toolcatalog ./internal/toolchecks ./cmd/gendocs`
- `make docs-tools && git diff --stat && git diff --exit-code -- web/data/tools.json cmd/gendocs/testdata/tools.golden.json`
- `go test ./...`
