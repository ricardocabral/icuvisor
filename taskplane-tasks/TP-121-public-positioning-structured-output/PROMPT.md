# Task: TP-121 - Public positioning for structured local-first output

**Created:** 2026-05-29
**Size:** S

## Review Level: 0 (None)

**Assessment:** This is a documentation-only positioning pass with no code or runtime behavior changes. It should be careful, factual, and avoid naming competitors or making unsupported claims.
**Score:** 1/8 — Blast radius: 0, Pattern novelty: 1, Security: 0, Reversibility: 0

## Canonical Task Folder

```
taskplane-tasks/TP-121-public-positioning-structured-output/
├── PROMPT.md   ← This file (immutable above --- divider)
├── STATUS.md   ← Execution state (worker updates this)
├── .reviews/   ← Reviewer output (created by the orchestrator runtime)
└── .DONE       ← Created when complete
```

## Mission

Improve icuvisor's public-facing positioning around its existing advantages: single-binary install, local-first/keychain credentials, terse structured JSON responses, unit-labelled fields, and registration-time delete safety. The goal is to make those differentiators visible to users without disparaging or naming competitor projects.

## Dependencies

- **None**

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- `README.md` — primary public landing copy in this repo.
- `docs/prd/PRD-icuvisor.md` — source of truth for product claims and differentiators.
- `SECURITY.md` — credential/keychain and installer integrity wording.
- `CONTRIBUTING.md` — documentation/process expectations.

## Environment

- **Workspace:** repository root
- **Services required:** None

## File Scope

- `README.md`
- `docs/prd/PRD-icuvisor.md`
- `docs/clients/*.md`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm no competitor source or GPL/copyleft material is opened or copied; use only icuvisor docs and the summarized public opportunity.

### Step 1: Add concise public differentiator copy

- [ ] Add or refine a short README section that explains why icuvisor is different: local-first, single binary, no API keys in tool arguments, terse structured JSON, unit-labelled fields, and delete-mode gating.
- [ ] Keep tone factual and user-centered; do not name, mock, or compare directly against specific repositories.
- [ ] Ensure every claim is already supported by current code/docs or soften it to an aspiration if roadmap-only.
- [ ] Run targeted documentation check: `markdownlint README.md` if available, otherwise manually review Markdown rendering.

**Artifacts:**
- `README.md` (modified)

### Step 2: Align docs if wording exposes drift

- [ ] Check PRD/security/client docs for conflicting wording introduced by the README change.
- [ ] Update only the smallest relevant docs if the README exposes stale or contradictory claims.
- [ ] Add a CHANGELOG documentation note if user-visible docs materially change.

**Artifacts:**
- `docs/prd/PRD-icuvisor.md` (modified only if needed)
- `docs/clients/*.md` (modified only if needed)
- `CHANGELOG.md` (modified if needed)

### Step 3: Testing & Verification

- [ ] Run FULL test suite: `make test` if docs changes affect generated references or examples; otherwise record why code tests were not necessary.
- [ ] Fix all failures or documentation issues
- [ ] Build passes: `make build` if docs changed generated/catalog-adjacent references; otherwise record not applicable.

### Step 4: Documentation & Delivery

- [ ] "Must Update" docs modified
- [ ] "Check If Affected" docs reviewed
- [ ] Discoveries logged in STATUS.md

## Documentation Requirements

**Must Update:**
- `README.md` — public differentiator copy.

**Check If Affected:**
- `docs/prd/PRD-icuvisor.md` — product claims must remain consistent.
- `SECURITY.md` — credential/security wording must remain consistent.
- `docs/clients/*.md` — update only if client setup claims become stale.
- `CHANGELOG.md` — add a docs note only if warranted.

## Completion Criteria

- [ ] README clearly communicates icuvisor's local-first structured-output advantages.
- [ ] No direct competitor naming or unsupported claims are introduced.
- [ ] Related docs are checked for consistency.

## Git Commit Convention

Commits happen at **step boundaries** (not after every checkbox). All commits for this task MUST include the task ID for traceability:

- **Step completion:** `docs(TP-121): complete Step N — description`
- **Bug fixes:** `fix(TP-121): description`
- **Tests:** `test(TP-121): description`
- **Hydration:** `hydrate: TP-121 expand Step N checkboxes`

## Do NOT

- Do not read, copy, paraphrase, or port GPL/copyleft competitor source code.
- Do not name or disparage competitor repositories in public copy.
- Do not claim completed installer/client support beyond current repository reality.
- Do not modify code.
- Do not skip consistency review.

---

## Amendments (Added During Execution)

<!-- Workers add amendments here if issues discovered during execution.
     Format:
     ### Amendment N — YYYY-MM-DD HH:MM
     **Issue:** [what was wrong]
     **Resolution:** [what was changed] -->
