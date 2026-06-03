Prompt: Plan health review
Scope: planned_start=2026-05-18, planned_end=2026-06-01, completed_lookback_days=21, race_date=2026-06-07, race_name=A Race.
Resources: icuvisor://athlete-profile, icuvisor://event-categories, icuvisor://analysis-formulas.
Tools: get_athlete_profile, get_events, get_training_plan, get_activities, compute_compliance_rate, get_fitness, get_training_summary, compute_load_balance, get_fitness_projection, get_wellness_data, icuvisor_list_advanced_capabilities.
Do:
- Read profile first for timezone, units, sport settings, and today's athlete-local date; convert all windows before comparing days.
- Separate completed-lookback, planned-window, and race-scenario dates; do not mix current-day or post-window wellness into completed adherence evidence.
- Read events and training plan for planned workouts and races; if no race event is found, say so and treat any supplied race_date as a scenario anchor only.
- Use compute_compliance_rate for scheduled-vs-completed adherence, then get_fitness, get_training_summary, compute_load_balance, and get_fitness_projection for load/form trajectory and future assumptions.
- Quote analyzer `_meta.method`, `_meta.assumptions`, `_meta.formula_ref`, missing-days, and sample-size caveats where present; call icuvisor_list_advanced_capabilities and name missing helpers when full-tool analyzers are unavailable.
- Read recent wellness for sleep/readiness/HRV caveats; treat current-day `_meta.as_of` as partial-day context only and do not infer readiness when data is stale, absent, or missing key fields.
- Treat planned deload or recovery weeks as intentional load reductions unless compliance, wellness, or form evidence shows a problem.
Guardrails:
- Do not request or accept intervals.icu API keys in chat.
- Prefer terse default tool responses; use include_full only when the user asks or evidence is missing.
- Do not invent a black-box plan-health score; use only surfaced values and label risk low/medium/high with evidence.
- Do not create a season plan, fill a calendar, or act as an autonomous physiology model.
- Do not call write or delete tools unless the user has reviewed and approved the exact proposal first.
Return: data coverage, adherence, load/form trajectory, transparent risk table, deload/recovery caveats, race-date risk when anchored, and reviewed proposal/questions before any write.
