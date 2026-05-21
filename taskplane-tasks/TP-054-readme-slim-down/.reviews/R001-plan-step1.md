# Plan Review R001 — Step 1: Pre-flight verification

## Verdict

Needs changes before Step 1 is considered adequately planned.

The Step 1 checklist in `STATUS.md` mirrors the prompt, and the dependency gates are directionally correct. However, the current plan is still too implicit for this cleanup task: it does not say how the "live website" requirement will be verified, and it does not pre-enumerate the one-row-per-deleted-section mapping that the acceptance criteria require.

## Findings

### 1. Live-site verification is not explicit

The prompt requires confirming that migrated content exists on the live site, and specifically that TP-051's generated catalog website page is live. The plan currently only has a generic destination table. If the implementer verifies only local `web/content/**` files or task `STATUS.md` files, they could delete the repo docs before the production site actually contains the replacement content.

Add an explicit check to Step 1 such as:

```bash
curl -fsSL https://icuvisor.app/reference/tools/ >/tmp/icuvisor-tools.html
curl -fsSL https://icuvisor.app/install/macos/ >/tmp/icuvisor-macos.html
# repeat for every destination URL recorded in the table
```

For each table row, record evidence that proves both availability and completeness, e.g. HTTP 200 plus page title/key headings or a short content grep. If network or deployment access is unavailable, record that as a blocker and hold the task instead of substituting local-only evidence.

### 2. The destination matrix should be enumerated before proceeding

The acceptance criteria require `STATUS.md` to include one row per deleted section/content item. The current table is empty and the plan does not define the expected row set. This is risky because several README sections split across multiple website destinations.

Before moving to Step 2, add rows for at least:

- README `Features (planned for v1.0)` → `https://icuvisor.app/`
- README `MCP tool catalog` → `https://icuvisor.app/reference/tools/`
- README `MCP resources` → `https://icuvisor.app/reference/resources-prompts/#resources`
- README `MCP prompts` → `https://icuvisor.app/reference/resources-prompts/#prompts`
- README macOS install/DMG prose and `docs/install/macos.md` → `https://icuvisor.app/install/macos/`
- README quickstart → `https://icuvisor.app/install/` and `https://icuvisor.app/guides/api-key/`
- README API-key/keychain prose → `https://icuvisor.app/guides/api-key/`
- README MCP transport prose → `https://icuvisor.app/guides/http-transport/` and `https://icuvisor.app/reference/cli/`
- README delete/write safety mode → `https://icuvisor.app/reference/safety-modes/` and `https://icuvisor.app/explain/safety-modes/`
- README toolset tiers → `https://icuvisor.app/reference/safety-modes/#toolset-tier` and `https://icuvisor.app/explain/safety-modes/`
- README post-upgrade/new-conversation line and `docs/post-update.md` → `https://icuvisor.app/guides/after-upgrade/`
- `docs/clients/claude-desktop.md` → `https://icuvisor.app/connect/claude-desktop/`
- `docs/clients/claude-code.md` → `https://icuvisor.app/connect/claude-code/`
- `docs/coach-mode.md` → `https://icuvisor.app/guides/coach-mode/` and `https://icuvisor.app/explain/coach-mode/`
- TP-053 dependency/tutorial evidence → `https://icuvisor.app/tutorials/getting-started-chatgpt/` plus, if relevant, `https://icuvisor.app/connect/chatgpt/`

Each row should include concrete evidence, not just a URL.

## Non-blocking notes

- I verified the dependency marker files currently exist for TP-051, TP-052, TP-053, and TP-055 in this worktree. Step 1 should still record that evidence in `STATUS.md` before any README/doc deletion work starts.
- Keep the Step 1 stop condition strict: if any dependency marker, status, live URL, or completeness check fails, update `STATUS.md` with the blocker and do not proceed to the link sweep/rewrite/delete steps.
