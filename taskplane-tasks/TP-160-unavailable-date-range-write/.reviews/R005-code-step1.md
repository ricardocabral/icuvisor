# Code Review: Step 1 — Design the range-write contract

Verdict: REVISE

I ran `go test ./internal/tools -run 'Unavailable|DateRange|Event'`; it fails to compile because the Step 2 implementation symbols (`newAddUnavailableDateRangeTool`, `addUnavailableDateRangeExternalID`, etc.) do not exist yet, which is expected for this Step 1 TDD state.

## Findings

1. **The external-id fingerprint is not pinned independently of the implementation.**  
   The tests compute expected IDs and fixtures with `addUnavailableDateRangeExternalID` itself (`internal/tools/add_unavailable_date_range_test.go:98`, `:130-131`, `:164`, `:198`). If Step 2 implements that helper incorrectly, for example omitting `description` or `name` from the hash, these tests can still pass because the production code and test fixtures share the same broken helper. Add direct assertions that changing each contract field (`date`, normalized category, defaulted/trimmed name, description) changes the ID, and preferably one golden expected ID/prefix-length assertion so the stable retry key contract is locked.

2. **The structured response contract is still mostly untested.**  
   The contract in `STATUS.md` requires `_meta.date_range`, `range_cap_days`, `include_full`, `skipped:[{date,event_id,reason}]`, and `same_day_conflicts:[...]`, but the new tests mainly assert counts (`internal/tools/add_unavailable_date_range_test.go:121-123`, `:155-157`, `:189-191`, `:213-215`). An implementation could omit the `skipped` and `same_day_conflicts` detail arrays, or return the wrong date-range metadata, and these tests would still pass. Add assertions on those fields, especially in the all-skipped and duplicate-plus-same-day-conflict cases.

3. **`include_full:true` is only covered for newly created rows.**  
   `TestAddUnavailableDateRangeIncludeFullAddsRawEventPayload` covers a created event (`internal/tools/add_unavailable_date_range_test.go:219-240`), but repeated/idempotent calls are expected to return existing skipped event rows too. Add an `include_full:true` repeated-range case with raw fixture fields so Step 2 cannot accidentally include raw payloads only for writes and drop them for skipped duplicates.
