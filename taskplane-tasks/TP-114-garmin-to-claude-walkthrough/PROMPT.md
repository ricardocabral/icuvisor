# Task: TP-114 - Garmin to Claude walkthrough

**Created:** 2026-05-27
**Size:** M

## Review Level: 1

**Assessment:** Creates a new user-facing tutorial with possible mock visual assets and navigation links; no runtime changes, but requires plan review to keep claims, privacy, and docs structure coherent.
**Score:** 2/8 — Blast radius: 1, Pattern novelty: 1, Security: 0, Reversibility: 0

## Canonical Task Folder

```
taskplane-tasks/TP-114-garmin-to-claude-walkthrough/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Create a polished, copy-pasteable walkthrough that shows the core user journey: Garmin (or another device provider) syncs to intervals.icu, icuvisor exposes that data locally through MCP, and Claude can answer grounded training questions. The deliverable should be useful even without private screenshots: use redacted/mock visuals or diagrams, include safe prompts, and make clear that icuvisor reads intervals.icu rather than connecting directly to Garmin.

## Dependencies

- **Task:** TP-113 (homepage/local-first messaging should be refreshed first if both tasks touch landing-page calls to action)

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- `CLAUDE.md` — repository rules and documentation conventions.
- `web/README.md` — website build and content conventions.
- `web/content/tutorials/_index.md` — tutorial section conventions.
- `web/content/connect/claude-desktop.md` — Claude setup path.
- `web/content/cookbook/prompt-library.md` — existing prompt style.
- `web/content/explain/interval-sources.md` — source/import caveats.

## Environment

- **Workspace:** repository root (`/Users/jusbrasil/prj/icuvisor`)
- **Services required:** None.

## File Scope

- `web/content/tutorials/**/*`
- `web/content/connect/claude-desktop.md`
- `web/content/connect/claude-code.md`
- `web/content/cookbook/prompt-library.md`
- `web/static/images/**/*`
- `web/assets/**/*`
- `web/layouts/index.html`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Existing tutorial, connect, and cookbook structure reviewed

### Step 1: Plan the tutorial and visual treatment

- [ ] Decide whether to use redacted mock screenshots, lightweight diagrams, or existing UI-free documentation patterns.
- [ ] Define the tutorial path: device sync to intervals.icu, install/setup icuvisor, connect Claude, ask first grounded questions.
- [ ] Identify privacy guardrails for visuals and prompts: no real athlete IDs, API keys, raw training screenshots, or identifiable values.

**Plan-review checkpoint**

**Artifacts:**
- `STATUS.md` discoveries/plan notes (modified)

### Step 2: Add the walkthrough content

- [ ] Create a new tutorial page for the Garmin → intervals.icu → icuvisor → Claude journey.
- [ ] Include copy-paste prompts for first call, weekly review, recovery check, and troubleshooting if data is missing.
- [ ] Explain source limitations plainly: icuvisor reads intervals.icu, device-provider data must already be synced there, and some imported fields may be unavailable.
- [ ] Add redacted/mock visuals or diagrams if they improve comprehension without using private data.

**Artifacts:**
- New `web/content/tutorials/*.md` page
- Optional `web/static/images/**/*` or `web/assets/**/*` mock/redacted assets

### Step 3: Link the walkthrough from discovery surfaces

- [ ] Link the tutorial from the tutorial index and relevant Claude connection docs.
- [ ] Add a small homepage or cookbook pointer only if it improves discoverability without duplicating content.
- [ ] Update `CHANGELOG.md` under `[Unreleased]` for the new tutorial if appropriate.

**Artifacts:**
- `web/content/tutorials/_index.md` (modified if needed)
- `web/content/connect/*.md` (modified if needed)
- `web/layouts/index.html` or cookbook pages (modified if needed)
- `CHANGELOG.md` (modified if needed)

### Step 4: Testing & Verification

- [ ] Run docs/site build: `make web-build`
- [ ] Check rendered links and image references in the generated site output or via Hugo warnings.
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
- `web/content/tutorials/` — add the new walkthrough.
- `STATUS.md` — record the chosen visual/content approach.

**Check If Affected:**
- `web/content/connect/claude-desktop.md` and `web/content/connect/claude-code.md` — link to the walkthrough if relevant.
- `web/content/cookbook/prompt-library.md` — reuse or link prompts rather than duplicating too much.
- `web/layouts/index.html` — add a pointer only if it keeps the homepage focused.
- `CHANGELOG.md` — add docs note if appropriate.

## Completion Criteria

- A new tutorial explains the Garmin/device-provider → intervals.icu → icuvisor → Claude flow.
- Tutorial includes safe copy-paste prompts and missing-data caveats.
- No private data, real athlete IDs, API keys, or sensitive screenshots are introduced.
- Website build passes or any pre-existing failures are documented.

## Git Commit Convention

Commits happen at **step boundaries**. All commits for this task MUST include `TP-114` for traceability. Examples:

- `docs(TP-114): complete step 2 — add garmin to claude walkthrough`
- `hydrate: TP-114 expand step checkboxes`

## Do NOT

- Do not imply icuvisor connects directly to Garmin or any device provider.
- Do not include real training screenshots, athlete IDs, API keys, access tokens, or identifiable values.
- Do not mention competitor projects, forum threads, private research notes, or issue/PR numbers.
- Do not change runtime behavior.
- Do not modify protected docs without explicit approval.

---

## Amendments

_Add amendments below this line only._
