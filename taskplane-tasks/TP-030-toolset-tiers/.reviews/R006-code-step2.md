# R006 code review — Step 2: Per-tool tier membership

Verdict: **APPROVE**

No blocking findings.

Reviewed the diff from `9254b31b3f765a15b59d6e73b36225224e168df2..HEAD`, including the per-tool constructor tier annotations, `tools.Tool.Toolset`/`EffectiveToolset`, MCP validation for invalid in-code toolsets, and the catalog membership test.

Verification run:

- `go test ./...` — pass
- `git diff --check 9254b31b3f765a15b59d6e73b36225224e168df2..HEAD` — pass

Notes for the next steps:

- Step 3 should ensure tier filtering uses the same MCP registration validation path before applying `EffectiveToolset`, so unknown non-empty toolsets continue to fail closed rather than being treated as `full`.
- The current membership matrix correctly pins the existing registered catalog and preserves the Step 2 boundary; `icuvisor_list_advanced_capabilities` remains intentionally deferred to Step 4.
