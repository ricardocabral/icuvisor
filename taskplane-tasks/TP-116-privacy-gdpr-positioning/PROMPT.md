# Task: TP-116 - Privacy and GDPR positioning

**Created:** 2026-05-27
**Size:** S

## Review Level: 1

**Assessment:** Documentation-only privacy positioning, but privacy/GDPR language needs plan review to avoid overclaiming legal compliance.
**Score:** 2/8 — Blast radius: 1, Pattern novelty: 0, Security: 1, Reversibility: 0

## Canonical Task Folder

```
taskplane-tasks/TP-116-privacy-gdpr-positioning/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Add clear privacy positioning for athletes and coaches evaluating icuvisor: local-first by default, no icuvisor-hosted athlete database, API keys stored in the OS keychain, loopback-only HTTP by default, and a smaller trust boundary than hosted coaching or connector services. The copy should be useful to EU/privacy-conscious users while avoiding legal conclusions such as “GDPR compliant” unless the repository already substantiates them.

## Dependencies

- **Task:** TP-113 (local-first positioning refresh should land first to avoid overlapping homepage/privacy copy)

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- `CLAUDE.md` — repository rules and documentation conventions.
- `SECURITY.md` — security policy and credential-handling commitments.
- `web/README.md` — website build and content conventions.
- `web/content/explain/local-first.md` — existing privacy/local-first explanation.
- `web/content/explain/safety-modes.md` — write/delete safety positioning.
- `web/content/explain/coach-mode.md` — coach-mode trust boundary.
- `web/content/guides/http-transport.md` — loopback/LAN bind behavior.

## Environment

- **Workspace:** repository root (`/Users/jusbrasil/prj/icuvisor`)
- **Services required:** None.

## File Scope

- `web/content/explain/privacy.md`
- `web/content/explain/local-first.md`
- `web/content/explain/coach-mode.md`
- `web/content/explain/_index.md`
- `web/content/guides/http-transport.md`
- `web/layouts/index.html`
- `README.md`
- `SECURITY.md`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Existing privacy, security, coach-mode, and HTTP transport claims reviewed

### Step 1: Define privacy claims and boundaries

- [ ] Inventory substantiated claims already present in docs/code: OS keychain, no icuvisor SaaS in local mode, loopback default, delete-mode gating, coach key stays local.
- [ ] Define explicit non-claims: not legal advice, not certified compliance, chosen AI client may process conversation content, intervals.icu remains the upstream data processor/controller per its own terms.
- [ ] Decide whether to add a dedicated privacy explanation page or strengthen existing local-first/coach-mode pages.

**Plan-review checkpoint**

**Artifacts:**
- `STATUS.md` discoveries/plan notes (modified)

### Step 2: Add privacy-conscious user-facing copy

- [ ] Add or update docs explaining the local trust boundary, credential storage, HTTP bind defaults, and coach-mode credential handling.
- [ ] Include EU/GDPR-conscious language framed as design posture and questions to ask, not as legal certification.
- [ ] Add a concise homepage/README pointer only if it improves discoverability without overloading the main pitch.

**Artifacts:**
- `web/content/explain/privacy.md` or existing explain pages (new/modified)
- `web/layouts/index.html` or `README.md` (modified if needed)

### Step 3: Link from relevant docs and update changelog

- [ ] Link privacy positioning from local-first, coach-mode, HTTP transport, and/or safety docs where relevant.
- [ ] Ensure `SECURITY.md` remains authoritative for security policy; link to it rather than duplicating details if possible.
- [ ] Update `CHANGELOG.md` under `[Unreleased]` if appropriate.

**Artifacts:**
- Relevant `web/content/**/*.md` pages (modified)
- `CHANGELOG.md` (modified if needed)

### Step 4: Testing & Verification

- [ ] Run docs/site build: `make web-build`
- [ ] Check rendered links and page placement.
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
- Privacy/local-first explanation docs — add clear privacy-conscious positioning.
- `STATUS.md` — record claim boundaries and any legal/privacy caveats discovered.

**Check If Affected:**
- `SECURITY.md` — link only if a small pointer helps; do not rewrite policy without need.
- `web/content/explain/coach-mode.md` — ensure coach-mode privacy boundary is clear.
- `web/content/guides/http-transport.md` — ensure loopback/LAN-bind behavior is clear.
- `web/layouts/index.html` and `README.md` — add concise pointer if appropriate.
- `CHANGELOG.md` — add docs note if appropriate.

## Completion Criteria

- Privacy-conscious users can understand what icuvisor does and does not custody.
- Docs mention OS keychain, local operation, loopback default, and AI-client/upstream caveats.
- No unsubstantiated “GDPR compliant” or legal-certification claims are introduced.
- Website build passes or any pre-existing failures are documented.

## Git Commit Convention

Commits happen at **step boundaries**. All commits for this task MUST include `TP-116` for traceability. Examples:

- `docs(TP-116): complete step 2 — add privacy positioning`
- `hydrate: TP-116 expand step checkboxes`

## Do NOT

- Do not claim icuvisor is certified GDPR compliant or provide legal advice.
- Do not mention competitor projects, forum threads, private research notes, or issue/PR numbers.
- Do not duplicate large parts of `SECURITY.md` into marketing pages.
- Do not change runtime behavior.
- Do not modify protected docs without explicit approval.

---

## Amendments

_Add amendments below this line only._
