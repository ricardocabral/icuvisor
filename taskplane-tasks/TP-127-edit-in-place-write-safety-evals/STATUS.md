# TP-127: Edit-in-place write safety evals — Status

**Current Step:** Step 1: Audit write/delete guidance
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-29
**Review Level:** 2
**Review Counter:** 4
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

### Step 1: Audit write/delete guidance
**Status:** ✅ Complete

- [x] Inspect create/update/delete workout and event tool descriptions, schemas, and safety tests, including `create_workout` as the unsafe recreate side.
- [x] Inspect registration-time delete gating coverage in `internal/safety/adversarial_test.go` and decide whether `go test ./internal/safety` is needed.
- [x] Identify whether existing descriptions already prefer update/edit in place and where eval coverage is missing.
- [x] Record the current safety contract and any token-budget tradeoff in STATUS.md Discoveries.
- [x] Run targeted tests: `go test ./internal/tools`
- [x] Record exact Step 1 test evidence and safety-test rationale requested by R003.

---

### Step 2: Add eval/adversarial coverage
**Status:** ⬜ Not Started

- [ ] Add at least one eval/adversarial scenario where the user asks to change tomorrow’s workout and the assistant must choose update/edit tools, not delete/create.
- [ ] Assert safe-mode/delete-mode messaging remains short and actionable when deletion is unavailable.
- [ ] Run targeted tests: `make eval-validate` and `go test ./internal/tools`

---

### Step 3: Harden guidance if necessary
**Status:** ⬜ Not Started

- [ ] Update concise tool descriptions or cookbook prompts only where tests show ambiguity.
- [ ] Do not add a model-controlled `confirm` flag or weaken registration-time gating.
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
| Existing write guidance already distinguishes in-place updates from destructive deletes: `add_or_update_event` updates when `event_id` is present and says it never deletes; `update_workout` is sparse by `workout_id`; delete tools are registered only in full delete mode and expose no `confirm`. | Preserve and add eval coverage before changing wording. | `internal/tools/add_or_update_event.go`, `internal/tools/update_workout.go`, `internal/tools/delete_event.go`, `internal/tools/delete_workout.go` |
| `create_workout` currently describes initial template creation but does not explicitly warn against using create to modify an existing template; this is a concise-description tradeoff to avoid bloating tool metadata. | Consider a terse warning only if Step 2 eval shows ambiguity. | `internal/tools/create_workout.go` |
| Existing eval/adversarial coverage covers safe-mode delete unavailability and no-confirm schemas, but cookbook scenarios do not cover changing tomorrow's existing workout/event in place. `go test ./internal/safety` was run because the audit relies on registration-time gating; result: pass (`ok github.com/ricardocabral/icuvisor/internal/safety (cached)`). | Add explicit edit-in-place scenario in Step 2. | `docs/safety/adversarial-prompts.md`, `scripts/eval/scenarios/cookbook_scenarios.json`, `internal/safety/adversarial_test.go` |

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-05-29 | Task staged | PROMPT.md and STATUS.md created |
| 2026-05-29 14:06 | Task started | Runtime V2 lane-runner execution |
| 2026-05-29 14:06 | Step 0 started | Preflight |
| 2026-05-29 14:06 | Step 1 safety audit | Inspected write/delete tool descriptions and `internal/safety/adversarial_test.go`; ran `go test ./internal/safety` -> pass (`ok`, cached). |
| 2026-05-29 14:06 | Step 1 targeted tests | Ran `go test ./internal/tools` -> pass (`ok`, 0.329s). |

---

## Blockers

*None*

---

## Notes

- Step 1 plan review R001 requested explicit `create_workout` audit coverage and registration-time delete gating coverage.
- Step 1 code review R003 requested exact test evidence in STATUS; its status-complete request is deferred until code review APPROVE per Review Level 2 protocol.

| 2026-05-29 14:10 | Review R001 | plan Step 1: REVISE |
| 2026-05-29 14:11 | Review R002 | plan Step 1: APPROVE |
| 2026-05-29 14:15 | Review R003 | code Step 1: REVISE |
| 2026-05-29 14:18 | Review R004 | code Step 1: APPROVE |
