# Plan Review — TP-003 Step 3: Add registry/test tool scaffolding

Verdict: **approve**.

I reviewed `PROMPT.md`, the revised `STATUS.md`, the current Step 2 MCP/tools skeleton, and the relevant PRD MCP/response-shaping sections. The revised Step 3 plan addresses the blockers from the previous review and is scoped appropriately for TP-003.

## Why this is ready

- The registry boundary remains SDK-free in `internal/tools`, with `internal/mcp` responsible for adapting tool definitions to `the MCP Go SDK`.
- The plan now defines registration invariants clearly enough for deterministic implementation and tests: snake_case names, non-empty descriptions, object schemas, non-nil handlers, duplicate-name rejection, and unsupported content rejection.
- The fake/noop tool is explicitly test-only, so the production v0.1 catalog will not expose unstable placeholder tools before `get_athlete_profile` is implemented.
- The sanitized error contract is now explicit: expected/public tool errors can return short user-facing messages, while unknown/internal errors are logged and mapped to a generic actionable tool error result with `isError=true`.
- The response content shape is defined for this scaffold: text content plus optional structured JSON object passthrough, with unsupported content types rejected rather than silently dropped.

## Non-blocking implementation notes

- Implement `UserError`/public errors with a separate public-message accessor or equivalent. Do not rely on `err.Error()` if it includes a wrapped internal cause.
- If `OutputSchema` is added to the internal `tools.Tool` contract, pass it through to the SDK and validate that it is an object when present.
- Track registered names inside the registrar so duplicates fail before relying on SDK replacement behavior.
- Be explicit in code/tests about call-time argument validation. If the adapter keeps handing raw JSON to handlers, document that handlers own argument unmarshalling/validation and ensure malformed handler errors still pass through the sanitizer.

Proceed with Step 3 implementation; no real intervals.icu tool behavior, Streamable HTTP, SSE, or production placeholder catalog entries are needed for this step.
