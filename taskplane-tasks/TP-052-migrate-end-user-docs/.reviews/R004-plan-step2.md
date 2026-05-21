# Plan Review — Step 2: Reference section first

Verdict: **blocked / changes required before executing Step 2**

I read `PROMPT.md`, the current `STATUS.md`, and checked the current `web/content` tree. The Step 2 checklist is directionally aligned with the requested reference pages, but the plan should not proceed while the previous Step 1 review blocker remains unresolved.

## Findings

1. **Step 1 is still unresolved after a changes-requested code review.**  
   `R003-code-step1.md` requested changes because `STATUS.md` records TP-050 as passing even though the claimed scaffold destination structure is not present. The current tree still only has `web/content/_index.md` and `web/content/reference/`; there is no TP-050 `.DONE`. Since TP-050 is a blocking dependency in the prompt, and `STATUS.md` has not been corrected after R003, Step 2 should not be marked in progress or executed yet. First resolve Step 1 by either creating/verifying the missing scaffold under TP-050 or updating `STATUS.md` to accurately reflect the dependency state and agreed scope.

2. **The Step 2 plan should explicitly reconcile the existing `reference/toolset-tiers.md` page.**  
   The prompt says README “Toolset tiers” content should be merged into `reference/safety-modes.md` and the paired explanation page, “one page per concept, not one per env var.” The current site already contains `web/content/reference/toolset-tiers.md`. If Step 2 adds `reference/safety-modes.md` without addressing that page, the reference section may have duplicate or drift-prone toolset-tier documentation. Add an explicit Step 2 action to fold any useful content from `toolset-tiers.md` into `safety-modes.md` (or otherwise remove/replace the duplicate website page, if allowed by the task boundaries).

## What looks good once the blocker is fixed

- Using `internal/app/testdata/help.golden` for `reference/cli.md` matches the Step 1 audit finding and avoids the stale prompt filename.
- The planned safety-mode page includes the important code-truth details from `internal/safety`: `safe/full/none`, `core/full`, registration effects, and `_meta` echoes.
- The plan correctly avoids hand-authoring the tool catalog and preserves the existing generated `reference/tools.md` shortcode page.
- The resources/prompts plan points to `internal/resources/` and `internal/prompts/`, which is the right source of truth for the four resources and five prompts.

## Recommended adjustment

Before Step 2 execution, update `STATUS.md` to resolve R003 and move Step 2 back out of “In Progress” until the dependency gate is accurate. Then add a Step 2 checklist item for the existing `reference/toolset-tiers.md` so the new safety/toolset reference has a single canonical home.
