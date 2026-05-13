# Plan Review: TP-020 Step 2 (`link_activity_to_event`)

**Verdict: Changes requested**

I read `PROMPT.md`, `STATUS.md`, the PRD/roadmap anchors for the event write cluster, and the existing event/activity client and tool patterns. I do not see a concrete Step 2 implementation plan beyond the generated checklist in `STATUS.md`. The checklist identifies the right feature, but it is not yet detailed enough to implement safely.

## Blocking gaps to address in the plan

1. **Upstream endpoint semantics are unspecified**
   - The plan must identify the exact intervals.icu method/path, HTTP verb, request body/query parameters, and expected response shape for linking an activity to an event.
   - Do this from clean-room sources only: public API docs, black-box testing, or the allowed MIT reference if needed. Do not consult GPL code.
   - Add a typed `internal/intervals` method/params type rather than driving the write with ad-hoc `map[string]any`.

2. **ID “normalization” is currently ambiguous**
   - `PROMPT.md` says `activity_id` and `event_id` are normalized via existing helpers, but the repo only has athlete-ID normalization in `internal/config` and simple trim/validation patterns for activity/event IDs.
   - The plan should explicitly define the contract: trim surrounding whitespace, reject empty values, preserve upstream IDs exactly otherwise, and do not run athlete-ID normalization on activity/event IDs.
   - If the upstream endpoint requires numeric event IDs or strips an `i` prefix for activity IDs, document that as endpoint-specific conversion with tests; otherwise preserve caller-provided IDs.

3. **Mismatched-date warning needs a concrete data source**
   - The required `_meta.warnings` case cannot be implemented from only two IDs unless the tool fetches or receives activity/event dates.
   - The plan should specify whether the tool will call `GetActivity` and `GetEvent` before or after the link, which fields it compares (`activity.start_date_local` date prefix vs `event.start_date_local` date), and how it behaves if either read fails.
   - Define a stable warning shape, e.g. objects with `code`, `message`, `activity_date`, and `event_date`, rather than an unstructured string list.

4. **Idempotent re-link behavior is not defined**
   - The plan must state what “idempotent re-link” means against upstream behavior: repeated link of the same `activity_id`/`event_id` should return success and the same confirmation, while conflicts/already-linked responses should be mapped deliberately if upstream returns a non-2xx.
   - If the HTTP verb is not naturally idempotent, the plan must also specify retry policy to avoid duplicate/ambiguous mutations. Follow the Step 1 write-helper precedent: do not blindly retry unsafe POSTs.

5. **Tool response contract is missing**
   - Define the public response shape before coding. It should include at least `activity_id`, `event_id`, a terse success/status field, and `_meta` with warnings and server/debug metadata through the existing shaping path where appropriate.
   - If the upstream returns linked activity/event objects, decide whether raw payloads are exposed behind `include_full:true` or deliberately omitted. Either way, document this in the schema and tests.

6. **Write capability registration and absence in `none` mode must be planned**
   - `link_activity_to_event` is a write tool. It must be marked `RequirementWrite`, registered in `safe` and `full`, absent in `none`, and must not accept a `confirm` argument.
   - Add real-tool registry/safety coverage or extend the existing capability tests so this acceptance criterion is verified for the new tool, not only for toy test tools.

7. **Schema/catalog/docs updates are not called out**
   - Add `internal/tools/link_activity_to_event.go` with strict input schema (`additionalProperties:false`) and LLM-readable descriptions, including that the tool is the manual escape hatch when auto-pairing misses (forum #97) and is non-destructive.
   - Extend the schema catalog stub/snapshot generation, commit `internal/tools/schema_snapshot/link_activity_to_event.json`, and update README and `CHANGELOG.md` under `[Unreleased]` for this specific tool.

8. **Test plan needs to cover client and tool layers**
   - Add `internal/intervals` tests proving the exact path, method, payload/query, response decoding, non-2xx error wrapping, and response-body closure behavior.
   - Add table-driven tool tests for: success, whitespace trimming/rejection of empty IDs, date-mismatch warning, same-date no warning, idempotent re-link/already-linked handling, cancellation propagation, sanitized user-facing errors, no `confirm` schema property, and write-mode registration.

## Suggested revised plan shape

A reviewable Step 2 plan should list the key edits and decisions, for example:

- `internal/intervals/events.go` or a focused new intervals file: add typed `LinkActivityToEventParams` and `LinkActivityToEvent(ctx, params)` using the verified upstream endpoint and deliberate retry policy.
- `internal/tools/link_activity_to_event.go`: strict request validation, optional supporting reads for date comparison, typed response with `_meta.warnings`, short sanitized errors, `RequirementWrite`.
- `internal/tools/registry.go` and `internal/toolchecks/schema_stability.go`: register and snapshot the new write tool.
- Tests in `internal/intervals` and `internal/tools`, plus README/CHANGELOG/schema snapshot updates.

Once those choices are documented, the Step 2 plan should be ready for another review.
