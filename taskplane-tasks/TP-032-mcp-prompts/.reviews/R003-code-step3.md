# Code Review — Step 3: Token discipline

Verdict: changes requested

## Findings

- [P2] Weekly planning prompt requires a full-only tool without surfacing the core discovery path. `internal/prompts/catalog.go:84-88` tells the LLM to use `get_training_plan`, but that tool is full-only (`internal/tools/catalog_tiers_test.go:48-52`) while icuvisor defaults to the core toolset and PRD §7.2.E says `icuvisor_list_advanced_capabilities` is the core path for prompts that need advanced tools (`docs/prd/PRD-icuvisor.md:342-346`). In the default app, a user selecting this prompt will get instructions to call a tool that is not registered/visible. Keep the prompt terse, but add an explicit fallback/discovery cue (for example include `icuvisor_list_advanced_capabilities` in the tool list or a one-line guardrail to use it when `get_training_plan` is unavailable) and lock the wording in the golden file.

- [P2] `race_week_taper` does not enforce its required `race_date` server-side. The prompt advertises `race_date` as required at `internal/prompts/catalog.go:102-108`, but the SDK does not validate prompt arguments before calling the handler, and `staticPromptHandler` will render a prompt with no race anchor. Clients that ignore prompt argument metadata can therefore retrieve an unusable taper prompt instead of a short actionable error. Add a small validating handler (like `coachRosterTriageHandler`) that returns a `UserError` when `race_date` is missing, and add a test.

- [P3] Unrelated TP-016 status edits are still present in this worktree. `taskplane-tasks/TP-016-v02-dogfood-validation/STATUS.md:145-149` is outside TP-032 and includes a truncated table entry. Please keep it out of the TP-032 step commit.

## Notes

- Per instructions, `git diff a328310d5c2dea6d4f879b7e9dcc1c8ee249df25..HEAD --name-only` and the full committed diff were empty, so I reviewed the current uncommitted worktree changes.
- `go test ./internal/prompts ./internal/mcp ./internal/app` currently hits the known `internal/mcp` Streamable HTTP shutdown flake (`TestProtocolTransportParity`); `internal/prompts` itself passed.
