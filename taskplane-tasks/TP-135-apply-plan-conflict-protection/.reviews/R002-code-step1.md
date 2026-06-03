# Code Review — Step 1

Result: **Request changes**

## Findings

1. **STATUS audit is still missing the required conflict-flow details.** `STATUS.md:64` summarizes the conflict behavior, but it still does not record the actual code path (`fetchApplyTrainingPlanEvents` → `applyTrainingPlanConflictsForParams` → `eventCreatePreflightFromEvents`), the duplicate short-circuit in `eventCreatePreflightFromEvents` that returns immediately and drops other same-day rows, or the non-dry-run per-day re-preflight path before create. Those details were called out in R001 and are critical for Step 2; otherwise the implementation can miss mixed days containing an exact workout duplicate plus a NOTE/race/unavailable item, and can protect only the initial range preflight while leaving the re-preflight delete path unsafe.

2. **The taxonomy is too vague to implement safely.** `STATUS.md:64` says protected conflicts include `RACE` and “UNAVAILABLE-like calendar annotations/blocks,” but the documented categories are concrete values such as `RACE_A`, `RACE_B`, `RACE_C`, `HOLIDAY`, `SICK`, and `INJURED`; it also does not decide whether all non-`WORKOUT` categories are protected by default. Step 2 needs explicit replaceable/protected category rules so `replace_existing` can delete only intended workout conflicts.

## Verification

- Ran `go test ./internal/tools` — passes.
