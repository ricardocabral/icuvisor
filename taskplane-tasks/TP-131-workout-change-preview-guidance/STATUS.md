# TP-131: Workout change preview guidance — Status

**Current Step:** Not Started
**Status:** 🔵 Ready for Execution
**Last Updated:** 2026-05-29
**Review Level:** 2
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

### Step 1: Audit current pre-write guidance
**Status:** ⬜ Not Started

- [ ] Inspect `validate_workout`, workout write tool descriptions/examples, and weekly-planning/build-workouts prompts.
- [ ] Identify whether assistants are instructed to summarize proposed changes before writes.
- [ ] Record current behavior and chosen changes in STATUS.md Discoveries.
- [ ] Run targeted tests: `go test ./internal/tools ./internal/prompts`

---

### Step 2: Harden preview guidance
**Status:** ⬜ Not Started

- [ ] Update prompts/tool examples so proposed changes include total duration, key steps, target intensities, load/distance/time changes, and what is being preserved.
- [ ] Recommend `validate_workout` preflight for uncertain DSL or structured workout changes.
- [ ] Do not introduce a model-controlled `confirm` override or bypass safety modes.
- [ ] Run targeted tests: `go test ./internal/tools ./internal/prompts`

---

### Step 3: Update cookbook examples
**Status:** ⬜ Not Started

- [ ] Update build-workouts cookbook with before/after preview language and approval workflow.
- [ ] Ensure examples distinguish prose description from structured `workout_doc`.
- [ ] Run targeted tests/docs validation as available.

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
