# Plan Review: Step 1 — Design the read-only contract

**Verdict: Needs revision before implementation.**

The plan has the right overall direction: compose existing read paths, keep terse defaults, gate raw upstream payloads behind `include_full`, and explicitly state no writes / no ATP synthesis. However, a few contract details are still ambiguous enough that implementation and tests could drift.

## Required fixes

1. **Define the default week anchor.**
   - The plan allows no required args and an optional `week_start`, but does not say what week is used when omitted.
   - Existing `weekly_planning` prompt guidance says to use the upcoming athlete-local week unless `week_start` is supplied. The contract should explicitly choose that or another deterministic rule, including how “today is Monday” behaves.

2. **Revisit or justify `core` tier placement.**
   - The discovery says `get_planning_context` “should be core-tier”, but one of its explicit source sections is `get_training_plan`, which is currently a full-tier tool.
   - If this new tool is core, it effectively exposes a training-plan summary through the core toolset. That may be acceptable for token efficiency, but the plan should state the rationale and update tier/catalog tests accordingly; otherwise mark it full-tier.

3. **Make window sizes exact.**
   - `fitness_context` is described as “current row plus recent 7-day rows or a compact summary”; choose one exact shape and date range.
   - `upcoming_races` says “near-future window”; define the number of days and event categories included.
   - These need to be contract-level decisions so tests can assert deterministic output.

4. **Define event classification rules.**
   - The plan says split week events into “planned workouts vs races/other events”, but should name the category logic (`WORKOUT`, `RACE`/`RACE_*`, `NOTE`, etc.) and how unknown categories are handled.

## Suggested additions

- Add `_meta.source_tools` plus counts for each section as planned, but consider whether profile/timezone resolution should be represented separately (for example `profile_source` or including `get_athlete_profile`) since timezone is part of the contract.
- Add caveat codes, not only prose strings, for stable tests: e.g. `no_active_training_plan`, `no_week_workouts`, `no_fitness_rows`, `read_only_no_atp`.
- Include validation behavior for `week_start` in the contract: trim, require `YYYY-MM-DD`, normalize to Monday, and reject invalid dates with a short user-facing error.

Once these are captured in `STATUS.md` discoveries, the plan should be ready for implementation.
