---
title: "From Garmin to Claude"
description: "Sync device-provider training data into intervals.icu, expose it locally with icuvisor, and ask Claude grounded training questions."
weight: 20
---

This walkthrough shows the common path from a Garmin watch or bike computer to a grounded Claude answer:

your device provider syncs to intervals.icu, icuvisor reads intervals.icu through a local MCP server, and Claude calls icuvisor when you ask training questions.

icuvisor does **not** connect directly to Garmin, Wahoo, Coros, Polar, Suunto, Strava, or any other device provider. The activity, wellness, and calendar data must already be visible in intervals.icu.

## What you'll need

- A Garmin or other device-provider account already connected to intervals.icu.
- Recent data visible in intervals.icu.
- icuvisor installed and configured with your intervals.icu athlete ID, timezone, and API key.
- Claude Desktop or Claude Code with icuvisor enabled.
- About 10 minutes after your intervals.icu sync is working.

## The data path

```text
Garmin / device provider
        |
        | provider sync or export already configured outside icuvisor
        v
intervals.icu account
        |
        | intervals.icu API over your local icuvisor server
        v
icuvisor on your computer
        |
        | Model Context Protocol (MCP), usually stdio
        v
Claude Desktop or Claude Code
```

Keep the direction in mind when troubleshooting. If an activity is missing in intervals.icu, Claude cannot fetch it through icuvisor yet. If it appears in intervals.icu but Claude cannot see it, check the icuvisor and Claude connection.

## Step 1 — Confirm the data reached intervals.icu

Open intervals.icu and confirm the thing you want to ask about is already there:

- The recent ride, run, swim, or workout appears in your activities.
- Wellness fields such as sleep, HRV, or resting heart rate appear if you plan to ask recovery questions.
- Planned events appear on the calendar if you want planned-vs-completed analysis.

Do not copy API keys, athlete IDs, screenshots, or private activity details into Claude. The goal is to let Claude request the structured data through icuvisor.

## Step 2 — Install and configure icuvisor

If you have not installed icuvisor yet, follow the [macOS install guide]({{< relref "../install/macos" >}}) and run setup:

```bash
/Applications/icuvisor.app/Contents/MacOS/icuvisor setup
```

Setup stores your intervals.icu API key in the OS keychain and writes only non-secret settings, such as athlete ID and timezone, to the config file. If you use another platform, start from the [install section]({{< relref "../install" >}}).

## Step 3 — Connect Claude

Choose the Claude client you use:

- [Claude Desktop]({{< relref "../connect/claude-desktop" >}}) — use the Desktop Extension when possible, or the manual JSON/keychain fallback.
- [Claude Code]({{< relref "../connect/claude-code" >}}) — add a project `.mcp.json` that starts the local icuvisor binary over stdio.

After changing the MCP config, start a new Claude chat or session so Claude refreshes the tool catalog.

## Step 4 — Ask the first grounded question

Paste this into Claude:

```text
Use icuvisor to answer from my intervals.icu data. What are my current FTP, threshold heart rate, preferred units, and timezone? Do not use memory or estimates; if a field is missing, say it is missing.
```

A good answer should mention that Claude used icuvisor, report only values that intervals.icu exposed, and avoid asking you to paste your API key into chat.

If that works, try a recent activity question:

```text
Use icuvisor to find my most recent completed activity in intervals.icu. Summarize its sport, duration, distance, load, intensity, and any unavailable fields. Tell me which icuvisor tool provided the main numbers.
```

Example answer shape, using synthetic values:

> Your most recent activity is an example ride from Monday: 1h 18m, 42.0 km, load 74, mostly endurance intensity. Power and heart-rate summaries were available, but detailed interval structure was not. I used `get_activities` for the activity row and did not estimate unavailable fields.

## Copy-paste prompts

Use these once the first call succeeds. Adjust the date window before sending.

### Weekly review

```text
You are my endurance coach. Using only my intervals.icu data through icuvisor, review my last 14 days of training.

Include total load, total time, sport mix, intensity distribution if available, the two most important sessions, and whether current form suggests productive fatigue or recovery risk. Tell me which icuvisor tool each key number came from. Do not invent metrics that are unavailable.
```

### Recovery check

```text
Use icuvisor and my intervals.icu data to check whether I look recovered enough for a hard session today.

Look at today's fitness/form, recent wellness, recent training load, and planned events. If today's wellness has not synced yet, say that clearly instead of guessing. Keep subjective scales labeled correctly and end with one practical recommendation.
```

### Missing-data troubleshooting

```text
Use icuvisor to troubleshoot missing training data in this Claude session.

Check whether my recent activities and wellness are visible from intervals.icu, identify any activities with unavailable power, heart-rate, duration, or source fields, and explain whether the likely issue is upstream sync to intervals.icu, an imported-source limitation, or this Claude/icuvisor connection. Do not ask me to paste API keys or private screenshots.
```

### First-week habit

```text
For the next week, when I ask about training, use icuvisor and my intervals.icu data first. Cite the icuvisor tool behind important numbers, state when data is missing or stale, and avoid calculating substitutes for fields intervals.icu does not expose.
```

For more reusable prompts, browse the [prompt library]({{< relref "../cookbook/prompt-library" >}}) and [weekly review recipe]({{< relref "../cookbook/weekly-review" >}}).

## Source limitations to know

- **icuvisor reads intervals.icu, not Garmin directly.** If the device-provider sync is broken upstream, fix that before troubleshooting Claude.
- **Imported-source fields can be unavailable.** Some provider/import paths expose only partial activity data. When a field is unavailable, ask Claude to say so rather than infer it.
- **Device laps are not always workout steps.** Auto-laps from a watch or head unit can look like intervals. For interval analysis, ask Claude to respect icuvisor's interval-source metadata and avoid claiming you hit or missed planned steps unless a structured workout was actually present. See [structured workouts vs. device laps]({{< relref "../explain/interval-sources" >}}).
- **Subjective scales have specific ranges.** Sleep quality, feel, and RPE are not interchangeable; icuvisor labels scales in responses.
- **Claude conversations can cache tool catalogs.** After upgrading icuvisor or changing config, start a new chat/session. If answers still look stale, use the [troubleshooting guide]({{< relref "../guides/troubleshooting#stale-conversations-and-cached-tool-catalogs" >}}).

## If Claude says it cannot see your data

1. Confirm the activity or wellness row is visible in intervals.icu.
2. Confirm icuvisor starts locally:

   ```bash
   /Applications/icuvisor.app/Contents/MacOS/icuvisor version
   ```

3. Start a new Claude chat or session after changing MCP configuration.
4. Ask the first grounded question again.
5. If only some fields are missing, treat them as upstream unavailable rather than asking Claude to estimate.

Once the first call, weekly review, and recovery check work, you have the core loop: device data lands in intervals.icu, icuvisor exposes it locally, and Claude answers from the structured source instead of from guesses.
