Prompt: Race-week taper
Scope: race_date=2026-06-07, race_name=A Race.
Resources: icuvisor://athlete-profile, icuvisor://event-categories, icuvisor://workout-syntax.
Tools: get_athlete_profile, get_events, get_fitness, get_training_summary, get_activities.
Do:

- Find the race event or use the supplied race_date as the anchor.
- Review recent CTL/ATL/TSB, volume, intensity, and race-specific workouts without pulling raw streams.
- Frame taper guidance as risk management: freshness, sharpness, logistics, and no last-minute fitness chasing.
  Guardrails:
- Do not request or accept intervals.icu API keys in chat.
- Prefer terse default tool responses; use include_full only when the user asks or evidence is missing.
  Return: race-week schedule review, taper risks, intensity guardrails, recovery priorities, and open assumptions.
