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

`07-repeat-trailing-cooldown-*` covers the regression where a named repeat main
set is followed by a top-level cooldown. The de-indented cooldown line is a
sibling after the repeat block, not a child that repeats with the main set.

`09-cross-sport-target-boundaries-*` locks absolute run and swim pace ranges,
metric and yard pool distances, duration-based mile pace, explicit percent-FTP
and HR ranges, and recovery/upper power-zone boundaries without deriving any athlete threshold.

`10-mixed-prose-description.txt` is the prose/header/comment envelope used by
the local validation, submit, re-fetch, and re-parse fixture flow. Its sentinel
is replaced by structured steps; every other byte is preserved.

The remaining UI smoke for issue #25 is to create the equivalent workout in the
Intervals.icu web editor, fetch its `description` and `workout_doc` with
`include_full:true`, and replace or add a fixture if the UI emits a different
DSL that preserves repeats/RPE.

`09-cross-sport-upstream-response.json` is a hand-authored local contract fixture,
not a live API capture. It uses the upstream read shape (numeric meter distances
and `text` labels), independently of the local serializer/parser. The tool flow
compares submitted DSL against the golden text, reads this fixture through
`get_workouts_in_folder`, and re-validates its unchanged description. Its mixed
sport targets exercise syntax only; this is not a prescribed training session.
No fixture in this flow calls upstream or proves current upstream acceptance.
