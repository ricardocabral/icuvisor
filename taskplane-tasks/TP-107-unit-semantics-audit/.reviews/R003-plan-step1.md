# Plan Review — Step 1 (R003)

Verdict: **APPROVE**

The Step 1 plan now has enough concrete scope to implement safely. It names the table-driven serializer test, states that each case builds a `WorkoutDoc`, calls `Serialize`, and asserts the exact DSL output, and lists the intended power, pace, HR, zone, watt/BPM, and text-pace coverage. It also makes the `MINS_KM`/`MINS_MILE` decision explicit and keeps serializer changes test-driven rather than speculative.

Minor implementation note: when writing the HR table, include all currently supported HR percent aliases from `workoutTargetUnits` (`%HR`/`HR` as well as `PERCENT_HR`/`PERCENT_MAX_HR`) or document why any alias is intentionally omitted. This does not block the plan.

Required verification remains `go test ./internal/workoutdoc`, with `STATUS.md` updated for the behavior proven by the matrix and any unsupported-unit discovery.
