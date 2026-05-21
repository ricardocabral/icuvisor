---
title: "Coach roster triage"
description: "Scan a roster of athletes for red flags and decide who needs attention this week (coach mode)."
weight: 90
---

A coach with a roster needs a fast way to see who is thriving, who is drifting, and who is heading for trouble. This recipe runs a per-athlete scan in [coach mode]({{< relref "/guides/coach-mode" >}}) and ranks the roster by who needs attention — without exposing any API key in the chat.

## When to use this

- A weekly roster review, to set your coaching priorities.
- Monday triage before writing check-in messages.
- After a hard training week, to catch athletes trending toward overtraining.

{{< callout type="warning" >}}
This recipe needs the server running in **coach mode** with a roster configured. The coach-held API key never enters the conversation. The `athlete_id` argument only *selects* whose data to read — it is a selector, not a credential. See the [coach mode guide]({{< relref "/guides/coach-mode" >}}).
{{< /callout >}}

## The recipe

```text
Coach-mode roster triage. Use icuvisor.

1. List the athletes on my roster.
2. For each athlete in turn, select them and pull: their activities for the
   last 7 days, planned events for the next 7 days, fitness trend
   (CTL / ATL / TSB), and recent wellness.
3. For each athlete, give a one-line status and a red / amber / green flag.

Then:
- Rank the roster by who needs my attention most this week.
- For the top two or three, write a short, specific coaching note I could
  send.

Rules: athlete_id selects whose data to read — it is not a credential, and no
API key should appear in this chat. Be explicit about any athlete whose data
you could not access. Do not modify any athlete's calendar or settings.
```

## What icuvisor does

| Step | Tool | Why |
| --- | --- | --- |
| 1 | [`list_athletes`]({{< relref "/reference/tools#list_athletes" >}}) | Returns the configured roster. |
| 2 | [`select_athlete`]({{< relref "/reference/tools#select_athlete" >}}) | Switches the active athlete for the calls that follow. |
| 3 | [`get_activities`]({{< relref "/reference/tools#get_activities" >}}), [`get_events`]({{< relref "/reference/tools#get_events" >}}), [`get_fitness`]({{< relref "/reference/tools#get_fitness" >}}), [`get_wellness_data`]({{< relref "/reference/tools#get_wellness_data" >}}) | The per-athlete read, repeated for each athlete. |

Per-athlete tool access follows your coach-mode ACLs — an athlete you have limited to read-only stays read-only here.

## A good answer looks like

> **Roster triage — week of [DATE].** 5 athletes scanned (`list_athletes`).
>
> | Athlete | Flag | Status |
> | --- | --- | --- |
> | A | 🔴 red | TSB -28, HRV down 5 days running, missed 2 of 3 planned sessions |
> | B | 🟡 amber | Big load week (+22%), wellness still fine — watch |
> | C | 🟢 green | On plan, form +4, wellness stable |
> | D | 🟢 green | Light week as planned, fully recovered |
> | E | ⚪ — | No activities in 7 days and no wellness — data may not be syncing |
>
> **Priority order:** A, then E, then B.
>
> **Athlete A:** "Your HRV and form both say you're deep in a hole — let's pull this week back. Swap tomorrow's intervals for an easy hour and we'll reassess Thursday. Anything going on with sleep or stress outside training?"
>
> **Athlete E:** "I'm not seeing any activity or wellness data from you this week — can you check that your device is still syncing to intervals.icu?"

## Variations

- **One athlete deep-dive:** "Just triage athlete [ATHLETE_ID] in detail" — pairs well with the [weekly review]({{< relref "weekly-review" >}}) recipe scoped to that athlete.
- **Pre-camp check:** "...flag anyone not recovered enough to start a training camp Monday."
- **Compliance focus:** "...rank by who is least compliant with their planned sessions" — adds [`compute_compliance_rate`]({{< relref "/reference/tools#compute_compliance_rate" >}}).

## Why this prompt works

- **One athlete at a time.** Forcing `select_athlete` then a scan, per athlete, keeps each athlete's data separate and the tool sequence legible — instead of a tangled cross-athlete query.
- **Flag-and-rank.** A red/amber/green flag plus a priority order turns a wall of data into a coaching to-do list.
- **Credential reminder.** Restating that `athlete_id` is a selector keeps the assistant from ever asking for a key — the coach key stays server-side.

{{< callout type="info" >}}
The `coach_roster_triage` [MCP prompt]({{< relref "/reference/resources-prompts" >}}) runs a single-athlete version of this scan with the selector-not-credential rule enforced server-side.
{{< /callout >}}
