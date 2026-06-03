# Plan Review — Step 1

Result: **request changes**

The Step 1 checklist is directionally correct, and the targeted tests currently pass:

```sh
go test ./internal/coach ./internal/config ./internal/tools
```

However, the audit plan is too narrow for the task’s stated “coach/local athlete routing” and “tool registration paths” scope.

## Required plan adjustments

1. **Include the MCP registration/routing layer in the audit.** The key public behavior is in `internal/mcp/registrar_tools.go` and `internal/mcp/schema.go`: athlete_id schema injection/stripping, selected-athlete fallback, `intervals.WithTargetAthleteID`, and conversion to public tool errors. Auditing only `internal/coach`, config normalization, `list_athletes`, and `select_athlete` can miss where unauthorized targets are collapsed into generic/public errors.

2. **Explicitly audit local-mode behavior.** Step 1 says coach/local, but the checklist is mostly coach-mode. The audit should identify expected behavior when coach mode is off: whether a supplied `athlete_id` is rejected before upstream calls, whether no model-supplied API-key path exists, and how local-athlete fallback is enforced.

3. **Make the audit deliverable concrete.** Since Step 1 is audit-only, add/update a `STATUS.md` discovery table (or equivalent note) mapping: entrypoint, current behavior, ambiguous/generic failure mode, expected public message, and follow-up test/implementation target for Step 2.

4. **Define message classes, not just one generic message.** The plan should distinguish at least invalid format, not-in-configured-roster/unauthorized athlete, and tool-not-allowed-for-selected-athlete. Public messages must remain short and avoid credentials/raw sensitive identifiers, but they should be actionable enough to satisfy the task mission.

5. **Consider running or at least inspecting MCP protocol tests during the audit.** Existing MCP tests already cover `tools/list`, `select_athlete`, visibility filtering, and public error text. Even if Step 2 owns hardening, Step 1 should include `internal/mcp` in the audit evidence; otherwise Step 2 may miss the actual client-facing surface.

