# Activity metadata write contract and evidence

Status: typed corrections with per-call verification (2026-10-27).

`update_activity` accepts `commute`, `feel` (1–5), and `rpe` (1–10, sent as `icu_rpe`) alongside tags, subtype and manual-only distance corrections. The two subjective scales remain distinct; no conversion is performed. Null is rejected for these new fields, and each supplied value must be observed before it is listed in `fields_updated`.

## Public sources

The [public Activity update schema](https://intervals.icu/api/v1/docs) documents native keys and types. [David's manual-activity API clarification](https://forum.intervals.icu/t/api-access-to-intervals-icu/609/306) names `icu_rpe`, `feel`, and `distance` explicitly. A generic full-Activity PUT schema alone still does not prove that every field is writable, or that a particular source accepts an edit.

## Runtime evidence

The client compares each new requested native field with the PUT response. If a value is missing or differs, it performs one bounded activity readback. Only matching fields appear in `fields_updated`; submitted values remain separately labelled `requested`. `_meta.verification` reports confirmed, missing and mismatched fields, evidence source, and an actionable caveat. An unverified or mismatched request is not returned as an unqualified successful update and is not automatically written again.

Fixture tests prove request mapping and these verification/refusal paths. They are not a live upstream round-trip claim. No live athlete probe was performed for the October 27 implementation.

## Remaining limits

- Distance corrections require explicit manual-source evidence and no recorded file.
- Recorded-file/unknown-source distance, moving/elapsed-time changes, gear assignment, native input aliases, and arbitrary field passthrough are not exposed.
- Fields computed or reset upstream may be rejected or fail runtime verification; inspect the activity before an intentional retry.
