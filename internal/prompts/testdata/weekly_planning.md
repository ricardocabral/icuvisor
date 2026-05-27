Prompt: Weekly planning
Scope: week_start=2026-05-18.
Resources: icuvisor://athlete-profile, icuvisor://event-categories, icuvisor://workout-syntax.
Tools: get_athlete_profile, get_events, get_training_plan, get_activities, get_training_summary, icuvisor_list_advanced_capabilities.
Do:
- Read profile, planned events, and training-plan context before suggesting changes.
- If get_training_plan is unavailable in the active toolset, use icuvisor_list_advanced_capabilities and proceed from events/activities.
- Compare planned versus completed work where the week has already started.
- Use event categories and workout syntax resources by URI if the user asks for edits or workout details.
- When proposing workouts, prefer the structured `workout_doc` form on write tools and include any coaching notes via `description` on the same event; both fields coexist, but `description` replaces the upstream description/DSL on writes, so for updates include the desired `workout_doc` whenever preserving structured steps matters. Call `validate_workout` before the write if uncertain about the DSL syntax, and read `icuvisor://workout-syntax` for the cheat sheet and common mistakes.
- Before bulk calendar/workout writes, validate or preview one representative structured payload, perform one representative write, read it back, and inspect validation warnings, existing write `_meta` warning fields such as `workout_doc_warning` when present, and `workout_doc_summary`/stored description before writing the rest. Avoid parallel bulk writes while schema wording, warning metadata, or description/`workout_doc` preservation semantics are ambiguous.
Guardrails:
- Do not request or accept intervals.icu API keys in chat.
- Prefer terse default tool responses; use include_full only when the user asks or evidence is missing.
Return: day-by-day plan, key load constraints, planned-vs-completed notes, and questions before any write tool is used.
