# Plan Review — TP-003 Step 3: Add registry/test tool scaffolding

Verdict: **revise before proceeding**.

I reviewed `PROMPT.md`, `STATUS.md`, the current `internal/mcp` and `internal/tools` skeleton from Step 2, and the PRD MCP/response-shaping requirements. The broad direction is right: keep `internal/tools` SDK-free, adapt it in `internal/mcp`, and use a fake registry/tool for protocol tests. However, the Step 3 plan as recorded is still too thin on two contract points that are acceptance-critical: test-only scaffolding boundaries and sanitized user-facing errors.

## Blocking revisions

1. **Define the error contract before wiring tool handlers.**
   The plan must say how handler errors become MCP call results without leaking internals. A plain `error` returned by a handler and converted with `err.Error()` is not sufficient; it can expose wrapped upstream/API/config details to the LLM. Add an explicit public-error mechanism, for example a `tools.UserError`/`ToolError` type with a short public message plus optional internal wrapped cause, or a sanitizer that maps unknown errors to a generic actionable message while logging details server-side.

2. **Keep the noop/fake tool test-only.**
   Step 3 should not register a production placeholder/noop tool in the default server catalog. The fake/noop registry should live in tests or a clearly test-only helper so v0.1 does not expose unstable catalog entries. Production can continue with an empty registry until the real `get_athlete_profile` implementation lands in a later task.

3. **Specify the registry contract precisely enough for Step 4 protocol tests.**
   The plan should name the fields and invariants the registrar enforces: snake_case tool name, non-empty description, valid input schema, non-nil handler, and preferably duplicate-name rejection before calling the SDK. This avoids relying on SDK panics/recovery for normal validation failures and makes Step 4 assertions deterministic.

4. **Include response content shape in the contract.**
   The prompt asks Step 3 to define response content, not only registration metadata. The plan should state which content types are supported for v0.1 scaffolding (text is enough), how structured content is passed through, and what happens to unsupported content types. Silent dropping is risky unless documented and tested.

## Guidance for the revised plan

- Preserve the Step 1 boundary decision: `internal/tools` should not import `github.com/modelcontextprotocol/go-sdk` types.
- Use the fake registry/tool to exercise `initialize`, `tools/list`, and `tools/call` in Step 4 via SDK in-memory/IO transports; do not require Claude Desktop or intervals.icu.
- Make the fake tool name obviously test-scoped, snake_case, and not part of the public catalog, e.g. `test_echo` inside `internal/mcp` tests.
- Keep returned MCP tool errors short and actionable, while logging detailed internal errors with `slog` away from stdout.
- Update `STATUS.md` with the finalized Step 3 contract so later tool implementations know how to conform.

Once the plan addresses those points, Step 3 should be safe to implement without expanding scope into real intervals.icu tool behavior.
