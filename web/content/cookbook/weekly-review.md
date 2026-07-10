---
title: "Weekly training review"
description: "A reusable prompt that turns the last 7-14 days of intervals.icu data into a coached weekly review."
weight: 20
---

A weekly review is the most common thing athletes ask an AI assistant for. This recipe makes the assistant pull load, volume, intensity distribution, and form in a fixed order, then deliver a coached summary with a concrete next step — without inventing anything. For reusable Claude-wide guardrails before you paste this recipe, see the [Claude Project instructions guide]({{< relref "../guides/claude-project-instructions" >}}).

## When to use this

- Every Sunday or Monday, to close out a training week.
- After a heavy block, to see whether form is trending into a hole.
- Any time you want a load-and-intensity readout instead of eyeballing charts.

Use `weekly_review` for a retrospective closeout and optional next-week preview. Use `plan_health_review` when the question is whether the current plan still looks safe and realistic: planned-vs-completed adherence, load/form trajectory, planned deloads, race-date risk, and missing wellness/readiness caveats. Use the [season and block plan]({{< relref "season-and-block-plan" >}}) workflow for 8+ week plan design or scheduling.

## The recipe

Copy this, set the window, and send it as one message.

```text
You are my endurance coach. Using only my intervals.icu data through icuvisor,
review my training for the last 14 days.

Pull, in this order:
1. My athlete profile, so every number uses my units, zones, and thresholds.
2. My fitness trend (CTL, ATL, TSB) across the window.
3. My training-load and volume summary for the window.
4. The list of activities in the window.
5. My time-in-zone and training-load balance for the window.
6. If the full toolset is enabled, my recorded external mechanical work in kJ by configured power zone. Keep this separate from calories or metabolic energy.

Then give me:
- A three-sentence summary of the period: load, volume, intensity mix.
- The two most significant sessions and why they mattered.
- Recovery risk: is form (TSB) trending into a hole?
- One specific thing to do in the next 48 hours.

Rules: do not invent metrics that aren't available — if something is missing,
say so. Keep subjective scales labeled correctly (sleep quality 1-4, feel 1-5).
Tell me which icuvisor tool each key number came from.
```

## What icuvisor does

The assistant should call, in roughly this order:

| Step | Tool | Why |
| --- | --- | --- |
| 1 | [`get_athlete_profile`]({{< relref "/reference/tools#get_athlete_profile" >}}) | Anchors zones, units, and thresholds. |
| 2 | [`get_fitness`]({{< relref "/reference/tools#get_fitness" >}}) | CTL / ATL / TSB across the window. |
| 3 | [`get_training_summary`]({{< relref "/reference/tools#get_training_summary" >}}) | Aggregated load, time, and volume. |
| 4 | [`get_activities`]({{< relref "/reference/tools#get_activities" >}}) | Per-session detail to name notable workouts. |
| 5 | [`compute_zone_time`]({{< relref "/reference/tools#compute_zone_time" >}}) and [`compute_load_balance`]({{< relref "/reference/tools#compute_load_balance" >}}) | Intensity distribution and whether the block is polarized, pyramidal, or threshold-heavy. |
| 6 (optional, full toolset) | [`compute_zone_energy`]({{< relref "/reference/tools#compute_zone_energy" >}}) | Timestamp-weighted external mechanical work by configured power zone; report missing coverage and never describe kJ as calories or metabolic expenditure. |

A focused follow-up can be as short as: `For 2026-06-01 through 2026-06-14, use compute_zone_energy to report recorded cycling mechanical work by configured power zone. Keep include_full false; state any partial coverage and do not convert kJ to calories.`

`analyze_trend` may also be used to state whether load is meaningfully rising or falling. For athlete-defined activity custom fields (for example a VO2Max estimate recorded on workouts), ask the assistant to request that field explicitly and use `analyze_correlation` with `metric_x` or `metric_y` set to `custom:<field_code>` plus `custom_fields: ["<field_code>"]`; do not draw a trend from one visible activity detail row.

## A good answer looks like

> **Last 14 days.** You completed 9 sessions for a load of 612 (`get_training_summary`), up about 11% on the prior fortnight — a normal build-week ramp (`analyze_trend`). 8h 40m total, mostly riding. Your intensity split was 78% low / 9% moderate / 13% high — pyramidal, leaning polarized (`compute_zone_time`, `compute_load_balance`).
>
> **Notable sessions.** Tuesday's 5x5 min VO2 ride (load 95) and Saturday's 3h 20m endurance ride (load 168) carried most of the work.
>
> **Recovery risk.** CTL 62 → 66, ATL 71, TSB -9 and falling (`get_fitness`). You are productively fatigued, not buried — but a second hard day soon would dig a hole.
>
> **Next 48 hours.** Keep tomorrow easy (Z1-Z2, under 60 min) so TSB stabilizes before your next quality session.

Every number is tagged with the tool it came from, and nothing is asserted that the tools did not return.

## Share a reviewed report manually

When you want a public-facing weekly recap, month-end story, race-prep note, or training-journey update, use the MCP prompt `shareable_training_report` instead of asking the assistant to publish anything for you.

```text
Use `shareable_training_report` for a race-prep report from [START_DATE] to [END_DATE], with race_date=[RACE_DATE] and audience=[AUDIENCE]. Draft Markdown first with highlights, one honest challenge, key numbers with icuvisor tool citations, and a private-data review checklist. Do not publish, host, upload, or auto-share the report. If I later ask for HTML, convert the reviewed Markdown to simple static HTML in chat only.
```

The local icuvisor server helps compose the draft from your own intervals.icu data. It does not host a report page, push to social media, or spend an icuvisor app-side credit quota. You use your chosen AI client or subscription, and that client's provider terms still apply. Review and redact private health details, locations, notes, identifiers, and race logistics before copying, exporting, or posting the report yourself.

## Variations

- **Shareable report:** ask for `shareable_training_report` when the output is meant to become a manually shared Markdown recap, journey update, or race-prep note. Keep no-publish/no-host/no-auto-share and review/redact instructions in the prompt.
- **Plan-health review:** ask for `plan_health_review` when you want a transparent audit of the upcoming plan. Require evidence for any risk label, caveat missing wellness/readiness data, and ask the assistant to state when a race date is only a scenario anchor because no race event was found.
- **Monthly review:** change the window to "the last 28 days" and ask for the polarization trend week-over-week.
- **Compare two periods:** "Compare the last 14 days with the 14 days before — what changed in load and intensity?"
- **Sport-specific:** add "Only consider rides" or "Only consider runs" to scope the summary.
- **Custom-field correlation:** "For the last 8 weeks, test whether my activity custom field `vo2max_est` tracks CTL, ATL, TSB, or training load. Use `analyze_correlation` with `custom_fields: ["vo2max_est"]` and `metric_x: "custom:vo2max_est"`; if there are too few paired samples, say so instead of inferring a trend."

## Why this prompt works

- **Numbered tool order** stops the assistant from skipping the profile and misreading zones.
- **"Do not invent metrics that aren't available"** is the line that cuts hallucinated zone percentages — the most-reported failure with AI training analysis.
- **"Tell me which tool each number came from"** makes the answer auditable, so you can spot a wrong number instead of trusting a fluent paragraph.

{{< callout type="info" >}}
If your client supports [MCP prompts]({{< relref "/reference/resources-prompts" >}}), `weekly_review` covers this retrospective workflow with server-side guardrails, `shareable_training_report` drafts a manually reviewed Markdown report without publishing or hosting it, and `plan_health_review` covers the current-plan risk audit without introducing an opaque score.
{{< /callout >}}
