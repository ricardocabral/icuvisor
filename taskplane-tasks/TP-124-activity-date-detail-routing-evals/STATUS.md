# TP-124: Activity date resolution and detail-routing evals — Status

**Current Step:** Step 4: Testing & Verification
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-29
**Review Level:** 2
**Review Counter:** 10
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
**Status:** ✅ Complete

- [x] Inspect `get_activities`, `get_activity_details`, `get_activity_intervals`, `get_activity_splits`, cookbook prompts, prompt testdata, and eval scenarios.
- [x] Identify where prompts/tool descriptions fail to instruct list-by-date before detail/interval fetch.
- [x] Record any gaps and chosen changes in STATUS.md Discoveries.
- [x] Run targeted tests: `go test ./internal/tools ./internal/prompts`

---

### Step 2: Add eval scenarios
**Status:** ✅ Complete

- [x] Add at least one eval scenario for “analyze my race last Sunday” that requires activity-date lookup then detail/interval fetch.
- [x] Add one scenario for “show/compare lap splits or reps for my run on [date]” that must not stop at session summaries.
- [x] Validate expected tool-use ordering and grounding rubric.
- [x] Run targeted tests: `make eval-validate`

---

### Step 3: Harden descriptions or cookbook guidance
**Status:** ✅ Complete

- [x] If gaps are found, update tool descriptions or cookbook text to make the list→detail/interval/splits path explicit, including `internal/tools/get_activity_streams.go` when split hints change.
- [x] Keep downstream `activity_id` hints concise: resolve described/date-based activities with `get_activities` over the athlete-local date window, then pass the returned `activity_id`.
- [x] If tool catalog text changes, regenerate generated tool docs/data with `make docs-tools` (or document why not needed before Step 5).
- [x] Run targeted tests: `go test ./internal/tools ./internal/prompts` and `make eval-validate`
- [x] R008: Update `cmd/gendocs/testdata/tools.golden.json` for changed generated catalog summaries and verify `go test ./cmd/gendocs`.

---

### Step 4: Testing & Verification
**Status:** 🟨 In Progress

- [x] FULL test suite passing: `make test`
- [x] Lint passes or pre-existing linter limitations are documented: `make lint`
- [x] Build passes: `make build`
- [x] All failures fixed or clearly documented as pre-existing

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
| R002 | Plan | 1 | APPROVE | `.reviews/R002-plan-step1.md` |
| R003 | Code | 1 | APPROVE | `.reviews/R003-code-step1.md` |
| R004 | Plan | 2 | APPROVE | `.reviews/R004-plan-step2.md` |
| R005 | Code | 2 | APPROVE | `.reviews/R005-code-step2.md` |
| R006 | Plan | 3 | REVISE | `.reviews/R006-plan-step3.md` |
| R007 | Plan | 3 | APPROVE | `.reviews/R007-plan-step3.md` |
| R008 | Code | 3 | REVISE | `.reviews/R008-code-step3.md` |
| R009 | Code | 3 | APPROVE | `.reviews/R009-code-step3.md` |
| R010 | Plan | 4 | APPROVE | `.reviews/R010-plan-step4.md` |

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|
| Relative-date prompts like "last Sunday" need explicit athlete-local date-window resolution before selecting an activity ID. | Add eval coverage and concise cookbook/tool hints where needed. | `get_activities` description/schema; `web/content/cookbook/activity-retrospective.md`; `scripts/eval/scenarios/cookbook_scenarios.json` |
| Detail, intervals, and splits tools all require `activity_id`; only `get_activities` currently hints it should precede detail/interval/splits fetches, so ID-routing is one-sided. | Harden concise activation hints on activity detail/interval/splits descriptions or cookbook guidance. | `internal/tools/get_activity_details.go`; `internal/tools/get_activity_streams.go` |
| Existing cookbook evals include most-recent activity/test scans but not race-by-date detail analysis or splits/reps-by-date scenarios. | Add two routing eval scenarios with expected list→detail/interval/splits ordering and anti-patterns. | `scripts/eval/scenarios/cookbook_scenarios.json` |

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

Step 1 inspection identified three routing gaps: detail/interval/splits tool descriptions require `activity_id` but do not remind assistants to resolve described/date-based activities through `get_activities`; activity-retrospective cookbook says to list recent activities when no ID is supplied but does not explicitly say to query the athlete-local date window for relative dates like "last Sunday"; existing eval scenarios lack a race-by-date and splits/reps-by-date regression.

Plan review R006 required Step 3 to include split hinting in `internal/tools/get_activity_streams.go`, concise athlete-local date-window ID-routing wording, and generated tool docs/data sync if tool catalog descriptions change.

Code review R008 found stale `cmd/gendocs/testdata/tools.golden.json` summaries after catalog text changes; update the golden fixture and run `go test ./cmd/gendocs`.
| 2026-05-29 13:22 | Review R001 | plan Step 1: REVISE |
| 2026-05-29 13:24 | Review R002 | plan Step 1: APPROVE |
| 2026-05-29 13:27 | Review R003 | code Step 1: APPROVE |
| 2026-05-29 13:29 | Review R004 | plan Step 2: APPROVE |
| 2026-05-29 13:33 | Review R005 | code Step 2: APPROVE |
| 2026-05-29 13:35 | Review R006 | plan Step 3: REVISE |
| 2026-05-29 13:37 | Review R007 | plan Step 3: APPROVE |
| 2026-05-29 13:42 | Review R008 | code Step 3: REVISE |
| 2026-05-29 13:45 | Review R009 | code Step 3: APPROVE |
| 2026-05-29 13:47 | Review R010 | plan Step 4: APPROVE |
