Prompt: Coach roster triage
Scope: athlete_id=i12345, start_date=2026-05-01, end_date=2026-05-14.
Resources: icuvisor://athlete-profile, icuvisor://event-categories.
Tools: get_athlete_profile, get_wellness_data, get_fitness, get_training_summary, get_events, get_activities.
Do:
- Treat athlete_id as a coach-mode selector for server-side calls, never as a credential; do not ask for API keys.
- Scan wellness, fitness/load, upcoming events, missed/completed activities, and stale data warnings.
- Prioritize interventions: urgent health/recovery flags, compliance drift, race/event risk, then routine follow-up.
Guardrails:
- Do not request or accept intervals.icu API keys in chat.
- Prefer terse default tool responses; use include_full only when the user asks or evidence is missing.
Return: triage status, top risks, evidence by tool, recommended coach action, and what to check next.
