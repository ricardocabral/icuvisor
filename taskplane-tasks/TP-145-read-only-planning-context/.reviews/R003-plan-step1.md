# Plan Review: Step 1 — Design the read-only contract

**Verdict: APPROVED for implementation.**

The Step 1 discoveries now define a sufficiently deterministic contract for `get_planning_context`. The prior blockers are addressed:

- Default and supplied `week_start` behavior is deterministic and athlete-local.
- Tier placement is corrected to **full-tier** because the tool composes `get_training_plan`.
- Fitness context is bounded to the current as-of 7-day window, independent of planning-week drift.
- Race and week-event windows use explicit limits, with conservative truncation metadata/caveats.
- Event classification, caveat codes/conditions, `_meta.source_tools`, read-only/no-ATP metadata, and `include_full` scope are explicit enough for tests.

Implementation reminders, not blockers:

- Register the tool in both runtime catalog and `internal/toolcatalog` so coach ACL/known-tool tests stay aligned.
- Keep the composed client interface read-only (`GetAthleteProfile`, `ListEvents`, `GetTrainingPlan`, `ListAthleteSummary`) and use tests/fakes that would fail if a write method were introduced.
- Add catalog tier coverage for `get_planning_context` as full-tier, plus tests for future `week_start` still using the current fitness window and for `len(events) >= 500` truncation caveats.
