# TP-126: Deterministic segment-comparison analyzer workflow — Status

**Current Step:** Step 1: Audit current segment analyzer activation
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

### Step 1: Audit current segment analyzer activation
**Status:** 🟨 In Progress

- [ ] Inspect `compute_activity_segment_stats` description/schema/tests and existing eval scenarios.
- [ ] Confirm it supports distance-bounded first/last segment stats for pace/power/HR and exposes audit metadata without raw streams in terse mode.
- [ ] Record whether a higher-level helper is warranted or whether prompt/eval hardening is sufficient.
- [ ] Run targeted tests: `go test ./internal/tools`

---

### Step 2: Add segment-comparison eval/docs
**Status:** ⬜ Not Started

- [ ] Add an eval scenario for comparing first 10 km vs last 10 km that expects `compute_activity_segment_stats` rather than raw `get_activity_streams` reduction in chat.
- [ ] Update activity retrospective cookbook guidance with a deterministic segment-comparison prompt.
- [ ] If needed, tighten tool activation text without bloating core tool descriptions.
- [ ] Run targeted tests: `make eval-validate` and `go test ./internal/tools`

---

### Step 3: Add missing tests for first/last distance segments
**Status:** ⬜ Not Started

- [ ] Add or extend unit tests for distance-bounded segment stats over first and last portions of a fixture stream.
- [ ] Assert insufficient/missing stream metadata remains explicit and terse output does not dump raw stream samples.
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
| 2026-05-29 13:46 | Task started | Runtime V2 lane-runner execution |
| 2026-05-29 13:46 | Step 0 started | Preflight |

---

## Blockers

*None*

---

## Notes

*Reserved for execution notes*
