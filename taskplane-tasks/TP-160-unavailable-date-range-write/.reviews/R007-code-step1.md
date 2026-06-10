# Code Review: Step 1 — Design the range-write contract

Verdict: REVISE

I ran `go test ./internal/tools -run 'Unavailable|DateRange|Event'`; it still fails to compile on the intentionally missing Step 2 symbols. I also ran `git diff --check 9594870..HEAD`.

## Findings

1. **Mid-range write failure behavior is contracted but untested.**
   `STATUS.md` says successful writes are not rolled back and retries are safe if an upstream write fails mid-range, but `writeError` in `fakeUnavailableDateRangeClient` only fails every write from the first call and no test uses it (`internal/tools/add_unavailable_date_range_test.go:19-26`). Add a case where day 1 succeeds and day 2 fails, assert two attempted writes, the short public write error, and no rollback/delete behavior.

2. **The preflight can still pass tests while requesting too few events.**
   The fake ignores `ListEventsParams.Limit` (`internal/tools/add_unavailable_date_range_test.go:35-51`), and the list-call assertions only check bounds/category (`:133`, `:179`, `:225`). An implementation using `Limit: 1` or leaving the upstream default could miss same-day workouts/conflicts in production while these tests still pass. Assert `Limit == maxEventsLimit` (or make the fake honor `Limit`) for the range preflight.

3. **Malformed date validation is not pinned.**
   The invalid-input table covers unsupported category, broad alias, reversed range, and excessive range, but not non-`YYYY-MM-DD` or impossible dates (`internal/tools/add_unavailable_date_range_test.go:355-358`). Since the public contract requires athlete-local `YYYY-MM-DD`, add malformed date cases for `start_date` and `end_date`.

4. **Whitespace check fails in added review artifacts.**
   `git diff --check 9594870..HEAD` reports trailing whitespace in the newly added R004/R005/R006 markdown files. Remove those trailing spaces before committing if these task artifacts are part of the branch diff.
