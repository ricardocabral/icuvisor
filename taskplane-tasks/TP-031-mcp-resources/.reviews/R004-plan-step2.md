# Plan Review R004 — Step 2: `icuvisor://workout-syntax`

**Verdict: REVISE**

I read `PROMPT.md`, the current `STATUS.md`, the Step 1 reviews, and the existing `internal/resources`, `internal/mcp`, and `internal/workoutdoc` code. The current Step 2 entry in `STATUS.md` is still only the task checklist, not an implementation plan. Please add a concrete Step 2 plan before coding.

## Required plan additions

1. **Define how the resource is derived from `internal/workoutdoc`.**
   `workoutdoc` currently has an implicit grammar spread across `types.go`, `serialize.go`, and `parse.go`; there is no exported grammar table yet. The plan needs to state what single source of truth will be added or exposed so `icuvisor://workout-syntax` is generated from `workoutdoc` rather than a hand-authored Markdown copy that can drift. Acceptable directions include a `workoutdoc` syntax/spec descriptor used by both docs generation and tests, or a focused refactor that makes serializer target families/unit variants table-driven. A static Markdown string in `internal/resources` is not enough.

2. **Specify the default resource registry wiring.**
   Step 1 added plumbing but no production default registry and `internal/app` still does not pass `ResourceRegistry` to `mcp.NewServer`. Step 2's checklist explicitly says to register `icuvisor://workout-syntax` in the default resource registry. The plan should say whether it will add `resources.NewRegistry`/`DefaultRegistry`, register the workout resource there, and wire `internal/app` to pass it now, or justify deferring app wiring. If app wiring is deferred, the resource will not be visible in normal server runs.

3. **Pin the resource contract.**
   Record the intended metadata and read behavior:
   - URI: `icuvisor://workout-syntax`
   - stable snake_case name, human title, short description
   - MIME type, likely `text/markdown`
   - static/no-network handler that honors context cancellation
   - one text `ResourceContents` item with URI/MIME/text populated

4. **List the serializer features that must be covered.**
   The parity test should cover every supported step/target category from `workoutdoc.Serialize`, including at least:
   - simple duration steps and distance steps (`mtr`, `km`, `mi` canonical output)
   - repeats, with the documented no-nested-repeat limitation
   - free-ride steps
   - ramps using `start`/`end`
   - cadence targets
   - power targets: `%FTP`, watts, power zones, scalar and ranges
   - heart-rate targets: `%HR`, `%LTHR`, bpm, HR zones, scalar and ranges
   - pace targets: percent threshold pace, pace zones, text pace / `PACE` handling as currently supported
   - RPE scalar/range as supported by the serializer

   The plan should explicitly include unsupported/limited cases in the docs, e.g. one primary target per step, ramp cannot use text targets, freeride cannot combine with ramp, and repeat blocks cannot also carry simple-step fields.

5. **Make the coverage parity test meaningful.**
   Please describe the test mechanism, not just that a test exists. It should fail when `workoutdoc` gains a new supported target family/unit/step feature and the resource docs are not updated. At minimum, include table-driven fixtures that serialize representative `workoutdoc.Step` values and assert the generated syntax resource includes the corresponding documented feature/examples. Prefer deriving examples through `workoutdoc.Serialize` rather than copying DSL snippets by hand.

6. **Add deterministic resource-content tests.**
   Because this is a static resource, the step should include a golden/snapshot-style test for the generated Markdown plus a registry/protocol test that `resources/list` exposes `icuvisor://workout-syntax` and `resources/read` returns the generated Markdown with the expected MIME type. Keep the golden generated from the source-of-truth data, not as the source itself.

7. **State the package/file layout.**
   The plan should name the intended files. A likely shape is a focused `internal/resources/workout_syntax.go` plus tests, with any grammar/spec helpers living in `internal/workoutdoc` so the resource does not need to inspect unexported serializer internals. If the plan instead puts resource text generation in `workoutdoc`, explain the boundary.

8. **Update task documentation before implementation.**
   Add the Step 2 plan to `STATUS.md` under Notes, and record any discovered mismatch between the PRD's desired DSL coverage and the current serializer's actual coverage. Do not consult or copy GPL source. README trimming can wait for Step 6, but if this step makes a user-visible resource available in normal runs, note whether `CHANGELOG.md` will be updated now or in the final documentation step.

## Suggested Step 2 scope

A good Step 2 should produce one real, readable static resource and the derivation/parity guard around it. Avoid broad changes to tool descriptions here; that is Step 6. Avoid implementing the other three resources in this step; just make the default registry shape extensible so Steps 3–5 can append their resources cleanly.
