# Plan Review — TP-228 Step 1

## Verdict: REVISE

No Step 1 implementation plan was submitted. The task prompt and unchecked STATUS checklist state requirements, but do not make the required boundary decisions or identify the changes/tests that will implement them.

A revised plan must explicitly specify:

1. The exact client contract: update is `PUT /athlete/{athleteId}/sport-settings/{id}?recalcHrZones=true|false` with the existing sparse JSON body; apply is a separate `PUT .../{id}/apply` with neither query parameters nor a semantic body; `UpdateSportSettings` never calls apply.
2. The MCP migration: remove `effective_date` from the request type, required fields, validation, error text, examples, params, response metadata, snapshots, and generated catalog. Since decoding is strict, a legacy `effective_date` must be rejected before any upstream call.
3. How `recalc_hr_zones` defaults to true while preserving an explicitly supplied false. A plain Go `bool` cannot distinguish an omitted value from false; use presence-aware decoding (for example `*bool`) and pass the resolved value through `WriteSportSettingsParams` to the query encoder.
4. The exact truthful metadata field and semantics (it may report the requested HR-zone recalculation boolean, not that any date-scoped or historical recomputation is pending/completed). Retire `effective_date` and `recompute_pending` claims.
5. Targeted regression cases for omitted/default-true and explicit-false query values, the absent apply call, and rejection of legacy/unknown arguments, in addition to the required no-query/no-body apply wire assertion.

This is safety-critical because the current implementation converts `EffectiveDate` into an implicit apply request (`internal/intervals/sport_settings.go`) and the existing public error/schema still instruct callers to provide that date (`internal/tools/update_sport_settings.go`).
