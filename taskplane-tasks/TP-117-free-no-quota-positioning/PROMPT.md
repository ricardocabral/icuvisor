# Task: TP-117 - Free and no-quota positioning

**Created:** 2026-05-27
**Size:** S

## Review Level: 0

**Assessment:** Documentation and marketing copy refresh clarifying existing open-source/free positioning; no runtime behavior changes.
**Score:** 1/8 — Blast radius: 1, Pattern novelty: 0, Security: 0, Reversibility: 0

## Canonical Task Folder

```
taskplane-tasks/TP-117-free-no-quota-positioning/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Make icuvisor’s free/no-quota value proposition clearer: the icuvisor binary is open source and free to use, with no icuvisor account, onboarding credits, SaaS quota, or subscription gate. Keep the message precise by distinguishing icuvisor’s lack of quotas from any usage limits imposed by the user’s chosen AI client, model provider, GitHub, package manager, or intervals.icu account.

## Dependencies

- **Task:** TP-113 (local-first positioning refresh should land first to avoid overlapping homepage copy)

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- `CLAUDE.md` — repository rules and documentation conventions.
- `README.md` — public repository summary and install entry point.
- `web/README.md` — website build and content conventions.
- `web/layouts/index.html` — current homepage free/open-source card.
- `web/content/install/_index.md` — install landing page.
- `web/content/explain/local-first.md` — local-first/no-SaaS context.

## Environment

- **Workspace:** repository root (`/Users/jusbrasil/prj/icuvisor`)
- **Services required:** None.

## File Scope

- `README.md`
- `web/layouts/index.html`
- `web/content/install/_index.md`
- `web/content/install/macos.md`
- `web/content/install/windows.md`
- `web/content/install/linux.md`
- `web/content/explain/local-first.md`
- `web/content/guides/troubleshooting.md`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Existing free/open-source/no-account claims reviewed

### Step 1: Clarify no icuvisor quota/account/subscription claim

- [ ] Update homepage and/or README copy to say icuvisor is free/open source with no icuvisor-hosted account, usage credits, or subscription quota.
- [ ] Add caveats where needed that AI clients/model providers, package registries, GitHub release downloads, and intervals.icu may have their own terms or limits.
- [ ] Keep copy concise and user-focused, avoiding defensive or competitor-comparison wording.

**Artifacts:**
- `web/layouts/index.html` (modified if needed)
- `README.md` (modified if needed)

### Step 2: Link from install or local-first docs

- [ ] Add a short explanation in install or local-first docs if users may wonder what “free” covers.
- [ ] Link to license/open-source source where helpful.
- [ ] Update troubleshooting copy only if quota/account confusion is already addressed there.
- [ ] Update `CHANGELOG.md` under `[Unreleased]` if appropriate.

**Artifacts:**
- `web/content/install/*.md` or `web/content/explain/local-first.md` (modified if needed)
- `CHANGELOG.md` (modified if needed)

### Step 3: Testing & Verification

- [ ] Run docs/site build: `make web-build`
- [ ] Run FULL test suite if non-doc/generated app files are touched: `make test`
- [ ] Build passes if app strings or generated assets are touched: `make build`
- [ ] Fix all failures or document pre-existing unrelated failures in STATUS.md

### Step 4: Documentation & Delivery

- [ ] "Must Update" docs modified
- [ ] "Check If Affected" docs reviewed
- [ ] `STATUS.md` discoveries and final notes updated
- [ ] Commit at step boundary with the task ID in the message

## Documentation Requirements

**Must Update:**
- Homepage or README free/open-source positioning — clarify no icuvisor account/credits/quota/subscription.

**Check If Affected:**
- `web/content/install/*.md` — clarify free install/use if relevant.
- `web/content/explain/local-first.md` — align no-SaaS/no-account language.
- `web/content/guides/troubleshooting.md` — add only if quota/account confusion is likely.
- `CHANGELOG.md` — add docs note if appropriate.

## Completion Criteria

- Public copy clearly communicates that icuvisor itself has no SaaS quota, onboarding credits, or subscription gate.
- Copy does not imply third-party AI clients, model providers, GitHub/package distribution, intervals.icu, or network services have no limits.
- Copy avoids competitor comparisons and unsupported claims.
- Website build passes or any pre-existing failures are documented.

## Git Commit Convention

Commits happen at **step boundaries**. All commits for this task MUST include `TP-117` for traceability. Examples:

- `docs(TP-117): complete step 1 — clarify free no-quota positioning`
- `hydrate: TP-117 expand step checkboxes`

## Do NOT

- Do not claim all AI usage is free or unlimited.
- Do not mention competitor projects, forum threads, private research notes, or issue/PR numbers.
- Do not change runtime behavior.
- Do not modify protected docs without explicit approval.

---

## Amendments

_Add amendments below this line only._
