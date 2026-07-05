Prompt: Recovery check
Scope: date=2026-05-14, lookback_days=10.
Resources: icuvisor://athlete-profile.
Tools: get_athlete_profile, get_wellness_data, get_fitness, get_today.
Do:
- Read wellness first; preserve sleepQuality 1-4 and sleepScore 0-100 as separate fields.
- Check HRV, resting HR, readiness, fatigue, soreness, mood, and any `_meta.stale`, `_meta.missing_fields`, or provenance warnings.
- When readiness is present, cite `_meta.provenance.readiness.source` and `native_scale`; treat Garmin Body Battery, Oura readiness, Polar nightly recharge/ANS charge, WHOOP recovery, and unknown upstream readiness as provider-native signals, not a universal recovery score.
- If readiness is missing or null, say that plainly before interpreting other signals; do not invent a readiness score.
- Use HRV, resting HR, sleepSecs, sleepQuality (1-4), sleepScore (0-100), fatigue, soreness, stress, feel, mood, motivation, and available `_native` provider fields only as cautious supporting evidence.
- Use fitness only to contextualize recent load; do not turn recovery into a full training analysis.
- For today-specific or indoor/outdoor questions, call get_today and use only its weather.status/provenance, planned_events[].indoor, tags, and completed-activity context; if weather.status is forecast_unavailable, say weather is unavailable from icuvisor and do not invent conditions.
- Do not infer separate indoor/outdoor FTP from planned_events[].indoor or zone boundaries. Use get_athlete_profile sport_settings[].indoor_ftp_watts only when present; otherwise ask or confirm how to adjust the workout.
- When suggesting an indoor alternative, present it as a chat recommendation or preview first; do not write calendar changes, and do not create a second active workout for the same planned session unless the user explicitly approves replacing or adding one.
Guardrails:
- Do not request or accept intervals.icu API keys in chat.
- Prefer terse default tool responses; use include_full only when the user asks or evidence is missing.
- Do not call write or delete tools for indoor/outdoor adaptation unless the user has reviewed and approved the exact change.
Return: green/yellow/red recovery guidance, the main evidence with provider/source labels, stale or missing fields, readiness-score absence when applicable, weather availability when relevant, and a 24-48h training adjustment.
