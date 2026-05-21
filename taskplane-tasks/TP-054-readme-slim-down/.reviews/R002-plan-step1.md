# Plan Review R002 — Step 1: Pre-flight verification

## Verdict

Approved.

The updated Step 1 plan now addresses the main gaps from R001: it requires live-site verification with `curl -fsSL` plus heading/content evidence, and it pre-enumerates the deleted-section-to-website destination matrix that the acceptance criteria require.

## Findings

No blocking findings for the plan.

## Execution notes

- For fragment URLs such as `#resources`, `#prompts`, and `#toolset-tier`, remember that `curl` fetches only the page URL. Record evidence that the built/live HTML contains the expected anchor or heading/content for that fragment, not just that the parent page returns 200.
- Record dependency evidence in `STATUS.md` before proceeding: TP-051, TP-052, TP-053, and TP-055 should each have current-tree evidence, not only a checked box. The dependency `STATUS.md` files currently indicate completion, but Step 1 should still paste the specific evidence used for this task.
- Keep the stop condition strict: any failed live URL, missing migrated content, or unresolved dependency should hold the task before the inbound link sweep or README rewrite begins.
