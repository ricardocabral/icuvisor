# Code Review — Step 1: Prompt registration plumbing

Verdict: changes requested

## Findings

- [P2] `prompts/get` can leak arbitrary handler errors to the client. In `internal/mcp/prompts.go:45-47`, handler errors are returned as JSON-RPC errors after `publicPromptErrorMessage`; however `publicPromptErrorMessage` returns the raw error text for any message that does not contain `api key`, `token`, or `secret` (`internal/mcp/prompts.go:112-117`). Since `PromptRegistry` is an extension point, future prompt handlers can expose upstream paths, stack/context, or raw identifiers. This does not match the existing tool/resource sanitization pattern or the MCP convention that client-facing errors be short/actionable and free of internal detail. Prefer a prompt-specific public/user error type for expected argument errors, and return `genericPromptErrorMessage` for all other errors.

- [P3] The worktree includes unrelated task artifacts: `taskplane-tasks/TP-016-v02-dogfood-validation/STATUS.md` and `taskplane-tasks/TP-032-mcp-prompts/.reviewer-state.json`. Keep these out of the TP-032 step commit; the TP-016 status additions are unrelated to prompt registration and include truncated table text.

## Notes

- `git diff a328310d5c2dea6d4f879b7e9dcc1c8ee249df25..HEAD` was empty, so I reviewed the current worktree changes.
- `go test ./...` passes.
