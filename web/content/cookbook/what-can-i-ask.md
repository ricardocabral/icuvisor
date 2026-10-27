---
title: "What can I ask icuvisor?"
description: "Beginner-friendly prompts for asking an AI assistant about your intervals.icu training data."
weight: 10
---

If icuvisor is connected, you do not need to know MCP, JSON, or tool names to get useful answers. Ask in plain language, name your intervals.icu data, give a date window, and tell the assistant not to guess when data is missing.

For deeper templates after these starters, use the [prompt library]({{< relref "prompt-library" >}}) or a full [cookbook recipe]({{< relref "_index" >}}).

## First five prompts to try

Copy one prompt into a fresh chat after you finish setup or connect the hosted connector.

```text
Use icuvisor and my intervals.icu data. What are my current FTP, threshold heart rate, training zones, and athlete timezone? Do not use memory or estimates.
```

```text
Use icuvisor to summarize my last 14 days of training: total time, total distance, training load, intensity mix, and the two most notable sessions. Say when a metric is unavailable.
```

```text
How recovered do I look today? Use icuvisor for today's fitness, wellness, planned events, and completed activities. If today's wellness has not synced, say that.
```

```text
Find my most recent activity with icuvisor and explain it in plain language: sport, duration, distance, training load, intensity, and anything missing because of source restrictions.
```

```text
Show my planned workouts and events for the next 7 days using icuvisor. Resolve dates in my athlete timezone and do not change my calendar.
```

## How prompts become capabilities

You ask for an outcome; the assistant chooses icuvisor capabilities behind the scenes. For example, "summarize the last 14 days" usually maps to training-summary and activity reads, while "am I recovered today?" maps to fitness, wellness, today, and calendar reads. You can ask the assistant to cite the source tool behind key numbers, but you should not have to pick tool names yourself.

Good icuvisor prompts usually include four parts:

1. **Name the connector:** "Use icuvisor" or "using my intervals.icu data".
2. **Give the window:** "today", "last 14 days", "race week", or exact dates.
3. **State the job:** review, compare, plan, troubleshoot, or write a draft.
4. **Set a guardrail:** "Do not guess", "do not change my calendar", or "say when data is unavailable".

If the assistant answers without using icuvisor, interrupt it and say: "Call the icuvisor tool first, then answer from the tool result."

## Everyday athlete prompts

### Training review

```text
Using icuvisor, compare my last 7 days with the 7 days before that. What changed in load, volume, intensity, and recovery risk? Cite the source tool for key numbers.
```

```text
Give me a month-to-date training review from intervals.icu: completed activities, total load, sport mix, hard days, easy days, and one practical adjustment for next week.
```

### Recovery and readiness

```text
Use icuvisor to check my sleep, HRV, resting heart rate, form, and recent load for the last 10 days. Am I trending toward freshness or fatigue? Use the correct scale labels.
```

```text
I planned a hard workout today. Use icuvisor to decide whether the available wellness and fitness data supports doing it, modifying it, or resting. If data is stale, say so.
```

### Single-activity analysis

```text
Use icuvisor to find my latest run or ride and explain what limited the session: pacing, heart rate, power or pace availability, time in zones, and any missing streams.
```

```text
Analyze activity [ACTIVITY_ID] with icuvisor. Summarize intervals or laps, whether efforts held target, and which advanced metrics were actually available.
```

### Planning and calendar

```text
Use icuvisor to review the next 14 days of planned workouts and races. Flag back-to-back hard sessions, missing easy days, and anything that conflicts with my current form. Do not write changes.
```

```text
Draft a simple endurance workout for [SPORT] on [DATE]. Show the intervals.icu workout syntax and ask for approval before saving anything.
```

### Trends, zones, and goals

```text
Use icuvisor to review my best efforts and power, pace, or heart-rate signals over the last 90 days. What looks improved, flat, or missing?
```

```text
Project my fitness to [DATE] if I keep roughly my current weekly load and take every fourth week easier. Explain the assumptions and do not call it a prediction.
```

## Race splits and repeated sessions

These workflows need the `full` toolset. If the tools are unavailable, use the advanced-capability list to confirm that before changing your local configuration or hosted preferences, then start a fresh chat.

```text
Use icuvisor to review kilometre or mile splits for activity [ACTIVITY_ID].
Distinguish exact stream-derived splits from device laps or upstream intervals.
Report coverage and missing channels instead of filling gaps or assuming every
lap is a prescribed workout step. Keep raw stream samples out of the answer.
```

```text
Use icuvisor to find the completed sessions I name: [SESSIONS AND DATES]. Show
their activity IDs, dates, and sports before comparing them. After I confirm
the ordered activity IDs, use compute_workout_progression and keep each delta
attached to its source sessions. Say when sessions are not comparable; do not
infer equivalent workouts from matching names or diagnose fatigue from a
cross-sport heart-rate difference.
```

## Coach prompts

Coach mode lets the server route requests to configured athletes. The `athlete_id` is a selector, not a credential, and the API key stays in server configuration.

```text
Use icuvisor coach mode to list my roster and give one line per athlete: recent load, current form, missing wellness, and the most urgent follow-up.
```

```text
Triage athlete [ATHLETE_ID] for this week with icuvisor: recent activities, upcoming plan, fitness trend, wellness flags, and one concise coaching note.
```

```text
For athlete [ATHLETE_ID], compare planned versus completed work over the last 14 days. Use explicit workout status values, do not treat future workouts as missed, and do not edit the calendar.
```

## Data-check and troubleshooting prompts

```text
Use icuvisor to check my last 10 activities for missing power, heart rate, pace, distance, duration, or streams. Identify Strava imports or restricted-source rows and tell me what I can do next.
```

```text
Ask icuvisor which advanced capabilities are visible in this session and explain, in plain language, what extra questions I can ask if I enable the full toolset.
```

```text
If the data looks stale, use icuvisor_check_server_version and diagnostics guidance to tell me whether I should reconnect tools, start a new chat, or check my local config. Do not ask for my API key.
```

## What to do when the answer looks wrong

- If numbers look old, start a new chat and reconnect/reload tools; clients can cache tool catalogs and old context.
- If streams are missing or an activity is marked as Strava-restricted, icuvisor cannot bypass upstream restrictions. Use native provider imports in intervals.icu when you need full historical streams.
- If the assistant invents unavailable metrics, ask it to rerun the query and report only fields returned by icuvisor.
- If you are stuck at setup or data access, use the [troubleshooting guide]({{< relref "../guides/troubleshooting" >}}).
