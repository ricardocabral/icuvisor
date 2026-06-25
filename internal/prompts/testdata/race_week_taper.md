Prompt: Race-week taper
Scope: race_date=2026-06-07, race_name=A Race.
Resources: icuvisor://athlete-profile, icuvisor://event-categories, icuvisor://workout-syntax.
Tools: get_athlete_profile, resolve_calendar_dates, get_events, get_training_plan, get_fitness, get_training_summary, get_activities, compute_compliance_rate, get_fitness_projection, icuvisor_list_advanced_capabilities.
Do:
- Find the race event by date/name and confirm priority/category, sport, distance, expected duration, and goal when missing; if the user supplied a relative race date, countdown, weekday, or weekday/date pairing, first call resolve_calendar_dates and use its athlete-local result instead of UTC, client-time, or model arithmetic.
- Review active plan, planned events, fitness, training summary, recent activities, compute_compliance_rate workout_status/status counts/caveats, and get_fitness_projection race-day form assumptions without pulling raw streams or inferring completion from calendar/activity co-occurrence.
- If advanced helpers are unavailable, call icuvisor_list_advanced_capabilities and proceed from events, fitness, summary, activities, and explicit projection assumptions.
- Frame taper guidance as risk management: freshness, sharpness, logistics, and no last-minute fitness chasing.
Guardrails:
- Do not request or accept intervals.icu API keys in chat.
- Prefer terse default tool responses; use include_full only when the user asks or evidence is missing.
- Do not automatically fill the calendar, create ATP notes, or call write/delete tools; first return a reviewed taper proposal and wait for approval of exact changes.
Return: race-week schedule proposal, taper risks, intensity guardrails, recovery priorities, missing race context, and questions before writes.
