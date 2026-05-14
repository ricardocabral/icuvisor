# R012 code review — Step 4: `icuvisor_list_advanced_capabilities`

Verdict: **REVISE**

## Findings

### 1. `go test ./...` fails because the adversarial static catalog was not updated

**Severity:** P1

The new discoverability tool is registered unconditionally as a read/core tool (`internal/tools/registry.go:259`), but `internal/safety/adversarial_test.go` still pins the old catalog in `v03ToolCatalog` (`internal/safety/adversarial_test.go:23-60`). As a result, the catalog-count assertion at `internal/safety/adversarial_test.go:104-105` now fails in every delete mode with one unexpected extra tool:

- `safe`: registered 31, want 30
- `none`: registered 21, want 20
- `full`: registered 38, want 37

Please add `icuvisor_list_advanced_capabilities` to that static catalog with `tools.RequirementRead` (or otherwise update the adversarial guardrail intentionally) so the whole repository test suite passes and the new tool is covered by the no-confirm/delete-mode adversarial checks.

## Verification

- `git diff 5c378a0989eacb00845aef08eab1c2d1e74c2fe5..HEAD --name-only` — reviewed
- `git diff 5c378a0989eacb00845aef08eab1c2d1e74c2fe5..HEAD` — reviewed
- `go test ./internal/tools ./internal/mcp` — pass
- `go test ./...` — **fail** in `internal/safety` as described above
