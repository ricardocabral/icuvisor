# TP-233: Support indoor FTP and missing sport-setting creation — Status

**Current Step:** Step 1: Extend the typed client
**Status:** 🟡 In Progress
**Last Updated:** 2026-07-10
**Review Level:** 2
**Review Counter:** 1
**Iteration:** 2
**Size:** L

---

### Step 0: Preflight

**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] TP-228 and TP-229 are complete
- [x] Public create contract confirmed

---

### Step 1: Extend the typed client

**Status:** 🟨 In Progress

- [ ] Indoor FTP update support added
- [ ] Typed CreateSportSettings operation added
- [ ] Corrected threshold pace conversion reused
- [ ] Threshold validation defined without invented constraints
- [ ] Targeted client tests pass
- [ ] R001 plan: exact typed boundary, sparse POST contract, validation, and client regression coverage recorded

---

### Step 2: Add and register MCP surfaces

**Status:** ⬜ Not Started

- [ ] update_sport_settings indoor_ftp field added
- [ ] create_sport_settings tool implemented
- [ ] Full-toolset registration and annotations added
- [ ] Schemas, examples, catalog, and snapshots updated

---

### Step 3: Regression and safety coverage

**Status:** ⬜ Not Started

- [ ] Exact create wire tests added
- [ ] Invalid input avoids network calls
- [ ] Credential/confirm/zone arguments excluded
- [ ] Tool counts and catalog guards updated

---

### Step 4: Generated docs and public contract

**Status:** ⬜ Not Started

- [ ] PRD catalog updated
- [ ] Website tool/schema data regenerated
- [ ] Changelog updated

---

### Step 5: Testing & Verification

**Status:** ⬜ Not Started

- [ ] FULL test suite passing
- [ ] Race suite passing
- [ ] Lint passing
- [ ] Build passes
- [ ] Generated docs clean

---

### Step 6: Documentation & Delivery

**Status:** ⬜ Not Started

- [ ] Must Update docs modified
- [ ] Check If Affected docs reviewed
- [ ] Discoveries logged

---

## Reviews

| # | Type | Step | Verdict | File |
|---|------|------|---------|------|
| R001 | Plan | 1 | REVISE | `.reviews/R001-plan-step1.md` |

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-07-10 | Task staged | PROMPT.md and STATUS.md created |
| 2026-07-10 22:35 | Task started | Runtime V2 lane-runner execution |
| 2026-07-10 22:35 | Step 0 started | Preflight |
| 2026-07-10 22:40 | Worker iter 1 | done in 264s, tools: 25 |

## Blockers

- TP-228 and TP-229 must complete first.

## Notes

*Reserved for execution notes*
| 2026-07-10 22:40 | Review R001 | plan Step 1: REVISE |
