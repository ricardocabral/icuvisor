# Task: TP-108 - Stale conversation troubleshooting docs

**Created:** 2026-05-26
**Size:** S

## Review Level: 0

**Assessment:** Documentation-only troubleshooting guidance with no runtime behavior changes.
**Score:** 1/8 — Blast radius: 0, Pattern novelty: 0, Security: 0, Reversibility: 1

## Canonical Task Folder

```
taskplane-tasks/TP-108-stale-conversation-troubleshooting-docs/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Document stale conversation and cached tool-catalog behavior so users know when to start a new chat, refresh/reconnect MCP tools, restart clients, or run diagnostics. This reduces confusing reports where old conversation state masks a fixed server, changed config, or updated tool schema.

Tracking issue: https://github.com/ricardocabral/icuvisor/issues/35

## Dependencies

- **Task:** TP-035 (CLI help/documentation foundation exists)
- **Task:** TP-040 (schema-change notification behavior exists or is scoped)

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- `CLAUDE.md` — repository rules and documentation conventions.
- `README.md` — public entry point and docs links.
- `docs/` — existing client setup/troubleshooting content.
- `ROADMAP.md` — post-update schema-change notification context.

## Environment

- **Workspace:** repository root (`/Users/jusbrasil/prj/icuvisor`)
- **Services required:** None.

## File Scope

- `README.md`
- `docs/**/*`
- `web/**/*`
- `internal/app/**/*diagnostics*`
- `internal/app/**/*help*`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Existing user docs structure identified

### Step 1: Add troubleshooting guidance

- [ ] Create or update a discoverable troubleshooting section for stale conversations and cached tool catalogs.
- [ ] Explain in plain language why stale state happens.
- [ ] Include safe first steps: start a new conversation, refresh/reconnect tools, verify version, run `icuvisor diagnostics`.
- [ ] Mention timezone/date, zones, tool visibility, and write failures as example symptoms.
- [ ] Reinforce that API keys must not be pasted into assistant conversations.

**Artifacts:**
- `docs/**/*` or `web/**/*` docs file (modified/new)
- `README.md` (modified if needed for discoverability)

### Step 2: Link from related docs and update changelog

- [ ] Add links from setup/client docs or README where users are likely to look.
- [ ] Update any post-update/onboarding copy that already mentions schema changes to point at the troubleshooting guidance.
- [ ] Update `CHANGELOG.md` under `[Unreleased]` for docs addition if project convention requires it.
- [ ] Run markdown/link/docs checks if available.

**Artifacts:**
- Related docs (modified)
- `CHANGELOG.md` (modified if needed)

### Step 3: Testing & Verification

- [ ] Run markdown/link checks if available
- [ ] Run docs generation if the docs site is generated from repository files
- [ ] Run FULL test suite if generated docs or embedded help strings are touched: `make test`
- [ ] Build passes if generated assets or app strings are touched: `make build`
- [ ] Fix all failures or document pre-existing unrelated failures in STATUS.md

### Step 4: Documentation & Delivery

- [ ] "Must Update" docs modified
- [ ] "Check If Affected" docs reviewed
- [ ] `STATUS.md` discoveries and final notes updated
- [ ] Commit at step boundary with the task ID in the message

## Documentation Requirements

**Must Update:**
- User-facing troubleshooting docs — add stale conversation / cached catalog guidance.
- `STATUS.md` — keep execution state current.

**Check If Affected:**
- `README.md` — add a concise pointer if needed.
- Existing client setup docs under `docs/` or `web/` — link to troubleshooting where relevant.
- `CHANGELOG.md` — update if documentation changes are tracked there.

## Completion Criteria

- Troubleshooting guidance is discoverable from the main user docs.
- Guidance explicitly recommends a new conversation after schema changes or suspicious stale behavior.
- Guidance explains refresh/reconnect at a client-neutral level.
- Guidance covers timezone/date, zones, tool visibility, and write-failure stale-state examples.
- Guidance preserves credential safety: never paste API keys into assistant conversations.

## Git Commit Convention

Commits happen at **step boundaries**. All commits for this task MUST include `TP-108` for traceability. Examples:

- `docs(TP-108): complete step 1 — add stale conversation troubleshooting`
- `hydrate: TP-108 expand step checkboxes`

## Do NOT

- Do not change runtime behavior unless a docs link requires a small help-string update.
- Do not include real API keys, athlete IDs, or private examples.
- Do not overfit instructions to one MCP client when client-neutral guidance is sufficient.
- Do not modify protected docs without explicit approval.

---

## Amendments

_Add amendments below this line only._
