# Plan Review: Step 2 — Update prompt/docs wording

**Verdict:** Approved.

The Step 2 plan targets the right surfaces: the weekly-planning prompt and golden file, the workout/notes docs, `CHANGELOG.md`, and prompt/tool targeted tests. It also stays within the task's wording-only scope and does not depend on unimplemented TP-109 runtime warning behavior.

## Non-blocking guidance

- In `weekly_planning`, avoid the current implication that `description` is just a safe notes attachment. Keep the useful “`description` and `workout_doc` coexist” guidance, but add the replacement caveat: for updates, include the desired `workout_doc` whenever preserving structured steps matters.
- In `build-workouts.md`, add the update semantics near the “revise an existing template” / “create, don't overwrite” guidance: supplied descriptions replace the upstream description/DSL, so bulk or template edits should read/retain the structured steps explicitly instead of using description-only prose.
- In `calendar-notes.md`, keep NOTE guidance simple: `description` is the plain-text note body, and on updates it replaces that body; `workout_doc` remains only for WORKOUT structure.
- If `internal/prompts/catalog.go` changes, update `internal/prompts/testdata/weekly_planning.md` in the same edit so the golden test failure is intentional and reviewable.
- Record the user-visible wording clarification under `[Unreleased]` → `Changed`, adjacent to the existing description/workout_doc merge entry if possible.

Run and record `go test ./internal/prompts ./internal/tools` after edits.
