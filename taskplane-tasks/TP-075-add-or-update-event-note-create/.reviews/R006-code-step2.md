# Review R006 — Step 2 code

Decision: **Request changes.**

## Findings

1. **Tool validation regression does not assert the invalid NOTE create is stopped before the writer call** (`internal/tools/add_or_update_event_test.go:139`).

   The Step 2 plan approved in R005 called for the NOTE-without-name case to verify both a user-facing validation error and that `fakeEventWriterClient.AddOrUpdateEvent` is not called. The new case only checks `err != nil` inside the generic bad-arguments loop. For a write tool, the important regression is that invalid NOTE creates are rejected before any upstream write attempt. Please split this NOTE case into a focused test (or add per-case assertions) that checks:

   - `errors.As(err, &userErr)` / public validation message as appropriate, and
   - `len(client.calls) == 0` after the handler returns.

   This keeps the test aligned with the live-probe discovery (`name` is required for NOTE create) and protects Step 3 from accidentally returning an error after invoking the writer.

## Verification run

- `go test ./internal/intervals ./internal/tools -run 'TestAddOrUpdateEvent(SendsNoteCreateBody|RejectsBadArguments)'` — fails as expected on the intended pre-fix regressions (`NOTE` date serialization and missing-name validation).
