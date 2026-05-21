# Code Review — TP-084 Step 1

**Verdict:** APPROVE

## Findings

No blocking findings.

## Notes

- The new activity fixture is valid JSON and contains synthetic/scrubbed-looking values only; I did not see API keys, live athlete identifiers, notes, GPS data, or activity names.
- The fixture shape is aligned with the Step 1 contract for numeric/no-`i` activity IDs and empty/minimal rows that should be classified as Strava-unavailable stubs.
- `STATUS.md` captures the expected marker contract for Strava stubs, activity subresource fallbacks, event detail/list inconsistency, and NOTE date-only serialization. Minor cleanup for a later edit: the `| 2026-05-20 13:32 | Review R001 | ... |` row is currently under `## Notes` instead of the `## Execution Log` table.
- Fixture provenance/redaction is documented in `internal/intervals/testdata/activities/README.md`. If the task runner requires provenance specifically in `STATUS.md`, mirror that note there before closing Step 4.

## Verification

- `python3 -m json.tool internal/intervals/testdata/activities/strava_sync_chain_empty_stubs.json >/dev/null`
- `go test ./internal/tools ./internal/intervals`
