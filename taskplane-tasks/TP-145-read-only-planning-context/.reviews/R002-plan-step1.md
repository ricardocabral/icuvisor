# Plan Review: Step 1 — Design the read-only contract

**Verdict: Needs revision before implementation.**

The revised plan addresses most R001 concerns: it moves the tool to full-tier because it composes `get_training_plan`, includes `get_athlete_profile` in `_meta.source_tools`, defines caveat codes, preserves read-only/no-ATP semantics, and gives deterministic event classification rules. Those are good contract choices.

Two contract details are still likely to create implementation/test drift.

## Required fixes

1. **Fix the fitness window for supplied `week_start`.**
   - Current discovery says fitness fetches `week_start - 7 days` through `as_of_date` inclusive.
   - This becomes invalid for a future planning week more than one week ahead (`start > as_of_date`) and unbounded/too large for historical weeks (`start` far before `as_of_date`).
   - Choose a bounded deterministic rule, for example “always fetch `as_of_date - 7 days` through `as_of_date` inclusive for current fitness context,” or explicitly clamp/validate supported `week_start` ranges.
   - Add the chosen rule to tests, including a future `week_start` case.

2. **Define event/race fetch limits and truncation behavior.**
   - Week events and the 84-day race scan need explicit limits, especially the race scan if implemented by fetching all categories then filtering `RACE`/`RACE_*`.
   - State whether the tool uses `Limit: 100`, `Limit: 500`, category-filtered calls, or another strategy.
   - Expose deterministic `_meta` fields/caveats for truncation so a race is not silently missed behind many non-race events.

## Suggested clarifications

- Specify exact caveat-code conditions, e.g. whether an empty week emits both `no_week_events` and `no_week_workouts`.
- Spell out `include_full` scope per section: event `full`, fitness `full`, and training-plan raw nested payloads only when `include_full:true`.
- Keep the implementation composing client/read helpers directly; do not route through write-capable tools or introduce any create/update/delete methods.

Once the fitness-window and event-limit contracts are captured in `STATUS.md`, the plan should be ready to implement.
