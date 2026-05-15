# Review R010 — Code review for Step 3: Tests

Verdict: **REVISE**

## Findings

1. **Step 3 did not add or update any tests.**  
   `git diff 2784b93..HEAD --name-only` contains only `STATUS.md` and the prior review artifact; no `*_test.go` files changed. However `STATUS.md:37-39` marks all Step 3 test items complete. The task prompt's Step 3 acceptance requires concrete coverage, including actual response JSON verification for `_meta.schema_changed` and stabilized golden/default catalog hash behavior. Marking these done without test changes leaves the acceptance criteria unimplemented.

2. **The status log contradicts the committed review artifact.**  
   `taskplane-tasks/TP-040-schema-change-notification/.reviews/R009-plan-step3.md:3` says `Verdict: REVISE`, but `STATUS.md:68` records `Review R009 | plan Step 3: APPROVE`. That makes the task state unreliable and bypasses the required plan revisions listed in R009, especially the missing tool-result JSON assertion and `list_advanced_capabilities` regression test.

3. **Known response-path regression remains untested.**  
   R009 explicitly required a regression test for `icuvisor_list_advanced_capabilities` server-version/catalog metadata. The current handler still calls `encodeShaped(..., "", ...)` at `internal/tools/list_advanced_capabilities.go:98`, and the existing tests in `internal/tools/list_advanced_capabilities_test.go:14-101` assert `toolset` but never assert `_meta.server_version` or `catalog_hash` in the actual result JSON. A Step 3 test suite should fail on this until the handler is wired to the registry/server version path.

## Tests run

- `go test ./internal/response ./internal/tools ./internal/mcp`
