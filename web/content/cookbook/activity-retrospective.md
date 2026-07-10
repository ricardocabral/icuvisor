---
title: "Activity retrospective"
description: "Break down a single ride, run, or race in detail — intervals, splits, zones, and what to learn."
weight: 40
---

After a key session you want more than "nice ride". This recipe makes the assistant pull one activity apart — intervals, splits, time-in-zone, decoupling — and turn it into two takeaways, while being honest about Strava-imported activities that come back nearly empty.

For logged carbohydrate evidence, missing-log coverage, and source-labelled grams-per-hour calculations without nutrition targets, use the dedicated [Fueling review]({{< relref "fueling-review" >}}) recipe.

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
   load, tags, activity fueling (`carbs_ingested_g`, `carbs_used_g`),
   and source/device. If I ask about an activity custom field such as VO2Max,
   request its field code explicitly with `custom_fields` instead of assuming
   it is present in default activity rows.
3. Get the intervals or laps with get_activity_intervals, check
   `_meta.interval_source`, `_meta.auto_lap_suspected`, and
   `_meta.interval_source_caveat`, and get the per-km or per-mile splits.
4. Get the time-in-zone for the session.
5. Get the extended metrics, and report only the ones actually present
   (decoupling, IF, VI, normalized power, RPE, feel).
6. If `get_activities`/`get_activity_details` returns `hypoxic_training_caveat`,
   or `get_extended_metrics` returns `_meta.hypoxic_training_caveat`, quote its
   provenance and wording. If I explicitly said the session used an altitude
   tent/chamber or reduced oxygen exposure, mention that CTL/ATL/Form still use
   logged `training_load` and do not add a hypoxia multiplier.

Then give me:
- What kind of session this was and how it went overall.
- A table with one row per work interval: target vs actual power, average
  HR, duration, any interval `custom_fields` that are relevant to the prompt
  (for example manually-entered lactate), and whether it held. If the activity
  has only laps and no structured intervals, say so and summarize the laps instead;
  if there is one unknown/collapsed interval row, say it is not proof of no
  intervals and use `compute_activity_segment_stats` over explicit time or
  distance bounds before judging sprint or anaerobic execution.
- How the hard parts held up — pacing, fade, heart-rate drift, decoupling.
- Two concrete takeaways for next time.

