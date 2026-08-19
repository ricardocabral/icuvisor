# Activity RPE write contract gap

Status: read-only support only (2026-08-19)

The verified activity read contract exposes these upstream keys and types:

- `commute`: boolean
- `feel`: integer, athlete-reported 1–5 feeling scale
- `icu_rpe`: integer, athlete-reported 1–10 rating of perceived exertion

The sanitized activity fixtures and OpenAPI evidence verify GET/list decoding and
preserve the native keys in full payloads. They do not verify a field-specific PUT
request or a successful write/read round trip for `icu_rpe`. The generic PUT
operation references the full Activity schema and identifies Strava restrictions,
but that is not evidence that an `icu_rpe` property is accepted for writes.

Accordingly, `update_activity` does not advertise `rpe`, `perceived_exertion`, or
native `icu_rpe`; these fields are rejected by strict request decoding before an
updater or HTTP client is called. `feel` and `commute` remain read-only as well.
No conversion between the 1–5 feel scale and the 1–10 RPE scale is performed.

Revisit this gap only with non-live, source-honest evidence of the exact writable
request key and a successful native-field round trip.
