# Plan Review R001 — Step 1

Verdict: Approved

The Step 1 plan is appropriately scoped for an audit step: it focuses on the two calendar write tools, asks the right idempotency/deduplication questions before implementation, requires documenting the chosen contract and upstream limitations in `STATUS.md`, and includes a targeted `go test ./internal/tools` baseline.

I verified the current targeted test command passes:

```sh
go test ./internal/tools
```

Execution notes to make the audit complete:

- Include the transport retry surface in the audit, especially `internal/intervals/client.go` `decideRetry` and `internal/intervals/events.go` `doJSONBody`/`AddOrUpdateEvent`. Current create writes use POST bulk while updates use PUT, so retry behavior differs.
- For `apply_training_plan`, explicitly document the existing conflict model: it fetches events across the plan range and treats any event on the date as a conflict. Note any limitations around matching the same plan/workout deterministically, same-day multiple planned events, list limits, and the race window between conflict fetch and create.
- For `add_or_update_event`, call out that the current tool client interface only has `AddOrUpdateEvent`; deterministic duplicate preflight for creates would require adding read capability (`ListEvents`) or settling for explicit duplicate warnings in Step 2.
- When recording the idempotency contract, distinguish repeated calls by the same caller from near-concurrent invocations. Pre-write checks can reduce duplicates but cannot fully eliminate races without upstream idempotency keys/conditional writes.
- Keep Step 1 read-only except for `STATUS.md` discoveries; implementation and new regression tests belong in Step 2.
