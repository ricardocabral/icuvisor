# Code Review R005 — Step 2

Verdict: Changes requested

## Findings

### P1 — Exact duplicate detection can match completed/actual metrics instead of writable targets

`eventMatchesWriteParams` treats an event as an exact duplicate when requested target fields match either the writable target field or the completed/actual field (`icu_training_load`, `distance`, `moving_time`, `elapsed_time`) at `internal/tools/add_or_update_event.go:292-318`. The write body only sends `load_target`, `distance_target`, `time_target`, and `elapsed_time_target`, so falling back to completed metrics can silently skip a create when there is a same-day completed/imported activity with the same name/type and actual load/distance/time, even though the writable planned targets are absent. This affects both `add_or_update_event` duplicate skips and `apply_training_plan` because they share this matcher.

Please make the exact-duplicate path compare the actual writable create shape only. If upstream sometimes echoes planned writes in non-target fields, gate that with a planned-event marker/fixture-backed normalization; otherwise classify those same-day matches as `existing_event_on_date` conflicts/warnings rather than `duplicate_existing_event`. Add a regression where a same-day event has matching `icu_training_load`/`distance`/`moving_time` but no target fields and confirm the create is not skipped.

## Verification

```sh
go test ./internal/tools
```

Result: passed (`ok`, cached).
