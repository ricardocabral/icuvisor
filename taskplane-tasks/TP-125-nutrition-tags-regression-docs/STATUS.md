# TP-125: Activity tags and fueling regression/docs pass — Status

**Current Step:** Step 1: Audit existing coverage
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-29
**Review Level:** 1
**Review Counter:** 0
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
**Status:** 🟨 In Progress

- [ ] Verify terse `get_activities` and `get_activity_details` tests cover present tags, empty tags, and fueling fields.
- [ ] Verify `get_today` preserves tags for completed activities and planned events.
- [ ] Record any missing coverage in STATUS.md Discoveries.
- [ ] Run targeted tests: `go test ./internal/tools`

---

### Step 2: Fill regression or docs gaps
**Status:** ⬜ Not Started

- [ ] Add missing regression tests rather than changing already-correct behavior.
- [ ] Update user-facing docs/cookbook text to mention tag-aware and fueling-aware activity reads where useful.
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

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-05-29 | Task staged | PROMPT.md and STATUS.md created |
| 2026-05-29 14:46 | Task started | Runtime V2 lane-runner execution |
| 2026-05-29 14:46 | Step 0 started | Preflight |
| 2026-05-29 | Clean-room preflight | Used only task prompt, project docs, and local repository files; no GPL/copyleft competitor source opened. |

---

## Blockers

*None*

---

## Notes

*Reserved for execution notes*
