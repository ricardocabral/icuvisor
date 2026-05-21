# Review: Step 1 plan

Verdict: **request one small plan adjustment before executing Step 1**.

The revised Step 1 plan is much better than R001: it now treats the prompt line numbers/conflict shape as stale, accounts for TP-051/TP-052/TP-054-era generated/reference surfaces, asks for file paths and current line numbers, and explicitly records which later remediation steps may be no-ops.

One important gap remains for Conflict B:

- The `get_planning_parameters` bullet currently checks ROADMAP and README/reference surfaces, but it does **not** explicitly verify the code/catalog registration truth. Conflict B's remediation depends on whether the tool is registered, so Step 1 should inspect the same source-of-truth path used for analyzer verification: `internal/tools/catalog.go` / `registryBaseTools`, `internal/tools/registry.go`, and generated catalog evidence such as `tools.Catalog()` output or `web/data/tools.json` if relevant. Documentation surfaces alone can only show whether the old contradiction remains; they cannot establish whether the remaining ROADMAP/README/reference statement matches the implemented catalog.

Suggested wording change:

> Verify current `get_planning_parameters` registration status using `internal/tools/catalog.go`, `internal/tools/registry.go`, and generated/catalog surfaces where relevant; then verify ROADMAP and README/reference surfaces, recording whether the prior contradiction still exists and whether the docs match the code.

With that adjustment, the Step 1 plan is appropriate. The existing bullets for analyzer re-verification and `update_wellness` error-contract verification are aligned with the task prompt and the current tree.
