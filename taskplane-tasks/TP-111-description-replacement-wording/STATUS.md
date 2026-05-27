# TP-111: Clarify description replacement wording — Status

**Current Step:** Not Started
**Status:** 🔵 Ready for Execution
**Last Updated:** 2026-05-27
**Review Level:** 1
**Review Counter:** 0
**Iteration:** 0
**Size:** S

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it.

---

### Step 0: Preflight
**Status:** ⬜ Not Started

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Existing write-tool descriptions and docs reviewed

---

### Step 1: Update write-tool wording
**Status:** ⬜ Not Started

> **Plan-review checkpoint**

- [ ] `add_or_update_event` wording clarifies replacement semantics and structured-step risk
- [ ] `create_workout` / `update_workout` wording checked and updated
- [ ] `update_activity` wording checked for consistency
- [ ] Schema snapshots updated
- [ ] Targeted tests passing: `go test ./internal/tools`

---

### Step 2: Update prompt/docs wording
**Status:** ⬜ Not Started

- [ ] Weekly-planning prompt reviewed/updated
- [ ] Cookbook/explainer docs reviewed/updated
- [ ] Prompt golden tests updated if needed
- [ ] `CHANGELOG.md` updated
- [ ] Targeted tests passing: `go test ./internal/prompts ./internal/tools`

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
- [ ] Step-boundary commit includes `TP-111`

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

---

## Blockers

*None*

---

## Notes

*Reserved for execution notes*
