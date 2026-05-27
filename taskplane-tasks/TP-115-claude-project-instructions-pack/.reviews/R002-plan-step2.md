# Plan Review: Step 2 — Add copy-paste Claude Project instructions

**Verdict: APPROVE**

No blocking issues. The Step 1 notes give enough direction for Step 2: create the new guide, include one base copy-paste block plus optional weekly review, recovery check, and race-week taper add-ons, and keep the content grounded in registered icuvisor prompts/tools.

Implementation guardrails to keep while writing the page:

- Add normal Hugo front matter (`title`, `description`, and likely `weight`) and use `text` fences for copy-paste instruction blocks.
- Keep secrets out of the blocks: explicitly forbid API keys, athlete IDs, local config file contents, paths, and private setup details in Project instructions.
- Anchor date/time behavior to the athlete-local timezone configured in icuvisor; do not require users to paste private config. If naming tools/prompts, use only verified names from the reference/data files.
- Make tool grounding explicit: cite the icuvisor tool/prompt behind key numbers, label subjective scales, call out missing/stale data, and do not invent metrics.
- For race-week taper, keep the add-on advisory-only: no calendar writes/deletes unless the user separately asks and the registered tools allow it.
- Avoid duplicating full cookbook recipes here; Step 3 can add links from existing recipes back to this reusable Project-instructions guide.
