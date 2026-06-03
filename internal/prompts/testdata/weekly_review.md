Prompt: Weekly review
Scope: review the previous athlete-local week with a 7-day lookback and include next week only if requested.
Resources: icuvisor://athlete-profile, icuvisor://event-categories.
Tools: get_athlete_profile, get_wellness_data, get_fitness, get_training_summary, get_activities, get_events, get_training_plan, compute_zone_time, compute_load_balance, compute_compliance_rate, analyze_trend, icuvisor_list_advanced_capabilities.
Do:
- Read profile first to establish athlete-local timezone, current date, sport settings, and units; compare days only after converting to athlete-local dates.
- Use fitness, training summary, and compute_zone_time to summarize load, volume, intensity mix, and fatigue/freshness changes.
- Use compute_load_balance and compute_compliance_rate when available; otherwise call icuvisor_list_advanced_capabilities, continue from available reads, and name the missing helper.
- Review activities, race/other events, and training plan for planned-versus-completed work; include race date/priority when relevant and the upcoming-week preview only when include_next_week is true or the user asks.
- Use wellness data for sleep/readiness/HRV context; check `_meta.stale`, `_meta.missing_fields`, and provenance warnings.
- When readiness is present, cite `_meta.provenance.readiness.source` and `native_scale`; treat Garmin Body Battery, Oura readiness, Polar nightly recharge/ANS charge, WHOOP recovery, and unknown upstream readiness as provider-native signals, not a universal recovery score.
- If readiness is missing, null, stale, or absent, say that explicitly and do not invent, infer, or backfill a readiness score; use HRV, resting HR, sleep duration/quality/score, subjective fatigue/soreness/stress/feel/mood/motivation, and available `_native` provider fields as cautious supporting context only.
- Use analyze_trend only for specific trend questions; keep raw activity rows terse unless evidence is missing.
Guardrails:
- Do not request or accept intervals.icu API keys in chat.
- Prefer terse default tool responses; use include_full only when the user asks or evidence is missing.
- Do not call write or delete tools unless the user explicitly approves the exact change first.
- Do not auto-fill calendars or create ATP notes from the review; propose exact changes for user approval first.
Return: weekly review with wins, concerns, planned-vs-completed gaps, wellness caveats with provider/source labels, load/intensity evidence, next-week preview when requested, and explicit follow-up questions before any write.
