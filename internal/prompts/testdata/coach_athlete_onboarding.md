Prompt: Coach athlete onboarding
Scope: athlete_id=i12345, start_date=2026-05-01, end_date=2026-05-28.
Resources: icuvisor://athlete-profile, icuvisor://event-categories.
Tools: list_athletes, select_athlete, get_athlete_profile, get_activities, get_training_summary, get_fitness, get_wellness_data, get_events, get_training_plan, icuvisor_list_advanced_capabilities.
Do:
- Start with list_athletes; if athlete_id is supplied, select_athlete for that normalized selector, otherwise ask the coach which roster athlete to onboard.
- Before summarizing data, confirm the selected athlete's canonical ID/label and state that the coach must already have authorization and athlete consent to view and analyze this data.
- Read profile first for identity, timezone, units, thresholds/zones, and `_meta.warnings`; then check recent activities, training summary, fitness, wellness/HRV, events/races, and training-plan context.
- Call icuvisor_list_advanced_capabilities when a checklist item depends on a missing or ACL-hidden tool; name unavailable data rather than guessing.
- Produce checklist rows for thresholds/zones, activity coverage, wellness/HRV baseline, races/events/goals, devices/sources/sync gaps, missing data warnings, and coach follow-up questions.
- Keep this onboarding read-only; propose any calendar/settings changes separately and wait for explicit reviewed approval before using write tools.
Guardrails:
- athlete_id selects a configured athlete; it is not a credential, consent artifact, invite token, or proof of upstream authorization.
- Do not request or accept intervals.icu API keys, OAuth tokens, invite links, or private identifiers in chat.
- Do not expose raw wellness/location details beyond what the coach needs for onboarding; ask the coach to review/redact any summary before sharing.
- Do not run live account tests or claim upstream roster import, consent capture, device inventory, or bulk team analytics exists.
Return: authorized-athlete confirmation, onboarding checklist with pass/warn/missing status, baseline profile, goals/races questions, device/source caveats, and first coach actions.
