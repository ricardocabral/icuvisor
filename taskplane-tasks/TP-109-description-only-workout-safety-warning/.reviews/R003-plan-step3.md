# Plan Review R003 — Step 3

Verdict: Request changes.

The Step 3 verification checklist has the right baseline commands (`make test`, `make lint`, `make build`) and the requirement to fix or document failures. However, before treating the full-suite run as sufficient, the plan should explicitly verify the Step 1/R002 trigger matrix that was approved for the new warning.

Blocking gap:

- Add a focused verification item for `update_workout` where `description` is supplied as `null` and `workout_doc` is omitted. The approved contract said this tool should key off field presence (`descriptionProvided && !workoutDocProvided`) so explicit null/clear semantics remain covered if decoding accepts them. A plain full-suite run will not catch this unless the targeted regression exists.

Recommended Step 3 additions:

- Re-run/extend the targeted `internal/tools` tests to cover the previously approved warning matrix, especially `update_workout` with `{"workout_id":"...","description":null}`.
- Then run the planned full commands and record exact outcomes in `STATUS.md`: `make test`, `make lint` (or document unavailable tool), and `make build`.

Once that focused regression is included, the rest of the Step 3 plan is sufficient.
