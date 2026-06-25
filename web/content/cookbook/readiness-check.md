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
2. Pull my wellness for the last 14 days: Intervals readiness when present,
   sleep duration, sleep quality, sleep score, HRV, resting heart rate, fatigue,
   soreness, stress, feel, mood, motivation, and any `_native` provider fields.
3. Pull my current fitness, fatigue, and form (CTL / ATL / TSB).
4. If a hard session is planned today, use `get_today` to read planned-event
   `workout_status`, tags, the planned `indoor` flag, weather
   availability/provenance, and any completed warm-up activity tags before
   judging context.

Then:
- Give today's training-readiness guidance in one word — green, amber, or red — and why.
- If the Intervals readiness field is missing or null, say that before using fallback signals.
- Name the one or two signals driving that call.
- Recommend whether to keep, modify, move, or execute indoors/outdoors today, naming the planned-event `workout_status`, any relevant tags, `indoor` flag, weather availability, or activity tags returned by icuvisor.

Rules: sleep quality is a 1-4 scale, sleep score is 0-100 when device-imported,
and feel is a 1-5 scale — use the labels icuvisor returns, do not rescale them
to 0-10. If readiness is missing, do not invent or backfill a readiness score
from Garmin Body Battery or any other native/supporting field; treat HRV,
resting HR, sleep, subjective scales, and `_native` fields as cautious evidence.
If today's wellness row has not synced yet (common with a timezone offset), tell
me the latest data is stale rather than guessing today's values. Do not invent
HRV or sleep numbers. Use weather only when `get_today.weather` or completed
activity weather fields provide it with provenance; if `get_today.weather.status`
is `forecast_unavailable`, say that weather is unavailable from icuvisor rather
than inventing conditions. If you suggest doing the workout indoors instead of
outdoors, keep it as advice or a preview unless I approve a calendar edit; do not
create a second active workout for the same session.
```

## What icuvisor does

| Step | Tool | Why |
| --- | --- | --- |
| 1 | [`get_athlete_profile`]({{< relref "/reference/tools#get_athlete_profile" >}}) | Confirms units and which wellness fields you track. |
| 2 | [`get_wellness_data`]({{< relref "/reference/tools#get_wellness_data" >}}) | Readiness when present; sleep, HRV, resting HR, subjective scales, missing fields, provenance, and provider `_native` fields over the window. |
| 3 | [`get_fitness`]({{< relref "/reference/tools#get_fitness" >}}) | CTL / ATL / TSB — the fatigue side of the picture. |
| 4 | [`get_today`]({{< relref "/reference/tools#get_today" >}}) | Optional context for today's planned workout status, `indoor` flag, weather availability/provenance, and completed activity tags. |

`analyze_correlation` can be added if you want to know whether a wellness metric tracks load over a longer window.

## A good answer looks like

> **Readiness guidance: amber.**
>
> Intervals readiness is missing for today (`_meta.missing_fields` includes `readiness`), so I am not treating Garmin/native fields as a replacement score. This call uses supporting wellness signals plus current form.
>
> **Driving signals:** HRV is 48 ms, down from a 14-day average of 58 ms and the lowest in that window. Sleep last night was 6h 10m at quality 2/4 (sleepScore 72/100). Resting HR is +4 bpm, and `_native.garmin` includes Body Battery min/max on its own 0-100 provider scale.
>
> **Form:** TSB is -14 (`get_fitness`) — you are carrying real fatigue.
>
> **Recommendation:** Modify today's session. A suppressed HRV plus short, poor sleep on top of a -14 TSB is not the day for VO2 work. Do the planned session as endurance (Z2), or move the intervals to tomorrow and reassess once today's wellness syncs.

## Variations

- **Pre-race morning:** "...tell me whether I am ready to race today" and ask for the one thing to manage during the event.
- **Specific session:** paste the planned workout — "I have 4x8min at threshold planned; is today the day?"
- **Indoor alternative:** "If the outdoor session looks risky from available weather or logistics, suggest an indoor trainer version, but do not change my calendar unless I approve it."
- **Trend only:** "Just show me the 14-day HRV, resting HR, and sleep trends with no recommendation."

## Why this prompt works

- **Explicit scale reminder.** Assistants routinely treat sleep quality as out of 10 and call a 3/4 "poor". Stating the scale stops a good night being read as a bad one.
- **Missing-readiness fallback.** Garmin and other providers can expose useful HRV, resting HR, sleep, subjective, and `_native` wellness fields even when Intervals readiness itself is null. The prompt tells the assistant to explain that absence first and use fallback fields cautiously instead of creating a synthetic score.
- **Staleness instruction.** intervals.icu wellness often lags by a day because of timezone offsets. Telling the assistant to flag stale data prevents it from inventing today's HRV.
- **One-word verdict first.** Forcing green/amber/red keeps the answer decision-shaped instead of a hedge.
- **Weather provenance and one active workout.** `get_today` says whether weather is sourced or unavailable, so the assistant can recommend an indoor alternative without making up a forecast. Keeping the alternative as a preview avoids double-counting planned load by placing both outdoor and indoor versions on the active calendar.

{{< callout type="info" >}}
The `recovery_check` [MCP prompt]({{< relref "/reference/resources-prompts" >}}) runs this workflow with the scale and staleness guardrails enforced server-side.
{{< /callout >}}
