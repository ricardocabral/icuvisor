# R008 Plan Review — Step 3: `get_activities.go` cleanups

Verdict: **REVISE**

The Step 3 checklist is too optimistic for the current tree. It says to confirm `stringSet` has no callers and delete it, but a current grep shows it is not dead:

- `internal/tools/get_activities.go:408` — `activitiesAfterCursor`
- `internal/tools/get_activities.go:489` — `advanceCursorPast`
- `internal/tools/get_activity_streams.go:169` — `shapeActivityStreams`

There is also an unrelated `stringSet` helper in `internal/toolchecks/schema_stability.go`, so the prompt's literal acceptance grep (`grep -rn "stringSet" internal/`) cannot return no hits without touching out-of-scope code or narrowing the grep.

Please revise the plan before implementation:

1. Do not simply delete `internal/tools/get_activities.go`'s `stringSet`; that would break `internal/tools` compilation.
2. Decide and document the minimal behavior-preserving replacement for the current callers if the goal is still to remove the helper name from `internal/tools`:
   - e.g. inline a small local set in `activitiesAfterCursor` and `shapeActivityStreams`, and use `slices.Contains` or equivalent explicit logic in `advanceCursorPast`; or
   - keep a helper under a more specific name if the task owner agrees this is not the requested dead-code deletion.
3. Clarify the acceptance grep in `STATUS.md`: either scope it to `internal/tools/` / the `get_activities.go` helper, or explicitly include a tiny rename of the unrelated `internal/toolchecks` helper if the task owner wants zero `stringSet` symbols in all of `internal/`.
4. The named struct part is fine: define a private type near `decodeGetActivitiesRequest` / `validateActivitiesTokenArgs` with the exact same JSON tags, reuse it for the `supplied` variable and the function signature, and avoid changing error ordering or validation behavior.

After this revision, Step 3 should still remain small, but the current plan would fail the build and has an impossible acceptance check as written.
