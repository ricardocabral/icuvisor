Prompt: Weekly planning
Scope: week_start=2026-05-18.
Resources: icuvisor://athlete-profile, icuvisor://event-categories, icuvisor://workout-syntax.
Tools: get_athlete_profile, get_events, get_training_plan, get_activities, get_training_summary, icuvisor_list_advanced_capabilities.
Do:
- Read profile, planned events, and training-plan context before suggesting changes.
- If get_training_plan is unavailable in the active toolset, use icuvisor_list_advanced_capabilities and proceed from events/activities.
- Compare planned versus completed work where the week has already started.
- Use event categories and workout syntax resources by URI if the user asks for edits or workout details.
Guardrails:
- Do not request or accept intervals.icu API keys in chat.
- Prefer terse default tool responses; use include_full only when the user asks or evidence is missing.
Return: day-by-day plan, key load constraints, planned-vs-completed notes, and questions before any write tool is used.
