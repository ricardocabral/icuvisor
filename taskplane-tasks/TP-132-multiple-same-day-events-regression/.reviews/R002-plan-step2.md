# Plan Review: Step 2 — Add regression coverage

**Verdict:** Approved

The Step 2 plan is correctly scoped to regression coverage only and targets the right surfaces: `get_today` and `get_events` same-day event shaping, with targeted `go test ./internal/tools` verification and `make eval-validate` only if the eval cookbook changes.

Targeted baseline verification during review:

```sh
go test ./internal/tools
# ok github.com/ricardocabral/icuvisor/internal/tools (cached)
```

Non-blocking suggestions to make the regression tests maximally useful:

- Add one focused `get_today` case with two same-date `WORKOUT` rows plus NOTE/race annotations, and assert `planned_events` has both distinct `event_id`/`name` values and `_meta.section_counts.planned_events == 2`.
- Add one focused `get_events` case with same-date workouts returned in a deliberately non-sorted order, then assert both rows survive and response ordering/identity are deterministic (`start_date_local`, then `event_id`).
- Prefer checking IDs, names, categories, and counts rather than only row length, so a future map-by-date collapse or overwrite fails clearly.
- Keep eval scenario changes optional. If added, ensure the scenario expects `resolve_calendar_dates` before `get_events` for “tomorrow” and run `make eval-validate`.
