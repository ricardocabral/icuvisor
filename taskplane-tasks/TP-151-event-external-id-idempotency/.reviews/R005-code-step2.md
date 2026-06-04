# Code Review: TP-151 Step 2 — Implement event write/read support

**Verdict: APPROVE**

No blocking findings.

## Notes

- `external_id` is trimmed and omitted when blank in the intervals write payload, preserving the Step 1 no-clear semantics.
- `add_or_update_event` maps the new argument through to `WriteEventParams`, exposes it in terse event rows, and handles same-day matching-`external_id` duplicate skips with explicit metadata.
- Existing exact duplicate preflight remains in place for requests without matching external IDs.

## Verification

- `go test ./internal/intervals ./internal/tools -run 'Event|AddOrUpdateEvent'`
- `go test ./internal/tools`
- `go test ./...`
- `make lint`
