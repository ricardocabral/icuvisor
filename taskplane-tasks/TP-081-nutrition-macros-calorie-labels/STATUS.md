# TP-081: Nutrition macros and calories-label clarification — Status

**Current Step:** Step 6: Documentation & Delivery
**Status:** ✅ Complete
**Last Updated:** 2026-05-20
**Review Level:** 2
**Review Counter:** 12
**Iteration:** 1
**Size:** M

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers may expand steps when runtime discoveries warrant it.

---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Confirm no protected docs are changed without explicit approval

---

### Step 1: Map nutrition fields from upstream fixtures
**Status:** ✅ Complete

- [x] Identify carbs/protein/fat/calorie fields exposed on wellness and activity payloads.
- [x] Add typed fields with JSON names and fixtures; if an expected field is absent, document the gap.
- [x] Choose disambiguated response keys such as `carbs_g`, `protein_g`, `fat_g`, `calories_burned`, `calories_intake`/`calories_total` only when upstream semantics support them.
- [x] R002: Assert exact wellness macro fixture values to prove JSON key-to-field mapping.
- [x] R002: Request `calories` in the activity-list fixture test field filter before asserting typed calories.

---

### Step 2: Shape activity and wellness responses
**Status:** ✅ Complete

- [x] Expose macro fields only when present; keep null stripping intact.
- [x] Ensure `calories_burned` remains active/exercise calories and does not collide with intake/total fields.
- [x] Add `_meta` labels where semantics could be confused.
- [x] R004 plan: translate public keys exactly (`calories` -> `calories_burned`; wellness `kcalConsumed`/`carbohydrates`/`protein`/`fatTotal` -> `calories_intake`/`carbs_g`/`protein_g`/`fat_g`) and do not add `calories_total` or activity macro keys without fixture evidence.
- [x] R004 plan: remove legacy wellness nutrition keys from top-level terse rows while retaining raw upstream names under `full` for `include_full`.
- [x] R004 plan: put nutrition/calorie labels in a stable `_meta.field_semantics` map merged with existing wrapper/row metadata.
- [x] R004 plan: preserve present zero values by keeping wellness pointer handling and making activity `calories_burned` pointer-shaped.

---

### Step 3: Regression tests
**Status:** ✅ Complete

- [x] Add fixture-backed tests for activities and wellness with nutrition present and absent.
- [x] Assert key names and null stripping behavior.
- [x] Run targeted read-tool tests.

---

### Step 4: Docs and full verification
**Status:** ✅ Complete

- [x] Update tool reference/examples for nutrition fields.
- [x] Update CHANGELOG.md.
- [x] Run full quality gate.

---


### Step 5: Testing & Verification
**Status:** ✅ Complete

- [x] Targeted tests passing
- [x] FULL test suite passing: `make test`
- [x] Build passes: `make build`
- [x] Lint passes: `make lint`
- [x] All failures fixed or documented as pre-existing unrelated failures

---

### Step 6: Documentation & Delivery
**Status:** ✅ Complete

- [x] "Must Update" docs modified
- [x] "Check If Affected" docs reviewed
- [x] Discoveries logged
- [x] Final commit includes task ID

---

## Reviews

| # | Type | Step | Verdict | File |
|---|------|------|---------|------|
| R001 | Plan | 1 | APPROVE | .reviews/R001-plan-step1.md |
| R002 | Code | 1 | REVISE | .reviews/R002-code-step1.md |
| R003 | Code | 1 | APPROVE | .reviews/R003-code-step1.md |
| R004 | Plan | 2 | REVISE | .reviews/R004-plan-step2.md |
| R005 | Plan | 2 | APPROVE | .reviews/R005-plan-step2.md |
| R006 | Code | 2 | APPROVE | .reviews/R006-code-step2.md |
| R007 | Plan | 3 | APPROVE | .reviews/R007-plan-step3.md |
| R008 | Code | 3 | APPROVE | .reviews/R008-code-step3.md |
| R009 | Plan | 4 | APPROVE | .reviews/R009-plan-step4.md |
| R010 | Code | 4 | APPROVE | .reviews/R010-code-step4.md |
| R011 | Plan | 5 | APPROVE | .reviews/R011-plan-step5.md |
| R012 | Code | 5 | APPROVE | .reviews/R012-code-step5.md |

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|
| Upstream typed fields identified: activity `calories` means exercise calories burned; wellness exposes `kcalConsumed`, `carbohydrates`, `protein`, and `fatTotal`; no activity macro fields are present in existing fixtures/schemas. | Use disambiguated response keys; document activity macro gap instead of inventing fields. | internal/intervals/activities.go; internal/intervals/wellness.go |

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-05-20 | Task staged | PROMPT.md and STATUS.md created |
| 2026-05-20 12:41 | Task started | Runtime V2 lane-runner execution |
| 2026-05-20 12:41 | Step 0 started | Preflight |
| 2026-05-20 13:29 | Worker iter 1 | done in 2897s, tools: 207 |
| 2026-05-20 13:29 | Task complete | .DONE created |

---

## Blockers

*None*

---

## Notes

- Step 1 key mapping: activity `calories` -> `calories_burned`; wellness `kcalConsumed` -> `calories_intake`; wellness `carbohydrates`/`protein`/`fatTotal` -> `carbs_g`/`protein_g`/`fat_g`. No `calories_total` key will be emitted because current upstream fixtures do not establish a total-calories semantic.
- R002 suggestion: Reviews table/status bookkeeping should stay current when review files are created.
- Step 2 plan amendment: New nutrition labels will use `_meta.field_semantics`; wrapper metadata will cover activity list/detail semantics, and row-level wellness metadata will merge field semantics with existing provenance/scales/missing-fields metadata without overwriting them.

| 2026-05-20 12:44 | Review R001 | plan Step 1: APPROVE |
| 2026-05-20 12:49 | Review R002 | code Step 1: UNKNOWN |
| 2026-05-20 12:52 | Review R003 | code Step 1: APPROVE |
| 2026-05-20 12:54 | Review R004 | plan Step 2: REVISE |
| 2026-05-20 12:56 | Review R005 | plan Step 2: APPROVE |
| 2026-05-20 13:03 | Review R006 | code Step 2: APPROVE |
| 2026-05-20 13:05 | Review R007 | plan Step 3: APPROVE |
| 2026-05-20 13:11 | Review R008 | code Step 3: APPROVE |
| 2026-05-20 13:14 | Review R009 | plan Step 4: APPROVE |
| 2026-05-20 13:20 | Review R010 | code Step 4: APPROVE |
| 2026-05-20 13:22 | Review R011 | plan Step 5: APPROVE |
| 2026-05-20 13:26 | Review R012 | code Step 5: APPROVE |
