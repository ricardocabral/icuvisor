# Task: TP-113 - Local-first positioning refresh

**Created:** 2026-05-27
**Size:** S

## Review Level: 0

**Assessment:** Documentation and website copy refresh using existing local-first positioning; no runtime behavior or schema changes.
**Score:** 1/8 — Blast radius: 1, Pattern novelty: 0, Security: 0, Reversibility: 0

## Canonical Task Folder

```
taskplane-tasks/TP-113-local-first-positioning-refresh/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Refresh the public positioning so icuvisor clearly leads with the value proposition: local-first MCP for intervals.icu, no icuvisor SaaS account, no third-party custody of the intervals.icu API key, and fewer moving parts than hosted connector/OAuth flows. The copy should be confident but not combative: explain what users get without naming competitors, forums, or private research notes.

## Dependencies

- **None**

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- `CLAUDE.md` — repository rules and documentation conventions.
- `README.md` — public repository landing page.
- `web/README.md` — website build and content conventions.
- `web/layouts/index.html` — current landing page copy.
- `web/content/explain/local-first.md` — existing local-first explanation.

## Environment

- **Workspace:** repository root (`/Users/jusbrasil/prj/icuvisor`)
- **Services required:** None.

## File Scope

- `README.md`
- `web/hugo.toml`
- `web/layouts/index.html`
- `web/content/explain/local-first.md`
- `web/content/explain/_index.md`
- `web/content/guides/troubleshooting.md`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Current homepage, README, and local-first explanation reviewed for existing claims

### Step 1: Refresh homepage and metadata positioning

- [ ] Update the landing-page hero/feature/privacy copy to foreground local-first operation, keychain credential storage, no icuvisor SaaS account, and simple MCP setup.
- [ ] Update site metadata or README summary copy if it undersells the local-first/no-custody value proposition.
- [ ] Keep copy accurate for all supported clients; avoid promising that external AI providers never receive conversation content.

**Artifacts:**
- `web/layouts/index.html` (modified)
- `web/hugo.toml` (modified if metadata changes)
- `README.md` (modified if summary changes)

### Step 2: Strengthen explanatory copy and links

- [ ] Update `web/content/explain/local-first.md` with a concise comparison of local binary + OS keychain versus hosted connector/OAuth-style flows.
- [ ] Add or adjust links from related docs so privacy/local-first claims are easy to verify.
- [ ] Update `CHANGELOG.md` under `[Unreleased]` if the user-facing docs change is notable.

**Artifacts:**
- `web/content/explain/local-first.md` (modified)
- Related docs links (modified if needed)
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
- `web/layouts/index.html` — refresh primary marketing copy.
- `web/content/explain/local-first.md` — explain the local-first/no-custody model.

**Check If Affected:**
- `README.md` — keep repository summary aligned with homepage.
- `web/hugo.toml` — keep search/social description aligned.
- `web/content/guides/troubleshooting.md` — link if local setup avoids a common failure mode.
- `CHANGELOG.md` — add docs note if appropriate.

## Completion Criteria

- Public copy clearly says icuvisor runs locally and does not custody API keys.
- Copy explains the reliability/privacy benefit without naming or disparaging competitors.
- Claims remain accurate about what the chosen AI client may send to its model provider.
- Website build passes or any pre-existing failures are documented.

## Git Commit Convention

Commits happen at **step boundaries**. All commits for this task MUST include `TP-113` for traceability. Examples:

- `docs(TP-113): complete step 1 — refresh local-first positioning`
- `hydrate: TP-113 expand step checkboxes`

## Do NOT

- Do not mention competitor projects, forum threads, private research notes, or issue/PR numbers.
- Do not use combative phrasing like “flaky” in public copy; translate it into calm reliability language.
- Do not claim external AI providers never receive conversation content.
- Do not change runtime behavior.
- Do not modify protected docs without explicit approval.

---

## Amendments

_Add amendments below this line only._
