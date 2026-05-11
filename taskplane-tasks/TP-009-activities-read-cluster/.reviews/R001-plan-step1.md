# Plan Review — TP-009 Step 1

Decision: **Changes requested**

## Summary

I cannot approve Step 1 yet because the plan has not actually been recorded. `STATUS.md` still has an empty `Step 1 Notes` section and all Step 1 checklist items are unchecked. The task prompt requires this step to identify endpoints, query parameters, response shapes, uncertainty, pagination contract, and `include_unnamed` filtering behavior before implementation begins.

## Blocking issues

1. **No endpoint/type plan is present**
   - Missing an endpoint table for all six tools (`get_activities`, details, intervals, streams, splits, messages): method, path, required path params, query params, expected status codes, and typed response shapes.
   - Missing notes on Strava-blocked markers and how they map into the required `unavailable: { reason: "strava_tos", workaround: ... }` shape.
   - Missing explicit uncertainty captured from the public docs / black-box validation plan.

2. **No pagination contract is defined**
   - Missing default `page_size` and max `page_size`.
   - Missing opaque `next_page_token` format and validation/error behavior.
   - Missing explanation of whether pagination is offset/date/id based and how it preserves deterministic ordering across pages.
   - Missing rationale that the default fits the PRD §7.2.D response-shaping/token-budget requirements.

3. **`include_unnamed` behavior is undecided**
   - The Step 1 requirement says to decide server-side vs client-side filtering and prefer server-side. `STATUS.md` does not record a decision or fallback.
   - The implementation needs to know whether the upstream activities endpoint supports an `include_unnamed`/equivalent query param, and if not, how client-side filtering interacts with pagination.

4. **Integration assumptions are not documented**
   - Missing how TP-007 response shaping and TP-008 stream/unit canonicalization will be applied at tool boundaries.
   - Missing response `_meta` shape, including `next_page_token`, `total_count` if available, `server_version`, unknown stream keys, and unit metadata.
   - Missing typed struct boundaries between `internal/intervals` upstream decode types and `internal/tools` shaped response types.

## Required updates before approval

Please update `STATUS.md` under `Step 1 Notes` with at least:

- A table of public intervals.icu endpoints for the six tools, including method/path/query params and known/unknown response fields.
- A typed data model outline for upstream activity list/detail/interval/stream/split/message payloads and the shaped MCP responses.
- The exact `get_activities` pagination contract: sort order, default/max page size, token contents before encoding, expiry/versioning if any, and error messages for invalid tokens.
- The `include_unnamed` decision, including fallback if upstream does not support server-side filtering and how pagination remains correct under fallback.
- Explicit uncertainties/blockers to validate with fixtures or black-box testing, especially Strava-blocked detection and upstream availability of split/message endpoints.

