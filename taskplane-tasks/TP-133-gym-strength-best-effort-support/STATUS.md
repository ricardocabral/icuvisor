# TP-133: Gym and strength best-effort support plan — Status

**Current Step:** Not Started
**Status:** 🔵 Ready for Execution
**Last Updated:** 2026-05-29
**Review Level:** 1
**Review Counter:** 0
**Iteration:** 0
**Size:** M

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it — aim for 2-5 outcome-level items per step, not exhaustive implementation scripts.

---

### Step 0: Preflight
**Status:** ⬜ Not Started

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm no GPL/copyleft competitor source is opened or copied; use only public forum behavior signals and project docs.

---

### Step 1: Scope current support and upstream gaps
**Status:** ⬜ Not Started

- [ ] Inspect current event/workout category handling and PRD/Roadmap strength-training mentions.
- [ ] Determine what can be represented today without inventing unsupported structured strength sets.
- [ ] Create or update an upstream-gap note for strength/gym support if missing.
- [ ] Run targeted checks/tests as relevant.

---

### Step 2: Add best-effort prompt/docs guidance
**Status:** ⬜ Not Started

- [ ] Update cookbook/prompt guidance to allow scheduling simple gym time blocks or notes when the user wants that, while explicitly saying detailed strength sets are future scope unless upstream support exists.
- [ ] Avoid adding a new write tool in this task unless upstream API support is already documented in this repository.
- [ ] Run targeted tests: `go test ./internal/prompts` if prompt fixtures change.

---

### Step 3: Capture follow-up implementation criteria
**Status:** ⬜ Not Started

- [ ] Record in docs what evidence is needed before adding first-class strength-training tools: upstream endpoints, schema fields, response shape, and safe write behavior.
- [ ] Update ROADMAP/PRD only if this clarifies existing future scope, not to expand v1 commitments.
- [ ] Run docs/test validation as available.

---

### Step 4: Testing & Verification
**Status:** ⬜ Not Started

- [ ] FULL test suite passing: `make test`
- [ ] Lint passes or pre-existing linter limitations are documented: `make lint`
- [ ] Build passes: `make build`
- [ ] All failures fixed or clearly documented as pre-existing

---

### Step 5: Documentation & Delivery
**Status:** ⬜ Not Started

- [ ] "Must Update" docs modified
- [ ] "Check If Affected" docs reviewed
- [ ] Discoveries logged

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
| 2026-05-29 | Task staged | PROMPT.md and STATUS.md created |

---

## Blockers

*None*

---

## Notes

*Reserved for execution notes*
