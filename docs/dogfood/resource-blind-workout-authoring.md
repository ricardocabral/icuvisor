# Resource-blind structured-workout authoring protocol

## Purpose and status

This is a manual test protocol, not a host-compatibility report. It checks
whether an MCP host can reliably author and verify structured workouts when it
lists Resources but does not provide Resource contents to the model. All matrix
results below are intentionally **not run**. Do not infer any host, transport,
or model compatibility from this document until a row contains reproducible
evidence from a safely configured test environment.

`icuvisor://workout-syntax` remains useful when the host actually delivers its
contents to the model. `resources/list`, registration, or a visible URI alone
does not show that the model read the Resource.

## Safety boundary

Do not run this protocol with a real athlete, a production API key, a paid-model
call by default, or a production calendar. If execution is approved, use a
dedicated test athlete and test library folder (or a disposable test calendar
date), write mode limited to that environment, and the smallest permitted model
budget. Record the transport and exact icuvisor version so another operator can
reproduce the result.

The protocol makes no compatibility conclusion unless it captures the observable
inputs and returned output described below. A successful upload marker, prose,
or canonical DSL is not evidence that intervals.icu rendered structured steps.

## Procedure

1. Record the host, model and version, date, icuvisor version, transport,
   toolset, write mode, and test-environment identity in the matrix. Never put a
   credential in this document.
2. Start a fresh conversation in the host. If it lists Resources, deliberately
   do not paste or otherwise deliver the contents of `icuvisor://workout-syntax`
   to the model. Record whether a Resource read was **observable to the model**
   from the host trace; `resources/list` by itself is `not observed`.
3. For one matrix case, supply the exact `workout_doc` below to
   `validate_workout`. Record `valid`, all diagnostics, canonical DSL, and
   estimated duration. Do not write unless `valid: true` and every
   meaning-changing diagnostic is resolved.
4. After an explicit approval in the test conversation, send the same structured
   `workout_doc` in a `create_workout` library write or an
   `add_or_update_event` calendar write. Record the exact write input with
   redacted test-only IDs.
5. Capture the returned evidence: `workout.workout_doc_summary` for a library
   write or `event.workout_doc_summary` for a calendar write, plus
   `_meta.workout_doc_warning`. Compare the summary with the intended steps.
   A missing, unrendered, or partial-fidelity warning is a result to record, not
   a reason to claim success.
6. Keep the row `not run` until all evidence is captured. If it is run, attach
   a redacted transcript or trace location and state only what that row proves;
   do not generalize to another host, model, sport, or Resource mode.

## Exact validation fixtures

Each fixture is an exact `validate_workout` argument. Use it unchanged unless a
host/model failure requires a documented retry. The library write wraps the same
`workout_doc` with a test-only `name`, `sport`, and `folder_id`; the calendar
write wraps it with a test-only athlete-local `date`, `category: "WORKOUT"`,
`type`, and `name`.

### B1 — bike

```json
{"workout_doc":{"steps":[{"description":"Warm up","duration":600,"power":{"value":60,"units":"PERCENT_FTP"}},{"description":"Endurance","duration":1800,"power":{"value":70,"units":"PERCENT_FTP"}}]}}
```

### R1 — run

```json
{"workout_doc":{"steps":[{"description":"Run warmup","duration":600,"pace":{"value":85,"units":"PERCENT_THRESHOLD"}},{"description":"Run steady","duration":1200,"pace":{"value":95,"units":"PERCENT_THRESHOLD"}}]}}
```

### SM1 — metric swim

```json
{"workout_doc":{"steps":[{"description":"Metric swim","distance":{"value":400,"unit":"m"},"pace":{"value":90,"units":"PERCENT_THRESHOLD"}}]}}
```

### SY1 — yard swim

```json
{"workout_doc":{"steps":[{"description":"Yard swim","distance":{"value":400,"unit":"yrd"},"pace":{"value":90,"units":"PERCENT_THRESHOLD"}}]}}
```

### RP1 — repeat

```json
{"workout_doc":{"steps":[{"description":"Main set","reps":3,"steps":[{"duration":120,"power":{"value":110,"units":"PERCENT_FTP"}},{"description":"Recovery","duration":120,"power":{"value":50,"units":"PERCENT_FTP"}}]}]}}
```

### RM1 — ramp

```json
{"workout_doc":{"steps":[{"description":"Build","duration":480,"ramp":true,"power":{"start":70,"end":95,"units":"PERCENT_FTP"}}]}}
```

## Unexecuted host matrix

Every result and evidence cell is `not run` until a safely configured execution
records the required data. “Exact validation/write inputs” refers to the fixture
above plus the redacted complete write request captured during execution.

| Case | Required outcome | Host / model / version | Date, transport, toolset, write mode | Observable Resource-read state to model | Exact validation/write inputs | Validation evidence (`valid`, diagnostics, canonical DSL) | Returned summary evidence | `_meta.workout_doc_warning` | Result / trace |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| B1 bike | Bike structured steps survive validation and returned-summary inspection | not run | not run | not run | not run | not run | not run | not run | not run |
| R1 run | Run pace structured steps survive validation and returned-summary inspection | not run | not run | not run | not run | not run | not run | not run | not run |
| SM1 metric swim | Metric swim distance/pace structured steps survive validation and returned-summary inspection | not run | not run | not run | not run | not run | not run | not run | not run |
| SY1 yard swim | Yard `yrd` distance/pace structured steps survive validation and returned-summary inspection | not run | not run | not run | not run | not run | not run | not run | not run |
| RP1 repeat | Repeat structure survives validation and returned-summary inspection | not run | not run | not run | not run | not run | not run | not run | not run |
| RM1 ramp | Numeric ramp structure survives validation and returned-summary inspection | not run | not run | not run | not run | not run | not run | not run | not run |

## Recording a completed row

A completed row must include all of the following before it can say the one
case rendered correctly:

- host name; model name and version; icuvisor version; date; transport/toolset;
  and test-only write destination;
- whether Resource content delivery or a Resource read was observable to the
  model, not merely whether the host listed the Resource;
- the exact `validate_workout` and write requests, with only test IDs redacted;
- `valid: true` plus recorded diagnostics and canonical DSL before the write;
- the returned `workout.workout_doc_summary` or `event.workout_doc_summary`;
- the exact `_meta.workout_doc_warning` value, including an unrendered or
  partial-fidelity warning when present; and
- a redacted trace or transcript location.

Until those fields exist, leave the outcome `not run`. Do not turn blank rows into a compatibility matrix or publish a host/model compatibility claim.
