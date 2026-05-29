# TP-124: Activity date resolution and detail-routing evals — Status

**Current Step:** Step 1: Map current routing hints
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

### Step 1: Map current routing hints
**Status:** 🟨 In Progress

- [ ] Inspect `get_activities`, `get_activity_details`, `get_activity_intervals`, `get_activity_splits`, cookbook prompts, prompt testdata, and eval scenarios.
- [ ] Identify where prompts/tool descriptions fail to instruct list-by-date before detail/interval fetch.
- [ ] Record any gaps and chosen changes in STATUS.md Discoveries.
- [ ] Run targeted tests: `go test ./internal/tools ./internal/prompts`

---

### Step 2: Add eval scenarios
**Status:** ⬜ Not Started

- [ ] Add at least one eval scenario for “analyze my race last Sunday” that requires activity-date lookup then detail/interval fetch.
- [ ] Add one scenario for “show/compare lap splits or reps for my run on [date]” that must not stop at session summaries.
- [ ] Validate expected tool-use ordering and grounding rubric.
- [ ] Run targeted tests: `make eval-validate`

---

### Step 3: Harden descriptions or cookbook guidance
**Status:** ⬜ Not Started

- [ ] If gaps are found, update tool descriptions or cookbook text to make the list→detail/interval path explicit.
- [ ] Avoid adding broad tool-description tokens unless the eval requires them; prefer concise activation hints.
- [ ] Run targeted tests: `go test ./internal/tools ./internal/prompts` and `make eval-validate`

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
| R001 | Plan | 1 | REVISE | `.reviews/R001-plan-step1.md` |

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-05-29 | Task staged | PROMPT.md and STATUS.md created |
| 2026-05-29 13:19 | Task started | Runtime V2 lane-runner execution |
| 2026-05-29 13:19 | Step 0 started | Preflight |

---

## Blockers

*None*

---

## Notes

Plan review R001 required adding `get_activity_splits` / `internal/tools/get_activity_streams.go` to the Step 1 mapping scope and separating date, ID-routing, and split-vs-interval discoveries.
| 2026-05-29 13:22 | Review R001 | plan Step 1: REVISE |
