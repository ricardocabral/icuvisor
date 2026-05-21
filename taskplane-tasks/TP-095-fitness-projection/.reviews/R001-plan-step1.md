# R001 Plan Review — Step 1: Define projection assumptions

**Verdict:** Needs changes / not yet reviewable

The task files only show the high-level Step 1 checklist. I do not see a concrete implementation plan for the projection contract, request schema, or `_meta` assumptions. Because `get_fitness_projection` is a new public analyzer-family tool, Step 1 needs to lock down those API details before Step 2 starts; otherwise the projection engine and tests will be under-specified.

## Required plan details before approval

1. **Request schema contract**
   - Define whether callers provide `horizon_days`, `horizon_date`, or exactly one of the two, including min/max bounds and inclusive/exclusive date semantics.
   - Define `ramp_percent` units precisely, e.g. weekly percentage change in daily/weekly training load, and set numeric validation bounds.
   - Define recovery-week cadence semantics. Cadence alone is insufficient unless the plan also defines the recovery-week load reduction/multiplier/default and whether it is user-configurable.
   - Define `planned_load` shape if supported: daily vs weekly rows, required fields, units (`training_load`/TSS-like load), date validation, gap handling, ordering, duplicate handling, and precedence versus `ramp_percent`/recovery cadence.
   - Include `include_full` behavior in the schema: terse default summary; projected curve only when `include_full:true` per repository rules.

2. **Deterministic model assumptions**
   - Specify the CTL/ATL update formula and constants to be used. If using standard time constants, document the exact CTL and ATL horizons and the resulting TSB definition.
   - State how the starting point is selected from `get_fitness`/athlete summary rows, including what counts as insufficient current fitness data.
   - State rounding rules, timezone/date boundary behavior, and maximum projection length to keep payloads bounded.
   - Avoid claiming upstream parity unless there is an upstream documented formula. The meta should label this as an icuvisor deterministic simulation, not a physiological prediction or intervals.icu forecast.

3. **Analyzer `_meta` contract**
   - Plan how to satisfy the existing analyzer meta skeleton: `_meta.method`, `source_tools`, `n`, `missing_days`, `missing_action`, `insufficient_sample`, and a stable `formula_ref` if applicable.
   - Add projection-specific assumptions to `_meta` in a typed way, such as horizon, ramp percent, recovery cadence, recovery multiplier, planned-load mode, CTL/ATL constants, source start date, and boundary/capping notes.
   - If a new formula ref is introduced, include updating `icuvisor://analysis-formulas`/the formula registry in scope or explicitly defer with rationale.

4. **Unsupported/free-form model rejection**
   - Make the schema `additionalProperties:false` and do not accept a free-text `model`, `physiology_model`, or formula override.
   - If an enum field is planned for model/version selection, it should only allow the single supported deterministic model initially and reject anything else with a short actionable error.

## Recommendation

Revise Step 1 with a concrete API/model contract before implementation. Once those choices are documented in `STATUS.md` (or a small design note referenced from it), the plan should be straightforward to approve and Step 2 can implement against stable assumptions.
