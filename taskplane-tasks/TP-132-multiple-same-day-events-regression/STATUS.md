# TP-132: Multiple same-day events regression pack — Status

**Current Step:** Step 1: Audit same-day event handling
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

### Step 1: Audit same-day event handling
**Status:** 🟨 In Progress

- [ ] Inspect `get_today` and `get_events` shaping/tests for multiple same-day planned workouts, notes, and races.
- [ ] Confirm athlete-local date filtering does not collapse rows by date.
- [ ] Record missing cases in STATUS.md Discoveries.
- [ ] Run targeted tests: `go test ./internal/tools`

---

### Step 2: Add regression coverage
**Status:** ⬜ Not Started

- [ ] Add tests with at least two WORKOUT events on the same date plus optional NOTE/race annotations.
- [ ] Assert both entries are present, separately identifiable, and not overwritten by map/date grouping.
- [ ] If useful, add an eval scenario for “what is on tomorrow?” with two planned sessions.
- [ ] Run targeted tests: `go test ./internal/tools` and `make eval-validate` if eval changed.

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
| 2026-05-29 16:03 | Task started | Runtime V2 lane-runner execution |
| 2026-05-29 16:03 | Step 0 started | Preflight |

---

## Blockers

*None*

---

## Notes

*Reserved for execution notes*
