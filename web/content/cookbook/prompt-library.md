---
title: "Prompt library"
description: "Short, single-message prompts for everyday training questions, grouped by task."
weight: 15
---

Single-message prompts you can copy, adjust, and send. Every one names your data, gives a window, and tells the assistant not to guess — see [how to get reliable answers]({{< relref "_index" >}}). For multi-step jobs, use a [recipe]({{< relref "_index" >}}) instead. If you are setting up Claude from device-provider data for the first time, start with the [Garmin to Claude walkthrough]({{< relref "../tutorials/garmin-to-claude" >}}). If you repeat the same guardrails in every Claude chat, add the reusable [Claude Project instructions]({{< relref "../guides/claude-project-instructions" >}}).

Replace bracketed placeholders like `[DATE]` or `[ACTIVITY]` before sending.

## Training reviews

```text
Using my intervals.icu data, summarize my last 7 days of training: total load, time, intensity mix, and the two most notable sessions. Do not invent metrics that aren't available.
```

```text
Compare my last 14 days of training with the 14 days before that. What changed in load, volume, and intensity distribution? Tell me which icuvisor tool each number came from.
```

```text
Give me a month-end review for [MONTH]: load trend, polarization of my intensity, longest sessions, and one thing to keep doing and one to change.
```

## Recovery and wellness

```text
How's today looking? Use my intervals.icu data for today only: current CTL/ATL/TSB, wellness, completed activities, planned events, and any NOTE or race annotations. Keep it concise, call out missing/stale data instead of guessing, and end with one practical recommendation for the rest of the day.
```

```text
Look at my wellness for the last 14 days. Summarize sleep, HRV, and resting heart rate trends. Use the correct scales (sleep quality is 1-4) and flag any clear downward trend.
```

```text
Am I recovered enough for a hard interval session today? Use my recent wellness and my current fitness, fatigue, and form. If today's wellness hasn't synced yet, say so instead of guessing.
```

```text
Over the last 60 days, did my sleep duration and training load move together? State plainly if the relationship is weak, and do not infer causation.
```

## Single-activity analysis

```text
Find my most recent ride in intervals.icu and explain what happened in it: sport, duration, distance, load, intensity, and any notes. If it was imported from Strava and fields are blank, tell me.
```

```text
Analyze activity [ACTIVITY]: break down the intervals or laps, time in each zone, and whether the hard efforts held their targets.
```

```text
For activity [ACTIVITY], tell me which advanced metrics are actually available — decoupling, IF, VI, RPE, feel — and report only those. Do not compute fields the server does not expose.
```

## Fitness, form, and trends

```text
What are my CTL, ATL, and TSB right now, and how have they moved over the last 6 weeks? Explain what the current form number means for the next few days.
```

```text
Is my training load trending up, down, or flat over the last 28 days? Say whether the change is meaningful or within normal noise.
```

```text
Project my fitness if I hold my current weekly load for the next 8 weeks with a recovery week every fourth week. Where do CTL and TSB land?
```

## FTP, zones, and best efforts

```text
What are my best efforts and power curve signals over the last 90 days? Compare sprint, threshold, and endurance durations and state the units.
```

```text
Have my best efforts improved versus the prior 90 days? Break it down by duration bucket and show where I actually gained.
```

```text
From my ramp test on [DATE], read out my FTP, threshold heart rate, and any VO2max estimate. Use only what's in that activity's data.
```

## Planning and calendar

```text
Show my planned events for the next 14 days. Flag any week with no easy day and any back-to-back hard sessions.
```

```text
Compare what I completed this week against what was planned. Where did I fall short or overshoot, and what is one practical adjustment? Do not change my calendar.
```

```text
What's on my calendar between [DATE] and [DATE]? Group by category and total the planned load.
```

## Workouts

```text
List the workouts in my intervals.icu library with a one-line summary of each. Group them by folder.
```

```text
Draft a VO2max bike workout: 5 x 4 min at 110% FTP with 4 min recovery, plus warm-up and cool-down. Show it in intervals.icu workout syntax before saving anything.
```

```text
Schedule the workout [WORKOUT] from my library onto [DATE]. Confirm the date and load with me before writing.
```

## Racing

```text
I have a race on [RACE_DATE]. Review my fitness trend and the planned sessions around race week, then suggest a taper outline. Do not write or delete events.
```

```text
Give me a post-race retrospective for activity [ACTIVITY]: pacing, decoupling, time in zone, and two lessons for the next race of this type.
```

## Nutrition and fueling

```text
For my last 7 rides, compare carbs ingested per hour with each session's load. Fetch them in one get_activities call over the last 30 days, read carbs_ingested_g and load from that payload, and flag underfueled sessions. Where carbs weren't logged, say so — don't estimate.
```

```text
Review my logged nutrition wellness fields for the last 14 days alongside training load and flag days where intake looked low for the work done.
```

## Coach mode

```text
List the athletes on my roster. For each, give a one-line status: recent load, current form, and any wellness red flag.
```

```text
Triage athlete [ATHLETE_ID] for this week: recent activities, upcoming planned work, fitness trend, and wellness flags. End with one concise coaching note.
```

## Data checks and troubleshooting

```text
Which of my last 10 activities are missing power, heart rate, or duration data? Give one table row per activity, and identify Strava imports by quoting the source field. If any rows come back truncated, fetch those before answering.
```

```text
List the advanced icuvisor capabilities available in this session so I know which analyses you can run.
```
