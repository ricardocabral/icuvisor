---
title: "Structured workouts vs. device laps"
description: "Why interval data carries a source label, and how to read it without inventing workout steps."
---

When you ask for the intervals of an activity, the rows you get back can mean two completely different things — and telling them apart matters.

## Two kinds of "interval"

An *interval* in a structured workout is a planned step: "5 minutes at threshold", "8 x 30 seconds". A *device lap* is something your bike computer or watch created on its own — most often an automatic split every 1 km or 1 mile.

Both arrive as interval rows from [`get_activity_intervals`]({{< relref "/reference/tools#get_activity_intervals" >}}), and from the numbers alone they look alike. An AI assistant that cannot tell them apart will happily report that you "missed interval 3 of your workout" when interval 3 was just the third kilometre of a steady ride.

## How icuvisor labels the source

To head off that guess, the response carries additive metadata:

- `_meta.interval_source` — `structured_workout`, `device_laps`, or `unknown`.
- `_meta.auto_lap_suspected` — `true` when the rows are generic and near-uniform (1 km, 1 mi, or another supported fixed duration), the signature of device auto-laps.

## What it means for analysis

When auto-laps are suspected, the rows are splits, not planned segments. An analysis should describe them as distance or time splits and must not claim the athlete hit or missed individual workout steps — there were no planned steps to hit. When the source is a genuine structured workout, step-by-step compliance is fair game.

This is the same terse, honest-by-default stance behind the rest of icuvisor: surface what the data actually supports, and signal uncertainty rather than paper over it. The [activity retrospective]({{< relref "/cookbook/activity-retrospective" >}}) recipe leans on this when it breaks a session down rep by rep.
