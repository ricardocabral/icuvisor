# TP-112: Bulk calendar write preflight guidance — Status

**Current Step:** Step 0: Preflight
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-27
**Review Level:** 1
**Review Counter:** 0
**Iteration:** 1
**Size:** S

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it.

---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Existing prompt and cookbook guidance reviewed

---

### Step 1: Add curated prompt guardrails
**Status:** ⬜ Not Started

> **Plan-review checkpoint**

- [ ] Bulk-write preflight rule added to relevant curated prompt text
- [ ] Rule includes representative validate/preview, one write, readback, warning inspection, then remaining writes
- [ ] Rule discourages parallel bulk writes when preservation semantics are ambiguous
- [ ] Prompt golden files updated
- [ ] Targeted tests passing: `go test ./internal/prompts`

---

### Step 2: Add user-facing cookbook guidance
**Status:** ⬜ Not Started

- [ ] Workout/calendar cookbook guidance updated
- [ ] Structured-step preservation risk explained concisely
- [ ] Guidance remains client-neutral
- [ ] `CHANGELOG.md` updated

---

### Step 3: Testing & Verification
**Status:** ⬜ Not Started

- [ ] FULL test suite passing: `make test`
- [ ] Lint passing or documented: `make lint`
- [ ] Build passes: `make build`
- [ ] All failures fixed or documented as pre-existing unrelated failures

---

### Step 4: Documentation & Delivery
**Status:** ⬜ Not Started

- [ ] "Must Update" docs modified
- [ ] "Check If Affected" docs reviewed
- [ ] Discoveries logged
- [ ] Step-boundary commit includes `TP-112`

---

## Reviews

| # | Type | Step | Verdict | File |
|---|------|------|---------|------|

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-05-27 | Task staged | PROMPT.md and STATUS.md created |
| 2026-05-27 19:06 | Task started | Runtime V2 lane-runner execution |
| 2026-05-27 19:06 | Step 0 started | Preflight |

---

## Blockers

*None*

---

## Notes

*Reserved for execution notes*
