# TP-077-create-workout-refused — Status

**Current Step:** Step 6: Close the issue
**Status:** ✅ Complete
**Last Updated:** 2026-05-17
**Review Level:** 2
**Review Counter:** 15
**Iteration:** 2
**Size:** M
**Closes:** #9
**Requires live API access:** YES (.env-dev test athlete)

---

### Step 1: Live probe to isolate the contract

**Status:** ✅ Complete

- [x] Source `.env-dev`.
- [x] Inventory the test athlete workout library and identify a real writable folder ID, or confirm top-level writes are allowed.
- [x] Probe minimal workout-library create payload permutations out of process without committing scratch files.
- [x] Capture the accepted sanitized request and response fixtures under `internal/intervals/testdata/workout_library/`.
- [x] Delete every probe-created workout and verify cleanup via the workout library.

**Probe notes:** Live API accepted `type: "Ride"` with an existing `folder_id` and accepted the same payload with `description`; it rejected `sport` without `type` (`Missing type`), rejected omitted `folder_id` (`Folder is required`), and rejected top-level create with `folder_id: null` (`Folder is required`). Probe-created workouts were deleted and the library returned to its starting count.

Reviewer revision items:

- [x] Probe and record the omitted-`folder_id` POST result separately from explicit `null`.
- [x] Reduce or synthesize account-derived training settings in the captured create response fixture.

### Step 2: Add a failing test

**Status:** ✅ Complete

- [x] Add/update intervals client coverage for the required non-empty create `folder_id` contract while keeping the accepted outbound body tied to Step 1 fixtures.
- [x] Add tool-boundary validation/schema tests requiring an existing `folder_id`, and update existing happy-path tool tests to include a sanitized folder ID.
- [x] Confirm the focused new required-folder assertions fail before the client/tool fix. Evidence: `go test ./internal/tools ./internal/intervals -run 'CreateWorkout|CreateLibraryWorkout'` fails in `TestCreateWorkoutRejectsBadArguments` and `TestCreateWorkoutRegistrationMetadata` before production changes.

Reviewer revision items:

- [x] Replace intervals missing-folder assertions with a no-network test that fails if a request is made and proves local validation.
- [x] Extend create_workout schema metadata tests so every example/input_example includes a non-blank `folder_id`.
- [x] Add valid sanitized `folder_id` values to bad-argument rows whose purpose is not missing-folder validation.

### Step 3: Fix the client + tool

**Status:** ✅ Complete

- [x] Require a non-empty create `folder_id` in the intervals client and tool validation while keeping sport serialized as JSON `type` and leaving update semantics unchanged.
- [x] Update public `create_workout` validation text plus input schema description, required list, and examples to document that `folder_id` must identify an existing folder owned by the athlete.
- [x] Run focused create-workout tests and ensure the Step 2 failures now pass. Evidence: `go test ./internal/tools ./internal/intervals -run 'CreateWorkout|CreateLibraryWorkout'` passed.

### Step 4: Build + lint + race + live re-validation

**Status:** ✅ Complete

- [x] Run `make build`, `make test`, `make test-race`, and `make lint` successfully. Evidence: all four commands passed.
- [x] Live re-validation via stdio MCP creates a workout in a real folder, confirms it appears in library/folder reads, deletes it, and confirms cleanup. Evidence: stdio MCP `create_workout` returned an ID, `get_workouts_in_folder` found it, `delete_workout` returned deleted, and cleanup re-read confirmed it gone.

### Step 5: Document amendment

**Status:** ✅ Complete

- [x] Document the workout-library create payload contract, including required existing `folder_id`, JSON `type`, and refused omitted/null folder cases.

### Step 6: Close the issue

**Status:** ✅ Complete

- [x] Update `CHANGELOG.md` under `[Unreleased] → Fixed` with the `create_workout` root cause and fix.
- [x] Record final status/PR-body issue-close instructions in `STATUS.md` without closing the issue before merge. PR body must include `Closes #9`; if it does not auto-close after merge, close issue #9 manually.
- [x] Commit the final delivery changes with a TP-077 conventional commit referencing `Closes #9`.

| 2026-05-17 02:43 | Task started | Runtime V2 lane-runner execution |
| 2026-05-17 02:43 | Step 1 started | Live probe to isolate the contract |
| 2026-05-17 02:46 | Review R001 | plan Step 1: APPROVE |

| 2026-05-17 03:23 | Worker iter 1 | done in 2384s, tools: 40 |
| 2026-05-17 03:23 | Step 2 started | Add a failing test |
| 2026-05-17 03:26 | Review R002 | code Step 1: UNKNOWN |
| 2026-05-17 03:30 | Review R003 | code Step 1: APPROVE |
| 2026-05-17 03:32 | Review R004 | plan Step 2: REVISE |
| 2026-05-17 03:34 | Review R005 | plan Step 2: APPROVE |
| 2026-05-17 03:38 | Review R006 | code Step 2: UNKNOWN |
| 2026-05-17 03:42 | Review R007 | code Step 2: UNKNOWN |
| 2026-05-17 03:45 | Review R008 | code Step 2: APPROVE |
| 2026-05-17 03:47 | Review R009 | plan Step 3: REVISE |
| 2026-05-17 03:48 | Review R010 | plan Step 3: APPROVE |
| 2026-05-17 03:52 | Review R011 | code Step 3: APPROVE |
| 2026-05-17 03:54 | Review R012 | plan Step 4: APPROVE |
| 2026-05-17 04:00 | Review R013 | code Step 4: APPROVE |
| 2026-05-17 04:02 | Review R014 | plan Step 5: APPROVE |
| 2026-05-17 04:04 | Review R015 | code Step 5: APPROVE |

| 2026-05-17 04:06 | Worker iter 2 | done in 2608s, tools: 148 |
| 2026-05-17 04:06 | Task complete | .DONE created |