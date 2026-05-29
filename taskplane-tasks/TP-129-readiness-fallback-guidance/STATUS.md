# TP-129: Readiness fallback guidance for null upstream readiness — Status

**Current Step:** Step 1: Audit wellness readiness semantics
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

### Step 1: Audit wellness readiness semantics
**Status:** 🟨 In Progress

- [x] Inspect wellness shaping, provenance metadata, native provider fields, and recovery/weekly prompt text.
- [x] Identify whether null readiness already appears in missing_fields and whether prompts instruct cautious fallback.
- [x] Record available fallback fields and non-goals in STATUS.md Discoveries.
- [x] Run targeted tests: `go test ./internal/tools ./internal/prompts`

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
| R001 | plan | 1 | APPROVE | `.reviews/R001-plan-step1.md` |

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|
| Wellness shaping strips nulls into `_meta.missing_fields`; a raw `readiness:null` should be absent from the row and listed there, but no existing fixture specifically covers Garmin/native fallback with null canonical readiness. | Add coverage in Step 2. | `internal/response/meta.go`, `internal/tools/get_wellness_data_test.go` |
| Available fallback evidence when readiness is absent: `hrv`, `hrvSDNN`, `restingHR`, `avgSleepingHR`, `sleepSecs`, `sleepScore`, `sleepQuality`, `fatigue`, `soreness`, `stress`, `mood`, `feel`, plus `_native.<source>` fields and `_meta.provenance` scale labels. | Use in prompt/docs guidance; state as supporting evidence, not a substitute score. | `internal/tools/get_wellness_data.go`, PRD §7.2.C |
| Non-goals: do not invent/normalize a readiness score from Garmin Body Battery or other native fields, do not claim device-specific readiness semantics unless the field/scale is present, and do not request API keys in chat. | Preserve in Step 2/3 wording. | `internal/prompts/testdata/recovery_check.md`, `web/content/cookbook/readiness-check.md` |
| Recovery prompt currently asks for readiness but does not explicitly say to explain missingness or use cautious fallback; weekly prompt says not to infer readiness when stale/absent but does not name fallback signals. | Update prompt golden files in Step 2. | `internal/prompts/testdata/recovery_check.md`, `internal/prompts/testdata/weekly_review.md` |

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
| 2026-05-29 13:54 | Review R001 | plan Step 1: APPROVE |
