# TP-112: Bulk calendar write preflight guidance — Status

**Current Step:** Step 1: Add curated prompt guardrails
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-27
**Review Level:** 1
**Review Counter:** 2
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
**Status:** ✅ Complete

> **Plan-review checkpoint**

- [x] Bulk-write preflight rule added to relevant curated prompt text
- [x] Rule includes representative validate/preview, one write, readback, warning inspection, then remaining writes
- [x] Rule discourages parallel bulk writes when preservation semantics are ambiguous
- [x] Prompt golden files updated
- [x] Targeted tests passing: `go test ./internal/prompts`
- [x] R001 plan wording recorded and scoped to current warning/readback behavior without replacing default guardrails

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

- R001 proposed Step 1 prompt text: Before bulk calendar/workout writes, validate or preview one representative structured payload (use `validate_workout` for `workout_doc`/DSL when uncertain), perform one representative write, read it back, and inspect validation warnings, existing write `_meta` warning fields such as `workout_doc_warning` when present, and `workout_doc_summary`/stored description to confirm structured steps were preserved before writing the rest. Avoid parallel bulk writes while schema wording, warning metadata, or description/`workout_doc` preservation semantics are ambiguous.
- Step 1 implementation note: add this as a `Do` item in `WeeklyPlanningPrompt` rather than a custom `Guardrails` slice, keeping the default guardrails intact and the rendered prompt terse.
| 2026-05-27 19:10 | Review R001 | plan Step 1: REVISE |
| 2026-05-27 19:12 | Review R002 | plan Step 1: APPROVE |
