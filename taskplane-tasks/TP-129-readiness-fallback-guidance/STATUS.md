# TP-129: Readiness fallback guidance for null upstream readiness — Status

**Current Step:** Step 3: Update cookbook docs
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-29
**Review Level:** 2
**Review Counter:** 6
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
**Status:** ✅ Complete

- [x] Inspect wellness shaping, provenance metadata, native provider fields, and recovery/weekly prompt text.
- [x] Identify whether null readiness already appears in missing_fields and whether prompts instruct cautious fallback.
- [x] Record available fallback fields and non-goals in STATUS.md Discoveries.
- [x] Run targeted tests: `go test ./internal/tools ./internal/prompts`
- [x] Address R002: include all audited fallback/supporting wellness fields (`motivation`, `spO2`, `respiration`, `steps`, `vo2max`, `baevskySI`) in Discoveries or explicitly scope them out.

---

### Step 2: Add fallback tests or prompt guidance
**Status:** ✅ Complete

- [x] Add tests if missing for null readiness with present HRV/RHR/sleep/native fields.
- [x] Update recovery/weekly prompts so assistants do not invent readiness scores and explain missingness before fallback interpretation.
- [x] Run targeted tests: `go test ./internal/tools ./internal/prompts`

---

### Step 3: Update cookbook docs
**Status:** 🟨 In Progress

- [x] Update readiness-check cookbook with Garmin/null-readiness fallback examples and language.
- [x] Keep scale labels explicit and avoid device-specific claims not backed by response fields.
- [x] Run targeted tests/docs validation as available.

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
| R002 | code | 1 | REVISE | `.reviews/R002-code-step1.md` |
| R003 | code | 1 | APPROVE | `.reviews/R003-code-step1.md` |
| R004 | plan | 2 | APPROVE | `.reviews/R004-plan-step2.md` |
| R005 | code | 2 | APPROVE | `.reviews/R005-code-step2.md` |
| R006 | plan | 3 | APPROVE | `.reviews/R006-plan-step3.md` |

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|
| Wellness shaping strips nulls into `_meta.missing_fields`; a raw `readiness:null` should be absent from the row and listed there, but no existing fixture specifically covers Garmin/native fallback with null canonical readiness. | Add coverage in Step 2. | `internal/response/meta.go`, `internal/tools/get_wellness_data_test.go` |
| Available fallback evidence when readiness is absent: `hrv`, `hrvSDNN`, `restingHR`, `avgSleepingHR`, `sleepSecs`, `sleepScore`, `sleepQuality`, `fatigue`, `soreness`, `stress`, `mood`, `feel`, `motivation`, and where present `spO2`, `respiration`, `steps`, `vo2max`, `baevskySI`, plus `_native.<source>` fields and `_meta.provenance` scale labels. | Use in prompt/docs guidance as supporting/context signals only; state they are not a substitute score and avoid over-weighting general activity metrics such as steps/VO2max/Baevsky. | `internal/tools/get_wellness_data.go`, PRD §7.2.C |
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
| 2026-05-29 13:58 | Review R002 | code Step 1: UNKNOWN |
| 2026-05-29 14:01 | Review R003 | code Step 1: APPROVE |
| 2026-05-29 14:04 | Review R004 | plan Step 2: APPROVE |
| 2026-05-29 14:08 | Review R005 | code Step 2: APPROVE |
| 2026-05-29 14:10 | Review R006 | plan Step 3: APPROVE |
