# Plan Review: TP-011 Step 2 — Implement typed decoding

**Verdict: Approve.**

The revised Step 2 plan now covers the critical decoding invariants for this task: a typed intervals client method, nullable fields for scale-sensitive wellness values, raw JSON preservation, conservative native sidecar extraction, response rows built from raw/custom keys with typed overlays, and shaping through `response.Shape` with `RowCollections: []string{"wellness"}`. That is sufficient to proceed with implementation.

## What is now acceptable

- `sleepQuality`, `sleepScore`, and `sleepSecs` are explicitly decoded as separate nullable fields, avoiding zero-value ambiguity and sleep-scale collapse.
- Custom wellness keys are preserved by starting shaped rows from `Wellness.Raw` instead of marshaling only a static typed struct.
- Native provider fields are hoisted into `_native.<source>.<field>` and recognized top-level aliases are removed from terse rows, with the untouched upstream payload available only via `full` for `include_full:true`.
- The plan recognizes both nested provider objects and top-level aliases for Polar, Garmin, and Oura, while keeping unrecognized top-level fields as custom fields rather than inventing provenance.
- Provenance/staleness inference is deliberately deferred to Step 3, which keeps Step 2 focused on carrying enough typed/raw data without fabricating source metadata.

## Implementation notes to keep in mind

- When using `Wellness.Raw` as the row base, make sure claimed nested provider containers are not left duplicated beside `_native` in terse rows. Either remove the provider container entirely when it only contains claimed native fields, or remove claimed fields from it and keep only truly unclaimed data.
- Do not pass a static `Fields` list from `get_wellness_data` by default unless custom fields are still guaranteed to be returned. Wellness custom fields are part of the core contract, so default terse mode should not accidentally ask upstream for only known static fields.
- Keep `wellness.json` documented in `STATUS.md` as the JSON realization of the public `/wellness{ext}` route, since Step 1 used the shorthand `/wellness` endpoint form.
- Even though Step 6 owns the full fixture matrix, Step 2 implementation should include enough focused tests to lock in decoder behavior: path/query parameters, distinct sleep fields, raw custom-null preservation, and native extraction for at least one nested and one top-level shape.

No further plan changes are required before coding Step 2.
