# R008 plan review — Step 3

Verdict: APPROVE

The revised Step 3 plan addresses the blocking concerns from R007:

- Date serialization is now scoped to the live-probed NOTE path while preserving existing WORKOUT behavior and leaving unprobed categories unchanged.
- NOTE `name` validation is explicitly scoped to creates (`event_id` omitted), avoiding an unprobed update-behavior regression.
- The schema text and public invalid-arguments message are included in the fix plan, so the new validation remains actionable to MCP clients/LLMs.
- The targeted NOTE regression and existing WORKOUT create/update tests are the right minimum verification set for this step.

Non-blocking implementation notes:

- Use the already-trimmed `EventID`, `Category`, and `Name` values when deciding whether this is a NOTE create missing a name.
- Keep category matching case-insensitive for internal validation/serialization, but continue sending the caller-provided category value rather than silently changing casing unless a separate schema/normalization change is intended.
- If the date helper is changed in `internal/intervals/events.go`, the NOTE fixture test should demonstrate the new serialized body; avoid broadening the helper beyond `WORKOUT` and `NOTE` without new probe evidence.
