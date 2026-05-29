Prompt: Recovery check
Scope: date=2026-05-14, lookback_days=10.
Resources: icuvisor://athlete-profile.
Tools: get_athlete_profile, get_wellness_data, get_fitness.
Do:
- Read wellness first; preserve sleepQuality 1-4 and sleepScore 0-100 as separate fields.
- Check HRV, resting HR, readiness, fatigue, soreness, mood, and any `_meta.stale`, `_meta.missing_fields`, or provenance warnings.
- If readiness is missing or null, say that plainly before interpreting other signals; do not invent a readiness score.
- Use HRV, resting HR, sleepSecs, sleepQuality (1-4), sleepScore (0-100), fatigue, soreness, stress, feel, mood, motivation, and available `_native` provider fields only as cautious supporting evidence.
- Use fitness only to contextualize recent load; do not turn recovery into a full training analysis.
Guardrails:
- Do not request or accept intervals.icu API keys in chat.
- Prefer terse default tool responses; use include_full only when the user asks or evidence is missing.
Return: green/yellow/red recovery guidance, the main evidence, stale or missing fields, readiness-score absence when applicable, and a 24-48h training adjustment.
