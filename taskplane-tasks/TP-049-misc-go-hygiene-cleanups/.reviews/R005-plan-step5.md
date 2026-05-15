# Plan Review: Step 5 — Fix registry error message

Verdict: Approved with implementation clarifications.

The step is correctly scoped to item 5, but the current worktree still has the hard-coded `getAthleteProfileName` references in `internal/tools/registry.go:67-70`, so this should not be marked as a TP-042 no-op unless the branch changes before implementation.

Implementation notes to keep the fix precise:

- The two current failures at lines 67-70 are registry precondition failures (`missing profile client`, `missing registrar`), not actual `get_athlete_profile` registration failures. Do not try to invent a per-tool name for those cases; use a registry-level subject such as `registering tools: missing profile client` and `registering tools: missing registrar`.
- If you also decide to wrap downstream `AddTool` errors with the actual tool name, keep it centralized and mechanical, for example a local `addTool(tool Tool) error` helper that returns `fmt.Errorf("registering %s: %w", tool.Name, err)`. However, that is broader than the cited lines; prefer the smallest change that satisfies the acceptance criterion unless you find existing tests/behavior require per-tool wrapping.
- Preserve error wrapping with `%w` for any existing underlying errors. The nil-precondition messages have no underlying error and should remain plain errors.
- Do not change tool names, schemas, registration ordering, capability/toolset gating, or `catalogCollectingRegistrar` behavior.
- A tiny regression test is worthwhile if not too much churn: cover `NewRegistryWithOptions(nil, ...).Register(...)` and/or `Register(ctx, nil)` and assert the message no longer mentions `get_athlete_profile`.
- Run the requested `make build`, `make test`, and `make lint` before committing. Step 6 will cover the full `test-race` sweep.

No blockers found.
