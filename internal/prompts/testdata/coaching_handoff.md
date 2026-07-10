Prompt: Coaching conversation handoff
Scope: lookback_days=42, race_context_days=180.
Resources: icuvisor://athlete-profile, icuvisor://event-categories.
Tools: get_athlete_profile, resolve_calendar_dates, get_events, get_training_plan, get_fitness, get_training_summary, get_activities, get_wellness_data, icuvisor_list_advanced_capabilities.
Do:
- Call get_athlete_profile for the athlete timezone and resolve_calendar_dates for today's athlete-local anchor or any relative date; state the generated-on date, timezone, lookback window, and race-context window.
- Return compact Markdown sections in this exact order: Handoff scope; Conversation-stated context (Goals, Constraints, Accepted decisions); Icuvisor evidence; Current plan state; Data gaps and unresolved questions; Next actions.
- Put only user-stated goals and constraints in Conversation-stated context. Put a decision in Accepted decisions only when the user explicitly stated or accepted it; never promote an assistant suggestion, model summary, or calendar row into a user decision.
- Keep tool-sourced facts separate. Render Icuvisor evidence as Claim | Source tool | Athlete-local evidence date/window | Freshness/as-of, and keep Current plan state limited to sourced event and training-plan state.
- Distinguish the date/window when evidence applies from returned as_of or provider freshness. Preserve trustworthy freshness markers; write 'not provided' when none exists, and never invent fetched_at or another retrieval timestamp.
- Use terse get_events, get_training_plan, get_fitness, get_training_summary, get_activities, and get_wellness_data only as needed for durable context; do not dump full history.
- Surface _meta.stale, _meta.missing_fields, unavailable or Strava-blocked data, current-day partial data, and unresolved tool failures. Never treat a missing value as zero or fill a tool gap from chat memory.
- When next_page_token is present, fetch every page needed before claiming completeness or label the evidence partial with the covered window/count; never include the opaque token in the handoff.
- If an advanced analyzer material to the conversation is unavailable, call icuvisor_list_advanced_capabilities, name the missing capability, and preserve the unresolved question instead of calculating a substitute in chat.
- Ask the athlete to review the Markdown, then manually copy it into a fresh Claude, ChatGPT, Cursor, or other client conversation; do not claim any client automatically imports, persists, or remembers it.
Guardrails:
- This workflow is read-only: do not call write or delete tools.
- Never use include_full, raw streams, raw tool payloads, or full histories for the handoff.
- Exclude credentials, API/OAuth tokens, secrets, raw athlete identifiers, local or config paths, pagination tokens, and transport or debug metadata.
- Omit health details, precise locations, and private free-text notes by default; include only the minimum the user explicitly approves.
- Do not make diagnoses, unsupported physiological conclusions, or claims sourced only from model memory.
Return: the six-section Markdown handoff with explicit source separation, athlete-local dates, freshness and coverage caveats, unresolved questions, next actions, and a manual-review reminder.
