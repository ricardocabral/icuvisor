# Press-lap workout steps: public-evidence contract

## Decision

**Decision: `unverified`.**

As of 2026-07-24, current public upstream evidence was unavailable for this
task because live or authenticated upstream access is prohibited. The local
baseline contains no portable, round-trippable representation for “press lap
when ready,” but that absence is not attributed to the current upstream API.
The decision rule is conservative: this request is `supported` only when a
public upstream grammar/schema names the control and a returned-document example
proves that it survives a write/read round trip; it is `unsupported` only when
upstream explicitly says the control is unavailable; otherwise it is
`unverified`. The current evidence cannot establish either positive condition.

This is a documentation and validation boundary, not a device feature. icuvisor
must not add a `press_lap`, `lap_button`, `manual_lap`, or similar structured
field, DSL token, serializer branch, or device-specific writer based on an
assistant request or a vendor convention.

## Reproducible evidence record

### Public source references (current contents not fetched)

These are the exact public sources identified for a future re-check. They are
recorded here for reproducibility, but are not claimed as consulted current
evidence: task policy prohibited a live request, and no public documentation
snapshot containing a press-lap contract is checked in. A future re-check must
record its retrieval date and the excerpt it observes.

| Source | Exact area to inspect | Evidence available in this task | Retrieval/provenance |
| --- | --- | --- | --- |
| <https://intervals.icu/api-docs.html> | API documentation landing page and linked workout/event operations | No current excerpt was available locally; the URL is referenced by existing repository evidence records. | Reference recorded in `docs/upstream-gaps/periodization-parameters.md`; URL recorded 2026-07-24; not fetched in this task. |
| <https://intervals.icu/api/v1/docs> | `components.schemas.Workout.properties.description`, `components.schemas.Workout.properties.workout_doc`, and workout/event create/read operations | Local snapshot only: `description` is `{"type":"string"}` and `workout_doc` is `{"type":"object","additionalProperties":{"type":"object"}}`; neither local entry defines a press-lap field. | `scripts/openapidiff/baseline/intervals-openapi.json`, revision `b61b6e6b431bb49473f5222cd761e29f68aa6892`, inspected 2026-07-24; URL not fetched. |

The local snapshot is implementation evidence, not current public-upstream
proof. In particular, its missing field cannot establish that the live API
supports or rejects press-lap control. The source URLs alone are not evidence
of a device capability.

### Local implementation baseline (not upstream proof)

The following local files describe what icuvisor can safely represent and test;
they do not establish Garmin, Wahoo, or intervals.icu execution behavior:

- `internal/workoutdoc/types.go`: `Step` has description, duration/distance,
  power/HR/pace/RPE/cadence targets, ramp/freeride, and repeat fields only.
- `internal/workoutdoc/parse.go` and `serialize.go`: the canonical grammar is
  line-oriented: `- [label] [duration|distance] [target]`, with `Nx` repeat
  headers and indented child steps. There is no press-lap token.
- `internal/workoutdoc/syntax.go`: the published local syntax reference lists
  duration/distance, repeats, ramps, freeride, targets, and cadence; it does
  not list device-control steps.
- `internal/workoutdoc/testdata/`: checked-in DSL/structured pairs prove local
  parser/serializer behavior. The `06-full-surface-upstream-response-workout-doc.json`
  file is a sanitized historical capture with documented partial fidelity loss,
  not public proof of press-lap support and not a new fixture for this task.
- `internal/tools/validate_workout.go` and `internal/tools/decode.go`:
  `validate_workout` is read-only and strict-decodes its request; unknown
  top-level arguments are rejected before validation. Structured step fields
  are validated through the same serializer path used for writes.
- `internal/tools/workout_doc_fidelity.go`: write responses expose a returned
  structured-summary/fidelity warning boundary. An upload marker or canonical
  DSL is not proof that upstream rendered the intended structure.
- `docs/prd/PRD-icuvisor.md` §7.2.C (Events & workouts) and §7.4 assumptions:
  the documented write channel is the description-string DSL, with structured
  round-trip and lossy-field requirements. Those product requirements do not
  add a press-lap grammar.

## What the evidence does and does not show

Positive evidence establishes only the existing endurance-workout grammar:
structured steps have a measure (duration or distance), optional supported
training targets, and known repeat/ramp/free-ride forms. A label such as
`Press lap when ready` remains a free-text description label. It does not
become a structured control merely because it appears beside a valid duration
or target, and prose such as `Press lap when ready before the next interval`
is not a verified instruction to a watch or head unit.

Negative evidence is equally important: the local schema snapshot and local
WorkoutDoc types do not name a press-lap/manual-lap field or portable
device-control token, while current public evidence was not available to
confirm whether upstream has added one. The available repository-held
returned-document evidence contains no such control field or public
round-trip example; it is a local historical fixture, not a current upstream
claim. A model-controlled invented JSON key must therefore fail strict
structured input validation rather than be silently converted into a step or
ordinary duration semantics. A prose fallback may be stored as prose only when
the user explicitly wants a note; it must never be presented as
device-compatible control.

Garmin and Wahoo execution behavior is not established by the intervals.icu
DSL. Device workout-upload capabilities, whether a device exposes a manual lap
button during a workout, and whether an uploaded step waits for a lap are
separate vendor/device contracts. No TSS, training-load, duration, or other
load semantics may be inferred from a hypothetical press-lap action. In
particular, a manual lap is not a substitute for a timed or distance step in
load calculations.

## Future supported-branch prerequisites

Before opening a separate implementation task, obtain all of the following
from public upstream documentation or an explicitly approved synthetic/local
fixture exercise:

1. **Grammar/schema provenance:** a named portable field or DSL token, accepted
   spelling, scope, and semantics for waiting on a manual lap.
2. **Fixture provenance:** a sanitized source description and returned
   `workout_doc`/schema fixture that identifies the upstream version/date and
   the source operation; no hand-written candidate fixture counts as evidence.
3. **Local round trip:** `parse → serialize → reparse` equality for the
   candidate, with any documented lossy fields enumerated and tested.
4. **Returned-document evidence:** after an approved write/read probe, the
   returned structured document must preserve the control field; an upload
   marker or prose echo is insufficient. The test must also assert that a
   missing/partial fidelity warning blocks a compatibility claim.
5. **Load semantics:** explicit evidence for estimated duration, TSS/training
   load, distance, and any wait/manual-lap behavior. If upstream does not define
   a value, the implementation must report it as unknown rather than estimate
   it from the control.
6. **Device caveats:** separate, device-specific public evidence for Garmin and
   Wahoo (and any other target), including capability differences and what
   “press lap” means on each. A portable upstream grammar cannot by itself
   establish either vendor's behavior.

Until those prerequisites are met, the safe authoring choices are a supported
timed/distance structured step or an explicitly prose note. Do not modify
`internal/workoutdoc/syntax.go`, add a field/token, or implement a writer in
this task.
