# Plan Review R003 — Step 2

Verdict: Approved

The Step 2 plan is aligned with the Step 1 discovery and the task requirements: add regression coverage first, implement deterministic duplicate/conflict handling using only existing upstream event fields, make dry-run conflict outcomes visible, and verify with `go test ./internal/tools`.

I verified the current targeted baseline still passes:

```sh
go test ./internal/tools
```

Execution notes to keep the implementation complete:

- Cover both write paths. `add_or_update_event` currently only depends on `EventWriterClient`; duplicate preflight for creates will need read capability (`ListEvents`) or an explicit, tested limitation. Do not apply duplicate checks to `event_id` updates as if they were creates.
- Define stable response metadata for exact duplicates vs same-day conflicts. Exact duplicate creates should be idempotent/skip-like; non-identical same-day events should be clearly warned/skipped according to the chosen contract.
- For `apply_training_plan`, update the in-memory conflict index after each create and handle duplicate workouts within the same plan/date range before writing. If feasible, re-check the target date immediately before each create to narrow the concurrent-call race; still document that upstream has no CAS/idempotency-key guarantee.
- Dry-run output should use explicit conflict reasons (for example exact duplicate vs existing event on date) and include the upstream event ID when available, with no mutations.
- Tests should include repeated `apply_training_plan` over the same range, duplicate same-day plan entries, direct same-day `add_or_update_event` planned-event creates, and clear assertions that skipped duplicates do not call `AddOrUpdateEvent`.
- Keep the change scoped to existing upstream fields; do not add model-controlled confirmation flags or store hidden idempotency state.
