# WorkoutDoc fixtures

The numbered `*-dsl.txt` and `*-structured.json` pairs are golden fixtures for
the Intervals.icu workout description DSL emitted by `workoutdoc.Serialize`.

`06-full-surface-upstream-candidate-*` covers issue #25's known risk areas in one
fixture: inline labels, a labelled repeat header with indented child steps,
duration and distance steps, ramps, freeride, power/HR/pace/RPE/cadence targets,
ranges, and zone suffixes.

`06-full-surface-upstream-response-workout-doc.json` is a sanitized live API
capture from posting that candidate DSL to the workout-library endpoint on
2026-06-04. It intentionally keeps only `workout_doc.steps`, omitting calculated
zone/load metrics. The capture proves partial fidelity loss: Intervals.icu
returned structured steps, but flattened the repeat children to the end and
dropped the RPE target from `Strides`.

The remaining UI smoke for issue #25 is to create the equivalent workout in the
Intervals.icu web editor, fetch its `description` and `workout_doc` with
`include_full:true`, and replace or add a fixture if the UI emits a different
DSL that preserves repeats/RPE.
