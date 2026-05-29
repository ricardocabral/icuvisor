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

## The assumptions are part of the answer

Because the result depends entirely on its assumptions, the tool reports them back to you. `_meta.assumptions` records the scenario it ran: horizon length, weekly ramp percentage, recovery-week cadence and load, the number of explicit planned loads supplied, and the CTL/ATL time constants. `_meta.boundaries` records the limits: the horizon is capped at 180 days, no hidden upstream periodization fields are read, and explicit `planned_daily_loads` replace the modelled ramp only on the dates they cover. Plan-health reviews should quote those assumptions instead of collapsing them into an opaque score.

Treat those fields as the fine print of the projection. If a scenario assumed a 5%-per-week ramp and you would never train that way, the curve is answering a different question than the one you asked — change the assumption and run it again.

## Reading the output

By default the tool returns only the summary. Set `include_full: true` to get the daily projected CTL/ATL/TSB curve — see [Terse by default]({{< relref "terse-by-default" >}}) for when that opt-in is worth the extra tokens. If wellness/readiness or race-event data is missing, the projection does not fill it in; the assistant should say what is missing and treat a user-supplied race date as a scenario anchor when no matching race event is found.
