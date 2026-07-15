---
title: "Build and schedule workouts"
description: "Create structured workouts in correct intervals.icu syntax and put them on the calendar."
weight: 70
---

AI assistants are good at designing a workout and bad at writing it in valid intervals.icu syntax — wrong repeat notation, bullets instead of steps, targets in the wrong units. The portable path below uses structured input and a read-only preflight, so it works even when an MCP client lists Resources but does not deliver their contents to the model. `icuvisor://workout-syntax` remains an authoritative reference when its contents are available.

## When to use this

- When you need a specific endurance session — "5x4min VO2", "a sweet-spot ramp" — built and scheduled.
- To turn an endurance workout idea from elsewhere into a valid intervals.icu structured workout.
- To stock your library with reusable templates.
- When you only need a gym/strength time block or note on the calendar, not detailed structured strength sets.

## The recipe

```text
Build me a structured workout and put it on my calendar. Use icuvisor with
my intervals.icu data.

1. Read my athlete profile so all targets use my zones, FTP, and threshold.
2. If this host has delivered the contents of `icuvisor://workout-syntax` to
you, you may use it as the authoritative syntax reference. A listed Resource,
its URI, or `resources/list` is not evidence that you can read it. Whether or
not its contents are available, use the portable structured-tool path below.
3. Draft this workout as a structured `workout_doc`: [DESCRIBE IT — e.g.
VO2max bike session, 5x4min at 110% FTP with 4min recoveries, plus warm-up and
cool-down]. Put duration and distance in structured fields, and targets in
structured power, pace, heart-rate, or RPE fields — not only in prose. Call
`validate_workout` with that `workout_doc`. Continue only when `valid: true`;
resolve every returned error or warning that changes the intended workout.
Use the returned `canonical_dsl` and estimated duration in the proposed-change
preview, with total duration, key steps, target intensities, planned
load/distance/time changes, and anything being preserved. If adapting an
existing planned session for indoor vs outdoor execution, read `get_today`
first and use only its weather availability/provenance, planned `indoor` flag,
tags, equipment context I provide, and athlete preference.
4. Ask for approval of the exact preview. Only after I approve, either save it
to my workout library or schedule it as a calendar event on [DATE]. Tell me
which you did and report the planned load.
5. Inspect the returned write evidence before claiming the workout rendered as
structured steps: for `create_workout`, check
`workout.workout_doc_summary`; for `add_or_update_event`, check
`event.workout_doc_summary`; in both responses check
`_meta.workout_doc_warning`. An upload marker, prose, or canonical DSL alone
does not verify structured rendering. A warning must be absent or understood,
and the summary must show the expected structure.

Rules: use structured `workout_doc` only for endurance workouts supported by
the intervals.icu DSL. If I ask for gym or strength work, schedule a simple
`NOTE` time block or a free-text supported calendar event; do not invent
exercises, sets, reps, or loads as structured workout steps unless my
intervals.icu account exposes documented strength-training support. Do not dump
the whole workout library into the chat: pick the relevant folder first, sample
only one or two examples, and use `include_full:true` only after selecting a
specific template that needs raw source detail. Do not overwrite an existing
library workout; create a new one unless I explicitly name one to update. For
multiple calendar or library writes, validate one representative workout, write
one, inspect its returned summary and warning, then continue with the rest. For
indoor/outdoor alternatives to the same planned session, keep only one active
calendar workout unless I explicitly ask for both; prefer editing/replacing the
existing event after approval so planned load is not double-counted.
```

## What icuvisor does

| Step | Tool / resource | Why |
| --- | --- | --- |
| 1 | [`get_athlete_profile`]({{< relref "/reference/tools#get_athlete_profile" >}}) | Targets resolve to your actual FTP and zones. |
| 2 | Optional `icuvisor://workout-syntax` resource | Use its authoritative DSL reference only when the host has made its contents readable to the model; Resource registration alone is not enough. |
| 3 | [`validate_workout`]({{< relref "/reference/tools#validate_workout" >}}) | Validates a structured `workout_doc` without network access and returns the canonical DSL and duration for the preview. |
| 4 | [`create_workout`]({{< relref "/reference/tools#create_workout" >}}) or [`add_or_update_event`]({{< relref "/reference/tools#add_or_update_event" >}}) | Saves to the library or schedules it — gated on write mode and approval. |
| 5 | Returned write response | Confirms the returned structured-step summary and surfaces any fidelity warning. |

A compact portable `workout_doc` for the VO2max session looks like this. It is
structured input, not a DSL cheat-sheet, and is sufficient even when the
Resource cannot be read:

```json
{
  "steps": [
    {"description": "Warm up", "duration": 900, "power": {"value": 60, "units": "PERCENT_FTP"}},
    {"reps": 5, "steps": [
      {"duration": 240, "power": {"value": 110, "units": "PERCENT_FTP"}},
      {"description": "Recovery", "duration": 240, "power": {"value": 50, "units": "PERCENT_FTP"}}
    ]},
    {"description": "Cool down", "duration": 600, "power": {"value": 50, "units": "PERCENT_FTP"}}
  ]
}
```

For manual calendar writes that may be retried, ask the assistant to pass a stable `external_id` to `add_or_update_event`, such as a namespace you own plus the workout/date. Intervals stores that value upstream and icuvisor exposes it in event reads for auditability, so do not put secrets or credential-like values in it. Blank `external_id` values are ignored rather than used to clear an existing upstream value, and IDs are best-effort idempotency aids alongside icuvisor's same-day preflight rather than a global dedupe guarantee. Use your own namespace and avoid provider-owned prefixes such as `strava-` or `hevy-`.

