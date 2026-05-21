# TP-078 — Installer/onboarding integration for keychain-backed credentials

**Created:** 2026-05-20
**Size:** M

## Review Level: 2

**Assessment:** Touches setup/config and secret handling, so it needs plan and code review even though patterns already exist.
**Score:** 5/8 — Blast radius: 2, Pattern novelty: 1, Security: 2, Reversibility: 0

## Canonical Task Folder

```
taskplane-tasks/TP-078-installer-onboarding-keychain-integration/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Make every non-CLI onboarding path write intervals.icu credentials through the same OS keychain abstraction used by `icuvisor setup`, so users never place API keys in JSON during GUI/basic setup. This closes the v0.5 gap between the shipped keychain backend and the installer/onboarding UX that non-technical athletes actually use.

## Dependencies

- **Task:** TP-036 (OS keychain credential storage must exist)
- **Task:** TP-038 (first-run onboarding/setup flow must exist)
- **Task:** TP-037 (installer/manual client config docs provide current packaging context)

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- CLAUDE.md — repository rules, clean-room constraints, Go/MCP conventions.
- ROADMAP.md — exact roadmap entry and phase positioning.
- docs/prd/PRD-icuvisor.md — product contract, tool catalog, response-shaping rules.
- internal/app/setup*.go — current setup/onboarding flow.
- internal/credstore/* — keychain abstraction and platform support.
- internal/config/* — config write/redaction rules.
- web/content/install/* and web/content/guides/api-key.md — user-facing install/API-key docs.

## Environment

- **Workspace:** repository root (`/Users/jusbrasil/prj/icuvisor`)
- **Services required:** None unless the task explicitly calls for live/manual validation; unit tests must not hit the network.

## File Scope

- `internal/app/setup*.go`
- `internal/credstore/*`
- `internal/config/*`
- `cmd/icuvisor/*`
- `web/content/install/*`
- `web/content/guides/api-key.md`
- `web/content/reference/config-file.md`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm no protected docs are changed without explicit approval

### Step 1: Audit current onboarding credential paths

- [ ] Identify every setup/installer path that accepts an API key.
- [ ] Confirm whether each path stores via `internal/credstore` or writes plaintext config.
- [ ] Record the desired source of truth in STATUS.md before implementation.

### Step 2: Route onboarding credential writes through the keychain

- [ ] Update setup/onboarding code so API keys are stored via `internal/credstore` only.
- [ ] Ensure generated config stores keychain references/non-secret metadata, not the secret itself.
- [ ] Preserve headless/power-user behavior without weakening redaction.

### Step 3: Add regression tests for secret handling

- [ ] Add tests proving setup writes no plaintext API key to config/loggable diagnostics.
- [ ] Add/update tests for keychain failure messages and recovery guidance.
- [ ] Run targeted tests for `internal/app`, `internal/config`, and `internal/credstore`.

### Step 4: Update install and API-key documentation

- [ ] Update end-user docs to describe the keychain-backed path.
- [ ] Remove or clearly mark any stale instruction that asks users to paste keys into JSON.
- [ ] Update CHANGELOG.md if behavior changes are user-visible.

### Step 5: Testing & Verification

- [ ] Run targeted tests added/affected by this task
- [ ] Run FULL test suite: `make test`
- [ ] Build passes: `make build`
- [ ] Lint passes: `make lint`
- [ ] Fix all failures or document pre-existing unrelated failures in STATUS.md

### Step 6: Documentation & Delivery

- [ ] "Must Update" docs modified
- [ ] "Check If Affected" docs reviewed
- [ ] `STATUS.md` discoveries and final notes updated
- [ ] Commit at step boundary with the task ID in the message

## Documentation Requirements

**Must Update:**
- CHANGELOG.md — record user-visible behavior under [Unreleased] if code or docs behavior changes.
- STATUS.md — keep execution state current.

**Check If Affected:**
- README.md — update if public setup/tool behavior changes.
- web/content/reference/tools.md — update if tool catalog descriptions or generated docs are affected.
- docs/prd/PRD-icuvisor.md — check only if behavior intentionally diverges from product scope.

## Completion Criteria

- All onboarding paths store API keys via the OS keychain abstraction.
- No generated config or diagnostics output contains plaintext API keys.
- Tests cover happy path and keychain failure guidance.
- `make test`, `make build`, and `make lint` pass.

## Git Commit Convention

Commits happen at **step boundaries**. All commits for this task MUST include `TP-078` for traceability. Examples:

- `feat(TP-078): complete step 1 — scope current behavior`
- `fix(TP-078): repair regression found during analyzer tests`
- `test(TP-078): add golden coverage for roadmap behavior`
- `hydrate: TP-078 expand step checkboxes`

## Do NOT

- Do not read, paste, paraphrase, or transliterate GPL/copyleft implementation code.
- Do not log or persist intervals.icu API keys outside the configured keychain path.
- Do not add LLM-controlled confirm/override flags for destructive behavior.
- Do not broaden the task into unrelated refactors; log follow-ups in STATUS.md discoveries.
- Do not introduce a second credential store or duplicate keychain service naming.

---

## Amendments

_Add amendments below this line only._
