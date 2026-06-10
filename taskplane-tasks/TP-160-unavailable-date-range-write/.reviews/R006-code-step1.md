# Code Review: Step 1 — Design the range-write contract

Verdict: REVISE

I ran `go test ./internal/tools -run 'Unavailable|DateRange|Event'`; it still fails to compile because the Step 2 implementation symbols do not exist yet, which is expected for this TDD step.

## Findings

1. **Same-day conflict tests can pass even if the implementation queries only unavailable categories.**  
   `fakeUnavailableDateRangeClient.ListEvents` filters only by `Oldest`/`Newest` and ignores `ListEventsParams.Category` (`internal/tools/add_unavailable_date_range_test.go:35-47`). As a result, an implementation could call `ListEvents` with `Category: "INJURED"`/`"HOLIDAY"`, which would miss workouts in production, but the fake would still return the workout fixture and the conflict assertion at `:234-236` would pass. Please have the fake honor `Category` and/or assert the preflight query leaves `Category` empty so the tests protect the “do not overwrite workouts / report same-day conflicts” safety contract.

2. **Exact duplicate idempotency without the generated external ID is untested.**  
   The contract says pre-existing same-day rows that exactly match writable fields should also skip, but the duplicate fixtures all use `addUnavailableDateRangeExternalID` (`internal/tools/add_unavailable_date_range_test.go:158-164`, `:204-208`). An implementation that only skips matching external IDs, and creates duplicates for manually-created/existing unavailable markers without an external ID, would pass. Add a case with no `external_id` but matching date/category/type/name/description and assert no write plus `skipped` metadata.

3. **The handler path does not test the optional `name` contract.**  
   The request contract includes optional `name`, and the external-id fingerprint includes the defaulted/trimmed name, but every handler test omits `name` and relies on defaults. The helper-level variation at `:80-85` does not prove the handler maps a custom name into `WriteEventParams.Name`, duplicate matching, and the generated `ExternalID`. Add a handler test with a custom/trimmed name to pin that public argument before implementation.
