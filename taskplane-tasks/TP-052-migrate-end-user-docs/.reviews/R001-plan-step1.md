# Plan Review — Step 1: Inventory + audit

Verdict: **blocked / changes required before proceeding past preflight**

I read the TP-052 prompt and current `STATUS.md`. The Step 1 checklist captures the two prompt bullets at a high level, but the plan is missing one required dependency gate and should make the audit evidence more explicit before content migration starts.

## Findings

1. **Required TP-050 dependency is not covered by the Step 1 plan.**
   The prompt lists TP-050 as a blocking dependency, but `STATUS.md` only plans to verify TP-055. Current local taskplane evidence is inconsistent: `taskplane-tasks/TP-050-hugo-hextra-site-scaffold/STATUS.md` says `Status: ⏳ Open` / `Current Step: Not started`, and there is no `.DONE` file in that task directory, even though `web/` content exists in this worktree. Step 1 should explicitly verify and record TP-050 landing/scaffold evidence before relying on the Hugo site as the migration destination.

2. **TP-055 should be verified first and recorded before the audit proceeds.**
   TP-055 appears to have landed locally (`taskplane-tasks/TP-055-reconcile-doc-conflicts/.DONE` exists and TP-055 `STATUS.md` is complete), but TP-052 `STATUS.md` still says “not verified yet.” Because the prompt says to stop if TP-055 has not landed, update the plan/status so this is the first action, with concrete evidence for the three conflicts:
   - analyzers remain planned PRD/ROADMAP scope, not registered/generated current catalog;
   - `get_planning_parameters` remains only the upstream-gap deferral, not a registered/shipped tool;
   - `update_wellness` public docs/catalog surface the read-only `sleepScore` error contract matching code.

3. **The inventory/audit should require a recorded evidence matrix, not just “read every source file.”**
   The current wording is easy to execute inconsistently on a large docs migration. Add a small `STATUS.md` table for Step 1 with at least: source path, claims audited, code/source-of-truth checked, discrepancy found, and follow-up note. Include the prompt’s code SoT paths (`internal/config/`, `internal/safety/`, `internal/toolset/`, `internal/prompts/`) plus the CLI help source (`internal/app/testdata/help_golden.txt` or fresh `./bin/icuvisor --help`) when checking flags/env vars/exit codes.

## Recommended Step 1 adjustment

Before reading/adapting pages, add a dependency preflight subsection to `STATUS.md`:

| Dependency | Required evidence | Result |
| --- | --- | --- |
| TP-050 | scaffold landed / `web/` structure and Hugo config from the landed task, or explicit explanation for stale task status | pass/block |
| TP-055 | `.DONE`/complete status plus conflict-specific grep evidence | pass/block |
| TP-051 | `.DONE`/generated catalog present; note that `reference/tools.md` should not be hand-authored | pass/note |

If TP-050 or TP-055 cannot be verified, mark Step 1 blocked and stop. Once these gates and the audit table are added, the Step 1 plan is otherwise aligned with the prompt.
