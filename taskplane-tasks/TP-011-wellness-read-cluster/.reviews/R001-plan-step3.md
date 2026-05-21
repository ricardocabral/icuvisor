# Plan Review: TP-011 Step 3 — Provenance and staleness `_meta`

**Verdict: Approve.**

The revised Step 3 plan addresses the prior blockers and is ready for implementation.

## What is now covered

- Provenance assembly stays in the wellness tool row builder, preserving the low-level decoder as a raw/typed payload layer rather than a semantic inference layer.
- Bridge timestamp selection explicitly excludes `updated`, which avoids treating a later manual row edit as proof that provider data refreshed.
- `_meta.provenance.<field>.fetched_at` is always present, using the non-null sentinel string `"unknown"` when no explicit bridge/import/provider timestamp exists, so terse null-stripping will not remove the required key.
- Unknown timestamps are excluded from staleness calculations.
- Staleness uses the wellness row date reference and a strict `>24h` boundary; exactly 24h old remains fresh.
- Source inference is conservative and based on explicit provider/native evidence. Always-device/normalized bridged fields still get provenance with `source: "unknown"` when provider evidence is missing, while dual-use manual/body fields only get provenance when bridge evidence exists.
- `native_scale` now remains informative when the field scale/unit is known even if provider source is unknown, instead of defaulting everything to `"unknown"`.
- `_meta.stale_reason` has deterministic one-line wording and is only emitted for stale rows.
- The response shaper change is explicitly required: default debug stripping must preserve `_meta.provenance.*.fetched_at` while still removing top-level/debug-only `fetched_at` and `query_type`.

## Implementation notes

- When implementing timestamp extraction, inspect both top-level raw keys and any retained nested provider objects from `row.Raw`; Step 2 supports nested provider native payloads, so fixture/provider timestamps may reasonably appear there.
- Normalize provider labels before comparing/emitting them, and fall back to `unknown` for unrecognized marker values rather than echoing arbitrary upstream/custom strings.
- If multiple stale provider timestamps are present in a single row, keep the stale reason deterministic, e.g. choose a stable provider order or use the generic bridge-data reason.

No further plan changes are required before Step 3 implementation.
