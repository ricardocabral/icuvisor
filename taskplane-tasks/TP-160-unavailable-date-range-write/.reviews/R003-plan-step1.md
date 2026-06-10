# Plan Review: Step 1 — Design the range-write contract

Verdict: APPROVE

The Step 1 contract now resolves the R001/R002 blockers. It chooses a dedicated `add_unavailable_date_range` write tool, keeps the category surface intentionally narrow, defines per-day write behavior around the current single-date event client, includes bounded inclusive date validation, and covers catalog/toolcheck/doc surfaces for Step 2.

A few implementation notes to preserve the approved contract:

1. Make the generated `external_id` fingerprint exact and stable over all fields that define the marker (`date`, normalized category, defaulted/trimmed name, and description, with nil vs empty-string semantics decided in code/tests). Avoid using the existing generic external-id duplicate shortcut in a way that would skip a same-external-id event whose writable fields do not match the requested unavailable marker.
2. Scan all same-day events when shaping duplicate/conflict metadata so duplicate skips do not accidentally hide unrelated same-day conflicts.
3. Keep the response status enum exactly test-covered: `created`, `partial`, and `skipped`, including mixed created/skipped ranges and conflict warnings.
4. Step 1 still has the “initial targeted event tests added/run” checkbox open; add the planned failing tests before moving to Step 2 implementation.

With those notes, the design is concrete enough to implement against.