For an indoor/outdoor adaptation of today's planned workout, ask for a preview first: what changes (route/trainer mode, target basis, duration/load, tags, and the `indoor` flag), what is preserved, and whether weather was sourced or unavailable. Do not ask the assistant to create an indoor clone next to the outdoor plan unless you explicitly want two active workouts and understand the planned-load/CTL double-counting risk.

To revise an existing template, name it and the assistant uses [`update_workout`]({{< relref "/reference/tools#update_workout" >}}). On updates, supplied `description` text is prose in the same upstream description/DSL field; it replaces that field rather than appending a note. `workout_doc` is the structured step plan that icuvisor serializes into intervals.icu workout syntax. If the existing template has structured steps you want to keep, ask the assistant to include the desired `workout_doc` explicitly and use `<!-- icuvisor:steps -->` to place the serialized steps around any prose. When `update_workout` changes `workout_doc` for a sport where zone targets are ambiguous, include the template `sport` in the same request so icuvisor can apply the athlete's sport priority settings and emit device-facing target suffixes instead of prose-only target hints. Run [`validate_workout`]({{< relref "/reference/tools#validate_workout" >}}) before every uncertain structured change and resolve diagnostics before approval. For bulk edits, avoid parallel writes until one representative returned response has the expected `workout_doc_summary` and `_meta.workout_doc_warning` is absent or understood; a warning can mean that intervals.icu did not render the uploaded structure or only partially preserved it.

## A good answer looks like

> I will use the portable structured path. The syntax Resource is optional; I will use it only if this host has actually delivered its contents to me.
>
> First I will send this `workout_doc` to `validate_workout`:
>
> ```json
> {"steps":[{"description":"Warm up","duration":900,"power":{"value":60,"units":"PERCENT_FTP"}},{"reps":5,"steps":[{"duration":240,"power":{"value":110,"units":"PERCENT_FTP"}},{"description":"Recovery","duration":240,"power":{"value":50,"units":"PERCENT_FTP"}}]},{"description":"Cool down","duration":600,"power":{"value":50,"units":"PERCENT_FTP"}}]}
> ```
>
> The preflight returned `valid: true`, a canonical DSL, and an estimated duration of 65 minutes. Preview: a 15 min warm-up, 5 × 4 min at 110% FTP (about 310 W for your 282 W FTP) with 4 min recoveries at 50% FTP, and a 10 min cool-down. Planned load is about 78. This is a new workout, so no existing template fields are being overwritten or preserved.
>
> Say "save to library" and I will create it with `create_workout`, or "schedule for Tuesday" and I will place it on your calendar with `add_or_update_event`. I will not write anything until you approve this exact preview. After the write, I will check the returned summary and any fidelity warning before saying it rendered as a structured workout.

For an existing-template edit, the assistant should make the delta just as explicit:

> Before I call `update_workout`, here is the proposed update to `Sweet Spot Builder`:
>
> - **Current:** 60 min total, 3 × 10 min at 88-92% FTP, 5 min recoveries at 55% FTP, load 68.
> - **Proposed:** 70 min total, 4 × 10 min at 88-92% FTP, same 5 min recoveries at 55% FTP, load about 78.
> - **Delta:** +10 min, +1 sweet-spot interval, +10 load, no distance target change.
> - **Preserved:** title, sport, folder, tags, warm-up/cool-down structure, recovery intensity, and the existing prose note about cadence.
>
> If you approve this exact before/after preview, I will send `update_workout` with the full desired `workout_doc` so the structured steps are preserved instead of replacing them with prose-only text.

## Variations

- **Library only:** "...just save it to my 'VO2' folder, do not schedule it."
- **A week of workouts:** describe 3-4 sessions and ask for them scheduled across named days — still review each before the writes.
- **From a description:** paste a coach's text endurance workout and ask for it converted to structured `workout_doc`, validated before saving.
- **Gym time block:** "Put 45 minutes of gym strength and mobility on Friday as a calendar note; keep the exercise details in free text and do not create structured sets."
- **Indoor option for today:** "Read `get_today`, then show me an indoor trainer version of today's outdoor ride. Do not write it unless I approve replacing the existing event."

## Why this prompt works

- **Portable preflight first.** A structured `workout_doc` plus `validate_workout` is sufficient even if a host lists Resources without exposing their contents to the model. When the Resource contents are available, they remain the authoritative reference rather than a prerequisite for authoring.
- **Draft, validate, then save.** Requiring `valid: true` before approval catches malformed structure before a write; the returned canonical DSL and duration make the preview reviewable.
- **Verify returned structure.** A successful upload or readable DSL does not prove that intervals.icu rendered graphical interval segments. Inspect the returned summary and fidelity warning before claiming success.
- **"Create, don't overwrite."** Without this, an assistant may update the closest-matching library workout. The explicit rule protects your existing templates; for any intentional bulk or template edit, read and retain structured steps explicitly instead of sending description-only prose.
- **One active version.** Indoor and outdoor variants of the same planned session are alternatives, not two workouts to publish by default. Preview the replacement and write only after approval so Intervals.icu does not double-count planned load.

{{< callout type="warning" >}}
`create_workout`, `update_workout`, and `add_or_update_event` only run when the server is in write mode. In read-only mode the assistant should still draft the workout and show you the structured plan to paste into intervals.icu yourself. See [safety modes]({{< relref "/reference/safety-modes" >}}).
{{< /callout >}}
