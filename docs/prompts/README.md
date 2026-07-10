# Icuvisor prompt packs

These downloadable prompt packs turn Icuvisor's MCP prompts and deterministic tools into copyable client modes. They are plain Markdown so they work in clients with custom assistant profiles and in clients that only support copy/paste chat instructions.

## Packs

| Pack | Use when | Source prompt |
|------|----------|---------------|
| [Weekly review](client-prompt-packs/weekly-review.md) | Reviewing the previous athlete-local training week and, optionally, previewing next week. | `weekly_review` |
| [Coaching conversation handoff](client-prompt-packs/coaching-handoff.md) | Preserving reviewed goals, constraints, decisions, and source-labelled evidence when moving to a fresh client conversation. | `coaching_handoff` |
| [Race-week taper](client-prompt-packs/race-week-taper.md) | Preparing the final week before a known race with calendar, plan, and projection context. | `race_week_taper` |
| [Ride analysis](client-prompt-packs/ride-analysis.md) | Analyzing one ride with activity details, intervals, streams only when needed, and deterministic analyzers. | `ride_analysis` |
| [Fueling review](client-prompt-packs/fueling-review.md) | Reviewing logged activity and daily nutrition evidence without estimating intake or prescribing targets. | `fueling_review` |
| [Coach roster triage](client-prompt-packs/coach-roster-triage.md) | Scanning one authorized coach-mode roster athlete for recovery, compliance, race/event risk, and stale data. | `coach_roster_triage` |

## How to use a pack

### Clients with custom modes or profiles

1. Connect Icuvisor as an MCP server and keep the relevant toolset enabled.
2. Open one pack and copy the `Copy/paste prompt` block into the client's custom mode, profile, instruction, or system-prompt field.
3. Name the mode after the pack, for example `Icuvisor weekly review`.
4. Start a new chat in that mode and provide the requested week, race, ride, handoff window, or athlete selector.

### Clients without custom modes

1. Start a new chat with Icuvisor MCP connected.
2. Paste the pack's `Copy/paste prompt` block as the first message.
3. Add the concrete request after the prompt, such as the week to review, the race anchor, the ride to inspect, or the coaching context to hand off.
4. Start a fresh chat when switching packs so old instructions do not conflict.

## Why use these instead of generic coaching prompts?

The packs are designed to make the assistant call Icuvisor tools for evidence and deterministic calculations instead of inventing baselines or formulas in chat. They steer the model toward:

- tool-returned `_meta.method`, `_meta.source_tools`, assumptions, caveats, and unit metadata;
- analyzer tools for zone time, histograms, load balance, projections, compliance, and segment statistics;
- terse default responses, with raw streams or `include_full` only when the question requires them;
- explicit missing-data, stale-data, Strava-import, permission, and provider-scale caveats.

Do not paste API keys, OAuth tokens, private identifiers, raw tool payloads, or confidential athlete data into a prompt pack. Icuvisor reads credentials from local setup or hosted OAuth, not from chat instructions.
