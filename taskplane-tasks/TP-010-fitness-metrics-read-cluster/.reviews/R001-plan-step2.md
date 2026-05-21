# Review R001 — Plan review for Step 2

Decision: **Approve**

I re-read `PROMPT.md`, the current `STATUS.md` Step 2 implementation plan, the PRD/Roadmap anchors, and the existing response/tool patterns. The revised plan now addresses the prior blockers:

- `get_power_curves` is tied to the athlete `power-curves.json` endpoint with a required/default sport strategy instead of recomputing curves client-side.
- `get_best_efforts` has an explicit default sport fan-out and metadata for unavailable/empty families.
- Curve-backed terse responses now have a compact default bucket contract, requested/default bucket echoing, missing-bucket metadata, and an `include_full:true` gate for raw upstream arrays/maps/ranks.
- Date validation, paired date semantics for all-time best-effort curves, strict JSON decoding, profile-derived timezone/unit metadata, `response.Shape`, and conditional registry wiring are all called out and align with current project conventions.

## Notes for implementation

These are not blockers, but they are worth preserving in code/tests:

- Keep unit-disambiguated response keys for any distance/pace values (`distance_km`/`distance_mi`, `pace_seconds_per_km`/`pace_seconds_per_mile`/swim-specific pace labels) rather than relying only on `_meta.units`; this follows the PRD response-shaping rule and existing activity/profile tools.
- For `get_training_summary`, make the neutral naming explicit in the structs (`training_load`, not upstream-specific or coach-language aliases) and aggregate zone arrays by upstream zone order with the zone order described in `_meta`.
- For all curve tools, do not zip and emit full `secs`/`values` arrays in terse mode; only requested/default buckets should appear unless `include_full:true`.
- Add table-driven tests around the now-defined bucket defaults and missing-bucket metadata, in addition to the Step 4 cases already listed.

The plan is sufficiently specific to implement Step 2.
