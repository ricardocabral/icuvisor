---
title: "Build and schedule workouts"
description: "Create structured workouts in correct intervals.icu syntax and put them on the calendar."
weight: 70
---

AI assistants are good at designing a workout and bad at writing it in valid intervals.icu syntax — wrong repeat notation, bullets instead of steps, targets in the wrong units. This recipe makes the assistant check the syntax reference and existing examples first, then show you the workout before saving.

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
2. Check the intervals.icu workout-syntax reference. Use get_workout_library to
   choose the relevant folder, then use get_workouts_in_folder for that folder
   and read one or two terse examples. Keep include_full off unless you need the
   full source for a specific template.
3. Draft this workout: [DESCRIBE IT — e.g. VO2max bike session, 5x4min at
   110% FTP with 4min recoveries, plus warm-up and cool-down]. Before saving,
   show a proposed-change preview with the total duration, key steps, target
   intensities, planned load/distance/time changes, and anything being preserved.
4. Ask for approval of the exact preview. Only after I approve, either save it
   to my workout library or schedule it as a calendar event on [DATE]. Tell me
   which you did and report the planned load.

Rules: get the DSL right — correct repeat syntax, steps not bullets, targets
as percentages of my threshold. For zone targets, rely on icuvisor's structured
`workout_doc` serializer so planned-workout DSL uses my sport settings and adds
metric suffixes like `Z2 Power`, `Z2 HR`, or `Z2 Pace` when needed. Use
structured `workout_doc` only for endurance workouts supported by the
intervals.icu DSL. If I ask for gym or strength work, schedule a simple `NOTE`
time block or a free-text supported calendar event;
do not invent exercises, sets, reps, or loads as structured workout steps unless
my intervals.icu account exposes documented strength-training support. Do not
dump the whole workout library into the chat: pick the relevant folder first,
sample only one or two examples, and use `include_full:true` only after selecting
a specific template that needs raw source detail. Do not overwrite an existing
library workout; create a new one unless I explicitly name one to update. For
multiple calendar or library writes, validate one representative workout, write
one, read it back, check warnings and
structured-step summaries, then continue with the rest.
```

## What icuvisor does

| Step | Tool / resource | Why |
| --- | --- | --- |
| 1 | [`get_athlete_profile`]({{< relref "/reference/tools#get_athlete_profile" >}}) | Targets resolve to your actual FTP and zones. |
| 2 | `icuvisor://workout-syntax` resource + [`get_workout_library`]({{< relref "/reference/tools#get_workout_library" >}}) / [`get_workouts_in_folder`]({{< relref "/reference/tools#get_workouts_in_folder" >}}) | The DSL spec plus folder-scoped, terse examples to copy the format from without flooding the chat. |
| 3 | — | Drafting happens in the chat; nothing is written yet. |
| 4 | [`create_workout`]({{< relref "/reference/tools#create_workout" >}}) or [`add_or_update_event`]({{< relref "/reference/tools#add_or_update_event" >}}) | Saves to the library or schedules it — gated on write mode. |

For manual calendar writes that may be retried, ask the assistant to pass a stable `external_id` to `add_or_update_event`, such as a namespace you own plus the workout/date. Intervals stores that value upstream and icuvisor exposes it in event reads for auditability, so do not put secrets or credential-like values in it. Blank `external_id` values are ignored rather than used to clear an existing upstream value, and IDs are best-effort idempotency aids alongside icuvisor's same-day preflight rather than a global dedupe guarantee. Use your own namespace and avoid provider-owned prefixes such as `strava-` or `hevy-`.

To revise an existing template, name it and the assistant uses [`update_workout`]({{< relref "/reference/tools#update_workout" >}}). On updates, supplied `description` text is prose in the same upstream description/DSL field; it replaces that field rather than appending a note. `workout_doc` is the structured step plan that icuvisor serializes into intervals.icu workout syntax. If the existing template has structured steps you want to keep, ask the assistant to include the desired `workout_doc` explicitly and use `<!-- icuvisor:steps -->` to place the serialized steps around any prose. When `update_workout` changes `workout_doc` for a sport where zone targets are ambiguous, include the template `sport` in the same request so icuvisor can apply the athlete's sport priority settings. When the structure or DSL is uncertain, have the assistant run [`validate_workout`]({{< relref "/reference/tools#validate_workout" >}}) first and use the returned canonical DSL plus estimated duration in the preview. For bulk edits, avoid parallel writes until one representative readback confirms `_meta.workout_doc_warning` is absent or understood and `workout_doc_summary` still shows the expected steps.

## A good answer looks like

> Here is the VO2max session before I save anything:
>
> ```text
> Warmup 15m ramp 45-75% FTP
> 5x
>   4m 110% FTP
>   4m 50% FTP
> Cooldown 10m 50% FTP
> ```
>
> Preview: total duration 53 min; key steps are a 15 min progressive warm-up, 5 × 4 min at 110% FTP (≈310 W for your 282 W FTP) with 4 min recoveries at 50% FTP, and a 10 min cool-down. Planned load is ~78. This is a new workout, so no existing template fields are being overwritten or preserved.
>
> Say "save to library" and I will create it with `create_workout`, or "schedule for Tuesday" and I will place it on your calendar with `add_or_update_event`. I will not write anything until you approve this exact preview.

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
- **From a description:** paste a coach's text endurance workout and ask for it converted to valid intervals.icu syntax.
- **Gym time block:** "Put 45 minutes of gym strength and mobility on Friday as a calendar note; keep the exercise details in free text and do not create structured sets."

## Why this prompt works

- **Read the syntax reference first.** The `icuvisor://workout-syntax` resource is the authoritative DSL spec. Forcing the assistant to consult it — plus one or two folder-filtered library examples — is what fixes the broken repeat/bullet syntax users hit without wasting context on unrelated templates.
- **Draft, then save.** Showing the workout in the chat before any write means you catch a wrong target before it lands on your calendar.
- **"Create, don't overwrite."** Without this, an assistant may update the closest-matching library workout. The explicit rule protects your existing templates; for any intentional bulk or template edit, read and retain structured steps explicitly instead of sending description-only prose.

{{< callout type="warning" >}}
`create_workout`, `update_workout`, and `add_or_update_event` only run when the server is in write mode. In read-only mode the assistant should still draft the workout and show you the syntax to paste into intervals.icu yourself. See [safety modes]({{< relref "/reference/safety-modes" >}}).
{{< /callout >}}
