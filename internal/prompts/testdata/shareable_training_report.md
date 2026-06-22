Prompt: Shareable training report
Scope: report_type=race_prep, start_date=2026-05-01, end_date=2026-06-07, race_date=2026-06-07, audience=family.
Resources: icuvisor://athlete-profile, icuvisor://event-categories.
Tools: get_athlete_profile, get_fitness, get_training_summary, get_activities, get_events, get_training_plan, get_wellness_data, compute_zone_time, compute_load_balance, analyze_trend, icuvisor_list_advanced_capabilities.
Do:
- Read profile first for athlete-local timezone, units, sport settings, and dates; define the report window before fetching data.
- Gather only the summary evidence needed for a public-facing story: fitness/form, volume/load, notable sessions, planned/race context, intensity mix, and wellness caveats when useful.
- Use analyzers such as compute_zone_time, compute_load_balance, or analyze_trend only when they support the requested story; if unavailable, call icuvisor_list_advanced_capabilities and continue from ordinary reads.
- Draft Markdown first with a short title, timeframe, highlights, one honest challenge, key numbers with tool citations, and a concise next-focus section.
- If the user asks for HTML, convert the reviewed Markdown to simple static HTML in chat; icuvisor does not generate, publish, upload, or host HTML.
- Ask the athlete to review and redact private health, location, notes, identifiers, and race logistics before copying, exporting, or posting anywhere.
Guardrails:
- Do not request or accept intervals.icu API keys in chat.
- Prefer terse default tool responses; do not use include_full, raw streams, or heavy payloads unless the user explicitly asks or evidence is missing.
- Do not publish, host, upload, auto-share, or connect to social platforms; the athlete manually shares only after review.
- Do not invent missing metrics, race details, locations, health claims, or emotional framing not supported by data or the user's words.
Return: Markdown report draft plus private-data review checklist, cited evidence, missing/stale-data caveats, and optional HTML-conversion offer only after user review.
