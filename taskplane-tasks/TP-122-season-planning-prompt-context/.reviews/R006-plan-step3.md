# Plan Review — Step 3

Verdict: **Changes requested before implementation.**

The Step 3 checklist has the right intent, but it is too optional given the current `add_or_update_event` surface. The existing examples only include one `RACE_B` update example and it lacks `target_load`; they do not clearly cover `RACE_A`/`RACE_C` or a full race-planning create shape with sport type, date, distance, expected duration, and load. Treat coverage as insufficient rather than leaving it as an “if needed” decision.

Required plan refinements:

1. Add a concrete implementation target: update `addOrUpdateEventInputExamples()` so race examples cover `RACE_A`, `RACE_B`, and `RACE_C` across the example set, and at least one race example includes `type`, `date`, `name`, `distance_meters`, expected duration (`moving_time_seconds` or `elapsed_time_seconds`), and `target_load`.
2. Add or strengthen a test that locks this planning-specific coverage, not just schema validity. A focused table/assertion in `add_or_update_event_test.go` or `event_category_schema_test.go` should fail if the public examples drop the race priority categories or the full race metric fields.
3. Keep the work examples/schema-only: do not change validation semantics, category pass-through behavior, write params, handler behavior, or response shaping unless an existing test exposes a real bug.
4. Reconcile the previous-step status before proceeding: `.reviews/R005-code-step2.md` says **REVISE** while `STATUS.md`/execution log says Step 2 was approved. Either fix the Step 2 findings and update status, or record why the review file is obsolete.
5. Run the targeted command from the prompt after the example/test change: `go test ./internal/tools -run 'AddOrUpdateEvent|InputExamples|EventCategory' -count=1`.

Baseline verification during review: `go test ./internal/tools -run 'AddOrUpdateEvent|InputExamples|EventCategory' -count=1` passes on the current code, but current race-example coverage is still insufficient for this task.
