# TP-130: Token-safe workout library positioning — Status

**Current Step:** Step 1: Audit workout-library response shape
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-29
**Review Level:** 1
**Review Counter:** 1
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

### Step 1: Audit workout-library response shape
**Status:** 🟨 In Progress

- [x] Inspect workout-library tools/tests for pagination, terse default, and folder scoping.
- [x] Record whether existing tests protect against huge raw payloads and `include_full` behavior.
- [x] Run targeted tests: `go test ./internal/tools`

---

### Step 2: Add docs/eval hardening
**Status:** ⬜ Not Started

- [ ] Update build-workouts cookbook to recommend folder-filtered and paginated library queries instead of dumping all templates.
- [ ] Add or update tests only if audit finds missing token-safety coverage.
- [ ] Mention the local/token-safe advantage without naming or disparaging competitors.
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
| Workout-library audit found `get_workout_library` returns folders by default and only fetches top-level workouts on opt-in; `get_workouts_in_folder` requires `folder_id`, filters server results client-side, and uses `include_full` for raw docs. No explicit page-size/page-token pagination exists for folder workouts; token safety currently relies on folder scoping and terse shaping. | Inform Step 2 docs/test hardening. | `internal/tools/get_workout_library.go`; `internal/tools/get_workouts_in_folder.go`; `internal/tools/get_workout_library_test.go` |
| Existing tests verify `get_workout_library` does not expose raw `workout_doc`, does not fetch workouts by default, and `get_workouts_in_folder` hides `workout_doc`/description unless `include_full:true`; they do not include a large-payload regression fixture proving many raw docs stay hidden. | Add focused test in Step 2. | `internal/tools/get_workout_library_test.go` |

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-05-29 | Task staged | PROMPT.md and STATUS.md created |
| 2026-05-29 15:46 | Task started | Runtime V2 lane-runner execution |
| 2026-05-29 15:46 | Step 0 started | Preflight |

---

## Blockers

*None*

---

## Notes

*Reserved for execution notes*
| 2026-05-29 15:49 | Review R001 | plan Step 1: APPROVE |
