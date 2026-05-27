# Task: TP-115 - Claude Project instructions pack

**Created:** 2026-05-27
**Size:** M

## Review Level: 1

**Assessment:** Adds reusable prompt/instruction documentation across cookbook and client docs; plan review is useful to avoid duplicating existing prompt library content and to keep instructions client-safe.
**Score:** 2/8 — Blast radius: 1, Pattern novelty: 1, Security: 0, Reversibility: 0

## Canonical Task Folder

```
taskplane-tasks/TP-115-claude-project-instructions-pack/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Publish ready-made Claude Project instructions that help athletes get reliable icuvisor answers across chats. The pack should include timezone/date discipline, weekly review behavior, recovery-check behavior, and race-week taper behavior. It should be copy-pasteable, grounded in available icuvisor tools/prompts, and clear that users must not paste API keys or private config into project instructions.

## Dependencies

- **Task:** TP-106 (weekly review MCP prompt exists or its intended behavior is documented)

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- `CLAUDE.md` — repository rules and documentation conventions.
- `web/README.md` — website build and content conventions.
- `web/content/connect/claude-desktop.md` — Claude Desktop setup and project context.
- `web/content/connect/claude-code.md` — Claude Code setup if relevant.
- `web/content/cookbook/prompt-library.md` — current single-message prompt style.
- `web/content/cookbook/weekly-review.md` — existing weekly review recipe.
- `web/content/cookbook/readiness-check.md` — existing recovery/readiness recipe.
- `web/content/cookbook/race-week-taper.md` — existing taper recipe.
- `web/content/reference/resources-prompts.md` — registered MCP prompts reference.

## Environment

- **Workspace:** repository root (`/Users/jusbrasil/prj/icuvisor`)
- **Services required:** None.

## File Scope

- `web/content/guides/claude-project-instructions.md`
- `web/content/guides/_index.md`
- `web/content/connect/claude-desktop.md`
- `web/content/connect/claude-code.md`
- `web/content/cookbook/prompt-library.md`
- `web/content/cookbook/weekly-review.md`
- `web/content/cookbook/readiness-check.md`
- `web/content/cookbook/race-week-taper.md`
- `web/content/reference/resources-prompts.md`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Existing Claude, cookbook, and MCP prompt docs reviewed for overlap

### Step 1: Design the instruction pack structure

- [ ] Decide whether the pack belongs in a new guide, cookbook page, or both via links.
- [ ] Define the reusable instruction blocks: base timezone/date discipline, weekly review, recovery check, race-week taper, and missing/stale data handling.
- [ ] Identify exact tool/prompt names to mention only where they are registered or documented.

**Plan-review checkpoint**

**Artifacts:**
- `STATUS.md` discoveries/plan notes (modified)

### Step 2: Add copy-paste Claude Project instructions

- [ ] Create a user-facing page with a complete base Project instruction block.
- [ ] Add optional add-on blocks for weekly review, recovery check, and race-week taper.
- [ ] Include usage guidance: where to paste instructions, when to start a new chat, and how to keep API keys/config out of the project.
- [ ] Include tool-grounding requirements: use athlete-local timezone, cite source tools, call out missing/stale data, and do not invent metrics.

**Artifacts:**
- `web/content/guides/claude-project-instructions.md` (new)

### Step 3: Link and deduplicate with existing recipes

- [ ] Link the new guide from Claude connection docs and the guides index.
- [ ] Add pointers from weekly review, readiness check, race-week taper, or prompt-library docs where the Project instructions reduce repeated setup.
- [ ] Update `web/content/reference/resources-prompts.md` only if it needs to explain the relationship between MCP Prompts and client Project instructions.
- [ ] Update `CHANGELOG.md` under `[Unreleased]` if appropriate.

**Artifacts:**
- `web/content/guides/_index.md` (modified if needed)
- `web/content/connect/*.md` (modified if needed)
- Cookbook/reference docs (modified if needed)
- `CHANGELOG.md` (modified if needed)

### Step 4: Testing & Verification

- [ ] Run docs/site build: `make web-build`
- [ ] Check internal links and page placement in the rendered docs navigation.
- [ ] Run FULL test suite if non-doc/generated app files are touched: `make test`
- [ ] Build passes if app strings or generated assets are touched: `make build`
- [ ] Fix all failures or document pre-existing unrelated failures in STATUS.md

### Step 5: Documentation & Delivery

- [ ] "Must Update" docs modified
- [ ] "Check If Affected" docs reviewed
- [ ] `STATUS.md` discoveries and final notes updated
- [ ] Commit at step boundary with the task ID in the message

## Documentation Requirements

**Must Update:**
- `web/content/guides/claude-project-instructions.md` — publish the instruction pack.
- `web/content/guides/_index.md` or docs navigation surfaces — make the new guide discoverable.

**Check If Affected:**
- `web/content/connect/claude-desktop.md` and `web/content/connect/claude-code.md` — link where appropriate.
- `web/content/cookbook/*.md` relevant recipes — link rather than duplicating instructions.
- `web/content/reference/resources-prompts.md` — clarify Project instructions versus MCP Prompts if needed.
- `CHANGELOG.md` — add docs note if appropriate.

## Completion Criteria

- Users can copy one base Claude Project instruction block and optional specialized blocks.
- Instructions explicitly enforce athlete-local timezone/date discipline and missing-data caveats.
- Instructions cover weekly review, recovery check, and race-week taper use cases.
- API-key and private-config safety warnings are included.
- Website build passes or any pre-existing failures are documented.

## Git Commit Convention

Commits happen at **step boundaries**. All commits for this task MUST include `TP-115` for traceability. Examples:

- `docs(TP-115): complete step 2 — add claude project instructions pack`
- `hydrate: TP-115 expand step checkboxes`

## Do NOT

- Do not instruct users to paste API keys, athlete IDs, local config files, or secrets into Claude Project instructions.
- Do not mention competitor projects, forum threads, private research notes, or issue/PR numbers.
- Do not invent tool names; only mention registered tools/prompts or link to generic behavior.
- Do not change runtime behavior.
- Do not modify protected docs without explicit approval.

---

## Amendments

_Add amendments below this line only._
