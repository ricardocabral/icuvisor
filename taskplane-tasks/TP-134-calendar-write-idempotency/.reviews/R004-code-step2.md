# Code Review R004 — Step 2

Verdict: Changes requested

## Findings

### P1 — Duplicate matching treats omitted create fields as wildcards

`eventMatchesWriteParams` only compares `name`, `description`, tags, indoor, and targets when the caller supplied a non-empty/non-nil value (`internal/tools/add_or_update_event.go:272-301`). For creates, omitted fields are not “leave unchanged”; they are part of the event that would be created as empty/absent. This means a request such as `{"date":"2026-06-01","category":"WORKOUT","type":"Ride"}` will be reported as `skip_duplicate` if there is any same-day Ride workout, even if the existing event has a different name, description, tags, or targets. That violates the Step 2 contract to skip exact duplicates and warn on non-identical same-day events, and can silently drop a valid user-requested create.

Please make create duplicate matching compare the full writable create shape (including absent/empty fields, with whatever normalization is needed for upstream defaults), and leave broader same-day matches as conflicts/warnings instead of duplicate skips.

## Verification

```sh
go test ./internal/tools
```

Result: passed (`ok`, cached).
