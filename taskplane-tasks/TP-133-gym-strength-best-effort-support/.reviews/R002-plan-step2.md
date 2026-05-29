# Plan Review R002 — Step 2

Verdict: Approved

The Step 2 plan is correctly scoped: it updates user-facing cookbook/prompt guidance for best-effort gym scheduling, explicitly avoids new strength write tools, and calls out the targeted prompt test when fixtures change.

Execution notes to keep the step safe:

- If updating the weekly-planning prompt, update the source in `internal/prompts/catalog.go` and regenerate/update `internal/prompts/testdata/weekly_planning.md`; do not edit the golden fixture only. This is a small justified scope expansion from the listed artifacts.
- Keep the prompt terse enough for `internal/prompts` invariants; prefer replacing or tightening existing workout guidance over adding multiple long bullets.
- Guidance should default to `NOTE` time blocks/free-text notes for gym work. Mention `WORKOUT` only for simple duration/name/description scheduling when an upstream-supported activity type is known; do not imply `GYM`/`STRENGTH` categories, structured sets/reps/load, or `workout_doc` strength semantics exist today.
- In `build-workouts.md`, preserve the distinction that the workout DSL is for structured endurance workouts, not first-class strength sessions.
- Run `go test ./internal/prompts` if `catalog.go` or prompt testdata changes; docs-only cookbook edits can document that no targeted Go test was needed.
