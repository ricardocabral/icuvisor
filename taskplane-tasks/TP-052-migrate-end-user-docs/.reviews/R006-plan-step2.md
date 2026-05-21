# Plan Review — Step 2: Reference section first

Verdict: **approved**

I read `PROMPT.md`, the current `STATUS.md`, and re-checked the relevant existing website state under `web/content/reference/`. The Step 2 plan now addresses the blockers from R004: Step 1 has been corrected/re-verified, and the existing duplicate `reference/toolset-tiers.md` page is explicitly included in the Step 2 checklist for reconciliation.

## What looks good

- `reference/cli.md` is planned against the actual fixture, `internal/app/testdata/help.golden`, with verbatim help output. This matches the Step 1 audit and avoids the stale prompt filename.
- `reference/safety-modes.md` now explicitly covers both `ICUVISOR_DELETE_MODE` and `ICUVISOR_TOOLSET`, registration effects, and `_meta.delete_mode` / `_meta.toolset` behavior.
- The plan includes R004’s required cleanup path for `web/content/reference/toolset-tiers.md`, so the site should end with one canonical safety/toolset reference instead of duplicate pages.
- `reference/config-file.md` is grounded in `internal/config/` and coach config code, which is the correct source of truth for accepted JSON fields and defaults.
- `reference/resources-prompts.md` points to `internal/resources/` and `internal/prompts/`, not the README, for the four resources and five prompts.
- `reference/tools.md` is correctly treated as generated/stubbed content; the plan avoids hand-authoring the tool catalog.

## Non-blocking implementation notes

- When authoring `config-file.md`, include the code-accepted legacy `api_key` field only as a discouraged/plaintext legacy compatibility field; the recommended path remains OS keychain/setup. Also include the `coach` object because the JSON decoder accepts it and Step 1 noted coach config as code truth.
- If `resources-prompts.md` lists tool names from prompt specs, link each tool name to the generated `/reference/tools/` anchor per the task’s cross-reference rule.
- If `toolset-tiers.md` is removed or replaced, make sure no section index, sidebar/frontmatter, or existing internal link still points at it before Step 8’s Hugo build/link check.

No plan changes are required before executing Step 2.
