---
title: "Build and schedule workouts"
description: "Create structured workouts in correct intervals.icu syntax and put them on the calendar."
weight: 70
---

AI assistants are good at designing a workout and bad at writing it in valid intervals.icu syntax — wrong repeat notation, bullets instead of steps, targets in the wrong units. This recipe makes the assistant check the syntax reference and existing examples first, then show you the workout before saving.

## When to use this

- When you need a specific session — "5x4min VO2", "a sweet-spot ramp" — built and scheduled.
- To turn a workout idea from elsewhere into a valid intervals.icu structured workout.
- To stock your library with reusable templates.

## The recipe

```text
Build me a structured workout and put it on my calendar. Use icuvisor with
my intervals.icu data.

1. Read my athlete profile so all targets use my zones, FTP, and threshold.
2. Check the intervals.icu workout-syntax reference, and read one or two
   existing library workouts as format examples.
3. Draft this workout: [DESCRIBE IT — e.g. VO2max bike session, 5x4min at
   110% FTP with 4min recoveries, plus warm-up and cool-down]. Show it in
   intervals.icu workout syntax and explain the structure before saving.
4. When I approve, either save it to my workout library or schedule it as a
   calendar event on [DATE]. Tell me which you did and confirm the planned load.

Rules: get the DSL right — correct repeat syntax, steps not bullets, targets
as percentages of my threshold. Do not overwrite an existing library workout;
create a new one unless I explicitly name one to update. For multiple calendar
or library writes, validate one representative workout, write one, read it back,
check warnings and structured-step summaries, then continue with the rest.
```

## What icuvisor does

| Step | Tool / resource | Why |
| --- | --- | --- |
| 1 | [`get_athlete_profile`]({{< relref "/reference/tools#get_athlete_profile" >}}) | Targets resolve to your actual FTP and zones. |
| 2 | `icuvisor://workout-syntax` resource + [`get_workout_library`]({{< relref "/reference/tools#get_workout_library" >}}) / [`get_workouts_in_folder`]({{< relref "/reference/tools#get_workouts_in_folder" >}}) | The DSL spec plus real examples to copy the format from. |
| 3 | — | Drafting happens in the chat; nothing is written yet. |
| 4 | [`create_workout`]({{< relref "/reference/tools#create_workout" >}}) or [`add_or_update_event`]({{< relref "/reference/tools#add_or_update_event" >}}) | Saves to the library or schedules it — gated on write mode. |

To revise an existing template, name it and the assistant uses [`update_workout`]({{< relref "/reference/tools#update_workout" >}}). On updates, supplied `description` text replaces the upstream template description/DSL rather than appending a note. If the existing template has structured steps you want to keep, ask the assistant to include the desired `workout_doc` explicitly and use `<!-- icuvisor:steps -->` to place the serialized steps around any prose. For bulk edits, avoid parallel writes until one representative readback confirms `_meta.workout_doc_warning` is absent or understood and `workout_doc_summary` still shows the expected steps.

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
> Structure: 15 min progressive warm-up, then 5 work intervals of 4 min at 110% FTP (≈310 W for your 282 W FTP) with equal 4 min recoveries, then a 10 min cool-down. Estimated duration 53 min, estimated load ~78.
>
> Say "save to library" and I will create it with `create_workout`, or "schedule for Tuesday" and I will place it on your calendar with `add_or_update_event`. I will not overwrite anything that already exists.

## Variations

- **Library only:** "...just save it to my 'VO2' folder, do not schedule it."
- **A week of workouts:** describe 3-4 sessions and ask for them scheduled across named days — still review each before the writes.
- **From a description:** paste a coach's text workout and ask for it converted to valid intervals.icu syntax.

## Why this prompt works

- **Read the syntax reference first.** The `icuvisor://workout-syntax` resource is the authoritative DSL spec. Forcing the assistant to consult it — plus real library examples — is what fixes the broken repeat/bullet syntax users hit.
- **Draft, then save.** Showing the workout in the chat before any write means you catch a wrong target before it lands on your calendar.
- **"Create, don't overwrite."** Without this, an assistant may update the closest-matching library workout. The explicit rule protects your existing templates; for any intentional bulk or template edit, read and retain structured steps explicitly instead of sending description-only prose.

{{< callout type="warning" >}}
`create_workout`, `update_workout`, and `add_or_update_event` only run when the server is in write mode. In read-only mode the assistant should still draft the workout and show you the syntax to paste into intervals.icu yourself. See [safety modes]({{< relref "/reference/safety-modes" >}}).
{{< /callout >}}
