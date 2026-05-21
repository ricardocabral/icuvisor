# R009 Plan Review — Step 4

Verdict: REVISE

## Findings

1. **The fallback smoke does not satisfy the Step 4 install/tool-call requirement.**
   The plan says that if Claude Desktop GUI automation cannot confirm the install, the worker will "record the limitation" and run a packaged `server/icuvisor` MCP `tools/list` smoke with dummy env. That is useful evidence, but it is not equivalent to "Test local installation in Claude Desktop by dragging/opening the `.mcpb` and confirming stdio tool call works" from the task. If Claude Desktop cannot be driven in this worker, the plan should explicitly leave that checkbox incomplete or record a blocker/manual-validation requirement rather than treating `tools/list` as completion.

2. **Use an actual MCP tool invocation for the non-GUI smoke, not only `tools/list`.**
   `tools/list` verifies catalog discovery, but the task asks for a stdio tool call. The fallback should initialize the packaged binary and invoke a no-network tool such as `icuvisor_list_advanced_capabilities` with `{}` using dummy/non-secret env. This better validates request/response handling without requiring a real intervals.icu API key or network access.

## Recommended plan adjustment

- Attempt the real Claude Desktop open/install path and document exact evidence in `STATUS.md` (bundle path, Claude version if available, whether extension config appeared, and result of a simple tool call).
- If GUI confirmation is not possible, document that as a limitation/blocker for the install checkbox and run a separate MCP stdio smoke against the staged/bundled binary that performs both `tools/list` and `tools/call` for `icuvisor_list_advanced_capabilities`.
- Keep the proposed docs/changelog updates, ensuring the public docs clearly state the current first supported release artifact slice is macOS `.mcpb`, the API key is entered only in Claude Desktop's sensitive extension config, and the manual JSON/keychain setup remains as fallback.
