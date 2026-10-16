# Calendar reconciliation fixtures

Synthetic, local-only contracts; no athlete data or upstream calls are used.

- `events.json`: distinct same-day sessions, occupied reschedule destinations,
  and one multi-day race whose target is not expanded into daily targets.
- `compliance.json`: explicitly linked strength/ride sessions plus a second
  provider copy. Compliance counts the two scheduled sessions once and retains
  the explicit activity links.
- `timezones.json`: UTC presentation in positive/negative offsets and across the
  New York spring DST transition.
- `linking.json`, `interrupted_create.json`, `partial_range.json`: retry,
  conflict, interrupted-write verification, and partial-range failure cases.
- `aggregates.json`: upstream weekly summary snapshots for linked sessions with
  a provider copy and for a three-stage race. Training summary, load balance,
  and fitness must preserve the supplied aggregate once, without adding calendar
  targets or re-summing provider copies. Fitness remains the upstream CTL value.

These tests do not assert that icuvisor repairs duplicate sessions already
included in upstream aggregate totals. Provider names, matching timestamps, and
external IDs from different namespaces alone are not proof of session identity.
The existing contracts consume upstream summaries and explicit links; they do
not infer a new calendar, deduplicate arbitrary provider uploads, or identify
race peaks. The multi-day event read remains one event with its end date.

The client-level `internal/intervals/testdata/events/write_confirmations.json`
checks malformed HTTP-success payloads. The tool/client integration test then
checks that an unconfirmed create can succeed only after one bounded calendar
read confirms it, and that a repeat skips the second write.
