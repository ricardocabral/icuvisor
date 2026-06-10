# Plan Review: Step 1

Verdict: approve with one required clarification before implementation.

The step scope is appropriate: add only a golden DSL/structured pair, document it in `internal/workoutdoc/testdata/README.md`, and run `go test ./internal/workoutdoc` before deciding whether Step 2 is needed.

Required clarification: be careful with the requested “blank/group boundary.” The existing golden harness asserts exact `Serialize(Parse(dsl)) == dsl`, while `Serialize` currently does not emit interior blank lines. If the new `*-dsl.txt` includes an actual blank line before `Cooldown`, Step 1 may fail for canonical formatting rather than proving the cooldown nesting regression. Either:

- make the golden DSL canonical and use the top-level dedent (`- Cooldown ...`) as the group boundary, or
- explicitly treat an actual blank-line fixture as an expected failing regression that will require Step 2 changes to serializer/test expectations.

Implementation checks for the fixture:

- Use a new numbered `*-repeat*cooldown*-dsl.txt` / `*-structured.json` pair.
- Ensure `Cooldown` is unindented/top-level in the DSL.
- Ensure the structured JSON has three top-level siblings: warmup, `Main Set` repeat with `reps: 3`, and final cooldown. The cooldown must not appear in the repeat’s `steps` array.
- Keep README wording clean-room and behavior-focused; do not reference or copy external implementation details.
- Record the targeted test outcome in `STATUS.md`; if the new golden fails, proceed to Step 2 rather than broadening Step 1.
