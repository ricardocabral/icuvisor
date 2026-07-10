---
title: "Masters plan review"
description: "Audit an existing endurance plan with transparent athlete-specific evidence, not age rules or opaque scores."
weight: 75
---

Use this read-only workflow to audit an existing plan when you want a conservative, transparent discussion of hard-session spacing, load context, recovery evidence, race proximity, and feasibility. “Masters” is an audience label only: it does not ask for or derive age, date of birth, an age cutoff, or an age-based training rule.

## Start here

- In a client that supports MCP prompts, invoke `masters_plan_review`.
- Otherwise, paste the canonical [Masters plan review prompt pack](https://github.com/ricardocabral/icuvisor/blob/main/docs/prompts/client-prompt-packs/masters-plan-review.md) into a new chat.

Both entry points use the same read-only evidence contract. They never call write or delete tools, even after you approve a possible change. A calendar adjustment remains a conditional, unapplied proposal for you to make yourself.

## Windows and inputs

By default, the workflow resolves a 14-day athlete-local planned window from today through day 13. It then uses the 28 completed-history days immediately before `planned_start` and the 56 personal-baseline days immediately before that history. The windows must not overlap.

You can provide optional `planned_start` and `planned_end` together as strict `YYYY-MM-DD` dates. A same-day review is valid, and the inclusive planned window is capped at 90 athlete-local days. `history_lookback_days` accepts 1-90 days (default 28); `baseline_lookback_days` accepts 1-180 days (default 56). A `race_name` requires its matching strict `race_date`.

## The review

Copy this only when your client does not surface the MCP prompt:

```text
Audit my existing endurance plan with icuvisor using athlete-local, source-labelled evidence. Treat “masters” only as an audience label: do not ask for or use my age or date of birth.

Resolve relative dates with get_athlete_profile and resolve_calendar_dates. Keep personal baseline/history, completed, planned, and race windows non-overlapping. Read events, training-plan rows, activities, fitness/load context, and wellness freshness for their assigned windows; fetch all pages needed for a completeness claim and call partial coverage partial. Confirm a calendar race, or call my supplied race date a scenario anchor.

Use compute_baseline for one supported metric at a time and retain its status, sample counts, missing days, freshness, method, and formula reference. Call a session hard only when I identify it or sourced plan/activity intensity detail supports it; titles, aggregate load, calendar proximity, absent/invalid zones, and age are not enough. Use get_fitness_projection only with copyable plan targets or values I explicitly supply, and expose every returned assumption. Never treat projection defaults as a recommendation.

Return these sections in order: Observed tool evidence (tool, athlete-local window, freshness/coverage); Athlete-stated preferences (availability and requested duration only); Cautious interpretation; Insufficient evidence and focused questions; Reviewable proposals.

For any unsupported dimension, name the missing evidence, make no conclusion for that dimension, and ask one focused question. Do not make medical, diagnostic, treatment, injury-risk, or opaque readiness/risk-score claims. Do not write or delete anything; every calendar idea is an unapplied conditional proposal.
```

## What counts as evidence

| Question | Supported route | Boundary |
| --- | --- | --- |
| Personal baseline | [`compute_baseline`]({{< relref "/reference/tools#compute_baseline" >}}) for one eligible metric | Preserve status, samples, missing days, freshness, method, and formula metadata. Do not combine metrics into a readiness or risk score. |
| Hard-session spacing | Athlete identification or sufficiently detailed, sourced plan/activity intensity evidence | Titles, aggregate load, date proximity, missing/invalid zones, and age do not classify a session as hard. |
| Load context | [`get_fitness`]({{< relref "/reference/tools#get_fitness" >}}), [`get_training_summary`]({{< relref "/reference/tools#get_training_summary" >}}), and [`compute_load_balance`]({{< relref "/reference/tools#compute_load_balance" >}}) | `compute_load_balance` is a window aggregate, not individual-session classification. |
| Projection | [`get_fitness_projection`]({{< relref "/reference/tools#get_fitness_projection" >}}) with copyable targets or values you supplied | Treat every projection as its returned assumptions, never as a universal masters ramp or recovery rule. |
| Wellness | [`get_wellness_data`]({{< relref "/reference/tools#get_wellness_data" >}}) | Current-day rows are partial; stale, missing, and provider-native readiness data cannot support a broad readiness conclusion. |

If an advanced helper is unavailable, ask the assistant to call [`icuvisor_list_advanced_capabilities`]({{< relref "/reference/tools#icuvisor_list_advanced_capabilities" >}}), name the gap, and avoid replacing it with chat-side calculations.

## Insufficient evidence is useful

Ambiguous plan detail or hard-session classification, absent/invalid zones, incomplete history, stale or missing wellness, provider-native readiness gaps, missing race context, and missing projection targets each suppress only the affected conclusion. The assistant should name that gap and ask one focused question; it should not treat missing data as complete, zero, or a reason to invent an age rule.

A confirmed calendar race is observed evidence. A date you supplied without a matching event is only a scenario anchor. Availability and requested duration are your stated preferences, not inferred hard constraints or an implied session count.

## Limits

This workflow is an evidence review, not a future science-backed plan-validation engine. It cannot derive a universal age rule, a medical or injury conclusion, an individualized treatment, an automatic hard-day gap, a ramp cap, a recovery cadence, a readiness/risk score, or an automatic calendar edit. The separate v2.2 rule-engine work will require transparent cited and versioned rules before it can make rule-based plan-validation claims.
