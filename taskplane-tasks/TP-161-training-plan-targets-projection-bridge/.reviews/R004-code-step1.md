# Code Review R004 — Step 1

Verdict: REVISE

Findings:

1. `internal/analysis/fitness_projection_test.go:80-81` expects `2026-06-08` to be `modeled_recovery_week`, but with the existing projection model day 0 is `2026-06-01`, day 1..7 are the first modeled week, and recovery cadence 2 starts on day 8 (`2026-06-09`). The Step 1 contract says uncovered dates keep the existing modeled ramp/recovery behavior, so this assertion would force an implementation to change existing recovery-week timing or special-case weekly targets incorrectly. Assert `modeled_ramp` on `2026-06-08` and/or `modeled_recovery_week` on `2026-06-09`.

2. `internal/tools/get_fitness_projection_test.go:149` expects `weekly_plan_target_filled_day_count == 6`, but the scenario itself asserts a weekly-target source on `2026-05-02` and has projected days `2026-05-02..2026-05-10` covered by the two weekly targets, with only `2026-05-05` overridden. That is 8 dates filled from weekly targets, not 6. Keeping this expectation will make the metadata undercount partial current-week fills even though the series uses them.

Verification:

- Ran `go test ./internal/analysis ./internal/tools -run 'FitnessProjection|TrainingPlan'`; failures are expected for missing bridge support, but the two assertions above are inconsistent with the intended contract/current model.
