# Plan Review — TP-093 Step 1: Design deterministic contracts

**Verdict:** Changes requested

The new `CONTRACT.md` is a substantial improvement over R001: it defines public inputs, default terse/full behavior, source priorities, analyzer `_meta`, and missing/insufficient states for all four tools. However, several parts still conflict with the PRD/current typed sources or leave deterministic behavior ambiguous enough that Step 2 would likely implement an incompatible public contract.

## Blocking findings

1. **`compute_zone_time` omits the PRD-required polarization output.**
   PRD §7.2.C defines `compute_zone_time` as “sum of time per power / HR / pace zone ... with polarization index” (`docs/prd/PRD-icuvisor.md:279`). The contract currently says `_meta.formula_ref` is omitted because summation has no formula ref and lists no `polarization_index` in the result (`CONTRACT.md:32`). Either add the same polarization fields/formula ref/boundary handling as load balance to `compute_zone_time`, or explicitly document a PRD divergence for approval before implementation.

2. **The summary-zone source does not satisfy sport-filtered or metric-specific zone requests as written.**
   The contract says summary-backed tools filter `SummaryWithCats.byCategory` for `sport` (`CONTRACT.md:12`) and makes `SummaryWithCats.TimeInZones` authoritative for `compute_zone_time` (`CONTRACT.md:26`). In the current typed source, `TimeInZones` is only on the all-up daily `SummaryWithCats`; `CategorySummary` has totals/load/distance but no zone arrays (`internal/intervals/fitness.go:35-37`, `internal/intervals/fitness.go:56-67`). Also, the summary `TimeInZones` is described as metric-agnostic (`CONTRACT.md:26`) while the request enum is explicitly `power` / `heart_rate` / `pace` (`CONTRACT.md:22`). The contract needs to pin deterministic behavior here: e.g. only use summary zones when no `sport` is requested and when the upstream zone family can be proven to match the requested `zone_metric`; otherwise fall through to activity-level precomputed arrays or return unavailable. Do not let a sport-filtered or HR/pace request silently use all-sport/unknown-family summary zones.

3. **Activity-level source acquisition is under-specified.**
   `compute_zone_time` says extended metrics are used “only when activity IDs are already available from `get_activities`” (`CONTRACT.md:27`), but the request schema has only a date window and optional sport (`CONTRACT.md:19-20`). Step 2 needs a deterministic rule for fetching activity IDs: whether the tool itself calls `get_activities`, how it handles pagination/page caps, whether `get_activities` is included in `_meta.source_tools`, and what happens when the activity list is truncated. Without that, source priority and missing-source counts will vary by implementation.

4. **`compute_baseline` is missing the PRD-required wellness interpretation.**
   PRD §7.2.C requires a z-score plus a “suppressed” / “elevated” flag for wellness metrics (`docs/prd/PRD-icuvisor.md:281`). The response shape currently has `z_score` but no interpretation/direction field (`CONTRACT.md:75`). Add a deterministic `interpretation`/`state` contract, including which metrics are inverted or non-directional, what z-score thresholds are used, and where those assumptions appear in `_meta.assumptions`.

5. **`compute_compliance_rate` does not fully cover the PRD/event-type contract and aggregate deltas.**
   PRD §7.2.C calls for “mean delta to target, per sport / event type” (`docs/prd/PRD-icuvisor.md:283`). The request schema includes `sport` and `category` but no event `type` filter or grouping (`CONTRACT.md:81-82`), and the result includes per-row percent difference only in full `series`, with no aggregate `mean_delta_percent` / `mean_delta_*` fields (`CONTRACT.md:94-96`). Add exact event-type filter/group semantics and aggregate delta fields to the terse result so the default response satisfies the catalog promise.

6. **Compliance pairing is still not deterministic enough for implementation.**
   The plan uses open-ended link keys (“or equivalent ID fields”) (`CONTRACT.md:87`), does not say whether one activity may be consumed by more than one scheduled event during auto-pairing (`CONTRACT.md:88`, `CONTRACT.md:94`), and leaves time-target matching ambiguous between moving vs elapsed target/actual fields (`CONTRACT.md:86`, `CONTRACT.md:94`). Pin exact accepted raw keys, one-to-one matching behavior, tie-break order across both events and activities, and the actual field selected for `time_target` vs `elapsed_time_target`.

## Non-blocking notes

- The explicit “no `get_activity_streams` fallback” decision is acceptable as a conservative contract if you keep it aligned with the roadmap’s statement that `compute_activity_segment_stats` is the only analyzer that touches streams by default (`ROADMAP.md:101`). Keep the “upstream coverage audit” follow-up in mind (`ROADMAP.md:106`).
- Consider adding maximum date-window limits (or explicitly stating there are none) for all four schemas. R001 asked for limits, and the current contract only rejects inverted windows (`CONTRACT.md:3`).
- Avoid “future/unknown” source language in the final contract where possible (`CONTRACT.md:27`); exact raw keys make golden tests and schema stability easier.

Once the above items are pinned in `CONTRACT.md`/`STATUS.md`, Step 1 should be ready to approve for implementation.
