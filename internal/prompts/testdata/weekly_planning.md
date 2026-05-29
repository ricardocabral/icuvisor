Prompt: Weekly planning
Scope: week_start=2026-05-18.
Resources: icuvisor://athlete-profile, icuvisor://event-categories, icuvisor://workout-syntax.
Tools: get_athlete_profile, get_events, get_training_plan, get_fitness, get_training_summary, get_activities, compute_compliance_rate, icuvisor_list_advanced_capabilities.
Do:
- Read profile/timezone, then ask or confirm the planning anchor: race date, priority/category, goal, and constraints when missing.
- Read planned events and active training-plan context before suggesting changes.
- Use fitness, training summary, recent activities, and compliance to summarize current load, fatigue/freshness, and planned-versus-completed work.
- If get_training_plan or compute_compliance_rate is unavailable, call icuvisor_list_advanced_capabilities and proceed from get_events, get_fitness, get_training_summary, and activities.
- Use event categories and workout syntax resources by URI if the user asks for edits or workout details.
- Draft a season/block/week proposal with assumptions, load constraints, and follow-up questions before any edits.
- When proposing endurance workouts, prefer the structured `workout_doc` form on write tools and include any coaching notes via `description` on the same event; both fields coexist, but `description` replaces the upstream description/DSL on writes, so for updates include the desired `workout_doc` whenever preserving structured steps matters. Call `validate_workout` before the write if uncertain about the DSL syntax, and read `icuvisor://workout-syntax` for the cheat sheet and common mistakes.
- When the user asks for gym or strength work, schedule a simple `NOTE` time block or free-text supported calendar event; do not invent structured exercises, sets, reps, loads, or rest periods unless documented upstream strength-training support is available.
- Before bulk calendar/workout writes, validate or preview one representative structured payload, perform one representative write, read it back, and inspect validation warnings, existing write `_meta` warning fields such as `workout_doc_warning` when present, and `workout_doc_summary`/stored description before writing the rest. Avoid parallel bulk writes while schema wording, warning metadata, or description/`workout_doc` preservation semantics are ambiguous.
- For approved writes, use event categories and workout syntax resources, validate workout_doc when uncertain, and write only the exact user-approved changes.
Guardrails:
- Do not request or accept intervals.icu API keys in chat.
- Prefer terse default tool responses; use include_full only when the user asks or evidence is missing.
- Do not automatically fill the calendar, create ATP notes, or call write/delete tools; first return a reviewed proposal and wait for approval of exact changes.
Return: season/block/week proposal, race context, current load, active-plan and event evidence, compliance notes, and questions before writes.
