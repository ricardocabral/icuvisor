# TP-111: Clarify description replacement wording — Status

**Current Step:** Step 3: Testing & Verification
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-27
**Review Level:** 1
**Review Counter:** 3
**Iteration:** 3
**Size:** S

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it.

---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Existing write-tool descriptions and docs reviewed

---

### Step 1: Update write-tool wording
**Status:** ✅ Complete

> **Plan-review checkpoint**

- [x] Plan-review R001 final wording pattern recorded in STATUS.md before source edits
- [x] Plan-review R001 snapshot verification owner included
- [x] `add_or_update_event` wording clarifies replacement semantics and structured-step risk
- [x] `create_workout` / `update_workout` wording checked and updated
- [x] `update_activity` wording checked for consistency
- [x] Schema snapshots updated
- [x] Targeted tests passing: `go test ./internal/tools`

---

### Step 2: Update prompt/docs wording
**Status:** ✅ Complete

- [x] Weekly-planning prompt reviewed/updated
- [x] Cookbook/explainer docs reviewed/updated
- [x] Prompt golden tests updated if needed
- [x] `CHANGELOG.md` updated
- [x] Targeted tests passing: `go test ./internal/prompts ./internal/tools`

---

### Step 3: Testing & Verification
**Status:** 🟨 In Progress

- [ ] FULL test suite passing: `make test`
- [ ] Lint passing or documented: `make lint`
- [ ] Build passes: `make build`
- [ ] All failures fixed or documented as pre-existing unrelated failures

---

### Step 4: Documentation & Delivery
**Status:** ⬜ Not Started

- [ ] "Must Update" docs modified
- [ ] "Check If Affected" docs reviewed
- [ ] Discoveries logged
- [ ] Step-boundary commit includes `TP-111`

---

## Reviews

| # | Type | Step | Verdict | File |
|---|------|------|---------|------|
| R001 | plan | 1 | REVISE | `.reviews/R001-plan-step1.md` |
| R002 | plan | 1 | APPROVE | `.reviews/R002-plan-step1.md` |
| R003 | plan | 2 | APPROVE | `.reviews/R003-plan-step2.md` |

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-05-27 | Task staged | PROMPT.md and STATUS.md created |
| 2026-05-27 18:13 | Task started | Runtime V2 lane-runner execution |
| 2026-05-27 18:13 | Step 0 started | Preflight |
| 2026-05-27 18:14 | Worker iter 1 | done in 89s, tools: 11 |
| 2026-05-27 18:22 | Dependency check | TP-109 still `🔵 Ready for Execution`; Step 0 not started |
| 2026-05-27 18:15 | Agent escalate | TP-111 is blocked at Step 0 `Dependencies satisfied`: dependency TP-109 remains `🔵 Ready for Execution` with Step 0 not started, so I cannot align warning terminology safely or proceed to wording upd |
| 2026-05-27 18:15 | Worker iter 2 | done in 75s, tools: 7 |
| 2026-05-27 18:15 | No progress | Iteration 2: 0 new checkboxes (1/3 stall limit) |
| 2026-05-27 18:30 | Steering | TP-109 is non-blocking; proceed with current-state wording-only implementation and no references to nonexistent TP-109 behavior |
| 2026-05-27 18:35 | Preflight review | Reviewed TP-109 terminology intent, write-tool description/schema text, weekly planning prompt, cookbook, calendar-notes explainer, and changelog |
| 2026-05-27 18:45 | Step 1 verification | Ran `go run ./scripts/snapshot_tool_schemas.go` and `go test ./internal/tools ./internal/toolchecks` successfully |
| 2026-05-27 18:55 | Step 2 verification | Ran `go test ./internal/prompts ./internal/tools` successfully |

---

## Blockers

- 2026-05-27: Dependency TP-109 is not complete (`taskplane-tasks/TP-109-description-only-workout-safety-warning/STATUS.md` shows task status `🔵 Ready for Execution` and Step 0 not started), so warning terminology cannot be aligned safely yet.
- 2026-05-27: Rechecked TP-109 during iteration 2; it remains not started, so TP-111 cannot complete the `Dependencies satisfied` preflight checkbox or proceed to wording updates without risking terminology drift from TP-109.
- 2026-05-27: Supervisor instructed that TP-109 was intentionally skipped in wave 1 and is non-blocking for this batch; proceed with current-state wording-only implementation, using TP-109 PROMPT.md only for terminology intent and avoiding references to nonexistent TP-109 behavior.

---

## Notes

- R001 plan wording pattern to apply:
  - `add_or_update_event`: `description` replaces the upstream event description/DSL; omit on updates to leave unchanged. For `WORKOUT` updates, supplying `description` without `workout_doc` can replace existing structured steps; include the desired `workout_doc` to preserve/merge structure, using `<!-- icuvisor:steps -->` to position serialized steps when also supplying prose.
  - `create_workout`: creation builds the initial upstream template description/DSL from `description`, `workout_doc`, or both. It has no existing template to preserve, but `description` is part of the single upstream description/DSL field, not an append-only notes channel.
  - `update_workout`: omitted sparse fields stay unchanged; supplied `description` and/or `workout_doc` replaces the template's upstream description/DSL. To preserve structured steps while changing prose, provide the desired `workout_doc` plus prose/sentinel.
  - `update_activity`: replacement free-text activity description only; do not imply activity descriptions carry planned workout structure. Keep the `set_activity_intervals` pointer.
- R001 scope guard: TP-109 runtime warning/guard behavior is not implemented; TP-111 must avoid references to nonexistent warning/guard/block metadata or behavior.
- Snapshot verification owner: after tool/schema wording edits, regenerate snapshots with `go run ./scripts/snapshot_tool_schemas.go`; targeted verification should include `go test ./internal/tools ./internal/toolchecks`.
| 2026-05-27 18:19 | Review R001 | plan Step 1: REVISE |
| 2026-05-27 18:21 | Review R002 | plan Step 1: APPROVE |
| 2026-05-27 18:25 | Review R003 | plan Step 2: APPROVE |
