# Plan Review: Step 2 — Investigate and implement activity tag handling if supported

**Verdict: Approved**

The revised Step 2 checklist in `STATUS.md` addresses the prior review concerns. It now commits to raw-payload string-array extraction, preserves the event-tag semantics, acknowledges shared `activityRow` consumers, includes schema/description updates, and defines targeted coverage before implementation.

## Implementation guardrails

- Keep activity tags sourced only from `activity.Raw["tags"]`; do not infer tags from names, notes, types, or custom fields.
- Preserve the same semantics as events: emit valid JSON string arrays in upstream order, emit explicit empty arrays, copy slices, and omit missing/null/non-array/mixed values.
- If adding the field to `getActivitiesRow`, initialize it before the `isStravaBlocked` early return if tags should survive on Strava-blocked rows, as the plan states.
- Update `getActivitiesOutputSchema()` and `activityReadOutputSchema()` when the field is added; consider whether `getTodayOutputSchema()` needs a short mention because completed activities inherit the row shape.
- Log the investigation result in `STATUS.md` discoveries, especially if current checked-in fixtures still lack activity `tags` and tests use synthetic payloads to exercise supported raw extraction.

With those constraints followed, the plan is safe to implement.
