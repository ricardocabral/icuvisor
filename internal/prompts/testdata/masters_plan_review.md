Prompt: Masters plan review
Scope: planned_start=2026-05-18, planned_end=2026-06-01, history_lookback_days=28, baseline_lookback_days=56, race_date=2026-06-07, race_name=A Race.
Resources: icuvisor://athlete-profile, icuvisor://event-categories, icuvisor://analysis-formulas.
Tools: get_athlete_profile, resolve_calendar_dates, get_events, get_training_plan, get_activities, get_fitness, get_training_summary, compute_baseline, compute_load_balance, get_fitness_projection, get_wellness_data, icuvisor_list_advanced_capabilities.
Do:
- Establish the athlete-local timezone from the profile and resolve every relative date, weekday, countdown, or stale conversation anchor with resolve_calendar_dates before comparing dates.
- Partition non-overlapping personal-baseline/history, completed, planned, and race windows; resolve completed history immediately before planned_start and personal baseline immediately before that history, do not mix a completed window with later planned rows, and retain current-day wellness only as partial context.
- Read sourced events, training-plan rows, and activities for their assigned windows. Fetch every needed page before claiming coverage; when pagination, truncation, an unavailable tool, or missing rows prevent that, label the coverage partial and do not treat it as complete.
- Use get_fitness, get_training_summary, and compute_load_balance only for sourced load context; compute_load_balance is a window aggregate and cannot classify an individual session as hard. Read get_wellness_data for freshness and provider-native caveats, not a universal readiness score.
- Confirm a calendar race by matching event evidence. If no matching event is found, label a supplied race_date a scenario anchor rather than observed race evidence.
- For a personal baseline, use compute_baseline for one eligible metric at a time. Retain its status, n_baseline, n_current, min_samples, missing-day counts, freshness_status, caveats, `_meta.method`, and `_meta.formula_ref`; never combine metric results into a readiness or risk score.
- Assess hard-session spacing only when the athlete identifies sessions as hard or sourced activity/plan rows provide sufficiently detailed intensity evidence. Titles, aggregate load, calendar proximity, zones that are absent or invalid, and age cannot classify a session as hard; otherwise report insufficient evidence rather than calculating a gap in chat.
- Use get_fitness_projection only with copyable plan targets or athlete-supplied values. Surface every returned `_meta.assumptions`, never present default weekly_ramp_pct or recovery_week_cadence values as plan evidence or a masters recommendation, and report insufficient explicit load targets instead of inventing them.
- Treat availability and requested duration as athlete-stated context, not inferred hard constraints or an implied session count. If compute_baseline, get_fitness_projection, or another needed analyzer is unavailable, call icuvisor_list_advanced_capabilities, name the gap, and do not calculate a substitute in chat.
- For ambiguous or unavailable hard-session or plan detail; absent or invalid zones; short, partial, truncated, or missing historical coverage; missing, stale, or partial wellness; missing or provider-native readiness; missing race context; or insufficient explicit projection targets: name the missing evidence, make no comparison or conclusion for that affected dimension, and ask one focused question.
- Return visibly separate sections in this order: Observed tool evidence (tool, athlete-local window, freshness/coverage); Athlete-stated preferences (availability and requested duration only); Cautious interpretation; Insufficient evidence and focused questions; Reviewable proposals.
Guardrails:
- This workflow is absolutely read-only: never call write or delete tools, including after approval. A calendar change remains a conditional, unapplied proposal.
- Do not request, infer, derive, or use age or date of birth. Masters is an audience label only, never an age-derived policy or universal cutoff.
- Do not make medical, diagnostic, treatment, or injury-risk claims, and do not invent a black-box readiness or risk score.
- Do not request or accept intervals.icu API keys in chat.
- Prefer terse default tool responses; use include_full only when evidence is missing.
Return: five visibly separate sections: sourced evidence, athlete-stated preferences, cautious interpretation, insufficient-evidence questions, and conditional unapplied proposals.
