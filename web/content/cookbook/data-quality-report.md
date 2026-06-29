---
title: "Data quality report"
description: "Diagnose why an assistant cannot see enough activity, stream, load, wellness, threshold, or calendar data."
weight: 25
---

When an answer feels thin or the assistant says it cannot see a metric, first ask icuvisor for a data-quality report. This read-only report explains whether the issue is sparse history, Strava-restricted activity imports, missing stream metadata, HR/TRIMP-only load, stale wellness sync, missing sport thresholds or zones, or missing race/calendar context.

## When to use this

- "Why can't you see my data?" or "Why did that answer ignore my workout details?"
- Before a coaching prompt that depends on streams, thresholds, wellness, or race calendar context.
- After reconnecting intervals.icu, changing device sync, or importing historical activities.

## The recipe

```text
Use icuvisor to explain why you may not be able to see enough of my data.

Run get_data_quality_report for 2026-01-01 through 2026-01-31. If I name a sport,
set the sport filter to that sport; otherwise inspect all visible data.

Then:
- Summarize the overall data-quality status in one paragraph.
- List each warning or critical diagnostic by code and plain-language meaning.
- Tell me the next action for each issue, naming the icuvisor follow-up tool when one is returned.
- Do not fetch raw activity streams unless the report specifically recommends checking one activity ID.
- Do not guess missing values; say whether the problem is restricted source data, missing streams,
  HR/TRIMP-only load, stale wellness, missing thresholds/zones, sparse activity history, or calendar/race data.
```

## What icuvisor does

| Step | Tool | Why |
| --- | --- | --- |
| 1 | [`get_data_quality_report`]({{< relref "/reference/tools#get_data_quality_report" >}}) | Reads bounded activity, fitness/load, wellness, profile, and calendar probes and returns actionable diagnostics. |
| 2 | Follow-up tools from report recommendations | Use only the targeted tool named by each diagnostic, such as `get_activities`, `get_activity_streams`, `get_training_summary`, `get_wellness_data`, `get_athlete_profile`, or `get_events`. |

## A good answer looks like

> Data-quality status: warning. icuvisor can see recent activities, but stream-dependent and load-dependent answers may be incomplete.
>
> - `restricted_source`: 5 Strava-imported activities look like restricted summaries. Re-import those workouts directly from Garmin/Wahoo/Coros or use available summary fields only.
> - `missing_stream_metadata`: visible activities do not advertise stream channels. For a specific workout, inspect that activity with `get_activity_streams` only if stream-level analysis is required.
> - `trimp_or_hr_load_available`: load appears HR/TRIMP-derived. Treat it as neutral training load, not as power-based TSS.
> - `stale_wellness`: the latest wellness row is older than the report end date. Refresh the wellness bridge or choose a window ending on the latest synced row.
>
> I should not estimate hidden watts, HRV, or race context from those gaps.

## Variations

- "Run this for only Ride data from the last 90 days."
- "Before my race-week review, check whether you can see race events and recent wellness."
- "Explain which data gaps I should fix first before asking for a training-plan review."

## Why this prompt works

- **One bounded preflight.** The report checks common visibility failures without downloading raw stream samples by default.
- **Actionable diagnostics.** Each diagnostic includes a code, severity, evidence, and next action so the assistant can explain the cause instead of apologizing vaguely.
- **No guessing.** The prompt tells the assistant to name the data gap and stop rather than fabricate hidden activity, wellness, load, or calendar values.
