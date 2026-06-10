# Plan Review: Step 1 — Design the range-write contract

Verdict: Changes requested

The current Step 1 plan is not yet a design; it mostly restates the task checklist. Before implementation, please record a concrete contract in `STATUS.md` (or another task artifact) covering these decisions:

1. Public API surface: choose either a dedicated `add_unavailable_date_range` tool or an `add_or_update_event` extension. If dedicated, plan the required catalog surfaces too: `internal/toolcatalog`, registry group, schema snapshots/hash, README/PRD, coach ACL eligibility, `RequirementWrite`, and toolset placement.

2. Exact request/response contract: define field names, required fields, inclusive date semantics, max range cap, terse/default `include_full` behavior, and `_meta` counts/details for created/skipped duplicates. Return shape is part of the MCP API and should be designed before tests.

3. Category and idempotency semantics: specify the closed accepted categories/aliases (likely only `HOLIDAY`, `SICK`, `INJURED`, with any allowed time-off aliases normalized explicitly) and whether the implementation creates one multi-day upstream event or per-day events. Given the current client only writes a single `Date` per event, if using per-day events define repeated-call behavior, partial duplicate behavior, external_id strategy, and whether existing nonmatching same-day events only warn or block.

4. Test plan: add failing tests for valid range creation, repeated range idempotency, mixed existing/missing days, invalid/reversed/excessive ranges, unsupported categories, and safety invariants (no delete-mode bypass, no workout overwrite by default). The targeted command in the prompt is fine, but the tests need to encode the above contract.

Once these decisions are captured, Step 1 should be reviewable and Step 2 can implement against a stable contract.
