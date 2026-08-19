# AlphaHRV and DFA alpha1 evidence boundary

Status: upstream evidence incomplete; no canonical AlphaHRV activity field is advertised.

## Verified evidence

The pinned public Intervals.icu OpenAPI snapshot exposes `average_dfa_a1` as a numeric field on `Interval` and `IntervalGroup`. Icuvisor surfaces that field only from `GET /api/v1/activity/{id}/intervals` as `get_extended_metrics.intervals[].dfa_alpha1`:

- scope: analyzed interval only;
- unit: unitless, as recorded in the response provenance;
- availability: conditional and omitted when the upstream value is null, absent, or malformed;
- provenance: `source_field: average_dfa_a1`, with no device or algorithm attribution.

The existing sanitized TP-251 interval fixture is the only committed upstream evidence fixture for this field. Its value is a contract sample, not a claim that every activity or device supplies DFA alpha1.

The same OpenAPI snapshot exposes wellness `hrv` and `hrvSDNN` on daily wellness rows. Those fields are a separate scope and are never joined to an activity interval or renamed as DFA alpha1/AlphaHRV.

## What is not verified

Permitted evidence does not establish any of the following:

- an AlphaHRV-branded activity or interval key;
- a universal `alpha_hrv` custom-field code;
- a unit, scale, algorithm, device source, or physiological meaning for an athlete-defined custom field;
- a relationship between `average_dfa_a1` and a named device, readiness score, threshold, training state, or medical conclusion;
- reliable population or device conditions beyond the conditional upstream field presence.

A field name, custom-item label, screenshot, or value alone is not sufficient to promote a custom field to a canonical metric. Canonicalization requires a documented upstream key, scope, unit/scale, provenance, and representative sanitized evidence.

## Safe custom-field workaround

When an athlete has an `ACTIVITY_FIELD` custom item, call `get_activities` or `get_activity_details` with that exact upstream field code in `custom_fields`. Icuvisor validates the code against the athlete's custom-item definitions, fetches it only when explicitly selected, preserves the original code, and returns only non-null scalar values. The response metadata marks unit, scale, algorithm, physiology, and source device as `not_provided`; absent, null, and non-scalar values remain unavailable rather than becoming zero.

An explicit custom field named `alpha_hrv` is still an athlete-defined custom value. It is not AlphaHRV evidence and must not be relabeled, converted, correlated to wellness HRV, or used to infer readiness, thresholds, injury risk, or training prescriptions. For verified DFA alpha1, use the interval `dfa_alpha1` field and its source metadata instead.
