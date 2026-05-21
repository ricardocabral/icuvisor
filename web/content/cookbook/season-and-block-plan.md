---
title: "Season and block plan"
description: "Build a periodized training plan toward a goal event, in three reviewable stages."
weight: 60
---

Asking an assistant to "make me a training plan" in one shot produces a generic plan and often overruns the context window. This recipe splits the job into three stages — assess, design, schedule — so each step is grounded in your real data and you approve the plan before anything touches your calendar.

## When to use this

- When you have a goal event and a runway of 8+ weeks.
- At the start of a season, to lay out base/build/peak/taper blocks.
- After a goal event, to plan the next macrocycle.

## The recipe

Send the stages as **separate messages**. Wait for each before sending the next.

### Stage 1 — assess

```text
Stage 1 of planning my season. Use icuvisor with my intervals.icu data.
Read my athlete profile, my fitness trend (CTL / ATL / TSB) for the last
8 weeks, my training-load summary for that period, and my upcoming events.
Summarize my current fitness, my recent weekly load range, and how much
training history you can actually see. Do not propose a plan yet.
```

### Stage 2 — design

```text
Stage 2. My goal is [GOAL EVENT] on [DATE]. It demands [e.g. 3-hour road
race, hilly]. Design a periodized plan from now to race day:
- Name each block (base, build, peak, taper), its length, and its purpose.
- Give a weekly training-load target and ramp rate per block.
- Place recovery weeks (every 3rd or 4th week).
- State a target CTL for race day.
Use get_fitness_projection to check the CTL path is realistic given my
current fitness and the ramp you chose. Present the plan as a week-by-week
table. Do not write anything to my calendar yet.
```

### Stage 3 — schedule

```text
Stage 3. The plan looks good. Add it to my calendar one week at a time:
create the block markers and the key sessions as events, show me each week
before moving to the next, and stop if I say so.
```

## What icuvisor does

| Stage | Tools |
| --- | --- |
| Assess | [`get_athlete_profile`]({{< relref "/reference/tools#get_athlete_profile" >}}), [`get_fitness`]({{< relref "/reference/tools#get_fitness" >}}), [`get_training_summary`]({{< relref "/reference/tools#get_training_summary" >}}), [`get_events`]({{< relref "/reference/tools#get_events" >}}), [`get_training_plan`]({{< relref "/reference/tools#get_training_plan" >}}) |
| Design | [`get_fitness_projection`]({{< relref "/reference/tools#get_fitness_projection" >}}) — simulates CTL/ATL/TSB under your proposed ramp and recovery cadence |
| Schedule | [`add_or_update_event`]({{< relref "/reference/tools#add_or_update_event" >}}) — one event per session, gated on write mode |

## A good answer looks like

Stage 2 should produce something like:

> **Plan: 12 weeks to [GOAL EVENT].** Current CTL 58; race-day target CTL ~72.
>
> | Block | Weeks | Focus | Weekly load | Ramp |
> | --- | --- | --- | --- | --- |
> | Base 2 | 1-3 | Aerobic volume, tempo | 480 → 520 | +6%/wk |
> | Recovery | 4 | Absorb | 300 | -40% |
> | Build 1 | 5-7 | Threshold | 540 → 590 | +5%/wk |
> | Recovery | 8 | Absorb | 340 | -40% |
> | Build 2 | 9-11 | VO2 + race-specific | 560 → 600 | +4%/wk |
> | Taper | 12 | Sharpen | 320 → race | -45% |
>
> `get_fitness_projection` confirms this lands CTL at 71 and TSB at +12 on race morning — within target. Pushing the Build ramps above 6%/wk would overshoot ATL and leave TSB negative on race day.

## Variations

- **No calendar writes:** stop after Stage 2 and apply the plan in intervals.icu yourself. Useful when the server runs read-only.
- **Apply a library plan:** if a structured plan already exists, ask the assistant to use [`apply_training_plan`]({{< relref "/reference/tools#apply_training_plan" >}}) instead of authoring one.
- **Re-plan mid-season:** "I missed two weeks to illness — re-assess and adjust the remaining blocks."

## Why this prompt works

- **Three stages, three messages.** Keeps each tool burst small and gives you a checkpoint before any write — the opposite of a single mega-prompt that gets interrupted.
- **`get_fitness_projection` as a reality check.** The assistant proposes a ramp; the tool tests whether the CTL path is actually reachable, so the plan is not just plausible prose.
- **Schedule last, week by week.** Calendar writes are [gated]({{< relref "/reference/safety-modes" >}}) and hard to undo in bulk. Incremental scheduling keeps you in control.