Rules: if this activity was imported from Strava and its fields are blank,
tell me that up front and analyze only what is actually there. Quote a
decoupling, NP, IF, or VI figure only when a tool returned it, and name the
source tool inline. Do not compute or estimate metrics intervals.icu did not
provide. For hypoxic training, require explicit provenance from the user or from
activity name/notes/tags/custom fields; do not treat altitude, elevation gain, or
SpO2 alone as proof. Keep the answer under about 400 words, leading with the
interval table.
```

## What icuvisor does

| Step | Tool                                                                                                                                                                    | Why                                                                                                                                                    |
| ---- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------ |
| 1    | [`get_activities`]({{< relref "/reference/tools#get_activities" >}})                                                                                                    | Finds the activity and its ID when you describe it; terse rows include upstream tags when intervals.icu returns them.                                  |
| 2    | [`get_activity_details`]({{< relref "/reference/tools#get_activity_details" >}})                                                                                        | Sport, timing, load, tags, activity fueling grams, device, Strava-import flag.                                                                         |
| 3    | [`get_activity_intervals`]({{< relref "/reference/tools#get_activity_intervals" >}}) and [`get_activity_splits`]({{< relref "/reference/tools#get_activity_splits" >}}) | Per-rep and per-distance breakdown. Check interval provenance/caveats before treating rows as reps; a single unknown/collapsed row can be an averaged import, not proof of no interval work. `get_activity_intervals` also includes scalar interval `custom_fields` such as lactate when upstream returns them. |
| 4    | [`get_activity_histogram`]({{< relref "/reference/tools#get_activity_histogram" >}})                                                                                    | Time-in-zone distribution for the session.                                                                                                             |
| 5    | [`get_extended_metrics`]({{< relref "/reference/tools#get_extended_metrics" >}})                                                                                        | Decoupling, IF, VI — only those upstream actually fitted.                                                                                              |

For a specific surge, climb, distance-bounded split, or sprint/anaerobic workout that appears as one averaged interval row, [`compute_activity_segment_stats`]({{< relref "/reference/tools#compute_activity_segment_stats" >}}) computes mean/median/p90, NP/IF, drift, or decoupling over an explicit time or distance range. For relative requests like "last 10 km", first use `get_activity_details` to get the activity distance, convert it to meters, and pass explicit bounds such as `start_distance_m = total_distance_m - 10000` and `end_distance_m = total_distance_m`. Do not fetch raw streams and average them in chat.

### Hypoxic-training caveat

Some athletes do workouts in reduced-oxygen environments such as altitude tents,
altitude chambers, or other explicitly logged hypoxic setups. icuvisor surfaces a
`hypoxic_training_caveat` on activity rows, and `_meta.hypoxic_training_caveat`
from `get_extended_metrics`, only when explicit evidence is present in the user
request or exposed activity name/notes/tags/custom fields. Do not infer hypoxic
stress from altitude, elevation gain, or a low SpO2 value by itself.

When the caveat appears, keep the interpretation conservative: CTL, ATL, Form,
and projections are based on logged `training_load`; icuvisor does not change TSS
or apply a hypoxia multiplier. If load is power-based, it may under-represent
extra physiological strain from reduced oxygen. If load is HR-based, it may
capture some acute cardiovascular response, but it is still not a complete
hypoxic-stress model. Use HR, RPE, feel, and recovery trends as supporting
context only.

## A good answer looks like

> **Session:** Tuesday VO2 ride, 1h 12m, 38.4 km, load 95, tags `vo2` and `trainer`, recorded on a Wahoo head unit (`get_activity_details`) — direct upload, full data. Fueling fields show 72 g carbs ingested and 138 g used.
>
> **The work:** 5 x 4 min intervals (`get_activity_intervals`). Targets held well for reps 1-4 (308-314 W, target 310 W) but rep 5 faded to 291 W. Recoveries were on target.
>
> **Aerobic response:** Pw:HR decoupling 6.1% across the work intervals (`get_extended_metrics`) — acceptable for a VO2 session. Time-in-zone was 41% Z5+ during the work blocks (`get_activity_histogram`).
>
> **Takeaways:** (1) The rep-5 fade suggests the session was at the right edge of doable — hold this format rather than adding a sixth rep. (2) Decoupling stayed controlled, so durability is not the limiter here; the ceiling is repeatability at VO2 power.

## Variations

- **Lactate test:** "Analyze this interval session and include the lactate values I entered on each rep. Use `custom_fields.lactate` from `get_activity_intervals` when present."
- **Activity custom field:** "For this ride, include my activity custom field `vo2max_est`. Pass `custom_fields: ["vo2max_est"]` to `get_activity_details`, and if I ask whether it is improving over time, switch to `analyze_correlation` with `metric_x: "custom:vo2max_est"` plus the same `custom_fields` selection over a date range."
- **Race debrief:** "...this was a race — focus on pacing discipline and where I lost time."
- **First-vs-last distance comparison:** "Compare the first 10 km with the last 10 km of this run. Use `get_activity_details` for total distance, then call `compute_activity_segment_stats` with explicit distance bounds (`0..10000 m` and `total_distance_m-10000..total_distance_m`) for average `velocity_smooth`, `watts` if available, and `heart_rate`. Convert velocity to pace in the final answer; do not reduce raw streams in chat."
- **Compare two sessions:** "Compare activity A and activity B — same workout, two weeks apart. Did the hard parts improve?"
- **A Strava import specifically:** "Analyze my most recent Strava-imported activity. Identify genuine Strava imports by the `source` field — a Garmin or Wahoo device means a native upload, not a Strava import — state the blank-field policy first, and confirm the blank payload with `get_activity_details` rather than inferring it from the list."
- **Leave a note:** add "If write tools are enabled, append a one-line summary as a comment on the activity." This uses [`add_activity_message`]({{< relref "/reference/tools#add_activity_message" >}}).

## Why this prompt works

- **Look-it-up step.** Letting the assistant resolve the activity with `get_activities` — including athlete-local date windows for phrases like "last Sunday" — means you never paste raw data or raw IDs, which is what overruns the context window.
- **"Report only the ones actually present."** Extended metrics vary by device and sport. This line stops the assistant inventing a decoupling figure for a run with no power.
- **Strava callout.** Strava-imported activities return blank fields by policy. Naming this up front turns a confusing "your power was 0 W" into an honest "this was a Strava import, so power is unavailable".
