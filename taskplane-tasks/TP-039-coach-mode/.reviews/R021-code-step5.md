# R021 code review — Step 5: Catalog-cache caveat + Tests

Verdict: **APPROVE**

## Findings

No blocking findings for Step 5.

The R020 gaps are addressed:

- `select_athlete.allowed_tools` now uses the MCP registrar's post-gate `visibleToolNamesForAthlete` helper via `coach.SelectionContext`, so the response and `_meta.requires_new_conversation` are based on the same catalog surface used by `tools/list`.
- The advanced-capabilities regression now parses structured JSON and asserts exact row names for the relevant delete-mode/toolset/coach-ACL combinations, including delete-mode-hidden and core-toolset-hidden cases.
- The fake-client end-to-end test now uses valid write/delete arguments and asserts the upstream request counter is unchanged after read-only athlete denials.
- `docs/coach-mode.md` documents the catalog-cache caveat and the new-conversation/reconnect guidance for `_meta.requires_new_conversation: true`.

## Tests run

- `go test ./internal/mcp ./internal/tools ./internal/coach`
- `go test ./...`
- `go test -race ./internal/mcp`
