# TP-129: Readiness fallback guidance for null upstream readiness — Status

**Current Step:** Step 0: Preflight
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-29
**Review Level:** 2
**Review Counter:** 0
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

### Step 1: Audit wellness readiness semantics
**Status:** ⬜ Not Started

- [ ] Inspect wellness shaping, provenance metadata, native provider fields, and recovery/weekly prompt text.
- [ ] Identify whether null readiness already appears in missing_fields and whether prompts instruct cautious fallback.
- [ ] Record available fallback fields and non-goals in STATUS.md Discoveries.
- [ ] Run targeted tests: `go test ./internal/tools ./internal/prompts`

---

### Step 2: Add fallback tests or prompt guidance
**Status:** ⬜ Not Started

- [ ] Add tests if missing for null readiness with present HRV/RHR/sleep/native fields.
- [ ] Update recovery/weekly prompts so assistants do not invent readiness scores and explain missingness before fallback interpretation.
- [ ] Run targeted tests: `go test ./internal/tools ./internal/prompts`

---

### Step 3: Update cookbook docs
**Status:** ⬜ Not Started

- [ ] Update readiness-check cookbook with Garmin/null-readiness fallback examples and language.
- [ ] Keep scale labels explicit and avoid device-specific claims not backed by response fields.
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
| 2026-05-29 13:51 | Task started | Runtime V2 lane-runner execution |
| 2026-05-29 13:51 | Step 0 started | Preflight |

---

## Blockers

*None*

---

## Notes

*Reserved for execution notes*
