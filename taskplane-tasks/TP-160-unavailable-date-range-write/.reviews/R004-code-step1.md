# Code Review: Step 1 — Design the range-write contract

Verdict: REVISE

I ran `go test ./internal/tools -run 'Unavailable|DateRange|Event'`; as expected for the Step 1 TDD state, it currently fails to compile because `newAddUnavailableDateRangeTool` and related implementation symbols do not exist yet.

## Findings

1. **Query bounds are not tested because the fake ignores `ListEventsParams`.**  
   `fakeUnavailableDateRangeClient.ListEvents` records params but always returns all fixtures (`internal/tools/add_unavailable_date_range_test.go:35-37`). That means an implementation could query the wrong date/range, omit `Oldest`/`Newest`, or use a single-day preflight incorrectly and these idempotency/conflict tests would still pass. Please either filter fixtures by `params.Oldest`/`params.Newest` in the fake or assert the expected list call(s), especially for inclusive range preflight.

2. **Created writes do not assert the stable idempotency fingerprint.**  
   The create test only checks the generated `ExternalID` prefix and per-day uniqueness (`internal/tools/add_unavailable_date_range_test.go:87-92`). This would allow a random/non-deterministic external ID and still pass, breaking retry idempotency. Assert each write's `ExternalID` equals `addUnavailableDateRangeExternalID(normalizedCategory, date, defaultedOrTrimmedName, description)`.

3. **The R003 duplicate/conflict edge cases are still uncovered.**  
   The duplicate fixture matches both `external_id` and all writable fields (`internal/tools/add_unavailable_date_range_test.go:112-123`), and the mixed conflict fixture puts the duplicate and unrelated conflict on different dates (`internal/tools/add_unavailable_date_range_test.go:143-149`). This does not catch an implementation that skips solely on matching external ID even when writable fields differ, or one that stops scanning a date after finding a duplicate and hides other same-day conflicts. Add tests for those two cases before Step 2.

4. **`include_full` shaping is listed in the Step 1 contract but not tested.**  
   The status contract says initial tests should cover `include_full` shaping, but all new handler calls omit `include_full`. Add a terse/default assertion that `full` is absent and an `include_full:true` case where created/skipped event rows include the raw upstream payload.
