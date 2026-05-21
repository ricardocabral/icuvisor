# Plan Review — TP-077 Step 4

Verdict: **revise**

## Blocking findings

1. **Live MCP environment is underspecified.**
   `create_workout` and `delete_workout` are full-tier workout-library tools, and `delete_workout` is additionally registered only in `ICUVISOR_DELETE_MODE=full`. The Step 4 plan should explicitly run the stdio MCP re-validation with the `.env-dev` test-athlete credentials plus command-local/non-logged settings such as:
   - `ICUVISOR_TOOLSET=full`
   - `ICUVISOR_DELETE_MODE=full` for the cleanup session, or an explicit alternate cleanup route if full-mode delete is unavailable.

   Without this, the live smoke can fail because the needed tools are absent from the catalog, and cleanup may be impossible.

2. **Cleanup needs a failure-safe path.**
   The plan says to delete the created workout, but it should require recording the created workout ID in a non-committed scratch variable/note and cleaning it up even if the read verification fails. If `delete_workout` is unavailable or fails, the worker must use the already-approved direct API/UI cleanup path from Step 1 or mark Step 4 blocked rather than leaving a synthetic workout behind.

## Required amendments

- Run automated validation first and record exact pass/fail evidence in `STATUS.md` for:
  - `make build`
  - `make test`
  - `make test-race`
  - `make lint`
- For the live stdio MCP smoke, use only the `.env-dev` test athlete, avoid printing/copying API keys, and keep raw workout/folder/athlete IDs out of committed notes.
- Use a unique synthetic TP-077 workout name, create it in the real folder identified in Step 1, then verify both:
  - `get_workout_library` shows the workout under the expected library tree/folder, and
  - `get_workouts_in_folder` for that folder returns the new workout.
- After deletion, verify both reads no longer return the workout. Record only sanitized/redacted evidence in `STATUS.md`.

With those amendments, Step 4 will cover the acceptance criteria without risking an unclean test-athlete state.
