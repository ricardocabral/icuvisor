# Plan Review — Step 1: Implement `LogValue`

Decision: **approved**

The revised Step 1 plan addresses the prior compile-shape and redaction concerns. Returning `slog.GroupValue(...)` directly with an explicit allowlist is the right approach for `slog.LogValuer`, and summarizing coach roster data as `coach_athletes_count` avoids exposing nested athlete IDs or labels.

## Notes for implementation

- Keep the allowlist exactly as planned: `api_base_url`, `default_athlete_id`, `http_bind`, `coach_athletes_count`, `delete_mode`, and `toolset`.
- Preserve the redacted/presence-marker semantics for `default_athlete_id`; do not emit raw `Config.AthleteID`, `Coach.DefaultAthleteID`, roster IDs, or labels.
- Use string conversions for enum-like values (`DeleteMode.String()`, `Toolset.String()`) so structured output remains stable even for zero values.
- Place `LogValue` next to `String()` and do not change the existing `String()` output in this step.

No further plan changes are required before coding Step 1.
