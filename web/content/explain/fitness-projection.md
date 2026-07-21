---
title: "How fitness projection works"
description: "Why get_fitness_projection is a deterministic scenario model, not a prediction of the future."
---

[`get_fitness_projection`]({{< relref "/reference/tools#get_fitness_projection" >}}) answers questions like "where will my form be on race day if I keep ramping load?" It is tempting to read its output as a forecast. It is not one — it is a deterministic *scenario model*, and understanding that distinction is what keeps an AI assistant honest about it.

## A scenario, not a forecast

A forecast claims what *will* happen. A scenario answers *if this, then that*: given a starting point and a set of assumptions, here is how the numbers move. icuvisor projects fitness with a closed `deterministic_ctl_atl_tsb` model — the same exponentially-weighted CTL/ATL/TSB maths intervals.icu itself uses — run forward from today.

It is deterministic on purpose. The same inputs always produce the same curve, so the assistant can explain *why* a projection looks the way it does. Free-form "physiology" models, where the model invents its own progression, are rejected: they would be unreproducible and impossible to audit.

## Where the curve starts

The projection seeds CTL, ATL, and TSB from the athlete-local `start_date` returned by [`get_fitness`]({{< relref "/reference/tools#get_fitness" >}}). The starting point is your real, current fitness — not an estimate — so a projection is only ever as current as your logged data.

## Logged load is the model input

CTL, ATL, Form, and `get_fitness_projection` all move from logged `training_load`.
icuvisor does not adjust TSS, invent a hypoxia multiplier, or separately model
reduced-oxygen exposure. Treat hypoxic-training caveats only when explicit
provenance exists — for example, the user says the session was in an altitude
tent/chamber or an activity name, note, tag, or selected custom field clearly says
hypoxic/reduced oxygen. Altitude, elevation gain, or SpO2 by itself is context,
not proof that the session should be interpreted as hypoxic training.

If a hypoxic session used power-based TSS/load, the logged load may
under-represent extra physiological strain from reduced oxygen. HR-based load may
capture some acute cardiovascular response, but it is not a complete
hypoxic-stress model. Use HR, RPE, feel, and recovery trends as supporting
context; do not inflate CTL/ATL/Form unless the logged `training_load` itself was
changed upstream.

## Per-sport load trends are a context aid

For runners, cyclists, swimmers, and triathletes, a single combined load can hide divergent sport fatigue. [`get_fitness`]({{< relref "/reference/tools#get_fitness" >}}) therefore has an opt-in `include_per_sport_load_trends: true` view that computes running, cycling, swimming, and other load trends from visible `byCategory[].training_load` in the athlete summary.

Those per-sport CTL/ATL/TSB-style values are warmed estimates, not upstream-native per-sport fitness fields. They are useful for questions like "is my run load rising while bike load is flat?" or "am I carrying swim fatigue into a run block?" They should not replace the combined upstream CTL/ATL/TSB when judging global form, fatigue, or race-day freshness. If category data is missing or the warm-up history is short, the response includes caveats in `_meta.per_sport_load_trends.caveats`; quote those caveats instead of treating the per-sport numbers as precise physiology.

## The assumptions are part of the answer

Because the result depends entirely on its assumptions, the tool reports them back to you. `_meta.assumptions` records the scenario it ran: horizon length, weekly ramp percentage, recovery-week cadence and load, the number of explicit planned loads supplied, and the CTL/ATL time constants. `_meta.boundaries` records the limits: the horizon is capped at 180 days, no hidden upstream periodization fields are read, and explicit `planned_daily_loads` replace the modelled ramp only on the dates they cover. Plan-health reviews should quote those assumptions instead of collapsing them into an opaque score.

Treat those fields as the fine print of the projection. If a scenario assumed a 5%-per-week ramp and you would never train that way, the curve is answering a different question than the one you asked — change the assumption and run it again.

## Planned targets are inputs, not evidence

Annual-plan tools can summarize explicit weekly TARGET loads already on your calendar or produce a read-only season proposal. Their projection bridge contains only weekly targets with an explicit load; it does not infer an athlete's preferred ramp, recovery cadence, taper, or intensity distribution from calendar text.

When those targets are supplied to `get_fitness_projection`, each `week_start_date` is an athlete-local ISO Monday. The tool distributes its `training_load` evenly across the seven dates of that week. A `planned_daily_loads` entry wins for its matching date and is not redistributed over the other days. Partial weeks remain partial rather than being reweighted.

That makes a calendar target a transparent scenario input, not proof that training happened or a prediction that it will. Compare the resulting curve with future completed load as it becomes available, and keep the assumptions in the answer.

## Reading the output

By default the tool returns only the summary. Set `include_full: true` to get the daily projected CTL/ATL/TSB curve — see [Terse by default]({{< relref "terse-by-default" >}}) for when that opt-in is worth the extra tokens. If wellness/readiness or race-event data is missing, the projection does not fill it in; the assistant should say what is missing and treat a user-supplied race date as a scenario anchor when no matching race event is found.
