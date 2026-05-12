# Event detail/list inconsistency fixtures

This directory stores sanitized reproducer payloads for cases where the intervals.icu event list endpoint returns an event ID but the event detail endpoint returns 404 for that same ID.

Fixture format:

- `*_list.json` — originating `GET /athlete/{athlete_id}/events` response containing the listed event. Athlete identifiers, calendar identifiers, names, notes, and descriptions must be synthetic or scrubbed.
- `*_detail_404.txt` — short note for the matching `GET /athlete/{athlete_id}/events/{event_id}` detail request and 404 outcome. Do not include API keys, athlete identifiers, personal notes, or raw upstream error bodies.

No real athlete reproducer was captured during TP-012 implementation. The current fixture is synthetic and exists to document the expected mismatch shape for tests and future root-cause analysis.
