# TP-125: Activity tags and fueling regression/docs pass — Status

**Current Step:** Step 2: Fill regression or docs gaps
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-29
**Review Level:** 1
**Review Counter:** 3
**Iteration:** 1
**Size:** S

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it — aim for 2-5 outcome-level items per step, not exhaustive implementation scripts.

---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Confirm no GPL/copyleft competitor source is opened or copied; use only public forum behavior signals and project docs.

---

### Step 1: Audit existing coverage
**Status:** ✅ Complete

- [x] Verify terse `get_activities` and `get_activity_details` tests cover present tags, empty tags, and fueling fields.
- [x] Verify `get_today` preserves tags for completed activities and planned events.
- [x] Record any missing coverage in STATUS.md Discoveries.
- [x] Run targeted tests: `go test ./internal/tools`

---

### Step 2: Fill regression or docs gaps
**Status:** 🟨 In Progress

- [ ] Add missing regression tests rather than changing already-correct behavior, including default/no-`include_full` tag preservation for present and empty tags in `get_activities` and `get_activity_details`.
- [ ] Update user-facing docs/cookbook text to mention tag-aware and fueling-aware activity reads where useful.
- [ ] Update `CHANGELOG.md` under `[Unreleased]` for the added regression coverage and docs.
- [ ] Avoid changing raw upstream field names; keep disambiguated grams suffixes.
- [ ] Run targeted tests: `go test ./internal/tools`

---

### Step 3: Testing & Verification
**Status:** ⬜ Not Started

- [ ] FULL test suite passing: `make test`
- [ ] Lint passes or pre-existing linter limitations are documented: `make lint`
- [ ] Build passes: `make build`
- [ ] All failures fixed or clearly documented as pre-existing

---

### Step 4: Documentation & Delivery
**Status:** ⬜ Not Started

- [ ] "Must Update" docs modified
- [ ] "Check If Affected" docs reviewed
- [ ] Discoveries logged

---

## Reviews

| # | Type | Step | Verdict | File |
|---|------|------|---------|------|
| 1 | plan | 1 | APPROVE | — |
| 2 | plan | 2 | REVISE | `.reviews/R002-plan-step2.md` |
| 3 | plan | 2 | REVISE | `.reviews/R003-plan-step2.md` |

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|
| `get_activity_details` lacks an empty-tags regression while `get_activities` already covers explicit empty arrays. | Add regression in Step 2. | `internal/tools/get_activity_details_test.go` |
| Existing tag tests use `include_full:true`, so they verify raw full payload preservation but not default terse tag preservation. | Add default/no-`include_full` present-and-empty tag regressions in Step 2. | `internal/tools/get_activities_test.go`, `internal/tools/get_activity_details_test.go` |

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-05-29 | Task staged | PROMPT.md and STATUS.md created |
| 2026-05-29 14:46 | Task started | Runtime V2 lane-runner execution |
| 2026-05-29 14:46 | Step 0 started | Preflight |
| 2026-05-29 | Clean-room preflight | Used only task prompt, project docs, and local repository files; no GPL/copyleft competitor source opened. |
| 2026-05-29 | Step 1 targeted tests | `go test ./internal/tools` passed (cached). |

---

## Blockers

*None*

---

## Notes

- Step 1 audit: `get_activities` covers present and empty tags plus activity fueling fields; `get_activity_details` covers present tags and fueling fields but does not yet cover empty tags.
- Step 1 audit: `get_today` covers tags on completed activities, planned workout events, and annotations in `TestGetTodayDigestUsesAthleteLocalDateAndSourceShapes`.
| 2026-05-29 14:49 | Review R001 | plan Step 1: APPROVE |
| 2026-05-29 14:53 | Review R002 | plan Step 2: UNKNOWN |
| 2026-05-29 14:54 | Review R003 | plan Step 2: REVISE |
