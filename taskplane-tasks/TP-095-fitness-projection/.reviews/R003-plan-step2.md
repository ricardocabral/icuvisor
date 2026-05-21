# R003 Plan Review — Step 2: Implement projection engine

**Verdict:** Needs changes / not yet reviewable

I read `PROMPT.md`, `STATUS.md`, the Step 1 review notes, and the current `get_fitness_projection` scaffolding. I do not see a concrete Step 2 implementation plan beyond the high-level checklist in `STATUS.md`. For a new public analyzer-family tool, the projection engine needs several implementation choices locked down before code is written; otherwise tests and response shape will be ambiguous.

## Required plan details before approval

1. **Starting-row extraction and insufficient-data behavior**
   - Define exactly how the engine fetches the starting fitness row, e.g. `ListAthleteSummary(ctx, {Start: start_date, End: start_date})`, how it handles zero rows, multiple rows, and rows whose `date` does not exactly match `start_date`.
   - Define how non-null CTL/ATL/TSB are detected. `intervals.SummaryWithCats` currently stores `Fitness`, `Fatigue`, and `Form` as `float64`, so null/missing upstream values are indistinguishable from real zeroes unless the engine checks `Raw` or the intervals model is changed to pointer fields. The plan must avoid treating missing current fitness data as a valid zero seed.
   - State whether insufficient current data returns a short user error or an analyzer response with `_meta.insufficient_sample:true`. Either can be acceptable if consistent, but the behavior needs to be explicit and tested.

2. **Deterministic projection formula**
   - Specify the exact update equations and order of operations. For example, whether each projected day uses `ctl_next = ctl_prev + (load - ctl_prev) / 42`, `atl_next = atl_prev + (load - atl_prev) / 7`, and `tsb_next = ctl_next - atl_next`, or whether TSB uses previous-day CTL/ATL. This choice materially changes the curve.
   - Define day indexing: whether day 1 is `start_date + 1`, whether the start row itself appears in the full series, and whether `horizon_days` is inclusive of the end date. `projectionHorizonDays` currently calculates `horizon_date - start_date`; the engine plan should match that.
   - Define rounding rules for all public numeric fields and whether internal recurrence keeps full precision until final shaping.

3. **Modeled load generation**
   - Define the ramp schedule precisely: daily load for week 1, when the first ramp increment applies, and whether `weekly_ramp_pct` compounds by week or is applied linearly.
   - Define recovery-week cadence semantics: whether week numbering is 1-based, whether the first recovery week for cadence 4 is week 4, and whether the recovery multiplier applies before or after ramping.
   - Define planned-load precedence. The Step 1 assumptions say planned loads replace modeled load on matching dates; the Step 2 plan should confirm that planned loads also bypass recovery-week reduction for those dates, or state otherwise.
   - Account for the Step 1 normalization issue from R002: planned-load dates are validated with `strings.TrimSpace` but currently returned untrimmed. The engine plan must either normalize them before use or build its override map from trimmed dates.

4. **Output contract and analyzer `_meta`**
   - Define the terse `result` fields, not just “summary”. A useful minimum would include start/end CTL/ATL/TSB, deltas, peak/min TSB or final TSB, horizon, and whether planned-load overrides were used.
   - Define the full `series` element shape and confirm it is emitted only with `include_full:true` via `encodeAnalyzerResponse`/`newAnalyzerResponsePayload`.
   - Define `_meta` values: `method`, `source_tools` (likely `get_fitness`), `n`, `missing_days`, `missing_action`, `min_samples`, `formula_ref` decision, assumptions, and boundaries. If a CTL/ATL projection `formula_ref` is added, include updating `icuvisor://analysis-formulas` in scope; if not, explicitly leave `formula_ref` empty and document assumptions in `_meta.assumptions`.

5. **Step 1 blockers should be resolved before engine work**
   - R002 requested changes for schema/decoder mismatch on `recovery_week_cadence`, horizon contract inconsistency, planned-load date normalization, and inaccurate `STATUS.md` review tracking. Step 2 should not proceed as if Step 1 is approved until those contract issues are fixed or consciously carried forward in the Step 2 plan.

## Recommendation

Revise `STATUS.md` (or add a short design note referenced from it) with the concrete engine algorithm, source-data handling, output shape, and analyzer meta contract above. Once those details are explicit, Step 2 should be straightforward to implement and review against deterministic golden tests.
