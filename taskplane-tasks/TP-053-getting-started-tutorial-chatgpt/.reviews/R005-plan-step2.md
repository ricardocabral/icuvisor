# Plan review — TP-053 Step 2

Decision: **approved** for drafting.

The current Step 2 plan addresses the blockers from the previous reviews. It now plans the missing Tutorials section/nav work, carries forward the local-stdio honesty constraints from Step 1, chooses a deterministic build-from-source install path, adds the temporary Go/Xcode/Git prerequisites, requires smoke-testing the exact `/Users/Shared/icuvisor/bin/icuvisor` path before using it in the page, preserves the required fenced `text` connector JSON, and includes the changelog update.

## Non-blocking watchout

- Keep the troubleshooting link out of **Where to next**. The plan currently says troubleshooting should appear only in the "footer/Where next area"; the prompt requires the **Where to next** section to contain exactly three links: `/connect/`, `/reference/tools/`, and `/guides/coach-mode/`. If you include troubleshooting at all, put it in a separate footer-style line outside the tutorial body, not as a fourth next-step link.

With that wording tightened during implementation, the plan is ready for Step 2.
