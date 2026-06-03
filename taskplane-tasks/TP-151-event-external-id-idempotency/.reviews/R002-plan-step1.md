# Plan Review: TP-151 Step 1 — Design external_id contract

**Verdict: APPROVE**

The revised `STATUS.md` now records the substantive contract decisions required before public schema/write-path changes:

- `add_or_update_event.external_id` is optional, trimmed, write-only when non-empty, omitted otherwise, and does not expose unsupported clear/null semantics.
- `apply_training_plan` uses an icuvisor-owned, versioned prefix plus a hash of a deterministic plan/workout/date tuple, avoiding raw ID leakage and known provider prefixes.
- Event read rows will expose `external_id` in terse mode when upstream returns it, with full mode continuing to include raw payloads.
- Same-day preflight behavior with `external_id` is defined, and upstream uncertainty is recorded conservatively.

## Non-blocking implementation notes

1. Pin the hash input serialization and digest length in code/tests (for example, canonical JSON array or length-delimited fields) so `icuvisor-plan-v1-...` stays stable across refactors.
2. When implementing external-ID duplicate skips where writable fields drift, include the existing event ID and an idempotency/duplicate warning in `_meta` so callers can distinguish retry success from an applied content change.
3. Consider making dry-run exposure deterministic rather than “may show” hashed `external_id`; tests should make the chosen output shape explicit.
4. Clean up the stale/conflicting R001 execution-log note in `STATUS.md` when closing the step, if the task runner expects review history consistency.

Proceed to Step 2.
