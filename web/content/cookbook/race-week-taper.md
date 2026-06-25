---
title: "Race-week taper"
description: "Plan the final days before a goal event — load, intensity, and target race-day form."
weight: 80
---

The taper is where good fitness is either sharpened or wasted. This recipe makes the assistant ground the taper in your real fitness trajectory and race-week calendar, then hand you a day-by-day outline — without touching your events. For a Claude Project dedicated to race prep, add the [race-week Project instruction block]({{< relref "../guides/claude-project-instructions#optional-block-race-week-taper" >}}) so the no-write and timezone rules are always present.

## When to use this

- 7-14 days out from a goal event.
- When you are unsure how much to cut and want a TSB target for race morning.
- To sanity-check a taper you have already drafted.

## The recipe

```text
I have a [RACE TYPE — e.g. hilly road race, ~3h] on [RACE_DATE]. Plan my
taper. Use icuvisor with my intervals.icu data.

1. Read my athlete profile and resolve any relative dates, countdowns, or
   weekday/date pairings with `resolve_calendar_dates` in my athlete timezone.
2. Read my calendar events around race week to confirm the race and what is
   already scheduled.
3. Pull my fitness trend (CTL / ATL / TSB) and training-load summary for
   the last 6 weeks.
4. Pull my recent activities and my recent wellness.
5. Project where form (TSB) lands on race day under a reduced taper load.

Then give me:
- A day-by-day taper outline from today to race day: load, intensity, and
  the key sharpening sessions.
- The target TSB for race morning, and whether my current trajectory
  reaches it.
- Wellness signals to watch this week.

Rules: do not write or delete calendar events — give me the outline to apply
myself. Do not invent fitness numbers; if the window is short on data, say so.
```

## What icuvisor does

| Step | Tool | Why |
| --- | --- | --- |
| 1 | [`get_athlete_profile`]({{< relref "/reference/tools#get_athlete_profile" >}}), [`resolve_calendar_dates`]({{< relref "/reference/tools#resolve_calendar_dates" >}}) | Pins today, countdowns, and weekday/date pairings to the athlete-local calendar before planning. |
| 2 | [`get_events`]({{< relref "/reference/tools#get_events" >}}) | Confirms the race and race-week calendar. |
| 3 | [`get_fitness`]({{< relref "/reference/tools#get_fitness" >}}), [`get_training_summary`]({{< relref "/reference/tools#get_training_summary" >}}) | CTL/ATL/TSB and recent load to taper from. |
| 4 | [`get_activities`]({{< relref "/reference/tools#get_activities" >}}), [`get_wellness_data`]({{< relref "/reference/tools#get_wellness_data" >}}) | Freshness and recovery signals. |
| 5 | [`get_fitness_projection`]({{< relref "/reference/tools#get_fitness_projection" >}}) | Simulates race-day TSB under the proposed taper load. |

## A good answer looks like

> **Taper to [RACE_DATE] — 8 days out.** Current CTL 71, ATL 84, TSB -13 (`get_fitness`).
>
> **Race-morning target:** TSB +15 to +20 for a 3-hour race. `get_fitness_projection` shows that holding ~45% of your recent weekly load this week lands TSB at +17 — on target. Cutting harder risks going flat; cutting less leaves you racing tired.
>
> | Day | Session | Load |
> | --- | --- | --- |
> | -8 Sat | Endurance 2h, 2x10min tempo | 95 |
> | -7 Sun | Easy 1h | 35 |
> | -6 Mon | Rest | 0 |
> | -5 Tue | 75min, 4x3min at race pace | 65 |
> | -4 Wed | Easy 45min | 28 |
> | -3 Thu | 50min, 3x90s openers | 38 |
> | -2 Fri | Rest or 30min spin | 0-20 |
> | -1 Sat | 40min, 3x1min openers | 30 |
>
> **Watch this week:** resting HR and HRV should drift up as fatigue clears — a *drop* mid-week means the taper is too light on rest, not too heavy. Sleep is your biggest lever now.
>
> Apply this in intervals.icu yourself — I have not changed any events.

## Variations

- **Short taper:** "...I can only taper for 5 days" — ask for a compressed outline.
- **Multi-day event:** "...this is a 3-day stage race" — ask for taper plus inter-stage recovery notes.
- **Already drafted:** paste your taper and ask the assistant to critique it against your fitness trend.

## Why this prompt works

- **Date-anchored.** `resolve_calendar_dates` prevents stale-chat, UTC, or client-time date math from shifting the taper to the wrong day.
- **Projection-anchored.** `get_fitness_projection` turns "rest up" into a specific load percentage that hits a TSB number — testable, not vibes.
- **No-write rule.** Race week is the worst time for an accidental calendar edit. Explicitly forbidding writes keeps the assistant advisory.
- **Wellness watch.** Adding the freshness signals to monitor makes the taper adaptive instead of a fixed script.

{{< callout type="info" >}}
The `race_week_taper` [MCP prompt]({{< relref "/reference/resources-prompts" >}}) runs this workflow and validates the race date before starting; use `resolve_calendar_dates` first when the date came from "today," a countdown, or a weekday phrase.
{{< /callout >}}
