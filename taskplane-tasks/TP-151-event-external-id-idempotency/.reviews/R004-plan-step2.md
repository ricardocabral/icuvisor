# Plan Review: TP-151 Step 2 — Implement event write/read support

**Verdict: APPROVE**

The Step 2 checklist is aligned with the accepted Step 1 contract: add typed/write-payload support, expose `external_id` through `add_or_update_event`, keep same-day preflight protections, expose the read-row audit field, and run targeted event tests.

## Implementation requirements to carry into the step

1. Preserve the Step 1 omit semantics in code and tests: trim `external_id`; forward only non-empty values; do not send empty string or `null` as a clear operation.
2. Test both upstream body shapes: POST bulk create wraps the event payload in an array, while PUT update sends a single object. Both should include `external_id` only when provided.
3. Make the duplicate-preflight tests explicit for the new contract: a same-day matching `external_id` should skip create even when other writable fields drift, and missing/different `external_id` should not disable the existing exact-field duplicate check.
4. Add/read-row coverage for terse `external_id` exposure and null/absent omission. Since `eventRow` is shared, a `get_events` or `get_event_by_id` test plus the write response shape is sufficient.
5. Update the `add_or_update_event` schema/description text when adding the field so it no longer claims there is no idempotency key.

Schema snapshots, docs, and broader routing expectations can remain in Step 4 as planned.
