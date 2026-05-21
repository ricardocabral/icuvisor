# TP-059-workout-training-typed-bodies — Status

**Current Step:** Step 3: Verify
**Status:** ✅ Complete
**Last Updated:** 2026-05-16
**Review Level:** 1
**Review Counter:** 3
**Iteration:** 1
**Size:** S

---

### Step 1: Define the typed structs

**Status:** ✅ Complete

- [x] Define the workout-library write request in `internal/intervals` with pointer fields so sparse updates preserve explicit empty values.
- [x] Define shaped tool response summary structs in `internal/tools` for workout docs and training plans with existing JSON keys preserved.
- [x] Replace opaque `full` response fields with an explicit non-map passthrough representation such as `json.RawMessage`.
- [x] Reuse `internal/workoutdoc` only where data is already in canonical workoutdoc form; preserve existing summary behavior for unknown top-level keys.

### Step 2: Swap request/response bodies

**Status:** ✅ Complete

- [x] Replace each remaining `map[string]any` request/response body in the cited files with the typed structs or explicit raw JSON passthrough.
- [x] Update or add round-trip tests proving the request/response wire shape is unchanged.
- [x] `make build` / `make test` / `make test-race` / `make lint`.

### Step 3: Verify

**Status:** ✅ Complete

- [x] `grep -n "map\[string\]any" internal/tools/get_workout_library.go internal/tools/get_training_plan.go internal/intervals/workout_library.go` — zero hits for request/response bodies. (Schema literals exempt.)
- [x] All `make` checks pass.
- [x] Commit: `TP-059 use typed structs for workout & training bodies`.

| 2026-05-16 23:10 | Task started | Runtime V2 lane-runner execution |
| 2026-05-16 23:10 | Step 1 started | Define the typed structs |
| 2026-05-16 23:13 | Review R001 | plan Step 1: UNKNOWN |
| 2026-05-16 23:15 | Review R002 | plan Step 1: APPROVE |
| 2026-05-16 23:23 | Review R003 | plan Step 2: APPROVE |

| 2026-05-16 23:32 | Worker iter 1 | done in 1292s, tools: 113 |
| 2026-05-16 23:32 | Task complete | .DONE created |