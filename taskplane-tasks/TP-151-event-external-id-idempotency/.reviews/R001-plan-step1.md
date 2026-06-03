# Plan Review: TP-151 Step 1 — Design external_id contract

**Verdict: Not approved — plan is missing substantive design decisions.**

`STATUS.md` only records Step 1 as in progress with unchecked outcome boxes. It does not state the actual public contract decisions required by the prompt, so there is nothing concrete to approve before schema/API changes.

## Blocking issues

1. **Create/update/omit/clear semantics are not specified.**
   The implementation must know whether `external_id` is non-empty only, whether omitted means “leave unchanged” on update, and whether empty string/null can clear. Given upstream uncertainty, the conservative contract should be explicit, e.g. omit leaves unchanged, non-empty writes, and clear is rejected/not supported unless tested.

2. **`apply_training_plan` deterministic ID format is not defined.**
   The plan needs an exact namespace and tuple/hash, versioned and collision-resistant, avoiding provider prefixes like `hevy-`/`strava-`. It should also state whether IDs include raw plan/workout IDs or a hash, and what is exposed in dry-run/results.

3. **Read exposure is undecided.**
   Step 1 must choose whether event rows expose `external_id` in terse mode, only under `full`, or `_meta`, and apply that consistently to `get_events`, `get_event_by_id`, and `add_or_update_event` confirmations.

4. **Retry/preflight behavior with `external_id` is not described.**
   The plan should say how same-day duplicate matching changes when `external_id` is present: whether a matching external ID is treated as an idempotent duplicate even if other writable fields differ, whether differing external IDs make otherwise-identical rows non-duplicates, and what warning/conflict is returned.

5. **Upstream uncertainty is not recorded.**
   The task explicitly requires recording uncertainty in `STATUS.md`. Current discoveries only note that the write path lacks typed support; they do not record whether upstream supports clearing, uniqueness enforcement, or matching behavior for event `external_id`.

## Required revision

Update `STATUS.md` with a concrete Step 1 design section covering the four required decisions and known upstream uncertainties. After that, request another plan review before changing `internal/intervals/events.go` or public tool schemas.
