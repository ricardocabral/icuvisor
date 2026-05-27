# Plan Review: Step 2 — Implement opt-in first-tool-call runner

**Verdict: Needs revision**

The hydrated Step 2 checklist points in the right direction, but it is still too high-level to approve. The riskiest parts of this task are exactly where the plan is vague: how the runner obtains the active tool schemas without handlers, how provider credentials are gated, and how the first-tool response is parsed and reported.

## Required clarifications before implementation

1. **Catalog-loading source**
   - Specify the exact mechanism for loading the four fixture modes (`core_safe`, `core_full_delete`, `full_safe`, `full_full_delete`).
   - Do not use only `toolcatalog.AllToolNames()` or `tools.Catalog()` for the runner; the model must receive registered tool definitions with descriptions/input schemas and with safety/toolset filtering applied.
   - Avoid requiring real intervals credentials or executing handlers. If adding an exported collector around the existing registry/safe registrar path, make that explicit and test that delete/toolset filtering matches expectations.

2. **Provider configuration and credential handling**
   - Name the required environment variables before coding, including provider/model/API key and any optional base URL.
   - Define behavior when they are absent. Recommended: skip with a clear message and zero exit only for an explicit “not configured” state; mismatches/errors from a configured provider should be non-zero.
   - Ensure API keys are read only from env, never echoed in logs/output, and never written into STATUS/results.

3. **Anthropic-compatible request/response contract**
   - Define how icuvisor tool definitions map to the provider’s tool schema.
   - Define extraction rules: first `tool_use` block wins; no `tool_use` is an explicit no-tool result; unknown returned tool is a failed case; malformed/provider errors are reported separately.
   - Include sensible request bounds (`max_tokens`, deterministic temperature if supported) so the smoke eval is repeatable.

4. **Output and exit semantics**
   - Specify the output format: per-case id, catalog mode, expected, actual/no-tool, pass/fail detail, and summary counts.
   - Specify exit codes for all-pass, assertion failures, provider misconfiguration, and provider/runtime errors.

5. **Network-free normal tests**
   - Add tests for catalog-mode selection, provider request construction, response parsing, no-tool handling, unknown-tool handling, and summary/exit classification using a fake HTTP client/RoundTripper.
   - Make sure `make test` cannot accidentally call the provider even if local provider env vars are set.

Once these details are reflected in the Step 2 plan, the implementation should be low-risk and aligned with the task’s opt-in/no-tool-execution requirements.
