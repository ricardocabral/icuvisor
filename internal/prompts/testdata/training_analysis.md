Prompt: Training analysis
Scope: start_date=2026-04-01, end_date=2026-04-30.
Resources: icuvisor://athlete-profile.
Tools: get_athlete_profile, get_fitness, get_training_summary, get_best_efforts, get_activities.
Do:
- Read profile first for timezone, sport settings, and units.
- Use fitness and summary rows for CTL/ATL/TSB, ramp, volume, load, and intensity mix.
- Use best efforts and recent activities only for context; keep raw rows terse unless the user asks for detail.
Guardrails:
- Do not request or accept intervals.icu API keys in chat.
- Prefer terse default tool responses; use include_full only when the user asks or evidence is missing.
Return: load/trend readout with notable changes, likely drivers, missing-data caveats, and 2-3 next-step questions or actions.
