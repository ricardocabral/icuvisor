# Plan Review R005 — Step 2: Inbound link sweep

## Verdict

Approved.

The revised Step 2 plan addresses the prior review: it scopes historical task artifacts out of the product-link sweep, enumerates the current grep hits, and assigns a context-sensitive action for each live product file.

## Review notes

- I re-ran the scoped command:
  ```bash
  git grep -nE 'docs/(install/macos|clients/claude-desktop|clients/claude-code|coach-mode|post-update)\.md' -- . ':!taskplane-tasks/**'
  ```
  The output matches the hit/action table in `STATUS.md`.
- The planned replacements are appropriate:
  - `SECURITY.md` should remove the stale maintainer-checklist pointer rather than link to the public install page.
  - `docs/dogfood/v0.2-findings.md` and `internal/app/setup.go` should point to `https://icuvisor.app/connect/claude-desktop/`.
  - `docs/internal-beta/onboarding-playbook.md` should point to the coach-mode website guide/explanation.
  - `internal/app/setup_test.go` should be updated with the setup output change.
  - README and files scheduled for deletion are correctly deferred to later steps.
- The explicit `taskplane-tasks/**` exclusion is reasonable because those files are task/review records, not live product inbound links. Keep that rationale in `STATUS.md` when recording final grep results.

## Implementation reminder

After editing the non-deleted files in this step, re-run the same scoped grep and record either zero remaining non-deferred product hits or the expected deferred README/deleted-doc hits for Steps 3 and 4.
