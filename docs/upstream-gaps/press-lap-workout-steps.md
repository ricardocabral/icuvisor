# Press-lap workout steps: public-evidence contract

## Decision

**Decision: `unverified`.**

As of 2026-07-24, icuvisor has no public, portable, round-trippable evidence that
an intervals.icu workout description can encode “press lap when ready” (or an
equivalent device-control instruction). The decision rule is conservative: this
request is `supported` only when a public upstream grammar/schema names the
control and a returned-document example proves that it survives a write/read
round trip; it is `unsupported` only when upstream explicitly says the control
is unavailable; otherwise it is `unverified`. The current evidence satisfies
neither positive condition.

This is a documentation and validation boundary, not a device feature. icuvisor
must not add a `press_lap`, `lap_button`, `manual_lap`, or similar structured
field, DSL token, serializer branch, or device-specific writer based on an
assistant request or a vendor convention.

## Reproducible evidence record

### Public sources consulted

The following are the public upstream documentation sources to consult for a
future re-check. They were identified from the repository's existing upstream
API references and compared with the repository-held OpenAPI baseline on
2026-07-24; this task made no authenticated or live API request.

| Source | Area consulted | Observed evidence | Retrieval/provenance |
| --- | --- | --- | --- |
| <https://intervals.icu/api-docs.html> | Public API documentation landing page; workout/event API documentation pointer | The public API surface is the authority for endpoint and payload contracts; this task found no published press-lap control contract in the repository's captured evidence. | URL recorded in `docs/upstream-gaps/periodization-parameters.md`; checked against local references on 2026-07-24. No network retrieval in this task. |
| <https://intervals.icu/api/v1/docs> | OpenAPI `Workout` schema and workout/event operations | The repository-held OpenAPI baseline describes `description` as a string and `workout_doc` as a generic object; it does not define a portable step field or control token for manual/press-lap behavior. | Local baseline `scripts/openapidiff/baseline/intervals-openapi.json`, revision `b61b6e6b431bb49473f5222cd761e29f68aa6892`, inspected 2026-07-24. The URL was not fetched during this task. |

The source URLs alone are not evidence of a device capability. Any future
re-check must record the retrieval date, relevant endpoint/schema path, and a
sanitized response or documentation excerpt rather than treating an absent
field in an old local snapshot as proof of upstream behavior.

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

Negative evidence is equally important: the public schema evidence does not
name a press-lap/manual-lap field or a portable device-control token, and the
available repository-held returned-document evidence contains no such control
field or public round-trip example. A model-controlled invented JSON key must
therefore fail strict structured input validation rather than be silently
converted into a step or ordinary duration semantics. A prose fallback may be
stored as prose only when the user explicitly wants a note; it must never be
presented as device-compatible control.

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
