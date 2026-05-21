# TP-073-get-workouts-in-folder-terse-description — Status

**Current Step:** Step 5: Close the issue
**Status:** ✅ Complete
**Last Updated:** 2026-05-16
**Review Level:** 1
**Review Counter:** 6
**Iteration:** 3
**Size:** XS
**Closes:** #12

---

### Step 1: Reproduce the verbose default in a test

**Status:** ✅ Complete

- [x] Update the terse-default test fixture with a non-empty `description` so the red test reproduces the verbose default.
- [x] In `TestGetWorkoutsInFolderHidesWorkoutDocByDefault`, add an assertion that `description` is also absent from the terse default row.
- [x] Confirm the test fails on `main`.

Step 1 evidence: `go test ./internal/tools -run TestGetWorkoutsInFolderHidesWorkoutDocByDefault -count=1` fails with `description present by default` before the production fix.

### Step 2: Fix the shaping function

**Status:** ✅ Complete

- [x] In `workoutInFolderToRow()`, only populate the `Description` field when `includeFull` is true.
- [x] Keep `workout_doc_summary` in terse mode if it's already present.

Step 2 evidence: `go test ./internal/tools -run TestGetWorkoutsInFolderHidesWorkoutDocByDefault -count=1` passes and keeps the existing `workout_doc_summary` assertion green.

### Step 3: Tests

**Status:** ✅ Complete

- [x] Update the include-full fixture with a non-empty `description` and assert the exact value is preserved when `include_full: true`.
- [x] Confirm the modified terse test from Step 1 now passes.
- [x] Run `make test` and `make test-race`.

Step 3 evidence: targeted include-full and terse tests pass; `make test` and `make test-race` pass.

### Step 4: Build + lint

**Status:** ✅ Complete

- [x] Run `make build` and `make lint`.

Step 4 evidence: `make build` and `make lint` pass.

### Step 5: Close the issue

**Status:** ✅ Complete

- [x] Update `CHANGELOG.md` under `[Unreleased] → Changed` with the `get_workouts_in_folder` default-shape change.
- [x] Record the `Closes #12` handoff details in `STATUS.md` for the PR body/merge auto-close flow.
- [x] Commit final task changes with a `Closes #12` reference.

Step 5 handoff: committed `3e66c6b` with `Closes #12`. PR body should include `Closes #12`. After merge, verify issue #12 is closed manually if needed.

| 2026-05-16 21:03 | Task started | Runtime V2 lane-runner execution |
| 2026-05-16 21:03 | Step 1 started | Reproduce the verbose default in a test |
| 2026-05-16 21:05 | Review R001 | plan Step 1: REVISE |
| 2026-05-16 21:07 | Review R002 | plan Step 1: APPROVE |
| 2026-05-16 22:15 | Review R004 | plan Step 3: REVISE |

| 2026-05-16 22:25 | Worker iter 1 | done in 4895s, tools: 41 |
| 2026-05-16 22:27 | Review R005 | plan Step 3: APPROVE |
| 2026-05-16 22:30 | Review R006 | plan Step 4: APPROVE |

| 2026-05-16 22:48 | Worker iter 2 | done in 1399s, tools: 39 |

| 2026-05-16 22:50 | Worker iter 3 | done in 121s, tools: 16 |
| 2026-05-16 22:50 | Task complete | .DONE created |