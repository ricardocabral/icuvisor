# Review R002 — Plan Review for Step 1

**Verdict:** Changes requested

The updated Step 1 notes now cover most of the required measurement method: they identify fixture roots, list the current precomputed zone-time field names, define per-family metrics, and correctly state that the current analyzers are precomputed-only rather than actually doing stream fallback. However, two parts remain ambiguous enough that Step 2 could produce non-reproducible or inflated results.

## Blocking issues

1. **Exclusion precedence is contradictory and will include event fixtures.**
   The plan says to exclude events, workout library, custom items, activity intervals, analyzer goldens, etc. “unless an object matches an eligible rule.” But the activity/detail-like rule is `id` plus `start_date` or `start_date_local`, which also matches multiple event fixtures such as `internal/intervals/testdata/events/detail.json`, `events/inconsistent/synthetic_list.json`, and `events/note_create_response.json`. Those are not activity/detail fixtures and would be counted as missing precomputed zones, inflating fallback coverage.

   Define path/type exclusions as taking precedence over generic shape detection, or tighten the activity rule so event-like objects are rejected (for example, exclude paths under `events/` and objects with event-only markers such as `category`, `workout_doc`, `show_as_note`, etc.). The script should not rely on broad `id + start_date_local` detection alone.

2. **Per-family opportunity eligibility is still underspecified.**
   The plan defines `precomputed_count`, `fallback_count`, and `unknown_count` per `(fixture_unit, metric_family)`, but it does not say how a fixture becomes “eligible for that family.” Without this, two implementations could disagree on whether a Ride without power zones is a `power` fallback, `heart_rate` fallback, three family fallbacks, or `unknown`.

   Add deterministic family-applicability rules before Step 2. For example: a family is precomputed if one of that family’s valid arrays is present; fallback only when there is a clear family signal but no valid array; unknown when the object is activity-like but lacks such signal. The exact signal set should be recorded (e.g. family-specific load/average/max fields, stream-key metadata if present, sport-type defaults if you choose to use them). If the intent is to count every eligible activity as an opportunity for all three families, state that explicitly, but be aware it will overstate fallback for fixtures where a metric was never recorded.

## Non-blocking suggestions

- Record that exclusion rules are applied before object-shape rules, then object-shape rules are applied per root object or root-array element.
- Keep the “positive total seconds” validation rule; it is a good usability criterion even though `zoneSliceForMetric` currently accepts any non-empty numeric slice, because a zero-sum slice still cannot produce a useful analyzer result.
- Consider naming the threshold policy as: “operator decision; for this audit, any non-zero measured missing-precomputed opportunity is documented as risky evidence, not as an automatic pass/fail threshold.” That avoids making a de facto threshold after seeing results.

Once the exclusion precedence and family-eligibility rules are made explicit in `STATUS.md`, Step 1 should be ready for implementation.
