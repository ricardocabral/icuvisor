---
title: "Cookbook"
description: "Copy-paste prompts and multi-step recipes for getting reliable training answers from icuvisor."
weight: 35
type: docs
cascade:
  type: docs
---

The cookbook is a library of prompts that work well with icuvisor. Each one is written so your AI assistant reaches for the right [tools]({{< relref "/reference/tools" >}}), grounds every number in your real intervals.icu data, and refuses to guess when data is missing.

Two kinds of content:

- [What can I ask icuvisor?]({{< relref "what-can-i-ask" >}}) — beginner-friendly first prompts grouped by athlete, coach, and troubleshooting use case.
- The [prompt library]({{< relref "prompt-library" >}}) — short, single-message prompts grouped by task. Copy one, adjust the dates, send.
- The [icuvisor Agent Skill]({{< relref "icuvisor-agent-skill" >}}) — reusable `SKILL.md` instructions for skills-compatible OpenAI, Anthropic, and other AI clients.
- **Recipes** — longer, reusable prompt templates that drive a multi-step job (a weekly review, a taper, a roster triage). Treat a recipe as a small agent skill: paste it, fill the blanks, and the assistant runs the whole workflow.

{{< cards >}}
  {{< card link="what-can-i-ask" title="What can I ask icuvisor?" subtitle="Start here after connecting: first prompts and plain-language use cases." >}}
  {{< card link="prompt-library" title="Prompt library" subtitle="One-line prompts for everyday questions, grouped by task." >}}
  {{< card link="icuvisor-agent-skill" title="icuvisor Agent Skill" subtitle="Reusable SKILL.md instructions for skills-compatible AI clients." >}}
  {{< card link="weekly-review" title="Weekly training review" subtitle="Summarize load, intensity, and recovery risk for the last 7-14 days." >}}
  {{< card link="data-quality-report" title="Data quality report" subtitle="Diagnose missing streams, stale sync, Strava restrictions, and other visibility gaps." >}}
  {{< card link="readiness-check" title="Readiness check" subtitle="Decide whether to train hard today from wellness and form." >}}
  {{< card link="activity-retrospective" title="Activity retrospective" subtitle="Break down a single ride, run, or race in detail." >}}
  {{< card link="ftp-and-zones" title="FTP and zones review" subtitle="Decide whether a threshold has moved and update zones." >}}
  {{< card link="season-and-block-plan" title="Season and block plan" subtitle="Build a periodized plan toward a goal event." >}}
  {{< card link="build-workouts" title="Build and schedule workouts" subtitle="Create structured workouts and put them on the calendar." >}}
  {{< card link="race-week-taper" title="Race-week taper" subtitle="Plan the final days before a goal event." >}}
  {{< card link="coach-roster-triage" title="Coach roster triage" subtitle="Scan a roster of athletes for red flags (coach mode)." >}}
{{< /cards >}}

## How to get reliable answers

icuvisor only helps if the assistant actually calls it instead of answering from memory. These habits — baked into every recipe below — make that happen.

1. **Name your data.** Phrases like "using my intervals.icu data" or "use icuvisor" make the assistant fetch real numbers instead of estimating. The recipes always include this.
2. **Give a concrete window.** "the last 14 days", "since 1 April", "race week" maps cleanly onto date-range tool arguments. Vague windows produce vague tool calls.
3. **Say what not to do.** "Do not invent metrics that aren't available" is the single most effective line for cutting hallucination. icuvisor returns explicit "unavailable" signals — tell the assistant to honor them.
4. **Ask for sources.** "Tell me which tool each number came from" turns a plausible-sounding answer into an auditable one.
5. **One job per message.** Building a season plan and authoring ten workouts in one turn overruns the context window. The recipes split big jobs into stages on purpose.
6. **Let it look things up.** If a recipe needs an activity or event, let the assistant find the ID with `get_activities` or `get_events` — you should not have to paste raw data.

{{< callout type="warning" >}}
**Subjective scales are not 0-10.** intervals.icu sleep quality is **1-4**, athlete-reported feel is **1-5**, and RPE is **1-10**. icuvisor labels every scale in its responses. If the assistant restates one wrong ("a 3 is a terrible sleep"), correct it — a 3/4 is good sleep.
{{< /callout >}}

## What icuvisor will not do

Knowing the limits keeps you from chasing answers icuvisor cannot give:

- **Strava-imported activities** are returned with most fields blank — Strava's API terms forbid third-party apps from reading them. icuvisor labels these so the assistant says "unavailable" instead of guessing. Connect your device directly to intervals.icu for full data.
- **Writes are gated.** Creating or updating events, workouts, and wellness rows works only when the server runs with write mode enabled; deletes need delete mode. See [safety modes]({{< relref "/reference/safety-modes" >}}). If a write recipe cannot run, ask for a preview of the change instead.
- **icuvisor does not compute what intervals.icu does not expose.** If a metric (decoupling, VI, a custom field) is not in the data, the recipes ask the assistant to say so rather than calculate a substitute.

## Cookbook recipes vs. MCP prompts

If your client supports [MCP prompts]({{< relref "/reference/resources-prompts" >}}), icuvisor ships nine curated ones (`training_analysis`, `recovery_check`, `weekly_planning`, `weekly_review`, `shareable_training_report`, `plan_health_review`, `race_week_taper`, `coach_athlete_onboarding`, `coach_roster_triage`) you can invoke directly — they carry the same guardrails server-side. The cookbook recipes are the portable equivalent: plain text that works in any client, including ones with no prompt support. Each recipe notes its matching MCP prompt where one exists.
