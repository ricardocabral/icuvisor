# TP-127: Edit-in-place write safety evals — Status

**Current Step:** Step 1: Audit write/delete guidance
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-29
**Review Level:** 2
**Review Counter:** 1
**Iteration:** 1
**Size:** M

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it — aim for 2-5 outcome-level items per step, not exhaustive implementation scripts.

---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Confirm no GPL/copyleft competitor source is opened or copied; use only public forum behavior signals and project docs.

---

### Step 1: Audit write/delete guidance
**Status:** 🟨 In Progress

- [ ] Inspect create/update/delete workout and event tool descriptions, schemas, and safety tests, including `create_workout` as the unsafe recreate side.
- [ ] Inspect registration-time delete gating coverage in `internal/safety/adversarial_test.go` and decide whether `go test ./internal/safety` is needed.
- [ ] Identify whether existing descriptions already prefer update/edit in place and where eval coverage is missing.
- [ ] Record the current safety contract and any token-budget tradeoff in STATUS.md Discoveries.
- [ ] Run targeted tests: `go test ./internal/tools`

---

### Step 2: Add eval/adversarial coverage
**Status:** ⬜ Not Started

- [ ] Add at least one eval/adversarial scenario where the user asks to change tomorrow’s workout and the assistant must choose update/edit tools, not delete/create.
- [ ] Assert safe-mode/delete-mode messaging remains short and actionable when deletion is unavailable.
- [ ] Run targeted tests: `make eval-validate` and `go test ./internal/tools`

---

### Step 3: Harden guidance if necessary
**Status:** ⬜ Not Started

- [ ] Update concise tool descriptions or cookbook prompts only where tests show ambiguity.
- [ ] Do not add a model-controlled `confirm` flag or weaken registration-time gating.
- [ ] Run targeted tests: `go test ./internal/tools`

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
| 2026-05-29 14:06 | Task started | Runtime V2 lane-runner execution |
| 2026-05-29 14:06 | Step 0 started | Preflight |

---

## Blockers

*None*

---

## Notes

- Step 1 plan review R001 requested explicit `create_workout` audit coverage and registration-time delete gating coverage.

| 2026-05-29 14:10 | Review R001 | plan Step 1: REVISE |
