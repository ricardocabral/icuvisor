# Code Review — Step 3: Lift coach concerns to `internal/coach`

Decision: **Approved**

No blocking findings.

## What I checked

- Reviewed the full diff from `20753ac..HEAD` and the changed files.
- Re-read the task prompt and Step 3 status/design notes.
- Verified the new `coach.ToolFilter` preserves the prior coach visibility semantics:
  - coach catalog tools stay always visible;
  - non-athlete-scoped tools remain visible;
  - athlete-scoped tools delegate ACL decisions to `coach.Evaluator`;
  - selected/default/supplied athlete target resolution still normalizes and roster-checks before authorization.
- Checked the moved coach-filtered advanced-capabilities handler path in `internal/tools/list_advanced_capabilities.go`; the R008 wire-compatibility concerns are addressed (`_meta.toolset` is active-toolset-derived, no invented `_meta.delete_mode`, and the legacy coach enable-instruction wording is preserved).
- Confirmed `internal/mcp/registrar_tools.go` now adapts SDK/session/schema/raw arguments while delegating coach policy decisions through `coach.ToolFilter`.

## Tests run

```sh
go test ./internal/mcp ./internal/tools ./internal/coach ./internal/toolcatalog
go test ./...
```

## Non-blocking note

`CHANGELOG.md` still does not mention TP-064. The earlier Step 2 review treated this as a finalization item rather than a per-step blocker; it should be updated before the task is closed.
