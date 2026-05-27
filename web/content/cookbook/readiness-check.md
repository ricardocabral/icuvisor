---
title: "Readiness check"
description: "Decide whether to train hard today using wellness and form, with correct scales and staleness handling."
weight: 30
---

Before a hard session, you want a straight answer: train as planned, modify, or back off. This recipe makes the assistant weigh wellness against current fatigue and give a green/amber/red call — while handling the two things AI assistants get wrong most often: subjective scales and not-yet-synced data. To make those scale and staleness rules stick across Claude chats, add the [Claude Project instructions]({{< relref "../guides/claude-project-instructions" >}}).

## When to use this

- In the morning, before a planned interval or long session.
- When TSB is deep negative and you are unsure whether that is productive fatigue or a warning.
- Any day you "feel off" and want the data to confirm or contradict it.

## The recipe

```text
Act as my coach and tell me whether I should train hard today. Use icuvisor
with my intervals.icu data.

1. Read my athlete profile.
2. Pull my wellness for the last 14 days: sleep duration, sleep quality, HRV,
   resting heart rate, fatigue, soreness, mood, and stress.
3. Pull my current fitness, fatigue, and form (CTL / ATL / TSB).

Then:
- Give today's readiness in one word — green, amber, or red — and why.
- Name the one or two signals driving that call.
- Recommend whether to keep, modify, or move today's planned hard session.

Rules: sleep quality is a 1-4 scale and feel is a 1-5 scale — use the labels
icuvisor returns, do not rescale them to 0-10. If today's wellness row has not
synced yet (common with a timezone offset), tell me the latest data is stale
rather than guessing today's values. Do not invent HRV or sleep numbers.
```

## What icuvisor does

| Step | Tool | Why |
| --- | --- | --- |
| 1 | [`get_athlete_profile`]({{< relref "/reference/tools#get_athlete_profile" >}}) | Confirms units and which wellness fields you track. |
| 2 | [`get_wellness_data`]({{< relref "/reference/tools#get_wellness_data" >}}) | Sleep, HRV, resting HR, and subjective scales over the window. |
| 3 | [`get_fitness`]({{< relref "/reference/tools#get_fitness" >}}) | CTL / ATL / TSB — the fatigue side of the picture. |

`analyze_correlation` can be added if you want to know whether a wellness metric tracks load over a longer window.

## A good answer looks like

> **Readiness: amber.**
>
> Your last fully synced wellness row is yesterday — today's has not arrived yet, so this call uses yesterday plus the 14-day trend (`get_wellness_data`).
>
> **Driving signals:** HRV is 48 ms, down from a 14-day average of 58 ms and the lowest in that window. Sleep last night was 6h 10m at quality 2/4 (below your usual 3/4). Resting HR is +4 bpm.
>
> **Form:** TSB is -14 (`get_fitness`) — you are carrying real fatigue.
>
> **Recommendation:** Modify today's session. A suppressed HRV plus short, poor sleep on top of a -14 TSB is not the day for VO2 work. Do the planned session as endurance (Z2), or move the intervals to tomorrow and reassess once today's wellness syncs.

## Variations

- **Pre-race morning:** "...tell me whether I am ready to race today" and ask for the one thing to manage during the event.
- **Specific session:** paste the planned workout — "I have 4x8min at threshold planned; is today the day?"
- **Trend only:** "Just show me the 14-day HRV, resting HR, and sleep trends with no recommendation."

## Why this prompt works

- **Explicit scale reminder.** Assistants routinely treat sleep quality as out of 10 and call a 3/4 "poor". Stating the scale stops a good night being read as a bad one.
- **Staleness instruction.** intervals.icu wellness often lags by a day because of timezone offsets. Telling the assistant to flag stale data prevents it from inventing today's HRV.
- **One-word verdict first.** Forcing green/amber/red keeps the answer decision-shaped instead of a hedge.

{{< callout type="info" >}}
The `recovery_check` [MCP prompt]({{< relref "/reference/resources-prompts" >}}) runs this workflow with the scale and staleness guardrails enforced server-side.
{{< /callout >}}
