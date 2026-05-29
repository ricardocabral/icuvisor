---
title: "Activity retrospective"
description: "Break down a single ride, run, or race in detail — intervals, splits, zones, and what to learn."
weight: 40
---

After a key session you want more than "nice ride". This recipe makes the assistant pull one activity apart — intervals, splits, time-in-zone, decoupling — and turn it into two takeaways, while being honest about Strava-imported activities that come back nearly empty.

## When to use this

- After an interval session, to check whether the hard parts held their targets.
- After a long endurance ride or run, to see aerobic durability (drift, decoupling).
- As a post-race debrief.

## The recipe

You can name an activity by ID or just describe it — the assistant will look it up. For relative dates like "last Sunday," it should resolve the athlete-local date window with `get_activities`, choose the matching activity, then pass that `activity_id` to detail, interval, and split tools.

```text
Give me a detailed retrospective of my most recent hard ride. Use icuvisor
with my intervals.icu data.

1. If I gave an activity ID, use it. Otherwise list my recent activities or
   the requested athlete-local date window, pick the one I described, and tell
   me which activity ID you chose.
2. Get the activity details: sport, local start time, duration, distance,
   load, and source/device.
3. Get the intervals or laps with get_activity_intervals, and the per-km or
   per-mile splits.
4. Get the time-in-zone for the session.
5. Get the extended metrics, and report only the ones actually present
   (decoupling, IF, VI, normalized power, RPE, feel).

Then give me:
- What kind of session this was and how it went overall.
- A table with one row per work interval: target vs actual power, average
  HR, duration, any interval `custom_fields` that are relevant to the prompt
  (for example manually-entered lactate), and whether it held. If the activity
  has only laps and no structured intervals, say so and summarize the laps instead.
- How the hard parts held up — pacing, fade, heart-rate drift, decoupling.
- Two concrete takeaways for next time.

Rules: if this activity was imported from Strava and its fields are blank,
tell me that up front and analyze only what is actually there. Quote a
decoupling, NP, IF, or VI figure only when a tool returned it, and name the
source tool inline. Do not compute or estimate metrics intervals.icu did not
provide. Keep the answer under about 400 words, leading with the interval table.
```

## What icuvisor does

| Step | Tool                                                                                                                                                                    | Why                                                                                                                                                    |
| ---- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------ |
| 1    | [`get_activities`]({{< relref "/reference/tools#get_activities" >}})                                                                                                    | Finds the activity and its ID when you describe it.                                                                                                    |
| 2    | [`get_activity_details`]({{< relref "/reference/tools#get_activity_details" >}})                                                                                        | Sport, timing, load, device, Strava-import flag.                                                                                                       |
| 3    | [`get_activity_intervals`]({{< relref "/reference/tools#get_activity_intervals" >}}) and [`get_activity_splits`]({{< relref "/reference/tools#get_activity_splits" >}}) | Per-rep and per-distance breakdown. `get_activity_intervals` also includes scalar interval `custom_fields` such as lactate when upstream returns them. |
| 4    | [`get_activity_histogram`]({{< relref "/reference/tools#get_activity_histogram" >}})                                                                                    | Time-in-zone distribution for the session.                                                                                                             |
| 5    | [`get_extended_metrics`]({{< relref "/reference/tools#get_extended_metrics" >}})                                                                                        | Decoupling, IF, VI — only those upstream actually fitted.                                                                                              |

For a specific surge or climb, [`compute_activity_segment_stats`]({{< relref "/reference/tools#compute_activity_segment_stats" >}}) computes mean/NP/decoupling over an explicit time range.

## A good answer looks like

> **Session:** Tuesday VO2 ride, 1h 12m, 38.4 km, load 95, recorded on a Wahoo head unit (`get_activity_details`) — direct upload, full data.
>
> **The work:** 5 x 4 min intervals (`get_activity_intervals`). Targets held well for reps 1-4 (308-314 W, target 310 W) but rep 5 faded to 291 W. Recoveries were on target.
>
> **Aerobic response:** Pw:HR decoupling 6.1% across the work intervals (`get_extended_metrics`) — acceptable for a VO2 session. Time-in-zone was 41% Z5+ during the work blocks (`get_activity_histogram`).
>
> **Takeaways:** (1) The rep-5 fade suggests the session was at the right edge of doable — hold this format rather than adding a sixth rep. (2) Decoupling stayed controlled, so durability is not the limiter here; the ceiling is repeatability at VO2 power.

## Variations

- **Lactate test:** "Analyze this interval session and include the lactate values I entered on each rep. Use `custom_fields.lactate` from `get_activity_intervals` when present."
- **Race debrief:** "...this was a race — focus on pacing discipline and where I lost time."
- **Compare two sessions:** "Compare activity A and activity B — same workout, two weeks apart. Did the hard parts improve?"
- **A Strava import specifically:** "Analyze my most recent Strava-imported activity. Identify genuine Strava imports by the `source` field — a Garmin or Wahoo device means a native upload, not a Strava import — state the blank-field policy first, and confirm the blank payload with `get_activity_details` rather than inferring it from the list."
- **Leave a note:** add "If write tools are enabled, append a one-line summary as a comment on the activity." This uses [`add_activity_message`]({{< relref "/reference/tools#add_activity_message" >}}).

## Why this prompt works

- **Look-it-up step.** Letting the assistant resolve the activity with `get_activities` — including athlete-local date windows for phrases like "last Sunday" — means you never paste raw data or raw IDs, which is what overruns the context window.
- **"Report only the ones actually present."** Extended metrics vary by device and sport. This line stops the assistant inventing a decoupling figure for a run with no power.
- **Strava callout.** Strava-imported activities return blank fields by policy. Naming this up front turns a confusing "your power was 0 W" into an honest "this was a Strava import, so power is unavailable".
